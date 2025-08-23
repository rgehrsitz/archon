package api

import (
	"context"
	"github.com/rgehrsitz/archon/internal/errors"
)

type GitService struct{}

func NewGitService() *GitService { return &GitService{} }

func (s *GitService) Status(ctx context.Context) (string, errors.Envelope) {
	return "", errors.Envelope{}
}
