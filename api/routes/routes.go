package routes

import (
	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/handlers"
	"aken_reporting_service/internal/middleware"
	"aken_reporting_service/internal/repositories"
	"aken_reporting_service/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes initializes all API routes with dependency injection following the household project pattern
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Apply global middleware
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.ResponseHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Create API version group
	v2 := router.Group("/api/v2")

	// Initialize repositories
	transactionRepo := repositories.NewTransactionRepository(db)

	// Initialize services
	transactionService := services.NewTransactionService(transactionRepo)

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// Register transaction routes
	RegisterTransactionRoutes(v2, transactionHandler)

	// Health check endpoint
	v2.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   config.ServiceName,
			"version":   config.APIVersion,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(startTime).Seconds(),
		})
	})

	// API info endpoint
	v2.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     config.ServiceName,
			"version":     config.APIVersion,
			"description": "Modern RESTful API for AKEN transaction reporting",
			"endpoints": gin.H{
				"transactions": gin.H{
					"list":           "GET /api/v2/transactions",
					"get":            "GET /api/v2/transactions/:id",
					"search":         "POST /api/v2/transactions/search",
					"export":         "POST /api/v2/transactions/export (coming soon)",
					"batch":          "POST /api/v2/transactions/batch (coming soon)",
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

// RegisterTransactionRoutes sets up all transaction-related routes
func RegisterTransactionRoutes(rg *gin.RouterGroup, handler *handlers.TransactionHandler) {
	transactions := rg.Group("/transactions")
	{
		// Core transaction endpoints
		transactions.GET("", handler.GetTransactions)
		transactions.GET("/:id", handler.GetTransactionByID)
		transactions.POST("/search", handler.AdvancedTransactionSearch)

		// Future endpoints (placeholders)
		transactions.POST("/export", handleNotImplemented("Transaction export"))
		transactions.POST("/batch", handleNotImplemented("Batch operations"))
		transactions.GET("/stream", handleNotImplemented("Real-time transaction stream"))
	}

	// Merchant-specific routes
	merchants := rg.Group("/merchants")
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

// handleNotImplemented returns a 501 Not Implemented response for future features
func handleNotImplemented(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": gin.H{
				"code":       config.ErrorCodeNotImplemented,
				"message":    feature + " not yet implemented",
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
				"request_id": c.GetHeader("X-Request-ID"),
			},
		})
	}
}

var startTime = time.Now()