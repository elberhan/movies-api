package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gitea.kood.tech/nikolailetshev/movies-api/db"
	"gitea.kood.tech/nikolailetshev/movies-api/handler"
)

func main() {
	var initDB bool
	flag.BoolVar(&initDB, "init", false, "wipe the example tables and insert the seed data before serving")
	flag.BoolVar(&initDB, "i", false, "shorthand for -init")

	port := flag.Int("port", 8080, "TCP port the server listens on")

	flag.Parse()

	if *port < 1025 || *port > 65535 {
		log.Fatalf("invalid port %d: must be between 1025 and 65535", *port)
	}

	addr := fmt.Sprintf(":%d", *port)

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

	fmt.Printf("server listening on http://localhost:%d/\n\n", *port)
	log.Printf("started")
	log.Fatal(http.ListenAndServe(addr, mux))
}
