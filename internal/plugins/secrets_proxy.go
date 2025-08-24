package plugins

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SecretsStore abstracts retrieval and listing of secrets for plugins
// Implementations may back this with OS keychain, file, or memory for tests.
type SecretsStore interface {
	Get(ctx context.Context, key string) (*SecretValue, bool, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

// InMemorySecretsStore is a simple map-based store for tests
type InMemorySecretsStore struct {
	items map[string]*SecretValue
}

func NewInMemorySecretsStore(kv map[string]string) *InMemorySecretsStore {
	items := make(map[string]*SecretValue)
	for k, v := range kv {
		items[k] = &SecretValue{Name: k, Value: v}
	}
	return &InMemorySecretsStore{items: items}
}

func (s *InMemorySecretsStore) Get(ctx context.Context, key string) (*SecretValue, bool, error) {
	v, ok := s.items[key]
	if !ok {
		return nil, false, nil
	}
	// Return a copy to avoid external mutation in tests
	c := *v
	return &c, true, nil
}

func (s *InMemorySecretsStore) List(ctx context.Context, prefix string) ([]string, error) {
	var out []string
	for k := range s.items {
		if len(prefix) == 0 || (len(k) >= len(prefix) && k[:len(prefix)] == prefix) {
			out = append(out, k)
		}
	}
	return out, nil
}

// ProxyPolicy defines network execution policies enforced before/after HTTP calls
type ProxyPolicy struct {
	// Allowed HTTP methods (uppercased). If empty, allow common safe defaults: GET, POST, PUT, DELETE, PATCH, HEAD
	AllowedMethods []string
	// Allow only hosts that end with these suffixes (case-insensitive). If empty, allow all (minus denies)
	AllowHostSuffixes []string
	// Deny hosts that end with these suffixes (case-insensitive). Evaluated before allowlist
	DenyHostSuffixes []string
	// Redact these response header names (case-insensitive) in the returned ProxyResponse
	RedactResponseHeaders []string
}

// PolicyProxyExecutor wraps a ProxyExecutor and enforces ProxyPolicy
type PolicyProxyExecutor struct {
	Inner  ProxyExecutor
	Policy ProxyPolicy
}

func NewPolicyProxyExecutor(inner ProxyExecutor, policy ProxyPolicy) *PolicyProxyExecutor {
	return &PolicyProxyExecutor{Inner: inner, Policy: policy}
}

func (p *PolicyProxyExecutor) Do(ctx context.Context, req ProxyRequest) (ProxyResponse, error) {
	// Method restriction
	if !p.methodAllowed(req.Method) {
		return ProxyResponse{}, &PolicyError{Reason: "method_not_allowed"}
	}

	// Host allow/deny enforcement
	if !p.hostAllowed(req.URL) {
		return ProxyResponse{}, &PolicyError{Reason: "host_not_allowed"}
	}

	// Delegate
	resp, err := p.Inner.Do(ctx, req)
	if err != nil {
		return resp, err
	}

	// Redact response headers as configured
	if len(p.Policy.RedactResponseHeaders) > 0 && resp.Headers != nil {
		resp.Headers = redactHeaders(resp.Headers, p.Policy.RedactResponseHeaders)
	}
	return resp, nil
}

// PolicyError represents a policy violation
type PolicyError struct {
	Reason string
}

func (e *PolicyError) Error() string { return e.Reason }

func (p *PolicyProxyExecutor) methodAllowed(method string) bool {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" { return false }
	if len(p.Policy.AllowedMethods) == 0 {
		switch m {
		case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD":
			return true
		default:
			return false
		}
	}
	for _, allowed := range p.Policy.AllowedMethods {
		if strings.EqualFold(allowed, m) {
			return true
		}
	}
	return false
}

func (p *PolicyProxyExecutor) hostAllowed(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	// Deny has precedence
	for _, d := range p.Policy.DenyHostSuffixes {
		if hasHostSuffix(host, d) {
			return false
		}
	}
	// Allow list (if provided)
	if len(p.Policy.AllowHostSuffixes) == 0 {
		return true
	}
	for _, a := range p.Policy.AllowHostSuffixes {
		if hasHostSuffix(host, a) {
			return true
		}
	}
	return false
}

func hasHostSuffix(host, suffix string) bool {
	s := strings.ToLower(strings.TrimSpace(suffix))
	if s == "" { return false }
	return strings.HasSuffix(host, s)
}

func redactHeaders(hdrs map[string]string, names []string) map[string]string {
	if hdrs == nil { return nil }
	out := make(map[string]string, len(hdrs))
	for k, v := range hdrs {
		redacted := false
		for _, n := range names {
			if strings.EqualFold(k, n) {
				redacted = true
				break
			}
		}
		if redacted {
			out[k] = "REDACTED"
		} else {
			out[k] = v
		}
	}
	return out
}

// SecretsPolicy defines behavior for exposing secrets to plugins
type SecretsPolicy struct {
	// If false, secret values are not returned to plugins; only metadata with Redacted=true
	ReturnValues bool
}

// PolicySecretsStore wraps a SecretsStore to apply SecretsPolicy
type PolicySecretsStore struct {
	Inner  SecretsStore
	Policy SecretsPolicy
}

func NewPolicySecretsStore(inner SecretsStore, policy SecretsPolicy) *PolicySecretsStore {
	return &PolicySecretsStore{Inner: inner, Policy: policy}
}

func (p *PolicySecretsStore) Get(ctx context.Context, key string) (*SecretValue, bool, error) {
	val, ok, err := p.Inner.Get(ctx, key)
	if err != nil || !ok || val == nil {
		return val, ok, err
	}
	if !p.Policy.ReturnValues {
		// Redact the value before returning
		redacted := &SecretValue{
			Name:     val.Name,
			Value:    "",
			Redacted: true,
			Metadata: val.Metadata,
		}
		return redacted, ok, nil
	}
	// Ensure Redacted flag is consistent
	if val.Redacted && val.Value != "" {
		// Clear the value if marked Redacted by backend
		val.Value = ""
	}
	return val, ok, nil
}

func (p *PolicySecretsStore) List(ctx context.Context, prefix string) ([]string, error) {
	return p.Inner.List(ctx, prefix)
}

// ProxyExecutor abstracts outbound HTTP execution.
// HostService uses this to perform network requests under PermissionNet.
type ProxyExecutor interface {
	Do(ctx context.Context, req ProxyRequest) (ProxyResponse, error)
}

// HTTPProxyExecutor is a basic implementation using net/http Client.
// It should only be used in controlled contexts; production may enforce
// additional allowlists, size limits, and redaction.
type HTTPProxyExecutor struct {
	Client *http.Client
}

func NewHTTPProxyExecutor(timeout time.Duration) *HTTPProxyExecutor {
	return &HTTPProxyExecutor{Client: &http.Client{Timeout: timeout}}
}

func (e *HTTPProxyExecutor) Do(ctx context.Context, preq ProxyRequest) (ProxyResponse, error) {
	// Build request
	var body io.Reader
	if len(preq.Body) > 0 {
		body = bytes.NewReader(preq.Body)
	}
	r, err := http.NewRequestWithContext(ctx, preq.Method, preq.URL, body)
	if err != nil {
		return ProxyResponse{}, err
	}
	for k, v := range preq.Headers {
		r.Header.Set(k, v)
	}

	// Execute
	resp, err := e.Client.Do(r)
	if err != nil {
		return ProxyResponse{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	out := ProxyResponse{
		Status:  resp.StatusCode,
		Headers: map[string]string{},
		Body:    respBody,
	}
	for k, vals := range resp.Header {
		if len(vals) > 0 {
			out.Headers[k] = vals[0]
		}
	}
	return out, nil
}
