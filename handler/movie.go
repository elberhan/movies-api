package handler

import (
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

type MovieHandler struct {
	service *service.MovieService
}

func NewMovieHandler(s *service.MovieService) *MovieHandler {
	return &MovieHandler{service: s}
}

// RegisterMovieRoutes wires every /api/movies endpoint onto the mux.
func (h *MovieHandler) RegisterMovieRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/movies", h.Create)               // create                       -> 201
	mux.HandleFunc("GET /api/movies", h.GetAll)                // list (paginated, filtered)   -> 200
	mux.HandleFunc("GET /api/movies/search", h.Search)         // search (?title=)             -> 200
	mux.HandleFunc("GET /api/movies/{id}", h.GetByID)          // read one, with links         -> 200
	mux.HandleFunc("GET /api/movies/{id}/actors", h.GetActors) // cast (paginated)             -> 200
	mux.HandleFunc("GET /api/movies/{id}/genres", h.GetGenres) // genres (paginated)           -> 200
	mux.HandleFunc("PATCH /api/movies/{id}", h.Update)         // partial update               -> 200
	mux.HandleFunc("DELETE /api/movies/{id}", h.Delete)        // delete (?force=)             -> 204
}

func (h *MovieHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		ReleaseYear int    `json:"releaseYear"`
		Duration    int    `json:"duration"`
		Genres      []int  `json:"genres"`
		Actors      []int  `json:"actors"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	movie, err := h.service.Create(req.Title, req.ReleaseYear, req.Duration, req.Genres, req.Actors)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, movie)
}

func (h *MovieHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter, err := movieFilter(r)
	if err != nil {
		writeError(w, err)
		return
	}

	page, size, err := pageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.service.List(filter, page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MovieHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, size, err := pageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.service.SearchByTitle(r.URL.Query().Get("title"), page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MovieHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	movie, err := h.service.GetDetail(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (h *MovieHandler) GetActors(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	page, size, err := pageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.service.ActorsInMovie(id, page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MovieHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	page, size, err := pageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}

	result, err := h.service.GenresInMovie(id, page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MovieHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req struct {
		Title       *string `json:"title"`
		ReleaseYear *int    `json:"releaseYear"`
		Duration    *int    `json:"duration"`
		Genres      []int   `json:"genres"`
		Actors      []int   `json:"actors"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	movie, err := h.service.UpdatePartial(id, req.Title, req.ReleaseYear, req.Duration, req.Genres, req.Actors)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (h *MovieHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	force := r.URL.Query().Get("force") == "true"
	if err := h.service.Delete(id, force); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
