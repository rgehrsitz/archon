package plugins

import (
	"testing"
	"time"
)

func TestPermissionManager_DeclareAndCheck(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"

	// Test declaring permissions
	permissions := []Permission{PermissionReadRepo, PermissionWriteRepo}
	pm.DeclarePermissions(pluginID, permissions)

	declared := pm.GetDeclaredPermissions(pluginID)
	if len(declared) != 2 {
		t.Errorf("Expected 2 declared permissions, got %d", len(declared))
	}

	// Initially no permissions should be granted
	if pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permission should not be granted initially")
	}
}

func TestPermissionManager_GrantPermission(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"
	
	// Declare permissions
	permissions := []Permission{PermissionReadRepo, PermissionWriteRepo}
	pm.DeclarePermissions(pluginID, permissions)

	// Test granting a declared permission
	envelope := pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
	if envelope.Code != "" {
		t.Errorf("Failed to grant permission: %v", envelope.Message)
	}

	if !pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permission should be granted")
	}

	if pm.HasPermission(pluginID, PermissionWriteRepo) {
		t.Error("Ungranted permission should not be available")
	}

	// Test granting undeclared permission
	envelope = pm.GrantPermission(pluginID, PermissionNet, false, 0)
	if envelope.Code == "" {
		t.Error("Should not be able to grant undeclared permission")
	}
}

func TestPermissionManager_TemporaryPermissions(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"
	
	// Declare permissions
	permissions := []Permission{PermissionReadRepo}
	pm.DeclarePermissions(pluginID, permissions)

	// Grant temporary permission
	duration := 100 * time.Millisecond
	envelope := pm.GrantPermission(pluginID, PermissionReadRepo, true, duration)
	if envelope.Code != "" {
		t.Errorf("Failed to grant temporary permission: %v", envelope.Message)
	}

	// Should initially have permission
	if !pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Temporary permission should be available immediately")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should no longer have permission
	if pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Temporary permission should have expired")
	}
}

func TestPermissionManager_PatternMatching(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"
	
	// Declare wildcard permission
	permissions := []Permission{"secrets:jira*"}
	pm.DeclarePermissions(pluginID, permissions)

	// Grant wildcard permission
	envelope := pm.GrantPermission(pluginID, "secrets:jira*", false, 0)
	if envelope.Code != "" {
		t.Errorf("Failed to grant wildcard permission: %v", envelope.Message)
	}

	// Should match specific secrets
	if !pm.HasPermission(pluginID, "secrets:jira.token") {
		t.Error("Wildcard permission should match specific pattern")
	}

	if !pm.HasPermission(pluginID, "secrets:jira.oauth.refresh") {
		t.Error("Wildcard permission should match nested pattern")
	}

	// Should not match other services
	if pm.HasPermission(pluginID, "secrets:github.token") {
		t.Error("Wildcard permission should not match different service")
	}
}

func TestPermissionManager_RevokePermission(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"
	
	// Declare and grant permission
	permissions := []Permission{PermissionReadRepo}
	pm.DeclarePermissions(pluginID, permissions)
	
	envelope := pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
	if envelope.Code != "" {
		t.Errorf("Failed to grant permission: %v", envelope.Message)
	}

	// Verify permission is granted
	if !pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permission should be granted")
	}

	// Revoke permission
	pm.RevokePermission(pluginID, PermissionReadRepo)

	// Verify permission is revoked
	if pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permission should be revoked")
	}
}

func TestPermissionManager_PermissionCategories(t *testing.T) {
	pm := NewPermissionManager()

	tests := []struct {
		permission Permission
		category   string
	}{
		{PermissionReadRepo, "LOW_RISK"},
		{PermissionUI, "LOW_RISK"},
		{PermissionAttachments, "MEDIUM_RISK"},
		{PermissionIndexWrite, "MEDIUM_RISK"},
		{PermissionWriteRepo, "HIGH_RISK"},
		{PermissionNet, "HIGH_RISK"},
		{"secrets:jira.token", "HIGH_RISK"},
		{"secrets:github*", "HIGH_RISK"},
	}

	for _, tt := range tests {
		t.Run(string(tt.permission), func(t *testing.T) {
			category := pm.GetPermissionCategory(tt.permission)
			if category != tt.category {
				t.Errorf("Expected category %s, got %s", tt.category, category)
			}
		})
	}
}

func TestPermissionManager_PermissionDescriptions(t *testing.T) {
	pm := NewPermissionManager()

	tests := []struct {
		permission  Permission
		shouldContain string
	}{
		{PermissionReadRepo, "Read project data"},
		{PermissionWriteRepo, "Modify project data"},
		{PermissionAttachments, "Access file attachments"},
		{PermissionNet, "Access network resources"},
		{PermissionIndexWrite, "Write to search index"},
		{PermissionUI, "Contribute to user interface"},
		{"secrets:jira.token", "Access secrets for jira.token"},
		{"secrets:github*", "Access secrets for githubany"},
		{"unknown:permission", "Unknown permission"},
	}

	for _, tt := range tests {
		t.Run(string(tt.permission), func(t *testing.T) {
			description := pm.GetPermissionDescription(tt.permission)
			if description == "" {
				t.Error("Description should not be empty")
			}
			if tt.shouldContain != "" && !contains(description, tt.shouldContain) {
				t.Errorf("Description '%s' should contain '%s'", description, tt.shouldContain)
			}
		})
	}
}

func TestPermissionManager_CleanupExpiredGrants(t *testing.T) {
	pm := NewPermissionManager()
	pluginID := "com.test.plugin"
	
	// Declare permissions
	permissions := []Permission{PermissionReadRepo, PermissionWriteRepo}
	pm.DeclarePermissions(pluginID, permissions)

	// Grant one permanent and one temporary permission
	pm.GrantPermission(pluginID, PermissionReadRepo, false, 0) // permanent
	pm.GrantPermission(pluginID, PermissionWriteRepo, true, 50*time.Millisecond) // temporary

	// Both should be available initially
	if !pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permanent permission should be available")
	}
	if !pm.HasPermission(pluginID, PermissionWriteRepo) {
		t.Error("Temporary permission should be available initially")
	}

	// Wait for temporary permission to expire
	time.Sleep(100 * time.Millisecond)

	// Clean up expired grants
	pm.CleanupExpiredGrants()

	// Permanent should still be available
	if !pm.HasPermission(pluginID, PermissionReadRepo) {
		t.Error("Permanent permission should still be available")
	}

	// Temporary should be cleaned up
	if pm.HasPermission(pluginID, PermissionWriteRepo) {
		t.Error("Expired temporary permission should be cleaned up")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}