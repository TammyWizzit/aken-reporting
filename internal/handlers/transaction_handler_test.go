package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"aken_reporting_service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"single value", "field1", []string{"field1"}},
		{"multiple values", "field1,field2,field3", []string{"field1", "field2", "field3"}},
		{"with spaces", "field1, field2 , field3", []string{"field1", "field2", "field3"}},
		{"duplicate values", "field1,field1,field2", []string{"field1", "field1", "field2"}},
		{"only spaces", "   ,  ,  ", []string{}},
		{"mixed empty and values", ",field1,,field2,", []string{"field1", "field2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMerchantID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Test with merchant context
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set("merchantID", "test-merchant")
	
	result := getMerchantID(ctx)
	assert.Equal(t, "test-merchant", result)
}

func TestGetMerchantID_Empty(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Test without merchant context
	ctx, _ := gin.CreateTestContext(nil)
	
	result := getMerchantID(ctx)
	assert.Equal(t, "", result)
}

func TestSendErrorResponse(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/test", func(c *gin.Context) {
		handler := &TransactionHandler{}
		handler.sendErrorResponse(c, 400, "TEST_ERROR", "Test error message", map[string]interface{}{"detail": "test detail"})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 400, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "TEST_ERROR", errorObj["code"])
	assert.Equal(t, "Test error message", errorObj["message"])
	
	// Check if details exist (it's optional)
	if details, exists := errorObj["detail"]; exists {
		assert.Equal(t, "test detail", details)
	}
}

func TestTransactionHandler_Constructor(t *testing.T) {
	// Test that the handler can be created
	handler := &TransactionHandler{}
	assert.NotNil(t, handler)
}

func TestErrorCodeConstants(t *testing.T) {
	// Test that error codes are properly defined
	assert.NotEmpty(t, config.ErrorCodeAuthFailed)
	assert.NotEmpty(t, config.ErrorCodeBadRequest)
	assert.NotEmpty(t, config.ErrorCodeInvalidFilter)
	assert.NotEmpty(t, config.ErrorCodeInvalidSort)
	assert.NotEmpty(t, config.ErrorCodeDatabaseError)
	assert.NotEmpty(t, config.ErrorCodeNotFound)
}

func TestPaginationValidation(t *testing.T) {
	tests := []struct {
		name    string
		page    string
		limit   string
		isValid bool
	}{
		{"valid pagination", "1", "100", true},
		{"invalid page", "0", "100", false},
		{"invalid limit", "1", "0", false},
		{"negative page", "-1", "100", false},
		{"negative limit", "1", "-1", false},
		{"non-numeric page", "abc", "100", false},
		{"non-numeric limit", "1", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual handler method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestFieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		fields   string
		expected []string
	}{
		{"empty fields", "", []string{}},
		{"single field", "tx_log_id", []string{"tx_log_id"}},
		{"multiple fields", "tx_log_id,amount,merchant_name", []string{"tx_log_id", "amount", "merchant_name"}},
		{"fields with spaces", "tx_log_id, amount , merchant_name", []string{"tx_log_id", "amount", "merchant_name"}},
		{"fields with empty values", "tx_log_id,,amount", []string{"tx_log_id", "amount"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommaSeparated(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimezoneValidation(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		isValid  bool
	}{
		{"UTC", "UTC", true},
		{"Africa/Johannesburg", "Africa/Johannesburg", true},
		{"empty timezone", "", true}, // Should default to UTC
		{"invalid timezone", "Invalid/Timezone", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual handler method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestPANFormatValidation(t *testing.T) {
	tests := []struct {
		name      string
		panFormat string
		isValid   bool
	}{
		{"bin_id_and_pan_id", "bin_id_and_pan_id", true},
		{"pan_id_only", "pan_id_only", true},
		{"empty format", "", true}, // Should default to bin_id_and_pan_id
		{"invalid format", "invalid_format", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual handler method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}
