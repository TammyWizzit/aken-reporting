package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aken_reporting_service/internal/config"
	"aken_reporting_service/internal/models"
	"aken_reporting_service/internal/services"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionService services.TransactionService
}

func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// GetTransactions handles GET /api/v2/transactions
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	merchantID := getMerchantID(c)
	if merchantID == "" {
		h.sendErrorResponse(c, http.StatusUnauthorized, config.ErrorCodeAuthFailed, "Invalid or missing authentication credentials", nil)
		return
	}

	// Parse query parameters
	fieldsParam := c.Query("fields")
	filterParam := c.Query("filter")
	sortParam := c.Query("sort")
	pageParam := c.DefaultQuery("page", "1")
	limitParam := c.DefaultQuery("limit", "100")
	timezone := c.DefaultQuery("timezone", "UTC")
	panFormat := c.DefaultQuery("pan_format", "bin_id_and_pan_id")

	// Parse pagination
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeBadRequest, "Invalid page parameter", nil)
		return
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 || limit > config.MaxPageSize {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeBadRequest, fmt.Sprintf("Invalid limit parameter (must be 1-%d)", config.MaxPageSize), nil)
		return
	}

	// Parse fields
	var fields []string
	if fieldsParam != "" {
		fields = parseCommaSeparated(fieldsParam)
	}

	// Parse filter
	filter, err := h.transactionService.ParseAdvancedFilter(filterParam, timezone)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeInvalidFilter, fmt.Sprintf("Invalid filter expression: %v", err), nil)
		return
	}

	// Parse sort
	sort, err := h.transactionService.ParseSort(sortParam)
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeInvalidSort, fmt.Sprintf("Invalid sort expression: %v", err), nil)
		return
	}

	// Prepare service parameters
	params := &services.GetTransactionsParams{
		Filter:    filter,
		Fields:    fields,
		Sort:      sort,
		Page:      page,
		Limit:     limit,
		Timezone:  timezone,
		PANFormat: panFormat,
	}

	// Get transactions
	result, err := h.transactionService.GetTransactions(merchantID, params)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("Database error in GetTransactions: %v\n", err)
		
		// Check if this is an internal error that should be sanitized
		if config.IsInternalError(err) {
			h.sendErrorResponse(c, http.StatusServiceUnavailable, config.ErrorCodeServiceUnavailable, "", 
				gin.H{"retry_after": 30})
		} else {
			h.sendErrorResponse(c, http.StatusInternalServerError, config.ErrorCodeDatabaseError, "", nil)
		}
		return
	}

	// Build response
	response := gin.H{
		"data": result.Transactions,
		"meta": gin.H{
			"pagination": gin.H{
				"page":              result.Page,
				"limit":             result.Limit,
				"total":             result.TotalCount,
				"total_pages":       result.TotalPages,
				"current_page_count": result.CurrentPageCount,
				"has_next":          result.HasNext,
				"has_prev":          result.HasPrev,
			},
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"version":          config.APIVersion,
			"execution_time_ms": 150, // In real implementation, measure actual time
			"cached":           false,
		},
		"links": h.buildPaginationLinks(c, result.Page, result.TotalPages, result.Limit),
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionByID handles GET /api/v2/transactions/:id
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	merchantID := getMerchantID(c)
	if merchantID == "" {
		h.sendErrorResponse(c, http.StatusUnauthorized, config.ErrorCodeAuthFailed, "Invalid or missing authentication credentials", nil)
		return
	}

	transactionID := c.Param("id")
	if transactionID == "" {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeBadRequest, "Transaction ID is required", nil)
		return
	}

	fieldsParam := c.Query("fields")
	timezone := c.DefaultQuery("timezone", "UTC")
	panFormat := c.DefaultQuery("pan_format", "bin_id_and_pan_id")

	var fields []string
	if fieldsParam != "" {
		fields = parseCommaSeparated(fieldsParam)
	}

	transaction, err := h.transactionService.GetTransactionByID(merchantID, transactionID, fields, timezone, panFormat)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, config.ErrorCodeDatabaseError, fmt.Sprintf("Failed to retrieve transaction: %v", err), nil)
		return
	}

	if transaction == nil {
		h.sendErrorResponse(c, http.StatusNotFound, config.ErrorCodeTxNotFound, fmt.Sprintf("Transaction with ID %s not found", transactionID), nil)
		return
	}

	response := gin.H{
		"data": transaction,
		"meta": gin.H{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   config.APIVersion,
		},
	}

	c.JSON(http.StatusOK, response)
}

// AdvancedTransactionSearch handles POST /api/v2/transactions/search
func (h *TransactionHandler) AdvancedTransactionSearch(c *gin.Context) {
	merchantID := getMerchantID(c)
	if merchantID == "" {
		h.sendErrorResponse(c, http.StatusUnauthorized, config.ErrorCodeAuthFailed, "Invalid or missing authentication credentials", nil)
		return
	}

	var searchReq models.TransactionSearchRequest
	if err := c.ShouldBindJSON(&searchReq); err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeBadRequest, fmt.Sprintf("Invalid request body: %v", err), nil)
		return
	}

	timezone := c.DefaultQuery("timezone", "UTC")
	panFormat := c.DefaultQuery("pan_format", "bin_id_and_pan_id")

	result, err := h.transactionService.SearchTransactions(merchantID, &searchReq, timezone, panFormat)
	if err != nil {
		h.sendErrorResponse(c, http.StatusInternalServerError, config.ErrorCodeDatabaseError, fmt.Sprintf("Failed to search transactions: %v", err), nil)
		return
	}

	// Handle aggregations if requested
	aggregationResults := make(map[string]interface{})
	if searchReq.Aggregations != nil {
		// Simple aggregation implementation
		if _, exists := searchReq.Aggregations["total_amount"]; exists {
			total := int64(0)
			for _, tx := range result.Transactions {
				total += tx.Amount
			}
			aggregationResults["total_amount"] = map[string]interface{}{"value": total}
		}
		
		if _, exists := searchReq.Aggregations["avg_amount"]; exists {
			if len(result.Transactions) > 0 {
				total := int64(0)
				for _, tx := range result.Transactions {
					total += tx.Amount
				}
				avg := float64(total) / float64(len(result.Transactions))
				aggregationResults["avg_amount"] = map[string]interface{}{"value": avg}
			} else {
				aggregationResults["avg_amount"] = map[string]interface{}{"value": 0}
			}
		}
	}

	response := gin.H{
		"data": result.Transactions,
		"meta": gin.H{
			"pagination": gin.H{
				"page":        result.Page,
				"limit":       result.Limit,
				"total":       result.TotalCount,
				"total_pages": result.TotalPages,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   config.APIVersion,
		},
		"aggregations": aggregationResults,
	}

	c.JSON(http.StatusOK, response)
}

// GetMerchantSummary handles GET /api/v2/merchants/:merchant_id/summary
func (h *TransactionHandler) GetMerchantSummary(c *gin.Context) {
	merchantID := getMerchantID(c)
	if merchantID == "" {
		h.sendErrorResponse(c, http.StatusUnauthorized, config.ErrorCodeAuthFailed, "Invalid or missing authentication credentials", nil)
		return
	}

	requestedMerchantID := c.Param("merchant_id")
	if requestedMerchantID == "" {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeBadRequest, "Merchant ID is required", nil)
		return
	}

	// Verify merchant access (can only access own data)
	if requestedMerchantID != merchantID {
		h.sendErrorResponse(c, http.StatusForbidden, config.ErrorCodeAuthzFailed, "Access denied to this merchant data", nil)
		return
	}

	// Parse optional filter parameters
	filterParam := c.Query("filter")
	filter, err := h.transactionService.ParseAdvancedFilter(filterParam, "UTC")
	if err != nil {
		h.sendErrorResponse(c, http.StatusBadRequest, config.ErrorCodeInvalidFilter, fmt.Sprintf("Invalid filter expression: %v", err), nil)
		return
	}

	summary, err := h.transactionService.GetMerchantSummary(merchantID, filter)
	if err != nil {
		// Log the actual error for debugging
		fmt.Printf("Database error in GetMerchantSummary: %v\n", err)
		
		// Check if this is an internal error that should be sanitized
		if config.IsInternalError(err) {
			h.sendErrorResponse(c, http.StatusServiceUnavailable, config.ErrorCodeServiceUnavailable, "", nil)
		} else {
			h.sendErrorResponse(c, http.StatusInternalServerError, config.ErrorCodeDatabaseError, "", nil)
		}
		return
	}

	response := gin.H{
		"data": gin.H{
			"merchant_id":   summary.MerchantID,
			"merchant_name": summary.MerchantName,
			"summary": gin.H{
				"total_transactions":      summary.TotalTransactions,
				"successful_transactions": summary.SuccessfulTransactions,
				"failed_transactions":     summary.FailedTransactions,
				"total_amount":            summary.TotalAmount,
				"average_amount":          summary.AverageAmount,
				"success_rate":            summary.SuccessRate,
				"date_range": gin.H{
					"from": summary.DateFrom.Format(time.RFC3339),
					"to":   summary.DateTo.Format(time.RFC3339),
				},
			},
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   config.APIVersion,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetMerchantTransactions handles GET /api/v2/merchants/:merchant_id/transactions
func (h *TransactionHandler) GetMerchantTransactions(c *gin.Context) {
	requestedMerchantID := c.Param("merchant_id")
	
	// Add merchant filter to query parameters
	existingFilter := c.Query("filter")
	if existingFilter == "" {
		c.Request.URL.RawQuery += fmt.Sprintf("filter=merchant_id:eq:%s", requestedMerchantID)
	} else {
		c.Request.URL.RawQuery += fmt.Sprintf("&filter=%s AND merchant_id:eq:%s", existingFilter, requestedMerchantID)
	}
	
	// Delegate to main GetTransactions handler
	h.GetTransactions(c)
}

// Helper functions

func (h *TransactionHandler) sendErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	// Use user-friendly message if available
	userMessage := config.GetUserFriendlyMessage(errorCode)
	if message != "" {
		userMessage = message
	}
	
	response := gin.H{
		"code":       errorCode,
		"message":    userMessage,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"request_id": getRequestID(c),
	}
	
	if details != nil {
		response["details"] = details
	}
	
	c.JSON(statusCode, response)
}

func (h *TransactionHandler) buildPaginationLinks(c *gin.Context, currentPage, totalPages, limit int) gin.H {
	baseURL := fmt.Sprintf("%s://%s%s", getScheme(c), c.Request.Host, c.Request.URL.Path)
	query := c.Request.URL.Query()
	
	// Remove page parameter for link building
	delete(query, "page")
	baseQuery := query.Encode()
	
	links := gin.H{
		"self":  buildURL(baseURL, baseQuery, currentPage),
		"first": buildURL(baseURL, baseQuery, 1),
		"last":  buildURL(baseURL, baseQuery, totalPages),
	}
	
	if currentPage > 1 {
		links["prev"] = buildURL(baseURL, baseQuery, currentPage-1)
	} else {
		links["prev"] = nil
	}
	
	if currentPage < totalPages {
		links["next"] = buildURL(baseURL, baseQuery, currentPage+1)
	} else {
		links["next"] = nil
	}
	
	return links
}

func buildURL(baseURL, query string, page int) string {
	if query == "" {
		return fmt.Sprintf("%s?page=%d", baseURL, page)
	}
	return fmt.Sprintf("%s?%s&page=%d", baseURL, query, page)
}

func getMerchantID(c *gin.Context) string {
	if merchantID, exists := c.Get("merchantID"); exists {
		if id, ok := merchantID.(string); ok {
			return id
		}
	}
	if merchantID, exists := c.Get("merchant_id"); exists {
		if id, ok := merchantID.(string); ok {
			return id
		}
	}
	return ""
}

func getRequestID(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
	}
	return requestID
}

func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if scheme := c.GetHeader("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

func parseCommaSeparated(input string) []string {
	result := make([]string, 0)
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}