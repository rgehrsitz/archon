package api

import (
	"context"
	"wailts/internal/errors"
)

type GitService struct{}

func NewGitService() *GitService { return &GitService{} }

func (s *GitService) Status(ctx context.Context) (string, errors.Envelope) {
	return "", errors.Envelope{}
}
