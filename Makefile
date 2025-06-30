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

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker containers"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  db-setup       - Set up PostgreSQL database"
	@echo "  migrate-create - Create a new migration file"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  migrate-status - Show migration status"
	@echo "  deps           - Install dependencies"
	@echo "  all            - Install dependencies and build"
