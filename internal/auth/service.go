package auth

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service contains auth business rules and keeps handlers thin.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (AuthResponse, error)
}

type service struct {
	repo       Repository
	jwtSecret  string
	ttlInHours int
}

func NewService(repo Repository, jwtSecret string, ttlInHours int) Service {
	return &service{repo: repo, jwtSecret: jwtSecret, ttlInHours: ttlInHours}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user := User{
		ID:           uuid.NewString(),
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(passwordHash),
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		return AuthResponse{}, err
	}
	return s.authResponse(user)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return AuthResponse{}, ErrInvalidCredential
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return AuthResponse{}, ErrInvalidCredential
	}
	return s.authResponse(user)
}

func (s *service) authResponse(user User) (AuthResponse, error) {
	token, expiresIn, err := s.generateToken(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{AccessToken: token, TokenType: "Bearer", ExpiresIn: expiresIn, User: NewUserResponse(user)}, nil
}

func (s *service) generateToken(userID string) (string, int, error) {
	expiresAt := time.Now().Add(time.Duration(s.ttlInHours) * time.Hour)
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
	return token, int(time.Until(expiresAt).Seconds()), err
}
