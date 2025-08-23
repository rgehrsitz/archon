package index

import (
	"fmt"

	"github.com/rgehrsitz/archon/internal/index/sqlite"
	"github.com/rgehrsitz/archon/internal/types"
)

type Manager struct {
	db     *sqlite.DB
	writer *sqlite.IndexWriter
	search *sqlite.SearchEngine
}

func NewManager(projectRoot string) (*Manager, error) {
	db, err := sqlite.Open(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open index database: %w", err)
	}

	return &Manager{
		db:     db,
		writer: sqlite.NewIndexWriter(db),
		search: sqlite.NewSearchEngine(db),
	}, nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}

func (m *Manager) IndexNode(node *types.Node, parentID string, depth int) error {
	return m.writer.IndexNode(node, parentID, depth)
}

func (m *Manager) RemoveNode(nodeID string) error {
	return m.writer.RemoveNode(nodeID)
}

func (m *Manager) UpdateNodeChildCount(nodeID string, childCount int) error {
	return m.writer.UpdateNodeChildCount(nodeID, childCount)
}

func (m *Manager) RebuildIndex(nodesWithContext []sqlite.NodeWithContext) error {
	return m.writer.RebuildIndex(nodesWithContext)
}

func (m *Manager) SearchNodes(query string, limit int) ([]sqlite.SearchResult, error) {
	return m.search.SearchNodes(query, limit)
}

func (m *Manager) SearchProperties(query string, limit int) ([]sqlite.PropertySearchResult, error) {
	return m.search.SearchProperties(query, limit)
}

func (m *Manager) SearchByPath(pathPrefix string, limit int) ([]sqlite.SearchResult, error) {
	return m.search.SearchByPath(pathPrefix, limit)
}

func (m *Manager) GetNodesByDepth(depth, limit int) ([]sqlite.SearchResult, error) {
	return m.search.GetNodesByDepth(depth, limit)
}

func (m *Manager) GetSchemaVersion() (string, error) {
	return m.db.GetSchemaVersion()
}

func (m *Manager) Health() error {
	return m.db.Ping()
}