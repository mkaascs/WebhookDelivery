# Webhook Service

![CI](https://github.com/mkaascs/WebhookDelivery/actions/workflows/ci.yaml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)

A reliable webhook delivery service. It accepts events from applications and guarantees delivery to subscribers over HTTP — with retries, exponential backoff, HMAC signing, and a full log of delivery attempts.

In essence, this is what sits behind the webhooks of Stripe, GitHub, or Telegram: the side that dispatches notifications to third-party URLs and takes on all the dirty work of reliable delivery.

## Stack

| Technology | Role |
| --- | --- |
| **Go** | The API layer is built on the standard `net/http` (using the routing improvements from Go 1.22+), and the delivery workers run as a pool of goroutines. No web framework: the standard library covers routing, middleware, and the outbound HTTP client with proper timeouts. |
| **PostgreSQL** | The source of truth. Stores endpoints, subscriptions, events, and the delivery log. Related tables (an event fans out into many deliveries, a delivery accumulates many attempts) are queried with JOINs for status and debugging views. |
| **Redis** | The delivery queue between the event receiver and the workers. The receiver enqueues delivery jobs and returns immediately; workers pull jobs and dispatch them in the background. |
| **Docker Compose** | Brings up the API service, PostgreSQL, and Redis with a single command. |

## API

### Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/endpoints` | Register a receiver URL (+ secret for signing) |
| `GET` | `/api/v1/endpoints` | List endpoints |
| `GET` | `/api/v1/endpoints/{id}` | Details + delivery stats |
| `PATCH` | `/api/v1/endpoints/{id}` | Update URL / pause |
| `DELETE` | `/api/v1/endpoints/{id}` | Delete an endpoint |

### Subscriptions

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/endpoints/{id}/subscriptions` | Subscribe to event types |
| `DELETE` | `/api/v1/subscriptions/{id}` | Remove a subscription |

### Events

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/v1/events` | Publish an event → fans out to subscribers |
| `GET` | `/api/v1/events/{id}` | Event status |

### Deliveries

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/deliveries?event_id=` | Delivery attempt history |
| `GET` | `/api/v1/deliveries/{id}` | Single delivery details |
| `POST` | `/api/v1/deliveries/{id}/retry` | Manually retry a failed delivery |

## Running

```bash
docker-compose up --build
```

Brings up the API service, PostgreSQL, and Redis. The API is available at `localhost:8080`.

## Signature verification on the receiver side

Every delivered request carries an `X-Signature` header — an HMAC-SHA256 of the request body computed with the endpoint's secret. The receiver repeats the computation with its own copy of the secret and compares — confirming the event came from `hookrelay` and was not forged.

## Status

Work in progress.