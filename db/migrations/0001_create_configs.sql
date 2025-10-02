CREATE TABLE IF NOT EXISTS configs (
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    PRIMARY KEY (name, version)
);
CREATE INDEX IF NOT EXISTS idx_configs_name ON configs(name);
