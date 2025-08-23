package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/rgehrsitz/archon/internal/types"
)

type IndexWriter struct {
	db *DB
}

func NewIndexWriter(db *DB) *IndexWriter {
	return &IndexWriter{db: db}
}

func (w *IndexWriter) IndexNode(node *types.Node, parentID string, depth int) error {
	tx, err := w.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	path := w.buildPathFromID(node.ID, node.Name)

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO nodes (id, name, description, parent_id, path, depth, child_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		node.ID, node.Name, node.Description, parentID, path, depth, len(node.Children), node.CreatedAt, node.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to index node: %w", err)
	}

	if err := w.indexProperties(tx, node.ID, node.Properties); err != nil {
		return fmt.Errorf("failed to index properties: %w", err)
	}

	return tx.Commit()
}

func (w *IndexWriter) RemoveNode(nodeID string) error {
	tx, err := w.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM nodes WHERE id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}

	_, err = tx.Exec("DELETE FROM properties WHERE node_id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("failed to remove properties: %w", err)
	}

	return tx.Commit()
}

func (w *IndexWriter) UpdateNodeChildCount(nodeID string, childCount int) error {
	_, err := w.db.conn.Exec("UPDATE nodes SET child_count = ? WHERE id = ?", childCount, nodeID)
	if err != nil {
		return fmt.Errorf("failed to update child count: %w", err)
	}
	return nil
}

type NodeWithContext struct {
	Node     *types.Node
	ParentID string
	Depth    int
}

func (w *IndexWriter) RebuildIndex(nodesWithContext []NodeWithContext) error {
	tx, err := w.db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM nodes"); err != nil {
		return fmt.Errorf("failed to clear nodes: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM properties"); err != nil {
		return fmt.Errorf("failed to clear properties: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO nodes (id, name, description, parent_id, path, depth, child_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare node statement: %w", err)
	}
	defer stmt.Close()

	for _, nodeCtx := range nodesWithContext {
		node := nodeCtx.Node
		path := w.buildPathFromID(node.ID, node.Name)
		_, err = stmt.Exec(node.ID, node.Name, node.Description, nodeCtx.ParentID, path, nodeCtx.Depth, len(node.Children), node.CreatedAt, node.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert node %s: %w", node.ID, err)
		}

		if err := w.indexProperties(tx, node.ID, node.Properties); err != nil {
			return fmt.Errorf("failed to index properties for node %s: %w", node.ID, err)
		}
	}

	_, err = tx.Exec("UPDATE index_meta SET value = datetime('now') WHERE key = 'last_rebuild'")
	if err != nil {
		return fmt.Errorf("failed to update rebuild timestamp: %w", err)
	}

	return tx.Commit()
}

func (w *IndexWriter) indexProperties(tx *sql.Tx, nodeID string, properties map[string]types.Property) error {
	_, err := tx.Exec("DELETE FROM properties WHERE node_id = ?", nodeID)
	if err != nil {
		return fmt.Errorf("failed to clear existing properties: %w", err)
	}

	if len(properties) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO properties (node_id, key, value, type_hint)
		VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare property statement: %w", err)
	}
	defer stmt.Close()

	for key, prop := range properties {
		_, err = stmt.Exec(nodeID, key, prop.Value, prop.TypeHint)
		if err != nil {
			return fmt.Errorf("failed to insert property %s: %w", key, err)
		}
	}

	return nil
}

func (w *IndexWriter) buildPathFromID(nodeID, nodeName string) string {
	var pathParts []string
	var err error

	currentID := nodeID
	for currentID != "" {
		var name, parentID string
		err = w.db.conn.QueryRow("SELECT name, parent_id FROM nodes WHERE id = ?", currentID).Scan(&name, &parentID)
		if err != nil {
			if currentID == nodeID {
				return nodeName
			}
			break
		}
		pathParts = append([]string{name}, pathParts...)
		currentID = parentID
	}

	if len(pathParts) == 0 {
		return nodeName
	}
	pathParts[len(pathParts)-1] = nodeName
	return strings.Join(pathParts, " > ")
}
