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
    CREATE TABLE IF NOT EXISTS data_sources (
        data_source_id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        data_source_type INTEGER NOT NULL,
        data_source_path TEXT NOT NULL,
        when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
    `

	_, err := db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

