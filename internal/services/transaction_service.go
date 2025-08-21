package services

import (
	"fmt"
	"strings"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/database"
	"aken_reporting_service/internal/models"
	"aken_reporting_service/internal/repositories"
	"crypto/md5"
)

type TransactionService interface {
	GetTransactions(merchantID string, params *GetTransactionsParams) (*TransactionServiceResult, error)
	GetTransactionByID(merchantID, transactionID string, fields []string, timezone string, panFormat string) (*models.Transaction, error)
	SearchTransactions(merchantID string, searchReq *models.TransactionSearchRequest, timezone string, panFormat string) (*TransactionServiceResult, error)
	GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error)
	GetTransactionTotals(merchantID string, request models.TransactionTotalsRequest) (*models.TransactionTotalsResponse, error)
	ParseAdvancedFilter(filterString, timezone string) (*models.TransactionFilter, error)
	ParseSort(sortString string) ([]models.SortParams, error)
	ValidateFields(fields []string) error
}

type transactionService struct {
	transactionRepo repositories.TransactionRepository
	cacheService    CacheService
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
	Transactions     []models.Transaction `json:"data"`
	TotalCount       int64                `json:"total_count"`
	Page             int                  `json:"page"`
	Limit            int                  `json:"limit"`
	TotalPages       int                  `json:"total_pages"`
	CurrentPageCount int                  `json:"current_page_count"`
	HasNext          bool                 `json:"has_next"`
	HasPrev          bool                 `json:"has_prev"`
	RequestedFields  []string             `json:"-"` // Internal field, not serialized
}

func NewTransactionService(transactionRepo repositories.TransactionRepository, cacheService CacheService) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		cacheService:    cacheService,
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
	// Don't set default fields - empty fields means return all fields
	
	// Validate fields if any are specified
	if len(params.Fields) > 0 {
		if err := s.ValidateFields(params.Fields); err != nil {
			return nil, fmt.Errorf("invalid fields: %v", err)
		}
	}
	
	// Skip caching for transaction data to ensure fresh data
	// Transaction data changes frequently and users need latest information
	
	pagination := models.PaginationParams{
		Page:  params.Page,
		Limit: params.Limit,
	}
	
	// Use retry logic for database operations
	retryConfig := database.DefaultRetryConfig()
	
	var result *repositories.TransactionListResult
	err := database.RetryWithBackoff(func() error {
		var dbErr error
		result, dbErr = s.transactionRepo.GetTransactions(
			merchantID,
			params.Filter,
			params.Fields,
			params.Sort,
			pagination,
			params.Timezone,
			params.PANFormat,
		)
		return dbErr
	}, retryConfig)
	
	if err != nil {
		// Don't wrap the error to avoid exposing internal details
		return nil, err
	}
	
	// Return fresh transaction data without caching
	serviceResult := &TransactionServiceResult{
		Transactions:     result.Transactions,
		TotalCount:       result.TotalCount,
		Page:             result.Page,
		Limit:            result.Limit,
		TotalPages:       result.TotalPages,
		CurrentPageCount: len(result.Transactions),
		HasNext:          result.Page < result.TotalPages,
		HasPrev:          result.Page > 1,
		RequestedFields:  result.RequestedFields,
	}
	
	return serviceResult, nil
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
		return nil, err
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
	// Don't set default fields - let the repository handle field filtering
	// Empty fields means return all fields
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
		return nil, err
	}
	
	return &TransactionServiceResult{
		Transactions:     result.Transactions,
		TotalCount:       result.TotalCount,
		Page:             result.Page,
		Limit:            result.Limit,
		TotalPages:       result.TotalPages,
		CurrentPageCount: len(result.Transactions),
		HasNext:          result.Page < result.TotalPages,
		HasPrev:          result.Page > 1,
		RequestedFields:  result.RequestedFields,
	}, nil
}

// GetMerchantSummary calculates merchant summary statistics
func (s *transactionService) GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error) {
	// Generate cache key for merchant summary
	cacheKey := s.generateMerchantSummaryCacheKey(merchantID, filter)
	
	// Try to get from cache first (summaries can be cached)
	if s.cacheService != nil {
		if cachedSummary, err := s.cacheService.GetCachedMerchantSummary(cacheKey); err == nil && cachedSummary != nil {
			return cachedSummary, nil
		}
	}
	
	// Cache miss - calculate summary from database
	summary, err := s.transactionRepo.GetMerchantSummary(merchantID, filter)
	if err != nil {
		// Don't wrap the error to avoid exposing internal details
		return nil, err
	}
	
	// Cache the summary for 30 minutes (aggregated data is safe to cache)
	if s.cacheService != nil {
		ttl := config.GetRedisTTL()
		s.cacheService.SetCachedMerchantSummary(cacheKey, summary, ttl)
	}
	
	return summary, nil
}

// ParseAdvancedFilter parses filter string into TransactionFilter struct
func (s *transactionService) ParseAdvancedFilter(filterString, timezone string) (*models.TransactionFilter, error) {
	if filterString == "" {
		return &models.TransactionFilter{}, nil
	}
	
	filter := &models.TransactionFilter{}
	
	// Split by AND, but preserve parenthesized groups
	conditions := s.splitPreservingParentheses(filterString, " AND ")
	
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}
		
		// Handle OR conditions within this AND group
		// If condition starts with '(' and ends with ')', it's a parenthesized group
		if strings.HasPrefix(condition, "(") && strings.HasSuffix(condition, ")") {
			// Remove outer parentheses and split by OR
			innerCondition := strings.Trim(condition, "()")
			orConditions := strings.Split(innerCondition, " OR ")
			for _, orCondition := range orConditions {
				if err := s.parseCondition(strings.TrimSpace(orCondition), filter, timezone); err != nil {
					return nil, err
				}
			}
		} else {
			// Handle regular condition (may contain OR)
			orConditions := strings.Split(condition, " OR ")
			for _, orCondition := range orConditions {
				if err := s.parseCondition(strings.TrimSpace(orCondition), filter, timezone); err != nil {
					return nil, err
				}
			}
		}
	}
	
	return filter, nil
}

// splitPreservingParentheses splits a string by delimiter while preserving parenthesized groups
func (s *transactionService) splitPreservingParentheses(input, delimiter string) []string {
	var result []string
	var current strings.Builder
	parenCount := 0
	i := 0
	
	for i < len(input) {
		if input[i] == '(' {
			parenCount++
			current.WriteByte(input[i])
			i++
		} else if input[i] == ')' {
			parenCount--
			current.WriteByte(input[i])
			i++
		} else if parenCount == 0 && i+len(delimiter) <= len(input) && input[i:i+len(delimiter)] == delimiter {
			// Found delimiter outside parentheses
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			i += len(delimiter)
		} else {
			current.WriteByte(input[i])
			i++
		}
	}
	
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	
	return result
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
			// If the date is just a date (not a datetime), extend it to end of day
			if isDateOnly(value) {
				date = date.Add(24*time.Hour - time.Nanosecond)
			}
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
			// If the "to" date is just a date (not a datetime), extend it to end of day
			if isDateOnly(strings.TrimSpace(parts[1])) {
				to = to.Add(24*time.Hour - time.Nanosecond)
			}
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
	validFields["currency_info"] = true
	
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

// isDateOnly checks if the date string is just a date (YYYY-MM-DD) without time
func isDateOnly(dateStr string) bool {
	dateOnlyFormats := []string{
		"2006-01-02",
		"2006/01/02",
	}
	
	for _, format := range dateOnlyFormats {
		if _, err := time.Parse(format, dateStr); err == nil {
			return true
		}
	}
	return false
}

// generateTransactionCacheKey creates a unique cache key for transaction queries
func (s *transactionService) generateTransactionCacheKey(merchantID string, params *GetTransactionsParams) string {
	// Create a string representation of the parameters
	keyParts := []string{
		"transactions",
		merchantID,
		fmt.Sprintf("page:%d", params.Page),
		fmt.Sprintf("limit:%d", params.Limit),
		fmt.Sprintf("timezone:%s", params.Timezone),
		fmt.Sprintf("pan_format:%s", params.PANFormat),
	}
	
	// Add fields
	if len(params.Fields) > 0 {
		keyParts = append(keyParts, fmt.Sprintf("fields:%s", strings.Join(params.Fields, ",")))
	}
	
	// Add sort parameters
	if len(params.Sort) > 0 {
		sortParts := make([]string, len(params.Sort))
		for i, sort := range params.Sort {
			sortParts[i] = fmt.Sprintf("%s:%s", sort.Field, sort.Direction)
		}
		keyParts = append(keyParts, fmt.Sprintf("sort:%s", strings.Join(sortParts, ",")))
	}
	
	// Add filter parameters
	if params.Filter != nil {
		if params.Filter.MerchantID != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_merchant_id:%s", *params.Filter.MerchantID))
		}
		if params.Filter.DeviceID != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_device_id:%s", *params.Filter.DeviceID))
		}
		if params.Filter.ResponseCode != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_response_code:%s", *params.Filter.ResponseCode))
		}
		if params.Filter.CurrencyCode != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_currency_code:%s", *params.Filter.CurrencyCode))
		}
		if params.Filter.AmountMin != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_amount_min:%d", *params.Filter.AmountMin))
		}
		if params.Filter.AmountMax != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_amount_max:%d", *params.Filter.AmountMax))
		}
		if params.Filter.DateTimeFrom != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_date_from:%s", params.Filter.DateTimeFrom.Format(time.RFC3339)))
		}
		if params.Filter.DateTimeTo != nil {
			keyParts = append(keyParts, fmt.Sprintf("filter_date_to:%s", params.Filter.DateTimeTo.Format(time.RFC3339)))
		}
	}
	
	// Join all parts and create a hash
	key := strings.Join(keyParts, "|")
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	
	return fmt.Sprintf("%s:%s", config.GetRedisKeyPrefix(), hash[:16])
}

// generateMerchantSummaryCacheKey creates a unique cache key for merchant summary queries
func (s *transactionService) generateMerchantSummaryCacheKey(merchantID string, filter *models.TransactionFilter) string {
	// Create a string representation of the parameters
	keyParts := []string{
		"summary",
		merchantID,
	}
	
	// Add filter parameters if present
	if filter != nil {
		if filter.DateTimeFrom != nil {
			keyParts = append(keyParts, fmt.Sprintf("from:%s", filter.DateTimeFrom.Format(time.RFC3339)))
		}
		if filter.DateTimeTo != nil {
			keyParts = append(keyParts, fmt.Sprintf("to:%s", filter.DateTimeTo.Format(time.RFC3339)))
		}
		if filter.AmountMin != nil {
			keyParts = append(keyParts, fmt.Sprintf("min_amount:%d", *filter.AmountMin))
		}
		if filter.AmountMax != nil {
			keyParts = append(keyParts, fmt.Sprintf("max_amount:%d", *filter.AmountMax))
		}
	}
	
	// Join all parts and create a hash
	key := strings.Join(keyParts, "|")
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	
	return fmt.Sprintf("%s:%s", config.GetRedisKeyPrefix(), hash[:16])
}

// GetTransactionTotals retrieves transaction totals by type for a specific date and device/terminal
func (s *transactionService) GetTransactionTotals(merchantID string, request models.TransactionTotalsRequest) (*models.TransactionTotalsResponse, error) {
	// Validate the date format
	if _, err := time.Parse("2006-01-02", request.Date); err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	
	// Call repository method
	result, err := s.transactionRepo.GetTransactionTotals(merchantID, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction totals: %w", err)
	}
	
	return result, nil
}