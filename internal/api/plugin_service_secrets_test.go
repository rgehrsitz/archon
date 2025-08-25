package api

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "testing"

    "github.com/rgehrsitz/archon/internal/errors"
    "github.com/rgehrsitz/archon/internal/logging"
    "github.com/rgehrsitz/archon/internal/plugins"
)

// helper to create a quiet test logger
func newTestLogger(t *testing.T) logging.Logger {
    t.Helper()
    cfg := logging.DefaultConfig()
    cfg.OutputConsole = false
    cfg.OutputFile = false
    l, _ := logging.NewLogger(cfg)
    return *l
}

// writeSecretsFile writes a map of key -> SecretValue to .archon/secrets.json
func writeSecretsFile(t *testing.T, projectPath string, data map[string]plugins.SecretValue) string {
    t.Helper()
    archonDir := filepath.Join(projectPath, ".archon")
    if err := os.MkdirAll(archonDir, 0o755); err != nil {
        t.Fatalf("failed to create .archon dir: %v", err)
    }
    secretsPath := filepath.Join(archonDir, "secrets.json")
    b, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        t.Fatalf("failed to marshal secrets: %v", err)
    }
    if err := os.WriteFile(secretsPath, b, 0o600); err != nil {
        t.Fatalf("failed to write secrets file: %v", err)
    }
    return secretsPath
}

func setupProjectAndPluginService(t *testing.T, settings map[string]any) (string, *ProjectService, *PluginService) {
    t.Helper()
    t.Setenv("ARCHON_DISABLE_INDEX", "1")

    projectPath := t.TempDir()

    // Create project with provided settings
    ps := NewProjectService()
    if _, env := ps.CreateProject(context.Background(), projectPath, settings); env.Code != "" {
        t.Fatalf("CreateProject failed: %+v", env)
    }

    // Init plugin service
    svc := NewPluginService(newTestLogger(t), ps)
    if env := svc.InitializePluginSystem(context.Background()); env.Code != "" {
        t.Fatalf("InitializePluginSystem failed: %+v", env)
    }

    return projectPath, ps, svc
}

func declarePermissions(t *testing.T, svc *PluginService, pluginID string, perms []plugins.Permission) {
    t.Helper()
    // Access the underlying permission manager via pluginManager
    if svc.pluginManager == nil {
        t.Fatalf("pluginManager not initialized")
    }
    svc.pluginManager.GetPermissionManager().DeclarePermissions(pluginID, perms)
}

func TestPluginService_SecretsPermissionsAndRedaction(t *testing.T) {
    ctx := context.Background()
    pluginID := "test.secrets.plugin"

    // Initial settings: redact values by default
    settings := map[string]any{
        "secretsPolicy": map[string]any{
            "returnValues": false,
        },
    }

    projectPath, ps, svc := setupProjectAndPluginService(t, settings)

    // Prepare secrets file in project
    secrets := map[string]plugins.SecretValue{
        "jira.token":   {Name: "jira.token", Value: "secret_jira_token", Redacted: false},
        "jira.user":    {Name: "jira.user", Value: "alice", Redacted: false},
        "db.password":  {Name: "db.password", Value: "supersecret", Redacted: false},
        "service.meta": {Name: "service.meta", Value: "v", Redacted: false, Metadata: map[string]any{"env": "dev"}},
    }
    _ = writeSecretsFile(t, projectPath, secrets)

    // Re-initialize plugin system to load secrets from file
    if env := svc.InitializePluginSystem(ctx); env.Code != "" {
        t.Fatalf("re-InitializePluginSystem failed: %+v", env)
    }

    // Declare permissions for our fake plugin
    declarePermissions(t, svc, pluginID, []plugins.Permission{
        plugins.Permission("secrets:jira*"),
        plugins.Permission("secrets:db*"),
    })

    // 1) Unauthorized access before grants
    if val, env := svc.PluginSecretsGet(ctx, pluginID, "jira.token"); env.Code != errors.ErrUnauthorized {
        t.Fatalf("expected UNAUTHORIZED for SecretsGet before grant, got: val=%+v err=%+v", val, env)
    }
    if keys, env := svc.PluginSecretsList(ctx, pluginID, "jira"); env.Code != errors.ErrUnauthorized {
        t.Fatalf("expected UNAUTHORIZED for SecretsList before grant, got: keys=%v err=%+v", keys, env)
    }

    // 2) Grant jira prefix and verify list/get behavior
    if env := svc.GrantPermission(ctx, pluginID, "secrets:jira*", false, 0); env.Code != "" {
        t.Fatalf("GrantPermission failed: %+v", env)
    }

    // List should return only jira.* keys
    keys, env := svc.PluginSecretsList(ctx, pluginID, "jira")
    if env.Code != "" {
        t.Fatalf("SecretsList failed: %+v", env)
    }
    sort.Strings(keys)
    expected := []string{"jira.token", "jira.user"}
    if len(keys) != len(expected) || keys[0] != expected[0] || keys[1] != expected[1] {
        t.Fatalf("unexpected keys: %v expected %v", keys, expected)
    }

    // db prefix should still be unauthorized
    if keys, env := svc.PluginSecretsList(ctx, pluginID, "db"); env.Code != errors.ErrUnauthorized {
        t.Fatalf("expected UNAUTHORIZED for db prefix list, got: keys=%v err=%+v", keys, env)
    }

    // Get returns redacted value by default policy
    sv, env := svc.PluginSecretsGet(ctx, pluginID, "jira.token")
    if env.Code != "" {
        t.Fatalf("SecretsGet failed: %+v", env)
    }
    if sv == nil || !sv.Redacted || sv.Value != "" || sv.Name != "jira.token" {
        t.Fatalf("expected redacted secret without value, got: %+v", sv)
    }

    // 3) Toggle policy to return values; re-init and re-declare/grant
    if env := ps.UpdateProjectSettings(ctx, map[string]any{
        "secretsPolicy": map[string]any{"returnValues": true},
    }); env.Code != "" {
        t.Fatalf("UpdateProjectSettings failed: %+v", env)
    }
    if env := svc.InitializePluginSystem(ctx); env.Code != "" {
        t.Fatalf("re-InitializePluginSystem (after policy toggle) failed: %+v", env)
    }
    declarePermissions(t, svc, pluginID, []plugins.Permission{
        plugins.Permission("secrets:jira*"),
    })
    if env := svc.GrantPermission(ctx, pluginID, "secrets:jira*", false, 0); env.Code != "" {
        t.Fatalf("GrantPermission failed after toggle: %+v", env)
    }

    // Now value should be returned with Redacted=false
    sv2, env := svc.PluginSecretsGet(ctx, pluginID, "jira.token")
    if env.Code != "" {
        t.Fatalf("SecretsGet after toggle failed: %+v", env)
    }
    if sv2 == nil || sv2.Redacted || sv2.Value != "secret_jira_token" {
        t.Fatalf("expected unredacted value after toggle, got: %+v", sv2)
    }

    // Metadata should be preserved regardless of redaction
    if sv3, env := svc.PluginSecretsGet(ctx, pluginID, "service.meta"); env.Code == "" {
        // With only jira* grant, this should be unauthorized
        t.Fatalf("expected UNAUTHORIZED for service.meta get, got value: %+v", sv3)
    }
}
