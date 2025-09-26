// main.go
package main

import (
	"aken_reporting_service/api/routes"
	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/database"
	"aken_reporting_service/internal/middleware"
	"aken_reporting_service/internal/services"
	"aken_reporting_service/internal/utils"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()     // Load environment variables from .env file
	database.ConnectDB() // Connect to the database

	// Set Gin mode based on environment variable
	ginMode := config.GetGinMode()
	gin.SetMode(ginMode)
	
	utils.LogInfo("Gin framework configured", map[string]interface{}{
		"mode": ginMode,
	})

	// Initialize Redis cache service
	cacheService, err := services.NewCacheService()
	if err != nil {
		utils.LogWarn("Failed to initialize Redis cache", map[string]interface{}{
			"error": err.Error(),
		})
		utils.LogInfo("Continuing without caching", nil)
	} else {
		utils.LogInfo("Redis cache service initialized successfully", nil)
	}

	utils.LogInfo("Starting AKEN Reporting Service", map[string]interface{}{
		"port": os.Getenv("PORT"),
	})

	// Create Gin router without default middleware
	r := gin.New()

	// Add custom recovery middleware
	r.Use(gin.Recovery())

	// Add custom logging middleware
	r.Use(middleware.LoggingMiddleware())

	// Disable automatic redirects to prevent 301 redirects for trailing slashes
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{
		"http://localhost:8080",
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:3001",
		"https://aken-eu.staging.wizzitdigital.com",
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-Request-ID"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// Add cache middleware
	r.Use(middleware.CacheControlMiddleware())
	r.Use(middleware.CacheMiddleware(cacheService))
	r.Use(middleware.CacheInvalidationMiddleware(cacheService))

	// Check DISABLE_AUTH and skip authentication for development
	utils.LogTrace("Starting DISABLE_AUTH check", nil)
	disableAuth := os.Getenv("DISABLE_AUTH")
	utils.LogTrace("DISABLE_AUTH environment variable", map[string]interface{}{
		"disable_auth": disableAuth,
	})

	if disableAuth == "true" {
		utils.LogInfo("DISABLE_AUTH=true, skipping authentication for development", nil)
		// Simple development middleware to set merchant info in context
		r.Use(func(c *gin.Context) {
			utils.LogTrace("Setting merchant info for development", map[string]interface{}{
				"path":        c.Request.URL.Path,
				"merchant_id": "9cda37a0-4813-11ef-95d7-c5ac867bb9fc",
			})
			c.Set("merchantID", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchant_id", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchantName", "NASS WALLET")
			c.Set("authenticated", true)
			c.Next()
		})
	} else {
		utils.LogInfo("DISABLE_AUTH not set to 'true', JWT authentication will be applied per route", nil)
		// JWT middleware is applied per route group in routes.go, not globally
	}

	// Debug endpoint to check env vars and auth status
	r.GET("/debug", func(c *gin.Context) {
		merchantID, _ := c.Get("merchantID")
		authenticated, _ := c.Get("authenticated")

		c.JSON(200, gin.H{
			"service":       "AKEN Reporting Service",
			"version":       "2.0.0",
			"DISABLE_AUTH":  os.Getenv("DISABLE_AUTH"),
			"ENV":           os.Getenv("ENV"),
			"merchantID":    merchantID,
			"authenticated": authenticated,
			"timestamp":     os.Getenv("TIMESTAMP"),
		})
	})

	// Set trusted proxies to localhost only for safety
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		utils.LogError("Failed to set trusted proxies", err, nil)
		os.Exit(1)
	}

	// Simple test endpoint
	r.GET("/simple-test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "AKEN Reporting Service is running!",
			"service": "aken-reporting-service",
			"version": "2.0.0",
		})
	})

	// Setup all API routes
	routes.SetupRoutes(r, database.DB, cacheService)

	// Handle 404 for unknown API routes
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"code":       "ENDPOINT_NOT_FOUND",
				"message":    fmt.Sprintf("API endpoint %s %s not found", c.Request.Method, c.Request.URL.Path),
				"timestamp":  "2025-01-28T10:30:00.000Z",
				"request_id": c.GetHeader("X-Request-ID"),
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "Endpoint not found"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // Different port from main aken service
	}

	utils.LogInfo("AKEN Reporting Service starting", map[string]interface{}{
		"port": port,
	})
	utils.LogTrace("Available endpoints", map[string]interface{}{
		"endpoints": []string{
			"GET /api/v2/health - Health check",
			"GET /api/v2/transactions - List transactions",
			"GET /api/v2/transactions/:id - Get single transaction",
			"POST /api/v2/transactions/search - Advanced search",
			"GET /api/v2/merchants/:id/summary - Merchant summary",
		},
	})

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		utils.LogError("Failed to start server", err, nil)
		os.Exit(1)
	}
}
