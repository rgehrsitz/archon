package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// NodeService provides Wails-bound node operations
type NodeService struct {
	projectService *ProjectService
	ctx            context.Context
}

// NewNodeService creates a new node service
func NewNodeService(projectService *ProjectService) *NodeService {
	return &NodeService{
		projectService: projectService,
		ctx:            context.Background(),
	}
}

// SetContext sets the context for the service (called by Wails during initialization)
func (s *NodeService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// CreateNode creates a new node under the specified parent
func (s *NodeService) CreateNode(req *types.CreateNodeRequest) (*types.Node, errors.Envelope) {
	if s.projectService != nil && s.projectService.readOnly {
		return nil, errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) GetNode(nodeID string) (*types.Node, errors.Envelope) {
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
func (s *NodeService) UpdateNode(req *types.UpdateNodeRequest) (*types.Node, errors.Envelope) {
	if s.projectService != nil && s.projectService.readOnly {
		return nil, errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) DeleteNode(nodeID string) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) MoveNode(req *types.MoveNodeRequest) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) ReorderChildren(req *types.ReorderChildrenRequest) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) ListChildren(nodeID string) []*types.Node {
	nodeStore, err := s.getNodeStore()
	if err.Code != "" {
		logging.GetLogger().Error().Str("nodeId", nodeID).Str("error", err.Message).Msg("Failed to get node store in ListChildren")
		return nil
	}
	
	logging.GetLogger().Debug().Str("nodeId", nodeID).Msg("Listing children")
	
	children, storeErr := nodeStore.ListChildren(nodeID)
	if storeErr != nil {
		logging.GetLogger().Error().Err(storeErr).Str("nodeId", nodeID).Msg("Failed to list children")
		return nil
	}
	
	logging.GetLogger().Debug().Str("nodeId", nodeID).Int("childCount", len(children)).Msg("Children retrieved successfully")
	return children
}

// GetNodePath returns the path from root to the specified node
func (s *NodeService) GetNodePath(nodeID string) ([]*types.Node, errors.Envelope) {
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
func (s *NodeService) GetRootNode() *types.Node {
	projectStore, currentPath := s.projectService.GetCurrentProject()
	if projectStore == nil {
		logging.GetLogger().Error().Msg("No project currently open in GetRootNode")
		return nil
	}
	
	project, err := projectStore.OpenProject()
	if err != nil {
		logging.GetLogger().Error().Err(err).Msg("Failed to get project info in GetRootNode")
		return nil
	}
	
	logging.GetLogger().Debug().Str("rootId", project.RootID).Msg("Getting root node")
	
	nodeStore := store.NewNodeStore(currentPath, s.projectService.currentProject.IndexManager)
	rootNode, storeErr := nodeStore.GetNode(project.RootID)
	if storeErr != nil {
		logging.GetLogger().Error().Err(storeErr).Str("rootId", project.RootID).Msg("Failed to get root node")
		return nil
	}
	
	logging.GetLogger().Debug().Str("nodeId", rootNode.ID).Str("nodeName", rootNode.Name).Msg("Root node retrieved successfully")
	return rootNode
}

// SetProperty sets a property on a node
func (s *NodeService) SetProperty(nodeID, key string, value any, typeHint string) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
func (s *NodeService) DeleteProperty(nodeID, key string) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
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
