package plugins

import (
	"time"

	"github.com/rgehrsitz/archon/internal/types"
)

// Permission represents a plugin permission type
type Permission string

const (
	PermissionReadRepo    Permission = "readRepo"
	PermissionWriteRepo   Permission = "writeRepo"
	PermissionAttachments Permission = "attachments"
	PermissionNet         Permission = "net"
	PermissionIndexWrite  Permission = "indexWrite"
	PermissionUI          Permission = "ui"
	// PermissionSecrets uses dynamic patterns like "secrets:jira*"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeImporter             PluginType = "Importer"
	PluginTypeExporter             PluginType = "Exporter"
	PluginTypeTransformer          PluginType = "Transformer"
	PluginTypeValidator            PluginType = "Validator"
	PluginTypePanel                PluginType = "Panel"
	PluginTypeProvider             PluginType = "Provider"
	PluginTypeAttachmentProcessor  PluginType = "AttachmentProcessor"
	PluginTypeConflictResolver     PluginType = "ConflictResolver"
	PluginTypeSearchIndexer        PluginType = "SearchIndexer"
	PluginTypeUIContrib            PluginType = "UIContrib"
)

// PluginManifest represents a plugin's manifest metadata
type PluginManifest struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Type          PluginType        `json:"type"`
	Description   string            `json:"description,omitempty"`
	Author        string            `json:"author,omitempty"`
	License       string            `json:"license,omitempty"`
	Permissions   []Permission      `json:"permissions"`
	EntryPoint    string            `json:"entryPoint"`
	ArchonVersion string            `json:"archonVersion,omitempty"`
	Integrity     string            `json:"integrity,omitempty"`
	Metadata      *PluginMetadata   `json:"metadata,omitempty"`
}

// PluginMetadata represents additional plugin metadata
type PluginMetadata struct {
	Category   string   `json:"category,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Website    string   `json:"website,omitempty"`
	Repository string   `json:"repository,omitempty"`
}

// PluginInstallation represents an installed plugin
type PluginInstallation struct {
	Manifest    PluginManifest `json:"manifest"`
	Path        string         `json:"path"`
	InstalledAt time.Time      `json:"installedAt"`
	Enabled     bool           `json:"enabled"`
	Source      string         `json:"source"` // "local", "url", "registry"
}

// PluginPermissionGrant represents a granted permission
type PluginPermissionGrant struct {
	PluginID   string      `json:"pluginId"`
	Permission Permission  `json:"permission"`
	Granted    bool        `json:"granted"`
	Temporary  bool        `json:"temporary"`
	ExpiresAt  *time.Time  `json:"expiresAt,omitempty"`
	GrantedAt  time.Time   `json:"grantedAt"`
}

// Mutation represents a plugin-requested repository change
type Mutation struct {
	Type     MutationType      `json:"type"`
	NodeID   string            `json:"nodeId,omitempty"`
	ParentID string            `json:"parentId,omitempty"`
	Data     *NodeData         `json:"data,omitempty"`
	Position int               `json:"position,omitempty"`
}

// MutationType represents the type of mutation
type MutationType string

const (
	MutationCreate  MutationType = "create"
	MutationUpdate  MutationType = "update"
	MutationDelete  MutationType = "delete"
	MutationMove    MutationType = "move"
	MutationReorder MutationType = "reorder"
)

// NodeData represents node data for mutations (matches TypeScript ArchonNode)
type NodeData struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Children    []string               `json:"children,omitempty"`
}

// PluginExecutionContext represents the context passed to plugin operations
type PluginExecutionContext struct {
	PluginID    string                    `json:"pluginId"`
	Permissions []Permission              `json:"permissions"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// PluginRequest represents a request from a plugin to the host
type PluginRequest struct {
	ID       string                 `json:"id"`
	Method   string                 `json:"method"`
	Args     []interface{}          `json:"args"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PluginResponse represents a response to a plugin request
type PluginResponse struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success  bool                   `json:"success"`
	Imported int                    `json:"imported"`
	Skipped  int                    `json:"skipped"`
	Errors   []string               `json:"errors"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// PluginConfig represents plugin configuration settings
type PluginConfig struct {
	PluginID string                 `json:"pluginId"`
	Settings map[string]interface{} `json:"settings"`
}

// ConvertToArchonNode converts NodeData to types.Node for internal use
func (nd *NodeData) ToArchonNode() *types.Node {
	properties := make(map[string]types.Property)
	for key, value := range nd.Properties {
		properties[key] = types.Property{
			Value: value,
		}
	}

	return &types.Node{
		ID:          nd.ID,
		Name:        nd.Name,
		Description: nd.Description,
		Properties:  properties,
		Children:    nd.Children,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// FromArchonNode converts types.Node to NodeData for plugin use
func NodeDataFromArchonNode(node *types.Node) *NodeData {
	properties := make(map[string]interface{})
	for key, prop := range node.Properties {
		properties[key] = prop.Value
	}

	return &NodeData{
		ID:          node.ID,
		Name:        node.Name,
		Description: node.Description,
		Properties:  properties,
		Children:    node.Children,
	}
}