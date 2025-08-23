package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
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
    return errors.Envelope{}
}

func (s *IndexService) Search(ctx context.Context, q string) (any, errors.Envelope) {
	return []any{}, errors.Envelope{}
}
