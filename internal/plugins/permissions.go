package plugins

import (
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
)

// PermissionManager handles plugin permissions and grants
type PermissionManager struct {
	grants   map[string]map[Permission]*PluginPermissionGrant
	declared map[string][]Permission
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		grants:   make(map[string]map[Permission]*PluginPermissionGrant),
		declared: make(map[string][]Permission),
	}
}

// DeclarePermissions registers the permissions declared by a plugin
func (pm *PermissionManager) DeclarePermissions(pluginID string, permissions []Permission) {
	pm.declared[pluginID] = permissions
	
	// Initialize grants map for this plugin if not exists
	if pm.grants[pluginID] == nil {
		pm.grants[pluginID] = make(map[Permission]*PluginPermissionGrant)
	}
}

// HasPermission checks if a plugin has a specific permission
func (pm *PermissionManager) HasPermission(pluginID string, permission Permission) bool {
	pluginGrants, exists := pm.grants[pluginID]
	if !exists {
		return false
	}

	// Check for exact match first
	if grant, exists := pluginGrants[permission]; exists {
		return pm.isGrantValid(grant)
	}

	// Check for pattern matching (e.g., secrets:jira* matches secrets:jira.token)
	for grantedPerm, grant := range pluginGrants {
		if pm.permissionMatches(string(grantedPerm), string(permission)) && pm.isGrantValid(grant) {
			return true
		}
	}

	return false
}

// GrantPermission grants a permission to a plugin
func (pm *PermissionManager) GrantPermission(pluginID string, permission Permission, temporary bool, duration time.Duration) errors.Envelope {
	// Check if permission is declared
	if !pm.isPermissionDeclared(pluginID, permission) {
		return errors.New(errors.ErrValidationFailure, "Permission not declared by plugin: "+string(permission))
	}

	if pm.grants[pluginID] == nil {
		pm.grants[pluginID] = make(map[Permission]*PluginPermissionGrant)
	}

	grant := &PluginPermissionGrant{
		PluginID:   pluginID,
		Permission: permission,
		Granted:    true,
		Temporary:  temporary,
		GrantedAt:  time.Now(),
	}

	if temporary && duration > 0 {
		expiresAt := time.Now().Add(duration)
		grant.ExpiresAt = &expiresAt
	}

	pm.grants[pluginID][permission] = grant
	return errors.Envelope{}
}

// RevokePermission revokes a permission from a plugin
func (pm *PermissionManager) RevokePermission(pluginID string, permission Permission) {
	if pluginGrants, exists := pm.grants[pluginID]; exists {
		delete(pluginGrants, permission)
	}
}

// GetGrantedPermissions returns all currently granted permissions for a plugin
func (pm *PermissionManager) GetGrantedPermissions(pluginID string) []*PluginPermissionGrant {
	var grants []*PluginPermissionGrant
	
	pluginGrants, exists := pm.grants[pluginID]
	if !exists {
		return grants
	}

	for _, grant := range pluginGrants {
		if pm.isGrantValid(grant) {
			grants = append(grants, grant)
		}
	}

	return grants
}

// GetDeclaredPermissions returns all permissions declared by a plugin
func (pm *PermissionManager) GetDeclaredPermissions(pluginID string) []Permission {
	if permissions, exists := pm.declared[pluginID]; exists {
		return permissions
	}
	return []Permission{}
}

// ValidatePermissions checks that all requested permissions are valid
func (pm *PermissionManager) ValidatePermissions(permissions []Permission) []string {
	var errors []string
	
	for _, perm := range permissions {
		if !pm.isValidPermission(perm) {
			errors = append(errors, "Invalid permission: "+string(perm))
		}
	}
	
	return errors
}

// CleanupExpiredGrants removes expired temporary permissions
func (pm *PermissionManager) CleanupExpiredGrants() {
	now := time.Now()
	
	for _, pluginGrants := range pm.grants {
		for _, grant := range pluginGrants {
			if grant.Temporary && grant.ExpiresAt != nil && grant.ExpiresAt.Before(now) {
				delete(pluginGrants, grant.Permission)
			}
		}
	}
}

// isGrantValid checks if a permission grant is still valid
func (pm *PermissionManager) isGrantValid(grant *PluginPermissionGrant) bool {
	if !grant.Granted {
		return false
	}

	// Check if temporary grant has expired
	if grant.Temporary && grant.ExpiresAt != nil {
		return time.Now().Before(*grant.ExpiresAt)
	}

	return true
}

// isPermissionDeclared checks if a permission is declared by a plugin
func (pm *PermissionManager) isPermissionDeclared(pluginID string, permission Permission) bool {
	declared, exists := pm.declared[pluginID]
	if !exists {
		return false
	}

	for _, declaredPerm := range declared {
		if declaredPerm == permission || pm.permissionMatches(string(declaredPerm), string(permission)) {
			return true
		}
	}

	return false
}

// permissionMatches checks if a declared permission pattern matches a requested permission
func (pm *PermissionManager) permissionMatches(declared, requested string) bool {
	// Handle exact matches
	if declared == requested {
		return true
	}

	// Handle wildcard patterns (e.g., "secrets:jira*" matches "secrets:jira.token")
	if strings.Contains(declared, "*") {
		prefix := strings.TrimSuffix(declared, "*")
		return strings.HasPrefix(requested, prefix)
	}

	return false
}

// isValidPermission checks if a permission string is valid
func (pm *PermissionManager) isValidPermission(permission Permission) bool {
	perm := string(permission)
	
	// Standard permissions
	switch permission {
	case PermissionReadRepo, PermissionWriteRepo, PermissionAttachments, 
		 PermissionNet, PermissionIndexWrite, PermissionUI:
		return true
	}

	// Secrets permissions (must start with "secrets:")
	if strings.HasPrefix(perm, "secrets:") && len(perm) > 8 {
		return true
	}

	return false
}

// GetPermissionCategory returns the risk category of a permission
func (pm *PermissionManager) GetPermissionCategory(permission Permission) string {
	switch permission {
	case PermissionReadRepo, PermissionUI:
		return "LOW_RISK"
	case PermissionAttachments, PermissionIndexWrite:
		return "MEDIUM_RISK"
	case PermissionWriteRepo, PermissionNet:
		return "HIGH_RISK"
	}

	// Secrets are always high risk
	if strings.HasPrefix(string(permission), "secrets:") {
		return "HIGH_RISK"
	}

	return "UNKNOWN"
}

// GetPermissionDescription returns a human-readable description of a permission
func (pm *PermissionManager) GetPermissionDescription(permission Permission) string {
	switch permission {
	case PermissionReadRepo:
		return "Read project data (nodes, properties, structure)"
	case PermissionWriteRepo:
		return "Modify project data (create, update, delete nodes)"
	case PermissionAttachments:
		return "Access file attachments (read and write)"
	case PermissionNet:
		return "Access network resources (HTTP requests)"
	case PermissionIndexWrite:
		return "Write to search index"
	case PermissionUI:
		return "Contribute to user interface (commands, panels, dialogs)"
	}

	perm := string(permission)
	if strings.HasPrefix(perm, "secrets:") {
		scope := strings.TrimPrefix(perm, "secrets:")
		scope = strings.ReplaceAll(scope, "*", "any")
		return "Access secrets for " + scope + " services"
	}

	return "Unknown permission: " + perm
}