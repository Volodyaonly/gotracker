.PHONY: build run test lint

build:
	go build -o bin/gotracker ./cmd/api

run:
	go run ./cmd/...

test:
	go vet ./...
	go test ./...

lint:
	golangci-lint run