.PHONY: build test clean run dev docker-build docker-up docker-down fmt lint vet

# Build the application
build:
	@echo "Building the application..."
	go build -o ./bin/server ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf ./bin
	rm -f coverage.out coverage.html

# Run the application locally
run: build
	@echo "Running the application..."
	./bin/server

# Run in development mode with hot reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air for hot reload..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t image-gallery:latest .

# Docker compose up
docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build -d

# Docker compose down
docker-down:
	@echo "Stopping services..."
	docker-compose down

# Install development dependencies
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing Atlas CLI..."
	@if ! command -v atlas > /dev/null; then \
		curl -sSf https://atlasgo.sh | sh -s -- --yes; \
	else \
		echo "Atlas CLI already installed"; \
	fi

# Atlas database schema management
atlas-validate:
	@echo "Validating Atlas configuration..."
	atlas schema validate --env local

atlas-inspect:
	@echo "Inspecting current database schema..."
	atlas schema inspect --env local

atlas-diff:
	@echo "Generating schema diff..."
	atlas migrate diff --env local

atlas-apply:
	@echo "Applying schema changes..."
	atlas schema apply --env local --auto-approve

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@if command -v atlas > /dev/null; then \
		atlas migrate apply --env local; \
	else \
		echo "Atlas not available, using built-in migration system"; \
		echo "Database migrations will be run automatically on startup"; \
	fi

# Database operations
db-start:
	@echo "Starting database services..."
	docker-compose up -d postgres

db-stop:
	@echo "Stopping database services..."
	docker-compose stop postgres

db-reset:
	@echo "Resetting database..."
	docker-compose down -v postgres
	docker-compose up -d postgres
	sleep 3
	make migrate

# Show help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  run             - Run the application locally"
	@echo "  dev             - Run in development mode with hot reload"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  vet             - Vet code"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-up       - Start services with Docker Compose"
	@echo "  docker-down     - Stop services"
	@echo "  install-tools   - Install development tools"
	@echo "  migrate         - Run database migrations"
	@echo "  atlas-validate  - Validate Atlas configuration"
	@echo "  atlas-inspect   - Inspect current database schema"
	@echo "  atlas-diff      - Generate schema diff"
	@echo "  atlas-apply     - Apply schema changes"
	@echo "  db-start        - Start database services only"
	@echo "  db-stop         - Stop database services"
	@echo "  db-reset        - Reset database with fresh schema"
	@echo "  help            - Show this help message"