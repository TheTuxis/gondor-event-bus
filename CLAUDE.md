# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview
Event Bus microservice for Gondor platform. Go/Gin service providing NATS JetStream-based message brokering with dead letter queue, webhook management, and event logging.

## Commands
- `make build` -- compile to bin/server
- `make run` -- run locally (needs PostgreSQL + Redis + NATS)
- `make test` -- run all tests with race detector
- `make lint` -- golangci-lint
- `make docker` -- build Docker image
- `make migrate-up` -- run database migrations
- `make migrate-down` -- rollback migrations

## Architecture
- `cmd/server/main.go` -- entry point, dependency injection, route registration
- `internal/config/` -- env-based configuration
- `internal/model/` -- GORM domain models (Webhook, DeadLetterMessage, EventLog)
- `internal/repository/` -- database access layer
- `internal/service/` -- business logic
- `internal/handler/` -- HTTP handlers (Gin)
- `internal/middleware/` -- JWT auth (validate-only), logging
- `internal/pkg/jwt/` -- JWT validation (tokens issued by gondor-users-security)
- `internal/nats/` -- NATS JetStream client wrapper (connect, publish, subscribe)

## Key Decisions
- JWT tokens are validated only (issued by gondor-users-security service)
- Port 8008
- Database: gondor_event_bus (PostgreSQL, database-per-service)
- NATS JetStream for event pub/sub with stream GONDOR_EVENTS
- Subject pattern: gondor.events.{event_type}
- Webhook secrets auto-generated if not provided
- DLQ statuses: pending, retrying, exhausted
- All routes under /v1/events/ prefix
- /health and /metrics skip JWT auth

## Database
PostgreSQL with GORM. Tables: webhooks, dead_letter_messages, event_logs.

## Environment Variables
- `PORT` (default: 8008)
- `DATABASE_URL` (default: postgres://gondor:gondor_dev@localhost:5432/gondor_event_bus?sslmode=disable)
- `JWT_SECRET` (default: change-me-in-production)
- `REDIS_URL` (default: localhost:6379)
- `NATS_URL` (default: nats://localhost:4222)
- `LOG_LEVEL` (default: info)
- `ENVIRONMENT` (default: development)
