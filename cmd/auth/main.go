package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	authgrpc "github.com/goconnect/internal/auth/grpc"
	"github.com/goconnect/internal/auth/repository"
	"github.com/goconnect/internal/auth/service"
	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/db"
	"github.com/goconnect/pkg/logger"
	"github.com/goconnect/pkg/notification"
	redispkg "github.com/goconnect/pkg/redis"
	pb "github.com/goconnect/api/shared/proto_gen/api/shared/proto"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.InitLogger(cfg.Server.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	database, err := db.Connect(db.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
	})
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	logger.Log.Info("Connected to database")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	logger.Log.Info("Connected to Redis")

	redisWrapper, err := redispkg.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		logger.Log.Fatal("Failed to create Redis client wrapper", zap.Error(err))
	}
	defer redisWrapper.Close()

	emailConfig := notification.EmailConfig{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		FromEmail:    cfg.Email.FromEmail,
		FromName:     cfg.Email.FromName,
	}
	emailService := notification.NewEmailService(emailConfig)
	logger.Log.Info("Email service initialized")

	// Initialize OTP service with all parameters
	otpService := notification.NewOTPService(
		redisWrapper,
		emailService,
		cfg.OTP.Length,
		cfg.OTP.Expiry,
		cfg.OTP.MaxVerifyAttempts,
		cfg.OTP.MaxResendAttempts,
		cfg.OTP.ResendCooldown,
		cfg.OTP.BlockDuration,
	)
	logger.Log.Info("OTP service initialized")

	// Initialize repositories
	userRepo := repository.NewUserRepository(database, cfg.Retry)
	oauthRepo := repository.NewOAuthRepository(database, cfg.Retry)
	tokenRepo := repository.NewTokenRepository(database)
	unverifiedUserRepo := repository.NewUnverifiedUserRepository(database)
	logger.Log.Info("Repositories initialized")

	// Initialize bloom filter
	bloomFilter := service.NewBloomFilterService(redisClient, userRepo)
	logger.Log.Info("Bloom filter initialized")

	// Initialize pending registration service
	pendingRegService := service.NewPendingRegistrationService(
		redisWrapper,
		unverifiedUserRepo,
		cfg.RateLimit.PendingRegTTL,
	)
	logger.Log.Info("Pending registration service initialized")

	// Initialize cleanup service
	cleanupService := service.NewCleanupService(unverifiedUserRepo, cfg.RateLimit.UnverifiedUserCleanup)
	logger.Log.Info("Cleanup service initialized")

	// Start cleanup background job
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go cleanupService.Start(ctx)
	logger.Log.Info("Cleanup service started")

	// Initialize auth service with new parameters
	authService, err := service.NewAuthService(
		userRepo,
		tokenRepo,
		bloomFilter,
		cfg.JWT,
		cfg.OTP,
		cfg.RateLimit,
		otpService,
		pendingRegService,
	)
	if err != nil {
		logger.Log.Fatal("Failed to initialize auth service", zap.Error(err))
	}
	logger.Log.Info("Auth service initialized")

	oauthService := service.NewOAuthService(userRepo, oauthRepo, tokenRepo, bloomFilter, cfg.JWT, cfg.OAuth)
	logger.Log.Info("OAuth service initialized")

	// Initialize gRPC server with OTP service
	grpcServer := authgrpc.NewAuthGRPCServer(authService, oauthService, bloomFilter, otpService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Log.Fatal("Failed to listen", zap.Error(err))
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, grpcServer)

	logger.Log.Info(fmt.Sprintf("Auth Service gRPC server listening on port %d", cfg.Server.Port))

	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down Auth Service...")
	server.GracefulStop()
}
