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
