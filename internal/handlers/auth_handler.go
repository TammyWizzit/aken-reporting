package handlers

import (
	"net/http"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/database"
	"aken_reporting_service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct{}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// TokenClaims represents the claims in our JWT token
type TokenClaims struct {
	MerchantID   string `json:"merchant_id"`
	MerchantName string `json:"merchant_name"`
	jwt.RegisteredClaims
}

// GenerateTokenRequest represents the request for token generation
type GenerateTokenRequest struct {
	MerchantID string `json:"merchant_id" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// TokenResponse represents the response containing the generated token
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	TokenType string `json:"token_type"`
}

// GenerateToken handles JWT token generation for development/testing
func (ah *AuthHandler) GenerateToken(c *gin.Context) {
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":       config.ErrorCodeBadRequest,
				"message":    "Invalid request body",
				"details":    err.Error(),
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"request_id": c.GetHeader("X-Request-ID"),
			},
		})
		return
	}

	// Validate merchant credentials (simplified for development)
	if !isValidMerchantCredentials(req.MerchantID, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":       config.ErrorCodeAuthFailed,
				"message":    "Invalid merchant credentials",
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"request_id": c.GetHeader("X-Request-ID"),
			},
		})
		return
	}

	// Generate JWT token
	token, expiresIn, err := generateJWTToken(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":       config.ErrorCodeInternalError,
				"message":    "Failed to generate token",
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"request_id": c.GetHeader("X-Request-ID"),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": expiresIn,
		"token_type": "Bearer",
	})
}

// VerifyToken handles JWT token verification
func (ah *AuthHandler) VerifyToken(c *gin.Context) {
	// Token validation is already handled by JWTAuthMiddleware
	merchantID, _ := c.Get("merchantID")
	merchantName, _ := c.Get("merchantName")

	c.JSON(http.StatusOK, gin.H{
		"valid":         true,
		"merchant_id":   merchantID,
		"merchant_name": merchantName,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
}

// Helper functions

func isValidMerchantCredentials(merchantID, password string) bool {
	// Look up merchant in database
	var merchant models.Merchant
	err := database.DB.Where("merchant_id = ? AND active = ?", merchantID, true).First(&merchant).Error
	if err != nil {
		// Merchant not found or inactive
		return false
	}

	// Validate password against database password field
	return password == merchant.Password
}

func generateJWTToken(merchantID string) (string, int64, error) {
	// Get merchant name (in production, fetch from database)
	merchantName := getMerchantNameByID(merchantID)

	// Set token expiration (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)
	expiresIn := expirationTime.Unix() - time.Now().Unix()

	// Create the claims
	claims := TokenClaims{
		MerchantID:   merchantID,
		MerchantName: merchantName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.GetJWTIssuer(),
			Subject:   merchantID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with secret
	tokenString, err := token.SignedString([]byte(config.GetJWTSecret()))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

func getMerchantNameByID(merchantID string) string {
	// Fetch merchant name from database
	var merchant models.Merchant
	err := database.DB.Select("name").Where("merchant_id = ? AND active = ?", merchantID, true).First(&merchant).Error
	if err != nil {
		// Merchant not found, return default
		return "Unknown Merchant"
	}

	return merchant.Name
}
