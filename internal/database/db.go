package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// ConnectDB initializes the database connection to the existing AKEN database
func ConnectDB() {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefault("PMT_TX_DB_HOST", "portal_db"),
		getEnvOrDefault("PMT_TX_DB_PORT", "5432"),
		getEnvOrDefault("PMT_TX_DB_USER", "wizzit_pay"),
		getEnvOrDefault("PMT_TX_DB_PASSWORD", "wizzit_pay"),
		getEnvOrDefault("PMT_TX_DB_DATABASE", "wizzit_pay"),
	)

	log.Printf("Connecting to database: host=%s port=%s dbname=%s user=%s", 
		getEnvOrDefault("PMT_TX_DB_HOST", "portal_db"),
		getEnvOrDefault("PMT_TX_DB_PORT", "5432"),
		getEnvOrDefault("PMT_TX_DB_DATABASE", "wizzit_pay"),
		getEnvOrDefault("PMT_TX_DB_USER", "wizzit_pay"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Disable automatic pluralization since we're using existing tables
		NamingStrategy: &NamingStrategy{},
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connection established successfully")
	
	// Test the connection with a simple query
	if err := testConnection(); err != nil {
		log.Fatalf("Database connection test failed: %v", err)
	}
	
	log.Println("✅ Database connection test passed")
}

// testConnection verifies the database connection works
func testConnection() error {
	var result int
	err := DB.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return fmt.Errorf("failed to execute test query: %v", err)
	}
	
	// Test that we can access the main tables we need
	var count int64
	if err := DB.Raw("SELECT COUNT(*) FROM merchants LIMIT 1").Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to access merchants table: %v", err)
	}
	
	if err := DB.Raw("SELECT COUNT(*) FROM payment_tx_log LIMIT 1").Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to access payment_tx_log table: %v", err)
	}
	
	log.Printf("✅ Successfully verified access to core tables")
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Custom naming strategy to work with existing AKEN table structure
type NamingStrategy struct{}

func (ns NamingStrategy) TableName(str string) string {
	// Return table names as-is since we're using existing AKEN tables
	return str
}

func (ns NamingStrategy) SchemaName(table string) string {
	return ""
}

func (ns NamingStrategy) ColumnName(table, column string) string {
	return column
}

func (ns NamingStrategy) JoinTableName(joinTable string) string {
	return joinTable
}

func (ns NamingStrategy) RelationshipFKName(relationship schema.Relationship) string {
	return relationship.Name
}

func (ns NamingStrategy) CheckerName(table, column string) string {
	return ""
}

func (ns NamingStrategy) IndexName(table, column string) string {
	return ""
}