package handler

import (
	"net/http"
	"strconv"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

type GenreHandler struct {
	service *service.GenreService
}

func NewGenreHandler(s *service.GenreService) *GenreHandler {
	return &GenreHandler{service: s}
}

// RegisterGenreRoutes wires every /api/genres endpoint onto the mux.
func (h *GenreHandler) RegisterGenreRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/genres", h.Create)        // create            -> 201
	mux.HandleFunc("GET /api/genres", h.GetAll)         // list (paginated)  -> 200
	mux.HandleFunc("GET /api/genres/{id}", h.GetByID)   // read one          -> 200
	mux.HandleFunc("PATCH /api/genres/{id}", h.Update)  // partial update    -> 200
	mux.HandleFunc("DELETE /api/genres/{id}", h.Delete) // delete (?force=)  -> 204
}


func (h *GenreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil { // broken JSON -> ErrInvalidInput -> 400
		writeError(w, err)
		return
	}

	genre, err := h.service.Create(req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, genre)
}

func (h *GenreHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 0)
	size := queryInt(r, "size", 0)

	result, err := h.service.GetAll(page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GenreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	genre, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, genre)
}

func (h *GenreHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	genre, err := h.service.Update(id, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, genre)
}

func (h *GenreHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

//helpers
func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, models.ErrNotFound
	}
	return id, nil
}


func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback // param absent entirely
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback // param present but nonsense
	}
	return v
}
