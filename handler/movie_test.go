package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"gitea.kood.tech/nikolailetshev/movies-api/repository"
	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

// newMovieTestHandler wires a real MovieRepository + MovieService over a
// throwaway SQLite database, so the handler tests run the whole stack.
func newMovieTestHandler(t *testing.T) *MovieHandler {
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

	repo := repository.NewMovieRepository(db)
	svc := service.NewMovieService(repo)
	return NewMovieHandler(svc)
}

// newMovieMux returns a full mux (needs genre/actor routes for linking too),
// with all three entities registered.
func newMovieMux(t *testing.T) *http.ServeMux {
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

	mux := http.NewServeMux()
	RegisterRoutes(mux, db)
	return mux
}

// createGenreAndActor seeds one genre and one actor, returning their IDs.
func createGenreAndActor(t *testing.T, mux *http.ServeMux) (genreID, actorID int) {
	t.Helper()
	genreID = createGenre(t, mux, "Sci-Fi")

	rec := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Leonardo DiCaprio","birthDate":"1974-11-11"}`)
	var a struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &a); err != nil {
		t.Fatalf("create actor: %v", err)
	}
	return genreID, a.ID
}

// createMovie posts a valid movie and returns its id. A genre + actor is the
// normal case; a movie may also be created with no links at all.
func createMovie(t *testing.T, mux *http.ServeMux, title string, year, dur int, genreIDs, actorIDs []int) int {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"title":       title,
		"releaseYear": year,
		"duration":    dur,
		"genres":      genreIDs,
		"actors":      actorIDs,
	})
	rec := doRequest(t, mux, http.MethodPost, "/api/movies", string(body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create movie: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var m struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode created movie: %v", err)
	}
	return m.ID
}

func TestMovieCreateHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)

	rec := doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"title":"Inception","releaseYear":2010,"duration":148,"genres":[`+
			strconv.Itoa(gid)+`],"actors":[`+strconv.Itoa(aid)+`]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var movie struct {
		ID          int                     `json:"id"`
		Title       string                  `json:"title"`
		ReleaseYear int                     `json:"releaseYear"`
		Duration    int                     `json:"duration"`
		Genres      []struct{ Name string } `json:"genres"`
		Actors      []struct{ Name string } `json:"actors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &movie); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if movie.ID < 1 || movie.Title != "Inception" {
		t.Fatalf("unexpected movie: %+v", movie)
	}
	if len(movie.Genres) != 1 || movie.Genres[0].Name != "Sci-Fi" {
		t.Fatalf("expected linked genre, got %+v", movie.Genres)
	}
	if len(movie.Actors) != 1 || movie.Actors[0].Name != "Leonardo DiCaprio" {
		t.Fatalf("expected linked actor, got %+v", movie.Actors)
	}
}

func TestMovieCreateInvalidBody(t *testing.T) {
	mux := newMovieMux(t)

	// missing required title/year/duration -> 400
	rec := doRequest(t, mux, http.MethodPost, "/api/movies", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("expected error object, got %s", rec.Body.String())
	}
}

func TestMovieCreateWithNoLinks(t *testing.T) {
	mux := newMovieMux(t)

	// genres and actors are optional; a movie may be created with no links
	rec := doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"title":"Orphan","releaseYear":2020,"duration":90,"genres":[],"actors":[]}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var movie struct {
		Genres []any `json:"genres"`
		Actors []any `json:"actors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &movie); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(movie.Genres) != 0 || len(movie.Actors) != 0 {
		t.Fatalf("expected empty genre/actor lists, got genres=%v actors=%v", movie.Genres, movie.Actors)
	}

	// omitting the lists entirely is also allowed
	rec = doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"title":"Orphan 2","releaseYear":2021,"duration":90}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("omitted lists: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMoviePatchClearsLinks(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	// an explicit empty list clears all of that link (like PATCH actor movies:[])
	rec := doRequest(t, mux, http.MethodPatch, "/api/movies/"+strconv.Itoa(id),
		`{"genres":[],"actors":[]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var movie struct {
		Genres []any `json:"genres"`
		Actors []any `json:"actors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &movie); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(movie.Genres) != 0 || len(movie.Actors) != 0 {
		t.Fatalf("expected cleared links, got genres=%v actors=%v", movie.Genres, movie.Actors)
	}
}

func TestMovieGetByIDHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	rec := doRequest(t, mux, http.MethodGet, "/api/movies/"+strconv.Itoa(id), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var movie struct {
		Title string `json:"title"`
	}
	json.Unmarshal(rec.Body.Bytes(), &movie)
	if movie.Title != "Inception" {
		t.Fatalf("unexpected movie: %+v", movie)
	}
}

func TestMovieGetByIDNotFoundHandler(t *testing.T) {
	mux := newMovieMux(t)

	rec := doRequest(t, mux, http.MethodGet, "/api/movies/99999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMovieGetAllPaginationShape(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	createMovie(t, mux, "One", 2000, 100, []int{gid}, []int{aid})
	createMovie(t, mux, "Two", 2001, 101, []int{gid}, []int{aid})

	rec := doRequest(t, mux, http.MethodGet, "/api/movies?page=0&size=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var page struct {
		Content       []any `json:"content"`
		Page          int   `json:"page"`
		Size          int   `json:"size"`
		TotalElements int   `json:"totalElements"`
		TotalPages    int   `json:"totalPages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	if page.Size != 1 || page.Page != 0 || page.TotalElements != 2 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if len(page.Content) != 1 {
		t.Fatalf("expected 1 movie on page, got %d", len(page.Content))
	}
}

func TestMovieSearchHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})
	createMovie(t, mux, "Incept Again", 2011, 149, []int{gid}, []int{aid})
	createMovie(t, mux, "The Matrix", 1999, 136, []int{gid}, []int{aid})

	rec := doRequest(t, mux, http.MethodGet, "/api/movies/search?title=incep", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var page struct {
		Content       []struct{ Title string } `json:"content"`
		TotalElements int                      `json:"totalElements"`
	}
	json.Unmarshal(rec.Body.Bytes(), &page)
	if page.TotalElements != 2 {
		t.Fatalf("expected 2 search matches, got %d", page.TotalElements)
	}
}

func TestMovieFilterByGenreHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	// a second genre so the filter only matches Inception
	other := createGenre(t, mux, "Action")
	createMovie(t, mux, "Matrix", 1999, 136, []int{other}, []int{aid})

	rec := doRequest(t, mux, http.MethodGet, "/api/movies?genre="+strconv.Itoa(gid), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var page struct {
		Content       []struct{ Title string } `json:"content"`
		TotalElements int                      `json:"totalElements"`
	}
	json.Unmarshal(rec.Body.Bytes(), &page)
	if page.TotalElements != 1 || page.Content[0].Title != "Inception" {
		t.Fatalf("expected 1 filtered movie, got %d: %+v", page.TotalElements, page.Content)
	}
}

func TestMovieGetCastAndGenres(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	recActors := doRequest(t, mux, http.MethodGet, "/api/movies/"+strconv.Itoa(id)+"/actors", "")
	if recActors.Code != http.StatusOK {
		t.Fatalf("get actors: expected 200, got %d", recActors.Code)
	}
	var actors struct {
		Content       []struct{ Name string } `json:"content"`
		TotalElements int                     `json:"totalElements"`
	}
	json.Unmarshal(recActors.Body.Bytes(), &actors)
	if actors.TotalElements != 1 || actors.Content[0].Name != "Leonardo DiCaprio" {
		t.Fatalf("unexpected cast: %+v", actors)
	}

	recGenres := doRequest(t, mux, http.MethodGet, "/api/movies/"+strconv.Itoa(id)+"/genres", "")
	if recGenres.Code != http.StatusOK {
		t.Fatalf("get genres: expected 200, got %d", recGenres.Code)
	}
	var genres struct {
		Content       []struct{ Name string } `json:"content"`
		TotalElements int                     `json:"totalElements"`
	}
	json.Unmarshal(recGenres.Body.Bytes(), &genres)
	if genres.TotalElements != 1 || genres.Content[0].Name != "Sci-Fi" {
		t.Fatalf("unexpected genres: %+v", genres)
	}
}

func TestMoviePartialUpdateHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	// partial PATCH: only title; year and duration must stay unchanged
	rec := doRequest(t, mux, http.MethodPatch, "/api/movies/"+strconv.Itoa(id),
		`{"title":"Inception Director's Cut"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var movie struct {
		Title       string `json:"title"`
		ReleaseYear int    `json:"releaseYear"`
		Duration    int    `json:"duration"`
	}
	json.Unmarshal(rec.Body.Bytes(), &movie)
	if movie.Title != "Inception Director's Cut" {
		t.Fatalf("expected updated title, got %q", movie.Title)
	}
	if movie.ReleaseYear != 2010 || movie.Duration != 148 {
		t.Fatalf("year/duration should be unchanged, got %+v", movie)
	}
}

func TestMovieDeleteHandler(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	rec := doRequest(t, mux, http.MethodDelete, "/api/movies/"+strconv.Itoa(id)+"?force=true", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := doRequest(t, mux, http.MethodGet, "/api/movies/"+strconv.Itoa(id), "")
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestMovieDeleteWithLinksRequiresForce(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)
	id := createMovie(t, mux, "Inception", 2010, 148, []int{gid}, []int{aid})

	// deleting a movie without ?force must be rejected (it has links)
	rec := doRequest(t, mux, http.MethodDelete, "/api/movies/"+strconv.Itoa(id), "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	rec2 := doRequest(t, mux, http.MethodDelete, "/api/movies/"+strconv.Itoa(id)+"?force=true", "")
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 with force, got %d", rec2.Code)
	}
}
