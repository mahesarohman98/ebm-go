package bookmanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

func newSqliteConnection(path string) (*sqlx.DB, error) {
	migrate := false
	// if path not exist, create path and do migate
	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, 0750); err != nil {
			return nil, fmt.Errorf("create ebm directory: %v", err)
		}
		migrate = true
	}

	db, err := sqlx.Connect("sqlite3", filepath.Join(path, "ebm.db"))
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	db.MustExec("PRAGMA foreign_keys = ON;")
	var foreignKeysEnabled int
	db.Get(&foreignKeysEnabled, "PRAGMA foreign_keys;") // Check if foreign keys are enabled
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

func runMigration(db *sqlx.DB, filepath string) error {
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
