package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goconnect/internal/gateway/handlers"
	gwmiddleware "github.com/goconnect/internal/gateway/middleware"
	"github.com/goconnect/pkg/middleware"
	pb "github.com/goconnect/api/shared/proto_gen/api/shared/proto"
)

func SetupRoutes(router *gin.Engine, authClient pb.AuthServiceClient, jwtSecret string, rateLimiter *gwmiddleware.RateLimiter, config struct {
	RegistrationIPLimit    int
	RegistrationEmailLimit int
	UsernameCheckLimit     int
}) {
	authHandler := handlers.NewAuthHandler(authClient)
	oauthHandler := handlers.NewOAuthHandler(authClient)

	// Apply IP extraction middleware globally
	router.Use(gwmiddleware.ExtractIP())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			// Registration with IP and email rate limiting
			auth.POST("/register", 
				rateLimiter.RegistrationRateLimit(config.RegistrationIPLimit, config.RegistrationEmailLimit),
				authHandler.Register,
			)
			
			// Email verification endpoints
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/resend-otp", authHandler.ResendOTP)
			auth.GET("/block-status", authHandler.GetBlockStatus)
			
			// TOTP setup and verification
			auth.POST("/setup-totp", authHandler.SetupTOTP)
			auth.POST("/verify-totp", authHandler.VerifyTOTP)
			
			// Login and token management
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			
			// Password reset flow
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/reset-password", authHandler.ResetPassword)
			
			// Username availability check with rate limiting
			auth.GET("/check-username", 
				rateLimiter.UsernameCheckRateLimit(config.UsernameCheckLimit),
				authHandler.CheckUsername,
			)

			// OAuth endpoints
			auth.GET("/oauth/:provider", oauthHandler.InitiateOAuth)
			auth.GET("/callback/:provider", oauthHandler.OAuthCallback)
		}

		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(jwtSecret))
		{
			protected.GET("/me", func(c *gin.Context) {
				userID := c.GetString("user_id")
				username := c.GetString("username")
				email := c.GetString("email")

				c.JSON(200, gin.H{
					"user_id":  userID,
					"username": username,
					"email":    email,
				})
			})
		}
	}
}
