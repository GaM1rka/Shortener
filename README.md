# Shortener

URL shortener service with redirect analytics. The backend is written in Go and exposes a small HTTP API for creating short links, redirecting visitors, and reading click analytics.

## Features

- `POST /shorten` creates a short link.
- `GET /s/{short_url}` redirects to the original URL and records a click.
- `GET /analytics/{short_url}` returns click count, raw clicks, and aggregations by day, month, and User-Agent.
- Supports custom aliases with letters, numbers, `_`, and `-`.
- Stores links and clicks in Postgres when `DATABASE_URL` is set.
- Falls back to in-memory storage for quick local runs without infrastructure.
- Serves a simple browser UI for creating links and viewing analytics.
- Docker Compose includes backend, Postgres, and Redis.

## Project Structure

```text
backend/
  cmd/shortener/          application entrypoint
  internal/app/           wiring
  internal/config/        environment config
  internal/domain/        response/domain models
  internal/httpapi/       HTTP handlers and router
  internal/migrations/    migration runner
  internal/service/       business logic
  internal/storage/       memory and postgres stores
  migrations/             SQL migrations
docs/
  openapi.yaml            OpenAPI 3.0 spec
frontend/
  index.html              browser UI
  app.js
  styles.css
docker-compose.yml
```

## Run With Docker Compose

```bash
docker compose up --build
```

The service listens on `http://localhost:8080`.

Postgres migrations are applied automatically by the backend container on startup.

Open the UI at `http://localhost:8080`.

## Run Backend Locally

Without Postgres, the service uses in-memory storage:

```bash
cd backend
go run ./cmd/shortener
```

With Postgres:

```bash
cd backend
DATABASE_URL="postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable" \
PUBLIC_BASE_URL="http://localhost:8080" \
go run ./cmd/shortener
```

## Environment

| Variable | Default | Description |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Backend bind address |
| `PUBLIC_BASE_URL` | `http://localhost:8080` | Base URL used in generated short links |
| `SHORT_CODE_LENGTH` | `7` | Generated code length |
| `DATABASE_URL` | empty | Postgres connection string |
| `REDIS_URL` | empty | Reserved for Redis caching |
| `MIGRATIONS_DIR` | `./migrations` | Directory with `*.up.sql` files |
| `FRONTEND_DIR` | `../frontend` | Directory with static UI files |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |

## API Examples

Create a link with generated code:

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/landing?page=summer"}'
```

Create a link with custom alias:

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/docs","custom_alias":"docs"}'
```

Follow a short link:

```bash
curl -i http://localhost:8080/s/docs
```

Read analytics:

```bash
curl http://localhost:8080/analytics/docs
```

## Swagger

OpenAPI spec is available at:

```text
docs/openapi.yaml
```

You can open it in Swagger Editor or import it into Postman/Insomnia.

## Tests

```bash
cd backend
GOCACHE=/tmp/shortener-go-cache go test ./...
```
