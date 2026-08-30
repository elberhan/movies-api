package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

//go:embed reset.sql
var reset string

//go:embed seed.sql
var seed string

// Open creates the SQLite database, enables foreign keys,
// and runs the schema to create all tables.
func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:movies.db?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("execute schema: %w", err)
	}

	return db, nil
}

// Seed clears the tables using reset.sql and fills them again with
// seed.sql. The reset and the insert share one transaction, so a
// failure leaves the database exactly as it was, unchanged.
func Seed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(reset); err != nil {
		return fmt.Errorf("reset seeded tables: %w", err)
	}

	if _, err := tx.Exec(seed); err != nil {
		return fmt.Errorf("insert seed data: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}
