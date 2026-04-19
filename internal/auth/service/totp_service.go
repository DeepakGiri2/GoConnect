package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"time"

	"github.com/goconnect/internal/auth/repository"
	"github.com/goconnect/pkg/redis"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPService struct {
	totpRepo    *repository.TOTPRepository
	redisClient *redis.RedisClient
}

func NewTOTPService(totpRepo *repository.TOTPRepository, redisClient *redis.RedisClient) *TOTPService {
	return &TOTPService{
		totpRepo:    totpRepo,
		redisClient: redisClient,
	}
}

func (s *TOTPService) GenerateSecret(userID, email string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: email,
		Period:      30,
		SecretSize:  32,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	pendingKey := fmt.Sprintf("totp:pending:%s", userID)
	if err := s.redisClient.Set(ctx, pendingKey, key.Secret(), 10*time.Minute); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *TOTPService) VerifyAndEnable(userID, code string) error {
	ctx := context.Background()
	
	pendingKey := fmt.Sprintf("totp:pending:%s", userID)
	secret, err := s.redisClient.Get(ctx, pendingKey)
	if err != nil {
		return errors.New("TOTP setup not initiated or expired")
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return errors.New("invalid TOTP code")
	}

	if err := s.totpRepo.SaveTOTPSecret(userID, secret); err != nil {
		return err
	}

	if err := s.totpRepo.MarkTOTPVerified(userID); err != nil {
		return err
	}

	s.redisClient.Delete(ctx, pendingKey)

	return nil
}

func (s *TOTPService) VerifyCode(userID, code string) error {
	ctx := context.Background()

	blockedKey := fmt.Sprintf("totp:blocked:%s", userID)
	blocked, _ := s.redisClient.Exists(ctx, blockedKey)
	if blocked > 0 {
		return errors.New("too many failed attempts, please try again in 15 minutes")
	}

	attemptsKey := fmt.Sprintf("totp:attempts:%s", userID)
	attempts, _ := s.redisClient.Increment(ctx, attemptsKey)
	s.redisClient.Expire(ctx, attemptsKey, 5*time.Minute)

	if attempts > 5 {
		s.redisClient.Set(ctx, blockedKey, "1", 15*time.Minute)
		return errors.New("too many failed attempts, account temporarily blocked")
	}

	secret, err := s.totpRepo.GetTOTPSecret(userID)
	if err != nil {
		return err
	}

	usedKey := fmt.Sprintf("totp:used:%s:%s", userID, code)
	used, _ := s.redisClient.Exists(ctx, usedKey)
	if used > 0 {
		return errors.New("TOTP code already used")
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return errors.New("invalid TOTP code")
	}

	s.redisClient.SetNX(ctx, usedKey, "1", 90*time.Second)
	s.redisClient.Delete(ctx, attemptsKey)

	return nil
}

func (s *TOTPService) Disable(userID string) error {
	ctx := context.Background()
	
	pendingKey := fmt.Sprintf("totp:pending:%s", userID)
	s.redisClient.Delete(ctx, pendingKey)
	
	return s.totpRepo.DisableTOTP(userID)
}

func (s *TOTPService) IsEnabled(userID string) (bool, error) {
	return s.totpRepo.IsTOTPEnabled(userID)
}

func generateRandomSecret() (string, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}
