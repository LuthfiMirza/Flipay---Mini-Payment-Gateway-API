package auth

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredential  = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)
