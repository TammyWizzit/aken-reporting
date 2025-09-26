package middleware

import (
	"github.com/gin-gonic/gin"
)

// UseMySQLMiddleware sets a flag in the context to use MySQL database
func UseMySQLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("useMysql", true)
		c.Next()
	}
}

// GetUseMysqlFlag retrieves the MySQL flag from context
func GetUseMysqlFlag(c *gin.Context) bool {
	if useMysql, exists := c.Get("useMysql"); exists {
		if flag, ok := useMysql.(bool); ok {
			return flag
		}
	}
	return false // Default to PostgreSQL
}