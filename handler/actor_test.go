package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"gitea.kood.tech/nikolailetshev/movies-api/repository"
	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

// newActorTestHandler wires a real ActorRepository + ActorService over a
// throwaway SQLite database, so the handler tests run the whole stack.
func newActorTestHandler(t *testing.T) *ActorHandler {
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

	repo := repository.NewActorRepository(db)
	svc := service.NewActorService(repo)
	return NewActorHandler(svc)
}

// newActorMux returns a mux with only the actor routes registered.
func newActorMux(h *ActorHandler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterActorRoutes(mux)
	return mux
}

func doRequest(t *testing.T, mux *http.ServeMux, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, target, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestActorCreateHandler(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	rec := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Tom Hanks","birthDate":"1956-07-09"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var actor struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		BirthDate string `json:"birthDate"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &actor); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if actor.ID < 1 || actor.Name != "Tom Hanks" {
		t.Fatalf("unexpected actor: %+v", actor)
	}
}

func TestActorCreateInvalidBirthDate(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	rec := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Bad","birthDate":"1990-13-40"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("expected error object, got %s", rec.Body.String())
	}
}

func TestActorCreateMissingName(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	rec := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"birthDate":"1956-07-09"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestActorGetByIDHandler(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	create := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Tom Hanks","birthDate":"1956-07-09"}`)
	var created struct {
		ID int `json:"id"`
	}
	json.Unmarshal(create.Body.Bytes(), &created)

	rec := doRequest(t, mux, http.MethodGet, "/api/actors/"+strconv.Itoa(created.ID), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestActorGetByIDNotFoundHandler(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	rec := doRequest(t, mux, http.MethodGet, "/api/actors/99999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestActorGetAllPaginationShape(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	rec := doRequest(t, mux, http.MethodGet, "/api/actors?page=0&size=5", "")
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
	if page.TotalElements < 0 || page.Size != 5 || page.Page != 0 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestActorPartialUpdateHandler(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	create := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Tom Hanks","birthDate":"1956-07-09"}`)
	var created struct {
		ID int `json:"id"`
	}
	json.Unmarshal(create.Body.Bytes(), &created)

	// partial PATCH: only name, birthDate must be untouched
	rec := doRequest(t, mux, http.MethodPatch, "/api/actors/"+strconv.Itoa(created.ID),
		`{"name":"Tom Hanks Jr."}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var actor struct {
		Name      string `json:"name"`
		BirthDate string `json:"birthDate"`
	}
	json.Unmarshal(rec.Body.Bytes(), &actor)
	if actor.Name != "Tom Hanks Jr." {
		t.Fatalf("expected updated name, got %q", actor.Name)
	}
	if actor.BirthDate != "1956-07-09" {
		t.Fatalf("birthDate should be unchanged, got %q", actor.BirthDate)
	}
}

func TestActorDeleteHandler(t *testing.T) {
	h := newActorTestHandler(t)
	mux := newActorMux(h)

	create := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"Tom Hanks","birthDate":"1956-07-09"}`)
	var created struct {
		ID int `json:"id"`
	}
	json.Unmarshal(create.Body.Bytes(), &created)

	rec := doRequest(t, mux, http.MethodDelete, "/api/actors/"+strconv.Itoa(created.ID)+"?force=true", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// now it should be gone
	rec2 := doRequest(t, mux, http.MethodGet, "/api/actors/"+strconv.Itoa(created.ID), "")
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestActorPATCHMoviesHandler(t *testing.T) {
	mux := newMovieMux(t)

	// a genre and two actors
	gid, aidA := createGenreAndActor(t, mux)
	aidB := createActor(t, mux, "Helen Hunt")

	// two movies, both starring actor A only
	mid1 := createMovie(t, mux, "Cast Away", 2000, 143, []int{gid}, []int{aidA})
	mid2 := createMovie(t, mux, "The Terminal", 2004, 128, []int{gid}, []int{aidA})

	// link actor B to both movies via PATCH
	rec := doRequest(t, mux, http.MethodPatch, "/api/actors/"+strconv.Itoa(aidB),
		`{"name":"Helen Hunt Jr","movies":[`+strconv.Itoa(mid1)+`,`+strconv.Itoa(mid2)+`]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var actor struct {
		Name string `json:"name"`
	}
	json.Unmarshal(rec.Body.Bytes(), &actor)
	if actor.Name != "Helen Hunt Jr" {
		t.Fatalf("expected updated name, got %q", actor.Name)
	}

	// both movies now show two actors
	for _, mid := range []int{mid1, mid2} {
		page := getMovieActors(t, mux, mid)
		if page.TotalElements != 2 {
			t.Fatalf("expected 2 actors in movie %d, got %d", mid, page.TotalElements)
		}
	}

	// replacing with a single movie must remove actor B from the other
	rec = doRequest(t, mux, http.MethodPatch, "/api/actors/"+strconv.Itoa(aidB),
		`{"movies":[`+strconv.Itoa(mid1)+`]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("replace movies: expected 200, got %d", rec.Code)
	}
	if page := getMovieActors(t, mux, mid1); page.TotalElements != 2 {
		t.Fatalf("expected actor B still in movie 1, got %d", page.TotalElements)
	}
	if page := getMovieActors(t, mux, mid2); page.TotalElements != 1 {
		t.Fatalf("expected actor B removed from movie 2, got %d", page.TotalElements)
	}

	// clearing with an empty list must remove actor B everywhere
	rec = doRequest(t, mux, http.MethodPatch, "/api/actors/"+strconv.Itoa(aidB), `{"movies":[]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("clear movies: expected 200, got %d", rec.Code)
	}
	if page := getMovieActors(t, mux, mid1); page.TotalElements != 1 {
		t.Fatalf("expected 1 actor after clear, got %d", page.TotalElements)
	}

	// an unknown movie id is invalid input
	rec = doRequest(t, mux, http.MethodPatch, "/api/actors/"+strconv.Itoa(aidB), `{"movies":[99999]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown movie, got %d: %s", rec.Code, rec.Body.String())
	}
}

// createActor posts a new actor and returns its id.
func createActor(t *testing.T, mux *http.ServeMux, name string) int {
	t.Helper()
	rec := doRequest(t, mux, http.MethodPost, "/api/actors",
		`{"name":"`+name+`","birthDate":"1963-06-15"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create actor %s: expected 201, got %d: %s", name, rec.Code, rec.Body.String())
	}
	var a struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &a); err != nil {
		t.Fatalf("decode created actor: %v", err)
	}
	return a.ID
}

// getMovieActors returns the actor page of a movie.
func getMovieActors(t *testing.T, mux *http.ServeMux, movieID int) struct {
	Content       []struct{ Name string } `json:"content"`
	TotalElements int                     `json:"totalElements"`
} {
	t.Helper()
	rec := doRequest(t, mux, http.MethodGet,
		"/api/movies/"+strconv.Itoa(movieID)+"/actors", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get movie %d actors: expected 200, got %d", movieID, rec.Code)
	}
	var page struct {
		Content       []struct{ Name string } `json:"content"`
		TotalElements int                     `json:"totalElements"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode movie actors: %v", err)
	}
	return page
}
