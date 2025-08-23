package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// NodeService provides Wails-bound node operations
type NodeService struct {
	projectService *ProjectService
}

// NewNodeService creates a new node service
func NewNodeService(projectService *ProjectService) *NodeService {
	return &NodeService{
		projectService: projectService,
	}
}

// CreateNode creates a new node under the specified parent
func (s *NodeService) CreateNode(ctx context.Context, req *types.CreateNodeRequest) (*types.Node, errors.Envelope) {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return nil, err
	}
	
	node, storeErr := nodeStore.CreateNode(req)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create node", storeErr)
	}
	
	return node, errors.Envelope{}
}

// GetNode retrieves a node by ID
func (s *NodeService) GetNode(ctx context.Context, nodeID string) (*types.Node, errors.Envelope) {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return nil, err
	}
	
	node, storeErr := nodeStore.GetNode(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get node", storeErr)
	}
	
	return node, errors.Envelope{}
}

// UpdateNode updates an existing node
func (s *NodeService) UpdateNode(ctx context.Context, req *types.UpdateNodeRequest) (*types.Node, errors.Envelope) {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return nil, err
	}
	
	node, storeErr := nodeStore.UpdateNode(req)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to update node", storeErr)
	}
	
	return node, errors.Envelope{}
}

// DeleteNode deletes a node and all its children
func (s *NodeService) DeleteNode(ctx context.Context, nodeID string) errors.Envelope {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return err
	}
	
	storeErr := nodeStore.DeleteNode(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to delete node", storeErr)
	}
	
	return errors.Envelope{}
}

// MoveNode moves a node to a new parent
func (s *NodeService) MoveNode(ctx context.Context, req *types.MoveNodeRequest) errors.Envelope {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return err
	}
	
	storeErr := nodeStore.MoveNode(req)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to move node", storeErr)
	}
	
	return errors.Envelope{}
}

// ReorderChildren reorders the children of a parent node
func (s *NodeService) ReorderChildren(ctx context.Context, req *types.ReorderChildrenRequest) errors.Envelope {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return err
	}
	
	storeErr := nodeStore.ReorderChildren(req)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to reorder children", storeErr)
	}
	
	return errors.Envelope{}
}

// ListChildren returns all direct children of a node
func (s *NodeService) ListChildren(ctx context.Context, nodeID string) ([]*types.Node, errors.Envelope) {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return nil, err
	}
	
	children, storeErr := nodeStore.ListChildren(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to list children", storeErr)
	}
	
	return children, errors.Envelope{}
}

// GetNodePath returns the path from root to the specified node
func (s *NodeService) GetNodePath(ctx context.Context, nodeID string) ([]*types.Node, errors.Envelope) {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return nil, err
	}
	
	path, storeErr := nodeStore.GetNodePath(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get node path", storeErr)
	}
	
	return path, errors.Envelope{}
}

// GetRootNode returns the root node of the current project
func (s *NodeService) GetRootNode(ctx context.Context) (*types.Node, errors.Envelope) {
	projectStore, currentPath := s.projectService.GetCurrentProject()
	if projectStore == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	project, err := projectStore.OpenProject()
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get project info", err)
	}
	
	nodeStore := store.NewNodeStore(currentPath, s.projectService.currentProject.IndexManager)
	rootNode, storeErr := nodeStore.GetNode(project.RootID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get root node", storeErr)
	}
	
	return rootNode, errors.Envelope{}
}

// SetProperty sets a property on a node
func (s *NodeService) SetProperty(ctx context.Context, nodeID, key string, value any, typeHint string) errors.Envelope {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return err
	}
	
	// Get current node
	node, storeErr := nodeStore.GetNode(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to get node", storeErr)
	}
	
	// Update property
	if node.Properties == nil {
		node.Properties = make(map[string]types.Property)
	}
	
	node.Properties[key] = types.Property{
		TypeHint: typeHint,
		Value:    value,
	}
	
	// Save updated node
	updateReq := &types.UpdateNodeRequest{
		ID:         nodeID,
		Properties: node.Properties,
	}
	
	_, storeErr = nodeStore.UpdateNode(updateReq)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to update node property", storeErr)
	}
	
	return errors.Envelope{}
}

// DeleteProperty removes a property from a node
func (s *NodeService) DeleteProperty(ctx context.Context, nodeID, key string) errors.Envelope {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		return err
	}
	
	// Get current node
	node, storeErr := nodeStore.GetNode(nodeID)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to get node", storeErr)
	}
	
	// Remove property
	if node.Properties != nil {
		delete(node.Properties, key)
	}
	
	// Save updated node
	updateReq := &types.UpdateNodeRequest{
		ID:         nodeID,
		Properties: node.Properties,
	}
	
	_, storeErr = nodeStore.UpdateNode(updateReq)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to delete node property", storeErr)
	}
	
	return errors.Envelope{}
}

// Helper method to get node store for current project
func (s *NodeService) getNodeStore() (*store.NodeStore, errors.Envelope) {
	_, currentPath := s.projectService.GetCurrentProject()
	if currentPath == "" {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	return store.NewNodeStore(currentPath, s.projectService.currentProject.IndexManager), errors.Envelope{}
}
