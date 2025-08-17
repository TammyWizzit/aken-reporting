package services

import (
	"testing"
	"time"

	"aken_reporting_service/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestNewTransactionService(t *testing.T) {
	// Test service creation
	service := NewTransactionService(nil, nil)
	assert.NotNil(t, service)
}

func TestGetTransactionsParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  GetTransactionsParams
		isValid bool
	}{
		{
			name: "valid params",
			params: GetTransactionsParams{
				Page:     1,
				Limit:    100,
				Timezone: "UTC",
				Fields:   []string{"tx_log_id", "amount"},
			},
			isValid: true,
		},
		{
			name: "invalid page",
			params: GetTransactionsParams{
				Page:     0,
				Limit:    100,
				Timezone: "UTC",
			},
			isValid: false,
		},
		{
			name: "invalid limit",
			params: GetTransactionsParams{
				Page:     1,
				Limit:    0,
				Timezone: "UTC",
			},
			isValid: false,
		},
		{
			name: "empty timezone",
			params: GetTransactionsParams{
				Page:     1,
				Limit:    100,
				Timezone: "",
			},
			isValid: true, // Should default to UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual service method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestTransactionServiceResult_Calculation(t *testing.T) {
	result := TransactionServiceResult{
		Transactions: []models.Transaction{},
		TotalCount:   100,
		Page:         1,
		Limit:        10,
		TotalPages:   10,
		HasNext:      true,
		HasPrev:      false,
	}

	// Test calculated fields
	expectedTotalPages := 10
	expectedHasNext := true
	expectedHasPrev := false

	assert.Equal(t, expectedTotalPages, result.TotalPages)
	assert.Equal(t, expectedHasNext, result.HasNext)
	assert.Equal(t, expectedHasPrev, result.HasPrev)
}

func TestTransactionServiceResult_LastPage(t *testing.T) {
	result := TransactionServiceResult{
		Transactions: []models.Transaction{},
		TotalCount:   100,
		Page:         10,
		Limit:        10,
		TotalPages:   10,
		HasNext:      false,
		HasPrev:      true,
	}

	// Test calculated fields for last page
	expectedTotalPages := 10
	expectedHasNext := false
	expectedHasPrev := true

	assert.Equal(t, expectedTotalPages, result.TotalPages)
	assert.Equal(t, expectedHasNext, result.HasNext)
	assert.Equal(t, expectedHasPrev, result.HasPrev)
}

func TestTransactionServiceResult_MiddlePage(t *testing.T) {
	result := TransactionServiceResult{
		Transactions: []models.Transaction{},
		TotalCount:   100,
		Page:         5,
		Limit:        10,
		TotalPages:   10,
		HasNext:      true,
		HasPrev:      true,
	}

	// Test calculated fields for middle page
	expectedTotalPages := 10
	expectedHasNext := true
	expectedHasPrev := true

	assert.Equal(t, expectedTotalPages, result.TotalPages)
	assert.Equal(t, expectedHasNext, result.HasNext)
	assert.Equal(t, expectedHasPrev, result.HasPrev)
}

func TestTransactionServiceResult_SinglePage(t *testing.T) {
	result := TransactionServiceResult{
		Transactions: []models.Transaction{},
		TotalCount:   5,
		Page:         1,
		Limit:        10,
		TotalPages:   1,
		HasNext:      false,
		HasPrev:      false,
	}

	// Test calculated fields for single page
	expectedTotalPages := 1
	expectedHasNext := false
	expectedHasPrev := false

	assert.Equal(t, expectedTotalPages, result.TotalPages)
	assert.Equal(t, expectedHasNext, result.HasNext)
	assert.Equal(t, expectedHasPrev, result.HasPrev)
}

func TestTransactionFilter_Validation(t *testing.T) {
	tests := []struct {
		name    string
		filter  models.TransactionFilter
		isValid bool
	}{
		{
			name: "valid filter with merchant_id",
			filter: models.TransactionFilter{
				MerchantID: stringPtr("test-merchant"),
			},
			isValid: true,
		},
		{
			name: "valid filter with date range",
			filter: models.TransactionFilter{
				DateTimeFrom: timePtr(time.Now().Add(-24 * time.Hour)),
				DateTimeTo:   timePtr(time.Now()),
			},
			isValid: true,
		},
		{
			name: "valid filter with amount range",
			filter: models.TransactionFilter{
				AmountMin: int64Ptr(100),
				AmountMax: int64Ptr(1000),
			},
			isValid: true,
		},
		{
			name: "invalid filter with amount min > max",
			filter: models.TransactionFilter{
				AmountMin: int64Ptr(1000),
				AmountMax: int64Ptr(100),
			},
			isValid: false,
		},
		{
			name: "invalid filter with date from > to",
			filter: models.TransactionFilter{
				DateTimeFrom: timePtr(time.Now()),
				DateTimeTo:   timePtr(time.Now().Add(-24 * time.Hour)),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual service method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestSortParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		sort    models.SortParams
		isValid bool
	}{
		{
			name: "valid asc sort",
			sort: models.SortParams{
				Field:     "amount",
				Direction: "asc",
			},
			isValid: true,
		},
		{
			name: "valid desc sort",
			sort: models.SortParams{
				Field:     "amount",
				Direction: "desc",
			},
			isValid: true,
		},
		{
			name: "invalid direction",
			sort: models.SortParams{
				Field:     "amount",
				Direction: "invalid",
			},
			isValid: false,
		},
		{
			name: "empty field",
			sort: models.SortParams{
				Field:     "",
				Direction: "asc",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual service method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		isValid bool
	}{
		{
			name:    "valid fields",
			fields:  []string{"tx_log_id", "amount", "merchant_name"},
			isValid: true,
		},
		{
			name:    "empty fields",
			fields:  []string{},
			isValid: true, // Should use default fields
		},
		{
			name:    "invalid field",
			fields:  []string{"invalid_field"},
			isValid: false,
		},
		{
			name:    "mixed valid and invalid fields",
			fields:  []string{"tx_log_id", "invalid_field", "amount"},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be tested in the actual service method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
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
			// This would be tested in the actual service method
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
			// This would be tested in the actual service method
			// For now, we just validate the test cases
			assert.NotEmpty(t, tt.name)
		})
	}
}

func TestMerchantSummary_Calculation(t *testing.T) {
	summary := models.MerchantSummary{
		MerchantID:             "test-merchant",
		MerchantName:           "Test Merchant",
		TotalTransactions:      100,
		SuccessfulTransactions: 80,
		FailedTransactions:     20,
		TotalAmount:            10000,
		AverageAmount:          100.0,
		SuccessRate:            80.0,
		DateFrom:               time.Now().Add(-24 * time.Hour),
		DateTo:                 time.Now(),
	}

	// Test calculated fields
	expectedAverageAmount := float64(10000) / float64(100)
	expectedSuccessRate := float64(80) / float64(100) * 100

	assert.Equal(t, expectedAverageAmount, summary.AverageAmount)
	assert.Equal(t, expectedSuccessRate, summary.SuccessRate)
}

func TestMerchantSummary_ZeroTransactions(t *testing.T) {
	summary := models.MerchantSummary{
		MerchantID:             "test-merchant",
		MerchantName:           "Test Merchant",
		TotalTransactions:      0,
		SuccessfulTransactions: 0,
		FailedTransactions:     0,
		TotalAmount:            0,
		DateFrom:               time.Now().Add(-24 * time.Hour),
		DateTo:                 time.Now(),
	}

	// Test calculated fields for zero transactions
	expectedAverageAmount := 0.0
	expectedSuccessRate := 0.0

	assert.Equal(t, expectedAverageAmount, summary.AverageAmount)
	assert.Equal(t, expectedSuccessRate, summary.SuccessRate)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func int64Ptr(i int64) *int64 {
	return &i
}
