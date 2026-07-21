# Webhook Delivery

![CI](https://github.com/mkaascs/WebhookDelivery/actions/workflows/ci.yaml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)

A reliable webhook delivery service. Applications register receiver endpoints, subscribe them to event types, and publish events — the service fans each event out to its subscribers and guarantees delivery over HTTP with retries, exponential backoff, and HMAC-signed requests. Every delivery attempt is tracked with its own status, retry schedule, and last response.

This is the machinery behind the webhooks of services like Stripe or GitHub: the side that dispatches notifications to third-party URLs and takes on the dirty work of reliable delivery.

---

## Features

- **Endpoint registration** — register a receiver URL; a per-endpoint signing secret (`whsec_…`) is generated automatically
- **Event subscriptions** — subscribe an endpoint to specific event types; subscribing is idempotent (`ON CONFLICT DO NOTHING`)
- **Event fan-out** — publishing an event creates one pending delivery per subscribed endpoint
- **Reliable delivery** — a background worker pool delivers over HTTP with retries, exponential backoff with jitter, and a max-attempt cap; each delivery tracks its status (`pending` / `delivered` / `failed`), attempt count, last response code and last error
- **HMAC request signing** — every delivered request carries an `X-Webhook-Signature` header (HMAC-SHA256 of the body)
- **SSRF protection** — endpoint URLs are validated on registration; loopback, link-local and `localhost` targets are rejected
- **Request validation** — incoming bodies are parsed and validated by generic middleware before reaching a handler
- **Postgres-backed queue** — no external broker; pending deliveries are claimed straight from Postgres via a partial index
- **OpenAPI docs** — handlers are annotated with swaggo; the spec is generated into `docs/`
- **Graceful shutdown** — clean drain of workers and the HTTP server on `SIGINT` / `SIGTERM`

---

## Tech Stack

| Component | Technology |
|----------------------|-------------------------------------------------|
| **Language** | Go 1.26+ |
| **HTTP Router** | [chi](https://github.com/go-chi/chi) |
| **Database / Queue** | PostgreSQL — native [pgx v5 / pgxpool](https://github.com/jackc/pgx) |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) |
| **Validation** | [go-playground/validator](https://github.com/go-playground/validator) |
| **JSON responses** | [go-chi/render](https://github.com/go-chi/render) |
| **API docs** | [swaggo/swag](https://github.com/swaggo/swag) (OpenAPI 2.0) |
| **Config** | [cleanenv](https://github.com/ilyakaznacheev/cleanenv) + godotenv (YAML + env) |
| **Logging** | slog (structured JSON / text) |
| **Testing** | testify + [gomock](https://github.com/golang/mock) |
| **Containerization** | Docker + Docker Compose |

---

## Architecture

Three layers with dependencies pointing inward, wired through interfaces:

- **`delivery`** — HTTP handlers (chi) plus generic middleware: body parsing (`BodyParser[T]`), validation (`Validator[T]`), pagination and request logging
- **`services`** — business logic; each service depends on repository interfaces, not concrete drivers
- **`infrastructure/pg`** — pgxpool repositories (one file per aggregate), sharing a small `poolQuery` seam so methods run identically on the pool or inside a transaction
- **`domain`** — entities, DTOs and error values, free of framework code

The delivery worker pool lives in `services/workers` and reaches Postgres through the same repository seam.

---

## API

All routes are mounted under `/api/v1`. Handlers are documented with swaggo annotations; run `task docs:swagger` to (re)generate the OpenAPI spec into `docs/` (`swagger.json` / `swagger.yaml`). Delivery dispatch is an internal pipeline and has no public endpoint.

### Endpoints

| Method | Path | Description |
|----------|------------------------------|--------------------------------------|
| `POST` | `/api/v1/endpoints` | Register a receiver URL (returns its signing secret) |
| `GET` | `/api/v1/endpoints` | List endpoints (paginated) |
| `GET` | `/api/v1/endpoints/{id}` | Get a single endpoint |
| `PATCH` | `/api/v1/endpoints/{id}` | Update URL / active state / description |
| `DELETE` | `/api/v1/endpoints/{id}` | Delete an endpoint |

### Subscriptions

| Method | Path | Description |
|----------|-------------------------------------------|-----------------------------------|
| `POST` | `/api/v1/endpoints/{id}/subscriptions` | Subscribe an endpoint to event types |
| `GET` | `/api/v1/endpoints/{id}/subscriptions` | List an endpoint's subscriptions |
| `DELETE` | `/api/v1/subscriptions/{id}` | Remove a subscription |

### Events

| Method | Path | Description |
|--------|--------------------------|----------------------------------------------|
| `POST` | `/api/v1/events` | Publish an event → fans out to subscribers |
| `GET` | `/api/v1/events/{id}` | Get an event by id |

### Deliveries

| Method | Path                                | Description                                        |
|--------|-------------------------------------|----------------------------------------------------|
| `POST` | `/api/v1/deliveries/{id}/retry`     | Retry delivery even endpoint is inactive right now |
| `GET`  | `/api/v1/deliveries/{id}` | Get a single delivery                              |
| `GET`  | `/api/v1/events/{id}/deliveries`    | Deliveries list of current event                   |

---

## Delivery Pipeline

Publishing an event fans it out into one `pending` delivery row per subscribed endpoint. A pool of worker goroutines then drives those rows to completion:

1. **Claim** — each worker claims a batch of due `pending` deliveries from Postgres (a partial index on `next_retry_at WHERE status = 'pending'` keeps this cheap).
2. **Deliver** — the payload is `POST`ed to the endpoint URL with an `X-Webhook-Signature` header. Any `2xx` marks the delivery `delivered`.
3. **Reschedule** — on a non-`2xx` response the delivery stays `pending` and `next_retry_at` is pushed back by an exponential backoff with jitter; once the attempt count reaches `max_attempts` it becomes `failed`. The last response code and error are recorded on the row either way.

Workers wake on a ticker and on an in-process notify signal, so freshly published events start delivering without waiting for the next tick.

```yaml
workers:
  max_goroutines: 5
  batch_size: 10
  max_attempts: 5
  ticker_duration: 30s
  timeout: 10s
  base_backoff: 10s
  max_backoff: 1h
```

---

## Signature Verification (receiver side)

Every delivered request carries an `X-Webhook-Signature` header — an HMAC-SHA256 of the raw request body, computed with the endpoint's secret. The receiver repeats the computation with its own copy of the secret and compares the two, confirming the request came from this service and was not forged or tampered with.

---

## CI

Every push and pull request to `master` / `development` runs the **CI** workflow:

- **test** — `go build`, `go vet`, and `go test -race` across all packages
- **lint** — `golangci-lint`

---

## Quick Start

### Prerequisites

- Docker + Docker Compose
- (for local runs) Go 1.26+, PostgreSQL and [Task](https://taskfile.dev)

### Environment variables

| Variable | Description | Required |
|---------------|-------------------------------------|----------|
| `CONFIG_PATH` | Path to the YAML config file | Yes |
| `DB_USERNAME` | Postgres user | Yes |
| `DB_PASSWORD` | Postgres password | Yes |

Provide them via a `.env` file (loaded automatically) or the environment. For Docker use `config/dev.yaml` (its `db.addr` points at the `postgres` service); for host runs use `config/local.yaml` (`localhost`).

### Run with Docker

```bash
docker-compose up --build
```

Brings up PostgreSQL and the service; migrations are applied automatically on startup. The API is exposed on `localhost:5428`.

### Run locally

```bash
task migrate:up               # apply database migrations
go run ./cmd/webhook-delivery # start the service (serves on :8080)
```

Other handy tasks: `task test` / `task test:cover` (coverage), `task docs:swagger` (regenerate OpenAPI), `task mocks` (regenerate mocks), `task migrate:down`.

### Config file (`config/local.yaml`)

```yaml
env: "local" # local, dev, prod
http_server:
  port: 8080
  write_timeout: 5s
  read_timeout: 10s
db:
  addr: "localhost:5432"
  connection_timeout: 5s
workers:
  max_goroutines: 5
  batch_size: 5
  max_attempts: 5
  ticker_duration: 10s
  timeout: 10s
  base_backoff: 10s
  max_backoff: 1h
```

---

## Security

- Each endpoint gets a signing secret (`whsec_…`, 32 random bytes from `crypto/rand`); payloads are signed with HMAC-SHA256 so receivers can verify authenticity
- Endpoint URLs are SSRF-checked on registration — loopback, link-local and `localhost` targets are rejected
- Postgres is accessed through a pooled connection with parameterized queries throughout
- Credentials (DB user / password) are supplied via environment, never committed to the config file
