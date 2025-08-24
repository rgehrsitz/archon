package plugins

import (
	"context"
	"fmt"

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
}

// NewHostService creates a new plugin host service
func NewHostService(
	logger logging.Logger,
	nodeStore *store.NodeStore,
	gitRepo git.Repository,
	indexManager *index.Manager,
	permissionManager *PermissionManager,
) *HostService {
	return &HostService{
		logger:            logger,
		nodeStore:         nodeStore,
		gitRepo:           gitRepo,
		indexManager:      indexManager,
		permissionManager: permissionManager,
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
		if envelope := h.applyMutation(ctx, mutation); envelope.Code != "" {
			return envelope
		}
	}

	return errors.Envelope{}
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
func (h *HostService) applyMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
	switch mutation.Type {
	case MutationCreate:
		return h.applyCreateMutation(ctx, mutation)
	case MutationUpdate:
		return h.applyUpdateMutation(ctx, mutation)
	case MutationDelete:
		return h.applyDeleteMutation(ctx, mutation)
	case MutationMove:
		return h.applyMoveMutation(ctx, mutation)
	case MutationReorder:
		return h.applyReorderMutation(ctx, mutation)
	default:
		return errors.New(errors.ErrValidationFailure, "Unknown mutation type: "+string(mutation.Type))
	}
}

// applyCreateMutation creates a new node
func (h *HostService) applyCreateMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
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
func (h *HostService) applyUpdateMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
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
func (h *HostService) applyDeleteMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
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
func (h *HostService) applyMoveMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
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
func (h *HostService) applyReorderMutation(ctx context.Context, mutation *Mutation) errors.Envelope {
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