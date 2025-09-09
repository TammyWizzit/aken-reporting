package routes

import (
	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/database"
	"aken_reporting_service/internal/handlers"
	"aken_reporting_service/internal/middleware"
	"aken_reporting_service/internal/repositories"
	"aken_reporting_service/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var startTime = time.Now()

// SetupRoutes initializes all API routes with dependency injection following the household project pattern
func SetupRoutes(router *gin.Engine, db *gorm.DB, cacheService services.CacheService) {
	// Apply global middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.ResponseHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Create API version groups
	v1 := router.Group("/api/v1")
	v2 := router.Group("/api/v2")

	// Initialize repositories with both databases
	transactionRepo := repositories.NewTransactionRepository(database.DB, database.MySQLDB)

	// Initialize services
	transactionService := services.NewTransactionService(transactionRepo, cacheService)

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	authHandler := handlers.NewAuthHandler()

	// Register authentication routes (for testing and development)
	RegisterAuthRoutes(v2, authHandler)

	// Register transaction routes
	RegisterTransactionRoutes(v2, transactionHandler)

	// Register v1 transaction lookup route
	RegisterV1TransactionRoutes(v1, transactionHandler)

	// Health check endpoint (supports both GET and HEAD)
	healthHandler := func(c *gin.Context) {
		// Check database health
		dbHealth := database.CheckDatabaseHealth()

		// Determine overall status
		status := "healthy"
		httpStatus := http.StatusOK

		if dbHealth.Status != "healthy" {
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    status,
			"service":   config.ServiceName,
			"version":   config.APIVersion,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(startTime).Seconds(),
			"database":  dbHealth,
		})
	}

	// Register health endpoint for both GET and HEAD requests
	v2.GET("/health", healthHandler)
	v2.HEAD("/health", healthHandler)

	// API info endpoint
	v2.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     config.ServiceName,
			"version":     config.APIVersion,
			"description": "Modern RESTful API for AKEN transaction reporting",
			"endpoints": gin.H{
				"transactions": gin.H{
					"list":   "GET /api/v2/transactions",
					"get":    "GET /api/v2/transactions/:id",
					"search": "POST /api/v2/transactions/search",
					"totals": "GET /api/v2/transactions/totals",
					"export": "POST /api/v2/transactions/export (coming soon)",
					"batch":  "POST /api/v2/transactions/batch (coming soon)",
				},
				"merchants": gin.H{
					"summary":      "GET /api/v2/merchants/:id/summary",
					"transactions": "GET /api/v2/merchants/:id/transactions",
				},
				"analytics": gin.H{
					"summary": "GET /api/v2/analytics/summary (coming soon)",
					"custom":  "POST /api/v2/analytics/custom (coming soon)",
				},
				"system": gin.H{
					"health": "GET /api/v2/health",
					"info":   "GET /api/v2/info",
				},
			},
			"features": []string{
				"Advanced filtering with operators (eq, ne, gt, gte, lt, lte, like, in, between)",
				"Field selection to reduce payload size",
				"Flexible sorting on any field",
				"Pagination with navigation links",
				"Merchant-specific transaction summaries",
				"Compatible with existing AKEN v1 authentication",
				"RESTful design with proper HTTP methods",
				"Comprehensive error handling",
			},
		})
	})
}

// RegisterAuthRoutes sets up authentication routes for token generation and verification
func RegisterAuthRoutes(rg *gin.RouterGroup, handler *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		// Public endpoints for token generation (development/testing)
		auth.POST("/generate-token", handler.GenerateToken)

		// Protected endpoint for token verification
		auth.GET("/verify-token", middleware.JWTAuthMiddleware(), handler.VerifyToken)
	}
}

// RegisterTransactionRoutes sets up all transaction-related routes
func RegisterTransactionRoutes(rg *gin.RouterGroup, handler *handlers.TransactionHandler) {
	// Apply JWT authentication to all transaction routes
	// Dev mode handling is done at the middleware level in main.go
	transactions := rg.Group("/transactions")
	transactions.Use(middleware.JWTAuthMiddleware())
	{
		// Core transaction endpoints - each merchant can only see their own data
		transactions.GET("", handler.GetTransactions)
		transactions.GET("/:id", handler.GetTransactionByID)
		transactions.POST("/search", handler.AdvancedTransactionSearch)
		transactions.GET("/totals", handler.GetTransactionTotals)

		// Future endpoints (placeholders)
		transactions.POST("/export", handleNotImplemented("Transaction export"))
		transactions.POST("/batch", handleNotImplemented("Batch operations"))
		transactions.GET("/stream", handleNotImplemented("Real-time transaction stream"))
	}

	// Merchant-specific routes - protected by JWT authentication
	merchants := rg.Group("/merchants")
	merchants.Use(middleware.JWTAuthMiddleware())
	{
		merchants.GET("/:merchant_id/summary", handler.GetMerchantSummary)
		merchants.GET("/:merchant_id/transactions", handler.GetMerchantTransactions)
	}

	// Analytics routes (future features)
	analytics := rg.Group("/analytics")
	{
		analytics.GET("/summary", handleNotImplemented("Analytics summary"))
		analytics.POST("/custom", handleNotImplemented("Custom analytics"))
	}

	// Export management routes (future features)
	exports := rg.Group("/exports")
	{
		exports.GET("/:export_id", handleNotImplemented("Export status checking"))
		exports.GET("/:export_id/download", handleNotImplemented("Export download"))
	}
}

// RegisterV1TransactionRoutes sets up v1 efinance transaction routes
func RegisterV1TransactionRoutes(rg *gin.RouterGroup, handler *handlers.TransactionHandler) {
	efinance := rg.Group("/efinance")
	efinance.Use(middleware.JWTAuthMiddleware())
	efinance.Use(middleware.UseMySQLMiddleware()) // Use MySQL for efinance APIs
	{
		transactions := efinance.Group("/transactions")
		{
			// V1 efinance transaction totals endpoint
			transactions.POST("/totals", handler.GetTransactionLookup)
			// V1 efinance transaction lookup endpoint
			transactions.POST("/lookup", handler.SearchTransactionDetails)
		}
	}
}

// handleNotImplemented returns a 501 Not Implemented response for future features
func handleNotImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"code":       config.ErrorCodeNotImplemented,
			"message":    feature + " not yet implemented",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": c.GetHeader("X-Request-ID"),
		})
	}
}
