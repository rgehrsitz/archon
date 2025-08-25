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

### Important: Default Proxy Behavior

- If `proxyPolicy` is NOT present in project settings, the network proxy is DISABLED.
- In this state, `HostService.NetRequest()` returns `NOT_IMPLEMENTED` ("Proxy executor not configured").

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
  - Builds a file-backed `FileSecretsStore` at `<projectPath>/.archon/secrets.json` and wraps with `PolicySecretsStore`
  - Both are injected into `HostService` via `NewHostService`
- Tests:
  - `internal/api/plugin_service_secrets_test.go` covers permission enforcement and redaction policy end-to-end via PluginService
  - `internal/plugins/secrets_file_store_test.go` covers file-backed secrets loading, listing, and concurrency safety
  - `internal/plugins/host_test.go` covers general host permission enforcement (read repo, query)
