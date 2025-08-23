package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type ImportService struct{}

func NewImportService() *ImportService { return &ImportService{} }

func (s *ImportService) Run(ctx context.Context, pluginID string, bytes []byte, opts map[string]any) (any, errors.Envelope) {
	return map[string]any{"pluginId": pluginID, "staged": true}, errors.Envelope{}
}
