package models

import "github.com/goconnect/pkg/utils"

type RegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Email   string `json:"email"`
}

type EmailVerificationResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	User    *User            `json:"user"`
	Tokens  *utils.TokenPair `json:"tokens"`
}

type LoginResponse struct {
	Success        bool             `json:"success"`
	Message        string           `json:"message,omitempty"`
	User           *User            `json:"user,omitempty"`
	Tokens         *utils.TokenPair `json:"tokens,omitempty"`
	RequiresTOTP   bool             `json:"requires_totp"`
}
