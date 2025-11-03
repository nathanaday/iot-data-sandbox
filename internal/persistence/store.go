package persistence

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles database operations
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance with initialized database
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func createTables(db *sql.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS projects (
        project_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS data_sources (
        data_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER,
        name TEXT NOT NULL,
        data_source_type INTEGER NOT NULL,
        data_source_path TEXT NOT NULL,
        row_count INTEGER NOT NULL DEFAULT 0,
        start_time TIMESTAMP,
        end_time TIMESTAMP,
        time_label TEXT NOT NULL DEFAULT 'time',
        value_label TEXT NOT NULL DEFAULT 'value',
        when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE SET NULL
    );

    CREATE TABLE IF NOT EXISTS data_layers (
        data_layer_id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER,
        data_source_id INTEGER,
        name TEXT NOT NULL,
        color TEXT NOT NULL DEFAULT '#3b82f6',
        z_index INTEGER NOT NULL DEFAULT 0,
        is_visible BOOLEAN NOT NULL DEFAULT 1,
        display_start_time TIMESTAMP,
        display_end_time TIMESTAMP,
        FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE,
        FOREIGN KEY (data_source_id) REFERENCES data_sources(data_source_id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS tools (
        tool_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        fx_name TEXT NOT NULL,
        timeout_s INTEGER NOT NULL DEFAULT 30,
        is_enabled BOOLEAN NOT NULL DEFAULT 1,
        when_last_call TIMESTAMP,
        num_calls INTEGER NOT NULL DEFAULT 0,
        max_calls INTEGER,
        num_call_reset INTEGER,
        UNIQUE(fx_name)
    );

    CREATE TABLE IF NOT EXISTS tool_auth_props (
        tool_id INTEGER PRIMARY KEY,
        hashed_api_key TEXT,
        hashed_username TEXT,
        hashed_password TEXT,
        FOREIGN KEY (tool_id) REFERENCES tools(tool_id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_tools_enabled ON tools(is_enabled);
    CREATE INDEX IF NOT EXISTS idx_data_sources_type ON data_sources(data_source_type);
    CREATE INDEX IF NOT EXISTS idx_data_sources_project ON data_sources(project_id);
    CREATE INDEX IF NOT EXISTS idx_data_layers_project ON data_layers(project_id);
    CREATE INDEX IF NOT EXISTS idx_data_layers_z_index ON data_layers(project_id, z_index);
    `

	_, err := db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

