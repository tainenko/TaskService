# TaskService

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-blue)](https://golang.org/)
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

- Go 1.26+
- PostgreSQL
- Docker & Docker Compose

## API Documentation

Full documentation available at: https://tainenko.github.io/TaskService/

### Authentication

Task endpoints require a JWT: `Authorization: Bearer <token>`. Get one from `POST /auth/register`
or `POST /auth/login` (JSON body `{"email": "...", "password": "..."}`, password 8-72 characters).
Each user only sees and modifies their own tasks; other users' tasks return 404.

Set the signing secret (at least 32 characters) via `TASK_AUTH_JWTSECRET`; the token lifetime is
`Auth.TokenTTLMinutes` (default 60). `config.prod.yaml` has no secret, so the service refuses to start
in prod until it is provided. The local/dev config ships with a development-only secret.

Existing tasks created before authentication have no owner and are not visible through the API;
assign them with `UPDATE task SET user_id = <id> WHERE user_id IS NULL`.

### Tags

Tasks can carry up to 10 tags of at most 50 characters. Tags belong to the user and are
normalized to lowercase and de-duplicated. Send `"tags": [...]` when creating or updating a task:
omitting it on update leaves the tags unchanged, `[]` clears them. Filter with `GET /tasks?tag=work`;
list your tags with usage counts via `GET /tags`; `DELETE /tags/{id}` removes a tag from all tasks.

### Metrics

`GET /metrics` serves Prometheus metrics: `http_requests_total`, `http_request_duration_seconds`
(labelled by method, route template and status, so cardinality stays bounded),
`http_requests_in_flight`, database pool stats (`go_sql_*`) and Go/process metrics.
Set `Server.MetricsToken` (env `TASK_SERVER_METRICSTOKEN`) to require `Authorization: Bearer <token>`;
when empty the endpoint is open, so restrict it at the network level or set a token in production.

### Rate limiting

`/auth/register` and `/auth/login` are protected against brute force, returning `429` with a
`Retry-After` header:

- **Per client IP**: `Auth.IPRatePerMinute` (default 20) with a burst of `Auth.IPBurst` (10).
- **Per account** (login only): after `Auth.LoginMaxFailures` (5) failed logins within
  `Auth.LoginFailureWindowMinutes` (15), that email is throttled from any IP, even with the
  right password. A successful login resets the counter.

Limits are kept in memory, so they apply per instance. Behind a reverse proxy or load balancer,
set `Server.TrustedProxies` (env `TASK_SERVER_TRUSTEDPROXIES`, comma separated IPs/CIDRs);
otherwise every client appears to come from the proxy's IP. It is empty by default, so a spoofed
`X-Forwarded-For` header is ignored.

### Endpoints

| Method | Endpoint    | Description          |
|--------|-------------|----------------------|
| POST   | /auth/register | Register, returns token |
| POST   | /auth/login | Log in, returns token |
| GET    | /tasks      | List your tasks      |
| GET    | /tasks/{id} | Get a task           |
| GET    | /tags       | List your tags       |
| DELETE | /tags/{id}  | Delete a tag         |
| PATCH  | /tasks/{id}/status | Update task status |
| POST   | /tasks      | Create a new task    |
| PUT    | /tasks/{id} | Update existing task |
| DELETE | /tasks/{id} | Delete a task        |

### Task Model

Each task has the following fields:

| Field      | Type              | Description                                     |
|------------|-------------------|-------------------------------------------------|
| id         | SERIAL            | Primary key, auto-incrementing identifier       |
| name       | VARCHAR(255)      | Task name (required)                            |
| status     | INTEGER           | 0 = todo, 1 = done, 2 = in progress (required)  |
| description | TEXT             | Task description, defaults to empty             |
| due_date   | TIMESTAMP WITH TZ | Due date (optional)                             |
| priority   | INTEGER           | 0 = low, 1 = medium, 2 = high                   |
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

Integration tests run the whole HTTP API against a real PostgreSQL container (applying the
migrations in `migration/`). They need Docker and are excluded from the default run:
```bash
make test-integration
```