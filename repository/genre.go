package repository

import (
	"database/sql"
	"errors"
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

func (r *GenreRepository) GetByID(id int) (*models.Genre, error) {
	row := r.db.QueryRow("SELECT id, name FROM Genre WHERE id = ?", id)
	var g models.Genre
	err := row.Scan(&g.ID, &g.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GenreRepository) GetAll(limit, offset int) ([]*models.Genre, int, error) {
	rows, err := r.db.Query("SELECT id, name FROM Genre ORDER BY name LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	genres := []*models.Genre{}
	for rows.Next() {
		var g models.Genre
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, 0, err
		}
		genres = append(genres, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM Genre").Scan(&total); err != nil {
		return nil, 0, err
	}
	return genres, total, nil
}
