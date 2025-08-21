package middleware

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/services"

	"github.com/gin-gonic/gin"
)

// CacheMiddleware provides Redis caching for API responses
func CacheMiddleware(cacheService services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip caching if cache service is nil
		if cacheService == nil {
			c.Next()
			return
		}

		// Skip caching for non-GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Skip caching if disabled
		if !config.IsRedisEnabled() {
			c.Next()
			return
		}

		// Skip caching for /transactions endpoints to prevent stale data during development/testing
		if strings.Contains(c.Request.URL.Path, "/transactions") {
			c.Next()
			return
		}

		// Generate cache key from request
		cacheKey := generateCacheKey(c)
		
		// Try to get from cache
		var cachedResponse map[string]interface{}
		if err := cacheService.Get(cacheKey, &cachedResponse); err == nil && cachedResponse != nil {
			// Return cached response
			c.JSON(http.StatusOK, cachedResponse)
			c.Abort()
			return
		}

		// Store original response writer
		originalWriter := c.Writer
		
		// Create custom response writer to capture response
		responseWriter := &responseCapture{
			ResponseWriter: originalWriter,
			body:          make([]byte, 0),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Cache successful responses
		if c.Writer.Status() == http.StatusOK && len(responseWriter.body) > 0 {
			var responseData map[string]interface{}
			if err := json.Unmarshal(responseWriter.body, &responseData); err == nil {
				// Add cache metadata
				responseData["cached"] = false
				responseData["cache_timestamp"] = time.Now().Unix()
				
				// Cache the response
				ttl := config.GetRedisTTL()
				cacheService.Set(cacheKey, responseData, ttl)
			}
		}
	}
}

// CacheInvalidationMiddleware invalidates cache when data changes
func CacheInvalidationMiddleware(cacheService services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Skip invalidation if cache service is nil
		if cacheService == nil {
			return
		}

		// Invalidate cache for data modification requests
		if shouldInvalidateCache(c.Request.Method) {
			merchantID := getMerchantID(c)
			if merchantID != "" {
				// Invalidate transaction cache
				cacheService.InvalidateTransactionCache(merchantID)
				
				// Invalidate merchant cache
				cacheService.InvalidateMerchantCache(merchantID)
			}
		}
	}
}

// responseCapture captures the response body for caching
type responseCapture struct {
	gin.ResponseWriter
	body []byte
}

func (r *responseCapture) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// generateCacheKey creates a unique cache key from the request
func generateCacheKey(c *gin.Context) string {
	// Include method, path, query parameters, and headers that affect response
	parts := []string{
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery,
	}

	// Include important headers that affect response
	importantHeaders := []string{"Authorization", "X-Request-ID", "Accept"}
	for _, header := range importantHeaders {
		if value := c.GetHeader(header); value != "" {
			parts = append(parts, header+":"+value)
		}
	}

	// Create hash of the key parts
	key := strings.Join(parts, "|")
	hash := md5.Sum([]byte(key))
	
	return fmt.Sprintf("%s:api:%x", config.GetRedisKeyPrefix(), hash[:8])
}

// shouldInvalidateCache determines if cache should be invalidated
func shouldInvalidateCache(method string) bool {
	invalidationMethods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, m := range invalidationMethods {
		if method == m {
			return true
		}
	}
	return false
}

// getMerchantID extracts merchant ID from context
func getMerchantID(c *gin.Context) string {
	if merchantID, exists := c.Get("merchantID"); exists {
		if id, ok := merchantID.(string); ok {
			return id
		}
	}
	return ""
}

// CacheControlMiddleware adds cache control headers
func CacheControlMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add cache control headers
		c.Header("Cache-Control", "private, max-age=300") // 5 minutes
		c.Header("Vary", "Accept, Authorization")
		
		c.Next()
	}
}
