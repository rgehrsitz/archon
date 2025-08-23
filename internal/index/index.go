package index

import (
	"fmt"
	"os"
	"strings"

	"github.com/rgehrsitz/archon/internal/index/sqlite"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/types"
)

type Manager struct {
	db       *sqlite.DB
	writer   *sqlite.IndexWriter
	search   *sqlite.SearchEngine
	disabled bool
}

func NewManager(projectRoot string) (*Manager, error) {
	if os.Getenv("ARCHON_DISABLE_INDEX") == "1" {
		// Provide a no-op manager for environments without FTS5 (e.g., CI or dev machines)
		return &Manager{disabled: true}, nil
	}
	db, err := sqlite.Open(projectRoot)
	if err != nil {
		// Auto-fallback: if the platform SQLite lacks FTS5, disable the index gracefully
		if strings.Contains(strings.ToLower(err.Error()), "no such module: fts5") {
			logging.Log().Warn().Msg("SQLite FTS5 not available; disabling search index for this run")
			return &Manager{disabled: true}, nil
		}
		return nil, fmt.Errorf("failed to open index database: %w", err)
	}

	return &Manager{
		db:     db,
		writer: sqlite.NewIndexWriter(db),
		search: sqlite.NewSearchEngine(db),
	}, nil
}

func (m *Manager) Close() error {
	if m.disabled || m.db == nil {
		return nil
	}
	return m.db.Close()
}

func (m *Manager) IndexNode(node *types.Node, parentID string, depth int) error {
	if m.disabled {
		return nil
	}
	return m.writer.IndexNode(node, parentID, depth)
}

func (m *Manager) RemoveNode(nodeID string) error {
	if m.disabled {
		return nil
	}
	return m.writer.RemoveNode(nodeID)
}

func (m *Manager) UpdateNodeChildCount(nodeID string, childCount int) error {
	if m.disabled {
		return nil
	}
	return m.writer.UpdateNodeChildCount(nodeID, childCount)
}

func (m *Manager) RebuildIndex(nodesWithContext []sqlite.NodeWithContext) error {
	if m.disabled {
		return nil
	}
	return m.writer.RebuildIndex(nodesWithContext)
}

func (m *Manager) SearchNodes(query string, limit int) ([]sqlite.SearchResult, error) {
	if m.disabled {
		return []sqlite.SearchResult{}, nil
	}
	return m.search.SearchNodes(query, limit)
}

func (m *Manager) SearchProperties(query string, limit int) ([]sqlite.PropertySearchResult, error) {
	if m.disabled {
		return []sqlite.PropertySearchResult{}, nil
	}
	return m.search.SearchProperties(query, limit)
}

func (m *Manager) SearchByPath(pathPrefix string, limit int) ([]sqlite.SearchResult, error) {
	if m.disabled {
		return []sqlite.SearchResult{}, nil
	}
	return m.search.SearchByPath(pathPrefix, limit)
}

func (m *Manager) GetNodesByDepth(depth, limit int) ([]sqlite.SearchResult, error) {
	if m.disabled {
		return []sqlite.SearchResult{}, nil
	}
	return m.search.GetNodesByDepth(depth, limit)
}

func (m *Manager) GetSchemaVersion() (string, error) {
	if m.disabled {
		return "0", nil
	}
	return m.db.GetSchemaVersion()
}

func (m *Manager) Health() error {
	if m.disabled {
		return nil
	}
	return m.db.Ping()
}
