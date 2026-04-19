package service

import (
	"context"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-redis/redis/v8"
	"github.com/goconnect/internal/auth/repository"
)

type BloomFilterService struct {
	filter     *bloom.BloomFilter
	redisClient *redis.Client
	userRepo   *repository.UserRepository
}

func NewBloomFilterService(redisClient *redis.Client, userRepo *repository.UserRepository) *BloomFilterService {
	filter := bloom.NewWithEstimates(1000000, 0.01)
	
	service := &BloomFilterService{
		filter:     filter,
		redisClient: redisClient,
		userRepo:   userRepo,
	}
	
	service.InitializeFromDB()
	
	return service
}

func (s *BloomFilterService) InitializeFromDB() error {
	usernames, err := s.userRepo.GetAllUsernames()
	if err != nil {
		return err
	}
	
	for _, username := range usernames {
		s.filter.AddString(username)
	}
	
	return nil
}

func (s *BloomFilterService) IsUsernamePossiblyTaken(username string) bool {
	return s.filter.TestString(username)
}

func (s *BloomFilterService) AddUsername(username string) {
	s.filter.AddString(username)
}

func (s *BloomFilterService) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	// Only use bloom filter for usernames > 3 characters
	if len(username) <= 3 {
		// For short usernames, skip bloom filter and query DB directly
		exists, err := s.userRepo.UsernameExists(username)
		if err != nil {
			return false, err
		}
		return !exists, nil
	}
	
	// Check bloom filter first (fast, probabilistic)
	if !s.IsUsernamePossiblyTaken(username) {
		// Definitely available
		return true, nil
	}
	
	// Bloom filter says "possibly taken" - verify with DB
	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return false, err
	}
	
	return !exists, nil
}
