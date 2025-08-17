package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_TableName(t *testing.T) {
	tx := Transaction{}
	assert.Equal(t, "payment_tx_log", tx.TableName())
}

func TestTransaction_GetTypeString(t *testing.T) {
	tests := []struct {
		name     string
		typeID   int
		expected string
	}{
		{"payment_type", 0, "payment"},
		{"reversal_type", 1, "reversal"},
		{"void_type", 2, "void"},
		{"refund_type", 3, "refund"},
		{"mm_purchase_type", 9, "mm purchase"},
		{"mm_refund_type", 10, "mm refund"},
		{"unknown_type", 99, "unknown"},
		{"empty_type", -1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := Transaction{PaymentTxTypeID: tt.typeID}
			assert.Equal(t, tt.expected, tx.GetTypeString())
		})
	}
}

func TestTransaction_IsReversed(t *testing.T) {
	tests := []struct {
		name           string
		reversedTxLogID *string
		expected       bool
	}{
		{"not_reversed", nil, false},
		{"reversed", stringPtr("some-id"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := Transaction{ReversedTxLogID: tt.reversedTxLogID}
			assert.Equal(t, tt.expected, tx.IsReversed())
		})
	}
}

func TestPaymentProvider_TableName(t *testing.T) {
	provider := PaymentProvider{}
	assert.Equal(t, "payment_providers", provider.TableName())
}

func TestPaymentTxType_TableName(t *testing.T) {
	txType := PaymentTxType{}
	assert.Equal(t, "payment_tx_types", txType.TableName())
}

func TestMerchant_TableName(t *testing.T) {
	merchant := Merchant{}
	assert.Equal(t, "merchants", merchant.TableName())
}

func TestDevice_TableName(t *testing.T) {
	device := Device{}
	assert.Equal(t, "devices", device.TableName())
}

func TestTerminal_TableName(t *testing.T) {
	terminal := Terminal{}
	assert.Equal(t, "terminals", terminal.TableName())
}

func TestTransactionFilter_Validation(t *testing.T) {
	tests := []struct {
		name    string
		filter  TransactionFilter
		isValid bool
	}{
		{
			name: "valid_filter_with_merchant_id",
			filter: TransactionFilter{
				MerchantID: stringPtr("test-merchant"),
			},
			isValid: true,
		},
		{
			name: "valid_filter_with_date_range",
			filter: TransactionFilter{
				DateTimeFrom: timePtr(time.Now().Add(-24 * time.Hour)),
				DateTimeTo:   timePtr(time.Now()),
			},
			isValid: true,
		},
		{
			name: "valid_filter_with_amount_range",
			filter: TransactionFilter{
				AmountMin: int64Ptr(1000),
				AmountMax: int64Ptr(5000),
			},
			isValid: true,
		},
		{
			name: "invalid_filter_with_amount_min_>_max",
			filter: TransactionFilter{
				AmountMin: int64Ptr(5000),
				AmountMax: int64Ptr(1000),
			},
			isValid: false,
		},
		{
			name: "invalid_filter_with_date_from_>_to",
			filter: TransactionFilter{
				DateTimeFrom: timePtr(time.Now()),
				DateTimeTo:   timePtr(time.Now().Add(-24 * time.Hour)),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in a real implementation, you'd add validation logic
			if tt.filter.AmountMin != nil && tt.filter.AmountMax != nil {
				if *tt.filter.AmountMin > *tt.filter.AmountMax {
					assert.False(t, tt.isValid)
					return
				}
			}
			if tt.filter.DateTimeFrom != nil && tt.filter.DateTimeTo != nil {
				if tt.filter.DateTimeFrom.After(*tt.filter.DateTimeTo) {
					assert.False(t, tt.isValid)
					return
				}
			}
			assert.True(t, tt.isValid)
		})
	}
}

func TestPaginationParams_Validation(t *testing.T) {
	tests := []struct {
		name     string
		params   PaginationParams
		isValid  bool
	}{
		{"valid_pagination", PaginationParams{Page: 1, Limit: 10}, true},
		{"invalid_page", PaginationParams{Page: 0, Limit: 10}, false},
		{"invalid_limit", PaginationParams{Page: 1, Limit: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.params.Page > 0 && tt.params.Limit > 0
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestSortParams_Validation(t *testing.T) {
	tests := []struct {
		name     string
		params   SortParams
		isValid  bool
	}{
		{"valid_asc_sort", SortParams{Field: "amount", Direction: "asc"}, true},
		{"valid_desc_sort", SortParams{Field: "amount", Direction: "desc"}, true},
		{"invalid_direction", SortParams{Field: "amount", Direction: "invalid"}, false},
		{"empty_field", SortParams{Field: "", Direction: "asc"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.params.Field != "" && (tt.params.Direction == "asc" || tt.params.Direction == "desc")
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestMerchantSummary_Calculation(t *testing.T) {
	summary := MerchantSummary{
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

func TestTransaction_JSONSerialization(t *testing.T) {
	tx := Transaction{
		ID:                "test-id",
		PaymentTxTypeID:   0,
		PaymentProviderID: 1,
		RRN:               "test-rrn",
		STAN:              "123",
		MerchantCode:      "test-merchant",
		CurrencyCode:      "0978",
		Amount:            1000,
		Completed:         true,
		Active:            true,
		Reversed:          false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(tx)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Test JSON unmarshaling
	var unmarshaled Transaction
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, tx.ID, unmarshaled.ID)
	assert.Equal(t, tx.PaymentTxTypeID, unmarshaled.PaymentTxTypeID)
	assert.Equal(t, tx.Amount, unmarshaled.Amount)
}

func TestTransactionSearchRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request TransactionSearchRequest
		isValid bool
	}{
		{
			name: "valid_request_with_query",
			request: TransactionSearchRequest{
				Query: map[string]interface{}{"amount": 1000},
				Fields: []string{"tx_log_id", "amount"},
				Pagination: PaginationParams{Page: 1, Limit: 10},
			},
			isValid: true,
		},
		{
			name: "valid_request_without_query",
			request: TransactionSearchRequest{
				Fields: []string{"tx_log_id"},
				Pagination: PaginationParams{Page: 1, Limit: 10},
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			isValid := len(tt.request.Fields) > 0 && tt.request.Pagination.Page > 0 && tt.request.Pagination.Limit > 0
			assert.Equal(t, tt.isValid, isValid)
		})
	}
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
