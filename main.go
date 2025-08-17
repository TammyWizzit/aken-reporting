// main.go
package main

import (
	"aken_reporting_service/api/routes"
	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/database"
	"aken_reporting_service/internal/middleware"
	"aken_reporting_service/internal/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv() // Load environment variables from .env file
	database.ConnectDB() // Connect to the database

	// Initialize Redis cache service
	cacheService, err := services.NewCacheService()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v", err)
		log.Println("Continuing without caching...")
	} else {
		log.Println("Redis cache service initialized successfully")
	}

	log.Printf("Starting AKEN Reporting Service on port %s", os.Getenv("PORT"))
	
	r := gin.Default()
	
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
	log.Println("=== DEBUG: Starting DISABLE_AUTH check ===")
	disableAuth := os.Getenv("DISABLE_AUTH")
	log.Printf("DISABLE_AUTH environment variable: '%s'", disableAuth)
	
	if disableAuth == "true" {
		log.Println("DISABLE_AUTH=true, skipping authentication for development")
		// Simple development middleware to set merchant info in context
		r.Use(func(c *gin.Context) {
			fmt.Printf("=== DEVELOPMENT MIDDLEWARE: Setting merchant info for %s ===\n", c.Request.URL.Path)
			c.Set("merchantID", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchant_id", "9cda37a0-4813-11ef-95d7-c5ac867bb9fc")
			c.Set("merchantName", "NASS WALLET")
			c.Set("authenticated", true)
			c.Next()
		})
	} else {
		log.Println("DISABLE_AUTH not set to 'true', enabling authentication...")
		r.Use(middleware.AuthMiddleware())
	}

	// Debug endpoint to check env vars and auth status
	r.GET("/debug", func(c *gin.Context) {
		merchantID, _ := c.Get("merchantID")
		authenticated, _ := c.Get("authenticated")
		
		c.JSON(200, gin.H{
			"service": "AKEN Reporting Service",
			"version": "2.0.0",
			"DISABLE_AUTH": os.Getenv("DISABLE_AUTH"),
			"ENV": os.Getenv("ENV"),
			"merchantID": merchantID,
			"authenticated": authenticated,
			"timestamp": os.Getenv("TIMESTAMP"),
		})
	})

	// Set trusted proxies to localhost only for safety
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
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
				"code": "ENDPOINT_NOT_FOUND",
				"message": fmt.Sprintf("API endpoint %s %s not found", c.Request.Method, c.Request.URL.Path),
				"timestamp": "2025-01-28T10:30:00.000Z",
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

	log.Printf("🚀 AKEN Reporting Service starting on port %s", port)
	log.Printf("📊 Available endpoints:")
	log.Printf("   GET /api/v2/health - Health check")
	log.Printf("   GET /api/v2/transactions - List transactions")
	log.Printf("   GET /api/v2/transactions/:id - Get single transaction")
	log.Printf("   POST /api/v2/transactions/search - Advanced search")
	log.Printf("   GET /api/v2/merchants/:id/summary - Merchant summary")

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}