package api

import (
	"context"
	"wailts/internal/errors"
)

type DiffService struct{}

func NewDiffService() *DiffService { return &DiffService{} }

func (s *DiffService) Diff(ctx context.Context, refA, refB string) (any, errors.Envelope) {
	return map[string]any{"refA": refA, "refB": refB, "changes": []any{}}, errors.Envelope{}
}
