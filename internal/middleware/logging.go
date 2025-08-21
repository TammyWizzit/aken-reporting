package middleware

import (
	"aken_reporting_service/internal/utils"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides structured JSON logging for HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Use our structured logger instead of default Gin logging
		utils.LogHTTPRequest(
			param.Method,
			param.Path,
			strconv.Itoa(param.StatusCode),
			fmt.Sprintf("%.3f", float64(param.Latency.Nanoseconds())/1000000), // Convert to milliseconds
			param.Request.Header.Get("X-Request-ID"),
		)
		return "" // Return empty string since we're handling logging ourselves
	})
}

// RequestLoggingMiddleware logs incoming requests with additional context
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		if raw != "" {
			path = path + "?" + raw
		}

		// Extract merchant info if available
		merchantID, _ := c.Get("merchant_id")
		merchantIDStr := ""
		if merchantID != nil {
			merchantIDStr = fmt.Sprintf("%v", merchantID)
		}

		// Log request start
		utils.LogTrace("Request received", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        path,
			"merchant_id": merchantIDStr,
			"user_agent":  c.Request.UserAgent(),
			"remote_addr": c.ClientIP(),
		})

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		
		// Log request completion
		utils.LogHTTPRequest(
			c.Request.Method,
			path,
			strconv.Itoa(c.Writer.Status()),
			fmt.Sprintf("%.3f", float64(latency.Nanoseconds())/1000000),
			c.GetHeader("X-Request-ID"),
		)
	}
}