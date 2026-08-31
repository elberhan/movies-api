package repository

import (
	"database/sql"
	"errors"
	"fmt"

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
	if err != nil {
		return nil, 0, err
	}

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

// GetActorsByMovie fetches one page of the actors starring in the given movie.
// totalActors is the total actor count of that movie.
func (r *MovieRepository) GetActorsByMovie(movieID, limit, offset int) (actors []*models.Actor, totalActors int, err error) {
	rows, err := r.db.Query(
		`SELECT a.id, a.name, a.birthDate
		FROM Actor AS a
		JOIN movie_actors AS ma ON ma.actor_id = a.id
		WHERE ma.movie_id = ?
		ORDER BY a.name
		LIMIT ? OFFSET ?`, movieID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var a models.Actor
		if err := rows.Scan(&a.ID, &a.Name, &a.BirthDate); err != nil {
			return nil, 0, err
		}
		actors = append(actors, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM movie_actors WHERE movie_id = ?", movieID).Scan(&totalActors); err != nil {
		return nil, 0, err
	}

	return actors, totalActors, nil
}

// GetGenresByMovie fetches one page of the genres of the given movie.
// totalGenres is the total genre count of that movie.
func (r *MovieRepository) GetGenresByMovie(movieID, limit, offset int) (genres []*models.Genre, totalGenres int, err error) {
	rows, err := r.db.Query(
		`SELECT g.id, g.name
		FROM Genre AS g
		JOIN movie_genres AS mg ON mg.genre_id = g.id
		WHERE mg.movie_id = ?
		ORDER BY g.name
		LIMIT ? OFFSET ?`, movieID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

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

	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM movie_genres WHERE movie_id = ?", movieID).Scan(&totalGenres); err != nil {
		return nil, 0, err
	}

	return genres, totalGenres, nil
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

// Update replaces the title, releaseYear and duration of the movie with the
// given id and, for every list that is not nil, its genre and actor links. It
// is one transaction, so an update that fails halfway changes nothing.
func (r *MovieRepository) Update(id int, m *models.Movie, genreIDs, actorIDs []int) (*models.Movie, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
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

	if genreIDs != nil {
		if err := setGenresTx(tx, id, genreIDs); err != nil {
			return nil, err
		}
	}
	if actorIDs != nil {
		if err := setActorsTx(tx, id, actorIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

// Delete removes the movie with the given id. If movie_genres or
// movie_actors still reference it, the RESTRICT foreign keys make the
// delete fail, so movies with relationships cannot be removed this way.
func (r *MovieRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM Movie WHERE id = ?", id)
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

// ForceDelete removes all genre and actor links of the movie (not actors/genres
// themselves) and then the movie itself. Everything runs in one transaction, so
// nothing would be deleted in case of failure.
func (r *MovieRepository) ForceDelete(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM movie_genres WHERE movie_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM movie_actors WHERE movie_id = ?", id); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM Movie WHERE id = ?", id)
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

	return tx.Commit()
}

// SetGenres makes genreIDs the new genre list of the movie, in one transaction.
// An empty genreIDs is an error.
func (r *MovieRepository) SetGenres(movieID int, genreIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := ensureMovieExists(tx, movieID); err != nil {
		return err
	}
	if err := setGenresTx(tx, movieID, genreIDs); err != nil {
		return err
	}

	return tx.Commit()
}

// SetActors makes actorIDs the new actor list of the movie, in one transaction.
// An empty actorIDs is an error: a movie must keep a minimum of one actor.
func (r *MovieRepository) SetActors(movieID int, actorIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := ensureMovieExists(tx, movieID); err != nil {
		return err
	}
	if err := setActorsTx(tx, movieID, actorIDs); err != nil {
		return err
	}

	return tx.Commit()
}

// setGenresTx replaces the genre links of the movie inside the caller's
// transaction, so that it can share one commit with other writes.
func setGenresTx(tx *sql.Tx, movieID int, genreIDs []int) error {
	if len(genreIDs) == 0 {
		return fmt.Errorf("%w: a movie must have a minimum of one genre", models.ErrInvalidInput)
	}

	// remove old links to genres and set new ones
	if _, err := tx.Exec("DELETE FROM movie_genres WHERE movie_id = ?", movieID); err != nil {
		return err
	}

	for _, genreID := range genreIDs {
		var exists int
		err := tx.QueryRow("SELECT 1 FROM Genre WHERE id = ?", genreID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: genre with id %d does not exist", models.ErrInvalidInput, genreID)
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO movie_genres (movie_id, genre_id) VALUES (?, ?)", movieID, genreID); err != nil {
			return err
		}
	}

	return nil
}

// setActorsTx replaces the actors linked with specified movie inside the caller's
// transaction, so that it can share one commit with other writes.
func setActorsTx(tx *sql.Tx, movieID int, actorIDs []int) error {
	if len(actorIDs) == 0 {
		return fmt.Errorf("%w: a movie must have a minimum of one actor", models.ErrInvalidInput)
	}

	// remove old links between movie and actors, and set new ones
	if _, err := tx.Exec("DELETE FROM movie_actors WHERE movie_id = ?", movieID); err != nil {
		return err
	}

	for _, actorID := range actorIDs {
		var exists int
		err := tx.QueryRow("SELECT 1 FROM Actor WHERE id = ?", actorID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: actor with id %d does not exist", models.ErrInvalidInput, actorID)
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO movie_actors (movie_id, actor_id) VALUES (?, ?)", movieID, actorID); err != nil {
			return err
		}
	}

	return nil
}

// ensureMovieExists returns ErrNotFound if the movie is not in the database.
func ensureMovieExists(tx *sql.Tx, movieID int) error {
	var exists int
	err := tx.QueryRow("SELECT 1 FROM Movie WHERE id = ?", movieID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrNotFound
	}
	return err
}
