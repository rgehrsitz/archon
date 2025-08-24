package semantic

import (
	"encoding/json"
	"path/filepath"
	"strings"

	ggit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/types"
)

// snapshot holds parsed project and node state for a ref

type snapshot struct {
	Project       *types.Project
	Nodes         map[string]*types.Node
	ParentByChild map[string]string // childID -> parentID
}

// loadSnapshot reads project.json and nodes/*.json from a given ref (commit/tag/hash)
// using go-git to walk the tree without checking out.

func loadSnapshot(repoPath, ref string) (*snapshot, errors.Envelope) {
	// Open repository
	repo, err := ggit.PlainOpen(repoPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrInvalidPath, "Failed to open repository", err)
	}

	// Resolve ref to commit
	commit, err := resolveCommit(repo, ref)
	if err != nil {
		return nil, errors.WrapError(errors.ErrNotFound, "Failed to resolve ref", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get tree", err)
	}

	// Helper to read a blob file content by path
	readJSON := func(path string, v any) error {
		file, err := tree.File(path)
		if err != nil {
			return err
		}
		rc, err := file.Blob.Reader()
		if err != nil {
			return err
		}
		defer rc.Close()
		dec := json.NewDecoder(rc)
		return dec.Decode(v)
	}

	// Load project.json (optional but preferred)
	var proj types.Project
	_ = readJSON("project.json", &proj)

	// Load nodes/*.json
	nodes := make(map[string]*types.Node)

	// Walk tree to find nodes directory
	_ = tree.Files().ForEach(func(f *object.File) error {
		if f == nil {
			return nil
		}
		if !strings.HasPrefix(f.Name, "nodes/") {
			return nil
		}
		if filepath.Ext(f.Name) != ".json" {
			return nil
		}
		rc, err := f.Blob.Reader()
		if err != nil {
			return nil // tolerate
		}
		defer rc.Close()
		var n types.Node
		if err := json.NewDecoder(rc).Decode(&n); err != nil {
			return nil // tolerate corrupt/malformed
		}
		if n.ID != "" {
			// Copy to avoid referencing same memory
			nn := n
			nodes[n.ID] = &nn
		}
		return nil
	})

	// Build parent map by scanning children arrays
	parent := make(map[string]string)
	for pid, p := range nodes {
		for _, cid := range p.Children {
			if cid == "" {
				continue
			}
			parent[cid] = pid
		}
	}

	return &snapshot{
		Project:       &proj,
		Nodes:         nodes,
		ParentByChild: parent,
	}, errors.Envelope{}
}

// resolveCommit resolves ref to a commit using go-git

func resolveCommit(repo *ggit.Repository, ref string) (*object.Commit, error) {
	if ref == "" {
		h, err := repo.Head()
		if err != nil {
			return nil, err
		}
		return repo.CommitObject(h.Hash())
	}
	if hash, err := repo.ResolveRevision(plumbing.Revision(ref)); err == nil && hash != nil {
		if c, err := repo.CommitObject(*hash); err == nil {
			return c, nil
		}
		if tagObj, err := repo.TagObject(*hash); err == nil && tagObj != nil {
			if c, err := repo.CommitObject(tagObj.Target); err == nil {
				return c, nil
			}
		}
	}
	// last attempt, treat as hash
	if h := plumbing.NewHash(ref); !h.IsZero() {
		if c, err := repo.CommitObject(h); err == nil {
			return c, nil
		}
	}
	return nil, errors.New(errors.ErrNotFound, "unable to resolve commit")
}
