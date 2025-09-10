package store

import (
	"context"
	"slices"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/types"
)

// NodeStore handles node-level CRUD operations with validation
type NodeStore struct {
	basePath     string
	loader       *Loader
	indexManager *index.Manager
}

// reindexSubtree re-indexes a node and all of its descendants based on current on-disk state.
// It is safe to call when the indexManager is nil (no-op).
func (ns *NodeStore) reindexSubtree(rootID string) error {
    if ns.indexManager == nil {
        return nil
    }
    // Load root node and determine its parent/depth
    node, err := ns.loader.LoadNode(rootID)
    if err != nil {
        return err
    }
    parent, err := ns.findParent(rootID)
    if err != nil {
        return err
    }
    parentID := ""
    if parent != nil {
        parentID = parent.ID
    }
    depth := ns.calculateDepth(parentID)
    if err := ns.indexManager.IndexNode(node, parentID, depth); err != nil {
        return err
    }
    return ns.reindexChildren(node, depth)
}

// reindexChildren recursively re-indexes all descendants given the current node and its depth.
func (ns *NodeStore) reindexChildren(parentNode *types.Node, parentDepth int) error {
    if ns.indexManager == nil {
        return nil
    }
    for _, childID := range parentNode.Children {
        child, err := ns.loader.LoadNode(childID)
        if err != nil {
            return err
        }
        if err := ns.indexManager.IndexNode(child, parentNode.ID, parentDepth+1); err != nil {
            return err
        }
        if err := ns.reindexChildren(child, parentDepth+1); err != nil {
            return err
        }
    }
    return nil
}

// NewNodeStore creates a new node store
func NewNodeStore(basePath string, indexManager *index.Manager) *NodeStore {
	return &NodeStore{
		basePath:     basePath,
		loader:       NewLoader(basePath),
		indexManager: indexManager,
	}
}

// CreateNode creates a new node under the specified parent
func (ns *NodeStore) CreateNode(req *types.CreateNodeRequest) (*types.Node, error) {
	return ns.CreateNodeWithContext(context.Background(), req)
}

// CreateNodeWithContext creates a new node with logging context
func (ns *NodeStore) CreateNodeWithContext(ctx context.Context, req *types.CreateNodeRequest) (*types.Node, error) {
	logger := logging.WithOperation("create_node").
		WithContext(map[string]interface{}{
			"parent_id": req.ParentID,
			"node_name": req.Name,
		})
	
	logger.Info().Msg("Starting node creation")
	
	// Validate request
	if validationErrors := ValidateCreateNodeRequest(req); len(validationErrors) > 0 {
		err := errors.FromValidationErrors(validationErrors)
		logger.Error().Err(err).Msg("Node creation validation failed")
		return nil, err
	}
	
	// Load parent node to validate it exists and check sibling names
	parent, err := ns.loader.LoadNode(req.ParentID)
	if err != nil {
		logger.Error().Err(err).Str("parent_id", req.ParentID).Msg("Failed to load parent node")
		return nil, err
	}
	
	// Check sibling name uniqueness
	siblings, err := ns.loadSiblings(req.ParentID)
	if err != nil {
		return nil, err
	}
	
	// Create temporary node to check name conflicts
	newNode := &types.Node{
		Name: req.Name,
	}
	siblings = append(siblings, newNode)
	
	if validationErrors := ValidateSiblingNames(siblings); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// Generate new node ID and timestamps
	nodeID := id.NewV7()
	now := time.Now()
	
	// Create the node
	node := &types.Node{
		ID:          nodeID,
		Name:        req.Name,
		Description: req.Description,
		Properties:  req.Properties,
		Children:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	if node.Properties == nil {
		node.Properties = make(map[string]types.Property)
	}
	
	// Validate the new node
	if validationErrors := ValidateNode(node); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// Save the new node
	if err := ns.loader.SaveNode(node); err != nil {
		return nil, err
	}
	
	// Index the new node
	parentID := ""
	if req.ParentID != "" {
		parentID = req.ParentID
	}
	depth := ns.calculateDepth(parentID)
	if err := ns.indexManager.IndexNode(node, parentID, depth); err != nil {
		_ = ns.loader.DeleteNode(nodeID)
		return nil, err
	}
	
	// Update parent's children list
	parent.Children = append(parent.Children, nodeID)
	if err := ns.loader.SaveNode(parent); err != nil {
		// Rollback: delete the node we just created
		_ = ns.loader.DeleteNode(nodeID)
		_ = ns.indexManager.RemoveNode(nodeID)
		return nil, err
	}
	
	// Update parent's child count in index
	if err := ns.indexManager.UpdateNodeChildCount(req.ParentID, len(parent.Children)); err != nil {
		// Non-critical error - log but don't fail
		logger.Warn().Err(err).
			Str("parent_id", req.ParentID).
			Msg("Failed to update parent child count in index")
	}
	
	logger.Info().
		Str("node_id", node.ID).
		Str("parent_id", req.ParentID).
		Msg("Node created successfully")
	
	return node, nil
}

func (ns *NodeStore) calculateDepth(parentID string) int {
	// Root node depth
	if parentID == "" {
		return 0
	}
	
	// Walk up the parent chain to compute depth: depth = parentDepth + 1
	// Fallback conservatively to 1 if any error occurs while traversing.
	depth := 1
	currentID := parentID
	for currentID != "" {
		parent, err := ns.findParent(currentID)
		if err != nil {
			// On traversal error, return best-effort depth computed so far
			return depth
		}
		if parent == nil {
			// Reached root
			break
		}
		depth++
		currentID = parent.ID
	}
	return depth
}

// GetNode retrieves a node by ID
func (ns *NodeStore) GetNode(nodeID string) (*types.Node, error) {
	return ns.loader.LoadNode(nodeID)
}

// UpdateNode updates an existing node
func (ns *NodeStore) UpdateNode(req *types.UpdateNodeRequest) (*types.Node, error) {
	// Validate request
	if validationErrors := ValidateUpdateNodeRequest(req); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// Load existing node
	node, err := ns.loader.LoadNode(req.ID)
	if err != nil {
		return nil, err
	}
	
	originalName := node.Name
	
	// Update fields if provided
	if req.Name != nil {
		node.Name = *req.Name
	}
	if req.Description != nil {
		node.Description = *req.Description
	}
	if req.Properties != nil {
		node.Properties = req.Properties
	}
	
	// If name changed, check sibling name uniqueness
	if req.Name != nil && *req.Name != originalName {
		parent, err := ns.findParent(node.ID)
		if err != nil {
			return nil, err
		}
		
		if parent != nil {
			siblings, err := ns.loadSiblings(parent.ID)
			if err != nil {
				return nil, err
			}
			
			// Filter out the current node from siblings check
			filteredSiblings := make([]*types.Node, 0, len(siblings))
			for _, sibling := range siblings {
				if sibling.ID != node.ID {
					filteredSiblings = append(filteredSiblings, sibling)
				}
			}
			filteredSiblings = append(filteredSiblings, node)
			
			if validationErrors := ValidateSiblingNames(filteredSiblings); len(validationErrors) > 0 {
				return nil, errors.FromValidationErrors(validationErrors)
			}
		}
	}
	
	// Validate the updated node
	if validationErrors := ValidateNode(node); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// Save the updated node
	if err := ns.loader.SaveNode(node); err != nil {
		return nil, err
	}

    // Incremental index updates
    if ns.indexManager != nil {
        // Determine parent and depth for this node
        parent, _ := ns.findParent(node.ID)
        parentID := ""
        if parent != nil {
            parentID = parent.ID
        }
        depth := ns.calculateDepth(parentID)
        if err := ns.indexManager.IndexNode(node, parentID, depth); err != nil {
            return nil, err
        }
        // If the node was renamed, descendants' paths change; reindex subtree below this node
        if req.Name != nil && *req.Name != originalName {
            if err := ns.reindexChildren(node, depth); err != nil {
                return nil, err
            }
        }
    }

    return node, nil
}

// DeleteNode deletes a node and all its children
func (ns *NodeStore) DeleteNode(nodeID string) error {
	if !id.IsValid(nodeID) {
		return errors.New(errors.ErrInvalidUUID, "Invalid node ID format")
	}
	
	// Load the node to be deleted
	node, err := ns.loader.LoadNode(nodeID)
	if err != nil {
		return err
	}
	
	// Recursively delete all children first
	for _, childID := range node.Children {
		if err := ns.DeleteNode(childID); err != nil {
			return err
		}
	}
	
	// Remove this node from its parent's children list
	parent, err := ns.findParent(nodeID)
	if err != nil {
		return err
	}
	
	if parent != nil {
		parent.Children = slices.DeleteFunc(parent.Children, func(id string) bool {
			return id == nodeID
		})
		if err := ns.loader.SaveNode(parent); err != nil {
			return err
		}
	}
	
	// Delete the node file
	return ns.loader.DeleteNode(nodeID)
}

// MoveNode moves a node to a new parent
func (ns *NodeStore) MoveNode(req *types.MoveNodeRequest) error {
	// Validate request
	if validationErrors := ValidateMoveNodeRequest(req); len(validationErrors) > 0 {
		return errors.FromValidationErrors(validationErrors)
	}
	
	// Load the node to move
	node, err := ns.loader.LoadNode(req.NodeID)
	if err != nil {
		return err
	}
	
	// Load new parent
	newParent, err := ns.loader.LoadNode(req.NewParentID)
	if err != nil {
		return err
	}
	
	// Check for circular reference (can't move node under itself or its descendants)
	if err := ns.checkCircularReference(req.NodeID, req.NewParentID); err != nil {
		return err
	}
	
	// Load current parent
	currentParent, err := ns.findParent(req.NodeID)
	if err != nil {
		return err
	}
	
	// If moving to same parent, this is just a reorder operation
	if currentParent != nil && currentParent.ID == req.NewParentID {
		return ns.reorderChild(req.NewParentID, req.NodeID, req.Position)
	}
	
	// Check sibling name uniqueness in new parent
	newSiblings, err := ns.loadSiblings(req.NewParentID)
	if err != nil {
		return err
	}
	
	newSiblings = append(newSiblings, node)
	if validationErrors := ValidateSiblingNames(newSiblings); len(validationErrors) > 0 {
		return errors.FromValidationErrors(validationErrors)
	}
	
	// Remove from current parent
	if currentParent != nil {
		currentParent.Children = slices.DeleteFunc(currentParent.Children, func(id string) bool {
			return id == req.NodeID
		})
		if err := ns.loader.SaveNode(currentParent); err != nil {
			return err
		}
	}
	
	// Add to new parent at specified position
	if req.Position >= 0 && req.Position < len(newParent.Children) {
		newParent.Children = slices.Insert(newParent.Children, req.Position, req.NodeID)
	} else {
		newParent.Children = append(newParent.Children, req.NodeID)
	}
	
    // Save new parent
    if err := ns.loader.SaveNode(newParent); err != nil {
        return err
    }

    // Incremental index updates
    if ns.indexManager != nil {
        // Update child counts for both parents in the index
        if currentParent != nil {
            if err := ns.indexManager.UpdateNodeChildCount(currentParent.ID, len(currentParent.Children)); err != nil {
                return err
            }
        }
        if err := ns.indexManager.UpdateNodeChildCount(newParent.ID, len(newParent.Children)); err != nil {
            return err
        }

        // Reindex moved node and all descendants to refresh path/parent/depth
        if err := ns.reindexSubtree(req.NodeID); err != nil {
            return err
        }
    }

    return nil
}

// ReorderChildren reorders the children of a parent node
func (ns *NodeStore) ReorderChildren(req *types.ReorderChildrenRequest) error {
	// Validate request
	if validationErrors := ValidateReorderChildrenRequest(req); len(validationErrors) > 0 {
		return errors.FromValidationErrors(validationErrors)
	}
	
	// Load parent node
	parent, err := ns.loader.LoadNode(req.ParentID)
	if err != nil {
		return err
	}
	
	// Verify all ordered IDs are actually children of this parent
	currentChildSet := make(map[string]bool)
	for _, childID := range parent.Children {
		currentChildSet[childID] = true
	}
	
	orderedChildSet := make(map[string]bool)
	for _, childID := range req.OrderedChildIDs {
		orderedChildSet[childID] = true
		if !currentChildSet[childID] {
			return errors.New(errors.ErrInvalidInput, "Child ID not found in parent's children: "+childID)
		}
	}
	
	// Verify we have all children and no missing ones
	if len(orderedChildSet) != len(currentChildSet) {
		return errors.New(errors.ErrInvalidInput, "Ordered children list doesn't match parent's current children")
	}
	
	// Update parent's children order
	parent.Children = req.OrderedChildIDs
	
	if err := ns.loader.SaveNode(parent); err != nil {
        return err
    }
    
    // Incremental index update: reindex the parent to refresh updated_at/child_count
    if ns.indexManager != nil {
        grandparent, _ := ns.findParent(parent.ID)
        parentID := ""
        if grandparent != nil {
            parentID = grandparent.ID
        }
        depth := ns.calculateDepth(parentID)
        if err := ns.indexManager.IndexNode(parent, parentID, depth); err != nil {
            return err
        }
    }
    
    return nil
}

// ListChildren returns all direct children of a node
func (ns *NodeStore) ListChildren(nodeID string) ([]*types.Node, error) {
	parent, err := ns.loader.LoadNode(nodeID)
	if err != nil {
		return nil, err
	}
	
	children := make([]*types.Node, 0, len(parent.Children))
	for _, childID := range parent.Children {
		child, err := ns.loader.LoadNode(childID)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	
	return children, nil
}

// GetNodePath returns the path from root to the specified node
func (ns *NodeStore) GetNodePath(nodeID string) ([]*types.Node, error) {
	var path []*types.Node
	currentID := nodeID
	
	for currentID != "" {
		node, err := ns.loader.LoadNode(currentID)
		if err != nil {
			return nil, err
		}
		
		path = append([]*types.Node{node}, path...)
		
		// Find parent
		parent, err := ns.findParent(currentID)
		if err != nil {
			return nil, err
		}
		
		if parent == nil {
			break // Reached root
		}
		currentID = parent.ID
	}
	
	return path, nil
}

// Helper methods

// loadSiblings loads all sibling nodes for a given parent
func (ns *NodeStore) loadSiblings(parentID string) ([]*types.Node, error) {
	parent, err := ns.loader.LoadNode(parentID)
	if err != nil {
		return nil, err
	}
	
	siblings := make([]*types.Node, 0, len(parent.Children))
	for _, childID := range parent.Children {
		child, err := ns.loader.LoadNode(childID)
		if err != nil {
			return nil, err
		}
		siblings = append(siblings, child)
	}
	
	return siblings, nil
}

// findParent finds the parent node of a given node ID
func (ns *NodeStore) findParent(nodeID string) (*types.Node, error) {
	// Load all node files and search for the one containing this nodeID as a child
	allNodeIDs, err := ns.loader.ListNodeFiles()
	if err != nil {
		return nil, err
	}
	
	for _, candidateID := range allNodeIDs {
		candidate, err := ns.loader.LoadNode(candidateID)
		if err != nil {
			continue // Skip corrupted nodes
		}
		
		if slices.Contains(candidate.Children, nodeID) {
			return candidate, nil
		}
	}
	
	return nil, nil // No parent found (this is the root node)
}

// checkCircularReference ensures moving a node won't create a circular reference
func (ns *NodeStore) checkCircularReference(nodeID, newParentID string) error {
	// Walk up from newParentID to ensure nodeID is not in the ancestry
	currentID := newParentID
	
	for currentID != "" {
		if currentID == nodeID {
			return errors.New(errors.ErrCircularReference, "Cannot move node under itself or its descendants")
		}
		
		parent, err := ns.findParent(currentID)
		if err != nil {
			return err
		}
		
		if parent == nil {
			break // Reached root
		}
		currentID = parent.ID
	}
	
	return nil
}

// reorderChild reorders a single child within its current parent
func (ns *NodeStore) reorderChild(parentID, childID string, position int) error {
	parent, err := ns.loader.LoadNode(parentID)
	if err != nil {
		return err
	}
	
	// Find current position
	currentPos := slices.Index(parent.Children, childID)
	if currentPos == -1 {
		return errors.New(errors.ErrInvalidInput, "Child not found in parent")
	}
	
	// Remove from current position
	parent.Children = slices.Delete(parent.Children, currentPos, currentPos+1)
	
	// Insert at new position
	if position >= 0 && position <= len(parent.Children) {
		parent.Children = slices.Insert(parent.Children, position, childID)
	} else {
		parent.Children = append(parent.Children, childID)
	}
	
	return ns.loader.SaveNode(parent)
}