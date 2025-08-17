package models

import (
	"encoding/json"
	"time"
	"fmt"
	"math"
)

// Transaction represents a payment transaction from the AKEN system
type Transaction struct {
	ID                 string          `json:"tx_log_id" gorm:"column:payment_tx_log_id;primaryKey"`
	PaymentTxTypeID    int             `json:"payment_tx_type_id" gorm:"column:payment_tx_type_id"`
	PaymentProviderID  int             `json:"payment_provider_id" gorm:"column:payment_provider_id"`
	ReversedTxLogID    *string         `json:"reversed_tx_log_id" gorm:"column:reversed_tx_log_id"`
	RRN                string          `json:"rrn" gorm:"column:rrn"`
	STAN               string          `json:"stan" gorm:"column:stan"`
	BinID              *string         `json:"bin_id" gorm:"column:bin_id"`
	PanID              *string         `json:"pan_id" gorm:"column:pan_id"`
	DeviceID           *string         `json:"device_id" gorm:"column:device_id"`
	MerchantCode       string          `json:"merchant_code" gorm:"column:merchant_code"`
	TerminalID         *string         `json:"terminal_id" gorm:"column:terminal_id"`
	CurrencyCode       string          `json:"currency_code" gorm:"column:currency_code"`
	Amount             int64           `json:"amount" gorm:"column:amount"`
	AuthCode           *string         `json:"auth_code" gorm:"column:auth_code"`
	ResultCode         *string         `json:"result_code" gorm:"column:result_code"`
	Description        *string         `json:"description" gorm:"column:description"`
	Completed          bool            `json:"completed" gorm:"column:completed"`
	Active             bool            `json:"active" gorm:"column:active"`
	CreatedAt          time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt          time.Time       `json:"updated_at" gorm:"column:updated_at"`
	MerchantID         *string         `json:"merchant_id" gorm:"column:merchant_id"`
	Reversed           bool            `json:"reversed" gorm:"column:reversed"`
	ProfileID          *string         `json:"profile_id" gorm:"column:profile_id"`
	PaymentTxRef       *string         `json:"payment_tx_ref" gorm:"column:payment_tx_ref"`
	Meta               json.RawMessage `json:"meta" gorm:"column:meta"`
	AdditionalAmount   *string         `json:"additional_amount" gorm:"column:additional_amount"`
	
	// Computed fields for API responses
	Type               string          `json:"tx_log_type" gorm:"-"` // Computed from PaymentTxTypeID
	MerchantName       string          `json:"merchant_name" gorm:"-"` // Joined from merchants table
	PAN                *string         `json:"pan" gorm:"-"` // Computed from BinID and PanID
	ResponseCode       *string         `json:"response_code" gorm:"-"` // Alias for ResultCode
	UserRef            *string         `json:"user_ref" gorm:"-"` // From meta field
	SettlementDate     *time.Time      `json:"settlement_date" gorm:"-"` // Not in current schema
	SettlementStatus   *string         `json:"settlement_status" gorm:"-"` // Not in current schema
	CardType           *string         `json:"card_type" gorm:"-"` // Not in current schema
	CurrencyInfo       *CurrencyInfo   `json:"currency_info" gorm:"-"` // Currency formatting information
}

// TableName returns the table name for GORM
func (Transaction) TableName() string {
	return "payment_tx_log"
}

// GetTypeString converts payment_tx_type_id to readable string
func (t *Transaction) GetTypeString() string {
	switch t.PaymentTxTypeID {
	case 0:
		return "payment"
	case 1:
		return "reversal"
	case 2:
		return "void"
	case 3:
		return "refund"
	case 9:
		return "mm purchase"
	case 10:
		return "mm refund"
	default:
		return "unknown"
	}
}

// IsReversed checks if transaction was reversed
func (t *Transaction) IsReversed() bool {
	return t.ReversedTxLogID != nil
}

// PaymentProvider represents payment provider information
type PaymentProvider struct {
	ID        int       `json:"payment_provider_id" gorm:"column:payment_provider_id;primaryKey"`
	Name      *string   `json:"name" gorm:"column:name"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for GORM
func (PaymentProvider) TableName() string {
	return "payment_providers"
}

// PaymentTxType represents payment transaction type information
type PaymentTxType struct {
	ID        int       `json:"payment_tx_type_id" gorm:"column:payment_tx_type_id;primaryKey"`
	Name      string    `json:"name" gorm:"column:name"`
	Active    bool      `json:"active" gorm:"column:active"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for GORM
func (PaymentTxType) TableName() string {
	return "payment_tx_types"
}

// Merchant represents merchant information
type Merchant struct {
	ID                       string    `json:"merchant_id" gorm:"column:merchant_id;primaryKey"`
	Name                     string    `json:"merchant_name" gorm:"column:name"`
	TerminalID               *string   `json:"terminal_id" gorm:"column:terminal_id"`
	MerchantCode             string    `json:"merchant_code" gorm:"column:merchant_code"`
	Address                  *string   `json:"address" gorm:"column:address"`
	Active                   bool      `json:"active" gorm:"column:active"`
	CreatedAt                time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt                time.Time `json:"updated_at" gorm:"column:updated_at"`
	CallbackURL              *string   `json:"callback_url" gorm:"column:callback_url"`
	Location                 *string   `json:"merchant_location" gorm:"column:merchant_location"`
	ProvisionerID            *string   `json:"provisioner_id" gorm:"column:provisioner_id"`
	IsProvisioner            bool      `json:"is_provisioner" gorm:"column:is_provisioner"`
	CurrencyCode             string    `json:"currency_code" gorm:"column:currency_code"`
	PaymentProviderID        int       `json:"payment_provider_id" gorm:"column:payment_provider_id"`
	PaymentMeta              json.RawMessage `json:"payment_meta" gorm:"column:payment_meta"`
	CallbackHeaders          json.RawMessage `json:"callback_headers" gorm:"column:callback_headers"`
	AllowDuplicateMID        bool      `json:"allow_duplicate_mid" gorm:"column:allow_duplicate_mid"`
	RequiresSupervisorPin    bool      `json:"requires_supervisor_pin" gorm:"column:requires_supervisor_pin"`
	SupervisorPin            string    `json:"supervisor_pin" gorm:"column:supervisor_pin"`
	Tip                      bool      `json:"tip" gorm:"column:tip"`
	CurrencyMerchantID       *string   `json:"currency_merchant_id" gorm:"column:currency_merchant_id"`
	CallbackProviderID       int       `json:"callback_provider_id" gorm:"column:callback_provider_id"`
	MobileApplicationID      *string   `json:"mobile_application_id" gorm:"column:mobile_application_id"`
	CountryCode              string    `json:"country_code" gorm:"column:country_code"`
	CityName                 string    `json:"city_name" gorm:"column:city_name"`
	PostalCode               string    `json:"postal_code" gorm:"column:postal_code"`
	MerchantCategoryCode     string    `json:"merchant_category_code" gorm:"column:merchant_category_code"`
	TelephoneNumber          string    `json:"telephone_number" gorm:"column:telephone_number"`
	CountrySubdivisionCode   string    `json:"country_subdivision_code" gorm:"column:country_subdivision_code"`
	Aggregator               bool      `json:"aggregator" gorm:"column:aggregator"`
	RecordID                 string    `json:"record_id" gorm:"column:record_id"`
}

// TableName returns the table name for GORM
func (Merchant) TableName() string {
	return "merchants"
}

// Device represents device information
type Device struct {
	DeviceID   string  `json:"device_id" gorm:"column:deviceId;primaryKey"`
	MSISDN     *string `json:"msisdn" gorm:"column:msisdn"`
	TerminalID *string `json:"terminal_id" gorm:"column:terminal_id"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for GORM  
func (Device) TableName() string {
	return "devices"
}

// Terminal represents terminal information
type Terminal struct {
	TerminalID     string    `json:"terminal_id" gorm:"column:terminal_id;primaryKey"`
	BankTerminalID *string   `json:"bank_terminal_id" gorm:"column:bank_terminal_id"`
	MerchantID     string    `json:"merchant_id" gorm:"column:merchant_id"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for GORM
func (Terminal) TableName() string {
	return "terminals"
}

// TransactionFilter represents filter parameters for transaction queries
type TransactionFilter struct {
	MerchantID         *string    `json:"merchant_id"`
	MerchantCode       *string    `json:"merchant_code"`
	DeviceID           *string    `json:"device_id"`
	ProfileID          *string    `json:"profile_id"`
	ResponseCode       *string    `json:"response_code"`
	ResultCode         *string    `json:"result_code"`
	DateTimeFrom       *time.Time `json:"datetime_from"`
	DateTimeTo         *time.Time `json:"datetime_to"`
	CurrencyCode       *string    `json:"currency_code"`
	PaymentProviderID  *int       `json:"payment_provider_id"`
	PaymentTxTypeID    *int       `json:"payment_tx_type_id"`
	TxLogType          *string    `json:"tx_log_type"` // For backward compatibility
	AmountMin          *int64     `json:"amount_min"`
	AmountMax          *int64     `json:"amount_max"`
	Completed          *bool      `json:"completed"`
	Active             *bool      `json:"active"`
	Reversed           *bool      `json:"reversed"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	Limit    int `json:"limit"`
	PageSize int `json:"page_size"` // For v1 compatibility
}

// SortParams represents sorting parameters  
type SortParams struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// TransactionSearchRequest represents advanced search request body
type TransactionSearchRequest struct {
	Query        interface{}       `json:"query"`
	Fields       []string          `json:"fields"`
	Sort         []SortParams      `json:"sort"`
	Pagination   PaginationParams  `json:"pagination"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

// CurrencyInfo represents currency formatting information
type CurrencyInfo struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Symbol         string  `json:"symbol"`
	Exponent       int     `json:"exponent"`
	FormattedAmount string `json:"formatted_amount"`
}

// Currency represents currency table information
type Currency struct {
	CurrencyCode string `json:"currency_code" gorm:"column:curr_code;primaryKey"`
	CurrencyName string `json:"currency_name" gorm:"column:curr_short"`
	CurrDelim    int    `json:"curr_delim" gorm:"column:curr_delim"`
}

// TableName returns the table name for GORM
func (Currency) TableName() string {
	return "currency"
}

// FormatAmount formats the amount using currency exponent
func (c *CurrencyInfo) FormatAmount(amount int64) string {
	divisor := int64(math.Pow(10, float64(c.Exponent)))
	if divisor == 0 {
		divisor = 1
	}
	
	major := amount / divisor
	minor := amount % divisor
	
	if c.Exponent == 0 {
		return fmt.Sprintf("%s %d", c.Symbol, major)
	}
	
	formatStr := fmt.Sprintf("%%s %%d.%%0%dd", c.Exponent)
	return fmt.Sprintf(formatStr, c.Symbol, major, minor)
}

// MerchantSummary represents merchant transaction summary
type MerchantSummary struct {
	MerchantID             string    `json:"merchant_id"`
	MerchantName           string    `json:"merchant_name"`
	TotalTransactions      int       `json:"total_transactions"`
	SuccessfulTransactions int       `json:"successful_transactions"`
	FailedTransactions     int       `json:"failed_transactions"`
	TotalAmount            int64     `json:"total_amount"`
	AverageAmount          float64   `json:"average_amount"`
	SuccessRate            float64   `json:"success_rate"`
	DateFrom               time.Time `json:"date_from"`
	DateTo                 time.Time `json:"date_to"`
}