package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/index/sqlite"
	"github.com/rgehrsitz/archon/internal/store"
)

type IndexService struct{
    projectService *ProjectService
}

func NewIndexService(projectService *ProjectService) *IndexService { return &IndexService{projectService: projectService} }

func (s *IndexService) Rebuild(ctx context.Context) errors.Envelope {
    if s.projectService == nil || s.projectService.currentProject == nil {
        return errors.New(errors.ErrNoProject, "No project is currently open")
    }
    if s.projectService.readOnly {
        return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
    }

    // Build a complete snapshot of nodes with parent and depth context
    basePath := s.projectService.currentPath
    ldr := store.NewLoader(basePath)

    // List all node IDs
    allIDs, err := ldr.ListNodeFiles()
    if err != nil {
        if env, ok := err.(errors.Envelope); ok {
            return env
        }
        return errors.WrapError(errors.ErrStorageFailure, "Failed to list node files", err)
    }

    // Load all nodes and construct child->parent map in one pass
    nodes := make(map[string]*sqlite.NodeWithContext, len(allIDs))

    for _, id := range allIDs {
        n, loadErr := ldr.LoadNode(id)
        if loadErr != nil {
            if env, ok := loadErr.(errors.Envelope); ok {
                return env
            }
            return errors.WrapError(errors.ErrStorageFailure, "Failed to load node", loadErr)
        }
        nodes[id] = &sqlite.NodeWithContext{Node: n}
    }

    // Compute parent relationships from children arrays
    for _, nctx := range nodes {
        for _, childID := range nctx.Node.Children {
            if child, ok := nodes[childID]; ok {
                child.ParentID = nctx.Node.ID
            }
        }
    }

    // Compute depths by walking parent chain
    for _, nctx := range nodes {
        depth := 0
        current := nctx.ParentID
        for current != "" {
            parent, ok := nodes[current]
            if !ok {
                // Broken parent reference; stop here
                break
            }
            depth++
            current = parent.ParentID
        }
        nctx.Depth = depth
    }

    // Convert to slice in stable order (original list order is fine)
    list := make([]sqlite.NodeWithContext, 0, len(nodes))
    for _, id := range allIDs {
        if nctx, ok := nodes[id]; ok {
            list = append(list, *nctx)
        }
    }

    if err := s.projectService.currentProject.IndexManager.RebuildIndex(list); err != nil {
        return errors.WrapError(errors.ErrSearchFailure, "Failed to rebuild search index", err)
    }

    return errors.Envelope{}
}

func (s *IndexService) Search(ctx context.Context, q string) (any, errors.Envelope) {
	return []any{}, errors.Envelope{}
}
