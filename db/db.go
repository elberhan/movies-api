package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

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
