package domain

import "time"

type Student struct {
	ID           int64     `json:"id"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}