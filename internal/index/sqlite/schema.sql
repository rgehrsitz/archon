-- Placeholder schema per ADR-005
CREATE TABLE IF NOT EXISTS nodes (id TEXT PRIMARY KEY, name TEXT);
CREATE TABLE IF NOT EXISTS properties (node_id TEXT, key TEXT, value TEXT);
