// Package docs contains Swagger metadata served by gin-swagger.
// This file is intentionally committed so the app compiles even before a developer runs `swag init`.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Returns API health status.",
                "produces": ["application/json"],
                "tags": ["Health"],
                "summary": "Health check",
                "responses": {"200": {"description": "OK"}}
            }
        },
        "/api/v1/auth/register": {
            "post": {
                "description": "Create a new user account and return a JWT access token.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Auth"],
                "summary": "Register user",
                "responses": {"201": {"description": "Created"}, "400": {"description": "Bad Request"}, "409": {"description": "Conflict"}}
            }
        },
        "/api/v1/auth/login": {
            "post": {
                "description": "Validate credentials and return a JWT access token.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Auth"],
                "summary": "Login user",
                "responses": {"200": {"description": "OK"}, "400": {"description": "Bad Request"}, "401": {"description": "Unauthorized"}}
            }
        },
        "/api/v1/payments": {
            "post": {
                "security": [{"BearerAuth": []}],
                "description": "Create a payment and enqueue async processing.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Payments"],
                "summary": "Create payment",
                "parameters": [{"type":"string","description":"Unique key for safe request retry","name":"Idempotency-Key","in":"header"}],
                "responses": {"201": {"description": "Created"}, "400": {"description": "Bad Request"}, "401": {"description": "Unauthorized"}, "409": {"description": "Conflict"}}
            }
        },
        "/api/v1/payments/{id}": {
            "get": {
                "security": [{"BearerAuth": []}],
                "description": "Get a single payment owned by the authenticated user.",
                "produces": ["application/json"],
                "tags": ["Payments"],
                "summary": "Get payment detail",
                "parameters": [{"type":"string","description":"Payment ID","name":"id","in":"path","required":true}],
                "responses": {"200": {"description": "OK"}, "401": {"description": "Unauthorized"}, "404": {"description": "Not Found"}}
            }
        },
        "/api/v1/payments/history": {
            "get": {
                "security": [{"BearerAuth": []}],
                "description": "List payments owned by the authenticated user.",
                "produces": ["application/json"],
                "tags": ["Payments"],
                "summary": "Get payment history",
                "responses": {"200": {"description": "OK"}, "401": {"description": "Unauthorized"}}
            }
        },
        "/api/v1/callbacks/payment": {
            "post": {
                "description": "Receive a simulated merchant callback with HMAC signature validation.",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Callbacks"],
                "summary": "Payment callback",
                "responses": {"200": {"description": "OK"}, "401": {"description": "Unauthorized"}}
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {"type": "apiKey", "name": "Authorization", "in": "header"}
    }
}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Flipay - Mini Payment Gateway API",
	Description:      "Fintech backend simulation with JWT auth, payment processing, Redis worker, and webhook callbacks.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
