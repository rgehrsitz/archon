package api

import (
	"context"
	"wailts/internal/errors"
)

type ProjectService struct{}

func NewProjectService() *ProjectService { return &ProjectService{} }

func (s *ProjectService) New(ctx context.Context, path string, settings map[string]any) (any, errors.Envelope) {
	// TODO: init repo, LFS, write project.json
	return map[string]any{"path": path, "created": true}, errors.Envelope{}
}

func (s *ProjectService) Open(ctx context.Context, path string) (any, errors.Envelope) {
	// TODO: open + migrate if needed
	return map[string]any{"path": path, "open": true}, errors.Envelope{}
}

func (s *ProjectService) Close(ctx context.Context) errors.Envelope {
	return errors.Envelope{}
}
