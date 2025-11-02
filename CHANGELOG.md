# Changelog

This file tracks all notable changes to the Image Gallery project.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.3](https://github.com/Smana/image-gallery/compare/v1.7.2...v1.7.3) (2025-11-02)


### Bug Fixes

* **database:** eliminate json_agg memory explosion causing immediate … ([78efae2](https://github.com/Smana/image-gallery/commit/78efae2bb2a7d15c1fd967a016514138e28abff3))
* **database:** eliminate json_agg memory explosion causing immediate OOMKills ([c82ff39](https://github.com/Smana/image-gallery/commit/c82ff39847036b625ae51a7f431f197374852c8a))
* **database:** remove deprecated json_agg code and fix test helper ([c75f4ac](https://github.com/Smana/image-gallery/commit/c75f4ac793b0ee8f98a2aa0eaf5a56eab9ff69a6))

## [1.7.2](https://github.com/Smana/image-gallery/compare/v1.7.1...v1.7.2) (2025-11-02)


### Bug Fixes

* **database:** add connection pool limits to prevent resource exhaustion ([37b13df](https://github.com/Smana/image-gallery/commit/37b13dfbf1156849ba31969b22bf91b369769c78))
* **database:** add connection pool limits to prevent resource exhaustion ([5e09eb7](https://github.com/Smana/image-gallery/commit/5e09eb7990bf6735d05293a5e4940af64020d449))
* **database:** fix code formatting in connection pool configuration ([d4e03e3](https://github.com/Smana/image-gallery/commit/d4e03e39be9b46d0500356ec2b0b8c05aff9d780))

## [1.7.1](https://github.com/Smana/image-gallery/compare/v1.7.0...v1.7.1) (2025-11-02)


### Bug Fixes

* **handlers:** fix database connection leak in slow query scenario ([5dd6b3e](https://github.com/Smana/image-gallery/commit/5dd6b3ec21af54df69f919d84ab25733ae80a779))
* **handlers:** fix database connection leak in slow query scenario ([f3ae6aa](https://github.com/Smana/image-gallery/commit/f3ae6aa7df4cddae12361763d94a5025ea606c6a))

## [1.7.0](https://github.com/Smana/image-gallery/compare/v1.6.1...v1.7.0) (2025-11-02)


### Features

* **config:** add S3 sync on startup configuration option ([d59745c](https://github.com/Smana/image-gallery/commit/d59745c9e2bf93f5b3ebefcdc613962814846ce5))
* **server:** add automatic memory limit configuration ([65d844c](https://github.com/Smana/image-gallery/commit/65d844cee03523486cc8b9982b3d267b4e521278))
* **server:** implement S3 bucket sync and automemlimit initialization ([ef005e2](https://github.com/Smana/image-gallery/commit/ef005e2f98577f5354aaef704a0ae2475345a21d))
* **settings:** set default background image with 40% opacity ([9dd4e13](https://github.com/Smana/image-gallery/commit/9dd4e131b76b116ef60ac4a137102774eff6dfca))


### Bug Fixes

* **upload:** critical memory leak fixes to prevent OOMKills ([316a264](https://github.com/Smana/image-gallery/commit/316a2644f47e04225acb263042c0b8cb178c77a9))
* **upload:** deduplicate tags to prevent validation error ([08d2c56](https://github.com/Smana/image-gallery/commit/08d2c562ebf245924bbe54453abccc2969eed678))
* **upload:** deduplicate tags to prevent validation error ([c628ea8](https://github.com/Smana/image-gallery/commit/c628ea8896c0b047863c9f9cc8fcdb965f7df295))
* **upload:** improve error handling and reduce cyclomatic complexity ([40f3d9b](https://github.com/Smana/image-gallery/commit/40f3d9b1514057479f27c020fa48e24874f9991c))

## [1.6.1](https://github.com/Smana/image-gallery/compare/v1.6.0...v1.6.1) (2025-11-01)


### Bug Fixes

* **main:** bug in defining default user-id ([7bc8def](https://github.com/Smana/image-gallery/commit/7bc8def460b1402351e5b0b66fd63286daa5cf00))
* **main:** bug in defining default user-id ([a688bbc](https://github.com/Smana/image-gallery/commit/a688bbc73ebf3e6b5816f1dbd7fe25a8213b1498))

## [1.6.0](https://github.com/Smana/image-gallery/compare/v1.5.3...v1.6.0) (2025-11-01)


### Features

* add user settings and upload handler with tag filtering system ([33a4aa7](https://github.com/Smana/image-gallery/commit/33a4aa779ce9641c6233f999936794836acfaf32))
* add user settings and upload handler with tag filtering system ([5f9d9a2](https://github.com/Smana/image-gallery/commit/5f9d9a227e92b549ee6bb0065030601789969854))


### Bug Fixes

* **tests:** correct sort field to uploaded_at in mock expectation ([23dbc13](https://github.com/Smana/image-gallery/commit/23dbc1357332febd1a3a43c62c47b43267906f67))
* **tests:** remove duplicate GetByTags mock call ([ce9d6c6](https://github.com/Smana/image-gallery/commit/ce9d6c638ec75d8519ada1fa53e889a0a57aeb4b))
* **tests:** update mock expectation for GetWithTags ([73e6367](https://github.com/Smana/image-gallery/commit/73e6367204a28b5da684c1d687232f72a4981c7b))

## [1.5.3](https://github.com/Smana/image-gallery/compare/v1.5.2...v1.5.3) (2025-10-31)


### Bug Fixes

* **observability:** add a test endpoint for database ([076ac68](https://github.com/Smana/image-gallery/commit/076ac68db14c0a1ce73e8b3684a55dd1bcec786c))
* **observability:** add a test endpoint for database ([59e2ea1](https://github.com/Smana/image-gallery/commit/59e2ea1decfc80eaa49f8bb359a4defae1079b63))

## [1.5.2](https://github.com/Smana/image-gallery/compare/v1.5.1...v1.5.2) (2025-10-31)


### Bug Fixes

* **observability:** trigger a releases for the previous change ([fdef7cd](https://github.com/Smana/image-gallery/commit/fdef7cd8b396dea91955a60b10d0eabc2e52a754))

## [1.5.1](https://github.com/Smana/image-gallery/compare/v1.5.0...v1.5.1) (2025-10-31)


### Bug Fixes

* **otel:** add missing trace sampler configuration ([d6ec07e](https://github.com/Smana/image-gallery/commit/d6ec07e178de9f3584113826c8af9d59ddf97d8e))
* **otel:** add missing trace sampler configuration ([e6a6f91](https://github.com/Smana/image-gallery/commit/e6a6f914279f472963c1a0d3095a3fd1cc6f1e1e))

## [1.5.0](https://github.com/Smana/image-gallery/compare/v1.4.2...v1.5.0) (2025-10-31)


### Features

* **observability:** add exemplars, exponential histograms, and configurable sampling ([9f20831](https://github.com/Smana/image-gallery/commit/9f20831d8287d26a100a560da1ada3f23cbb793a))
* **observability:** add exemplars, exponential histograms, and configurable sampling ([04e46e2](https://github.com/Smana/image-gallery/commit/04e46e2e1d93d2bd4fcd52f3f1fb25cd1702e5c1))

## [1.4.2](https://github.com/Smana/image-gallery/compare/v1.4.1...v1.4.2) (2025-10-30)


### Bug Fixes

* **observability:** remove WithInsecure() when using WithEndpointURL ([68a9b11](https://github.com/Smana/image-gallery/commit/68a9b117876935ba5c6cf22158457572eb7dbeda))
* **observability:** remove WithInsecure() when using WithEndpointURL ([130b51f](https://github.com/Smana/image-gallery/commit/130b51f69ffa6663e962a6a7d8e042ebce1b3fb3))

## [1.4.1](https://github.com/Smana/image-gallery/compare/v1.4.0...v1.4.1) (2025-10-29)


### Bug Fixes

* **observability:** correct OTLP endpoint URL handling ([9e402fa](https://github.com/Smana/image-gallery/commit/9e402fa56e3b4cfb24fe22ba0669c0e28a4246f2))
* **observability:** use WithEndpointURL for OTLP exporters ([78b4c46](https://github.com/Smana/image-gallery/commit/78b4c46c0b9c6bc606f06e157b16b4342c88c357))

## [1.4.0](https://github.com/Smana/image-gallery/compare/v1.3.0...v1.4.0) (2025-10-29)


### Features

* **observability:** add comprehensive OpenTelemetry instrumentation ([c7df893](https://github.com/Smana/image-gallery/commit/c7df893208798bb31b5bb6dca0bcc0daf4d2a501))
* **observability:** add comprehensive OpenTelemetry instrumentation ([6f2f695](https://github.com/Smana/image-gallery/commit/6f2f695ee14a327c0c1e21e1e3419186772f81e3))


### Bug Fixes

* **observability:** resolve linting issues - errcheck, gofmt, and cyclomatic complexity ([c77afd5](https://github.com/Smana/image-gallery/commit/c77afd56aeb5a5174307669f08256fb2b84cd496))

## [1.3.0](https://github.com/Smana/image-gallery/compare/v1.2.0...v1.3.0) (2025-10-12)


### Features

* **storage:** support EKS Pod Identity and IAM roles for S3 access ([4873b10](https://github.com/Smana/image-gallery/commit/4873b10dd10676139de7f5854ba4bec30ecc6675))
* **storage:** support EKS Pod Identity and IAM roles for S3 access ([a43960a](https://github.com/Smana/image-gallery/commit/a43960af5b0cc76ba5b72631d566920112241d56))

## [1.2.0](https://github.com/Smana/image-gallery/compare/v1.1.0...v1.2.0) (2025-10-12)


### Features

* add healthchecks handlers ([20059a3](https://github.com/Smana/image-gallery/commit/20059a3e4ba7ece4fe34ffd1a3c65bbdc5030ba1))
* add healthchecks handlers ([f424924](https://github.com/Smana/image-gallery/commit/f424924acc57847b701a6274341be71144d20e17))

## [1.1.0](https://github.com/Smana/image-gallery/compare/v1.0.5...v1.1.0) (2025-09-27)


### Features

* **atlas:** first configuration with atlas local env ([fb2e9f0](https://github.com/Smana/image-gallery/commit/fb2e9f0e86d419a97e2f0af9d71399d6ccc77aeb))
* **atlas:** first configuration with atlas local env ([0324b9c](https://github.com/Smana/image-gallery/commit/0324b9cc6a2939dd3307a75f5869925636209268))

## [1.0.5](https://github.com/Smana/image-gallery/compare/v1.0.4...v1.0.5) (2025-09-14)


### Bug Fixes

* **ci:** remove redundant security scanning ([9d1517a](https://github.com/Smana/image-gallery/commit/9d1517a1e1b7d7dd829cc9d50816f02fc0d7d618))

## [1.0.4](https://github.com/Smana/image-gallery/compare/v1.0.3...v1.0.4) (2025-09-14)


### Bug Fixes

* **ci:** add category in trivy steps ([faa7b08](https://github.com/Smana/image-gallery/commit/faa7b080ac4c0360e7649579e73524c92ebc6d8f))
* **ci:** add category in trivy steps ([cdea3dd](https://github.com/Smana/image-gallery/commit/cdea3dd747e15d9826bbbb527d0b57be90f18645))
* **ci:** integrate trivy with goreleaser ([33e9ce2](https://github.com/Smana/image-gallery/commit/33e9ce211485dfc380ec0e6526bd4b1e52edb079))
* **ci:** integrate trivy with goreleaser ([418c430](https://github.com/Smana/image-gallery/commit/418c430654bb9aedb16b9439921f8599b33854b1))
* **ci:** remove useless steps for building images ([248eabd](https://github.com/Smana/image-gallery/commit/248eabdc36e159b6754236103723d7808e01e080))
* **ci:** replace dagger module with official action ([c07e7e7](https://github.com/Smana/image-gallery/commit/c07e7e71d3a0dc6a40d6770829241afc07e497d4))
* **ci:** replace dagger module with official action ([bd3f834](https://github.com/Smana/image-gallery/commit/bd3f834ffb55a32eaf79bf2f3a837c5b06704147))
* **ci:** use the same hash for both goreleaser and trivy ([7020267](https://github.com/Smana/image-gallery/commit/7020267264b6f62f3daae4176a25259117d435f7))
* **ci:** use the same hash for both goreleaser and trivy ([3787b4e](https://github.com/Smana/image-gallery/commit/3787b4ed82200fce8f3483f2f45463176c072264))

## [1.0.3](https://github.com/Smana/image-gallery/compare/v1.0.2...v1.0.3) (2025-09-14)


### Bug Fixes

* **ci:** docker image to lower ([446e88b](https://github.com/Smana/image-gallery/commit/446e88bdc3c3565ba65fd867f6aa52fd2d5934b9))
* **ci:** docker image to lower ([52698d6](https://github.com/Smana/image-gallery/commit/52698d65f9f1f53c6f3f334a61ca659b34bb97f6))

## [1.0.2](https://github.com/Smana/image-gallery/compare/v1.0.1...v1.0.2) (2025-09-14)


### Bug Fixes

* **ci:** use official github action ([c23ea4b](https://github.com/Smana/image-gallery/commit/c23ea4b9c2a7f62dcd2b209e78dc3d7ac5113f81))
* **ci:** use official github action ([8c254f4](https://github.com/Smana/image-gallery/commit/8c254f467c873298892365bd6388b8790e39ff8c))

## [1.0.1](https://github.com/Smana/image-gallery/compare/v1.0.0...v1.0.1) (2025-09-14)


### Bug Fixes

* **ci:** gorelease workflow ([6bb2786](https://github.com/Smana/image-gallery/commit/6bb2786637785cb49ee4a0d7e778e2855a902202))
* **ci:** gorelease workflow ([55ad87e](https://github.com/Smana/image-gallery/commit/55ad87e162018b7c71e08c50b984e0d62b2d7cd4))

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
