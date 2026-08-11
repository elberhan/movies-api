package repository

import (
	"database/sql"
)

// holds the database handle
type GenreRepository struct {
	db *sql.DB
}
