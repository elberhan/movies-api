package main

import (
	"flag"
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/db"
	"gitea.kood.tech/nikolailetshev/movies-api/handler"
)

func main() {
	var initDB bool
	flag.BoolVar(&initDB, "init", false, "wipe the example tables and insert the seed data before serving")
	flag.BoolVar(&initDB, "i", false, "shorthand for -init")
	flag.Parse()

	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if initDB {
		if err := db.Seed(database); err != nil {
			log.Fatal(err)
		}
		log.Println("database seeded with example data")
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, database)

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
