package store

import (
	"strings"
	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/types"
)

// ValidateNode performs comprehensive validation on a node
func ValidateNode(node *types.Node) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate ID
	if node.ID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "Node ID is required",
			Code:    errors.ErrInvalidUUID,
		})
	} else if !id.IsValid(node.ID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "Node ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate name
	if strings.TrimSpace(node.Name) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "name",
			Message: "Node name is required",
			Code:    errors.ErrNameRequired,
		})
	}
	
	// Validate children IDs
	for i, childID := range node.Children {
		if !id.IsValid(childID) {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   "children",
				Message: "Invalid child ID at position " + string(rune(i)),
				Code:    errors.ErrInvalidUUID,
			})
		}
	}
	
	// Check for duplicate child IDs
	childIDSet := make(map[string]bool)
	for _, childID := range node.Children {
		if childIDSet[childID] {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   "children",
				Message: "Duplicate child ID: " + childID,
				Code:    errors.ErrInvalidInput,
			})
		}
		childIDSet[childID] = true
	}
	
	return validationErrors
}

// ValidateProject performs validation on a project
func ValidateProject(project *types.Project) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate root ID
	if project.RootID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "rootId",
			Message: "Root ID is required",
			Code:    errors.ErrInvalidUUID,
		})
	} else if !id.IsValid(project.RootID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "rootId",
			Message: "Root ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate schema version
	if project.SchemaVersion <= 0 {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "schemaVersion",
			Message: "Schema version must be positive",
			Code:    errors.ErrInvalidInput,
		})
	}
	
	return validationErrors
}

// ValidateSiblingNames checks if names are unique among siblings (case-insensitive)
func ValidateSiblingNames(siblings []*types.Node) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	nameSet := make(map[string]string) // lowercase -> original name
	
	for _, node := range siblings {
		lowerName := strings.ToLower(strings.TrimSpace(node.Name))
		if lowerName == "" {
			continue // Skip empty names (handled by ValidateNode)
		}
		
		if existingName, exists := nameSet[lowerName]; exists {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   "name",
				Message: "Name '" + node.Name + "' conflicts with '" + existingName + "' (case-insensitive)",
				Code:    errors.ErrDuplicateName,
			})
		} else {
			nameSet[lowerName] = node.Name
		}
	}
	
	return validationErrors
}

// ValidateCreateNodeRequest validates a request to create a new node
func ValidateCreateNodeRequest(req *types.CreateNodeRequest) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate parent ID
	if req.ParentID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "parentId",
			Message: "Parent ID is required",
			Code:    errors.ErrInvalidParent,
		})
	} else if !id.IsValid(req.ParentID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "parentId",
			Message: "Parent ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "name",
			Message: "Node name is required",
			Code:    errors.ErrNameRequired,
		})
	}
	
	return validationErrors
}

// ValidateUpdateNodeRequest validates a request to update a node
func ValidateUpdateNodeRequest(req *types.UpdateNodeRequest) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate ID
	if req.ID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "Node ID is required",
			Code:    errors.ErrInvalidUUID,
		})
	} else if !id.IsValid(req.ID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "id",
			Message: "Node ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate name if provided
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "name",
			Message: "Node name cannot be empty",
			Code:    errors.ErrNameRequired,
		})
	}
	
	return validationErrors
}

// ValidateMoveNodeRequest validates a request to move a node
func ValidateMoveNodeRequest(req *types.MoveNodeRequest) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate node ID
	if req.NodeID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "nodeId",
			Message: "Node ID is required",
			Code:    errors.ErrInvalidUUID,
		})
	} else if !id.IsValid(req.NodeID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "nodeId",
			Message: "Node ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate new parent ID
	if req.NewParentID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "newParentId",
			Message: "New parent ID is required",
			Code:    errors.ErrInvalidParent,
		})
	} else if !id.IsValid(req.NewParentID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "newParentId",
			Message: "New parent ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Check for circular reference (can't move to self)
	if req.NodeID == req.NewParentID {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "newParentId",
			Message: "Cannot move node to itself",
			Code:    errors.ErrCircularReference,
		})
	}
	
	return validationErrors
}

// ValidateReorderChildrenRequest validates a request to reorder children
func ValidateReorderChildrenRequest(req *types.ReorderChildrenRequest) []errors.ValidationError {
	var validationErrors []errors.ValidationError
	
	// Validate parent ID
	if req.ParentID == "" {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "parentId",
			Message: "Parent ID is required",
			Code:    errors.ErrInvalidParent,
		})
	} else if !id.IsValid(req.ParentID) {
		validationErrors = append(validationErrors, errors.ValidationError{
			Field:   "parentId",
			Message: "Parent ID must be a valid UUID",
			Code:    errors.ErrInvalidUUID,
		})
	}
	
	// Validate child IDs
	for i, childID := range req.OrderedChildIDs {
		if !id.IsValid(childID) {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   "orderedChildIds",
				Message: "Invalid child ID at position " + string(rune(i)),
				Code:    errors.ErrInvalidUUID,
			})
		}
	}
	
	// Check for duplicates
	childIDSet := make(map[string]bool)
	for _, childID := range req.OrderedChildIDs {
		if childIDSet[childID] {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   "orderedChildIds",
				Message: "Duplicate child ID: " + childID,
				Code:    errors.ErrInvalidInput,
			})
		}
		childIDSet[childID] = true
	}
	
	return validationErrors
}
