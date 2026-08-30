package repository

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// newTestDB opens a throwaway SQLite database (a temp file, so foreign keys
// are enforced) and runs the real schema so every query runs against the
// same table shape as production.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("run schema: %v", err)
	}
	return db
}

func TestActorCreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	a, err := repo.Create("Tom Hanks", "1956-07-09")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.ID < 1 {
		t.Fatalf("expected generated id, got %d", a.ID)
	}

	got, err := repo.GetByID(a.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Tom Hanks" || got.BirthDate != "1956-07-09" {
		t.Fatalf("unexpected actor: %+v", got)
	}
}

func TestActorGetByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	_, err := repo.GetByID(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestActorGetAllPagination(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	for _, name := range []string{"Zoe", "Anna", "Mike"} {
		if _, err := repo.Create(name, "2000-01-01"); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	actors, total, err := repo.GetAll(2, 0) // limit 2, offset 0
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(actors) != 2 {
		t.Fatalf("expected 2 rows on page, got %d", len(actors))
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	// rows are ordered by name
	if actors[0].Name != "Anna" || actors[1].Name != "Mike" {
		t.Fatalf("unexpected ordering: %q, %q", actors[0].Name, actors[1].Name)
	}
}

func TestActorGetByNameCaseInsensitive(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	if _, err := repo.Create("Tom Hanks", "1956-07-09"); err != nil {
		t.Fatalf("create: %v", err)
	}

	matches, err := repo.GetByName("hanks", 10, 0)
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if len(matches) != 1 || matches[0].Name != "Tom Hanks" {
		t.Fatalf("expected one partial match, got %+v", matches)
	}
}

func TestActorUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	a, err := repo.Create("Tom Hanks", "1956-07-09")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := repo.Update(a.ID, "Tom H. Jr", "1956-07-10")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Tom H. Jr" || updated.BirthDate != "1956-07-10" {
		t.Fatalf("unexpected updated actor: %+v", updated)
	}
}

func TestActorUpdateNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	_, err := repo.Update(999, "Nobody", "2000-01-01")
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestActorDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewActorRepository(db)

	a, err := repo.Create("Tom Hanks", "1956-07-09")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(a.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	if err := repo.Delete(a.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("second delete should be ErrNotFound, got %v", err)
	}
}

func TestActorMovieCountAndLinks(t *testing.T) {
	db := newTestDB(t)
	actorRepo := NewActorRepository(db)

	actor, err := actorRepo.Create("Tom Hanks", "1956-07-09")
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}

	// insert a movie + a join row directly to exercise the join-table methods
	if _, err := db.Exec("INSERT INTO Movie (title, releaseYear, duration) VALUES (?, ?, ?)",
		"Cast Away", 2000, 143); err != nil {
		t.Fatalf("insert movie: %v", err)
	}
	if _, err := db.Exec("INSERT INTO movie_actors (movie_id, actor_id) VALUES (?, ?)",
		1, actor.ID); err != nil {
		t.Fatalf("insert link: %v", err)
	}

	count, err := actorRepo.MovieCount(actor.ID)
	if err != nil {
		t.Fatalf("movie count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	movies, err := actorRepo.GetMoviesByActor(actor.ID)
	if err != nil {
		t.Fatalf("get movies by actor: %v", err)
	}
	if len(movies) != 1 || movies[0].Title != "Cast Away" {
		t.Fatalf("unexpected movies: %+v", movies)
	}

	if err := actorRepo.DeleteActorMovieLinks(actor.ID); err != nil {
		t.Fatalf("delete links: %v", err)
	}
	if count, _ := actorRepo.MovieCount(actor.ID); count != 0 {
		t.Fatalf("expected 0 links after cleanup, got %d", count)
	}
}
