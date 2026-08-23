package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

/* writeJSON puts the JSON label on the response,
   stamps it with the caller's status code, and
   then writes the data as JSON into the body,
   logging (not forwarding) any encoding error. */

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func decodeJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON body: %w", models.ErrInvalidInput)
	}
	return nil
}
