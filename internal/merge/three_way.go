package merge

import (
	"encoding/json"
	"path/filepath"
	"time"

	semdiff "github.com/rgehrsitz/archon/internal/diff/semantic"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// ThreeWay computes a basic 3-way semantic merge scaffolding between refs.
// repoPath: path to the git repo; base|ours|theirs are git revs.
// It detects conflicting logical fields (name, parent, properties, order).
// It does NOT write changes; it only reports conflicts and non-conflicting changes.
func ThreeWay(repoPath, base, ours, theirs string) (*Resolution, error) {
	// Normalize repoPath (light touch)
	rp := filepath.Clean(repoPath)

	// Diffs vs base
	diffBO, env1 := semdiff.Diff(rp, base, ours)
	if env1.Code != "" {
		return nil, env1
	}
	diffBT, env2 := semdiff.Diff(rp, base, theirs)
	if env2.Code != "" {
		return nil, env2
	}

	res := &Resolution{Base: base, Ours: ours, Theirs: theirs}

	// Index changes by node+field
	oursMap := indexSemantic(diffBO.Changes)
	theirsMap := indexSemantic(diffBT.Changes)

	// Detect conflicts: same node+field changed on both sides with different values
	for key, oc := range oursMap {
		if tc, ok := theirsMap[key]; ok {
			if conflicts(oc, tc) {
				res.Conflicts = append(res.Conflicts, toConflict(key, oc, tc))
				delete(theirsMap, key)
				continue
			}
		}
		res.OursOnly = append(res.OursOnly, oc)
	}
	// Remaining theirs-only
	for _, tc := range theirsMap {
		res.TheirsOnly = append(res.TheirsOnly, tc)
	}
	return res, nil
}

// indexSemantic maps a change to a logical key (node+field)
func indexSemantic(changes []semdiff.Change) map[string]semdiff.Change {
	m := make(map[string]semdiff.Change)
	for _, c := range changes {
		switch c.Type {
		case semdiff.ChangeNodeRenamed:
			m[c.NodeID+"|name"] = c
		case semdiff.ChangeNodeMoved:
			m[c.NodeID+"|parent"] = c
		case semdiff.ChangeOrderChanged:
			// key by parent order
			m[c.ParentID+"|order"] = c
		case semdiff.ChangePropertyChanged:
			for _, p := range c.ChangedProperties {
				m[c.NodeID+"|property:"+p.Key] = c
			}
		case semdiff.ChangeNodeAdded, semdiff.ChangeNodeRemoved, semdiff.ChangeAttachmentChanged:
			// treat as whole-node change keyed by node
			m[c.NodeID+"|node"] = c
		}
	}
	return m
}

// conflicts determines whether two semantic changes on the same logical key are conflicting
func conflicts(a, b semdiff.Change) bool {
	if a.Type != b.Type {
		return true
	}
	switch a.Type {
	case semdiff.ChangeNodeRenamed:
		return a.NameTo != b.NameTo
	case semdiff.ChangeNodeMoved:
		return a.ParentTo != b.ParentTo
	case semdiff.ChangeOrderChanged:
		// any differing target order is a conflict for now
		if len(a.OrderTo) != len(b.OrderTo) {
			return true
		}
		for i := range a.OrderTo {
			if a.OrderTo[i] != b.OrderTo[i] {
				return true
			}
		}
		return false
	case semdiff.ChangePropertyChanged:
		// if any property delta for same key differs
		amap := mapProps(a)
		bmap := mapProps(b)
		if len(amap) != len(bmap) {
			return true
		}
		for k, va := range amap {
			vb, ok := bmap[k]
			if !ok || va != vb {
				return true
			}
		}
		return false
	default:
		// Node add/remove/attachment: treat as conflicting if both present
		return true
	}
}

func mapProps(c semdiff.Change) map[string]string {
	m := make(map[string]string)
	for _, p := range c.ChangedProperties {
		m[p.Key+"/"+p.Kind] = string(p.New)
	}
	return m
}

func toConflict(key string, a, b semdiff.Change) Conflict {
	return Conflict{Field: key, NodeID: a.NodeID, Ours: a, Theirs: b, Rule: "same field changed"}
}

// Apply applies non-conflicting changes from the resolution to the working tree
// This modifies the current working tree files (no git operations)
func (r *Resolution) Apply(repoPath string) error {
	rp := filepath.Clean(repoPath)
	loader := store.NewLoader(rp)
	
	// Apply changes in order: first theirs, then ours (ours takes precedence for same-priority changes)
	allChanges := append(r.TheirsOnly, r.OursOnly...)
	
	for _, change := range allChanges {
		if err := r.applyChange(loader, change); err != nil {
			return err
		}
		r.Applied = append(r.Applied, change)
	}
	
	return nil
}

// applyChange applies a single semantic change to the working tree
func (r *Resolution) applyChange(loader *store.Loader, change semdiff.Change) error {
	switch change.Type {
	case semdiff.ChangeNodeRenamed:
		return r.applyRename(loader, change)
	case semdiff.ChangeNodeMoved:
		return r.applyMove(loader, change)
	case semdiff.ChangePropertyChanged:
		return r.applyPropertyChange(loader, change)
	case semdiff.ChangeOrderChanged:
		return r.applyOrderChange(loader, change)
	case semdiff.ChangeNodeAdded:
		// TODO: Implement node addition - requires creating new node files
		// This is more complex as it requires reconstructing node data from semantic diff
		return nil
	case semdiff.ChangeNodeRemoved:
		// TODO: Implement node removal - requires careful cascade deletion
		return nil
	case semdiff.ChangeAttachmentChanged:
		// TODO: Implement attachment changes - reserved for future
		return nil
	default:
		return nil // Unknown change type, skip
	}
}

// applyRename changes the name of a node
func (r *Resolution) applyRename(loader *store.Loader, change semdiff.Change) error {
	node, err := loader.LoadNode(change.NodeID)
	if err != nil {
		return err
	}
	
	node.Name = change.NameTo
	node.UpdatedAt = time.Now()
	
	return loader.SaveNode(node)
}

// applyMove changes the parent of a node
func (r *Resolution) applyMove(loader *store.Loader, change semdiff.Change) error {
	// Remove from old parent
	if change.ParentFrom != "" {
		oldParent, err := loader.LoadNode(change.ParentFrom)
		if err != nil {
			return err
		}
		
		oldParent.Children = removeFromSlice(oldParent.Children, change.NodeID)
		oldParent.UpdatedAt = time.Now()
		if err := loader.SaveNode(oldParent); err != nil {
			return err
		}
	}
	
	// Add to new parent
	if change.ParentTo != "" {
		newParent, err := loader.LoadNode(change.ParentTo)
		if err != nil {
			return err
		}
		
		newParent.Children = append(newParent.Children, change.NodeID)
		newParent.UpdatedAt = time.Now()
		if err := loader.SaveNode(newParent); err != nil {
			return err
		}
	}
	
	return nil
}

// applyPropertyChange updates properties on a node
func (r *Resolution) applyPropertyChange(loader *store.Loader, change semdiff.Change) error {
	node, err := loader.LoadNode(change.NodeID)
	if err != nil {
		return err
	}
	
	if node.Properties == nil {
		node.Properties = make(map[string]types.Property)
	}
	
	// Apply each property delta
	for _, delta := range change.ChangedProperties {
		switch delta.Kind {
		case "added", "updated":
			if delta.Key == "description" {
				// Description is a special case - unmarshal directly as string
				var desc string
				if err := json.Unmarshal(delta.New, &desc); err != nil {
					return err
				}
				node.Description = desc
			} else {
				// Regular properties - unmarshal the Property structure
				var prop types.Property
				if err := json.Unmarshal(delta.New, &prop); err != nil {
					return err
				}
				node.Properties[delta.Key] = prop
			}
		case "removed":
			if delta.Key == "description" {
				node.Description = ""
			} else {
				delete(node.Properties, delta.Key)
			}
		}
	}
	
	node.UpdatedAt = time.Now()
	return loader.SaveNode(node)
}

// applyOrderChange reorders children of a parent node
func (r *Resolution) applyOrderChange(loader *store.Loader, change semdiff.Change) error {
	parent, err := loader.LoadNode(change.ParentID)
	if err != nil {
		return err
	}
	
	// Apply the new ordering
	parent.Children = change.OrderTo
	parent.UpdatedAt = time.Now()
	
	return loader.SaveNode(parent)
}

// removeFromSlice removes a value from a string slice
func removeFromSlice(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
