package config

import "strings"

// API version constants
const (
	APIVersion = "2.0.0"
	ServiceName = "AKEN Reporting Service"
)

// Default pagination constants
const (
	DefaultPageSize = 100
	MaxPageSize     = 10000
	MinPageSize     = 1
)

// Field mapping constants for transaction queries
var FieldMappings = map[string]string{
	"tx_log_id":        "p.payment_tx_log_id",
	"tx_log_type":      "CASE WHEN p.payment_tx_type_id = 0 THEN 'payment' WHEN p.payment_tx_type_id = 1 THEN 'reversal' WHEN p.payment_tx_type_id = 2 THEN 'void' WHEN p.payment_tx_type_id = 3 THEN 'refund' WHEN p.payment_tx_type_id = 9 THEN 'mm purchase' WHEN p.payment_tx_type_id = 10 THEN 'mm refund' ELSE 'unknown' END",
	"tx_date_time":     "p.updated_at",
	"amount":           "p.amount",
	"merchant_id":      "m.merchant_id",
	"merchant_name":    "m.name",
	"device_id":        "p.device_id",
	"response_code":    "p.result_code",
	"auth_code":        "p.auth_code",
	"rrn":              "p.rrn",
	"pan":              "pan",
	"reversed":         "CASE WHEN p.reversed_tx_log_id IS NOT NULL THEN true ELSE false END",
	"settlement_status": "COALESCE(p.settlement_status, 'pending')",
	"stan":             "p.stan",
	"user_ref":         "p.meta->>'reference'",
	"meta":             "p.meta",
	"settlement_date":  "p.settlement_date",
	"card_type":        "p.card_type",
}

// Filter operator mappings
var FilterOperators = map[string]string{
	"eq":       "=",
	"ne":       "!=", 
	"gt":       ">",
	"gte":      ">=",
	"lt":       "<",
	"lte":      "<=",
	"like":     "LIKE",
	"ilike":    "ILIKE",
	"in":       "IN",
	"nin":      "NOT IN",
	"isnull":   "IS NULL",
	"isnotnull": "IS NOT NULL",
}

// PAN format mappings
var PANFormats = map[string]string{
	"bin_id_and_pan_id": "CONCAT(SUBSTRING(p.bin_id,1,4),' ',SUBSTRING(p.bin_id,5,2),'** **** ',p.pan_id)",
	"pan_id_only":       "CONCAT('***** ',p.pan_id)",
}

// Default fields to return if none specified
var DefaultFields = []string{
	"tx_log_id", "tx_log_type", "tx_date_time", "amount",
	"merchant_name", "response_code", "rrn", "pan",
}

// Error codes
const (
	ErrorCodeAuthFailed      = "AUTHENTICATION_FAILED"
	ErrorCodeAuthzFailed     = "AUTHORIZATION_FAILED"
	ErrorCodeInvalidFilter   = "INVALID_FILTER"
	ErrorCodeInvalidField    = "INVALID_FIELD"
	ErrorCodeInvalidSort     = "INVALID_SORT"
	ErrorCodeNotFound        = "NOT_FOUND"
	ErrorCodeTxNotFound      = "TRANSACTION_NOT_FOUND"
	ErrorCodeMerchantNotFound = "MERCHANT_NOT_FOUND"
	ErrorCodeDatabaseError   = "DATABASE_ERROR"
	ErrorCodeInternalError   = "INTERNAL_SERVER_ERROR"
	ErrorCodeNotImplemented  = "NOT_IMPLEMENTED"
	ErrorCodeBadRequest      = "BAD_REQUEST"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// User-friendly error messages
var ErrorMessages = map[string]string{
	ErrorCodeAuthFailed:      "Authentication failed. Please check your credentials.",
	ErrorCodeAuthzFailed:     "Access denied. You don't have permission to access this resource.",
	ErrorCodeInvalidFilter:   "Invalid filter expression. Please check your filter syntax.",
	ErrorCodeInvalidField:    "Invalid field specified. Please check the field name.",
	ErrorCodeInvalidSort:     "Invalid sort expression. Please check your sort syntax.",
	ErrorCodeNotFound:        "The requested resource was not found.",
	ErrorCodeTxNotFound:      "Transaction not found.",
	ErrorCodeMerchantNotFound: "Merchant not found.",
	ErrorCodeDatabaseError:   "Service temporarily unavailable. Please try again later.",
	ErrorCodeInternalError:   "An internal error occurred. Please try again later.",
	ErrorCodeNotImplemented:  "This feature is not yet implemented.",
	ErrorCodeBadRequest:      "Invalid request. Please check your parameters.",
	ErrorCodeServiceUnavailable: "Service temporarily unavailable. Please try again later.",
}

// Rate limiting constants
const (
	RateLimitStandard  = 1000  // requests per hour
	RateLimitPremium   = 5000  // requests per hour
	RateLimitEnterprise = 50000 // requests per hour
	RateLimitWindow    = 3600   // seconds (1 hour)
)

// GetUserFriendlyMessage returns a user-friendly error message for the given error code
func GetUserFriendlyMessage(errorCode string) string {
	if message, exists := ErrorMessages[errorCode]; exists {
		return message
	}
	return "An unexpected error occurred. Please try again later."
}

// IsInternalError checks if an error should be sanitized before exposing to API
func IsInternalError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for database-related errors that should be sanitized
	errorStr := err.Error()
	internalKeywords := []string{
		"column", "does not exist", "SQLSTATE", "syntax error",
		"foreign key", "constraint", "duplicate", "unique",
		"connection", "timeout", "deadlock", "lock",
	}
	
	for _, keyword := range internalKeywords {
		if strings.Contains(strings.ToLower(errorStr), strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}