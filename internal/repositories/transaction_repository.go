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
	GetTransactionTotals(merchantID string, request models.TransactionTotalsRequest) (*models.TransactionTotalsResponse, error)
	GetTransactionLookup(request models.TransactionLookupRequest) (*models.TransactionLookupResponse, error)
	SearchTransactionDetails(request models.IsoTransactionSearchRequest) (*models.IsoTransactionSearchResponse, error)
	SetUseMysql(useMysql bool) // Add method to set MySQL flag
}

type transactionRepository struct {
	postgresDB *gorm.DB  // For v2 APIs
	mysqlDB    *gorm.DB  // For v1 efinance APIs
	useMysql   bool      // Flag to determine which database to use
}

type TransactionListResult struct {
	Transactions    []models.Transaction `json:"data"`
	TotalCount      int64                `json:"total_count"`
	Page            int                  `json:"page"`
	Limit           int                  `json:"limit"`
	TotalPages      int                  `json:"total_pages"`
	RequestedFields []string             `json:"-"` // Internal field, not serialized
}

func NewTransactionRepository(postgresDB *gorm.DB, mysqlDB *gorm.DB) TransactionRepository {
	return &transactionRepository{
		postgresDB: postgresDB,
		mysqlDB:    mysqlDB,
		useMysql:   false, // Default to PostgreSQL
	}
}

// SetUseMysql sets the flag to determine which database to use
func (r *transactionRepository) SetUseMysql(useMysql bool) {
	r.useMysql = useMysql
}

// getDB returns the appropriate database based on the current flag
func (r *transactionRepository) getDB() *gorm.DB {
	if r.useMysql && r.mysqlDB != nil {
		return r.mysqlDB
	}
	// Default to PostgreSQL
	if r.postgresDB != nil {
		return r.postgresDB
	}
	// Fallback to MySQL if PostgreSQL is not available
	return r.mysqlDB
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

	// Apply sorting - for DISTINCT ON queries, we need special handling
	query = r.applySortingWithDistinct(query, sort)

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

	// Debug: Log the SQL query
	// sql := r.getDB().ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return query.Find(&transactions)
	// })
	// fmt.Printf("DEBUG SQL: %s\n", sql)

	// Execute query
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}

	// Post-process results
	r.postProcessTransactions(transactions)

	totalPages := int((totalCount + int64(pagination.Limit) - 1) / int64(pagination.Limit))

	return &TransactionListResult{
		Transactions:    transactions,
		TotalCount:      totalCount,
		Page:            pagination.Page,
		Limit:           pagination.Limit,
		TotalPages:      totalPages,
		RequestedFields: fields,
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
		MerchantID     string     `gorm:"column:merchant_id"`
		MerchantName   string     `gorm:"column:merchant_name"`
		TotalTxns      int        `gorm:"column:total_transactions"`
		SuccessfulTxns int        `gorm:"column:successful_transactions"`
		TotalAmount    int64      `gorm:"column:total_amount"`
		MinDate        *time.Time `gorm:"column:min_date"`
		MaxDate        *time.Time `gorm:"column:max_date"`
	}

	var result summaryResult

	query := r.getDB().Table("payment_tx_log p").
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

	if err := query.Take(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.MerchantSummary{
				MerchantID:   merchantID,
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
	var query *gorm.DB

	if len(fields) == 0 {
		// No field filtering - select all fields plus computed fields from joins
		query = r.getDB().Table("payment_tx_log p").
			Select("DISTINCT ON (p.payment_tx_log_id) p.*, m.name as merchant_name, c.curr_short as currency_name, c.curr_delim").
			Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id").
			Joins("LEFT JOIN currency c ON p.currency_code = c.curr_code")
	} else {
		// Field filtering requested - select only specific fields
		selectedFields := r.buildFieldSelection(fields, timezone, panFormat)
		query = r.getDB().Table("payment_tx_log p").
			Select("DISTINCT ON (p.payment_tx_log_id) " + selectedFields).
			Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id").
			Joins("LEFT JOIN currency c ON p.currency_code = c.curr_code")
	}

	return query
}

// buildCountQuery constructs a query for counting records
func (r *transactionRepository) buildCountQuery() *gorm.DB {
	return r.getDB().Table("payment_tx_log p").
		Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id").
		Joins("LEFT JOIN currency c ON p.currency_code = c.curr_code")
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
		case "payment_tx_log_id":
			selectedFields = append(selectedFields, "p.payment_tx_log_id")
		case "amount":
			selectedFields = append(selectedFields, "p.amount")
		case "merchant_name":
			selectedFields = append(selectedFields, "m.name as merchant_name")
		case "response_code":
			selectedFields = append(selectedFields, "p.result_code as response_code")
		case "pan":
			selectedFields = append(selectedFields, fmt.Sprintf("%s as pan", panFormatSQL))
		case "tx_date_time":
			selectedFields = append(selectedFields, fmt.Sprintf("TO_CHAR(TIMEZONE('%s', p.updated_at), 'YYYY-MM-DD\"T\"HH24:MI:SS.MS\"Z\"') as tx_date_time", timezone))
		case "tx_log_type":
			selectedFields = append(selectedFields, fmt.Sprintf("%s as tx_log_type", config.FieldMappings["tx_log_type"]))
		case "currency_info":
			selectedFields = append(selectedFields, "p.currency_code")
			selectedFields = append(selectedFields, "c.curr_short as currency_name")
			selectedFields = append(selectedFields, "c.curr_delim")
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

// applySortingWithDistinct adds ORDER BY clauses for DISTINCT ON queries
func (r *transactionRepository) applySortingWithDistinct(query *gorm.DB, sort []models.SortParams) *gorm.DB {
	// For DISTINCT ON (p.payment_tx_log_id), we must order by p.payment_tx_log_id first
	orderBy := []string{"p.payment_tx_log_id"}

	if len(sort) == 0 {
		// Default sort - add updated_at after the required DISTINCT column
		orderBy = append(orderBy, "p.updated_at DESC")
	} else {
		// Add user-specified sorts after the required DISTINCT column
		for _, s := range sort {
			direction := "ASC"
			if strings.ToLower(s.Direction) == "desc" {
				direction = "DESC"
			}

			var sqlField string
			if mappedField, exists := config.FieldMappings[s.Field]; exists {
				sqlField = mappedField
			} else {
				sqlField = fmt.Sprintf("p.%s", s.Field)
			}

			orderBy = append(orderBy, fmt.Sprintf("%s %s", sqlField, direction))
		}
	}

	return query.Order(strings.Join(orderBy, ", "))
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

		// Set type string based on payment_tx_type_id (if not already set by SQL)
		if tx.Type == "" {
			tx.Type = tx.GetTypeString()
		}

		// Set reversed flag
		tx.Reversed = tx.IsReversed()

		// Set response_code as alias for result_code (if not already set by SQL)
		if tx.ResponseCode == nil {
			tx.ResponseCode = tx.ResultCode
		}

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

		// Create currency info from joined currency data
		r.populateCurrencyInfo(tx)

		// Generate PAN from bin_id and pan_id if available
		if tx.BinID != nil && tx.PanID != nil && *tx.BinID != "" && *tx.PanID != "" {
			binID := *tx.BinID
			panID := *tx.PanID
			if len(binID) >= 6 {
				pan := fmt.Sprintf("%s %s** **** %s", binID[:4], binID[4:6], panID)
				tx.PAN = &pan
			}
		}
	}
}

// populateCurrencyInfo populates currency information for a transaction
func (r *transactionRepository) populateCurrencyInfo(tx *models.Transaction) {
	// Use joined currency data if available, otherwise query database
	if tx.CurrencyInfo == nil && tx.CurrencyCode != "" {
		currInfo := &models.CurrencyInfo{
			Code:     tx.CurrencyCode,
			Name:     tx.CurrencyName, // From joined currency table
			Symbol:   "R",             // Default to R for South African Rand
			Exponent: tx.CurrDelim,    // From joined currency table
		}

		// If joined data is empty, fall back to database query
		if currInfo.Name == "" || currInfo.Exponent == 0 {
			var currency models.Currency
			if err := r.getDB().Where("curr_code = ?", tx.CurrencyCode).First(&currency).Error; err == nil {
				currInfo.Name = currency.CurrencyName
				currInfo.Exponent = currency.CurrDelim
			}
		}

		// Format the amount using the currency info
		currInfo.FormattedAmount = currInfo.FormatAmount(tx.Amount)
		tx.CurrencyInfo = currInfo
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

// GetTransactionTotals returns transaction totals by type for a specific date and device/terminal
func (r *transactionRepository) GetTransactionTotals(merchantID string, request models.TransactionTotalsRequest) (*models.TransactionTotalsResponse, error) {
	type TotalResult struct {
		PaymentTxTypeID int     `gorm:"column:payment_tx_type_id"`
		TypeName        string  `gorm:"column:type_name"`
		TrxType         string  `gorm:"column:trx_type"`
		TrxDescr        string  `gorm:"column:trx_descr"`
		TotalAmount     float64 `gorm:"column:total_amount"`
	}

	var results []TotalResult

	// Build base query with JOIN to payment_tx_types and merchant tables
	query := r.getDB().Table("payment_tx_log p").
		Select(`
			p.payment_tx_type_id,
			COALESCE(pt.name, 'Unknown') as type_name,
			CASE 
				WHEN p.payment_tx_type_id = 0 THEN 'payment'
				WHEN p.payment_tx_type_id = 1 THEN 'reversal'
				WHEN p.payment_tx_type_id = 2 THEN 'void'
				WHEN p.payment_tx_type_id = 3 THEN 'refund'
				WHEN p.payment_tx_type_id = 9 THEN 'mm_purchase'
				WHEN p.payment_tx_type_id = 10 THEN 'mm_refund'
				ELSE 'unknown'
			END as trx_type,
			COALESCE(pt.name, 'Unknown') as trx_descr,
			SUM(CAST(p.amount as DECIMAL(15,2)) / 100.0) as total_amount
		`).
		Joins("LEFT JOIN payment_tx_types pt ON p.payment_tx_type_id = pt.payment_tx_type_id").
		Joins("LEFT JOIN merchants m ON p.merchant_id = m.merchant_id").
		Where("DATE(p.created_at) = ?", request.Date).
		Where("m.merchant_id = ? OR m.provisioner_id = ?", merchantID, merchantID).
		Group("p.payment_tx_type_id, pt.name").
		Order("p.payment_tx_type_id")

	// Apply device/terminal filters
	if request.DeviceID != "" {
		query = query.Where("p.device_id = ?", request.DeviceID)
	}
	if request.TerminalID != "" {
		query = query.Where("p.terminal_id = ?", request.TerminalID)
	}
	if request.BankTerminalID != "" {
		query = query.Where("p.bank_terminal_id = ?", request.BankTerminalID)
	}

	// Execute query
	if err := query.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction totals: %w", err)
	}

	// Convert results to response format
	totals := make([]models.TransactionTotal, len(results))
	for i, result := range results {
		totals[i] = models.TransactionTotal{
			TrxType:     result.TrxType,
			TrxDescr:    result.TrxDescr,
			TotalAmount: result.TotalAmount,
		}
	}

	// Build response
	response := &models.TransactionTotalsResponse{
		Date:   request.Date,
		Totals: totals,
	}

	// Add device/terminal info to response if provided
	if request.DeviceID != "" {
		response.DeviceID = request.DeviceID
	}
	if request.TerminalID != "" {
		response.TerminalID = request.TerminalID
	}
	if request.BankTerminalID != "" {
		response.BankTerminalID = request.BankTerminalID
	}

	return response, nil
}

// GetTransactionLookup returns transaction totals by description for a specific date and device
func (r *transactionRepository) GetTransactionLookup(request models.TransactionLookupRequest) (*models.TransactionLookupResponse, error) {
	type LookupResult struct {
		TrxDescr       string  `gorm:"column:trx_descr"`
		TotalAmountEGP float64 `gorm:"column:total_amount_egp"`
	}

	var results []LookupResult
	var query *gorm.DB

	// Build query based on whether device_id is provided
	if request.DeviceID != "" {
		// Query with device_id filter
		query = r.getDB().Raw(`
			SELECT 
				TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."43"'))) AS trx_descr,
				SUM(trx_amt) / 100 AS total_amount_egp
			FROM iso_trx
			WHERE DATE(trx_datetime) = ? 
			AND TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."42"')) = ?
			AND trx_rsp_code = '00'
			GROUP BY trx_descr
		`, request.Date, request.DeviceID)
	} else {
		// Query without device_id filter (all devices)
		query = r.getDB().Raw(`
			SELECT 
				TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."43"'))) AS trx_descr,
				SUM(trx_amt) / 100 AS total_amount_egp
			FROM iso_trx
			WHERE DATE(trx_datetime) = ? 
			AND trx_rsp_code = '00'
			GROUP BY trx_descr
		`, request.Date)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction lookup: %w", err)
	}

	// Convert results to response format
	totals := make([]models.TransactionLookupTotal, len(results))
	for i, result := range results {
		totals[i] = models.TransactionLookupTotal{
			TrxDescr:       result.TrxDescr,
			TotalAmountEGP: result.TotalAmountEGP,
		}
	}

	// Build response
	response := &models.TransactionLookupResponse{
		Date:     request.Date,
		DeviceID: request.DeviceID,
		Totals:   totals,
	}

	return response, nil
}
// SearchTransactionDetails returns detailed transaction information based on search criteria
func (r *transactionRepository) SearchTransactionDetails(request models.IsoTransactionSearchRequest) (*models.IsoTransactionSearchResponse, error) {
	type SearchResult struct {
		Datetime        string  `gorm:"column:datetime"`
		STAN            int     `gorm:"column:STAN"`
		TrxRRN          string  `gorm:"column:trx_rrn"`
		BIN             string  `gorm:"column:BIN"`
		PANID           string  `gorm:"column:PANID"`
		DeviceID        string  `gorm:"column:device_id"`
		GroupID         string  `gorm:"column:group_id"`
		TrxDescr        string  `gorm:"column:trx_descr"`
		TrxType         string  `gorm:"column:trx_type"`
		BankGroupID     *string `gorm:"column:bank_group_id"`
		TransactionCode *string `gorm:"column:transaction_code"`
		TxID            *string `gorm:"column:tx_id"`
		Amount          int     `gorm:"column:amount"`
		RC              string  `gorm:"column:RC"`
		TrxAuthCode     *string `gorm:"column:trx_auth_code"`
	}
	
	var results []SearchResult
	
	// Build dynamic query based on provided filters
	baseQuery := `
		SELECT
			trx_datetime AS datetime,
			trx_stan AS STAN,
			COALESCE(trx_rrn, '') AS trx_rrn,
			LEFT(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."35"')), 6) AS BIN,
			RIGHT(SUBSTRING_INDEX(SUBSTRING_INDEX(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."35"')),'=',1),'D',1), 4) AS PANID,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."42"'))) AS device_id,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."41"'))) AS group_id,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."43"'))) AS trx_descr,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(trx_snd, '$."3"'))) AS trx_type,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(JSON_EXTRACT(trx_snd, '$."request_meta"'), '$."bank_group_id"'))) AS bank_group_id,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(JSON_EXTRACT(trx_snd, '$."request_meta"'), '$."transaction_code"'))) AS transaction_code,
			TRIM(TRIM(BOTH '"' FROM JSON_EXTRACT(JSON_EXTRACT(trx_snd, '$."request_meta"'), '$."trx_id"'))) AS tx_id,
			trx_amt AS amount,
			trx_rsp_code AS RC,
			COALESCE(trx_auth_code, '') AS trx_auth_code
		FROM iso_trx
		WHERE 1=1
	`
	
	// Build WHERE conditions dynamically
	var conditions []string
	var args []interface{}
	
	// Add date filter if provided
	if request.Date != "" {
		conditions = append(conditions, "DATE(trx_datetime) = ?")
		args = append(args, request.Date)
	}
	
	// Add device_id filter if provided
	if request.DeviceID != "" {
		conditions = append(conditions, "TRIM(TRIM(BOTH '\"' FROM JSON_EXTRACT(trx_snd, '$.\"42\"'))) = ?")
		args = append(args, request.DeviceID)
	}
	
	// Add trx_rrn filter if provided
	if request.TrxRRN != "" {
		conditions = append(conditions, "trx_rrn = ?")
		args = append(args, request.TrxRRN)
	}
	
	// Add amount filter if provided (non-zero)
	if request.Amount != 0 {
		conditions = append(conditions, "trx_amt = ?")
		args = append(args, request.Amount)
	}
	
	// Add panid filter if provided
	if request.PanID != "" {
		conditions = append(conditions, "RIGHT(SUBSTRING_INDEX(SUBSTRING_INDEX(TRIM(BOTH '\"' FROM JSON_EXTRACT(trx_snd, '$.\"35\"')),'=',1),'D',1), 4) = ?")
		args = append(args, request.PanID)
	}
	
	// Add group_id filter if provided
	if request.GroupID != "" {
		conditions = append(conditions, "TRIM(TRIM(BOTH '\"' FROM JSON_EXTRACT(trx_snd, '$.\"41\"'))) = ?")
		args = append(args, request.GroupID)
	}
	
	// Add bank_group_id filter if provided
	if request.BankGroupID != "" {
		conditions = append(conditions, "TRIM(TRIM(BOTH '\"' FROM JSON_EXTRACT(JSON_EXTRACT(trx_snd, '$.\"request_meta\"'), '$.\"bank_group_id\"'))) = ?")
		args = append(args, request.BankGroupID)
	}
	
	// Add trx_descr filter if provided
	if request.TrxDescr != "" {
		conditions = append(conditions, "TRIM(TRIM(BOTH '\"' FROM JSON_EXTRACT(trx_snd, '$.\"43\"'))) = ?")
		args = append(args, request.TrxDescr)
	}
	
	// Add tx_id filter if provided
	if request.TxID != "" {
		conditions = append(conditions, "TRIM(TRIM(BOTH '\"' FROM JSON_EXTRACT(JSON_EXTRACT(trx_snd, '$.\"request_meta\"'), '$.\"trx_id\"'))) = ?")
		args = append(args, request.TxID)
	}
	
	// Add response_code filter if provided
	if request.ResponseCode != "" {
		conditions = append(conditions, "trx_rsp_code = ?")
		args = append(args, request.ResponseCode)
	}
	
	// Combine all conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}
	
	baseQuery += " ORDER BY trx_datetime"
	
	// Execute the dynamic query
	query := r.getDB().Raw(baseQuery, args...)
	
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search transaction details: %w", err)
	}
	
	// Convert results to response format
	transactions := make([]models.TransactionSearchItem, len(results))
	for i, result := range results {
		transactions[i] = models.TransactionSearchItem{
			Datetime:        result.Datetime,
			STAN:            result.STAN,
			RRN:             result.TrxRRN,
			BIN:             result.BIN,
			PANID:           result.PANID,
			DeviceID:        result.DeviceID,
			GroupID:         result.GroupID,
			TrxDescr:        result.TrxDescr,
			TrxType:         result.TrxType,
			BankGroupID:     result.BankGroupID,
			TransactionCode: result.TransactionCode,
			TxID:            result.TxID,
			Amount:          result.Amount,
			RC:              result.RC,
			TrxAuthCode:     result.TrxAuthCode,
		}
	}
	
	// Build response
	response := &models.IsoTransactionSearchResponse{
		Transactions: transactions,
	}
	
	return response, nil
}