package models

import (
	"encoding/json"
	"time"
)

// Transaction represents a payment transaction from the AKEN system
type Transaction struct {
	ID               string          `json:"tx_log_id" gorm:"column:payment_tx_log_id;primaryKey"`
	Type             string          `json:"tx_log_type" gorm:"column:payment_tx_type_id"`
	ReversedTxLogID  *string         `json:"reversed_tx_log_id" gorm:"column:reversed_tx_log_id"`
	DateTime         time.Time       `json:"tx_date_time" gorm:"column:updated_at"`
	MerchantID       string          `json:"merchant_id" gorm:"column:merchant_id"`
	MerchantName     string          `json:"merchant_name" gorm:"column:merchant_name"`
	DeviceID         *string         `json:"device_id" gorm:"column:device_id"`
	Amount           int64           `json:"amount" gorm:"column:amount"`
	PAN              *string         `json:"pan" gorm:"column:pan"`
	ResponseCode     string          `json:"response_code" gorm:"column:result_code"`
	RRN              *string         `json:"rrn" gorm:"column:rrn"`
	AuthCode         *string         `json:"auth_code" gorm:"column:auth_code"`
	STAN             *string         `json:"stan" gorm:"column:stan"`
	Reversed         bool            `json:"reversed"`
	UserRef          *string         `json:"user_ref"`
	Meta             json.RawMessage `json:"meta" gorm:"column:meta"`
	Description      *string         `json:"description" gorm:"column:description"`
	SettlementDate   *time.Time      `json:"settlement_date" gorm:"column:settlement_date"`
	SettlementStatus *string         `json:"settlement_status" gorm:"column:settlement_status"`
	CardType         *string         `json:"card_type" gorm:"column:card_type"`
	CreatedAt        time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt        time.Time       `json:"updated_at" gorm:"column:updated_at"`
}

// TableName returns the table name for GORM
func (Transaction) TableName() string {
	return "payment_tx_log"
}

// GetTypeString converts payment_tx_type_id to readable string
func (t *Transaction) GetTypeString() string {
	switch t.Type {
	case "0":
		return "payment"
	case "1":
		return "reversal"
	case "2":
		return "void"
	case "3":
		return "refund"
	case "9":
		return "mm purchase"
	case "10":
		return "mm refund"
	default:
		return "unknown"
	}
}

// IsReversed checks if transaction was reversed
func (t *Transaction) IsReversed() bool {
	return t.ReversedTxLogID != nil
}

// Merchant represents merchant information
type Merchant struct {
	ID             string    `json:"merchant_id" gorm:"column:merchant_id;primaryKey"`
	Name           string    `json:"merchant_name" gorm:"column:name"`
	TerminalID     *string   `json:"terminal_id" gorm:"column:terminal_id"`
	MerchantCode   string    `json:"merchant_code" gorm:"column:merchant_code"`
	Address        *string   `json:"address" gorm:"column:address"`
	Active         bool      `json:"active" gorm:"column:active"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
	CallbackURL    *string   `json:"callback_url" gorm:"column:callback_url"`
	Location       *string   `json:"merchant_location" gorm:"column:merchant_location"`
	ProvisionerID  *string   `json:"provisioner_id" gorm:"column:provisioner_id"`
	IsProvisioner  bool      `json:"is_provisioner" gorm:"column:is_provisioner"`
	CurrencyCode   *string   `json:"currency_code" gorm:"column:currency_code"`
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
	MerchantID       *string    `json:"merchant_id"`
	DeviceID         *string    `json:"device_id"`
	ProfileID        *string    `json:"profile_id"`
	ResponseCode     *string    `json:"response_code"`
	DateTimeFrom     *time.Time `json:"datetime_from"`
	DateTimeTo       *time.Time `json:"datetime_to"`
	CurrencyCode     *string    `json:"currency_code"`
	PaymentProviderID *string   `json:"payment_provider_id"`
	TxLogType        *string    `json:"tx_log_type"`
	AmountMin        *int64     `json:"amount_min"`
	AmountMax        *int64     `json:"amount_max"`
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