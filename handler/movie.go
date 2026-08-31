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
	// todo
}
