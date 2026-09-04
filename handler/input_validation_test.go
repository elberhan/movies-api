package handler

import (
	"net/http"
	"strconv"
	"testing"
)

// The reviewer's example: a POST carrying an unknown field must be rejected.
// Previously json.Decode silently dropped "xxx"; now DisallowUnknownFields makes
// it a 400.
func TestCreateMovieRejectsUnknownField(t *testing.T) {
	mux := newMovieMux(t)

	rec := doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"xxx":"ssss","title":"Egoist3","releaseYear":2004,"duration":722}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown field, got %d: %s", rec.Code, rec.Body.String())
	}
}

// A well-formed request (no unknown fields) still succeeds.
func TestCreateMovieAcceptsKnownFieldsNoUnknown(t *testing.T) {
	mux := newMovieMux(t)
	gid, aid := createGenreAndActor(t, mux)

	rec := doRequest(t, mux, http.MethodPost, "/api/movies",
		`{"title":"Egoist3","releaseYear":2004,"duration":722,`+
			`"genres":[`+strconv.Itoa(gid)+`],"actors":[`+strconv.Itoa(aid)+`]}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for a valid request, got %d: %s", rec.Code, rec.Body.String())
	}
}
