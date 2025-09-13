# Documentation Index

This directory contains comprehensive documentation for the Image Gallery application.

## 📚 Documentation Structure

### Development & Setup
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Complete development setup guide
  - Prerequisites and installation
  - Local development environment
  - Testing strategies (unit, integration, Dagger)
  - Available commands and tools
  - Database management
  - Docker development

### Architecture & Design
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed architecture documentation
  - Clean architecture overview
  - Layer responsibilities and interactions
  - Data flow diagrams
  - Database design
  - Testing architecture
  - Performance considerations

### Operations & CI/CD
- **[DAGGER_CI.md](DAGGER_CI.md)** - Containerized CI/CD pipeline
  - Dagger module usage
  - GitHub Actions workflows (split CI/Build)
  - Local pipeline execution
  - Pipeline configuration and features

### Security
- **[SECURITY.md](SECURITY.md)** - Security practices and tools
  - Security scanning tools (Trivy, govulncheck)
  - CI/CD security integration
  - Container and application security
  - Best practices and incident response

## 🔗 Quick Navigation

### For Developers
Start with [DEVELOPMENT.md](DEVELOPMENT.md) for local setup, then review [ARCHITECTURE.md](ARCHITECTURE.md) for understanding the codebase structure.

### For DevOps Engineers
Focus on [DAGGER_CI.md](DAGGER_CI.md) for CI/CD pipelines and [SECURITY.md](SECURITY.md) for security practices.

### For Security Teams
Review [SECURITY.md](SECURITY.md) for comprehensive security measures and [DAGGER_CI.md](DAGGER_CI.md) for pipeline security integration.

### For Architects
Start with [ARCHITECTURE.md](ARCHITECTURE.md) for system design and [DEVELOPMENT.md](DEVELOPMENT.md) for implementation details.

## 📖 Additional Documentation

### External References
- **Go Documentation**: [pkg.go.dev](https://pkg.go.dev/)
- **Dagger Documentation**: [docs.dagger.io](https://docs.dagger.io/)
- **Atlas Migrations**: [atlasgo.io](https://atlasgo.io/)
- **Chi Router**: [github.com/go-chi/chi](https://github.com/go-chi/chi)
- **Testcontainers**: [testcontainers.com](https://testcontainers.com/)

### Tools Documentation
- **PostgreSQL**: Database setup and management
- **MinIO**: S3-compatible object storage
- **Valkey**: Redis-compatible caching
- **Trivy**: Security vulnerability scanning
- **GitHub Actions**: CI/CD workflow automation

## 🚀 Getting Started

1. **New Developer**: [DEVELOPMENT.md](DEVELOPMENT.md) → [ARCHITECTURE.md](ARCHITECTURE.md)
2. **CI/CD Setup**: [DAGGER_CI.md](DAGGER_CI.md) → [SECURITY.md](SECURITY.md)  
3. **Security Review**: [SECURITY.md](SECURITY.md) → [DAGGER_CI.md](DAGGER_CI.md)
4. **Architecture Review**: [ARCHITECTURE.md](ARCHITECTURE.md) → [DEVELOPMENT.md](DEVELOPMENT.md)

## 📝 Contributing to Documentation

When adding new documentation:

1. **Follow the structure**: Use clear headings and sections
2. **Link between docs**: Create cross-references where helpful
3. **Update this index**: Add new documents to the relevant sections
4. **Include diagrams**: Use mermaid for visual explanations
5. **Keep it current**: Update docs when features change

### Documentation Standards

- Use **Markdown** format for all documentation
- Include **mermaid diagrams** for complex workflows
- Provide **code examples** with proper syntax highlighting  
- Link to **external resources** when helpful
- Keep **table of contents** updated in longer documents