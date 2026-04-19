package service

import (
	"context"
	"fmt"
	"time"

	"github.com/goconnect/internal/auth/repository"
	"github.com/goconnect/pkg/models"
	"github.com/goconnect/pkg/redis"
	"github.com/goconnect/pkg/utils"
)

// PendingRegistrationService manages dual storage for pending user registrations
// Stores in both Redis (fast access, TTL auto-cleanup) and DB (persistence, failover)
type PendingRegistrationService struct {
	redis              *redis.RedisClient
	unverifiedUserRepo *repository.UnverifiedUserRepository
	pendingRegTTL      time.Duration
}

func NewPendingRegistrationService(
	redisClient *redis.RedisClient,
	unverifiedUserRepo *repository.UnverifiedUserRepository,
	pendingRegTTL time.Duration,
) *PendingRegistrationService {
	return &PendingRegistrationService{
		redis:              redisClient,
		unverifiedUserRepo: unverifiedUserRepo,
		pendingRegTTL:      pendingRegTTL,
	}
}

// CreatePendingUser creates a pending user in both Redis and unverified_users table
func (s *PendingRegistrationService) CreatePendingUser(ctx context.Context, username, email, passwordHash string) (string, error) {
	userID := utils.GenerateGUID()
	expiresAt := time.Now().Add(48 * time.Hour)

	// 1. Store in Redis (fast access, TTL auto-cleanup)
	redisKey := fmt.Sprintf("pending:user:%s", email)
	userData := map[string]interface{}{
		"id":            userID,
		"username":      username,
		"email":         email,
		"password_hash": passwordHash,
		"expires_at":    expiresAt.Format(time.RFC3339),
	}
	
	if err := s.redis.HSet(ctx, redisKey, userData); err != nil {
		return "", fmt.Errorf("failed to store pending user in Redis: %w", err)
	}
	
	if err := s.redis.Expire(ctx, redisKey, s.pendingRegTTL); err != nil {
		return "", fmt.Errorf("failed to set Redis TTL: %w", err)
	}

	// 2. Store in DB (persistence, failover)
	unverifiedUser := &models.UnverifiedUser{
		ID:           userID,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
	}
	
	if err := s.unverifiedUserRepo.Create(unverifiedUser); err != nil {
		// Try to clean up Redis if DB insert fails
		s.redis.Delete(ctx, redisKey)
		return "", fmt.Errorf("failed to store pending user in DB: %w", err)
	}

	return userID, nil
}

// GetPendingUser retrieves a pending user, checking Redis first, then DB
func (s *PendingRegistrationService) GetPendingUser(ctx context.Context, email string) (*models.UnverifiedUser, error) {
	// Try Redis first (faster)
	redisKey := fmt.Sprintf("pending:user:%s", email)
	userData, err := s.redis.HGetAll(ctx, redisKey)
	
	if err == nil && len(userData) > 0 {
		// Parse Redis data
		expiresAt, _ := time.Parse(time.RFC3339, userData["expires_at"])
		createdAt, _ := time.Parse(time.RFC3339, userData["created_at"])
		
		return &models.UnverifiedUser{
			ID:           userData["id"],
			Username:     userData["username"],
			Email:        userData["email"],
			PasswordHash: userData["password_hash"],
			CreatedAt:    createdAt,
			ExpiresAt:    expiresAt,
		}, nil
	}

	// Fallback to DB
	return s.unverifiedUserRepo.GetByEmail(email)
}

// DeletePendingUser removes pending user from both Redis and DB
func (s *PendingRegistrationService) DeletePendingUser(ctx context.Context, email string) error {
	// Clear Redis
	redisKey := fmt.Sprintf("pending:user:%s", email)
	s.redis.Delete(ctx, redisKey)

	// Delete from unverified_users table
	return s.unverifiedUserRepo.DeleteByEmail(email)
}

// EmailExists checks if email exists in pending registrations
func (s *PendingRegistrationService) EmailExists(ctx context.Context, email string) (bool, error) {
	// Check Redis first
	redisKey := fmt.Sprintf("pending:user:%s", email)
	exists, err := s.redis.Exists(ctx, redisKey)
	if err == nil && exists > 0 {
		return true, nil
	}

	// Check DB
	return s.unverifiedUserRepo.EmailExists(email)
}

// UsernameExists checks if username exists in pending registrations
func (s *PendingRegistrationService) UsernameExists(ctx context.Context, username string) (bool, error) {
	// For username, we need to check DB (Redis is keyed by email)
	return s.unverifiedUserRepo.UsernameExists(username)
}
