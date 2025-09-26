package database

import (
	"fmt"
	"log"
	"strings"

	"aken_reporting_service/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB      // PostgreSQL database for v2 APIs
var MySQLDB *gorm.DB // MySQL database for efinance v1 APIs

// ConnectDB initializes both database connections
func ConnectDB() {
	connectPostgreSQL()
	connectMySQL()
}

// connectPostgreSQL initializes the PostgreSQL connection for v2 APIs
func connectPostgreSQL() {
	var err error

	// Get PostgreSQL configuration from config package
	host, port, user, password, dbname := config.GetPostgreSQLConfig()

	// Check if PostgreSQL configuration is available
	if host == "" || user == "" || dbname == "" {
		log.Printf("⚠️ PostgreSQL configuration incomplete: host=%s user=%s dbname=%s", host, user, dbname)
		log.Printf("Continuing without PostgreSQL connection...")
		return
	}

	// Set default port if not provided
	if port == "" {
		port = "5432"
	}

	// PostgreSQL connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("Connecting to PostgreSQL database: host=%s port=%s dbname=%s user=%s",
		host, port, dbname, user)

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

	// Get MySQL configuration from config package
	mysql_host, port, user, password, database := config.GetMySQLConfig()

	// Check if MySQL configuration is available
	if mysql_host == "" || user == "" || database == "" {
		log.Printf("⚠️ MySQL configuration incomplete: host=%s user=%s database=%s", mysql_host, user, database)
		log.Printf("Continuing without MySQL connection...")
		return
	}

	// Set default port if not provided
	if port == "" {
		port = "3306"
	}

	// MySQL connection string for efinance APIs
	mysql_dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, mysql_host, port, database)

	log.Printf("Connecting to MySQL database: host=%s port=%s dbname=%s user=%s",
		mysql_host, port, database, user)

	MySQLDB, err = gorm.Open(mysql.Open(mysql_dsn), &gorm.Config{})
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
