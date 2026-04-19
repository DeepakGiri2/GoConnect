package notification

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/goconnect/pkg/redis"
)

type OTPService struct {
	redis             *redis.RedisClient
	email             *EmailService
	otpLength         int
	otpExpiry         time.Duration
	maxVerifyAttempts int
	maxResendAttempts int
	resendCooldown    time.Duration
	blockDuration     time.Duration
}

func NewOTPService(
	redisClient *redis.RedisClient,
	emailService *EmailService,
	otpLength int,
	otpExpiry time.Duration,
	maxVerifyAttempts int,
	maxResendAttempts int,
	resendCooldown time.Duration,
	blockDuration time.Duration,
) *OTPService {
	return &OTPService{
		redis:             redisClient,
		email:             emailService,
		otpLength:         otpLength,
		otpExpiry:         otpExpiry,
		maxVerifyAttempts: maxVerifyAttempts,
		maxResendAttempts: maxResendAttempts,
		resendCooldown:    resendCooldown,
		blockDuration:     blockDuration,
	}
}

// GenerateNumericOTP generates a cryptographically secure random N-digit OTP
func GenerateNumericOTP(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("OTP length must be positive")
	}

	// Calculate the maximum value (10^length - 1)
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	
	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

// GenerateAndSendEmailOTP generates and sends OTP via email
func (s *OTPService) GenerateAndSendEmailOTP(ctx context.Context, email string) error {
	// Check if user is blocked
	blockKey := fmt.Sprintf("otp:block:%s", email)
	if blocked, _ := s.redis.Exists(ctx, blockKey); blocked > 0 {
		blockTTL, _ := s.redis.TTL(ctx, blockKey)
		return fmt.Errorf("too many failed attempts - blocked for %v", blockTTL.Round(time.Second))
	}

	// Check resend attempts
	resendKey := fmt.Sprintf("otp:resend:%s", email)
	resendCount, err := s.redis.Get(ctx, resendKey)
	if err == nil {
		count, _ := strconv.Atoi(resendCount)
		if count >= s.maxResendAttempts {
			return errors.New("maximum OTP resend limit reached")
		}
	}

	// Check resend cooldown
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", email)
	if exists, _ := s.redis.Exists(ctx, cooldownKey); exists > 0 {
		ttl, _ := s.redis.TTL(ctx, cooldownKey)
		return fmt.Errorf("please wait %v before requesting new OTP", ttl)
	}

	// Generate 6-digit OTP
	otp, err := GenerateNumericOTP(s.otpLength)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store in Redis with TTL
	otpKey := fmt.Sprintf("otp:%s", email)
	if err := s.redis.Set(ctx, otpKey, otp, s.otpExpiry); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send email
	if err := s.email.SendOTP(email, otp); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	// Track resend attempts
	s.redis.Increment(ctx, resendKey)
	s.redis.Expire(ctx, resendKey, 20*time.Minute) // Full registration window

	// Set cooldown
	s.redis.Set(ctx, cooldownKey, "1", s.resendCooldown)

	return nil
}

// VerifyEmailOTP verifies the OTP for email verification
func (s *OTPService) VerifyEmailOTP(ctx context.Context, email, otp string) (bool, int, error) {
	// Check if email is blocked
	blockKey := fmt.Sprintf("otp:block:%s", email)
	if exists, _ := s.redis.Exists(ctx, blockKey); exists > 0 {
		ttl, _ := s.redis.TTL(ctx, blockKey)
		return false, 0, fmt.Errorf("account blocked, try again in %v", ttl)
	}

	// Check global verify attempts (across all OTPs)
	attemptKey := fmt.Sprintf("otp:attempts:%s", email)
	attempts, _ := s.redis.Get(ctx, attemptKey)
	attemptCount, _ := strconv.Atoi(attempts)

	if attemptCount >= s.maxVerifyAttempts {
		// Block email for configured duration
		s.redis.Set(ctx, blockKey, "1", s.blockDuration)

		// Send block notification email (async, non-blocking)
		go s.email.SendBlockNotification(email, s.blockDuration)

		// Clear OTP and attempts
		s.redis.Delete(ctx, fmt.Sprintf("otp:%s", email), attemptKey)
		return false, 0, errors.New("too many failed attempts, account blocked for 5 minutes")
	}

	// Get stored OTP
	otpKey := fmt.Sprintf("otp:%s", email)
	storedOTP, err := s.redis.Get(ctx, otpKey)
	if err != nil {
		return false, s.maxVerifyAttempts - attemptCount, errors.New("OTP expired or not found")
	}

	// Verify OTP
	if storedOTP != otp {
		// Increment global attempt counter
		s.redis.Increment(ctx, attemptKey)
		s.redis.Expire(ctx, attemptKey, 20*time.Minute)

		remainingAttempts := s.maxVerifyAttempts - (attemptCount + 1)
		return false, remainingAttempts, fmt.Errorf("invalid OTP, %d attempts remaining", remainingAttempts)
	}

	// Success - clear all OTP-related keys
	s.redis.Delete(ctx,
		otpKey,
		attemptKey,
		fmt.Sprintf("otp:resend:%s", email),
		fmt.Sprintf("otp:cooldown:%s", email),
	)

	return true, s.maxVerifyAttempts, nil
}

// GetBlockTimeRemaining returns the remaining time for email block
func (s *OTPService) GetBlockTimeRemaining(ctx context.Context, email string) (time.Duration, error) {
	blockKey := fmt.Sprintf("otp:block:%s", email)
	return s.redis.TTL(ctx, blockKey)
}

// GetRemainingAttempts returns the number of remaining verification attempts
func (s *OTPService) GetRemainingAttempts(ctx context.Context, email string) (int, error) {
	attemptKey := fmt.Sprintf("otp:attempts:%s", email)
	attempts, err := s.redis.Get(ctx, attemptKey)
	if err != nil {
		return s.maxVerifyAttempts, nil
	}

	attemptCount, _ := strconv.Atoi(attempts)
	remaining := s.maxVerifyAttempts - attemptCount
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// InvalidateOTP invalidates the OTP for an email
func (s *OTPService) InvalidateOTP(ctx context.Context, email string) error {
	otpKey := fmt.Sprintf("otp:%s", email)
	return s.redis.Delete(ctx, otpKey)
}

// Deprecated: Use GenerateAndSendEmailOTP instead
func (s *OTPService) GenerateAndSend(ctx context.Context, userID, email string) error {
	return s.GenerateAndSendEmailOTP(ctx, email)
}

// GetBlockStatus returns block status information for an email
func (s *OTPService) GetBlockStatus(ctx context.Context, email string) (bool, int64, int, error) {
	blockKey := fmt.Sprintf("otp:block:%s", email)
	
	// Check if blocked
	blocked, _ := s.redis.Exists(ctx, blockKey)
	if blocked == 0 {
		// Not blocked - return remaining attempts
		attempts, _ := s.GetRemainingAttempts(ctx, email)
		return false, 0, attempts, nil
	}
	
	// Get remaining block time
	ttl, _ := s.redis.TTL(ctx, blockKey)
	remainingSeconds := int64(ttl.Seconds())
	
	return true, remainingSeconds, 0, nil
}

// Deprecated: Use VerifyEmailOTP instead
func (s *OTPService) Verify(ctx context.Context, userID, otp string) (bool, error) {
	valid, _, err := s.VerifyEmailOTP(ctx, userID, otp)
	return valid, err
}
