package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/index/sqlite"
)

// SearchService provides Wails-bound search operations
type SearchService struct {
	projectService *ProjectService
}

// NewSearchService creates a new search service
func NewSearchService(projectService *ProjectService) *SearchService {
	return &SearchService{
		projectService: projectService,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results []sqlite.SearchResult `json:"results"`
	Count   int                   `json:"count"`
}

// PropertySearchResponse represents a property search response
type PropertySearchResponse struct {
	Results []sqlite.PropertySearchResult `json:"results"`
	Count   int                           `json:"count"`
}

// SearchNodes performs full-text search on node names, descriptions, and paths
func (s *SearchService) SearchNodes(ctx context.Context, req SearchRequest) (*SearchResponse, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}

	results, err := s.projectService.currentProject.IndexManager.SearchNodes(req.Query, req.Limit)
	if err != nil {
		return nil, errors.WrapError(errors.ErrSearchFailure, "Search failed", err)
	}

	return &SearchResponse{
		Results: results,
		Count:   len(results),
	}, errors.Envelope{}
}

// SearchProperties performs full-text search on node properties
func (s *SearchService) SearchProperties(ctx context.Context, req SearchRequest) (*PropertySearchResponse, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	if req.Limit <= 0 {
		req.Limit = 50
	}

	results, err := s.projectService.currentProject.IndexManager.SearchProperties(req.Query, req.Limit)
	if err != nil {
		return nil, errors.WrapError(errors.ErrSearchFailure, "Property search failed", err)
	}

	return &PropertySearchResponse{
		Results: results,
		Count:   len(results),
	}, errors.Envelope{}
}

// SearchByPath searches nodes by path prefix
func (s *SearchService) SearchByPath(ctx context.Context, pathPrefix string, limit int) (*SearchResponse, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	if limit <= 0 {
		limit = 50
	}

	results, err := s.projectService.currentProject.IndexManager.SearchByPath(pathPrefix, limit)
	if err != nil {
		return nil, errors.WrapError(errors.ErrSearchFailure, "Path search failed", err)
	}

	return &SearchResponse{
		Results: results,
		Count:   len(results),
	}, errors.Envelope{}
}

// GetNodesByDepth retrieves all nodes at a specific depth level
func (s *SearchService) GetNodesByDepth(ctx context.Context, depth, limit int) (*SearchResponse, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	if limit <= 0 {
		limit = 50
	}

	results, err := s.projectService.currentProject.IndexManager.GetNodesByDepth(depth, limit)
	if err != nil {
		return nil, errors.WrapError(errors.ErrSearchFailure, "Depth query failed", err)
	}

	return &SearchResponse{
		Results: results,
		Count:   len(results),
	}, errors.Envelope{}
}

// GetIndexHealth checks the health of the search index
func (s *SearchService) GetIndexHealth(ctx context.Context) (map[string]any, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	health := make(map[string]any)

	// Check if index is accessible
	if err := s.projectService.currentProject.IndexManager.Health(); err != nil {
		health["status"] = "unhealthy"
		health["error"] = err.Error()
	} else {
		health["status"] = "healthy"
	}

	// Get schema version
	if version, err := s.projectService.currentProject.IndexManager.GetSchemaVersion(); err == nil {
		health["schema_version"] = version
	}

	return health, errors.Envelope{}
}