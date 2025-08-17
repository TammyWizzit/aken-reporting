package database

import (
	"context"
	"time"
)

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Latency   int64  `json:"latency_ms"`
}

// CheckDatabaseHealth checks if the database is healthy
func CheckDatabaseHealth() HealthStatus {
	start := time.Now()
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test database connection
	sqlDB, err := DB.DB()
	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   "Database connection failed",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	}
	
	// Ping the database
	err = sqlDB.PingContext(ctx)
	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   "Database ping failed",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	}
	
	// Test a simple query
	var result int
	err = DB.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   "Database query test failed",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	}
	
	return HealthStatus{
		Status:    "healthy",
		Message:   "Database is responding normally",
		Timestamp: time.Now(),
		Latency:   time.Since(start).Milliseconds(),
	}
}

// IsDatabaseHealthy returns true if database is healthy
func IsDatabaseHealthy() bool {
	health := CheckDatabaseHealth()
	return health.Status == "healthy"
}

// GetDatabaseStatus returns a user-friendly status message
func GetDatabaseStatus() string {
	if IsDatabaseHealthy() {
		return "Database is available"
	}
	return "Database is temporarily unavailable"
}
