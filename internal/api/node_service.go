package api

import (
	"context"
	"wailts/internal/errors"
)

type NodeService struct{}

func NewNodeService() *NodeService { return &NodeService{} }

func (s *NodeService) Get(ctx context.Context, id string) (any, errors.Envelope) {
	return map[string]any{"id": id}, errors.Envelope{}
}

func (s *NodeService) Create(ctx context.Context, parentID, name string) (any, errors.Envelope) {
	return map[string]any{"parentId": parentID, "name": name}, errors.Envelope{}
}
