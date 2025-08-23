-- SQLite search index schema per ADR-005
-- This index is rebuildable and not authoritative (source of truth is JSON files)

-- Index metadata table
CREATE TABLE IF NOT EXISTS index_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial metadata
INSERT OR REPLACE INTO index_meta (key, value) VALUES 
    ('schema_version', '1'),
    ('created_at', datetime('now')),
    ('last_rebuild', datetime('now'));

-- Main nodes table with all searchable content
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    parent_id TEXT,
    path TEXT,  -- Full path from root for display
    depth INTEGER DEFAULT 0,
    child_count INTEGER DEFAULT 0,
    created_at DATETIME,
    updated_at DATETIME,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Properties table for key-value pairs
CREATE TABLE IF NOT EXISTS properties (
    node_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    type_hint TEXT,
    indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (node_id, key),
    FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

-- Full-text search virtual table for nodes
CREATE VIRTUAL TABLE IF NOT EXISTS nodes_fts USING fts5(
    id UNINDEXED,  -- Don't index ID for FTS, but include for results
    name,
    description,
    path,
    content='nodes',  -- Use external content table
    content_rowid='rowid'
);

-- Full-text search virtual table for properties
CREATE VIRTUAL TABLE IF NOT EXISTS properties_fts USING fts5(
    node_id UNINDEXED,
    key,
    value,
    content='properties',
    content_rowid='rowid'
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_nodes_parent ON nodes(parent_id);
CREATE INDEX IF NOT EXISTS idx_nodes_depth ON nodes(depth);
CREATE INDEX IF NOT EXISTS idx_nodes_updated ON nodes(updated_at);
CREATE INDEX IF NOT EXISTS idx_properties_node ON properties(node_id);
CREATE INDEX IF NOT EXISTS idx_properties_key ON properties(key);

-- Triggers to maintain FTS5 indexes
CREATE TRIGGER IF NOT EXISTS nodes_fts_insert AFTER INSERT ON nodes BEGIN
    INSERT INTO nodes_fts(rowid, id, name, description, path) 
    VALUES (new.rowid, new.id, new.name, new.description, new.path);
END;

CREATE TRIGGER IF NOT EXISTS nodes_fts_update AFTER UPDATE ON nodes BEGIN
    INSERT INTO nodes_fts(nodes_fts, rowid, id, name, description, path) 
    VALUES ('delete', old.rowid, old.id, old.name, old.description, old.path);
    INSERT INTO nodes_fts(rowid, id, name, description, path) 
    VALUES (new.rowid, new.id, new.name, new.description, new.path);
END;

CREATE TRIGGER IF NOT EXISTS nodes_fts_delete AFTER DELETE ON nodes BEGIN
    INSERT INTO nodes_fts(nodes_fts, rowid, id, name, description, path) 
    VALUES ('delete', old.rowid, old.id, old.name, old.description, old.path);
END;

CREATE TRIGGER IF NOT EXISTS properties_fts_insert AFTER INSERT ON properties BEGIN
    INSERT INTO properties_fts(rowid, node_id, key, value) 
    VALUES (new.rowid, new.node_id, new.key, new.value);
END;

CREATE TRIGGER IF NOT EXISTS properties_fts_update AFTER UPDATE ON properties BEGIN
    INSERT INTO properties_fts(properties_fts, rowid, node_id, key, value) 
    VALUES ('delete', old.rowid, old.node_id, old.key, old.value);
    INSERT INTO properties_fts(rowid, node_id, key, value) 
    VALUES (new.rowid, new.node_id, new.key, new.value);
END;

CREATE TRIGGER IF NOT EXISTS properties_fts_delete AFTER DELETE ON properties BEGIN
    INSERT INTO properties_fts(properties_fts, rowid, node_id, key, value) 
    VALUES ('delete', old.rowid, old.node_id, old.key, old.value);
END;
