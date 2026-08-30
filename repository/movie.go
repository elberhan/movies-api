package repository

import (
	"database/sql"
	"errors"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// holds the database handle
type MovieRepository struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

// Create creates a new record in the Movie table, returning *models.Movie (pointer to the
// [models.Movie] object) and nil, or nil pointer and error.
func (r *MovieRepository) Create(title string, releaseYear int, duration int) (*models.Movie, error) {
	result, err := r.db.Exec(
		`INSERT INTO Movie (title, releaseYear, duration) VALUES (?, ?, ?)`,
		title, releaseYear, duration)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Movie{
		ID:          int(id),
		Title:       title,
		ReleaseYear: releaseYear,
		Duration:    duration,
	}, nil
}

// GetByID fetches a specific movie record from the Movie table by its ID.
func (r *MovieRepository) GetByID(id int) (*models.Movie, error) {
	movieRow := r.db.QueryRow(
		`SELECT id, title, releaseYear, duration FROM Movie WHERE id = ?`, id)

	var m models.Movie

	err := movieRow.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}

	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetAll fetches one page of movies: limit rows, starting after offset rows.
// totalRows is the total row count in the Movie table.
func (r *MovieRepository) GetAll(limit, offset int) (page []*models.Movie, totalRows int, err error) {
	rows, err := r.db.Query(
		`SELECT id, title, releaseYear, duration FROM Movie 
        ORDER BY id 
        LIMIT ? OFFSET ?`, limit, offset)

	for rows.Next() {
		var m models.Movie
		err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration)
		if err != nil {
			return nil, 0, err
		}
		page = append(page, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.db.QueryRow("SELECT COUNT(*) FROM Movie").Scan(&totalRows); err != nil {
		return nil, 0, err
	}

	return page, totalRows, nil
}

// GetByGenre fetches one page of the movies that belongs to the given genre.
// totalRows is the total movie count of specified genre in the Movie table.
func (r *MovieRepository) GetByGenre(genreID, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	rows, err := r.db.Query(
		`SELECT m.id, m.title, m.releaseYear, m.duration
		FROM Movie AS m
		JOIN movie_genres AS mg ON mg.movie_id = m.id
		WHERE mg.genre_id = ?
		ORDER BY m.id
		LIMIT ? OFFSET ?`, genreID, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration); err != nil {
			return nil, 0, err
		}
		page = append(page, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// get the number of total movies of specified genre
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM movie_genres WHERE genre_id = ?", genreID).Scan(&totalMovies); err != nil {
		return nil, 0, err
	}

	return page, totalMovies, nil
}

// GetByYear fetches one page of movies released in the given year.
// totalMovies is the count of all movies released in the given year.
func (r *MovieRepository) GetByYear(year, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	rows, err := r.db.Query(
		`SELECT id, title, releaseYear, duration FROM Movie
		WHERE releaseYear = ?
		ORDER BY id
		LIMIT ? OFFSET ?`, year, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration); err != nil {
			return nil, 0, err
		}
		page = append(page, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM Movie WHERE releaseYear = ?", year).Scan(&totalMovies); err != nil {
		return nil, 0, err
	}

	return page, totalMovies, nil
}

// GetByActor fetches one page of the movies the given actor plays in.
// totalMovies is the total movie count of that actor in the Movie table.
func (r *MovieRepository) GetByActor(actorID, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	rows, err := r.db.Query(
		`SELECT m.id, m.title, m.releaseYear, m.duration
		FROM Movie AS m
		JOIN movie_actors AS ma ON ma.movie_id = m.id
		WHERE ma.actor_id = ?
		ORDER BY m.id
		LIMIT ? OFFSET ?`, actorID, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration); err != nil {
			return nil, 0, err
		}
		page = append(page, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM movie_actors WHERE actor_id = ?", actorID).Scan(&totalMovies); err != nil {
		return nil, 0, err
	}

	return page, totalMovies, nil
}

// SearchByTitle fetches one page of movies whose title contains the query,
// case-insensitively. totalMovies is the total count of matching movies.
func (r *MovieRepository) SearchByTitle(query string, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	rows, err := r.db.Query(
		`SELECT id, title, releaseYear, duration FROM Movie
		WHERE title LIKE '%' || ? || '%'
		ORDER BY id
		LIMIT ? OFFSET ?`, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.ReleaseYear, &m.Duration); err != nil {
			return nil, 0, err
		}
		page = append(page, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM Movie WHERE title LIKE '%' || ? || '%'", query).Scan(&totalMovies); err != nil {
		return nil, 0, err
	}

	return page, totalMovies, nil
}

// Update replaces the title, releaseYear and duration of the movie with the given id.
func (r *MovieRepository) Update(id int, m *models.Movie) (*models.Movie, error) {
	result, err := r.db.Exec(
		`UPDATE Movie SET title = ?, releaseYear = ?, duration = ? WHERE id = ?`,
		m.Title, m.ReleaseYear, m.Duration, id)
	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 { // UPDATE matched no row
		return nil, models.ErrNotFound
	}

	return r.GetByID(id)
}

func (r *MovieRepository) Delete(id int) error {
	// todo
	return nil
}

func (r *MovieRepository) SetGenres(movieID int, genreIDs []int) error {
	// todo
	return nil
}

func (r *MovieRepository) SetActors(movieID int, actorIDs []int) error {
	// todo
	return nil
}

func (r *MovieRepository) GetActorsByMovie(movieID int) ([]*models.Actor, error) {
	// todo
	return nil, nil
}

func (r *MovieRepository) GetGenresByMovie(movieID int) ([]*models.Genre, error) {
	//todo
	return nil, nil
}
