package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type MergeService struct{}

func NewMergeService() *MergeService { return &MergeService{} }

func (s *MergeService) ThreeWay(ctx context.Context, base, ours, theirs string) (any, errors.Envelope) {
	return map[string]any{"conflicts": []any{}}, errors.Envelope{}
}
