package store

import (
	"testing"
	"time"
	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/types"
)

func TestValidateNode(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name           string
		node           *types.Node
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid node",
			node: &types.Node{
				ID:          "01234567-89ab-cdef-0123-456789abcdef",
				Name:        "Test Node",
				Description: "A test node",
				Properties:  make(map[string]types.Property),
				Children:    []string{"11111111-2222-3333-4444-555555555555"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty ID",
			node: &types.Node{
				ID:   "",
				Name: "Test Node",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid ID format",
			node: &types.Node{
				ID:   "invalid-uuid",
				Name: "Test Node",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Empty name",
			node: &types.Node{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: "",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
		{
			name: "Whitespace only name",
			node: &types.Node{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: "   ",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
		{
			name: "Invalid child ID",
			node: &types.Node{
				ID:       "01234567-89ab-cdef-0123-456789abcdef",
				Name:     "Test Node",
				Children: []string{"invalid-child-id"},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Duplicate child IDs",
			node: &types.Node{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: "Test Node",
				Children: []string{
					"11111111-2222-3333-4444-555555555555",
					"11111111-2222-3333-4444-555555555555",
				},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidInput},
		},
		{
			name: "Multiple errors",
			node: &types.Node{
				ID:   "invalid-id",
				Name: "",
				Children: []string{
					"invalid-child",
					"11111111-2222-3333-4444-555555555555",
					"11111111-2222-3333-4444-555555555555",
				},
			},
			expectedErrors: 4,
			expectedCodes:  []string{errors.ErrInvalidUUID, errors.ErrNameRequired, errors.ErrInvalidUUID, errors.ErrInvalidInput},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateNode(tt.node)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name           string
		project        *types.Project
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid project",
			project: &types.Project{
				RootID:        "01234567-89ab-cdef-0123-456789abcdef",
				SchemaVersion: 1,
				Settings:      map[string]any{"test": "value"},
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty root ID",
			project: &types.Project{
				RootID:        "",
				SchemaVersion: 1,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid root ID format",
			project: &types.Project{
				RootID:        "invalid-uuid",
				SchemaVersion: 1,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid schema version",
			project: &types.Project{
				RootID:        "01234567-89ab-cdef-0123-456789abcdef",
				SchemaVersion: 0,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidInput},
		},
		{
			name: "Negative schema version",
			project: &types.Project{
				RootID:        "01234567-89ab-cdef-0123-456789abcdef",
				SchemaVersion: -1,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidInput},
		},
		{
			name: "Multiple errors",
			project: &types.Project{
				RootID:        "invalid-uuid",
				SchemaVersion: 0,
			},
			expectedErrors: 2,
			expectedCodes:  []string{errors.ErrInvalidUUID, errors.ErrInvalidInput},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateProject(tt.project)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateSiblingNames(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name           string
		siblings       []*types.Node
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Unique names",
			siblings: []*types.Node{
				{Name: "Node A", CreatedAt: now, UpdatedAt: now},
				{Name: "Node B", CreatedAt: now, UpdatedAt: now},
				{Name: "Node C", CreatedAt: now, UpdatedAt: now},
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Case insensitive duplicate",
			siblings: []*types.Node{
				{Name: "Node A", CreatedAt: now, UpdatedAt: now},
				{Name: "node a", CreatedAt: now, UpdatedAt: now},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrDuplicateName},
		},
		{
			name: "Multiple duplicates",
			siblings: []*types.Node{
				{Name: "Node A", CreatedAt: now, UpdatedAt: now},
				{Name: "Node B", CreatedAt: now, UpdatedAt: now},
				{Name: "NODE A", CreatedAt: now, UpdatedAt: now},
				{Name: "node b", CreatedAt: now, UpdatedAt: now},
			},
			expectedErrors: 2,
			expectedCodes:  []string{errors.ErrDuplicateName, errors.ErrDuplicateName},
		},
		{
			name: "Whitespace handling",
			siblings: []*types.Node{
				{Name: "  Node A  ", CreatedAt: now, UpdatedAt: now},
				{Name: "Node A", CreatedAt: now, UpdatedAt: now},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrDuplicateName},
		},
		{
			name: "Empty names are skipped",
			siblings: []*types.Node{
				{Name: "", CreatedAt: now, UpdatedAt: now},
				{Name: "   ", CreatedAt: now, UpdatedAt: now},
				{Name: "Node A", CreatedAt: now, UpdatedAt: now},
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateSiblingNames(tt.siblings)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateCreateNodeRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.CreateNodeRequest
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid request",
			request: &types.CreateNodeRequest{
				ParentID:    "01234567-89ab-cdef-0123-456789abcdef",
				Name:        "Test Node",
				Description: "A test node",
				Properties:  make(map[string]types.Property),
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty parent ID",
			request: &types.CreateNodeRequest{
				ParentID: "",
				Name:     "Test Node",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidParent},
		},
		{
			name: "Invalid parent ID format",
			request: &types.CreateNodeRequest{
				ParentID: "invalid-uuid",
				Name:     "Test Node",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Empty name",
			request: &types.CreateNodeRequest{
				ParentID: "01234567-89ab-cdef-0123-456789abcdef",
				Name:     "",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
		{
			name: "Whitespace only name",
			request: &types.CreateNodeRequest{
				ParentID: "01234567-89ab-cdef-0123-456789abcdef",
				Name:     "   ",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateCreateNodeRequest(tt.request)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateUpdateNodeRequest(t *testing.T) {
	validName := "Valid Name"
	emptyName := ""
	whitespaceName := "   "
	
	tests := []struct {
		name           string
		request        *types.UpdateNodeRequest
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid request with name",
			request: &types.UpdateNodeRequest{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: &validName,
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Valid request without name",
			request: &types.UpdateNodeRequest{
				ID: "01234567-89ab-cdef-0123-456789abcdef",
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty ID",
			request: &types.UpdateNodeRequest{
				ID:   "",
				Name: &validName,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid ID format",
			request: &types.UpdateNodeRequest{
				ID:   "invalid-uuid",
				Name: &validName,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Empty name provided",
			request: &types.UpdateNodeRequest{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: &emptyName,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
		{
			name: "Whitespace only name provided",
			request: &types.UpdateNodeRequest{
				ID:   "01234567-89ab-cdef-0123-456789abcdef",
				Name: &whitespaceName,
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrNameRequired},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateUpdateNodeRequest(tt.request)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateMoveNodeRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.MoveNodeRequest
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid request",
			request: &types.MoveNodeRequest{
				NodeID:      "01234567-89ab-cdef-0123-456789abcdef",
				NewParentID: "11111111-2222-3333-4444-555555555555",
				Position:    0,
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty node ID",
			request: &types.MoveNodeRequest{
				NodeID:      "",
				NewParentID: "11111111-2222-3333-4444-555555555555",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid node ID",
			request: &types.MoveNodeRequest{
				NodeID:      "invalid-uuid",
				NewParentID: "11111111-2222-3333-4444-555555555555",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Empty new parent ID",
			request: &types.MoveNodeRequest{
				NodeID:      "01234567-89ab-cdef-0123-456789abcdef",
				NewParentID: "",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidParent},
		},
		{
			name: "Invalid new parent ID",
			request: &types.MoveNodeRequest{
				NodeID:      "01234567-89ab-cdef-0123-456789abcdef",
				NewParentID: "invalid-uuid",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Circular reference (same node)",
			request: &types.MoveNodeRequest{
				NodeID:      "01234567-89ab-cdef-0123-456789abcdef",
				NewParentID: "01234567-89ab-cdef-0123-456789abcdef",
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrCircularReference},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateMoveNodeRequest(tt.request)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}

func TestValidateReorderChildrenRequest(t *testing.T) {
	tests := []struct {
		name           string
		request        *types.ReorderChildrenRequest
		expectedErrors int
		expectedCodes  []string
	}{
		{
			name: "Valid request",
			request: &types.ReorderChildrenRequest{
				ParentID: "01234567-89ab-cdef-0123-456789abcdef",
				OrderedChildIDs: []string{
					"11111111-2222-3333-4444-555555555555",
					"22222222-3333-4444-5555-666666666666",
				},
			},
			expectedErrors: 0,
			expectedCodes:  []string{},
		},
		{
			name: "Empty parent ID",
			request: &types.ReorderChildrenRequest{
				ParentID:        "",
				OrderedChildIDs: []string{"11111111-2222-3333-4444-555555555555"},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidParent},
		},
		{
			name: "Invalid parent ID",
			request: &types.ReorderChildrenRequest{
				ParentID:        "invalid-uuid",
				OrderedChildIDs: []string{"11111111-2222-3333-4444-555555555555"},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Invalid child ID",
			request: &types.ReorderChildrenRequest{
				ParentID: "01234567-89ab-cdef-0123-456789abcdef",
				OrderedChildIDs: []string{
					"11111111-2222-3333-4444-555555555555",
					"invalid-uuid",
				},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidUUID},
		},
		{
			name: "Duplicate child IDs",
			request: &types.ReorderChildrenRequest{
				ParentID: "01234567-89ab-cdef-0123-456789abcdef",
				OrderedChildIDs: []string{
					"11111111-2222-3333-4444-555555555555",
					"11111111-2222-3333-4444-555555555555",
				},
			},
			expectedErrors: 1,
			expectedCodes:  []string{errors.ErrInvalidInput},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := ValidateReorderChildrenRequest(tt.request)
			
			if len(validationErrors) != tt.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d", tt.expectedErrors, len(validationErrors))
				for i, err := range validationErrors {
					t.Errorf("  Error %d: %s - %s", i, err.Code, err.Message)
				}
			}
			
			if len(tt.expectedCodes) > 0 {
				foundCodes := make(map[string]bool)
				for _, err := range validationErrors {
					foundCodes[err.Code] = true
				}
				
				for _, expectedCode := range tt.expectedCodes {
					if !foundCodes[expectedCode] {
						t.Errorf("Expected error code %s not found", expectedCode)
					}
				}
			}
		})
	}
}