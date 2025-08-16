package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		log.Println("Continuing with system environment variables...")
	} else {
		log.Println("Environment variables loaded from .env file")
	}
}

// GetDatabaseURL returns the database connection URL
func GetDatabaseURL() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT") 
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "wizzit_pay"
	}
	if password == "" {
		password = "wizzit_pay"
	}
	if dbname == "" {
		dbname = "wizzit_pay"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		host, port, user, password, dbname)
}

// GetPort returns the server port
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	return port
}

// IsDevMode returns true if running in development mode
func IsDevMode() bool {
	return os.Getenv("ENV") == "development" || os.Getenv("DISABLE_AUTH") == "true"
}