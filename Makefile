export APP := $(shell basename $(abspath $(dir $(lastword $(MAKEFILE_LIST)))))
ENV ?= staging
REGISTRY = registry.$(ENV).wizzitdigital.com
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")

default: build

# Go build (equivalent to browserify for Go projects)
gobuild:
	@echo "🔨 Building Go application..."
	@go mod download
	@go mod tidy
	@echo "✅ Go dependencies ready"

build:
	@echo "🏗️  Building production image..."
	make gobuild
	docker-compose build prod
	@echo "✅ Production build complete"

test:
	@echo "🧪 Running tests..."
	make gobuild
	docker-compose build dev
	docker-compose run --rm dev go test ./...
	@echo "✅ Tests completed"

dev:
	@echo "🚀 Starting development environment..."
	make gobuild
	docker-compose build dev
	@echo "📊 Starting AKEN Reporting Service..."
	@echo "🌐 API will be available at: http://localhost:8090"
	@echo "🏥 Health check: http://localhost:8090/api/v2/health"
	@echo "📈 Transactions: http://localhost:8090/api/v2/transactions"
	docker-compose up dev
	@echo "🛑 Development environment stopped"

# Local development without Docker
dev-local:
	@echo "🚀 Starting local development server..."
	@echo "📊 AKEN Reporting Service v2.0"
	@echo "🗄️  Checking database connection..."
	@docker-compose up postgres -d
	@sleep 2
	@echo "🌐 Starting API server on http://localhost:8090"
	@echo "🏥 Health endpoint: http://localhost:8090/api/v2/health"
	@echo "📈 Transactions endpoint: http://localhost:8090/api/v2/transactions"
	@echo "🛠️  Press Ctrl+C to stop"
	@DISABLE_AUTH=true ENV=development go run main.go

cleanup:
	@echo "🧹 Cleaning up containers and volumes..."
	docker-compose down --volumes
	@echo "✅ Cleanup complete"

.PHONY: build test dev dev-local

# Database operations
migrate:
	@echo "🗄️  Running database migrations..."
	@echo "📂 Loading production backup data..."
	docker-compose run --rm $(ENV) sh -c "psql -h postgres -U wizzit_pay -d wizzit_pay < 20250816_backup.sql"
	@echo "✅ Database migration complete"

# Development database setup
db-setup:
	@echo "🗄️  Setting up development database..."
	@docker-compose up postgres -d
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "📊 Loading production backup data..."
	@docker exec -i aken-postgres psql -U wizzit_pay -d wizzit_pay < 20250816_backup.sql 2>/dev/null || echo "Data already loaded"
	@echo "✅ Database setup complete"
	@echo "🔍 Database contains production-like transaction data"

# Testing targets
test-unit:
	@echo "🧪 Running unit tests..."
	@go test -v ./...
	@echo "✅ Unit tests complete"

test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-race:
	@echo "🧪 Running race condition tests..."
	@go test -race ./...
	@echo "✅ Race condition tests complete"

test-all: test-unit test-coverage test-race
	@echo "✅ All tests completed successfully"

# Code quality
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "✅ Linting complete"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt:
	@echo "💅 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Build targets
build-binary:
	@echo "🔨 Building binary..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/aken-reporting-service .
	@echo "✅ Binary built: bin/aken-reporting-service"

# Docker registry operations
push: push-$(ENV)

push-$(ENV):
	@echo "📤 Pushing $(APP) to $(REGISTRY)"
	@echo "🏷️  Version: $(VERSION)"
	docker tag $(APP):latest $(REGISTRY)/$(APP):$(VERSION)
	docker push $(REGISTRY)/$(APP):$(VERSION)
	@echo "✅ Push complete: $(REGISTRY)/$(APP):$(VERSION)"

.PHONY: push push-$(ENV) migrate db-setup test-unit test-coverage test-race test-all lint fmt build-binary

# Health and status checks
health:
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8090/api/v2/health | jq '.' || echo "❌ Service not responding"

status:
	@echo "📊 Service Status Check"
	@echo "======================="
	@echo "🐳 Docker containers:"
	@docker-compose ps 2>/dev/null || echo "Docker Compose not running"
	@echo ""
	@echo "🔌 Port 8090 status:"
	@lsof -i :8090 | head -2 || echo "Port 8090 not in use"
	@echo ""
	@echo "🏥 Health check:"
	@curl -s http://localhost:8090/api/v2/health | jq '.status' 2>/dev/null || echo "Service not responding"

# Quick start for new developers
quick-start:
	@echo "🚀 AKEN Reporting Service - Quick Start"
	@echo "======================================"
	@echo "1️⃣  Setting up database..."
	$(MAKE) db-setup
	@echo ""
	@echo "2️⃣  Starting development server..."
	@echo "🌐 API will be available at: http://localhost:8090"
	@echo "🏥 Health: http://localhost:8090/api/v2/health"
	@echo "📈 Transactions: http://localhost:8090/api/v2/transactions"
	@echo ""
	$(MAKE) dev-local

# Development workflow
dev-workflow: fmt lint test-unit dev-local

.PHONY: health status quick-start dev-workflow

# Help target
help:
	@echo "🔧 AKEN Reporting Service v2.0 - Available Commands"
	@echo "=================================================="
	@echo ""
	@echo "🚀 Development:"
	@echo "  make dev          - Start with Docker"
	@echo "  make dev-local    - Start locally (recommended)"
	@echo "  make quick-start  - Complete setup for new developers"
	@echo ""
	@echo "🏗️  Building:"
	@echo "  make build        - Build production Docker image"
	@echo "  make build-binary - Build Go binary"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test         - Run tests in Docker"
	@echo "  make test-all     - Run all test types"
	@echo "  make test-unit    - Unit tests only"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  make db-setup     - Setup development database"
	@echo "  make migrate      - Run migrations"
	@echo ""
	@echo "🔍 Quality:"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo ""
	@echo "📊 Status:"
	@echo "  make health       - Check service health"
	@echo "  make status       - Show service status"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  make cleanup      - Remove containers and volumes"