package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/goconnect/api/shared/proto_gen/api/shared/proto"
)

type AuthHandler struct {
	authClient pb.AuthServiceClient
}

func NewAuthHandler(authClient pb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Register(context.Background(), &pb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": resp.Success,
		"message": resp.Message,
		"email":   resp.Email,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Login(context.Background(), &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": resp.Message})
		return
	}

	// Check if requires TOTP or not verified
	if resp.RequiresTotp {
		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"requires_totp": true,
			"message":       "TOTP verification required",
			"user_id":       resp.UserId,
		})
		return
	}

	if !resp.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"success":     false,
			"is_verified": false,
			"message":     resp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
		"user": gin.H{
			"id":       resp.UserId,
			"username": resp.Username,
			"email":    resp.Email,
		},
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.GenerateOTP(context.Background(), &pb.GenerateOTPRequest{
		Email:   req.Email,
		Purpose: "password_reset",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OTP sent to your email",
	})
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email: req.Email,
		Otp:   req.OTP,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
		"user_id": resp.UserId,
	})
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.ResetPassword(context.Background(), &pb.ResetPasswordRequest{
		Email:       req.Email,
		Otp:         req.OTP,
		NewPassword: req.NewPassword,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
	})
}

func (h *AuthHandler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username parameter required"})
		return
	}

	resp, err := h.authClient.CheckUsernameAvailability(context.Background(), &pb.CheckUsernameRequest{
		Username: username,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"available": resp.Available,
	})
}

// New handlers for email verification flow

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.VerifyEmail(context.Background(), &pb.VerifyEmailRequest{
		Email: req.Email,
		Otp:   req.OTP,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   resp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
		"user": gin.H{
			"id":       resp.UserId,
			"username": resp.Username,
			"email":    resp.Email,
		},
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.ResendVerificationOTP(context.Background(), &pb.ResendOTPRequest{
		Email: req.Email,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   resp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
	})
}

func (h *AuthHandler) GetBlockStatus(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter required"})
		return
	}

	resp, err := h.authClient.GetBlockStatus(context.Background(), &pb.GetBlockStatusRequest{
		Email: email,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_blocked":         resp.IsBlocked,
		"remaining_seconds":  resp.RemainingSeconds,
		"remaining_attempts": resp.RemainingAttempts,
	})
}

type SetupTOTPRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	var req SetupTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.SetupTOTP(context.Background(), &pb.SetupTOTPRequest{
		UserId: req.UserID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   resp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      resp.Message,
		"secret":       resp.Secret,
		"qr_code":      resp.QrCode,
		"issuer":       resp.Issuer,
		"account_name": resp.AccountName,
	})
}

type VerifyTOTPRequest struct {
	Username string `json:"username" binding:"required"`
	TOTPCode string `json:"totp_code" binding:"required,len=6"`
}

func (h *AuthHandler) VerifyTOTP(c *gin.Context) {
	var req VerifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.VerifyTOTP(context.Background(), &pb.VerifyTOTPRequest{
		Username: req.Username,
		TotpCode: req.TOTPCode,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   resp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}
