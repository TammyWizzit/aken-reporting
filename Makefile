# AKEN Reporting Service Makefile
# Following the patterns from humble-household-overhaul-project

.PHONY: help build run test clean docker-build docker-run dev lint deps mod-tidy

# Default target
help: ## Show this help message
	@echo "AKEN Reporting Service v2.0"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
dev: ## Run in development mode with hot reload
	@echo "🚀 Starting AKEN Reporting Service in development mode..."
	@export DISABLE_AUTH=true ENV=development && go run main.go

run: ## Run the application
	@echo "▶️  Starting AKEN Reporting Service..."
	@go run main.go

build: ## Build the application binary
	@echo "🔨 Building AKEN Reporting Service..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/aken-reporting-service .
	@echo "✅ Build complete: bin/aken-reporting-service"

# Testing targets
test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

benchmark: ## Run benchmark tests
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Code quality targets
lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "💅 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet passed"

# Dependency management
deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@echo "✅ Dependencies downloaded"

mod-tidy: ## Tidy go modules
	@echo "🧹 Tidying go modules..."
	@go mod tidy
	@echo "✅ Modules tidied"

mod-verify: ## Verify go modules
	@echo "✅ Verifying go modules..."
	@go mod verify

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t aken-reporting-service:latest .
	@echo "✅ Docker image built: aken-reporting-service:latest"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -d \
		-p 8090:8090 \
		-e DISABLE_AUTH=true \
		-e ENV=development \
		--name aken-reporting-service \
		aken-reporting-service:latest
	@echo "✅ Container started on http://localhost:8090"

docker-stop: ## Stop Docker container
	@echo "🛑 Stopping Docker container..."
	@docker stop aken-reporting-service || true
	@docker rm aken-reporting-service || true
	@echo "✅ Container stopped"

docker-compose-up: ## Start all services with docker-compose
	@echo "🐳 Starting all services with docker-compose..."
	@docker-compose up -d
	@echo "✅ Services started. API available at http://localhost:8090"

docker-compose-down: ## Stop all services
	@echo "🛑 Stopping all services..."
	@docker-compose down
	@echo "✅ Services stopped"

docker-compose-logs: ## View service logs
	@docker-compose logs -f aken-reporting-service

# Database targets
db-status: ## Check database connection
	@echo "🗄️  Checking database connection..."
	@curl -s http://localhost:8090/api/v2/health | jq '.status' 2>/dev/null || echo "Service not running or jq not installed"

# Utility targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@docker system prune -f 2>/dev/null || true
	@echo "✅ Cleaned"

env-setup: ## Set up environment file
	@echo "⚙️  Setting up environment..."
	@cp .env.example .env
	@echo "✅ Environment file created. Edit .env with your settings"

health-check: ## Check service health
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8090/api/v2/health || echo "❌ Service not responding"

api-info: ## Get API information
	@echo "ℹ️  Getting API information..."
	@curl -s http://localhost:8090/api/v2/info | jq '.' 2>/dev/null || echo "Service not running or jq not installed"

# Performance targets
load-test: ## Run basic load test (requires hey)
	@echo "⚡ Running load test..."
	@if command -v hey >/dev/null 2>&1; then \
		hey -n 1000 -c 10 -H "Authorization: Basic $(shell echo -n 'test:test' | base64)" http://localhost:8090/api/v2/transactions; \
	else \
		echo "❌ 'hey' not installed. Install with: go install github.com/rakyll/hey@latest"; \
	fi

# Development workflow targets
dev-setup: env-setup deps ## Complete development setup
	@echo "🎯 Development setup complete!"
	@echo "Next steps:"
	@echo "  1. Edit .env with your database settings"
	@echo "  2. Run 'make dev' to start development server"
	@echo "  3. Visit http://localhost:8090/api/v2/health to verify"

dev-test: fmt vet test ## Run development tests
	@echo "✅ Development tests passed!"

ci: deps fmt vet test ## Run CI pipeline
	@echo "✅ CI pipeline completed successfully!"

# Release targets  
release-build: clean deps test build ## Build release version
	@echo "🚀 Release build complete!"

# Monitoring targets
logs: ## Show application logs (when running locally)
	@echo "📋 Application logs:"
	@tail -f /tmp/aken-reporting-service.log 2>/dev/null || echo "No log file found. Run with: make run > /tmp/aken-reporting-service.log"

metrics: ## Show basic metrics
	@echo "📊 Basic metrics:"
	@echo "Build info:" && go version
	@echo "Dependencies:" && go list -m all | wc -l
	@echo "Lines of code:" && find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1

# Documentation targets
docs-serve: ## Serve documentation (if you add godoc)
	@echo "📚 Starting documentation server..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "❌ godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Quick commands for common workflows
quick-start: docker-compose-up health-check ## Quick start with Docker
	@echo "🎉 AKEN Reporting Service is running!"
	@echo "📡 API: http://localhost:8090/api/v2/"
	@echo "🏥 Health: http://localhost:8090/api/v2/health" 
	@echo "ℹ️  Info: http://localhost:8090/api/v2/info"

all: deps fmt vet test build ## Run all build steps