package auth

// RegisterRequest is the payload sent by a client when creating a new account.
// Gin validates this struct before the request reaches the service layer.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email,max=160"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest is intentionally small: never accept fields that are not needed.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=160"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// UserResponse hides sensitive fields such as the password hash.
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// AuthResponse is returned by both register and login endpoints.
type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
}
