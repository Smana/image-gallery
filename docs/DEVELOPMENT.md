# Development Guide

This guide provides comprehensive information for developing the Image Gallery application locally.

## 📋 Prerequisites

- **Go 1.25+** - Latest version required
- **Docker & Docker Compose** - For local development environment
- **Make** - For running development commands
- **Atlas CLI** - For database schema management (optional)
- **Dagger CLI** - For containerized CI/CD (optional, installed via `make install-tools`)

## 🛠️ Local Development Setup

### 1. Clone and Setup

```bash
git clone <repository-url>
cd image-gallery
```

### 2. Install Development Tools

```bash
# Install all development tools (Go tools, Dagger, Atlas)
make install-tools
```

This installs:
- Air (hot reload)
- golangci-lint (linting)
- Atlas CLI (database schema management)
- Dagger CLI (containerized CI/CD)

### 3. Start Development Environment

```bash
# Start PostgreSQL, MinIO, and Valkey services
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 4. Configure Environment

```bash
# Copy environment template (if available)
cp .env.example .env

# Edit .env with your configuration
```

### 5. Database Setup

```bash
# Run database migrations
make migrate

# Or using Atlas (if installed)
atlas migrate apply --env local
```

### 6. Build and Run

```bash
# Build the application
make build

# Run the server
./bin/server

# Or run directly with Go
go run cmd/server/main.go

# Or use hot reload for development
make dev
```

## 🧪 Testing

### Unit Tests

Run fast unit tests with mocking:

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/services/implementations/...

# Run with race detection
go test -race ./...
```

### Integration Tests

**Note**: Integration tests require Docker to be running.

```bash
# Start test dependencies
docker-compose up -d postgres minio valkey

# Run integration tests
go test -v ./internal/services/integrationtests/...
```

### Dagger-based Testing (Containerized)

```bash
# Run complete CI pipeline locally
make dagger-ci

# Individual steps
make dagger-lint          # Lint code
make dagger-test          # Run tests
make dagger-vulncheck     # Security scan
make dagger-trivy-fs      # Filesystem security scan
make dagger-build         # Build binary
```

## 📊 Code Quality

### Linting

```bash
# Run linter (requires golangci-lint)
make lint

# Or use Dagger (no local installation required)
make dagger-lint

# Auto-fix issues (local only)
make lint-fix
```

### Code Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

### Formatting and Vetting

```bash
# Format code
make fmt

# Vet code
make vet
```

## 🗄️ Database Management

### Atlas Schema Management

```bash
# Generate new migration
atlas migrate diff --env local

# Apply migrations
atlas migrate apply --env local

# Check migration status
atlas migrate status --env local

# Validate Atlas configuration
make atlas-validate

# Inspect current database schema
make atlas-inspect
```

### Manual Database Operations

```bash
# Start only database services
make db-start

# Stop database services
make db-stop

# Reset database with fresh schema
make db-reset
```

## 🐳 Docker Development

### Local Services

```bash
# Start all services including the app
docker-compose up --build

# Start only infrastructure dependencies
docker-compose up -d postgres minio valkey

# View logs
docker-compose logs -f app

# Follow logs for specific services
docker-compose logs -f postgres minio valkey
```

### Docker Image Building

```bash
# Build Docker image locally
make docker-build

# Or use Dagger for multi-arch builds
make dagger-docker
```

## 🔧 Available Make Commands

### Build & Development
```bash
make build          # Build the application
make run            # Run the application
make dev            # Run in development mode with hot reload
make clean          # Clean build artifacts
make fmt            # Format code
make lint           # Lint code
make vet            # Vet code
make test           # Run unit tests
make test-coverage  # Run tests with coverage
```

### Dagger CI/CD (Containerized)
```bash
make dagger-lint     # Run linting with Dagger
make dagger-test     # Run tests with Dagger
make dagger-vulncheck # Run vulnerability scan with Dagger
make dagger-trivy-fs # Run filesystem security scan with Trivy
make dagger-trivy    # Run container image security scan with Trivy
make dagger-build    # Build application with Dagger
make dagger-docker   # Build Docker image with Dagger
make dagger-ci       # Run complete CI pipeline with Dagger
```

### Docker
```bash
make docker-build   # Build Docker image
make docker-up      # Start services with Docker Compose
make docker-down    # Stop services
```

### Database
```bash
make migrate        # Run database migrations
make atlas-validate # Validate Atlas configuration
make atlas-inspect  # Inspect current database schema
make atlas-diff     # Generate schema diff
make atlas-apply    # Apply schema changes
make db-start       # Start database services only
make db-stop        # Stop database services
make db-reset       # Reset database with fresh schema
```

### Tools
```bash
make install-tools  # Install development tools (including Dagger)
make help          # Show all available commands
```

## 🌐 Development Environment

### Service Ports

- **Application**: http://localhost:8080
- **PostgreSQL**: localhost:5432 (user: `testuser`, password: `testpass`, db: `image_gallery_test`)
- **MinIO**: http://localhost:9000 (console: http://localhost:9001, credentials: `minioadmin`/`minioadmin`)
- **Valkey**: localhost:6379 (caching layer, can be disabled)

### Environment Variables

For local development, these are the key environment variables:

```bash
# Database
DATABASE_URL=postgres://testuser:testpass@localhost:5432/image_gallery_test?sslmode=disable

# Storage (MinIO local)
STORAGE_ENDPOINT=localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_BUCKET=images
STORAGE_USE_SSL=false
STORAGE_REGION=us-east-1

# Cache (Valkey)
CACHE_ENABLED=true
CACHE_ADDRESS=localhost:6379
CACHE_PASSWORD=""
CACHE_DATABASE=0
CACHE_DEFAULT_TTL=1h

# Server
PORT=8080
HOST=0.0.0.0
GO_ENV=development
```

### AWS S3 Configuration

For production or testing with real AWS S3:

```bash
# Leave access keys empty to use AWS credentials chain (EKS Pod Identity)
STORAGE_ENDPOINT=s3.amazonaws.com
STORAGE_ACCESS_KEY=""
STORAGE_SECRET_KEY=""
STORAGE_BUCKET=your-bucket-name
STORAGE_USE_SSL=true
STORAGE_REGION=us-west-2
```

## 🔍 Debugging and Troubleshooting

### Common Issues

#### Dagger Not Found
```bash
make install-tools  # Install Dagger CLI
```

#### Docker Permission Issues
```bash
# Check Docker permissions
docker ps
```

#### Database Connection Issues
```bash
# Restart database services
make db-reset
```

#### Missing Dependencies
```bash
# Update Go modules
go mod tidy

# Reinstall tools
make install-tools
```

### Getting Help

```bash
# Makefile help
make help

# Dagger module help
dagger -m MODULE_NAME functions
dagger -m MODULE_NAME call FUNCTION_NAME --help
```

## 🤝 Development Guidelines

### Code Standards
- Follow Go idioms and conventions
- Use meaningful variable and function names
- Write tests before implementation (TDD)
- Keep functions small and focused
- Document public APIs

### Git Workflow
- Use conventional commit messages
- Create feature branches from main
- Write descriptive commit messages
- Include tests with all changes

### Testing Requirements
- All new code must have unit tests
- Integration tests for repository layers
- Minimum 80% code coverage required
- Tests must pass before merging

### Performance Considerations
- Use caching appropriately (Valkey)
- Optimize database queries
- Consider memory usage for large images
- Profile performance-critical paths

## 🚀 Production Deployment

### Environment Variables

Required environment variables for production:

```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/db

# Storage
STORAGE_ENDPOINT=s3.amazonaws.com
STORAGE_ACCESS_KEY=access_key  # or leave empty for IAM roles
STORAGE_SECRET_KEY=secret_key  # or leave empty for IAM roles
STORAGE_BUCKET=images
STORAGE_USE_SSL=true
STORAGE_REGION=us-west-2

# Cache
CACHE_ENABLED=true
CACHE_ADDRESS=cache-host:6379
CACHE_PASSWORD=cache_password
CACHE_DATABASE=0
CACHE_DEFAULT_TTL=1h

# Server
PORT=8080
HOST=0.0.0.0
GO_ENV=production
```

### Production Build

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server cmd/server/main.go

# Or use Docker
docker build -t image-gallery .

# Or use Dagger for multi-arch builds
make dagger-docker
```

## 📝 API Documentation

The application exposes RESTful APIs for image management:

- `GET /api/images` - List images with pagination
- `POST /api/images` - Upload new image
- `GET /api/images/:id` - Get specific image
- `PUT /api/images/:id` - Update image metadata
- `DELETE /api/images/:id` - Delete image

For detailed API documentation, start the server and visit `/docs` (when implemented).