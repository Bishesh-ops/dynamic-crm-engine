.PHONY: db-up db-down run build

db-up:
	docker compose up -d

db-down:
	docker compose down

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go
