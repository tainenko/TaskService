# TaskService

[![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue)](https://golang.org/)
[![Documentation](https://img.shields.io/badge/docs-online-brightgreen)](https://tainenko.github.io/TaskService/)

A RESTful Task Management API service built with Go.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [API Documentation](#api-documentation)
- [Setup](#setup)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)

## Overview

TaskService is a RESTful API application that provides endpoints for managing tasks. It uses PostgreSQL for data storage
and can be run using Docker.

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
