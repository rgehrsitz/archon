package plugins

import (
	"context"
	"testing"

	"github.com/rgehrsitz/archon/internal/logging"
)

// fakeProxy implements ProxyExecutor for tests
type fakeProxy struct {
	last ProxyRequest
	resp ProxyResponse
	err  error
}

func TestHostService_ProxyDisabled_WhenNotConfigured(t *testing.T) {
	basePath, _, _, ns, idx, pm, _, pluginID := setupHostTest(t)
	_ = basePath

	logger := logging.NewTestLogger()
	// secretsStore=nil, proxyExecutor=nil -> proxy disabled
	hs := NewHostService(logger, ns, nil, idx, pm, nil, nil)

	ctx := context.Background()

	// Declare and grant net permission so failure reason isn't authorization
	declared := pm.GetDeclaredPermissions(pluginID)
	declared = append(declared, PermissionNet)
	pm.DeclarePermissions(pluginID, declared)
	env := pm.GrantPermission(pluginID, PermissionNet, false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant net: %+v", env)
	}

	req := ProxyRequest{Method: "GET", URL: "https://example.test/hello"}
	_, env = hs.NetRequest(ctx, pluginID, req)
	if env.Code != "NOT_IMPLEMENTED" { // Proxy executor not configured
		t.Fatalf("expected NOT_IMPLEMENTED when proxy disabled, got: %+v", env)
	}
}

func (f *fakeProxy) Do(ctx context.Context, req ProxyRequest) (ProxyResponse, error) {
	f.last = req
	return f.resp, f.err
}

func TestHostService_Secrets_PermissionEnforced(t *testing.T) {
	basePath, _, _, ns, idx, pm, _, pluginID := setupHostTest(t)
	_ = basePath

	// Configure secrets store
	ss := NewInMemorySecretsStore(map[string]string{
		"jira.token":   "abc123",
		"jira.user":    "alice",
		"github.token": "zzz",
	})

	logger := logging.NewTestLogger()
	hs := NewHostService(logger, ns, nil, idx, pm, ss, nil)

	ctx := context.Background()

	// Declare secrets permission pattern for plugin (append to existing declarations)
	declared := pm.GetDeclaredPermissions(pluginID)
	declared = append(declared, Permission("secrets:jira*"))
	pm.DeclarePermissions(pluginID, declared)

	// Without grant -> unauthorized
	val, env := hs.SecretsGet(ctx, pluginID, "jira.token")
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
	}
	if val != nil {
		t.Fatalf("expected nil secret on unauthorized access")
	}

	// Grant permission -> can get jira.*
	env = pm.GrantPermission(pluginID, Permission("secrets:jira*"), false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant permission: %+v", env)
	}

	val, env = hs.SecretsGet(ctx, pluginID, "jira.token")
	if env.Code != "" {
		t.Fatalf("unexpected error: %+v", env)
	}
	if val == nil || val.Value != "abc123" || val.Name != "jira.token" {
		t.Fatalf("unexpected secret value: %#v", val)
	}

	// List by prefix -> only jira.* keys
	keys, env := hs.SecretsList(ctx, pluginID, "jira")
	if env.Code != "" {
		t.Fatalf("unexpected error listing: %+v", env)
	}
	// Expect at least jira.token and jira.user
	found := map[string]bool{}
	for _, k := range keys {
		found[k] = true
	}
	if !found["jira.token"] || !found["jira.user"] {
		t.Fatalf("expected jira.* keys, got %v", keys)
	}
	if found["github.token"] {
		t.Fatalf("did not expect github.token in jira* list: %v", keys)
	}

	// Access outside scope -> unauthorized
	_, env = hs.SecretsGet(ctx, pluginID, "github.token")
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED for github.token, got: %+v", env)
	}
}

func TestHostService_NetworkProxy_PermissionEnforced(t *testing.T) {
	basePath, _, _, ns, idx, pm, _, pluginID := setupHostTest(t)
	_ = basePath

	fp := &fakeProxy{resp: ProxyResponse{Status: 200, Headers: map[string]string{"X-Test": "1"}, Body: []byte("ok")}}
	logger := logging.NewTestLogger()
	hs := NewHostService(logger, ns, nil, idx, pm, nil, fp)

	ctx := context.Background()

	// Declare net permission (append)
	declared := pm.GetDeclaredPermissions(pluginID)
	declared = append(declared, PermissionNet)
	pm.DeclarePermissions(pluginID, declared)

	req := ProxyRequest{Method: "GET", URL: "https://example.test/hello"}

	// Without grant -> unauthorized
	resp, env := hs.NetRequest(ctx, pluginID, req)
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
	}
	_ = resp

	// Grant -> success
	env = pm.GrantPermission(pluginID, PermissionNet, false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant net: %+v", env)
	}

	resp, env = hs.NetRequest(ctx, pluginID, req)
	if env.Code != "" {
		t.Fatalf("unexpected error: %+v", env)
	}
	if resp.Status != 200 || string(resp.Body) != "ok" || fp.last.Method != "GET" || fp.last.URL != req.URL {
		t.Fatalf("unexpected response or proxy invocation: resp=%#v last=%#v", resp, fp.last)
	}
}
