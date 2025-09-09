package database

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Latency   int64     `json:"latency_ms"`
}

// CheckDatabaseHealth checks if both databases are healthy
func CheckDatabaseHealth() HealthStatus {
	start := time.Now()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postgresHealthy := checkSingleDatabase(DB, "PostgreSQL", ctx)
	mysqlHealthy := checkSingleDatabase(MySQLDB, "MySQL", ctx)

	// Determine overall health status
	if postgresHealthy && mysqlHealthy {
		return HealthStatus{
			Status:    "healthy",
			Message:   "Both databases are responding normally",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	} else if postgresHealthy || mysqlHealthy {
		return HealthStatus{
			Status:    "degraded",
			Message:   "One database is unavailable but service can continue",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	} else {
		return HealthStatus{
			Status:    "unhealthy",
			Message:   "Both databases are unavailable",
			Timestamp: time.Now(),
			Latency:   time.Since(start).Milliseconds(),
		}
	}
}

// checkSingleDatabase checks the health of a single database
func checkSingleDatabase(db *gorm.DB, dbType string, ctx context.Context) bool {
	if db == nil {
		return false
	}

	// Test database connection
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}

	// Ping the database
	err = sqlDB.PingContext(ctx)
	return err == nil
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
