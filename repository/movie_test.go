package repository

import (
	"database/sql"
	"errors"
	"testing"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

func TestMovieCreateAndGetByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)

	g, a := mustGenreAndActor(t, db)
	m, err := repo.Create("Inception", 2010, 148, []int{g.ID}, []int{a.ID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if m.ID < 1 {
		t.Fatalf("expected generated id, got %d", m.ID)
	}

	got, err := repo.GetByID(m.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Title != "Inception" || got.ReleaseYear != 2010 || got.Duration != 148 {
		t.Fatalf("unexpected movie: %+v", got)
	}
}

// mustGenreAndActor creates a genre and an actor directly through their repos,
// satisfying the "a movie must have at least one genre and one actor" rule.
func mustGenreAndActor(t *testing.T, db *sql.DB) (*models.Genre, *models.Actor) {
	t.Helper()
	g, err := NewGenreRepository(db).Create("Sci-Fi")
	if err != nil {
		t.Fatalf("create genre: %v", err)
	}
	a, err := NewActorRepository(db).Create("Leo", "1974-11-11")
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	return g, a
}

func TestMovieCreateWithLinks(t *testing.T) {
	db := newTestDB(t)
	movieRepo := NewMovieRepository(db)
	genreRepo := NewGenreRepository(db)
	actorRepo := NewActorRepository(db)

	genre, _ := genreRepo.Create("Sci-Fi")
	actor, _ := actorRepo.Create("Leonardo DiCaprio", "1974-11-11")

	m, err := movieRepo.Create("Inception", 2010, 148,
		[]int{genre.ID}, []int{actor.ID})
	if err != nil {
		t.Fatalf("create with links: %v", err)
	}

	genres, _, err := movieRepo.GetGenresByMovie(m.ID, 10, 0)
	if err != nil {
		t.Fatalf("get genres by movie: %v", err)
	}
	if len(genres) != 1 || genres[0].Name != "Sci-Fi" {
		t.Fatalf("unexpected genres: %+v", genres)
	}

	actors, _, err := movieRepo.GetActorsByMovie(m.ID, 10, 0)
	if err != nil {
		t.Fatalf("get actors by movie: %v", err)
	}
	if len(actors) != 1 || actors[0].Name != "Leonardo DiCaprio" {
		t.Fatalf("unexpected actors: %+v", actors)
	}
}

func TestMovieGetByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)

	_, err := repo.GetByID(999)
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMovieGetAllPagination(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)
	g, a := mustGenreAndActor(t, db)

	for _, title := range []string{"Zeta", "Alpha", "Mike"} {
		if _, err := repo.Create(title, 2000, 100, []int{g.ID}, []int{a.ID}); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
	}

	movies, total, err := repo.GetAll(2, 0)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(movies) != 2 {
		t.Fatalf("expected 2 rows on page, got %d", len(movies))
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	// rows are ordered by id
	if movies[0].Title != "Zeta" || movies[1].Title != "Alpha" {
		t.Fatalf("unexpected ordering: %q, %q", movies[0].Title, movies[1].Title)
	}
}

func TestMovieSearchByTitle(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)
	g, a := mustGenreAndActor(t, db)

	for _, title := range []string{"Inception", "Incept", "The Matrix"} {
		if _, err := repo.Create(title, 2000, 100, []int{g.ID}, []int{a.ID}); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
	}

	movies, total, err := repo.SearchByTitle("incep", 10, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if total != 2 || len(movies) != 2 {
		t.Fatalf("expected 2 matches, got %d (len %d)", total, len(movies))
	}
}

func TestMovieUpdateNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)

	_, err := repo.Update(999, &models.Movie{Title: "Ghost", ReleaseYear: 2000, Duration: 100}, nil, nil)
	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMovieDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovieRepository(db)
	g, a := mustGenreAndActor(t, db)

	m, err := repo.Create("Inception", 2010, 148, []int{g.ID}, []int{a.ID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.Delete(m.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(m.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMovieFindByFilters(t *testing.T) {
	db := newTestDB(t)
	movieRepo := NewMovieRepository(db)
	genreRepo := NewGenreRepository(db)
	actorRepo := NewActorRepository(db)

	g, _ := genreRepo.Create("Drama")
	g2, _ := genreRepo.Create("Comedy")
	a, _ := actorRepo.Create("Tom Hanks", "1956-07-09")
	a2, _ := actorRepo.Create("Helen Hunt", "1963-06-15")

	movieRepo.Create("The Terminal", 2004, 128, []int{g.ID}, []int{a.ID})
	movieRepo.Create("Cast Away", 2000, 143, []int{g.ID}, []int{a2.ID})
	movieRepo.Create("The Big Bang Theory Special", 2002, 90, []int{g2.ID}, []int{a.ID})

	// filter by genre
	movies, total, err := movieRepo.Find(models.MovieFilter{GenreID: &g.ID}, 10, 0)
	if err != nil {
		t.Fatalf("filter by genre: %v", err)
	}
	if total != 2 || len(movies) != 2 {
		t.Fatalf("expected 2 by genre, got %d (len %d)", total, len(movies))
	}

	// filter by actor (only Cast Away has Helen Hunt)
	movies, total, err = movieRepo.Find(models.MovieFilter{ActorID: &a2.ID}, 10, 0)
	if err != nil {
		t.Fatalf("filter by actor: %v", err)
	}
	if total != 1 || len(movies) != 1 || movies[0].Title != "Cast Away" {
		t.Fatalf("expected 1 by actor, got %d (len %d) %+v", total, len(movies), movies)
	}

	// filter by year
	year := 2000
	movies, total, err = movieRepo.Find(models.MovieFilter{Year: &year}, 10, 0)
	if err != nil {
		t.Fatalf("filter by year: %v", err)
	}
	if total != 1 || len(movies) != 1 || movies[0].Title != "Cast Away" {
		t.Fatalf("expected 1 by year, got %d (len %d)", total, len(movies))
	}
}

func TestMovieForceDeleteClearsLinks(t *testing.T) {
	db := newTestDB(t)
	movieRepo := NewMovieRepository(db)
	genreRepo := NewGenreRepository(db)
	actorRepo := NewActorRepository(db)

	g, _ := genreRepo.Create("Sci-Fi")
	a, _ := actorRepo.Create("Leo", "1974-11-11")

	m, err := movieRepo.Create("Inception", 2010, 148, []int{g.ID}, []int{a.ID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := movieRepo.ForceDelete(m.ID); err != nil {
		t.Fatalf("force delete: %v", err)
	}

	// links must be gone and the movie gone
	if _, err := movieRepo.GetByID(m.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after force delete, got %v", err)
	}
	if count, _ := genreRepo.MovieCount(g.ID); count != 0 {
		t.Fatalf("expected 0 genre links, got %d", count)
	}
	if count, _ := actorRepo.MovieCount(a.ID); count != 0 {
		t.Fatalf("expected 0 actor links, got %d", count)
	}
}

func TestMovieUpdateClearsLinksWithEmptyLists(t *testing.T) {
	db := newTestDB(t)
	movieRepo := NewMovieRepository(db)
	genreRepo := NewGenreRepository(db)
	actorRepo := NewActorRepository(db)

	g, _ := genreRepo.Create("Sci-Fi")
	a, _ := actorRepo.Create("Leo", "1974-11-11")

	m, err := movieRepo.Create("Inception", 2010, 148, []int{g.ID}, []int{a.ID})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// an explicit (non-nil) empty list must clear that link, not error
	updated := &models.Movie{Title: "Inception", ReleaseYear: 2010, Duration: 148}
	if _, err := movieRepo.Update(m.ID, updated, []int{}, []int{}); err != nil {
		t.Fatalf("update with empty lists: %v", err)
	}

	if count, _ := genreRepo.MovieCount(g.ID); count != 0 {
		t.Fatalf("expected 0 genre links after clear, got %d", count)
	}
	if count, _ := actorRepo.MovieCount(a.ID); count != 0 {
		t.Fatalf("expected 0 actor links after clear, got %d", count)
	}
}
