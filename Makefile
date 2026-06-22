.PHONY: dev db-up db-down migrate seed build run scrape test

dev: db-up migrate seed run

db-up:
	docker compose up -d
	@echo "Waiting for PostgreSQL..."
	@sleep 3

db-down:
	docker compose down

migrate:
	migrate -path migrations -database "$$DATABASE_URL" up

seed:
	go run cmd/scraper/main.go --seed

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/scraper cmd/scraper/main.go

run:
	go run cmd/api/main.go serve

scrape:
	go run cmd/scraper/main.go --scrape --output data/scraped.json

test:
	go test ./...
