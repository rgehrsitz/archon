package merge

import (
	"path/filepath"

	semdiff "github.com/rgehrsitz/archon/internal/diff/semantic"
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
