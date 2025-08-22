package api

import (
	"context"
	"wailts/internal/errors"
)

type SnapshotService struct{}

func NewSnapshotService() *SnapshotService { return &SnapshotService{} }

func (s *SnapshotService) Create(ctx context.Context, tag, message string, notes map[string]any) (any, errors.Envelope) {
	return map[string]any{"tag": tag, "created": true}, errors.Envelope{}
}

func (s *SnapshotService) List(ctx context.Context) ([]map[string]any, errors.Envelope) {
	return []map[string]any{}, errors.Envelope{}
}
