# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `make build` - Build the application binary to `./bin/server`
- `make run` - Build and run the application locally
- `make dev` - Run with hot reload using air (installs air if needed)

### Testing
- `make test` - Run all unit tests
- `make test-coverage` - Run tests with coverage report (generates coverage.html)

### Code Quality
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint (installs if needed)
- `make vet` - Run go vet

### Docker and Infrastructure
- `docker-compose up -d` - Start PostgreSQL, MinIO, and Redis services
- `docker-compose up -d postgres minio` - Start only database and storage for local development
- `make docker-build` - Build Docker image
- `make docker-up` - Start all services including the app
- `make docker-down` - Stop all services

### Database Management
- `make migrate` - Run database migrations (uses Atlas if available, falls back to built-in)
- `make db-start` - Start only PostgreSQL service
- `make db-reset` - Reset database with fresh schema
- Atlas commands: `atlas-validate`, `atlas-inspect`, `atlas-diff`, `atlas-apply`

## Architecture Overview

This is a clean architecture Go application with strict separation of concerns:

### Core Layers
- **Domain** (`internal/domain/`): Pure business logic with no external dependencies
  - `image/`: Contains Image, Tag, Album domain models, interfaces, events, and validation
- **Services** (`internal/services/`): Application logic orchestrating domain operations
  - `implementations/`: Concrete service implementations and repository adapters
  - `integrationtests/`: Integration test suites using testcontainers
  - `container.go`: Dependency injection container
- **Platform** (`internal/platform/`): Infrastructure implementations
  - `database/`: PostgreSQL repositories and migration management
  - `storage/`: MinIO S3-compatible storage service
  - `server/`: HTTP server configuration
- **Web** (`internal/web/`): HTTP handlers and routing

### Key Design Patterns
- **Dependency Injection**: All services are wired through a container in `internal/services/container.go`
- **Repository Pattern**: Domain interfaces implemented by platform layer
- **Clean Architecture**: Dependencies point inward toward domain
- **Test-Driven Development**: Comprehensive unit and integration tests

### Technology Stack
- **Runtime**: Go 1.24
- **Database**: PostgreSQL 15 with Atlas schema management
- **Storage**: MinIO S3-compatible object storage
- **Testing**: Testcontainers for isolated integration tests
- **HTTP**: Chi router for REST API endpoints

### Development Environment
- PostgreSQL runs on port 5432 (user: testuser, password: testpass, db: image_gallery_test)
- MinIO runs on port 9000 (console on 9001, credentials: minioadmin/minioadmin)
- Redis runs on port 6379 (optional, for future caching)
- Application runs on port 8080

### Storage Configuration
The application supports multiple storage backends:

#### Local Development (MinIO)
```bash
STORAGE_ENDPOINT=localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_USE_SSL=false
```

#### AWS S3 with Static Credentials
```bash
STORAGE_ENDPOINT=s3.amazonaws.com
STORAGE_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
STORAGE_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
STORAGE_USE_SSL=true
```

#### AWS S3 with EKS Pod Identity / IAM Roles
```bash
STORAGE_ENDPOINT=s3.amazonaws.com
STORAGE_ACCESS_KEY=
STORAGE_SECRET_KEY=
STORAGE_USE_SSL=true
```
When credentials are empty, the application uses AWS credentials chain (Pod Identity, IAM roles, credentials file, environment variables).

### API Endpoints
- `GET /api/images` - List images with pagination
- `POST /api/images` - Upload new image
- `GET /api/images/:id` - Get specific image
- `GET /api/images/:id/view` - View image (proxy endpoint)
- `PUT /api/images/:id` - Update image metadata
- `DELETE /api/images/:id` - Delete image

### Testing Strategy
- **Unit Tests**: Fast, isolated tests alongside implementation files
- **Integration Tests**: Real database/storage using testcontainers in `internal/services/integrationtests/`
- **Test Utilities**: Shared test infrastructure in `internal/testutils/`

### Important Notes
- Always run infrastructure services before running the application locally
- Integration tests require Docker to be running
- The application follows strict TDD methodology
- All new code requires corresponding tests
- Use the dependency injection container for service resolution