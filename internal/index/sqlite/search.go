package sqlite

import (
	"fmt"
	"strings"
)

type SearchResult struct {
	NodeID      string  `json:"nodeId"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Path        string  `json:"path"`
	Depth       int     `json:"depth"`
	ChildCount  int     `json:"childCount"`
	Rank        float64 `json:"rank"`
	Snippet     string  `json:"snippet,omitempty"`
}

type PropertySearchResult struct {
	NodeID string  `json:"nodeId"`
	Key    string  `json:"key"`
	Value  string  `json:"value"`
	Rank   float64 `json:"rank"`
}

type SearchEngine struct {
	db *DB
}

func NewSearchEngine(db *DB) *SearchEngine {
	return &SearchEngine{db: db}
}

func (s *SearchEngine) SearchNodes(query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	ftsQuery := s.buildFTSQuery(query)
	
	rows, err := s.db.conn.Query(`
		SELECT n.id, n.name, n.description, n.path, n.depth, n.child_count,
		       fts.rank, snippet(nodes_fts, 1, '<mark>', '</mark>', '...', 32) as snippet
		FROM nodes_fts fts
		JOIN nodes n ON fts.id = n.id
		WHERE nodes_fts MATCH ?
		ORDER BY fts.rank
		LIMIT ?`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var description, snippet *string
		
		err := rows.Scan(
			&result.NodeID, &result.Name, &description, &result.Path,
			&result.Depth, &result.ChildCount, &result.Rank, &snippet)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if description != nil {
			result.Description = *description
		}
		if snippet != nil {
			result.Snippet = *snippet
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search results: %w", err)
	}

	return results, nil
}

func (s *SearchEngine) SearchProperties(query string, limit int) ([]PropertySearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	ftsQuery := s.buildFTSQuery(query)
	
	rows, err := s.db.conn.Query(`
		SELECT fts.node_id, fts.key, fts.value, fts.rank
		FROM properties_fts fts
		WHERE properties_fts MATCH ?
		ORDER BY fts.rank
		LIMIT ?`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute property search query: %w", err)
	}
	defer rows.Close()

	var results []PropertySearchResult
	for rows.Next() {
		var result PropertySearchResult
		var value *string
		
		err := rows.Scan(&result.NodeID, &result.Key, &value, &result.Rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan property search result: %w", err)
		}

		if value != nil {
			result.Value = *value
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating property search results: %w", err)
	}

	return results, nil
}

func (s *SearchEngine) SearchByPath(pathPrefix string, limit int) ([]SearchResult, error) {
	if pathPrefix == "" {
		return nil, fmt.Errorf("path prefix cannot be empty")
	}

	rows, err := s.db.conn.Query(`
		SELECT id, name, description, path, depth, child_count, 1.0 as rank
		FROM nodes
		WHERE path LIKE ? || '%'
		ORDER BY depth, name
		LIMIT ?`, pathPrefix, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute path search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var description *string
		
		err := rows.Scan(
			&result.NodeID, &result.Name, &description, &result.Path,
			&result.Depth, &result.ChildCount, &result.Rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan path search result: %w", err)
		}

		if description != nil {
			result.Description = *description
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating path search results: %w", err)
	}

	return results, nil
}

func (s *SearchEngine) GetNodesByDepth(depth, limit int) ([]SearchResult, error) {
	rows, err := s.db.conn.Query(`
		SELECT id, name, description, path, depth, child_count, 1.0 as rank
		FROM nodes
		WHERE depth = ?
		ORDER BY name
		LIMIT ?`, depth, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute depth query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var description *string
		
		err := rows.Scan(
			&result.NodeID, &result.Name, &description, &result.Path,
			&result.Depth, &result.ChildCount, &result.Rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan depth result: %w", err)
		}

		if description != nil {
			result.Description = *description
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating depth results: %w", err)
	}

	return results, nil
}

func (s *SearchEngine) buildFTSQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	// Handle quoted phrases first
	var ftsTerms []string
	inQuotes := false
	currentPhrase := ""
	
	words := strings.Fields(query)
	for _, word := range words {
		if strings.HasPrefix(word, `"`) && strings.HasSuffix(word, `"`) && len(word) > 1 {
			// Complete quoted word
			ftsTerms = append(ftsTerms, word)
		} else if strings.HasPrefix(word, `"`) {
			// Start of quoted phrase
			inQuotes = true
			currentPhrase = word
		} else if strings.HasSuffix(word, `"`) && inQuotes {
			// End of quoted phrase
			currentPhrase += " " + word
			ftsTerms = append(ftsTerms, currentPhrase)
			inQuotes = false
			currentPhrase = ""
		} else if inQuotes {
			// Middle of quoted phrase
			currentPhrase += " " + word
		} else {
			// Regular term - add wildcard if not already ending with *
			if strings.HasSuffix(word, "*") {
				ftsTerms = append(ftsTerms, word)
			} else {
				ftsTerms = append(ftsTerms, word+"*")
			}
		}
	}

	return strings.Join(ftsTerms, " ")
}