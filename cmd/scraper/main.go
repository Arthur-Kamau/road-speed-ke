package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kamau/speed/internal/db"
	"github.com/kamau/speed/internal/scraper"
)

func main() {
	seed := flag.Bool("seed", false, "Load seed data from data/geojson/ into the database")
	scrape := flag.Bool("scrape", false, "Scrape Kenya Law for speed limit data")
	output := flag.String("output", "", "Output scraped data to file instead of database")
	flag.Parse()

	godotenv.Load()

	if *scrape {
		results, err := scraper.ScrapeKenyaLaw()
		if err != nil {
			log.Fatalf("Scraping failed: %v", err)
		}

		if *output != "" {
			data, _ := json.MarshalIndent(results, "", "  ")
			if err := os.WriteFile(*output, data, 0644); err != nil {
				log.Fatalf("Failed to write output: %v", err)
			}
			log.Printf("Wrote %d segments to %s", len(results), *output)
			return
		}

		fmt.Printf("Scraped %d segments (use --output to save)\n", len(results))
		return
	}

	if *seed {
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			log.Fatal("DATABASE_URL required for seeding")
		}

		pool, err := db.Connect(databaseURL)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer pool.Close()

		if err := scraper.SeedFromGeoJSON(context.Background(), pool, "data/geojson"); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}

		log.Println("Seed data loaded successfully")
		return
	}

	flag.Usage()
}
