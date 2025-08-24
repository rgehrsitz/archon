package plugins

import (
    "context"
    "testing"
)

func TestPolicyProxy_MethodDenied(t *testing.T) {
    fp := &fakeProxy{resp: ProxyResponse{Status: 200}}
    policy := ProxyPolicy{AllowedMethods: []string{"GET"}}
    p := NewPolicyProxyExecutor(fp, policy)

    _, err := p.Do(context.Background(), ProxyRequest{Method: "POST", URL: "https://example.com"})
    if err == nil {
        t.Fatalf("expected method_not_allowed error")
    }
    if pe, ok := err.(*PolicyError); !ok || pe.Reason != "method_not_allowed" {
        t.Fatalf("expected PolicyError method_not_allowed, got %#v", err)
    }
}

func TestPolicyProxy_HostDenied(t *testing.T) {
    fp := &fakeProxy{resp: ProxyResponse{Status: 200}}
    policy := ProxyPolicy{DenyHostSuffixes: []string{"internal"}}
    p := NewPolicyProxyExecutor(fp, policy)

    _, err := p.Do(context.Background(), ProxyRequest{Method: "GET", URL: "https://svc.internal/path"})
    if err == nil {
        t.Fatalf("expected host_not_allowed error")
    }
    if pe, ok := err.(*PolicyError); !ok || pe.Reason != "host_not_allowed" {
        t.Fatalf("expected PolicyError host_not_allowed, got %#v", err)
    }
}

func TestPolicyProxy_HeaderRedaction(t *testing.T) {
    fp := &fakeProxy{resp: ProxyResponse{Status: 200, Headers: map[string]string{
        "Authorization": "Bearer abc",
        "Set-Cookie":    "a=b",
        "X-Other":       "ok",
    }, Body: []byte("ok")}}

    policy := ProxyPolicy{RedactResponseHeaders: []string{"authorization", "set-cookie"}}
    p := NewPolicyProxyExecutor(fp, policy)

    resp, err := p.Do(context.Background(), ProxyRequest{Method: "GET", URL: "https://example.com"})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if resp.Headers["Authorization"] != "REDACTED" {
        t.Fatalf("Authorization not redacted: %#v", resp.Headers)
    }
    if resp.Headers["Set-Cookie"] != "REDACTED" {
        t.Fatalf("Set-Cookie not redacted: %#v", resp.Headers)
    }
    if resp.Headers["X-Other"] != "ok" {
        t.Fatalf("Unexpected change to X-Other: %#v", resp.Headers)
    }
}

func TestPolicySecretsStore_RedactValues(t *testing.T) {
    base := NewInMemorySecretsStore(map[string]string{"k1": "v1"})
    ps := NewPolicySecretsStore(base, SecretsPolicy{ReturnValues: false})

    val, ok, err := ps.Get(context.Background(), "k1")
    if err != nil || !ok || val == nil {
        t.Fatalf("unexpected result: val=%#v ok=%v err=%v", val, ok, err)
    }
    if !val.Redacted || val.Value != "" || val.Name != "k1" {
        t.Fatalf("expected redacted secret without value, got %#v", val)
    }
}
