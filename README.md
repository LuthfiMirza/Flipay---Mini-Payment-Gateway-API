# Flipay - Mini Payment Gateway API

![Go](https://img.shields.io/badge/Go-1.22-blue)
![Gin](https://img.shields.io/badge/Gin-Framework-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-blue)
![Redis](https://img.shields.io/badge/Redis-Queue-red)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED)
![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-85EA2D)
![CI](https://github.com/LuthfiMirza/Flipay---Mini-Payment-Gateway-API/actions/workflows/ci.yml/badge.svg)

Flipay is a **portfolio-grade fintech backend project** that simulates a mini payment gateway API. It demonstrates backend engineering practices commonly used in payment systems: JWT authentication, idempotent payment creation, PostgreSQL transactions, Redis queue processing, goroutine workers, webhook callbacks, structured logging, rate limiting, Swagger documentation, and CI/CD validation.

> Built to showcase clean backend architecture, production-style API design, and DevOps readiness for Golang backend engineer recruitment.

## Demo Preview

| Landing Page | Swagger API Docs |
| --- | --- |
| ![Flipay Landing Page](docs/images/landing-page.svg) | ![Flipay Swagger UI](docs/images/swagger-ui.svg) |

Local demo URLs after running the app:

- Landing page: `http://localhost:8080/`
- Health check: `http://localhost:8080/health`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Why This Project Is Strong for GitHub Portfolio

- **Clean project structure** using handler, service, repository, middleware, worker, queue, and webhook layers.
- **Clear README** with architecture, setup instructions, API examples, CI/CD, deployment notes, and screenshots.
- **Real backend patterns** such as JWT auth, idempotency keys, async processing, webhooks, and rate limiting.
- **DevOps readiness** with Docker, GitHub Actions, health check endpoint, and deployment guide.
- **Fintech relevance** because it models payment gateway flows like VA, QRIS, callbacks, and payment statuses.

## Core Features

### Authentication

- Register endpoint.
- Login endpoint.
- Password hashing with bcrypt.
- JWT token generation.
- JWT middleware for protected routes.

### Payment Gateway

- Create payment.
- Generate virtual account number.
- Generate QRIS simulation string.
- Get payment detail.
- Get payment history.
- Payment statuses: `PENDING`, `SUCCESS`, `FAILED`, `EXPIRED`.
- Idempotency key support to prevent duplicate payment creation.

### Async Processing

- Redis-backed queue simulation.
- Goroutine worker for payment processing.
- Retry mechanism skeleton with exponential backoff.
- Dead-letter queue skeleton for exhausted jobs.

### Webhook Callback

- Webhook sender simulation.
- HMAC signature generation and validation.
- Callback delivery logs.

### Production-Style Engineering

- PostgreSQL transactions.
- Structured JSON logging with Zap.
- Request/correlation ID middleware.
- Redis-based rate limiter.
- Graceful shutdown for `SIGINT` and `SIGTERM`.
- Swagger API documentation.
- GitHub Actions CI pipeline.

## Tech Stack

| Category | Technology |
| --- | --- |
| Language | Go 1.22 |
| Web Framework | Gin |
| Database | PostgreSQL |
| Cache / Queue | Redis |
| Auth | JWT, bcrypt |
| Logging | Zap JSON Logger |
| API Docs | Swagger / OpenAPI |
| DevOps | Docker, Docker Compose, GitHub Actions |
| Migration | SQL migration files |

## Architecture

```mermaid
flowchart TD
    Client[Client / Merchant App] --> API[Gin API]
    API --> Middleware[Middleware Layer]
    Middleware --> JWT[JWT Auth]
    Middleware --> RateLimit[Redis Rate Limiter]
    Middleware --> Logger[Request ID + Zap Logs]
    Middleware --> PaymentService[Payment Service]
    PaymentService --> Postgres[(PostgreSQL)]
    PaymentService --> Redis[(Redis Queue)]
    Redis --> Worker[Goroutine Worker]
    Worker --> Postgres
    Worker --> Webhook[Webhook Sender]
    Webhook --> Merchant[Merchant Callback URL]
```

More detailed architecture notes are available in `architecture.md`.

## Project Structure

```text
cmd/api              Application entrypoint, routes, graceful shutdown
configs              Environment configuration
internal/auth        Auth model, DTO, repository, service, handler
internal/payment     Payment model, DTO, repository, service, handler
internal/middleware  JWT, request logger, request ID, rate limiter
internal/queue       Redis queue abstraction
internal/worker      Async payment worker and expiration loop
internal/webhook     Callback sender and receiver
internal/database    PostgreSQL and Redis clients
internal/utils       Shared response, generators, landing page
migrations           SQL database migrations
docs                 Swagger and deployment documentation
.github/workflows    CI pipeline
```

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/LuthfiMirza/Flipay---Mini-Payment-Gateway-API.git
cd Flipay---Mini-Payment-Gateway-API
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Example environment variables:

```env
APP_ENV=development
APP_PORT=8080
DATABASE_URL=postgres://flipay:flipay@localhost:5432/flipay?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=change-me
WEBHOOK_SECRET=webhook-secret
PAYMENT_EXPIRY_MINUTES=30
```

### 3. Start PostgreSQL and Redis

```bash
docker compose up -d postgres redis
```

### 4. Run Database Migration

```bash
migrate -path migrations -database "postgres://flipay:flipay@localhost:5432/flipay?sslmode=disable" up
```

### 5. Run Application

```bash
go mod tidy
go run ./cmd/api
```

Open:

```text
http://localhost:8080/
```

## Docker Setup

Run the full stack with Docker Compose:

```bash
cp .env.example .env
docker compose up --build
```

Health check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## API Documentation

Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

Swagger notes and regeneration guide are available in `docs/swagger.md`.

## API Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Flip User",
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Use JWT Token

```bash
TOKEN="paste-access-token-here"

curl http://localhost:8080/api/v1/payments/history \
  -H "Authorization: Bearer $TOKEN"
```

### Create Bank Transfer Payment

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: pay-demo-001" \
  -d '{
    "amount": 150000,
    "payment_method": "bank_transfer"
  }'
```

### Create QRIS Payment

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: pay-demo-002" \
  -d '{
    "amount": 75000,
    "payment_method": "qris"
  }'
```

### Get Payment Detail

```bash
curl http://localhost:8080/api/v1/payments/payment-uuid \
  -H "Authorization: Bearer $TOKEN"
```

### Manual Webhook Callback Test

```bash
PAYLOAD='{"payment_id":"payment-uuid","reference_no":"FLP-2026-000001","status":"SUCCESS","amount":150000}'
SIGNATURE=$(printf '%s' "$PAYLOAD" | openssl dgst -sha256 -hmac "webhook-secret" -hex | sed 's/^.* //')

curl -X POST http://localhost:8080/api/v1/callbacks/payment \
  -H "Content-Type: application/json" \
  -H "X-Flipay-Signature: $SIGNATURE" \
  -d "$PAYLOAD"
```

## Testing and Validation

Run local validation before pushing changes:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

Docker Compose config validation:

```bash
docker compose config
```

## CI/CD

GitHub Actions workflow: `.github/workflows/ci.yml`.

The CI pipeline runs automatically on push and pull request to `main`:

1. Checkout repository.
2. Setup Go 1.22.
3. Download dependencies.
4. Validate formatting with `gofmt`.
5. Run `go vet ./...`.
6. Run `go test ./...`.
7. Run `go build ./...`.

Trigger CI by pushing to GitHub:

```bash
git push origin main
```

## Deployment

Deployment guide is available in `docs/deployment.md`.

Supported deployment targets:

- Railway with Dockerfile deployment.
- Render with Docker runtime.

Recommended production health check endpoint:

```http
GET /health
```

## Learning Notes

This project intentionally includes beginner-friendly comments in important backend areas:

- Middleware and request lifecycle.
- PostgreSQL transaction usage.
- Redis queue and retry skeleton.
- Worker processing flow.
- Webhook signature validation.
- Graceful shutdown.

## Future Improvements

- Add unit tests with mocked repositories.
- Add integration tests using testcontainers.
- Replace Redis list queue with Redis Streams.
- Add OpenTelemetry tracing.
- Add refresh token and logout support.
- Add merchant dashboard and callback URL management.
- Add database migration automation in deployment pipeline.
- Add Kubernetes deployment manifests.

## Author

Built by [Luthfi Mirza](https://github.com/LuthfiMirza) as a Golang fintech backend portfolio project.
