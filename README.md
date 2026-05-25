# Flipay - Mini Payment Gateway API

Backend fintech simulation built with Golang, Gin, PostgreSQL, Redis queue worker, JWT authentication, idempotency, and webhook callback flow.

## Current Focus: JWT Authentication

Implemented auth endpoints:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Auth module includes:

- User model and DTOs
- Request validation using Gin binding tags
- Password hashing using bcrypt
- JWT access token generation
- JWT middleware for protected routes
- PostgreSQL repository with transaction on register
- Clean JSON success and error responses
- User migration SQL in `migrations/000001_init.up.sql`

## Tech Stack

- Go + Gin
- PostgreSQL + pgx
- Redis
- JWT
- bcrypt
- Zap logger
- Docker Compose
- golang-migrate compatible SQL migrations

## Quick Start

```bash
cp .env.example .env
docker compose up -d postgres redis
migrate -path migrations -database "postgres://flipay:flipay@localhost:5432/flipay?sslmode=disable" up
go mod tidy
go run ./cmd/api
```

If using Docker for the API too:

```bash
cp .env.example .env
docker compose up --build
```

## Auth API

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

Example response:

```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 86399,
    "user": {
      "id": "uuid",
      "name": "Flip User",
      "email": "user@example.com",
      "created_at": "2026-05-26T10:00:00Z"
    }
  }
}
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

## Sample JWT Usage

Copy the `access_token` from register/login response, then call protected endpoints:

```bash
TOKEN="paste-access-token-here"

curl http://localhost:8080/api/v1/payments/history \
  -H "Authorization: Bearer $TOKEN"
```

Create payment with JWT and idempotency key:

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: demo-key-001" \
  -d '{
    "amount": 150000,
    "payment_method": "bank_transfer"
  }'
```

## Endpoint List

### Health

```http
GET /health
```

### Auth

```http
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Payments

Protected by `Authorization: Bearer <token>`.

```http
POST /api/v1/payments
GET /api/v1/payments/:id
GET /api/v1/payments/history
```

### Callback

```http
POST /api/v1/callbacks/payment
```

Callbacks must include `X-Flipay-Signature`, generated with HMAC-SHA256 over the raw JSON payload using `WEBHOOK_SECRET`.

## Project Structure

```text
cmd/api              Application entrypoint and route wiring
configs              Environment configuration
internal/auth        Auth model, DTO, repository, service, handler
internal/payment     Payment model, DTO, repository, service, handler
internal/queue       Redis queue abstraction
internal/worker      Async worker and expiration loop
internal/webhook     Callback sender and receiver
internal/middleware  JWT middleware
internal/database    PostgreSQL and Redis clients
migrations           SQL database migrations
```

## Payment Gateway API

Payment endpoints are protected with JWT. Login first, then use the token as `Authorization: Bearer <token>`.

### Create Bank Transfer Payment

```bash
TOKEN="paste-access-token-here"

curl -X POST http://localhost:8080/api/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: pay-demo-001" \
  -d '{
    "amount": 150000,
    "payment_method": "bank_transfer"
  }'
```

The response includes a simulated virtual account number:

```json
{
  "success": true,
  "message": "payment created",
  "data": {
    "payment": {
      "id": "payment-uuid",
      "reference_no": "FLP-2026-000001",
      "amount": 150000,
      "payment_method": "bank_transfer",
      "va_number": "8808123456789",
      "status": "PENDING"
    },
    "idempotent": false,
    "queued": true,
    "status_check": "/api/v1/payments/payment-uuid"
  }
}
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

### Replay Idempotent Request

Send the same body with the same `Idempotency-Key` to safely replay a payment creation request:

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

If the same key is reused with a different body, the API returns `409 Conflict`.

### Get Payment Detail

```bash
curl http://localhost:8080/api/v1/payments/payment-uuid \
  -H "Authorization: Bearer $TOKEN"
```

### Get Payment History

```bash
curl http://localhost:8080/api/v1/payments/history \
  -H "Authorization: Bearer $TOKEN"
```

### Webhook Callback Simulation

The worker updates payment status asynchronously and sends a callback to `WEBHOOK_URL` with `X-Flipay-Signature`.

Manual callback test example:

```bash
PAYLOAD='{"payment_id":"payment-uuid","reference_no":"FLP-2026-000001","status":"SUCCESS","amount":150000}'
SIGNATURE=$(printf '%s' "$PAYLOAD" | openssl dgst -sha256 -hmac "webhook-secret" -hex | sed 's/^.* //')

curl -X POST http://localhost:8080/api/v1/callbacks/payment \
  -H "Content-Type: application/json" \
  -H "X-Flipay-Signature: $SIGNATURE" \
  -d "$PAYLOAD"
```

## Payment Processing Flow

1. Client creates a payment with JWT and optional `Idempotency-Key`.
2. API validates request and stores `PENDING` payment in PostgreSQL transaction.
3. API pushes a compact job to Redis queue.
4. Goroutine worker pops the job and simulates external payment processing.
5. Worker updates status to `SUCCESS` or `FAILED`.
6. Webhook sender posts callback payload to merchant callback URL.
7. Expiration loop marks old `PENDING` payments as `EXPIRED`.

## Payment Statuses

- `PENDING`: payment created and waiting to be processed.
- `SUCCESS`: payment simulation succeeded.
- `FAILED`: payment simulation failed.
- `EXPIRED`: payment was not completed before `expired_at`.
