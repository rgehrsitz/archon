# Policy Configuration (Proxy and Secrets)

Archon loads network proxy and secrets policies from project settings at runtime in `PluginService.InitializePluginSystem()` and injects policy-wrapped backends into `HostService`.

## Settings Schema (JSON)

```json
{
  "proxyPolicy": {
    "allowedMethods": ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"],
    "allowHostSuffixes": ["example.com", "api.example.net"],
    "denyHostSuffixes": ["internal", "corp.local"],
    "redactResponseHeaders": ["authorization", "cookie", "set-cookie"]
  },
  "secretsPolicy": {
    "returnValues": false
  }
}
```

## Defaults

- proxyPolicy.allowedMethods: if empty, defaults to [GET, POST, PUT, DELETE, PATCH, HEAD]
- proxyPolicy.allowHostSuffixes: if empty, allow all hosts except those in deny list
- proxyPolicy.denyHostSuffixes: empty by default
- proxyPolicy.redactResponseHeaders: empty by default (no redaction)
- secretsPolicy.returnValues: false by default (values are redacted)

## Behavior

- Proxy policy is enforced by `PolicyProxyExecutor` before/after HTTP execution.
  - Method not allowed -> policy error; returned to plugin as remote failure
  - Host not allowed -> policy error; returned to plugin as remote failure
  - Configured response headers are redacted to `REDACTED`
- Secrets policy is enforced by `PolicySecretsStore`.
  - If `returnValues=false`, `SecretsGet` returns metadata with `Redacted=true` and empty `Value`
  - `SecretsList` is unaffected

## Implementation Notes

- Wired in `internal/api/plugin_service.go`:
  - Builds `HTTPProxyExecutor` and wraps with `PolicyProxyExecutor`
  - Builds an `InMemorySecretsStore` and wraps with `PolicySecretsStore`
  - Both are injected into `HostService` via `NewHostService`
- Tests:
  - `internal/plugins/policy_test.go` covers method deny, host deny, header redaction, and secrets value redaction
  - Existing `host_secrets_proxy_test.go` covers permission enforcement
