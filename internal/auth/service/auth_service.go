package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goconnect/internal/auth/repository"
	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/crypto"
	"github.com/goconnect/pkg/models"
	"github.com/goconnect/pkg/notification"
	"github.com/goconnect/pkg/utils"
)

type AuthService struct {
	userRepo        *repository.UserRepository
	tokenRepo       *repository.TokenRepository
	bloomFilter     *BloomFilterService
	jwtConfig       config.JWTConfig
	otpConfig       config.OTPConfig
	rateLimit       config.RateLimitConfig
	encryptor       *crypto.Encryptor
	otpService      *notification.OTPService
	pendingRegService *PendingRegistrationService
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	bloomFilter *BloomFilterService,
	jwtConfig config.JWTConfig,
	otpConfig config.OTPConfig,
	rateLimit config.RateLimitConfig,
	otpService *notification.OTPService,
	pendingRegService *PendingRegistrationService,
) (*AuthService, error) {
	encryptor, err := crypto.NewEncryptor(otpConfig.EncryptionKey)
	if err != nil {
		return nil, err
	}
	
	return &AuthService{
		userRepo:          userRepo,
		tokenRepo:         tokenRepo,
		bloomFilter:       bloomFilter,
		jwtConfig:         jwtConfig,
		otpConfig:         otpConfig,
		rateLimit:         rateLimit,
		encryptor:         encryptor,
		otpService:        otpService,
		pendingRegService: pendingRegService,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*models.RegistrationResponse, error) {
	if !utils.IsValidUsername(username) {
		return nil, errors.New("invalid username format")
	}
	
	if !utils.IsValidEmail(email) {
		return nil, errors.New("invalid email format")
	}
	
	valid, msg := utils.IsValidPassword(password)
	if !valid {
		return nil, errors.New(msg)
	}
	
	// Check verified users
	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already taken")
	}
	
	exists, err = s.userRepo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}
	
	// Check pending registrations
	pendingEmailExists, err := s.pendingRegService.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending email: %w", err)
	}
	if pendingEmailExists {
		// User already has a pending registration - resend OTP
		if err := s.otpService.GenerateAndSendEmailOTP(ctx, email); err != nil {
			return nil, fmt.Errorf("failed to resend verification email: %w", err)
		}
		return &models.RegistrationResponse{
			Success: true,
			Message: "A verification email has been resent. Please check your email.",
			Email:   email,
		}, nil
	}
	
	pendingUsernameExists, err := s.pendingRegService.UsernameExists(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check pending username: %w", err)
	}
	if pendingUsernameExists {
		return nil, errors.New("username already taken")
	}
	
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create pending user in dual storage (Redis + unverified_users table)
	_, err = s.pendingRegService.CreatePendingUser(ctx, username, email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending registration: %w", err)
	}
	
	// Send OTP via email
	if err := s.otpService.GenerateAndSendEmailOTP(ctx, email); err != nil {
		// Clean up pending registration if email fails
		s.pendingRegService.DeletePendingUser(ctx, email)
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}
	
	return &models.RegistrationResponse{
		Success: true,
		Message: "Registration successful! Please check your email for the verification code.",
		Email:   email,
	}, nil
}

// VerifyEmailAndCreateUser verifies the OTP and creates a verified user
func (s *AuthService) VerifyEmailAndCreateUser(ctx context.Context, email, otp string) (*models.EmailVerificationResponse, error) {
	// Verify OTP
	valid, remainingAttempts, err := s.otpService.VerifyEmailOTP(ctx, email, otp)
	if err != nil {
		return nil, err
	}
	
	if !valid {
		return nil, fmt.Errorf("invalid OTP, %d attempts remaining", remainingAttempts)
	}
	
	// Get pending user
	pendingUser, err := s.pendingRegService.GetPendingUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("pending registration not found")
	}
	
	// Create verified user
	now := time.Now()
	user := &models.User{
		ID:              pendingUser.ID,
		Username:        pendingUser.Username,
		Email:           pendingUser.Email,
		PasswordHash:    pendingUser.PasswordHash,
		IsActive:        true,
		IsVerified:      true,
		EmailVerifiedAt: &now,
		TOTPEnabled:     false,
	}
	
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Add username to bloom filter
	s.bloomFilter.AddUsername(user.Username)
	
	// Delete pending registration
	s.pendingRegService.DeletePendingUser(ctx, email)
	
	// Generate tokens
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}
	
	return &models.EmailVerificationResponse{
		Success: true,
		Message: "Email verified successfully! Your account has been created.",
		User:    user,
		Tokens:  tokens,
	}, nil
}

// ResendVerificationOTP resends the verification OTP
func (s *AuthService) ResendVerificationOTP(ctx context.Context, email string) error {
	// Check if pending registration exists
	pendingExists, err := s.pendingRegService.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check pending registration: %w", err)
	}
	
	if !pendingExists {
		return errors.New("no pending registration found for this email")
	}
	
	// Resend OTP
	if err := s.otpService.GenerateAndSendEmailOTP(ctx, email); err != nil {
		return err
	}
	
	return nil
}

// SetupTOTP generates TOTP secret for a verified user (optional step after email verification)
func (s *AuthService) SetupTOTP(userID string) (*utils.TOTPSetupData, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	
	if !user.IsVerified {
		return nil, errors.New("user must verify email before setting up TOTP")
	}
	
	if user.TOTPEnabled {
		return nil, errors.New("TOTP is already enabled for this user")
	}
	
	totpSetup, err := utils.GenerateTOTPSecret("GoConnect", user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	
	encryptedSecret, err := s.encryptor.EncryptTOTPSecret(totpSetup.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}
	
	if err := s.userRepo.UpdateTOTPSecret(user.ID, encryptedSecret); err != nil {
		return nil, fmt.Errorf("failed to store TOTP secret: %w", err)
	}
	
	return totpSetup, nil
}

func (s *AuthService) Login(username, password string) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	
	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}
	
	if !user.IsVerified {
		return &models.LoginResponse{
			Success:      false,
			Message:      "Please verify your email address before logging in. Check your email for the verification code.",
			RequiresTOTP: false,
		}, nil
	}
	
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}
	
	if user.TOTPEnabled {
		return &models.LoginResponse{
			Success:      true,
			User:         user,
			RequiresTOTP: true,
		}, nil
	}
	
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}
	
	return &models.LoginResponse{
		Success:      true,
		User:         user,
		Tokens:       tokens,
		RequiresTOTP: false,
	}, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(token, s.jwtConfig.Secret)
}

func (s *AuthService) RefreshTokens(refreshToken string) (*utils.TokenPair, error) {
	storedToken, err := s.tokenRepo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	
	if storedToken.IsRevoked {
		return nil, errors.New("refresh token has been revoked")
	}
	
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}
	
	user, err := s.userRepo.GetUserByID(storedToken.UserID)
	if err != nil {
		return nil, err
	}
	
	if err := s.tokenRepo.RevokeRefreshToken(refreshToken); err != nil {
		return nil, err
	}
	
	return s.generateTokenPair(user)
}

func (s *AuthService) GenerateOTP(email string) error {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return errors.New("email not found")
	}
	
	ctx := context.Background()
	if err := s.otpService.GenerateAndSend(ctx, user.ID, email); err != nil {
		return fmt.Errorf("failed to generate and send OTP: %w", err)
	}
	
	return nil
}

func (s *AuthService) VerifyOTP(email, otp string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("email not found")
	}
	
	ctx := context.Background()
	valid, err := s.otpService.Verify(ctx, user.ID, otp)
	if err != nil {
		return nil, err
	}
	
	if !valid {
		return nil, errors.New("invalid OTP")
	}
	
	return user, nil
}

func (s *AuthService) ResetPassword(email, otp, newPassword string) error {
	user, err := s.VerifyOTP(email, otp)
	if err != nil {
		return err
	}
	
	valid, msg := utils.IsValidPassword(newPassword)
	if !valid {
		return errors.New(msg)
	}
	
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	
	if err := s.userRepo.UpdatePassword(user.ID, passwordHash); err != nil {
		return err
	}
	
	return s.tokenRepo.RevokeAllUserTokens(user.ID)
}

func (s *AuthService) GetTOTPSetup(userID string) (*utils.TOTPSetupData, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	if user.TOTPSecret == "" {
		return nil, errors.New("TOTP not initialized for this user")
	}
	
	decryptedSecret, err := s.encryptor.DecryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return nil, err
	}
	
	qrURL := fmt.Sprintf("otpauth://totp/GoConnect:%s?secret=%s&issuer=GoConnect", user.Email, decryptedSecret)
	
	return &utils.TOTPSetupData{
		Secret:    decryptedSecret,
		QRCodeURL: qrURL,
	}, nil
}

func (s *AuthService) VerifyAndEnableTOTP(userID, code string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}
	
	if user.TOTPEnabled {
		return errors.New("TOTP is already enabled")
	}
	
	if user.TOTPSecret == "" {
		return errors.New("TOTP secret not found")
	}
	
	decryptedSecret, err := s.encryptor.DecryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return err
	}
	
	if !utils.VerifyTOTPCode(decryptedSecret, code) {
		return errors.New("invalid TOTP code")
	}
	
	return s.userRepo.EnableTOTP(userID)
}

func (s *AuthService) VerifyTOTPLogin(username, code string) (*utils.TokenPair, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username")
	}
	
	if !user.TOTPEnabled {
		return nil, errors.New("TOTP is not enabled for this user")
	}
	
	if user.TOTPSecret == "" {
		return nil, errors.New("TOTP secret not found")
	}
	
	decryptedSecret, err := s.encryptor.DecryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return nil, err
	}
	
	if !utils.VerifyTOTPCode(decryptedSecret, code) {
		return nil, errors.New("invalid TOTP code")
	}
	
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}
	
	return tokens, nil
}

func (s *AuthService) DisableTOTP(userID string) error {
	return s.userRepo.DisableTOTP(userID)
}

// VerifyTOTPAndLogin is an alias for VerifyTOTPLogin for gRPC compatibility
func (s *AuthService) VerifyTOTPAndLogin(username, code string) (*utils.TokenPair, error) {
	return s.VerifyTOTPLogin(username, code)
}

func (s *AuthService) generateTokenPair(user *models.User) (*utils.TokenPair, error) {
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Email,
		s.jwtConfig.Secret,
		s.jwtConfig.AccessExpiry,
	)
	if err != nil {
		return nil, err
	}
	
	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.jwtConfig.Secret,
		s.jwtConfig.RefreshExpiry,
	)
	if err != nil {
		return nil, err
	}
	
	tokenRecord := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtConfig.RefreshExpiry),
		IsRevoked: false,
	}
	
	if err := s.tokenRepo.CreateRefreshToken(tokenRecord); err != nil {
		return nil, err
	}
	
	return &utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
