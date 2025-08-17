package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aken_reporting_service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidBasicAuth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		merchantID, exists := c.Get("merchantID")
		assert.True(t, exists)
		assert.Equal(t, "d1a3fefe-101d-11ea-8d71-362b9e155667", merchantID)
		
		authenticated, exists := c.Get("authenticated")
		assert.True(t, exists)
		assert.True(t, authenticated.(bool))
		
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with valid Basic Auth (using known valid merchant ID)
	credentials := "d1a3fefe-101d-11ea-8d71-362b9e155667:test-password"
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCredentials)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request without Authorization header
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Missing Authorization header")
}

func TestAuthMiddleware_InvalidAuthFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with invalid Authorization format
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid Authorization header format")
}

func TestAuthMiddleware_InvalidBase64(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with invalid Base64
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic invalid-base64")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid Base64 encoding")
}

func TestAuthMiddleware_InvalidCredentialsFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with invalid credentials format
	credentials := "invalid-format-without-colon"
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCredentials)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid credential format")
}

func TestAuthMiddleware_InvalidCredentials(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with invalid credentials
	credentials := "invalid-merchant:wrong-password"
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCredentials)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid merchant credentials")
}

func TestRequestIDMiddleware_WithExistingRequestID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("requestID")
		assert.True(t, exists)
		assert.Equal(t, "existing-request-id", requestID)
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with existing X-Request-ID
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)
	// The header is set in the response by ResponseHeadersMiddleware, not RequestIDMiddleware
	// So we check that the requestID is properly set in context
	// The test above already verifies this in the handler function
}

func TestRequestIDMiddleware_WithoutRequestID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("requestID")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		assert.True(t, strings.HasPrefix(requestID.(string), "req_"))
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request without X-Request-ID
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)
	// The header is set by ResponseHeadersMiddleware, so we check the context instead
	// The test above already verifies this in the handler function
}

func TestResponseHeadersMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ResponseHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, config.APIVersion, w.Header().Get("X-API-Version"))
	assert.Equal(t, config.ServiceName, w.Header().Get("X-Service-Name"))
}

func TestResponseHeadersMiddleware_WithRequestID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ResponseHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})

	// Create request with X-Request-ID
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test-request-id", w.Header().Get("X-Request-ID"))
}

func TestAuthMiddleware_DevelopmentMode(t *testing.T) {
	// This test is skipped because we can't easily mock the dev mode
	// In a real implementation, we would temporarily enable dev mode for testing
	t.Skip("Skipping dev mode test - requires environment setup")
}

func TestIsValidMerchant(t *testing.T) {
	// Test valid merchant credentials
	tests := []struct {
		name       string
		merchantID string
		password   string
		expected   bool
	}{
		{"valid system user", "d1a3fefe-101d-11ea-8d71-362b9e155667", "test-password", true},
		{"valid UUID format", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc", "test-password", true},
		{"invalid merchant", "invalid-merchant", "test-password", false},
		{"empty merchant", "", "test-password", false},
		{"empty password", "test-merchant", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMerchant(tt.merchantID, tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMerchantName(t *testing.T) {
	// Test merchant name retrieval
	tests := []struct {
		name       string
		merchantID string
		expected   string
	}{
		{"known system user", "d1a3fefe-101d-11ea-8d71-362b9e155667", "Wizzit Test User"},
		{"known merchant", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc", "NASS WALLET"},
		{"unknown merchant", "unknown-merchant", "Unknown Merchant"},
		{"empty merchant", "", "Unknown Merchant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMerchantName(tt.merchantID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSendAuthError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		sendAuthError(c, "Test error message")
	})

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Test error message")
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}
