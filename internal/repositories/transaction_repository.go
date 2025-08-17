package repositories

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/models"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	GetTransactions(merchantID string, filter *models.TransactionFilter, fields []string, sort []models.SortParams, pagination models.PaginationParams, timezone string, panFormat string) (*TransactionListResult, error)
	GetTransactionByID(merchantID, transactionID string, fields []string, timezone string, panFormat string) (*models.Transaction, error)
	GetTransactionCount(merchantID string, filter *models.TransactionFilter) (int64, error)
	GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error)
	SearchTransactions(merchantID string, searchReq *models.TransactionSearchRequest, timezone string, panFormat string) (*TransactionListResult, error)
}

type transactionRepository struct {
	db *gorm.DB
}

type TransactionListResult struct {
	Transactions []models.Transaction `json:"data"`
	TotalCount   int64                `json:"total_count"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
	TotalPages   int                  `json:"total_pages"`
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// GetTransactions retrieves filtered, sorted, and paginated transactions
func (r *transactionRepository) GetTransactions(merchantID string, filter *models.TransactionFilter, fields []string, sort []models.SortParams, pagination models.PaginationParams, timezone string, panFormat string) (*TransactionListResult, error) {
	var transactions []models.Transaction
	
	// Build the query
	query := r.buildBaseQuery(fields, timezone, panFormat)
	
	// Apply merchant filter
	query = query.Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID)
	
	// Apply additional filters
	query = r.applyFilters(query, filter)
	
	// Apply sorting
	query = r.applySorting(query, sort)
	
	// Get total count for pagination
	var totalCount int64
	countQuery := r.buildCountQuery()
	countQuery = countQuery.Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID)
	countQuery = r.applyFilters(countQuery, filter)
	
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, err
	}
	
	// Apply pagination
	offset := (pagination.Page - 1) * pagination.Limit
	query = query.Limit(pagination.Limit).Offset(offset)
	
	// Execute query
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	
	// Post-process results
	r.postProcessTransactions(transactions)
	
	totalPages := int((totalCount + int64(pagination.Limit) - 1) / int64(pagination.Limit))
	
	return &TransactionListResult{
		Transactions: transactions,
		TotalCount:   totalCount,
		Page:         pagination.Page,
		Limit:        pagination.Limit,
		TotalPages:   totalPages,
	}, nil
}

// GetTransactionByID retrieves a single transaction by ID
func (r *transactionRepository) GetTransactionByID(merchantID, transactionID string, fields []string, timezone string, panFormat string) (*models.Transaction, error) {
	var transaction models.Transaction
	
	query := r.buildBaseQuery(fields, timezone, panFormat)
	query = query.Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID)
	query = query.Where("p.payment_tx_log_id = ?", transactionID)
	
	if err := query.First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	// Post-process single transaction
	r.postProcessTransactions([]models.Transaction{transaction})
	
	return &transaction, nil
}

// GetTransactionCount returns the total count of transactions matching the filter
func (r *transactionRepository) GetTransactionCount(merchantID string, filter *models.TransactionFilter) (int64, error) {
	query := r.buildCountQuery()
	query = query.Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID)
	query = r.applyFilters(query, filter)
	
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	
	return count, nil
}

// GetMerchantSummary calculates summary statistics for a merchant
func (r *transactionRepository) GetMerchantSummary(merchantID string, filter *models.TransactionFilter) (*models.MerchantSummary, error) {
	type summaryResult struct {
		MerchantID     string  `gorm:"column:merchant_id"`
		MerchantName   string  `gorm:"column:merchant_name"`
		TotalTxns      int     `gorm:"column:total_transactions"`
		SuccessfulTxns int     `gorm:"column:successful_transactions"`
		TotalAmount    int64   `gorm:"column:total_amount"`
		MinDate        *time.Time `gorm:"column:min_date"`
		MaxDate        *time.Time `gorm:"column:max_date"`
	}
	
	var result summaryResult
	
	query := r.db.Table("payment_tx_log p").
		Select(`
			m.merchant_id,
			m.name as merchant_name,
			COUNT(*) as total_transactions,
			SUM(CASE WHEN p.result_code IN ('00', '10') THEN 1 ELSE 0 END) as successful_transactions,
			SUM(COALESCE(p.amount, 0)) as total_amount,
			MIN(p.updated_at) as min_date,
			MAX(p.updated_at) as max_date
		`).
		Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id").
		Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID).
		Group("m.merchant_id, m.name")
	
	query = r.applyFilters(query, filter)
	
	if err := query.First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.MerchantSummary{
				MerchantID: merchantID,
				MerchantName: "Unknown",
			}, nil
		}
		return nil, err
	}
	
	summary := &models.MerchantSummary{
		MerchantID:             result.MerchantID,
		MerchantName:           result.MerchantName,
		TotalTransactions:      result.TotalTxns,
		SuccessfulTransactions: result.SuccessfulTxns,
		FailedTransactions:     result.TotalTxns - result.SuccessfulTxns,
		TotalAmount:            result.TotalAmount,
	}
	
	if result.TotalTxns > 0 {
		summary.AverageAmount = float64(result.TotalAmount) / float64(result.TotalTxns)
		summary.SuccessRate = (float64(result.SuccessfulTxns) / float64(result.TotalTxns)) * 100
	}
	
	if result.MinDate != nil {
		summary.DateFrom = *result.MinDate
	}
	if result.MaxDate != nil {
		summary.DateTo = *result.MaxDate
	}
	
	return summary, nil
}

// SearchTransactions performs advanced search with complex query body
func (r *transactionRepository) SearchTransactions(merchantID string, searchReq *models.TransactionSearchRequest, timezone string, panFormat string) (*TransactionListResult, error) {
	// For now, convert the search request to basic filters
	// In a full implementation, this would parse Elasticsearch-style queries
	filter := r.convertSearchToFilter(searchReq.Query)
	
	return r.GetTransactions(merchantID, filter, searchReq.Fields, searchReq.Sort, searchReq.Pagination, timezone, panFormat)
}

// buildBaseQuery constructs the base query with joins and field selection
func (r *transactionRepository) buildBaseQuery(fields []string, timezone string, panFormat string) *gorm.DB {
	if len(fields) == 0 {
		fields = config.DefaultFields
	}
	
	selectedFields := r.buildFieldSelection(fields, timezone, panFormat)
	
	query := r.db.Table("payment_tx_log p").
		Select(selectedFields).
		Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id")
	
	return query
}

// buildCountQuery constructs a query for counting records
func (r *transactionRepository) buildCountQuery() *gorm.DB {
	return r.db.Table("payment_tx_log p").
		Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id")
}

// buildFieldSelection creates the SELECT clause based on requested fields
func (r *transactionRepository) buildFieldSelection(fields []string, timezone string, panFormat string) string {
	var selectedFields []string
	
	panFormatSQL := config.PANFormats[panFormat]
	if panFormatSQL == "" {
		panFormatSQL = config.PANFormats["bin_id_and_pan_id"]
	}
	
	for _, field := range fields {
		switch field {
		case "pan":
			selectedFields = append(selectedFields, fmt.Sprintf("%s as pan", panFormatSQL))
		case "tx_date_time":
			selectedFields = append(selectedFields, fmt.Sprintf("TO_CHAR(TIMEZONE('%s', p.updated_at), 'YYYY-MM-DD\"T\"HH24:MI:SS.MS\"Z\"') as tx_date_time", timezone))
		case "tx_log_type":
			selectedFields = append(selectedFields, fmt.Sprintf("%s as tx_log_type", config.FieldMappings["tx_log_type"]))
		default:
			if sqlField, exists := config.FieldMappings[field]; exists {
				selectedFields = append(selectedFields, fmt.Sprintf("%s as %s", sqlField, field))
			} else {
				selectedFields = append(selectedFields, fmt.Sprintf("p.%s", field))
			}
		}
	}
	
	return strings.Join(selectedFields, ", ")
}

// applyFilters adds WHERE conditions based on the filter
func (r *transactionRepository) applyFilters(query *gorm.DB, filter *models.TransactionFilter) *gorm.DB {
	if filter == nil {
		return query
	}
	
	// Note: DeviceID filtering is disabled since we're not joining with devices table
	// if filter.DeviceID != nil {
	// 	query = query.Where("d.deviceid = ?", *filter.DeviceID)
	// }
	
	if filter.ResponseCode != nil {
		query = query.Where("p.result_code = ?", *filter.ResponseCode)
	}
	
	if filter.DateTimeFrom != nil {
		query = query.Where("p.updated_at >= ?", *filter.DateTimeFrom)
	}
	
	if filter.DateTimeTo != nil {
		query = query.Where("p.updated_at <= ?", *filter.DateTimeTo)
	}
	
	if filter.CurrencyCode != nil {
		query = query.Where("p.currency_code = ?", *filter.CurrencyCode)
	}
	
	if filter.AmountMin != nil {
		query = query.Where("p.amount >= ?", *filter.AmountMin)
	}
	
	if filter.AmountMax != nil {
		query = query.Where("p.amount <= ?", *filter.AmountMax)
	}
	
	if filter.TxLogType != nil {
		// Convert string type to numeric type
		var typeID string
		switch *filter.TxLogType {
		case "payment":
			typeID = "0"
		case "reversal":
			typeID = "1"
		case "void":
			typeID = "2"
		case "refund":
			typeID = "3"
		case "mm purchase":
			typeID = "9"
		case "mm refund":
			typeID = "10"
		}
		if typeID != "" {
			query = query.Where("p.payment_tx_type_id = ?", typeID)
		}
	}
	
	return query
}

// applySorting adds ORDER BY clauses
func (r *transactionRepository) applySorting(query *gorm.DB, sort []models.SortParams) *gorm.DB {
	if len(sort) == 0 {
		// Default sort
		return query.Order("p.updated_at DESC")
	}
	
	for _, s := range sort {
		direction := "ASC"
		if strings.ToLower(s.Direction) == "desc" {
			direction = "DESC"
		}
		
		if sqlField, exists := config.FieldMappings[s.Field]; exists {
			query = query.Order(fmt.Sprintf("%s %s", sqlField, direction))
		} else {
			query = query.Order(fmt.Sprintf("p.%s %s", s.Field, direction))
		}
	}
	
	return query
}

// postProcessTransactions performs any necessary post-processing on transactions
func (r *transactionRepository) postProcessTransactions(transactions []models.Transaction) {
	for i := range transactions {
		tx := &transactions[i]
		
		// Set type string based on payment_tx_type_id
		tx.Type = tx.GetTypeString()
		
		// Set reversed flag
		tx.Reversed = tx.IsReversed()
		
		// Extract user_ref from meta if needed
		if tx.Meta != nil {
			// Parse meta JSON to extract reference
			var meta map[string]interface{}
			if err := json.Unmarshal(tx.Meta, &meta); err == nil {
				if ref, ok := meta["reference"].(string); ok && ref != "" {
					tx.UserRef = &ref
				}
			}
		}
	}
}

// convertSearchToFilter converts search query to basic filter (simplified implementation)
func (r *transactionRepository) convertSearchToFilter(query interface{}) *models.TransactionFilter {
	// This is a simplified implementation
	// In a real implementation, you'd parse Elasticsearch-style queries
	filter := &models.TransactionFilter{}
	
	if queryMap, ok := query.(map[string]interface{}); ok {
		if boolQuery, exists := queryMap["bool"]; exists {
			if boolMap, ok := boolQuery.(map[string]interface{}); ok {
				if mustClauses, exists := boolMap["must"]; exists {
					if clauses, ok := mustClauses.([]interface{}); ok {
						for _, clause := range clauses {
							r.parseClause(clause, filter)
						}
					}
				}
			}
		}
	}
	
	return filter
}

// parseClause parses individual query clauses (simplified)
func (r *transactionRepository) parseClause(clause interface{}, filter *models.TransactionFilter) {
	if clauseMap, ok := clause.(map[string]interface{}); ok {
		// Parse term queries
		if term, exists := clauseMap["term"]; exists {
			if termMap, ok := term.(map[string]interface{}); ok {
				for field, value := range termMap {
					switch field {
					case "merchant_id":
						if strVal, ok := value.(string); ok {
							filter.MerchantID = &strVal
						}
					case "response_code":
						if strVal, ok := value.(string); ok {
							filter.ResponseCode = &strVal
						}
					}
				}
			}
		}
		
		// Parse range queries  
		if rangeQuery, exists := clauseMap["range"]; exists {
			if rangeMap, ok := rangeQuery.(map[string]interface{}); ok {
				for field, value := range rangeMap {
					if field == "amount" {
						if valueMap, ok := value.(map[string]interface{}); ok {
							if gte, exists := valueMap["gte"]; exists {
								if intVal, ok := gte.(float64); ok {
									val := int64(intVal)
									filter.AmountMin = &val
								}
							}
							if lte, exists := valueMap["lte"]; exists {
								if intVal, ok := lte.(float64); ok {
									val := int64(intVal)
									filter.AmountMax = &val
								}
							}
						}
					}
				}
			}
		}
	}
}