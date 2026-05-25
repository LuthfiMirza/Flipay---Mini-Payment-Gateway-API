# Deployment Guide

This guide explains how to deploy Flipay with Docker on Railway or Render.

## Required Environment Variables

```env
APP_ENV=production
APP_PORT=8080
DATABASE_URL=postgres://user:password@host:5432/database?sslmode=require
REDIS_ADDR=host:6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0
JWT_SECRET=use-a-long-random-secret
JWT_TTL_HOURS=24
WEBHOOK_URL=https://your-merchant.example.com/callbacks/payment
WEBHOOK_SECRET=use-a-long-random-webhook-secret
PAYMENT_EXPIRY_MINUTES=30
```

## Health Check

Use this endpoint for platform health checks:

```http
GET /health
```

Expected response:

```json
{"status":"ok"}
```

## Railway Deployment

1. Create a new Railway project.
2. Add PostgreSQL service.
3. Add Redis service.
4. Create a service from the GitHub repository.
5. Select Dockerfile deployment.
6. Add all environment variables from this guide.
7. Set the health check path to `/health`.
8. Deploy the service.
9. Run migrations from your local machine or a Railway job:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

## Render Deployment

1. Create a new Web Service on Render.
2. Connect the GitHub repository.
3. Select Docker runtime.
4. Add PostgreSQL and Redis instances.
5. Add environment variables from this guide.
6. Set health check path to `/health`.
7. Deploy the service.
8. Run database migrations before sending production traffic.

## Production Notes

- Always use strong `JWT_SECRET` and `WEBHOOK_SECRET` values.
- Use managed PostgreSQL and Redis with backups enabled.
- Keep `APP_ENV=production`.
- Restrict database and Redis network access where the platform supports it.
- Run migrations during deployment, not from application startup.
