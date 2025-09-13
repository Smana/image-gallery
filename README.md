# Image Gallery - Modern Go Application with TDD

A comprehensive image gallery application built with Go 1.25, following Test-Driven Development methodology, clean architecture principles, and modern Go best practices.

## 🚀 Features

- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Test-Driven Development**: Comprehensive unit and integration testing
- **Dependency Injection**: Container-based DI for better testability
- **Database Management**: Atlas schema management with PostgreSQL
- **Object Storage**: MinIO S3-compatible storage for images
- **Valkey Caching**: Performance optimization with Valkey caching layer
- **Testcontainers**: Docker-based integration testing
- **Modern Tooling**: Latest Go dependencies and development tools

## 🏗️ Project Structure

```
.
├── cmd/server/             # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── domain/            # Domain layer (entities, interfaces, business logic)
│   │   └── image/         # Image domain with models, validation, events
│   ├── platform/          # Infrastructure layer
│   │   ├── database/      # Database repositories and models
│   │   ├── storage/       # File storage (MinIO) implementations
│   │   └── server/        # HTTP server setup
│   ├── services/          # Application services layer
│   │   ├── implementations/ # Service implementations and adapters
│   │   └── integrationtests/ # Integration test suites
│   ├── testutils/         # Testing utilities and containers
│   └── web/              # Web layer (handlers, middleware)
├── migrations/           # Database migration files
├── docker-compose.yml    # Development environment
├── atlas.hcl            # Atlas database schema management
├── Makefile             # Development commands
└── README.md            # This file
```

## 📋 Prerequisites

- **Go 1.25+** - Latest version required
- **Docker & Docker Compose** - For local development environment
- **Make** - For running development commands
- **Atlas CLI** - For database schema management (optional)

## 🛠️ Local Development Setup

### 1. Clone and Setup

```bash
git clone <repository-url>
cd image-gallery
```

### 2. Start Development Environment

```bash
# Start PostgreSQL, MinIO, and Valkey services
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 3. Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
```

### 4. Database Setup

```bash
# Run database migrations
make migrate-up

# Or using Atlas (if installed)
atlas migrate apply --env local
```

### 5. Build and Run

```bash
# Build the application
make build

# Run the server
./server

# Or run directly with Go
go run cmd/server/main.go
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
docker-compose up -d postgres minio

# Run integration tests
make test-integration

# Run specific integration test
go test -v ./internal/services/integrationtests/...
```

### Test with Testcontainers

The project includes comprehensive testcontainers setup for isolated integration testing:

```bash
# Run tests that spin up their own containers
go test -v ./internal/testutils/...
```

## 📊 Code Quality

### Linting

```bash
# Run linter
make lint

# Auto-fix issues
make lint-fix
```

### Code Coverage

```bash
# Generate coverage report
make coverage

# Open coverage report in browser
make coverage-html
```

## 🗄️ Database Management

### Migrations with Atlas

```bash
# Generate new migration
atlas migrate diff --env local

# Apply migrations
atlas migrate apply --env local

# Check migration status
atlas migrate status --env local
```

### Manual Migrations

```bash
# Create new migration
make migrate-create name=add_new_table

# Apply migrations
make migrate-up

# Rollback migrations  
make migrate-down
```

## 🐳 Docker Development

### Build Docker Image

```bash
make docker-build
```

### Development with Docker

```bash
# Full development environment (all services including app)
docker-compose up --build

# Just the infrastructure dependencies (for local Go development)
docker-compose up postgres minio redis

# Run only the app service (assumes infrastructure is running)
docker-compose up app

# View logs
docker-compose logs -f app

# Follow logs for specific services
docker-compose logs -f postgres minio
```

## 🔧 Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Run the application
make test           # Run unit tests
make test-coverage  # Run tests with coverage
make test-integration # Run integration tests
make lint           # Run linter
make lint-fix       # Fix linter issues
make clean          # Clean build artifacts
make docker-build   # Build Docker image
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
```

## 🏛️ Architecture Overview

### Domain Layer (`internal/domain/`)
- **Pure business logic** with no external dependencies
- **Rich domain models** with validation and business rules
- **Domain interfaces** defining contracts for infrastructure

### Application Layer (`internal/services/`)
- **Use case implementations** orchestrating domain logic
- **Repository adapters** bridging domain and infrastructure
- **Dependency injection container** for service management

### Infrastructure Layer (`internal/platform/`)
- **Database repositories** implementing domain interfaces
- **External service integrations** (MinIO, etc.)
- **Infrastructure concerns** (logging, monitoring)

### Presentation Layer (`internal/web/`)
- **HTTP handlers** and routing
- **Request/response models**
- **Middleware** for cross-cutting concerns

## 🧪 Testing Strategy

### Test Pyramid Implementation

1. **Unit Tests (Fast, Isolated)**
   - Domain model validation
   - Repository adapters with mocking
   - Service logic with dependency injection
   - Located alongside implementation files

2. **Integration Tests (Medium Speed)**
   - Repository implementations with testcontainers
   - Service integration with real databases
   - Located in `internal/services/integrationtests/`

3. **End-to-End Tests (Slow, Complete)**
   - Full application testing with HTTP requests
   - Complete user journey testing
   - Located in dedicated test packages

### TDD Methodology

The project follows strict Test-Driven Development:
1. **Red**: Write failing tests first
2. **Green**: Implement minimum code to pass
3. **Refactor**: Improve code while keeping tests green

## 🚀 Deployment

### Environment Variables

Required environment variables for production:

```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/db

# Storage
STORAGE_ENDPOINT=minio:9000
STORAGE_ACCESS_KEY=access_key
STORAGE_SECRET_KEY=secret_key
STORAGE_BUCKET=images

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

## 📝 API Documentation

The application exposes RESTful APIs for image management:

- `GET /api/images` - List images with pagination
- `POST /api/images` - Upload new image
- `GET /api/images/:id` - Get specific image
- `PUT /api/images/:id` - Update image metadata
- `DELETE /api/images/:id` - Delete image

For detailed API documentation, start the server and visit `/docs` (when implemented).

## 🎯 Next Steps

- [ ] Implement HTMX-powered web interface
- [ ] Add comprehensive API documentation  
- [ ] Set up CI/CD pipeline with GitHub Actions
- [ ] Add metrics and monitoring
- [ ] Implement caching layer
- [ ] Add search functionality

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.