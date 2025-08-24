package plugins

import (
	"bytes"
	"context"
	"io"
	"net/http"
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
