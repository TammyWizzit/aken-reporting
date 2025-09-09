package middleware

import (
	"net/http"
	"strings"
	"time"

	"aken_reporting_service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the claims in our JWT token
type TokenClaims struct {
	MerchantID   string `json:"merchant_id"`
	MerchantName string `json:"merchant_name"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware provides JWT authentication middleware
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if disabled for development
		if config.IsDevMode() {
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
			sendJWTAuthError(c, "Missing Authorization header")
			return
		}

		// Check for Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			sendJWTAuthError(c, "Invalid Authorization header format. Expected Bearer token")
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			sendJWTAuthError(c, "Empty token")
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GetJWTSecret()), nil
		})

		if err != nil {
			sendJWTAuthError(c, "Invalid token: "+err.Error())
			return
		}

		// Verify token is valid
		if !token.Valid {
			sendJWTAuthError(c, "Invalid token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*TokenClaims)
		if !ok {
			sendJWTAuthError(c, "Invalid token claims")
			return
		}

		// Set merchant info in context
		c.Set("merchantID", claims.MerchantID)
		c.Set("merchant_id", claims.MerchantID)
		c.Set("merchantName", claims.MerchantName)
		c.Set("authenticated", true)

		c.Next()
	}
}

// sendJWTAuthError sends a JWT authentication error response
func sendJWTAuthError(c *gin.Context, message string) {
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
