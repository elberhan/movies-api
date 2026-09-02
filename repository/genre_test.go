package repository

import (
	"errors"
	"testing"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

func TestGenreCreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	g, err := repo.Create("Sci-Fi")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if g.ID < 1 {
		t.Fatalf("expected generated id, got %d", g.ID)
	}

	got, err := repo.GetByID(g.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != "Sci-Fi" {
		t.Fatalf("unexpected genre: %+v", got)
	}
}

func TestGenreCreateDuplicateConflict(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	if _, err := repo.Create("Sci-Fi"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := repo.Create("Sci-Fi")
	if !errors.Is(err, models.ErrConflict) {
		t.Fatalf("expected ErrConflict for duplicate, got %v", err)
	}
}

func TestGenreGetByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	_, err := repo.GetByID(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGenreGetAllPagination(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	for _, name := range []string{"Zombie", "Comedy", "Action"} {
		if _, err := repo.Create(name); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	genres, total, err := repo.GetAll(2, 0)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(genres) != 2 {
		t.Fatalf("expected 2 rows on page, got %d", len(genres))
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	// rows are ordered by name
	if genres[0].Name != "Action" || genres[1].Name != "Comedy" {
		t.Fatalf("unexpected ordering: %q, %q", genres[0].Name, genres[1].Name)
	}
}

func TestGenreUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	g, err := repo.Create("Sci-Fi")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := repo.Update(g.ID, "Science Fiction")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Science Fiction" {
		t.Fatalf("unexpected updated genre: %+v", updated)
	}
}

func TestGenreUpdateNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	_, err := repo.Update(999, "Nobody")
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGenreDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewGenreRepository(db)

	g, err := repo.Create("Sci-Fi")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(g.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(g.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
	if err := repo.Delete(g.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("second delete should be ErrNotFound, got %v", err)
	}
}

func TestGenreMovieCountAndLinks(t *testing.T) {
	db := newTestDB(t)
	genreRepo := NewGenreRepository(db)

	genre, err := genreRepo.Create("Sci-Fi")
	if err != nil {
		t.Fatalf("create genre: %v", err)
	}

	// insert a movie + a join row directly to exercise the join-table methods
	if _, err := db.Exec("INSERT INTO Movie (title, releaseYear, duration) VALUES (?, ?, ?)",
		"Inception", 2010, 148); err != nil {
		t.Fatalf("insert movie: %v", err)
	}
	if _, err := db.Exec("INSERT INTO movie_genres (movie_id, genre_id) VALUES (?, ?)",
		1, genre.ID); err != nil {
		t.Fatalf("insert link: %v", err)
	}

	count, err := genreRepo.MovieCount(genre.ID)
	if err != nil {
		t.Fatalf("movie count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	movies, err := genreRepo.GetMoviesByGenre(genre.ID, 10, 0)
	if err != nil {
		t.Fatalf("get movies by genre: %v", err)
	}
	if len(movies) != 1 || movies[0].Title != "Inception" {
		t.Fatalf("unexpected movies: %+v", movies)
	}

	if err := genreRepo.DeleteGenreMovieLinks(genre.ID); err != nil {
		t.Fatalf("delete links: %v", err)
	}
	if count, _ := genreRepo.MovieCount(genre.ID); count != 0 {
		t.Fatalf("expected 0 links after cleanup, got %d", count)
	}
}
