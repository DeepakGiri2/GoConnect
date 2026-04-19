package models

import (
	"time"
)

type User struct {
	ID               string     `json:"id" db:"id"`
	Username         string     `json:"username" db:"username"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	IsVerified       bool       `json:"is_verified" db:"is_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	TOTPSecret       string     `json:"-" db:"totp_secret"`
	TOTPEnabled      bool       `json:"totp_enabled" db:"totp_enabled"`
	TOTPVerifiedAt   *time.Time `json:"totp_verified_at,omitempty" db:"totp_verified_at"`
}

type UnverifiedUser struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
}

type OAuthAccount struct {
	ID             int       `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Provider       string    `json:"provider" db:"provider"`
	ProviderUserID string    `json:"provider_user_id" db:"provider_user_id"`
	AccessToken    string    `json:"-" db:"access_token"`
	RefreshToken   string    `json:"-" db:"refresh_token"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
}
