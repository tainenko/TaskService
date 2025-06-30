.PHONY: build test clean docker-build docker-up docker-down db-setup migrate-create migrate-up migrate-down migrate-status

# Build the application
build:
	go build -o bin/taskservice main.go

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Database setup
db-setup:
	psql -U postgres -f build/database/psql_dump.sql

# Database migrations
migrate-create:
	goose -dir migrations create $(name) sql

migrate-up:
	goose -dir migration up

migrate-down:
	goose -dir migration down

migrate-status:
	goose -dir migration status

# Install dependencies
deps:
	go mod download

# Default target
all: deps build
