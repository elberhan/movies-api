package repository

import (
	"database/sql"
	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// holds the database handle
type GenreRepository struct {
	db *sql.DB
}

func NewGenreRepository(db *sql.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

func (r *GenreRepository) Create(name string) (*models.Genre, error) {
	result, err := r.db.Exec("INSERT INTO Genre (name) VALUES (?)", name)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Genre{ID: int(id), Name: name}, nil
}
