package semantic

import (
	"encoding/json"
	"sort"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/types"
)

// Diff computes a semantic diff between two refs in the given repo path
func Diff(repoPath, from, to string) (*Result, errors.Envelope) {
	a, errA := loadSnapshot(repoPath, from)
	if errA.Code != "" {
		return nil, errA
	}
	b, errB := loadSnapshot(repoPath, to)
	if errB.Code != "" {
		return nil, errB
	}

	changes := make([]Change, 0, len(a.Nodes)+len(b.Nodes))

	// Index IDs
	idsA := make(map[string]struct{}, len(a.Nodes))
	for id := range a.Nodes {
		idsA[id] = struct{}{}
	}
	idsB := make(map[string]struct{}, len(b.Nodes))
	for id := range b.Nodes {
		idsB[id] = struct{}{}
	}

	// Added
	for id := range idsB {
		if _, ok := idsA[id]; !ok {
			changes = append(changes, Change{Type: ChangeNodeAdded, NodeID: id})
		}
	}
	// Removed
	for id := range idsA {
		if _, ok := idsB[id]; !ok {
			changes = append(changes, Change{Type: ChangeNodeRemoved, NodeID: id})
		}
	}

	// Shared IDs: inspect for rename/move/property/order
	for id := range idsA {
		if _, ok := idsB[id]; !ok {
			continue
		}
		na := a.Nodes[id]
		nb := b.Nodes[id]
		if na == nil || nb == nil {
			continue
		}
		// Rename
		if na.Name != nb.Name {
			changes = append(changes, Change{Type: ChangeNodeRenamed, NodeID: id, NameFrom: na.Name, NameTo: nb.Name})
		}
		// Move (parent change)
		pa := a.ParentByChild[id]
		pb := b.ParentByChild[id]
		if pa != pb {
			changes = append(changes, Change{Type: ChangeNodeMoved, NodeID: id, ParentFrom: pa, ParentTo: pb})
		}
		// Properties (description + properties map)
		propChanges := diffPropsAndDesc(na, nb)
		if len(propChanges) > 0 {
			changes = append(changes, Change{Type: ChangePropertyChanged, NodeID: id, ChangedProperties: propChanges})
		}
	}

	// Order changes: for any parent present in both where child IDs set equal but order differs
	// Build parent list from union of parent IDs
	parentIDs := make(map[string]struct{})
	for pid := range a.Nodes {
		// parent if it has children
		if len(a.Nodes[pid].Children) > 0 {
			parentIDs[pid] = struct{}{}
		}
	}
	for pid := range b.Nodes {
		if len(b.Nodes[pid].Children) > 0 {
			parentIDs[pid] = struct{}{}
		}
	}

	for pid := range parentIDs {
		pa, oka := a.Nodes[pid]
		pb, okb := b.Nodes[pid]
		if !oka || !okb {
			continue
		}
		// Report order change whenever the order slices differ
		if !sameOrder(pa.Children, pb.Children) {
			changes = append(changes, Change{Type: ChangeOrderChanged, ParentID: pid, OrderFrom: clone(pa.Children), OrderTo: clone(pb.Children)})
		}
	}

	// Ensure deterministic ordering within property deltas for each change
	for i := range changes {
		if len(changes[i].ChangedProperties) > 0 {
			sort.SliceStable(changes[i].ChangedProperties, func(a, b int) bool {
				if changes[i].ChangedProperties[a].Key == changes[i].ChangedProperties[b].Key {
					return changes[i].ChangedProperties[a].Kind < changes[i].ChangedProperties[b].Kind
				}
				return changes[i].ChangedProperties[a].Key < changes[i].ChangedProperties[b].Key
			})
		}
	}

	// Sort changes deterministically: by a fixed type order, then by key (nodeId or parentId), then by name fields
	sort.SliceStable(changes, func(i, j int) bool {
		ti, tj := rankType(changes[i].Type), rankType(changes[j].Type)
		if ti != tj {
			return ti < tj
		}
		// Key for tie-break: for OrderChanged use ParentID; otherwise NodeID
		ki := changes[i].NodeID
		if changes[i].Type == ChangeOrderChanged {
			ki = changes[i].ParentID
		}
		kj := changes[j].NodeID
		if changes[j].Type == ChangeOrderChanged {
			kj = changes[j].ParentID
		}
		if ki != kj {
			return ki < kj
		}
		// Final tie-breaker on NameFrom/NameTo to keep output stable across equal IDs
		if changes[i].NameFrom != changes[j].NameFrom {
			return changes[i].NameFrom < changes[j].NameFrom
		}
		return changes[i].NameTo < changes[j].NameTo
	})

	sum := summarize(changes)
	return &Result{From: from, To: to, Changes: changes, Summary: sum}, errors.Envelope{}
}

func diffPropsAndDesc(a, b *types.Node) []PropertyDelta {
	var deltas []PropertyDelta
	// Description
	if a.Description != b.Description {
		oldB, _ := jsonMarshal(a.Description)
		newB, _ := jsonMarshal(b.Description)
		deltas = append(deltas, PropertyDelta{Key: "description", Kind: "updated", Old: oldB, New: newB})
	}
	// Properties map
	keys := make(map[string]struct{})
	for k := range a.Properties {
		keys[k] = struct{}{}
	}
	for k := range b.Properties {
		keys[k] = struct{}{}
	}
	for k := range keys {
		va, oka := a.Properties[k]
		vb, okb := b.Properties[k]
		switch {
		case !oka && okb:
			newB, _ := jsonMarshal(vb)
			deltas = append(deltas, PropertyDelta{Key: k, Kind: "added", New: newB})
		case oka && !okb:
			oldB, _ := jsonMarshal(va)
			deltas = append(deltas, PropertyDelta{Key: k, Kind: "removed", Old: oldB})
		case oka && okb:
			if !equalProperty(va, vb) {
				oldB, _ := jsonMarshal(va)
				newB, _ := jsonMarshal(vb)
				deltas = append(deltas, PropertyDelta{Key: k, Kind: "updated", Old: oldB, New: newB})
			}
		}
	}
	return deltas
}

func equalProperty(a, b types.Property) bool {
	if a.TypeHint != b.TypeHint {
		return false
	}
	aj, _ := jsonMarshal(a)
	bj, _ := jsonMarshal(b)
	return string(aj) == string(bj)
}

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

func sameOrder(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func clone(s []string) []string { out := make([]string, len(s)); copy(out, s); return out }

func summarize(changes []Change) Summary {
	var s Summary
	s.Total = len(changes)
	for _, c := range changes {
		switch c.Type {
		case ChangeNodeAdded:
			s.Added++
		case ChangeNodeRemoved:
			s.Removed++
		case ChangeNodeRenamed:
			s.Renamed++
		case ChangeNodeMoved:
			s.Moved++
		case ChangePropertyChanged:
			s.PropertyChanged++
		case ChangeOrderChanged:
			s.OrderChanged++
		case ChangeAttachmentChanged:
			s.AttachmentChanged++
		}
	}
	return s
}

// rankType provides a deterministic ordering for change types in listings
func rankType(t ChangeType) int {
	switch t {
	case ChangeNodeAdded:
		return 0
	case ChangeNodeRemoved:
		return 1
	case ChangeNodeRenamed:
		return 2
	case ChangeNodeMoved:
		return 3
	case ChangePropertyChanged:
		return 4
	case ChangeOrderChanged:
		return 5
	case ChangeAttachmentChanged:
		return 6
	default:
		return 99
	}
}
