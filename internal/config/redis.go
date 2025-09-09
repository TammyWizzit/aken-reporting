package config

import (
	"os"
	"strconv"
	"time"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
	Timeout  time.Duration
}

// GetRedisConfig returns Redis configuration from environment variables
func GetRedisConfig() *RedisConfig {
	host := getEnvOrDefault("REDIS_HOST", "localhost")
	port := getEnvOrDefault("REDIS_PORT", "6379")
	password := getEnvOrDefault("REDIS_PASSWORD", "")

	// If password is empty string, set it to empty string (no authentication)
	if password == "" {
		password = ""
	}

	db, _ := strconv.Atoi(getEnvOrDefault("REDIS_DB", "0"))
	poolSize, _ := strconv.Atoi(getEnvOrDefault("REDIS_POOL_SIZE", "10"))

	timeout, _ := strconv.Atoi(getEnvOrDefault("REDIS_TIMEOUT", "5"))

	return &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
		Timeout:  time.Duration(timeout) * time.Second,
	}
}

// IsRedisEnabled checks if Redis is enabled
func IsRedisEnabled() bool {
	return getEnvOrDefault("REDIS_ENABLED", "true") == "true"
}

// GetRedisTTL returns the default TTL for cached items
func GetRedisTTL() time.Duration {
	ttl, _ := strconv.Atoi(getEnvOrDefault("REDIS_TTL_SECONDS", "1800")) // 30 minutes default
	return time.Duration(ttl) * time.Second
}

// GetRedisKeyPrefix returns the prefix for Redis keys
func GetRedisKeyPrefix() string {
	return getEnvOrDefault("REDIS_KEY_PREFIX", "aken:reporting:")
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
