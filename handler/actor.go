package handler

import (
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

type ActorHandler struct {
	service *service.ActorService
}

func NewActorHandler(s *service.ActorService) *ActorHandler {
	return &ActorHandler{service: s}
}

// RegisterActorRoutes wires every /api/actors endpoint onto the mux.
func (h *ActorHandler) RegisterActorRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/actors", h.Create)        // create            -> 201
	mux.HandleFunc("GET /api/actors", h.GetAll)         // list (paginated, ?name=) -> 200
	mux.HandleFunc("GET /api/actors/{id}", h.GetByID)   // read one          -> 200
	mux.HandleFunc("PATCH /api/actors/{id}", h.Update)  // partial update    -> 200
	mux.HandleFunc("DELETE /api/actors/{id}", h.Delete) // delete (?force=)  -> 204
}

func (h *ActorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		BirthDate string `json:"birthDate"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	actor, err := h.service.Create(req.Name, req.BirthDate)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, actor)
}

func (h *ActorHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 0)
	size := queryInt(r, "size", 0)

	if name := r.URL.Query().Get("name"); name != "" {
		result, err := h.service.GetByName(name, page, size)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	result, err := h.service.GetAll(page, size)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ActorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	actor, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, actor)
}

func (h *ActorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req struct {
		Name      *string `json:"name"`
		BirthDate *string `json:"birthDate"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	actor, err := h.service.UpdatePartial(id, req.Name, req.BirthDate)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, actor)
}

func (h *ActorHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
