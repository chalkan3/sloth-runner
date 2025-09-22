package ui

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "sloth_runner_history.db"

// InitDB initializes the SQLite database. It ensures the database file exists
// and the necessary tables are created.
func InitDB() (*sql.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, ".sloth-runner")
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("could not create .sloth-runner directory: %w", err)
	}

	db, err := sql.Open("sqlite3", filepath.Join(dbPath, dbFileName))
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("could not create tables: %w", err)
	}

	return db, nil
}

// createTables executes the SQL statements to create the database schema.
func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS pipeline_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_name TEXT NOT NULL,
		status TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME
	);

	CREATE TABLE IF NOT EXISTS task_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		task_name TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		message TEXT NOT NULL,
		FOREIGN KEY(run_id) REFERENCES pipeline_runs(id)
	);
	`
	_, err := db.Exec(schema)
	return err
}
