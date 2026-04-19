package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractIP extracts the real client IP from the request
func ExtractIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Check X-Forwarded-For header (for proxies/load balancers)
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			// X-Forwarded-For can contain multiple IPs, take the first one
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				clientIP = strings.TrimSpace(ips[0])
			}
		}
		
		// Check X-Real-IP header (some proxies use this)
		if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}
		
		// Validate IP
		if net.ParseIP(clientIP) == nil {
			clientIP = c.ClientIP() // Fallback to Gin's default
		}
		
		// Store in context for later use
		c.Set("client_ip", clientIP)
		
		c.Next()
	}
}

// GetClientIP retrieves the client IP from the context
func GetClientIP(c *gin.Context) string {
	if ip, exists := c.Get("client_ip"); exists {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return c.ClientIP()
}
