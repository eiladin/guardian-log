# GitHub Actions CI/CD

Complete guide for GitHub Actions workflows in Guardian-Log.

## Overview

This project uses GitHub Actions for:
- ✅ **Docker releases** - Multi-arch builds pushed to GHCR on tag
- ✅ **PR validation** - Build and test on every PR commit
- ✅ **Testing** - Run Go and frontend tests automatically

## Quick Start

### Release a New Version

```bash
# 1. Test everything locally
make test
make lint
make docker-build

# 2. Create and push a version tag
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions automatically:
#    - Builds for amd64 + arm64
#    - Pushes to ghcr.io/OWNER/guardian-log
#    - Tags: v1.0.0, v1.0, v1, sha-xxxxx

# 4. Pull and use
docker pull ghcr.io/OWNER/guardian-log:1.0.0
```

### Work on a Pull Request

```bash
# 1. Create feature branch and make changes
git checkout -b feature/my-feature
# ... make changes ...
git push origin feature/my-feature

# 2. Create PR on GitHub

# 3. GitHub Actions automatically:
#    - Runs all tests (Go + frontend)
#    - Builds Docker image (amd64)
#    - Comments on PR with status
#    - Must pass before merge
```

## Workflows

### 1. Docker Release (`.github/workflows/docker-release.yml`)

**Trigger:** Push a version tag (`v*.*.*`)

**Multi-Architecture:**
- `linux/amd64` - Intel/AMD processors
- `linux/arm64` - ARM processors (Raspberry Pi, M1 Mac, etc.)

**Outputs:**
```
ghcr.io/OWNER/guardian-log:1.2.3   # Exact version
ghcr.io/OWNER/guardian-log:1.2     # Latest 1.2.x
ghcr.io/OWNER/guardian-log:1       # Latest 1.x.x
ghcr.io/OWNER/guardian-log:sha-... # Specific commit
```

**Build time:** ~8-12 minutes (multi-arch)

**Features:**
- Build caching for speed
- Automated tagging
- Build attestations for security
- Zero configuration needed

### 2. Docker PR Build (`.github/workflows/docker-pr.yml`)

**Trigger:** Commits to PR targeting `main` or `develop`

**What it does:**
1. Builds Docker image for amd64 (fast)
2. Tests container starts successfully
3. Comments on PR with status
4. Uploads artifact (1 day retention)

**Skips if:** Only docs or non-code files changed

**Build time:** ~3-5 minutes (single arch + cache)

**Features:**
- Fast feedback on PRs
- Automatic PR comments
- Container smoke test
- Artifact for manual testing

### 3. Test (`.github/workflows/test.yml`)

**Trigger:** PR commits or push to `main`/`develop`

**Jobs:**

**test-go:**
- Go version check
- Run `go vet`
- Run tests with race detector
- Generate coverage report
- Upload to Codecov (optional)

**test-frontend:**
- Node.js setup
- Install dependencies
- Run ESLint
- Build frontend
- Verify build output

**lint:**
- Run golangci-lint
- Check code quality
- Enforce standards

**Build time:** ~2-4 minutes

## Semantic Versioning

Use semantic version tags: `vMAJOR.MINOR.PATCH`

### Examples

**Major version (breaking changes):**
```bash
git tag v2.0.0  # Breaking changes
```

**Minor version (new features):**
```bash
git tag v1.1.0  # New features, backward compatible
```

**Patch version (bug fixes):**
```bash
git tag v1.0.1  # Bug fixes only
```

**Pre-release:**
```bash
git tag v1.0.0-alpha.1  # Alpha
git tag v1.0.0-beta.1   # Beta
git tag v1.0.0-rc.1     # Release candidate
```

## Using Released Images

### Pull from GitHub Container Registry

**Public package:**
```bash
docker pull ghcr.io/OWNER/guardian-log:1.0.0
```

**Private package:**
```bash
# Create Personal Access Token with read:packages
echo $GITHUB_PAT | docker login ghcr.io -u USERNAME --password-stdin
docker pull ghcr.io/OWNER/guardian-log:1.0.0
```

### Run Container

```bash
docker run -d \
  --name guardian-log \
  -p 8080:8080 \
  -v ./data:/app/data \
  --env-file .env \
  ghcr.io/OWNER/guardian-log:1.0.0
```

### Docker Compose

```yaml
version: '3.8'
services:
  guardian-log:
    image: ghcr.io/OWNER/guardian-log:1.0.0
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    env_file:
      - .env
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guardian-log
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: guardian-log
        image: ghcr.io/OWNER/guardian-log:1.0.0
        ports:
        - containerPort: 8080
```

## Making a Release

### Complete Release Checklist

- [ ] **1. Update version references**
  ```bash
  # Update README, docs, etc.
  vim README.md
  ```

- [ ] **2. Run tests locally**
  ```bash
  make test
  make lint
  ```

- [ ] **3. Test Docker build**
  ```bash
  make docker-build
  make docker-run
  # Test at http://localhost:8080
  ```

- [ ] **4. Commit any changes**
  ```bash
  git add .
  git commit -m "Prepare v1.0.0 release"
  git push origin main
  ```

- [ ] **5. Create annotated tag**
  ```bash
  git tag -a v1.0.0 -m "Release v1.0.0

  Changes:
  - Feature 1
  - Feature 2
  - Bug fix 3"
  ```

- [ ] **6. Push tag**
  ```bash
  git push origin v1.0.0
  ```

- [ ] **7. Monitor GitHub Actions**
  - Go to: `github.com/OWNER/REPO/actions`
  - Watch "Docker Release" workflow
  - ~10 minutes for multi-arch build

- [ ] **8. Verify image**
  ```bash
  docker pull ghcr.io/OWNER/guardian-log:1.0.0
  docker run --rm ghcr.io/OWNER/guardian-log:1.0.0 --version
  ```

- [ ] **9. Create GitHub Release**
  - Go to: `github.com/OWNER/REPO/releases`
  - Draft new release
  - Select tag: `v1.0.0`
  - Add release notes
  - Publish

### Automated Release Notes

GitHub can auto-generate release notes:

1. Go to Releases → Draft new release
2. Select tag
3. Click "Generate release notes"
4. Review and edit
5. Publish

## Package Visibility

By default, GitHub packages are **private**.

### Make Package Public

1. Go to repository
2. Packages (right sidebar)
3. Click package name
4. Package settings
5. Change visibility → Public
6. Confirm

### Permissions

The package inherits repository permissions:
- Repository members can pull private packages
- Everyone can pull public packages
- No additional setup needed

## Caching

All workflows use GitHub Actions cache:

**Go modules:**
- Cached by `setup-go` action
- Keyed by `go.sum`

**npm packages:**
- Cached by `setup-node` action
- Keyed by `package-lock.json`

**Docker layers:**
- Cached with `type=gha`
- Speeds up rebuilds significantly

**Cache limits:**
- 10 GB per repository
- Oldest caches evicted first
- 7 day retention

## Debugging Workflows

### View Logs

1. Go to repository → Actions
2. Click workflow run
3. Click failed job
4. Expand step to see logs
5. Download logs (top right) for offline viewing

### Test Locally

**Docker build:**
```bash
make docker-build
```

**Go tests:**
```bash
make test
```

**Frontend build:**
```bash
cd web && npm run build
```

**Lint:**
```bash
make lint
```

### Re-run Failed Workflow

1. Go to failed workflow run
2. Click "Re-run all jobs" (top right)
3. Or "Re-run failed jobs"

### Common Issues

**Permission denied:**
- Check workflow file permissions
- Ensure `GITHUB_TOKEN` is available
- Verify package/contents permissions in workflow

**Build timeout:**
- Default: 6 hours
- Multi-arch can take 10+ minutes
- Increase: `timeout-minutes: 30` in job

**Cache miss:**
- New branches don't inherit cache
- First build will be slower
- Subsequent builds use cache

**Multi-arch fails:**
- Check QEMU step
- Try single platform first
- Verify Dockerfile syntax

## Status Badges

Add to `README.md`:

```markdown
![Docker Release](https://github.com/OWNER/REPO/actions/workflows/docker-release.yml/badge.svg)
![Docker PR](https://github.com/OWNER/REPO/actions/workflows/docker-pr.yml/badge.svg)
![Tests](https://github.com/OWNER/REPO/actions/workflows/test.yml/badge.svg)
```

Replace `OWNER/REPO` with your repository.

## Security

### Secrets

**No secrets required!** Workflows use `GITHUB_TOKEN` (automatic).

**Optional:**
- `CODECOV_TOKEN` - For coverage reports

### Permissions

Workflows use minimal permissions:

**docker-release.yml:**
```yaml
permissions:
  contents: read      # Read code
  packages: write     # Push to GHCR
  id-token: write     # Attestations
```

**docker-pr.yml:**
```yaml
permissions:
  contents: read         # Read code
  pull-requests: write   # Comment on PRs
```

### Build Attestations

Release builds create attestations:
- Verifiable build provenance
- Cryptographically signed
- Links image to source code
- Stored in GHCR metadata

## Advanced Usage

### Manual Workflow Trigger

Add to workflow:
```yaml
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to build'
        required: true
        default: 'latest'
```

Then trigger from Actions tab.

### Matrix Builds

Build multiple versions:
```yaml
strategy:
  matrix:
    go: ['1.25', '1.24']
    os: [ubuntu-latest, macos-latest]
```

### Scheduled Runs

Run nightly builds:
```yaml
on:
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight
```

## Best Practices

1. **Test locally first**
   - Always run `make test` and `make docker-build`
   - Catch issues before CI

2. **Keep workflows fast**
   - Use caching
   - Run only necessary tests
   - Parallel jobs when possible

3. **Semantic versioning**
   - Use meaningful version numbers
   - Follow semver rules
   - Tag major.minor.patch

4. **Good commit messages**
   - Clear and descriptive
   - Reference issues: `Fixes #123`
   - Include context

5. **Monitor builds**
   - Check Actions tab regularly
   - Fix failures quickly
   - Don't accumulate broken builds

## Examples

See [`.github/workflows/README.md`](.github/workflows/README.md) for detailed examples.

## Summary

**For releases:**
```bash
git tag v1.0.0
git push origin v1.0.0
# → Multi-arch build → ghcr.io
```

**For PRs:**
```bash
git push origin feature-branch
# → Auto build + test → PR comment
```

**For development:**
```bash
make dev-backend
make dev-frontend
# → Local hot-reload development
```

---

**Questions?** Check [.github/workflows/README.md](.github/workflows/README.md) for complete documentation.
