package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB        // PostgreSQL database for v2 APIs
var MySQLDB *gorm.DB   // MySQL database for efinance v1 APIs

// ConnectDB initializes both database connections
func ConnectDB() {
	connectPostgreSQL()
	connectMySQL()
}

// connectPostgreSQL initializes the PostgreSQL connection for v2 APIs
func connectPostgreSQL() {
	var err error
	
	// PostgreSQL connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefault("POSTGRES_HOST", "10.100.18.32"),
		getEnvOrDefault("POSTGRES_PORT", "5432"),
		getEnvOrDefault("POSTGRES_USER", "wizzit_pay"),
		getEnvOrDefault("POSTGRES_PASSWORD", "wizzit_pay"),
		getEnvOrDefault("POSTGRES_DB", "wizzit_pay"),
	)
	
	log.Printf("Connecting to PostgreSQL database: host=%s port=%s dbname=%s user=%s",
		getEnvOrDefault("POSTGRES_HOST", "10.100.18.32"),
		getEnvOrDefault("POSTGRES_PORT", "5432"),
		getEnvOrDefault("POSTGRES_DB", "wizzit_pay"),
		getEnvOrDefault("POSTGRES_USER", "wizzit_pay"),
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("⚠️ Failed to connect to PostgreSQL database: %v", err)
		log.Printf("Continuing without PostgreSQL connection...")
	} else {
		log.Println("✅ PostgreSQL database connection established successfully")
	}
}

// connectMySQL initializes the MySQL connection for efinance v1 APIs  
func connectMySQL() {
	var err error
	
	// MySQL connection string for efinance APIs
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnvOrDefault("MYSQL_USER", "atlas"),
		getEnvOrDefault("MYSQL_PASSWORD", "wizzitSTG11!#"),
		getEnvOrDefault("MYSQL_HOST", "10.100.18.31"),
		getEnvOrDefault("MYSQL_PORT", "3306"),
		getEnvOrDefault("MYSQL_DATABASE", "atlas"),
	)
	
	log.Printf("Connecting to MySQL database: host=%s port=%s dbname=%s user=%s",
		getEnvOrDefault("MYSQL_HOST", "10.100.18.31"),
		getEnvOrDefault("MYSQL_PORT", "3306"),
		getEnvOrDefault("MYSQL_DATABASE", "atlas"),
		getEnvOrDefault("MYSQL_USER", "atlas"),
	)

	MySQLDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("⚠️ Failed to connect to MySQL database: %v", err)
		log.Printf("Continuing without MySQL connection...")
	} else {
		log.Println("✅ MySQL database connection established successfully")
	}
}

// testConnection verifies the database connection works
func testConnection(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return fmt.Errorf("failed to execute test query: %v", err)
	}

	if result != 1 {
		return fmt.Errorf("test query returned unexpected result: %d", result)
	}

	log.Printf("✅ Successfully verified database connection")
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultPort returns the default port for the given database type
func getDefaultPort(dbType string) string {
	switch strings.ToLower(dbType) {
	case "postgres", "postgresql":
		return "5432"
	case "mysql":
		return "3306"
	default:
		return "3306"
	}
}

