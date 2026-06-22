package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kamau/speed/internal/api"
	"github.com/kamau/speed/internal/db"
)

func main() {
	godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations(databaseURL)
		return
	}

	pool, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.NewQueries(pool)
	router := api.NewRouter(queries)

	log.Printf("Starting server on :%s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runMigrations(databaseURL string) {
	log.Println("Running migrations...")
	// Migrations are handled via golang-migrate CLI or can be embedded.
	// For now, use: migrate -path migrations -database $DATABASE_URL up
	log.Println("Use: migrate -path migrations -database \"" + databaseURL + "\" up")
}
