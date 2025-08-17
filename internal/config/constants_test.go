package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIVersion(t *testing.T) {
	assert.Equal(t, "2.0.0", APIVersion)
	assert.Equal(t, "AKEN Reporting Service", ServiceName)
}

func TestPaginationConstants(t *testing.T) {
	assert.Equal(t, 100, DefaultPageSize)
	assert.Equal(t, 10000, MaxPageSize)
	assert.Equal(t, 1, MinPageSize)
}

func TestFieldMappings(t *testing.T) {
	// Test that all expected field mappings exist
	expectedFields := []string{
		"tx_log_id", "tx_log_type", "tx_date_time", "amount",
		"merchant_id", "merchant_name", "device_id", "response_code",
		"auth_code", "rrn", "pan", "reversed", "settlement_status",
		"stan", "user_ref", "meta", "settlement_date", "card_type",
	}

	for _, field := range expectedFields {
		t.Run(field, func(t *testing.T) {
			mapping, exists := FieldMappings[field]
			assert.True(t, exists, "Field mapping for %s should exist", field)
			assert.NotEmpty(t, mapping, "Field mapping for %s should not be empty", field)
		})
	}
}

func TestFilterOperators(t *testing.T) {
	// Test that all expected operators exist
	expectedOperators := []string{
		"eq", "ne", "gt", "gte", "lt", "lte",
		"like", "ilike", "in", "nin", "isnull", "isnotnull",
	}

	for _, op := range expectedOperators {
		t.Run(op, func(t *testing.T) {
			operator, exists := FilterOperators[op]
			assert.True(t, exists, "Filter operator %s should exist", op)
			assert.NotEmpty(t, operator, "Filter operator %s should not be empty", op)
		})
	}
}

func TestPANFormats(t *testing.T) {
	// Test that all expected PAN formats exist
	expectedFormats := []string{
		"bin_id_and_pan_id", "pan_id_only",
	}

	for _, format := range expectedFormats {
		t.Run(format, func(t *testing.T) {
			panFormat, exists := PANFormats[format]
			assert.True(t, exists, "PAN format %s should exist", format)
			assert.NotEmpty(t, panFormat, "PAN format %s should not be empty", format)
		})
	}
}

func TestDefaultFields(t *testing.T) {
	// Test that default fields are not empty
	assert.NotEmpty(t, DefaultFields)
	assert.Len(t, DefaultFields, 8) // Should have 8 default fields

	// Test that all default fields have mappings
	for _, field := range DefaultFields {
		t.Run(field, func(t *testing.T) {
			_, exists := FieldMappings[field]
			assert.True(t, exists, "Default field %s should have a mapping", field)
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined and not empty
	errorCodes := []string{
		ErrorCodeAuthFailed, ErrorCodeAuthzFailed, ErrorCodeInvalidFilter,
		ErrorCodeInvalidField, ErrorCodeInvalidSort, ErrorCodeNotFound,
		ErrorCodeTxNotFound, ErrorCodeMerchantNotFound, ErrorCodeDatabaseError,
		ErrorCodeInternalError, ErrorCodeNotImplemented, ErrorCodeBadRequest,
	}

	for _, code := range errorCodes {
		t.Run(code, func(t *testing.T) {
			assert.NotEmpty(t, code, "Error code should not be empty")
		})
	}
}

func TestRateLimitingConstants(t *testing.T) {
	// Test rate limiting constants
	assert.Equal(t, 1000, RateLimitStandard)
	assert.Equal(t, 5000, RateLimitPremium)
	assert.Equal(t, 50000, RateLimitEnterprise)
	assert.Equal(t, 3600, RateLimitWindow)

	// Test that rate limits are in ascending order
	assert.True(t, RateLimitStandard < RateLimitPremium)
	assert.True(t, RateLimitPremium < RateLimitEnterprise)
}

func TestFieldMappingValues(t *testing.T) {
	// Test specific field mapping values
	tests := []struct {
		field    string
		expected string
	}{
		{"tx_log_id", "p.payment_tx_log_id"},
		{"amount", "p.amount"},
		{"merchant_id", "m.merchant_id"},
		{"merchant_name", "m.name"},
		{"response_code", "p.result_code"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			mapping, exists := FieldMappings[tt.field]
			assert.True(t, exists)
			assert.Equal(t, tt.expected, mapping)
		})
	}
}

func TestFilterOperatorValues(t *testing.T) {
	// Test specific filter operator values
	tests := []struct {
		operator string
		expected string
	}{
		{"eq", "="},
		{"ne", "!="},
		{"gt", ">"},
		{"gte", ">="},
		{"lt", "<"},
		{"lte", "<="},
		{"like", "LIKE"},
		{"ilike", "ILIKE"},
		{"in", "IN"},
		{"nin", "NOT IN"},
		{"isnull", "IS NULL"},
		{"isnotnull", "IS NOT NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.operator, func(t *testing.T) {
			operator, exists := FilterOperators[tt.operator]
			assert.True(t, exists)
			assert.Equal(t, tt.expected, operator)
		})
	}
}
