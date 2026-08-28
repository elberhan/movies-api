package repository

import (
	"database/sql"
	"errors"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// ActorRepository is the only place that runs SQL against the Actor tables.
type ActorRepository struct {
	db *sql.DB
}

// NewActorRepository stores the shared database handle.
func NewActorRepository(db *sql.DB) *ActorRepository {
	return &ActorRepository{db: db}
}

func (r *ActorRepository) Create(name, birthDate string) (*models.Actor, error) {
	result, err := r.db.Exec("INSERT INTO Actor (name, birthDate) VALUES (?, ?)", name, birthDate)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId() // SQLite auto-generates the id
	if err != nil {
		return nil, err
	}
	return &models.Actor{ID: int(id), Name: name, BirthDate: birthDate}, nil
}

func (r *ActorRepository) GetByID(id int) (*models.Actor, error) {
	row := r.db.QueryRow("SELECT id, name, birthDate FROM Actor WHERE id = ?", id)
	var a models.Actor
	err := row.Scan(&a.ID, &a.Name, &a.BirthDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ActorRepository) GetAll(limit, offset int) ([]*models.Actor, int, error) {
	rows, err := r.db.Query("SELECT id, name, birthDate FROM Actor ORDER BY name LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close() // always release the cursor

	actors := []*models.Actor{}
	for rows.Next() {
		var a models.Actor
		if err := rows.Scan(&a.ID, &a.Name, &a.BirthDate); err != nil {
			return nil, 0, err
		}
		actors = append(actors, &a)
	}
	if err := rows.Err(); err != nil { // check cursor errors after loop
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM Actor").Scan(&total); err != nil {
		return nil, 0, err
	}
	return actors, total, nil
}

func (r *ActorRepository) GetByName(name string, limit, offset int) ([]*models.Actor, error) {
	rows, err := r.db.Query(
		"SELECT id, name, birthDate FROM Actor WHERE name LIKE '%' || ? || '%' ORDER BY name LIMIT ? OFFSET ?",
		name, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actors := []*models.Actor{}
	for rows.Next() {
		var a models.Actor
		if err := rows.Scan(&a.ID, &a.Name, &a.BirthDate); err != nil {
			return nil, err
		}
		actors = append(actors, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actors, nil
}

func (r *ActorRepository) Update(id int, name, birthDate string) (*models.Actor, error) {
	result, err := r.db.Exec("UPDATE Actor SET name = ?, birthDate = ? WHERE id = ?", name, birthDate, id)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, models.ErrNotFound // UPDATE matched no row
	}
	return r.GetByID(id)
}

func (r *ActorRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM Actor WHERE id = ?", id)
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

func (r *ActorRepository) MovieCount(id int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM movie_actors WHERE actor_id = ?", id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ActorRepository) GetMoviesByActor(actorID int) ([]*models.Movie, error) {
	rows, err := r.db.Query(`
		SELECT m.id, m.title, m.releaseYear, m.duration
		FROM Movie m
		JOIN movie_actors ma ON ma.movie_id = m.id
		WHERE ma.actor_id = ?
		ORDER BY m.title`, actorID)
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

func (r *ActorRepository) DeleteActorMovieLinks(actorID int) error {
	_, err := r.db.Exec("DELETE FROM movie_actors WHERE actor_id = ?", actorID)
	return err
}
