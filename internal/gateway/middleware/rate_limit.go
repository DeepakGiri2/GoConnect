package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goconnect/pkg/ratelimit"
)

type RateLimiter struct {
	limiter *ratelimit.SlidingWindowLimiter
}

func NewRateLimiter(limiter *ratelimit.SlidingWindowLimiter) *RateLimiter {
	return &RateLimiter{
		limiter: limiter,
	}
}

// RegistrationRateLimit applies rate limiting for registration endpoint
func (rl *RateLimiter) RegistrationRateLimit(ipLimit, emailLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		clientIP := GetClientIP(c)
		
		// Check IP rate limit
		allowed, resetIn, err := rl.limiter.CheckIPRegistrationLimit(ctx, clientIP, ipLimit)
		if err == nil && !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "Too many registration attempts from this IP",
				"retry_in": fmt.Sprintf("%d seconds", int(resetIn.Seconds())),
			})
			c.Abort()
			return
		}
		
		// Increment IP counter on success
		defer rl.limiter.IncrementIPRegistration(ctx, clientIP)
		
		c.Next()
	}
}

// EmailRateLimit applies rate limiting based on email (for OTP requests)
func (rl *RateLimiter) EmailRateLimit(maxAttempts int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		
		// Get email from request body
		var body struct {
			Email string `json:"email"`
		}
		
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			c.Abort()
			return
		}
		
		// Rebind the body for the handler
		c.Set("email", body.Email)
		
		// Check email rate limit
		allowed, resetIn, err := rl.limiter.CheckEmailRegistrationLimit(ctx, body.Email, maxAttempts)
		if err == nil && !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "Too many OTP requests for this email",
				"retry_in": fmt.Sprintf("%d seconds", int(resetIn.Seconds())),
			})
			c.Abort()
			return
		}
		
		// Increment email counter on success
		defer rl.limiter.IncrementEmailRegistration(ctx, body.Email)
		
		c.Next()
	}
}

// UsernameCheckRateLimit applies rate limiting for username availability checks
func (rl *RateLimiter) UsernameCheckRateLimit(maxAttempts int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		clientIP := GetClientIP(c)
		
		// Check IP rate limit
		allowed, resetIn, err := rl.limiter.CheckUsernameCheckLimit(ctx, clientIP, maxAttempts)
		if err == nil && !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "Too many username checks",
				"retry_in": fmt.Sprintf("%d seconds", int(resetIn.Seconds())),
			})
			c.Abort()
			return
		}
		
		// Increment counter on success
		defer rl.limiter.IncrementUsernameCheck(ctx, clientIP)
		
		c.Next()
	}
}
