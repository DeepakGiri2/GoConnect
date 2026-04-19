package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/goconnect/internal/gateway/routes"
	gwmiddleware "github.com/goconnect/internal/gateway/middleware"
	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/logger"
	"github.com/goconnect/pkg/middleware"
	"github.com/goconnect/pkg/ratelimit"
	"github.com/goconnect/pkg/redis"
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

	authServiceAddr := fmt.Sprintf("%s:%d", cfg.AuthService.Host, cfg.AuthService.Port)
	conn, err := grpc.Dial(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("Failed to connect to Auth Service", zap.Error(err))
	}
	defer conn.Close()

	authClient := pb.NewAuthServiceClient(conn)
	logger.Log.Info("Connected to Auth Service")

	// Initialize Redis client
	redisClient, err := redis.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password)
	if err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Log.Info("Connected to Redis")

	// Initialize rate limiter
	slidingWindowLimiter := ratelimit.NewSlidingWindowLimiter(redisClient)
	rateLimiter := gwmiddleware.NewRateLimiter(slidingWindowLimiter)

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.LoggerMiddleware(logger.Log))

	// Setup routes with rate limiting
	routeConfig := struct {
		RegistrationIPLimit    int
		RegistrationEmailLimit int
		UsernameCheckLimit     int
	}{
		RegistrationIPLimit:    cfg.RateLimit.RegistrationIP,
		RegistrationEmailLimit: cfg.RateLimit.RegistrationEmail,
		UsernameCheckLimit:     cfg.RateLimit.UsernameCheck,
	}
	routes.SetupRoutes(router, authClient, cfg.JWT.Secret, rateLimiter, routeConfig)

	port := cfg.Server.Port
	logger.Log.Info(fmt.Sprintf("API Gateway starting on port %d", port))

	go func() {
		if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
			logger.Log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down API Gateway...")
}
