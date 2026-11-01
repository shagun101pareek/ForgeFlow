.PHONY: run build tidy fmt vet test sqlc migrate-up migrate-version

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

sqlc:
	sqlc generate

migrate-up:
	go run ./cmd/migrate up

migrate-version:
	go run ./cmd/migrate version
