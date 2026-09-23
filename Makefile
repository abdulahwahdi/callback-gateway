.PHONY: run build tidy fmt test docker-up docker-down

run:
	go run ./cmd/webhook-middleware

build:
	go build -o bin/webhook-middleware ./cmd/webhook-middleware

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
