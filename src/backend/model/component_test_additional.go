package model

import (
	"testing"
)

func TestNewComponentTree(t *testing.T) {
	// Create test components
	root := NewComponent("root", "Root", "system")
	child1 := NewComponent("child1", "Child 1", "device")
	child1.ParentID = "root"
	child2 := NewComponent("child2", "Child 2", "device")
	child2.ParentID = "root"
	grandchild := NewComponent("grandchild", "Grandchild", "component")
	grandchild.ParentID = "child1"
	
	tests := []struct {
		name       string
		components []*Component
		wantErr    bool
		rootCount  int
	}{
		{
			name:       "Valid tree",
			components: []*Component{root, child1, child2, grandchild},
			wantErr:    false,
			rootCount:  1,
		},
		{
			name:       "Duplicate IDs",
			components: []*Component{root, child1, child1}, // Duplicate child1
			wantErr:    true,
			rootCount:  0,
		},
		{
			name: "Invalid parent reference",
			components: []*Component{
				root,
				&Component{
					ID:       "orphan",
					Name:     "Orphan",
					Type:     "device",
					ParentID: "nonexistent",
				},
			},
			wantErr:   true,
			rootCount: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := NewComponentTree(tt.components)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewComponentTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(tree.RootIDs) != tt.rootCount {
					t.Errorf("Expected %d root components, got %d", tt.rootCount, len(tree.RootIDs))
				}
				
				if len(tree.Components) != len(tt.components) {
					t.Errorf("Expected %d total components, got %d", len(tt.components), len(tree.Components))
				}
			}
		})
	}
}
