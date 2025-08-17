package database

import (
	"fmt"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Timeout     time.Duration
}

// DefaultRetryConfig returns default retry settings
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       1 * time.Second,
		Timeout:     5 * time.Second,
	}
}

// RetryableOperation represents a database operation that can be retried
type RetryableOperation func() error

// RetryWithBackoff executes a database operation with retry logic
func RetryWithBackoff(operation RetryableOperation, config RetryConfig) error {
	var lastErr error
	
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute operation
		err := operation()
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Don't retry on last attempt
		if attempt == config.MaxAttempts {
			break
		}
		
		// Log retry attempt (for debugging)
		fmt.Printf("Database operation failed (attempt %d/%d): %v. Retrying in %v...\n", 
			attempt, config.MaxAttempts, err, config.Delay)
		
		// Wait before retry
		time.Sleep(config.Delay)
		
		// Exponential backoff
		config.Delay *= 2
	}
	
	return fmt.Errorf("operation failed after %d attempts: %v", config.MaxAttempts, lastErr)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common retryable database errors
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"no route to host",
		"broken pipe",
		"EOF",
	}
	
	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
