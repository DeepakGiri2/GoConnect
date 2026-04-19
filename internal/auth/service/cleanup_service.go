package service

import (
	"context"
	"log"
	"time"

	"github.com/goconnect/internal/auth/repository"
)

// CleanupService handles periodic cleanup of expired unverified users
type CleanupService struct {
	unverifiedUserRepo *repository.UnverifiedUserRepository
	interval           time.Duration
}

func NewCleanupService(unverifiedUserRepo *repository.UnverifiedUserRepository, interval time.Duration) *CleanupService {
	return &CleanupService{
		unverifiedUserRepo: unverifiedUserRepo,
		interval:           interval,
	}
}

// Start begins the cleanup background job
// This should be called as a goroutine: go cleanupService.Start(ctx)
func (s *CleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("[CleanupService] Started with interval: %v", s.interval)

	// Run immediately on start
	s.runCleanup()

	for {
		select {
		case <-ticker.C:
			s.runCleanup()
		case <-ctx.Done():
			log.Println("[CleanupService] Stopping cleanup service")
			return
		}
	}
}

// runCleanup executes the cleanup operation
func (s *CleanupService) runCleanup() {
	deleted, err := s.unverifiedUserRepo.CleanupExpired()
	if err != nil {
		log.Printf("[CleanupService] Error cleaning up expired users: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("[CleanupService] Cleaned up %d expired unverified users", deleted)
	}
}

// RunOnce executes cleanup once (useful for testing or manual triggers)
func (s *CleanupService) RunOnce() (int64, error) {
	return s.unverifiedUserRepo.CleanupExpired()
}
