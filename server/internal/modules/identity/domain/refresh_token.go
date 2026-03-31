package domain

import "time"

type RefreshToken struct {
	ID         int
	UserID     int
	TokenHash  string
	ExpiresAt  time.Time
	Revoked    bool
	CreatedAt  time.Time
	CreateUser string
	UpdatedAt  time.Time
	UpdateUser string
}
