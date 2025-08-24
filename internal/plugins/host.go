package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// HostService provides plugin host services with permission enforcement
type HostService struct {
	logger            logging.Logger
	nodeStore         *store.NodeStore
	gitRepo           git.Repository
	indexManager      *index.Manager
	permissionManager *PermissionManager
	secretsStore      SecretsStore
	proxyExecutor     ProxyExecutor
}

// NewHostService creates a new plugin host service
func NewHostService(
	logger logging.Logger,
	nodeStore *store.NodeStore,
	gitRepo git.Repository,
	indexManager *index.Manager,
	permissionManager *PermissionManager,
	secretsStore SecretsStore,
	proxyExecutor ProxyExecutor,
) *HostService {
	return &HostService{
		logger:            logger,
		nodeStore:         nodeStore,
		gitRepo:           gitRepo,
		indexManager:      indexManager,
		permissionManager: permissionManager,
		secretsStore:      secretsStore,
		proxyExecutor:     proxyExecutor,
	}
}

// GetNode retrieves a node by ID (requires readRepo permission)
func (h *HostService) GetNode(ctx context.Context, pluginID string, nodeID string) (*NodeData, errors.Envelope) {
	if !h.checkPermission(pluginID, PermissionReadRepo) {
		return nil, errors.New(errors.ErrUnauthorized, "Plugin lacks readRepo permission")
	}

	node, err := h.nodeStore.GetNode(nodeID)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get node", err)
	}

	if node == nil {
		return nil, errors.Envelope{} // Return nil node, no error
	}

	return NodeDataFromArchonNode(node), errors.Envelope{}
}

// ListChildren lists the children of a node (requires readRepo permission)
func (h *HostService) ListChildren(ctx context.Context, pluginID string, nodeID string) ([]string, errors.Envelope) {
	if !h.checkPermission(pluginID, PermissionReadRepo) {
		return nil, errors.New(errors.ErrUnauthorized, "Plugin lacks readRepo permission")
	}

	node, err := h.nodeStore.GetNode(nodeID)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get node", err)
	}

	if node == nil {
		return nil, errors.New(errors.ErrNotFound, "Node not found: "+nodeID)
	}

	return node.Children, errors.Envelope{}
}

// Query searches for nodes (requires readRepo permission)
func (h *HostService) Query(ctx context.Context, pluginID string, query string, limit int) ([]*NodeData, errors.Envelope) {
	if !h.checkPermission(pluginID, PermissionReadRepo) {
		return nil, errors.New(errors.ErrUnauthorized, "Plugin lacks readRepo permission")
	}

	if limit <= 0 {
		limit = 100 // Default limit
	}

	searchResults, err := h.indexManager.SearchNodes(query, limit)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Search failed", err)
	}

	var results []*NodeData
	for _, result := range searchResults {
		node, storeErr := h.nodeStore.GetNode(result.NodeID)
		if storeErr != nil {
			h.logger.Warn().Msg("Failed to get node from search result")
			continue
		}
		if node != nil {
			results = append(results, NodeDataFromArchonNode(node))
		}
	}

	return results, errors.Envelope{}
}

// ApplyMutations applies a batch of mutations (requires writeRepo permission)
func (h *HostService) ApplyMutations(ctx context.Context, pluginID string, mutations []*Mutation) errors.Envelope {
	if !h.checkPermission(pluginID, PermissionWriteRepo) {
		return errors.New(errors.ErrUnauthorized, "Plugin lacks writeRepo permission")
	}

	h.logger.Info().Msg("Applying plugin mutations")

	// Convert plugin mutations to store operations
	for _, mutation := range mutations {
		if envelope := h.applyMutation(mutation); envelope.Code != "" {
			return envelope
		}
	}

	return errors.Envelope{}
}

// SecretsGet returns a secret value by key (requires matching secrets:<key> permission)
func (h *HostService) SecretsGet(ctx context.Context, pluginID string, key string) (*SecretValue, errors.Envelope) {
    if key == "" {
        return nil, errors.New(errors.ErrValidationFailure, "Secret key is required")
    }

    // Enforce secrets permission with exact key match; wildcard grants are honored by PermissionManager
    reqPerm := Permission("secrets:" + key)
    if !h.checkPermission(pluginID, reqPerm) {
        return nil, errors.New(errors.ErrUnauthorized, "Plugin lacks secrets permission for "+key)
    }

    if h.secretsStore == nil {
        return nil, errors.New(errors.ErrNotImplemented, "Secrets backend not configured")
    }

    val, ok, err := h.secretsStore.Get(ctx, key)
    if err != nil {
        return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to read secret", err)
    }
    if !ok || val == nil {
        return nil, errors.New(errors.ErrNotFound, "Secret not found: "+key)
    }

    return val, errors.Envelope{}
}

// SecretsList lists secret keys by prefix (requires matching secrets:<prefix>* permission)
func (h *HostService) SecretsList(ctx context.Context, pluginID string, prefix string) ([]string, errors.Envelope) {
    if prefix == "" {
        return nil, errors.New(errors.ErrValidationFailure, "Prefix is required")
    }

    // Require permission corresponding to the prefix; wildcard grants are honored by PermissionManager
    reqPerm := Permission("secrets:" + prefix + "*")
    if !h.checkPermission(pluginID, reqPerm) {
        return nil, errors.New(errors.ErrUnauthorized, "Plugin lacks secrets permission for prefix "+prefix)
    }

    if h.secretsStore == nil {
        return nil, errors.New(errors.ErrNotImplemented, "Secrets backend not configured")
    }

    keys, err := h.secretsStore.List(ctx, prefix)
    if err != nil {
        return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to list secrets", err)
    }
    return keys, errors.Envelope{}
}

// NetRequest performs an outbound HTTP request via the host proxy (requires PermissionNet)
func (h *HostService) NetRequest(ctx context.Context, pluginID string, req ProxyRequest) (ProxyResponse, errors.Envelope) {
    if !h.checkPermission(pluginID, PermissionNet) {
        return ProxyResponse{}, errors.New(errors.ErrUnauthorized, "Plugin lacks net permission")
    }
    if req.URL == "" || req.Method == "" {
        return ProxyResponse{}, errors.New(errors.ErrValidationFailure, "Method and URL are required")
    }
    if h.proxyExecutor == nil {
        return ProxyResponse{}, errors.New(errors.ErrNotImplemented, "Proxy executor not configured")
    }

    // Apply timeout if provided
    if req.TimeoutMs > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
        defer cancel()
    }

    resp, err := h.proxyExecutor.Do(ctx, req)
    if err != nil {
        return ProxyResponse{}, errors.WrapError(errors.ErrRemoteFailure, "Proxy request failed", err)
    }
    return resp, errors.Envelope{}
}

// Commit creates a commit with the current changes (requires writeRepo permission)
func (h *HostService) Commit(ctx context.Context, pluginID string, message string) (string, errors.Envelope) {
	if !h.checkPermission(pluginID, PermissionWriteRepo) {
		return "", errors.New(errors.ErrUnauthorized, "Plugin lacks writeRepo permission")
	}

	if message == "" {
		message = fmt.Sprintf("Plugin %s commit", pluginID)
	}

	commit, envelope := h.gitRepo.Commit(ctx, message, nil)
	if envelope.Code != "" {
		return "", envelope
	}

	h.logger.Info().Msg("Plugin created commit")
	return commit.Hash, errors.Envelope{}
}

// Snapshot creates a snapshot (requires writeRepo permission)
func (h *HostService) Snapshot(ctx context.Context, pluginID string, message string) (string, errors.Envelope) {
	if !h.checkPermission(pluginID, PermissionWriteRepo) {
		return "", errors.New(errors.ErrUnauthorized, "Plugin lacks writeRepo permission")
	}

	if message == "" {
		message = fmt.Sprintf("Plugin %s snapshot", pluginID)
	}

	// First commit current changes
	commit, envelope := h.gitRepo.Commit(ctx, message, nil)
	if envelope.Code != "" {
		return "", envelope
	}

	h.logger.Info().Msg("Plugin created snapshot")
	return commit.Hash, errors.Envelope{}
}

// IndexPut adds content to the search index (requires indexWrite permission)
func (h *HostService) IndexPut(ctx context.Context, pluginID string, nodeID string, content string) errors.Envelope {
	if !h.checkPermission(pluginID, PermissionIndexWrite) {
		return errors.New(errors.ErrUnauthorized, "Plugin lacks indexWrite permission")
	}

	// Get the node first to ensure it exists
	node, err := h.nodeStore.GetNode(nodeID)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to get node", err)
	}

	if node == nil {
		return errors.New(errors.ErrNotFound, "Node not found: "+nodeID)
	}

	// Add to index
	if err := h.indexManager.IndexNode(node, "", 0); err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to index node", err)
	}

	return errors.Envelope{}
}

// applyMutation applies a single mutation
func (h *HostService) applyMutation(mutation *Mutation) errors.Envelope {
	switch mutation.Type {
	case MutationCreate:
		return h.applyCreateMutation(mutation)
	case MutationUpdate:
		return h.applyUpdateMutation(mutation)
	case MutationDelete:
		return h.applyDeleteMutation(mutation)
	case MutationMove:
		return h.applyMoveMutation(mutation)
	case MutationReorder:
		return h.applyReorderMutation(mutation)
	default:
		return errors.New(errors.ErrValidationFailure, "Unknown mutation type: "+string(mutation.Type))
	}
}

// applyCreateMutation creates a new node
func (h *HostService) applyCreateMutation(mutation *Mutation) errors.Envelope {
	if mutation.Data == nil {
		return errors.New(errors.ErrValidationFailure, "Create mutation missing data")
	}

	if mutation.ParentID == "" {
		return errors.New(errors.ErrValidationFailure, "Create mutation missing parent ID")
	}

	req := &types.CreateNodeRequest{
		ParentID:    mutation.ParentID,
		Name:        mutation.Data.Name,
		Description: mutation.Data.Description,
		Properties:  convertToArchonProperties(mutation.Data.Properties),
	}

	_, err := h.nodeStore.CreateNode(req)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to create node", err)
	}

	return errors.Envelope{}
}

// applyUpdateMutation updates an existing node
func (h *HostService) applyUpdateMutation(mutation *Mutation) errors.Envelope {
	if mutation.NodeID == "" {
		return errors.New(errors.ErrValidationFailure, "Update mutation missing node ID")
	}

	if mutation.Data == nil {
		return errors.New(errors.ErrValidationFailure, "Update mutation missing data")
	}

	req := &types.UpdateNodeRequest{
		ID:         mutation.NodeID,
		Properties: convertToArchonProperties(mutation.Data.Properties),
	}

	if mutation.Data.Name != "" {
		req.Name = &mutation.Data.Name
	}

	if mutation.Data.Description != "" {
		req.Description = &mutation.Data.Description
	}

	_, err := h.nodeStore.UpdateNode(req)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to update node", err)
	}

	return errors.Envelope{}
}

// applyDeleteMutation deletes a node
func (h *HostService) applyDeleteMutation(mutation *Mutation) errors.Envelope {
	if mutation.NodeID == "" {
		return errors.New(errors.ErrValidationFailure, "Delete mutation missing node ID")
	}

	err := h.nodeStore.DeleteNode(mutation.NodeID)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to delete node", err)
	}

	return errors.Envelope{}
}

// applyMoveMutation moves a node to a new parent
func (h *HostService) applyMoveMutation(mutation *Mutation) errors.Envelope {
	if mutation.NodeID == "" {
		return errors.New(errors.ErrValidationFailure, "Move mutation missing node ID")
	}

	if mutation.ParentID == "" {
		return errors.New(errors.ErrValidationFailure, "Move mutation missing parent ID")
	}

	req := &types.MoveNodeRequest{
		NodeID:      mutation.NodeID,
		NewParentID: mutation.ParentID,
		Position:    mutation.Position,
	}

	err := h.nodeStore.MoveNode(req)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to move node", err)
	}

	return errors.Envelope{}
}

// applyReorderMutation reorders children of a parent node
func (h *HostService) applyReorderMutation(mutation *Mutation) errors.Envelope {
	if mutation.ParentID == "" {
		return errors.New(errors.ErrValidationFailure, "Reorder mutation missing parent ID")
	}

	if mutation.Data == nil || len(mutation.Data.Children) == 0 {
		return errors.New(errors.ErrValidationFailure, "Reorder mutation missing children list")
	}

	req := &types.ReorderChildrenRequest{
		ParentID:        mutation.ParentID,
		OrderedChildIDs: mutation.Data.Children,
	}

	err := h.nodeStore.ReorderChildren(req)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to reorder children", err)
	}

	return errors.Envelope{}
}

// checkPermission verifies that a plugin has the required permission
func (h *HostService) checkPermission(pluginID string, permission Permission) bool {
	return h.permissionManager.HasPermission(pluginID, permission)
}

// convertToArchonProperties converts plugin properties to internal property format
func convertToArchonProperties(pluginProps map[string]interface{}) map[string]types.Property {
	properties := make(map[string]types.Property)
	for key, value := range pluginProps {
		properties[key] = types.Property{
			Value: value,
		}
	}
	return properties
}