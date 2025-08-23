package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type SnapshotService struct{
    projectService *ProjectService
}

func NewSnapshotService(projectService *ProjectService) *SnapshotService { return &SnapshotService{projectService: projectService} }

func (s *SnapshotService) Create(ctx context.Context, tag, message string, notes map[string]any) (any, errors.Envelope) {
    if s.projectService == nil || s.projectService.currentProject == nil {
        return nil, errors.New(errors.ErrNoProject, "No project is currently open")
    }
    if s.projectService.readOnly {
        return nil, errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
    }
    return map[string]any{"tag": tag, "created": true}, errors.Envelope{}
}

func (s *SnapshotService) List(ctx context.Context) ([]map[string]any, errors.Envelope) {
	return []map[string]any{}, errors.Envelope{}
}
