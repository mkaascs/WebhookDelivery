# Webhook Service

A reliable webhook delivery service. It accepts events from applications and guarantees delivery to subscribers over HTTP — with retries, exponential backoff, HMAC signing, and a full log of delivery attempts.

In essence, this is what sits behind the webhooks of Stripe, GitHub, or Telegram: the side that dispatches notifications to third-party URLs and takes on all the dirty work of reliable delivery.

## Stack

- **Go** — the API layer is built on the standard `net/http` (using the routing improvements from Go 1.22+), and the delivery workers run as a pool of goroutines. No web framework: the standard library covers routing, middleware, and the outbound HTTP client with proper timeouts.
- **PostgreSQL** — the source of truth. Stores endpoints, subscriptions, events, and the delivery log. Related tables (an event fans out into many deliveries, a delivery accumulates many attempts) are queried with JOINs for status and debugging views.
- **Redis** — the delivery queue between the event receiver and the workers. The receiver enqueues delivery jobs and returns immediately; workers pull jobs and dispatch them in the background.
- **Docker Compose** — brings up the API service, PostgreSQL, and Redis with a single command.

## API

```
POST   /api/v1/endpoints              # register a receiver URL (+ secret for signing)
GET    /api/v1/endpoints              # list endpoints
GET    /api/v1/endpoints/{id}         # details + delivery stats
PATCH  /api/v1/endpoints/{id}         # update URL / pause
DELETE /api/v1/endpoints/{id}

POST   /api/v1/endpoints/{id}/subscriptions   # subscribe to event types
DELETE /api/v1/subscriptions/{id}

POST   /api/v1/events                 # publish an event → fans out to subscribers
GET    /api/v1/events/{id}            # event status

GET    /api/v1/deliveries?event_id=   # delivery attempt history
GET    /api/v1/deliveries/{id}
POST   /api/v1/deliveries/{id}/retry  # manually retry a failed delivery
```

## Running

```bash
docker-compose up --build
```

Brings up the API service, PostgreSQL, and Redis. The API is available at `localhost:8080`.

## Signature verification on the receiver side

Every delivered request carries an `X-Signature` header — an HMAC-SHA256 of the request body computed with the endpoint's secret. The receiver repeats the computation with its own copy of the secret and compares — confirming the event came from `hookrelay` and was not forged.

## Status

Work in progress.