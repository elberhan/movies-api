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
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})

	case errors.Is(err, models.ErrInvalidInput), errors.Is(err, models.ErrHasRelations):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})

	default:
		log.Printf("internal: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
