package payment

import "errors"

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrIdempotencyConflict  = errors.New("idempotency key was already used with a different request")
	ErrInvalidPaymentStatus = errors.New("invalid payment status transition")
)
