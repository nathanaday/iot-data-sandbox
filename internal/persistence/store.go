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

    CREATE TABLE IF NOT EXISTS dataframes (
        dataframe_id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        description TEXT,
        column_definitions TEXT,
        row_count INTEGER NOT NULL DEFAULT 0,
        start_time TIMESTAMP,
        end_time TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS data_layers (
        data_layer_id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER,
        dataframe_id INTEGER,
        name TEXT NOT NULL,
        color TEXT NOT NULL DEFAULT '#3b82f6',
        z_index INTEGER NOT NULL DEFAULT 0,
        is_visible BOOLEAN NOT NULL DEFAULT 1,
        FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE,
        FOREIGN KEY (dataframe_id) REFERENCES dataframes(dataframe_id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_dataframes_project ON dataframes(project_id);
    CREATE INDEX IF NOT EXISTS idx_data_layers_project ON data_layers(project_id);
    CREATE INDEX IF NOT EXISTS idx_data_layers_dataframe ON data_layers(dataframe_id);
    CREATE INDEX IF NOT EXISTS idx_data_layers_z_index ON data_layers(project_id, z_index);
    `

	_, err := db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
