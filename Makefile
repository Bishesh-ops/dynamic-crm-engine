.PHONY: db-up db-down run build

db-up:
	docker compose up --build -d api

db-down:
	docker compose down -v

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

db-logs:
	docker compose logs -f api
