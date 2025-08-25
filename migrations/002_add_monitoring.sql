-- Migration: Add monitoring tables to replace KV usage

-- Current status/state for monitored services
CREATE TABLE IF NOT EXISTS monitoring_state (
  key TEXT PRIMARY KEY,
  connected INTEGER NOT NULL,   -- 0/1
  session INTEGER,              -- 0/1
  status_json TEXT,             -- Full JSON payload from Z-API status
  last_changed DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- History of status changes (only persisted when state flips)
CREATE TABLE IF NOT EXISTS monitoring_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL,
  connected INTEGER NOT NULL,   -- 0/1
  session INTEGER,              -- 0/1
  status_json TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_monitoring_history_key_id
  ON monitoring_history(key, id DESC);

