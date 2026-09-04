package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/models"
)

// writeJSON puts the JSON label on the response,
// stamps it with the caller's status code, and
// then writes the data as JSON into the body,
// logging (not forwarding) any encoding error.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json response: %v", err)
	}
}

// decodeJSON reads a request body into out (any struct pointer).
func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // reject unknown fields instead of silently dropping them
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("%w: invalid JSON body: %v", models.ErrInvalidInput, err)
	}
	return nil
}
