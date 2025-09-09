package middleware

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"aken_reporting_service/internal/config"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides Basic Authentication middleware compatible with AKEN v1
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip authentication if disabled for development
		if config.IsDevMode() {
			log.Println("Development mode: skipping authentication")
			c.Set("merchantID", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchant_id", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchantName", "NASS WALLET")
			c.Set("authenticated", true)
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			sendAuthError(c, "Missing Authorization header")
			return
		}

		// Check for Basic Auth
		if !strings.HasPrefix(authHeader, "Basic ") {
			sendAuthError(c, "Invalid Authorization header format. Expected Basic Auth")
			return
		}

		// Decode Basic Auth credentials
		encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
		credentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			sendAuthError(c, "Invalid Base64 encoding in Authorization header")
			return
		}

		// Parse merchant_id:password
		credentialParts := strings.SplitN(string(credentials), ":", 2)
		if len(credentialParts) != 2 {
			sendAuthError(c, "Invalid credential format. Expected merchant_id:password")
			return
		}

		merchantID := credentialParts[0]
		password := credentialParts[1]

		// Validate credentials (simplified - in production, check against database/service)
		if !isValidMerchant(merchantID, password) {
			sendAuthError(c, "Invalid merchant credentials")
			return
		}

		// Set merchant info in context
		c.Set("merchantID", merchantID)
		c.Set("merchant_id", merchantID)
		c.Set("merchantName", getMerchantName(merchantID))
		c.Set("authenticated", true)

		log.Printf("Authenticated merchant: %s", merchantID)
		c.Next()
	})
}

// RequestIDMiddleware adds request ID tracking
func RequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
			c.Request.Header.Set("X-Request-ID", requestID)
		}
		c.Set("requestID", requestID)
		c.Next()
	})
}

// ResponseHeadersMiddleware adds standard response headers
func ResponseHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-API-Version", config.APIVersion)
		c.Header("X-Service-Name", config.ServiceName)

		if requestID, exists := c.Get("requestID"); exists {
			if id, ok := requestID.(string); ok {
				c.Header("X-Request-ID", id)
			}
		}

		// Rate limiting headers (placeholder values)
		c.Header("X-RateLimit-Limit", "1000")
		c.Header("X-RateLimit-Remaining", "999")
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Unix()+3600))
		c.Header("X-RateLimit-Window", "3600")

		c.Next()
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins or all origins in development
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost:5173",
			"https://aken-eu.staging.wizzitdigital.com",
		}

		if config.IsDevMode() {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

// Helper functions

func sendAuthError(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":       config.ErrorCodeAuthFailed,
			"message":    message,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": c.GetHeader("X-Request-ID"),
		},
	})
	c.Abort()
}

// isValidMerchant validates merchant credentials
// In production, this would query the merchants table or call an auth service
func isValidMerchant(merchantID, password string) bool {
	// System test user (hardcoded for development compatibility with AKEN v1)
	if merchantID == "d1a3fefe-101d-11ea-8d71-362b9e155667" {
		return true
	}

	// For now, accept any valid UUID as merchant ID with any password
	// In production, implement proper credential validation
	if len(merchantID) >= 32 && strings.Contains(merchantID, "-") {
		return true
	}

	return false
}

// getMerchantName returns the merchant name for a given merchant ID
// In production, this would query the merchants table
func getMerchantName(merchantID string) string {
	// Hardcoded mappings for development
	merchantNames := map[string]string{
		"d1a3fefe-101d-11ea-8d71-362b9e155667": "Wizzit Test User",
		"9cda37a0-4813-11ef-95d7-c5ac867bb9fc": "NASS WALLET",
	}

	if name, exists := merchantNames[merchantID]; exists {
		return name
	}

	return "Unknown Merchant"
}
