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
	host := os.Getenv("AKEN_REPORTING_DB_HOST")
	port := os.Getenv("AKEN_REPORTING_DB_PORT")
	user := os.Getenv("AKEN_REPORTING_DB_USER")
	password := os.Getenv("AKEN_REPORTING_DB_PASSWORD")
	dbname := os.Getenv("AKEN_REPORTING_DB_NAME")

	if host == "" {
		log.Fatal("AKEN_REPORTING_DB_HOST environment variable is required")
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		log.Fatal("AKEN_REPORTING_DB_USER environment variable is required")
	}
	if password == "" {
		log.Fatal("AKEN_REPORTING_DB_PASSWORD environment variable is required")
		}
	if dbname == "" {
		log.Fatal("AKEN_REPORTING_DB_NAME environment variable is required")}

	return fmt.Sprintf("host=%s port=%s user=%s  dbname=%s sslmode=disable",
		host, port, user, dbname)
}

	// 		// host = "10.100.18.31"
	// if host == "" {
	// 	log.Fatal("AKEN_REPORTING_DB_HOST environment variable is required")
	// }
	// if port == "" {
	// 	port = "9088"
	// }
	// if user == "" {
	// 	user = "atlas"
	// }
	// if password == "" {
	// 	password = "wizzitSTG11!#"
	// }
	// if dbname == "" {
	// 	dbname = "atlas_db"
	// }


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

// GetJWTSecret returns the JWT signing secret
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Default secret for development (DO NOT use in production)
		secret = "aken-reporting-service-dev-secret-change-in-production"
	}
	return secret
}

// GetJWTIssuer returns the JWT issuer
func GetJWTIssuer() string {
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "aken-reporting-service"
	}
	return issuer
}
