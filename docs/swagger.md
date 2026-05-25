# Swagger Documentation

Flipay exposes interactive API documentation at:

```http
GET /swagger/index.html
```

## How It Works

- `swaggo/swag` reads comment annotations from Go handlers.
- `gin-swagger` serves the Swagger UI inside the Gin application.
- The generated `docs` package stores OpenAPI metadata used by the Swagger route.

## Regenerate Docs

Install the CLI:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Generate OpenAPI files:

```bash
swag init -g cmd/api/main.go
```

Then run the API and open:

```text
http://localhost:8080/swagger/index.html
```

## Screenshot Placeholder

Add a Swagger UI screenshot here after running the app locally:

```text
docs/images/swagger-ui.png
```
