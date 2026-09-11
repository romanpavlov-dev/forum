package models

import "time"

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Created_at   time.Time
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
