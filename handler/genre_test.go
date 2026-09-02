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

// newGenreTestHandler wires a real GenreRepository + GenreService over a
// throwaway SQLite database, so the handler tests run the whole stack.
func newGenreTestHandler(t *testing.T) *GenreHandler {
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

	repo := repository.NewGenreRepository(db)
	svc := service.NewGenreService(repo)
	return NewGenreHandler(svc)
}

// newGenreMux returns a mux with only the genre routes registered.
func newGenreMux(h *GenreHandler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterGenreRoutes(mux)
	return mux
}

func createGenre(t *testing.T, mux *http.ServeMux, name string) int {
	t.Helper()
	rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"`+name+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create genre: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var g struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil {
		t.Fatalf("decode created genre: %v", err)
	}
	return g.ID
}

func TestGenreCreateHandler(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"Sci-Fi"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var genre struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &genre); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if genre.ID < 1 || genre.Name != "Sci-Fi" {
		t.Fatalf("unexpected genre: %+v", genre)
	}
}

func TestGenreCreateDuplicateConflict(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	// first create succeeds
	rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"Sci-Fi"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// exact duplicate must be a 409 Conflict, not a 500
	rec = doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"Sci-Fi"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create: expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "already exists") {
		t.Fatalf("expected descriptive conflict message, got %s", rec.Body.String())
	}
}

func TestGenreCreateDuplicateCaseInsensitiveConflict(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	if rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"Action"}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", rec.Code)
	}
	// the name column is COLLATE NOCASE, so a different-case duplicate also conflicts
	rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{"name":"action"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("case-insensitive duplicate: expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGenreCreateMissingName(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	rec := doRequest(t, mux, http.MethodPost, "/api/genres", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("expected error object, got %s", rec.Body.String())
	}
}

func TestGenreGetByIDHandler(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	id := createGenre(t, mux, "Sci-Fi")
	rec := doRequest(t, mux, http.MethodGet, "/api/genres/"+strconv.Itoa(id), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var genre struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	json.Unmarshal(rec.Body.Bytes(), &genre)
	if genre.Name != "Sci-Fi" {
		t.Fatalf("unexpected genre: %+v", genre)
	}
}

func TestGenreGetByIDNotFoundHandler(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	rec := doRequest(t, mux, http.MethodGet, "/api/genres/99999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGenreGetAllPaginationShape(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	createGenre(t, mux, "Sci-Fi")
	createGenre(t, mux, "Comedy")

	rec := doRequest(t, mux, http.MethodGet, "/api/genres?page=0&size=1", "")
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
		t.Fatalf("expected 1 genre on page, got %d", len(page.Content))
	}
}

func TestGenrePartialUpdateHandler(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	id := createGenre(t, mux, "Sci-Fi")

	rec := doRequest(t, mux, http.MethodPatch, "/api/genres/"+strconv.Itoa(id), `{"name":"Science Fiction"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var genre struct {
		Name string `json:"name"`
	}
	json.Unmarshal(rec.Body.Bytes(), &genre)
	if genre.Name != "Science Fiction" {
		t.Fatalf("expected updated name, got %q", genre.Name)
	}
}

func TestGenreDeleteHandler(t *testing.T) {
	h := newGenreTestHandler(t)
	mux := newGenreMux(h)

	id := createGenre(t, mux, "Sci-Fi")

	rec := doRequest(t, mux, http.MethodDelete, "/api/genres/"+strconv.Itoa(id)+"?force=true", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := doRequest(t, mux, http.MethodGet, "/api/genres/"+strconv.Itoa(id), "")
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestGenreDeleteWithRelationsRequiresForce(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)

	// create a movie linked to the genre and actor
	create := doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"title":"Inception","releaseYear":2010,"duration":148,"genres":[`+
			strconv.Itoa(gid)+`],"actors":[`+strconv.Itoa(aid)+`]}`)
	var created struct {
		ID int `json:"id"`
	}
	json.Unmarshal(create.Body.Bytes(), &created)
	if created.ID < 1 {
		t.Fatalf("create movie failed: %s", create.Body.String())
	}

	// deleting a genre that still has movies without ?force= must be rejected
	rec := doRequest(t, mux, http.MethodDelete, "/api/genres/"+strconv.Itoa(gid), "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// with ?force=true it succeeds
	rec2 := doRequest(t, mux, http.MethodDelete, "/api/genres/"+strconv.Itoa(gid)+"?force=true", "")
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 with force, got %d", rec2.Code)
	}
}
