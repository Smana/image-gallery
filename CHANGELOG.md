# Changelog

This file tracks all notable changes to the Image Gallery project.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup with Go 1.25
- Clean architecture implementation with domain-driven design
- RESTful API for image management
- PostgreSQL database with Atlas migrations
- S3-compatible storage (MinIO/AWS) with EKS Pod Identity support
- Valkey (Redis-compatible) caching layer
- Comprehensive testing with testcontainers
- Dagger-based CI/CD pipeline using community modules
- Security scanning with Trivy and govulncheck
- Multi-platform binary builds (Linux, macOS, Windows)
- Multi-architecture container images (AMD64, ARM64)
- Split GitHub Actions workflows (CI validation and build/push)
- Release Please integration for automated releases
- Conventional commits validation
- Comprehensive documentation structure

### Security
- Container vulnerability scanning with Trivy
- Go dependency vulnerability scanning with govulncheck
- Distroless container base images
- SARIF security report uploads to GitHub Security tab

### Documentation
- Complete development setup guide
- Architecture documentation with mermaid diagrams
- Security practices and incident response guide
- CI/CD pipeline documentation
- Release process with conventional commits

---

**Note**: This changelog will be automatically maintained by Release Please based on conventional commit messages starting with the first release.