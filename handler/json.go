package handler

import (
	"encoding/json"
	"fmt"
	"gitea.kood.tech/nikolailetshev/movies-api/models"
	"log"
	"net/http"
)

// Used for every success response: 200 OK, 201 Created, paginated lists...
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json response: %v", err)
	}
}

// decodeJSON reads a request body into out (any struct pointer).
func decodeJSON(r *http.Request, out any) error {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		return fmt.Errorf("%w: invalid JSON body: %v", models.ErrInvalidInput, err)
	}
	return nil
}
