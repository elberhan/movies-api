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

func (r *GenreRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM Genre WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *GenreRepository) MovieCount(id int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM movie_genres WHERE genre_id = ?", id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GenreRepository) GetMoviesByGenre(genreID int, limit, offset int) ([]*models.Movie, error) {
	rows, err := r.db.Query(`
		SELECT m.id, m.title, m.releaseYear, m.duration
		FROM Movie m
		JOIN movie_genres mg ON mg.movie_id = m.id
		WHERE mg.genre_id = ?
		LIMIT ? OFFSET ?`, genreID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []*models.Movie{}
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration); err != nil {
			return nil, err
		}
		movies = append(movies, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *GenreRepository) DeleteGenreMovieLinks(genreID int) error {
	_, err := r.db.Exec("DELETE FROM movie_genres WHERE genre_id = ?", genreID)
	return err
}
