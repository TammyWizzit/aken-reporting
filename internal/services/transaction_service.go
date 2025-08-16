package services

import (
	"fmt"
	"strings"
	"time"

	"aken_reporting_service/internal/models"
	"aken_reporting_service/internal/repositories"
	"aken_reporting_service/internal/config"
)

type TransactionService interface {
	GetTransactions(merchantID string, params *GetTransactionsParams) (*TransactionServiceResult, error)
	GetTransactionByID(merchantID, transactionID string, fields []string, timezone string, panFormat string) (*models.Transaction, error)
	SearchTransactions(merchantID string, searchReq *models.TransactionSearchRequest, timezone string, panFormat string) (*TransactionServiceResult, error)
	GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error)
	ParseAdvancedFilter(filterString, timezone string) (*models.TransactionFilter, error)
	ParseSort(sortString string) ([]models.SortParams, error)
	ValidateFields(fields []string) error
}

type transactionService struct {
	transactionRepo repositories.TransactionRepository
}

type GetTransactionsParams struct {
	Filter    *models.TransactionFilter
	Fields    []string
	Sort      []models.SortParams
	Page      int
	Limit     int
	Timezone  string
	PANFormat string
}

type TransactionServiceResult struct {
	Transactions []models.Transaction `json:"data"`
	TotalCount   int64                `json:"total_count"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
	TotalPages   int                  `json:"total_pages"`
	HasNext      bool                 `json:"has_next"`
	HasPrev      bool                 `json:"has_prev"`
}

func NewTransactionService(transactionRepo repositories.TransactionRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

// GetTransactions retrieves filtered, sorted, and paginated transactions
func (s *transactionService) GetTransactions(merchantID string, params *GetTransactionsParams) (*TransactionServiceResult, error) {
	// Validate and set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < config.MinPageSize {
		params.Limit = config.DefaultPageSize
	}
	if params.Limit > config.MaxPageSize {
		params.Limit = config.MaxPageSize
	}
	if params.Timezone == "" {
		params.Timezone = "UTC"
	}
	if params.PANFormat == "" {
		params.PANFormat = "bin_id_and_pan_id"
	}
	if len(params.Fields) == 0 {
		params.Fields = config.DefaultFields
	}
	
	// Validate fields
	if err := s.ValidateFields(params.Fields); err != nil {
		return nil, fmt.Errorf("invalid fields: %v", err)
	}
	
	pagination := models.PaginationParams{
		Page:  params.Page,
		Limit: params.Limit,
	}
	
	result, err := s.transactionRepo.GetTransactions(
		merchantID,
		params.Filter,
		params.Fields,
		params.Sort,
		pagination,
		params.Timezone,
		params.PANFormat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %v", err)
	}
	
	return &TransactionServiceResult{
		Transactions: result.Transactions,
		TotalCount:   result.TotalCount,
		Page:         result.Page,
		Limit:        result.Limit,
		TotalPages:   result.TotalPages,
		HasNext:      result.Page < result.TotalPages,
		HasPrev:      result.Page > 1,
	}, nil
}

// GetTransactionByID retrieves a single transaction by ID
func (s *transactionService) GetTransactionByID(merchantID, transactionID string, fields []string, timezone string, panFormat string) (*models.Transaction, error) {
	if len(fields) == 0 {
		fields = config.DefaultFields
	}
	if timezone == "" {
		timezone = "UTC"
	}
	if panFormat == "" {
		panFormat = "bin_id_and_pan_id"
	}
	
	// Validate fields
	if err := s.ValidateFields(fields); err != nil {
		return nil, fmt.Errorf("invalid fields: %v", err)
	}
	
	transaction, err := s.transactionRepo.GetTransactionByID(merchantID, transactionID, fields, timezone, panFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %v", err)
	}
	
	return transaction, nil
}

// SearchTransactions performs advanced search
func (s *transactionService) SearchTransactions(merchantID string, searchReq *models.TransactionSearchRequest, timezone string, panFormat string) (*TransactionServiceResult, error) {
	// Set defaults
	if searchReq.Pagination.Page < 1 {
		searchReq.Pagination.Page = 1
	}
	if searchReq.Pagination.Limit < config.MinPageSize {
		searchReq.Pagination.Limit = config.DefaultPageSize
	}
	if searchReq.Pagination.Limit > config.MaxPageSize {
		searchReq.Pagination.Limit = config.MaxPageSize
	}
	if len(searchReq.Fields) == 0 {
		searchReq.Fields = config.DefaultFields
	}
	if timezone == "" {
		timezone = "UTC"
	}
	if panFormat == "" {
		panFormat = "bin_id_and_pan_id"
	}
	
	// Validate fields
	if err := s.ValidateFields(searchReq.Fields); err != nil {
		return nil, fmt.Errorf("invalid fields: %v", err)
	}
	
	result, err := s.transactionRepo.SearchTransactions(merchantID, searchReq, timezone, panFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to search transactions: %v", err)
	}
	
	return &TransactionServiceResult{
		Transactions: result.Transactions,
		TotalCount:   result.TotalCount,
		Page:         result.Page,
		Limit:        result.Limit,
		TotalPages:   result.TotalPages,
		HasNext:      result.Page < result.TotalPages,
		HasPrev:      result.Page > 1,
	}, nil
}

// GetMerchantSummary calculates merchant summary statistics
func (s *transactionService) GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error) {
	summary, err := s.transactionRepo.GetMerchantSummary(merchantID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant summary: %v", err)
	}
	
	return summary, nil
}

// ParseAdvancedFilter parses filter string into TransactionFilter struct
func (s *transactionService) ParseAdvancedFilter(filterString, timezone string) (*models.TransactionFilter, error) {
	if filterString == "" {
		return &models.TransactionFilter{}, nil
	}
	
	filter := &models.TransactionFilter{}
	
	// Split by AND/OR (simplified parser)
	conditions := strings.Split(filterString, " AND ")
	
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}
		
		// Handle OR conditions within this AND group
		orConditions := strings.Split(condition, " OR ")
		for _, orCondition := range orConditions {
			if err := s.parseCondition(strings.TrimSpace(orCondition), filter, timezone); err != nil {
				return nil, err
			}
		}
	}
	
	return filter, nil
}

// parseCondition parses a single filter condition
func (s *transactionService) parseCondition(condition string, filter *models.TransactionFilter, timezone string) error {
	parts := strings.Split(condition, ":")
	if len(parts) < 3 {
		return fmt.Errorf("invalid filter condition: %s", condition)
	}
	
	field := parts[0]
	operator := parts[1]
	value := strings.Join(parts[2:], ":")
	
	// Validate operator
	if _, exists := config.FilterOperators[operator]; !exists {
		return fmt.Errorf("invalid operator '%s' for field '%s'", operator, field)
	}
	
	switch field {
	case "merchant_id":
		if operator == "eq" {
			filter.MerchantID = &value
		}
	case "device_id":
		if operator == "eq" {
			filter.DeviceID = &value
		}
	case "response_code":
		if operator == "eq" {
			filter.ResponseCode = &value
		}
	case "currency_code":
		if operator == "eq" {
			filter.CurrencyCode = &value
		}
	case "tx_log_type":
		if operator == "eq" {
			filter.TxLogType = &value
		}
	case "amount":
		return s.parseAmountCondition(operator, value, filter)
	case "tx_date_time":
		return s.parseDateCondition(operator, value, filter, timezone)
	default:
		return fmt.Errorf("unsupported filter field: %s", field)
	}
	
	return nil
}

// parseAmountCondition parses amount-related conditions
func (s *transactionService) parseAmountCondition(operator, value string, filter *models.TransactionFilter) error {
	switch operator {
	case "gte":
		if amount, err := parseAmount(value); err == nil {
			filter.AmountMin = &amount
		} else {
			return fmt.Errorf("invalid amount value: %s", value)
		}
	case "lte":
		if amount, err := parseAmount(value); err == nil {
			filter.AmountMax = &amount
		} else {
			return fmt.Errorf("invalid amount value: %s", value)
		}
	case "between":
		parts := strings.Split(value, ",")
		if len(parts) != 2 {
			return fmt.Errorf("between operator requires two comma-separated values")
		}
		if min, err := parseAmount(strings.TrimSpace(parts[0])); err == nil {
			filter.AmountMin = &min
		} else {
			return fmt.Errorf("invalid min amount: %s", parts[0])
		}
		if max, err := parseAmount(strings.TrimSpace(parts[1])); err == nil {
			filter.AmountMax = &max
		} else {
			return fmt.Errorf("invalid max amount: %s", parts[1])
		}
	}
	
	return nil
}

// parseDateCondition parses date-related conditions
func (s *transactionService) parseDateCondition(operator, value string, filter *models.TransactionFilter, timezone string) error {
	switch operator {
	case "gte":
		if date, err := parseDateTime(value); err == nil {
			filter.DateTimeFrom = &date
		} else {
			return fmt.Errorf("invalid date format: %s", value)
		}
	case "lte":
		if date, err := parseDateTime(value); err == nil {
			filter.DateTimeTo = &date
		} else {
			return fmt.Errorf("invalid date format: %s", value)
		}
	case "between":
		parts := strings.Split(value, ",")
		if len(parts) != 2 {
			return fmt.Errorf("between operator requires two comma-separated values")
		}
		if from, err := parseDateTime(strings.TrimSpace(parts[0])); err == nil {
			filter.DateTimeFrom = &from
		} else {
			return fmt.Errorf("invalid from date: %s", parts[0])
		}
		if to, err := parseDateTime(strings.TrimSpace(parts[1])); err == nil {
			filter.DateTimeTo = &to
		} else {
			return fmt.Errorf("invalid to date: %s", parts[1])
		}
	}
	
	return nil
}

// ParseSort parses sort string into SortParams slice
func (s *transactionService) ParseSort(sortString string) ([]models.SortParams, error) {
	if sortString == "" {
		return []models.SortParams{{Field: "tx_date_time", Direction: "desc"}}, nil
	}
	
	var sortParams []models.SortParams
	parts := strings.Split(sortString, ",")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		fieldDirection := strings.Split(part, ":")
		field := fieldDirection[0]
		direction := "asc"
		
		if len(fieldDirection) > 1 {
			direction = strings.ToLower(fieldDirection[1])
		}
		
		if direction != "asc" && direction != "desc" {
			return nil, fmt.Errorf("invalid sort direction '%s' for field '%s'", direction, field)
		}
		
		// Validate field exists in mappings
		if _, exists := config.FieldMappings[field]; !exists && field != "tx_date_time" {
			return nil, fmt.Errorf("invalid sort field: %s", field)
		}
		
		sortParams = append(sortParams, models.SortParams{
			Field:     field,
			Direction: direction,
		})
	}
	
	return sortParams, nil
}

// ValidateFields validates that all requested fields are valid
func (s *transactionService) ValidateFields(fields []string) error {
	validFields := make(map[string]bool)
	for field := range config.FieldMappings {
		validFields[field] = true
	}
	
	// Add some additional valid fields
	validFields["created_at"] = true
	validFields["updated_at"] = true
	validFields["description"] = true
	
	for _, field := range fields {
		if !validFields[field] {
			return fmt.Errorf("invalid field: %s", field)
		}
	}
	
	return nil
}

// Helper functions

func parseAmount(value string) (int64, error) {
	// In cents - multiply by 100 if it looks like decimal
	if strings.Contains(value, ".") {
		var amount float64
		if _, err := fmt.Sscanf(value, "%f", &amount); err != nil {
			return 0, err
		}
		return int64(amount * 100), nil
	}
	
	var amount int64
	if _, err := fmt.Sscanf(value, "%d", &amount); err != nil {
		return 0, err
	}
	return amount, nil
}

func parseDateTime(value string) (time.Time, error) {
	// Try different date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unsupported date format: %s", value)
}