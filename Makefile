.PHONY: db-up db-down run build test db-logs

db-up:
	docker compose up --build -d api

db-down:
	docker compose down -v

run:
	go run ./cmd/api/

build:
	go build -o bin/api ./cmd/api/

test:
	go test -v ./...

db-logs:
	docker compose logs -f api