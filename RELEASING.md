# Release Process

This document describes how to create a new release of lazygit-mcp-bridge.

## Prerequisites

1. **Permissions**: You need write access to the repository
2. **Homebrew Tap**: For automatic Homebrew tap updates, you need:
   - A GitHub personal access token with `repo` scope
   - The token added as `HOMEBREW_TAP_GITHUB_TOKEN` secret in the repository settings

## Release Steps

### 1. Prepare the Release

1. **Create a release branch** (optional but recommended):
   ```bash
   git checkout -b release/v1.0.0
   ```

2. **Update version information** if needed:
   - Version information is automatically injected during build
   - No manual version updates needed in code

3. **Update documentation** if necessary:
   - README.md
   - docs/ files
   - CHANGELOG.md (if maintained)

4. **Test the build**:
   ```bash
   make build
   ./build/lazygit-mcp-bridge --version
   ./build/lazygit-mcp-bridge --help
   ```

### 2. Create and Push the Tag

1. **Commit any final changes**:
   ```bash
   git add .
   git commit -s -m "Prepare for v1.0.0 release"
   ```

2. **Create the release tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   ```

3. **Push the tag**:
   ```bash
   git push origin v1.0.0
   ```

### 3. Automated Release Process

Once you push the tag, GitHub Actions will automatically:

1. **Run tests** on Go 1.24 and 1.25
2. **Build binaries** for multiple platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64, arm64)
3. **Create GitHub release** with:
   - Release notes
   - Binary attachments
   - Checksums
4. **Update Homebrew tap** (if configured)

### 4. Verify the Release

1. **Check GitHub releases page**:
   ```
   https://github.com/cappyzawa/lazygit-mcp-bridge/releases
   ```

2. **Test installation**:
   ```bash
   # Via Go install
   go install github.com/cappyzawa/lazygit-mcp-bridge/cmd/lazygit-mcp-bridge@v1.0.0
   
   # Via Homebrew (if tap is configured)
   brew tap cappyzawa/tap
   brew install lazygit-mcp-bridge
   ```

3. **Verify version**:
   ```bash
   lazygit-mcp-bridge --version
   ```

## Post-Release

1. **Merge release branch** (if used):
   ```bash
   git checkout main
   git merge release/v1.0.0
   git push origin main
   ```

2. **Update documentation** if needed:
   - Update installation instructions in README.md
   - Update any version-specific documentation

## Troubleshooting

### Release Failed

1. **Check GitHub Actions logs**:
   - Go to Actions tab in GitHub repository
   - Check the failed workflow for details

2. **Common issues**:
   - Missing `GITHUB_TOKEN`: This should be automatically available
   - Missing `HOMEBREW_TAP_GITHUB_TOKEN`: Add this secret if you want Homebrew tap updates
   - Build failures: Check Go version compatibility and dependencies

### Tag Issues

1. **Delete and recreate tag** (if needed):
   ```bash
   # Delete local tag
   git tag -d v1.0.0
   
   # Delete remote tag
   git push origin :refs/tags/v1.0.0
   
   # Recreate and push
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

## Semantic Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes
- **MINOR** version (v1.1.0): New functionality in a backward compatible manner
- **PATCH** version (v1.0.1): Backward compatible bug fixes

## Configuration Files

### `.goreleaser.yml`
- Defines build targets, archives, and release configuration
- Handles cross-compilation for multiple platforms
- Manages Homebrew tap updates

### `.github/workflows/release.yml`
- Triggers on tag push
- Uses GoReleaser to build and release
- Requires Go 1.24+

### `.github/workflows/test.yml`
- Runs on pull requests and main branch pushes
- Tests on Go 1.24 and 1.25
- Includes build verification

## Examples

### Creating v1.0.1 Patch Release

```bash
# Make bug fixes
git commit -s -m "Fix message duplication issue"

# Tag and release
git tag -a v1.0.1 -m "Release v1.0.1 - Fix message duplication"
git push origin v1.0.1
```

### Creating v1.1.0 Feature Release

```bash
# Add new features
git commit -s -m "Add support for custom message formatting"

# Tag and release  
git tag -a v1.1.0 -m "Release v1.1.0 - Add custom message formatting"
git push origin v1.1.0
```