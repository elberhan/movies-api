package handler

import (
	"database/sql"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/repository"
	"gitea.kood.tech/nikolailetshev/movies-api/service"
)

// RegisterRoutes wires all entity handlers onto the mux.
func RegisterRoutes(mux *http.ServeMux, db *sql.DB) {
	// Genre
	genreRepo := repository.NewGenreRepository(db)
	genreService := service.NewGenreService(genreRepo)
	genreHandler := NewGenreHandler(genreService)
	genreHandler.RegisterGenreRoutes(mux)

	// Actor
	actorRepo := repository.NewActorRepository(db)
	actorService := service.NewActorService(actorRepo)
	actorHandler := NewActorHandler(actorService)
	actorHandler.RegisterActorRoutes(mux)

	// TODO: Movie
}
