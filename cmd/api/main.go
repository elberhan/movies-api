package main

import (
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/db"
	"gitea.kood.tech/nikolailetshev/movies-api/handler"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, database)

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
