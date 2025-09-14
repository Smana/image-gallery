# Changelog

This file tracks all notable changes to the Image Gallery project.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.0.0 (2025-09-14)


### Features

* add comprehensive configuration validation and enhanced setting… ([621107a](https://github.com/Smana/image-gallery/commit/621107aaf6c8ae5715fd187918f0e8a085886e96))
* add comprehensive configuration validation and enhanced settings management ([ae1912f](https://github.com/Smana/image-gallery/commit/ae1912fb601d039a1ee977bade1db929c4d1df8d))
* add comprehensive database layer testing with mocks and integration tests ([be96c99](https://github.com/Smana/image-gallery/commit/be96c992f76623a9928b357fd5c7b42df5e5c436))
* add valkey support ([c95011c](https://github.com/Smana/image-gallery/commit/c95011c7ba38f9c698a60caa1092919934ccb8ed))
* add valkey support ([537d44f](https://github.com/Smana/image-gallery/commit/537d44f7551e02c6f2adfbd7dc72a287a5a77655))
* **aws:** being able to use EKS pod Identity ([9321253](https://github.com/Smana/image-gallery/commit/93212539462992abc6fc29844ee24eec1e0e818c))
* **ci:** configure release-please ([fb827b2](https://github.com/Smana/image-gallery/commit/fb827b2aba2101787d088d0dd01b4be7c894b110))
* **ci:** integrate GoReleaser for standardized build and release process ([6662b0e](https://github.com/Smana/image-gallery/commit/6662b0e6802cb2475859fd3715f7ed116e91a807))
* **ci:** use dagger for ci steps ([3a421a9](https://github.com/Smana/image-gallery/commit/3a421a93eda8271a2cc7e2e2a250b59a5f8ad195))
* enhance storage service with object listing capabilities ([053c2c4](https://github.com/Smana/image-gallery/commit/053c2c48c19719470052670790650a6dafd2a7cf))
* implement complete image gallery with viewing capabilities ([677058f](https://github.com/Smana/image-gallery/commit/677058f4360ee791c9ed98300a38d844e7de3fa7))
* implement comprehensive domain layer with business logic, validation, and events ([56a2ea9](https://github.com/Smana/image-gallery/commit/56a2ea9f2a180013aa1848cb1ff8e5c94c28c016))
* implement comprehensive storage layer with MinIO integration ([00eb2f3](https://github.com/Smana/image-gallery/commit/00eb2f36a90297b3068eaef3fff933a73bfa8d9b))
* implement comprehensive testcontainers integration with PostgreSQL and MinIO ([df354b9](https://github.com/Smana/image-gallery/commit/df354b9953f28a1401a0b383bf404d0a582203c0))
* implement dependency injection container with service interfaces ([f9c78e2](https://github.com/Smana/image-gallery/commit/f9c78e273dc4edec859b47b716f0e5f149b1fb95))
* implement TDD repository pattern with comprehensive unit testing ([cceacae](https://github.com/Smana/image-gallery/commit/cceacae41e14f88f2cd7464eea9486a7df29e1a8))
* integrate Atlas database schema management system ([1f8107f](https://github.com/Smana/image-gallery/commit/1f8107fd9a70dab6ee62facbad97ac0ab9758dd0))
* modernize project structure following Go 2025 best practices ([17ae35b](https://github.com/Smana/image-gallery/commit/17ae35b2863b6a031a4bb2c0835cba741bdf9af8))
* modernize project structure following Go 2025 best practices ([a366cee](https://github.com/Smana/image-gallery/commit/a366cee3d59cab3dc07e593854b271dc02bcc2af))
* **release:** coordinate release-please versions with Docker image tags ([925e6f4](https://github.com/Smana/image-gallery/commit/925e6f4fd1ffde76975d303381a3581f72f25b93))


### Bug Fixes

* add missing application service to docker-compose ([b9c2f39](https://github.com/Smana/image-gallery/commit/b9c2f3915afd859b45abf04ee318a3f92238bcb1))
* **ci:** add missing --platform flag to dagger with-platform commands ([35e28aa](https://github.com/Smana/image-gallery/commit/35e28aaedd5ef757d8d47fb27cf30f38b4239097))
* **ci:** correct Dagger module syntax for build command ([01b090e](https://github.com/Smana/image-gallery/commit/01b090edce2729bb9323627add0f8f061c922518))
* **ci:** correct package path for binary build ([a2a0191](https://github.com/Smana/image-gallery/commit/a2a0191e67ec383b4027e2144040ada6542ba549))
* **ci:** resolve shell syntax errors in build-push workflow ([50b3e80](https://github.com/Smana/image-gallery/commit/50b3e80e825d08f14ccf6799dd6bede097290fef))
* **ci:** resolve shell syntax errors in dagger build steps ([53e5f7a](https://github.com/Smana/image-gallery/commit/53e5f7a188b0f797279b484fe1b8439ce391a53c))
* **ci:** use full module path for Go package build ([ccc72d4](https://github.com/Smana/image-gallery/commit/ccc72d410ecb48d1570de5911a993a44966e7d7c))
* downgrade Go version and clean up dependencies ([cf6cf99](https://github.com/Smana/image-gallery/commit/cf6cf99d5ce759a9b9bf6110590728f7be99cc4d))
* update Docker configuration for Go 1.24 compatibility ([2402410](https://github.com/Smana/image-gallery/commit/2402410b3b0a20ec19d19578b9ceaa827e2ea964))


### Documentation

* comprehensive cleanup and README update ([9f93283](https://github.com/Smana/image-gallery/commit/9f9328386005854040dadc0040e494fbfa91479f))
* refactor structure ([5077dc8](https://github.com/Smana/image-gallery/commit/5077dc8e4cd6481a5199ffc77559af5c3206d8bf))


### Code Refactoring

* enhance project structure following golang-standards layout ([cbf7d1b](https://github.com/Smana/image-gallery/commit/cbf7d1b35afdc92b0889f8882d58541028f4be3b))
* rename repository from golang-helloworld to image-gallery ([27f0be9](https://github.com/Smana/image-gallery/commit/27f0be92a9d6f41369a1dbf3ef91c11f23eac598))

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
