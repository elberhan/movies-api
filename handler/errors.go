package handler

import (
	"errors"
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound) // 404

	case errors.Is(err, models.ErrInvalidInput),
		errors.Is(err, models.ErrHasRelations):
		http.Error(w, err.Error(), http.StatusBadRequest) // 400

	default:
		log.Printf("unhandled error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError) // 500
	}
}
