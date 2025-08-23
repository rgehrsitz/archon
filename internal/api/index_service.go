package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type IndexService struct{}

func NewIndexService() *IndexService { return &IndexService{} }

func (s *IndexService) Rebuild(ctx context.Context) errors.Envelope {
	return errors.Envelope{}
}

func (s *IndexService) Search(ctx context.Context, q string) (any, errors.Envelope) {
	return []any{}, errors.Envelope{}
}
