package bookmanager

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

func newSqliteConnection(path string) (*sql.DB, error) {
	migrate := false
	// if path not exist, create path and do migate
	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, 0750); err != nil {
			return nil, fmt.Errorf("create ebm directory: %v", err)
		}
		migrate = true
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=rwc", filepath.Join(path, "ebm.db")))
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}

	var foreignKeysEnabled int
	row := db.QueryRow("PRAGMA foreign_keys;") // Check if foreign keys are enabled
	if err := row.Scan(&foreignKeysEnabled); err != nil {
		return nil, err
	}

	if foreignKeysEnabled != 1 {
		return nil, errors.New("foreign keys are disable")
	}

	if migrate {
		if err := runMigration(db, "sql/schema.sql"); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func runMigration(db *sql.DB, filepath string) error {
	// Read schema.sql file
	sqlBytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to apply migration: %w", err)
	}

	fmt.Println("Migration applied successfully.")
	return nil
}
