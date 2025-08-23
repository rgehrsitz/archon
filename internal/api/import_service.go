package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type ImportService struct{
    projectService *ProjectService
}

func NewImportService(projectService *ProjectService) *ImportService { return &ImportService{projectService: projectService} }

func (s *ImportService) Run(ctx context.Context, pluginID string, bytes []byte, opts map[string]any) (any, errors.Envelope) {
    if s.projectService == nil || s.projectService.currentProject == nil {
        return nil, errors.New(errors.ErrNoProject, "No project is currently open")
    }
    if s.projectService.readOnly {
        return nil, errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
    }
    return map[string]any{"pluginId": pluginID, "staged": true}, errors.Envelope{}
}
