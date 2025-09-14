# Release Coordination: Docker Images and Versions

This document explains how Docker image tags are coordinated with release-please versions to ensure consistency across releases.

## Overview

The release process uses two main workflows:
1. **`build-push.yml`** - Builds and pushes development images
2. **`release-please.yml`** - Creates releases and versioned images

## Version Tagging Strategy

### Development Images (build-push.yml)
- **Trigger**: Pushes to `main` branch after successful CI
- **Tags**:
  - `ghcr.io/smana/image-gallery:latest` (for main branch)
  - `ghcr.io/smana/image-gallery:main-<sha>` (SHA-based tags)
- **Purpose**: Continuous deployment of latest development changes

### Release Images (release-please.yml)
- **Trigger**: When release-please creates a new release
- **Tags**:
  - `ghcr.io/smana/image-gallery:v1.2.3` (semantic version)
- **Purpose**: Stable, versioned releases

## Coordination Mechanisms

### 1. Avoiding Duplicate Images
The `build-push.yml` workflow includes logic to skip building when release-please is creating a release:

```bash
# Check if this is a release commit
if git log -1 --pretty=format:"%s" | grep -E "^(chore|release).*release"; then
  echo "Release commit detected - skipping to avoid duplicate images"
  exit 0
fi
```

### 2. Version Injection
Release images include the version in the binary:
```bash
--ldflags="-w -s -X main.version=${{ needs.release-please.outputs.version }}"
```

### 3. Proper Container Labels
Release images include OCI-compliant labels:
```bash
with-label org.opencontainers.image.version=${{ needs.release-please.outputs.version }}
with-label org.opencontainers.image.source=${{ github.server_url }}/${{ github.repository }}
with-label org.opencontainers.image.revision=${{ github.sha }}
```

## Usage Examples

### Pull Latest Development Image
```bash
docker pull ghcr.io/smana/image-gallery:latest
```

### Pull Specific Release Version
```bash
docker pull ghcr.io/smana/image-gallery:v1.2.3
```

### Check Version in Running Container
```bash
docker run --rm ghcr.io/smana/image-gallery:v1.2.3 --version
```

## Release Process Flow

1. **Development**:
   - Push changes to feature branches
   - CI validates changes
   - Merge to `main` triggers development image build

2. **Release Creation**:
   - release-please analyzes conventional commits
   - Creates PR with version bump and CHANGELOG
   - Merge PR triggers release workflow

3. **Release Workflow**:
   - Creates GitHub release
   - Builds versioned Docker image
   - Uploads release artifacts (binaries, checksums)
   - Runs security scans on release image

4. **Deployment**:
   - Use versioned images for production deployments
   - Use latest images for development/staging

## Configuration Files

### release-please-config.json
```json
{
  ".": {
    "release-type": "go",
    "bump-minor-pre-major": true,
    "bump-patch-for-minor-pre-major": true,
    "draft": false,
    "prerelease": false
  }
}
```

### .release-please-manifest.json
```json
{
  ".": "0.1.0"
}
```

## Best Practices

1. **Use Semantic Versioning**: Follow conventional commits for automatic version bumping
2. **Pin Production Images**: Use specific version tags in production, not `latest`
3. **Security Scanning**: All release images are automatically scanned with Trivy
4. **Multi-Architecture**: Images are built for `linux/amd64` and `linux/arm64`
5. **Distroless Base**: Release images use distroless base for security

## Troubleshooting

### Issue: Duplicate Images
- **Cause**: Both workflows building simultaneously
- **Solution**: The build-push workflow now detects release commits and skips

### Issue: Missing Version Tag
- **Cause**: release-please workflow failed or not triggered
- **Solution**: Check conventional commit format and workflow logs

### Issue: Version Mismatch
- **Cause**: Manual version changes not reflected in release-please manifest
- **Solution**: Update `.release-please-manifest.json` manually if needed