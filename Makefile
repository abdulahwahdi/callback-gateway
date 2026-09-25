.PHONY: run build tidy fmt test migration docker-up docker-down

migration:
	@go run cmd/migration/migration.go up

run:
	go run .

build:
	go build -o bin/webhook-middleware .

tidy:
	go mod tidy

fmt:
	gofmt -w .

test:
	go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
