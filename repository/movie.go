package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// holds the database handle
type MovieRepository struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

// Create inserts a new movie together with its genre and actor links, in one
// transaction. A failure halfway changes nothing.
func (r *MovieRepository) Create(title string, releaseYear int, duration int, genreIDs, actorIDs []int) (*models.Movie, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO Movie (title, releaseYear, duration) VALUES (?, ?, ?)`,
		title, releaseYear, duration)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	movieID := int(id)

	if err := setGenresTx(tx, movieID, genreIDs); err != nil {
		return nil, err
	}
	if err := setActorsTx(tx, movieID, actorIDs); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Movie{
		ID:          movieID,
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

func (r *MovieRepository) Find(f models.MovieFilter, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	var joins, conditions []string
	var args []any

	if f.GenreID != nil {
		joins = append(joins, "JOIN movie_genres AS mg ON mg.movie_id = m.id AND mg.genre_id = ?")
		args = append(args, *f.GenreID)
	}
	if f.ActorID != nil {
		joins = append(joins, "JOIN movie_actors AS ma ON ma.movie_id = m.id AND ma.actor_id = ?")
		args = append(args, *f.ActorID)
	}
	if f.Year != nil {
		conditions = append(conditions, "m.releaseYear = ?")
		args = append(args, *f.Year)
	}
	if f.Title != "" {
		conditions = append(conditions, `m.title LIKE '%' || ? || '%'`)
		args = append(args, f.Title)
	}

	source := "FROM Movie AS m " + strings.Join(joins, " ")
	if len(conditions) > 0 {
		source += " WHERE " + strings.Join(conditions, " AND ")
	}

	if err := r.db.QueryRow("SELECT COUNT(*) "+source, args...).Scan(&totalMovies); err != nil {
		return nil, 0, err
	}

	pageArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.Query(
		"SELECT m.id, m.title, m.releaseYear, m.duration "+source+" ORDER BY m.id LIMIT ? OFFSET ?",
		pageArgs...)
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

	return page, totalMovies, nil
}

// GenreExists reports whether the given genre id is in the database. A filter
// on an unknown id must not look the same as one that matches no movie.
func (r *MovieRepository) GenreExists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM Genre WHERE id = ?)", id).Scan(&exists)
	return exists, err
}

// ActorExists reports whether the given actor id is in the database.
func (r *MovieRepository) ActorExists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM Actor WHERE id = ?)", id).Scan(&exists)
	return exists, err
}

// GetAll fetches one page of movies: limit rows, starting after offset rows.
// totalRows is the total row count in the Movie table.
func (r *MovieRepository) GetAll(limit, offset int) (page []*models.Movie, totalRows int, err error) {
	return r.Find(models.MovieFilter{}, limit, offset)
}

// GetByGenre fetches one page of the movies that belongs to the given genre.
// totalRows is the total movie count of specified genre in the Movie table.
func (r *MovieRepository) GetByGenre(genreID, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	return r.Find(models.MovieFilter{GenreID: &genreID}, limit, offset)
}

// GetByYear fetches one page of movies released in the given year.
// totalMovies is the count of all movies released in the given year.
func (r *MovieRepository) GetByYear(year, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	return r.Find(models.MovieFilter{Year: &year}, limit, offset)
}

// GetByActor fetches one page of the movies the given actor plays in.
// totalMovies is the total movie count of that actor in the Movie table.
func (r *MovieRepository) GetByActor(actorID, limit, offset int) (page []*models.Movie, totalMovies int, err error) {
	return r.Find(models.MovieFilter{ActorID: &actorID}, limit, offset)
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
	return r.Find(models.MovieFilter{Title: query}, limit, offset)
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

// setGenresTx replaces the genre links of the movie inside the caller's
// transaction, so that it can share one commit with other writes.
func setGenresTx(tx *sql.Tx, movieID int, genreIDs []int) error {
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
