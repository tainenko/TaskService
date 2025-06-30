# TaskService

[![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue)](https://golang.org/)
[![Documentation](https://img.shields.io/badge/docs-online-brightgreen)](https://tainenko.github.io/TaskService/)

A RESTful Task Management API service built with Go.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Dependencies](#dependencies)
- [API Documentation](#api-documentation)
- [Setup](#setup)
- [Testing](#testing)


## Overview
TaskService is a RESTful API application that provides endpoints for managing tasks. It uses PostgreSQL for data storage
and can be run using Docker.

## Dependencies

- **Gin (v1.9.0+)**: High-performance HTTP web framework
- **GORM (v1.25.0+)**: The fantastic ORM library for Golang
- **Viper (v1.15.0+)**: Complete configuration solution
- **Goose (v3.7.0+)**: Database migration tool
- **PostgreSQL Driver (v1.5.0+)**: PostgreSQL driver for Go's database/sql package

## Requirements

- Go 1.18+
- PostgreSQL
- Docker & Docker Compose

## API Documentation

Full documentation available at: https://tainenko.github.io/TaskService/

### Endpoints

| Method | Endpoint    | Description          |
|--------|-------------|----------------------|
| GET    | /tasks      | List all tasks       |
| POST   | /tasks      | Create a new task    |
| PUT    | /tasks/{id} | Update existing task |
| DELETE | /tasks/{id} | Delete a task        |

### Task Model

Each task has the following fields:

| Field      | Type              | Description                                     |
|------------|-------------------|-------------------------------------------------|
| id         | SERIAL            | Primary key, auto-incrementing identifier       |
| name       | VARCHAR(255)      | Task name (required)                            |
| status     | INTEGER           | Task status code (required)                     |
| created_at | TIMESTAMP WITH TZ | Creation timestamp, defaults to current time    |
| updated_at | TIMESTAMP WITH TZ | Last update timestamp, defaults to current time |
| deleted_at | TIMESTAMP WITH TZ | Soft deletion timestamp (optional)              |

## Setup

There are two ways to run the TaskService: directly with Go or using Docker Compose.

### Local Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/tainenko/TaskService.git
   cd TaskService
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up PostgreSQL database and run migrations:
   ```bash
   psql -U postgres -f build/database/psql_dump.sql
   ```

4. Run the server:
   ```bash
   go run cmd/main.go
   ```

The server will start at `http://localhost:8080`

### Docker Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/tainenko/TaskService.git
   cd TaskService
   ```

2. Build and start the containers:
   ```bash
   docker-compose up -d
   ```

The server will be available at `http://localhost:8080`

To stop the containers:

```bash
docker-compose down
```

## Testing

Run all tests with verbose output using:
```bash
go test ./...
```