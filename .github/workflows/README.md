# GitHub Actions Workflows

This directory contains CI/CD workflows for Guardian-Log.

## Workflows

### 1. Docker Release (`docker-release.yml`)

**Triggers:** When a version tag is pushed (e.g., `v1.0.0`)

**What it does:**
- Builds Docker image for `linux/amd64` and `linux/arm64`
- Pushes to GitHub Container Registry (ghcr.io)
- Tags with semantic version, major.minor, major, and git SHA
- Creates build attestation for security

**Usage:**
```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0

# Or create a release through GitHub UI
```

**Output:**
- `ghcr.io/OWNER/guardian-log:1.0.0`
- `ghcr.io/OWNER/guardian-log:1.0`
- `ghcr.io/OWNER/guardian-log:1`
- `ghcr.io/OWNER/guardian-log:sha-abc123`

### 2. Docker PR Build (`docker-pr.yml`)

**Triggers:** When commits are pushed to a PR targeting `main` or `develop`

**What it does:**
- Builds Docker image for `linux/amd64` (fast validation)
- Tests that the container starts successfully
- Comments on PR with build status
- Uses GitHub Actions cache for faster builds

**Skips build if:** Only documentation or non-code files changed

**Output:**
- PR comment with build status
- Docker image artifact (1 day retention)

### 3. Test (`test.yml`)

**Triggers:**
- PR commits to `main` or `develop`
- Pushes to `main` or `develop`

**What it does:**
- **Go tests:** Runs tests, vet, and race detector
- **Frontend tests:** Lints and builds React app
- **Lint:** Runs golangci-lint on Go code
- **Coverage:** Uploads test coverage to Codecov (optional)

## How to Use

### Creating a Release

1. **Ensure all tests pass:**
   ```bash
   make test
   make lint
   make build
   ```

2. **Build and test Docker locally:**
   ```bash
   make docker-build
   make docker-run
   # Test at http://localhost:8080
   ```

3. **Create and push tag:**
   ```bash
   # Tag format: v<major>.<minor>.<patch>
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **Monitor GitHub Actions:**
   - Go to your repository → Actions tab
   - Watch the "Docker Release" workflow
   - Takes ~5-10 minutes for multi-arch build

5. **Pull and use the image:**
   ```bash
   # Pull from GHCR
   docker pull ghcr.io/OWNER/guardian-log:1.0.0

   # Run it
   docker run -p 8080:8080 \
     -e AGH_URL=http://192.168.1.1:8080 \
     -e GEMINI_API_KEY=your-key \
     ghcr.io/OWNER/guardian-log:1.0.0
   ```

### Working with Pull Requests

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes and commit:**
   ```bash
   git add .
   git commit -m "Add new feature"
   git push origin feature/my-feature
   ```

3. **Create PR on GitHub**

4. **Workflows run automatically:**
   - `test.yml` - Runs tests and lints
   - `docker-pr.yml` - Builds Docker image
   - Both must pass before merge

5. **Check PR comments:**
   - Bot comments with Docker build status
   - See test results in "Checks" tab

### Debugging Failed Builds

**Docker build fails:**
```bash
# Test locally first
make docker-build

# Check Dockerfile syntax
docker build -t test .

# View GitHub Actions logs
# Go to Actions → Failed workflow → Click job → View logs
```

**Tests fail:**
```bash
# Run tests locally
make test

# Run linter
make lint

# Fix issues and push again
```

## Environment Variables & Secrets

### Required Secrets

None! The workflows use `GITHUB_TOKEN` which is automatically provided.

### Optional Secrets

**For Codecov (if you want coverage reports):**
1. Sign up at https://codecov.io
2. Add `CODECOV_TOKEN` to GitHub Secrets
3. Uncomment `fail_ci_if_error: true` in `test.yml`

## Permissions

The workflows use minimal permissions:

**docker-release.yml:**
- `contents: read` - Read repository code
- `packages: write` - Push to GHCR
- `id-token: write` - Create attestations

**docker-pr.yml:**
- `contents: read` - Read repository code
- `pull-requests: write` - Comment on PRs

**test.yml:**
- Default permissions (read-only)

## Caching

All workflows use GitHub Actions cache:
- **Go modules:** Cached by `setup-go`
- **npm packages:** Cached by `setup-node`
- **Docker layers:** Cached with `type=gha`

This makes subsequent builds much faster!

## Multi-Architecture Builds

### Release (Both Platforms)
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64-bit, Raspberry Pi, M1 Macs)

### PR Builds (Fast Validation)
- `linux/amd64` only (faster CI)
- Full multi-arch tested on release

## Image Tags Explained

When you push `v1.2.3`:
- `ghcr.io/OWNER/guardian-log:1.2.3` - Exact version
- `ghcr.io/OWNER/guardian-log:1.2` - Latest patch for 1.2.x
- `ghcr.io/OWNER/guardian-log:1` - Latest minor for 1.x.x
- `ghcr.io/OWNER/guardian-log:sha-abc123` - Specific commit

**Recommended:** Pin to major or minor versions in production
```yaml
# Good - gets patch updates
image: ghcr.io/OWNER/guardian-log:1.2

# Better - pin exact version for reproducibility
image: ghcr.io/OWNER/guardian-log:1.2.3
```

## GitHub Container Registry (GHCR)

### Visibility

By default, packages are **private**. To make public:

1. Go to repository → Packages (right sidebar)
2. Click your package
3. Package settings → Change visibility → Public

### Pulling Images

**Public packages:**
```bash
docker pull ghcr.io/OWNER/guardian-log:1.0.0
```

**Private packages:**
```bash
# Create a Personal Access Token (PAT) with `read:packages` scope
echo $GITHUB_PAT | docker login ghcr.io -u USERNAME --password-stdin
docker pull ghcr.io/OWNER/guardian-log:1.0.0
```

## Common Patterns

### Pre-release Tags

```bash
# Alpha release
git tag v1.0.0-alpha.1
git push origin v1.0.0-alpha.1

# Beta release
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1

# Release candidate
git tag v1.0.0-rc.1
git push origin v1.0.0-rc.1
```

### Hotfix Workflow

```bash
# Create hotfix branch from tag
git checkout -b hotfix/1.0.1 v1.0.0

# Fix bug and commit
git commit -m "Fix critical bug"

# Create new patch tag
git tag v1.0.1
git push origin hotfix/1.0.1
git push origin v1.0.1
```

### Manual Workflow Trigger

Add to any workflow to allow manual runs:

```yaml
on:
  workflow_dispatch:  # Manual trigger
    inputs:
      version:
        description: 'Version to build'
        required: true
```

## Monitoring

### View Workflow Runs
- Repository → Actions tab
- See all runs, filter by workflow
- Click run to see jobs and logs

### Status Badges

Add to README.md:
```markdown
![Docker Release](https://github.com/OWNER/REPO/actions/workflows/docker-release.yml/badge.svg)
![Docker PR](https://github.com/OWNER/REPO/actions/workflows/docker-pr.yml/badge.svg)
![Tests](https://github.com/OWNER/REPO/actions/workflows/test.yml/badge.svg)
```

## Troubleshooting

### Build Times Out
- Multi-arch builds can take 10+ minutes
- Increase timeout: `timeout-minutes: 30` in job

### Permission Denied
- Check workflow permissions
- Ensure `GITHUB_TOKEN` has correct scopes

### Cache Not Working
- Caches are scoped to branch
- New branches don't inherit cache from main
- First build will be slower

### Multi-Arch Build Fails
- Usually emulation issue
- Check QEMU setup step
- Try building single platform first

## Best Practices

1. **Always test locally first:**
   ```bash
   make test
   make docker-build
   ```

2. **Use semantic versioning:**
   - `v1.0.0` - Major release
   - `v1.1.0` - Minor update
   - `v1.0.1` - Patch/bugfix

3. **Write good commit messages:**
   - Clear, descriptive
   - Reference issues: `Fixes #123`

4. **Keep PRs focused:**
   - One feature per PR
   - Easier to review and test

5. **Monitor builds:**
   - Check Actions tab
   - Fix failures quickly

## Examples

### Example: Release v1.0.0

```bash
# 1. Prepare release
make test
make lint
make docker-build

# 2. Update version in docs
vim README.md  # Update version references

# 3. Commit changes
git add .
git commit -m "Prepare v1.0.0 release"
git push origin main

# 4. Create tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# 5. Wait for GitHub Actions
# Monitor at: github.com/OWNER/REPO/actions

# 6. Verify image
docker pull ghcr.io/OWNER/guardian-log:1.0.0
docker run -p 8080:8080 \
  --env-file .env \
  ghcr.io/OWNER/guardian-log:1.0.0
```

### Example: Test PR Build

```bash
# 1. Create feature branch
git checkout -b feature/add-metrics

# 2. Make changes
vim internal/metrics/metrics.go

# 3. Test locally
make test
make docker-build

# 4. Commit and push
git add .
git commit -m "Add Prometheus metrics"
git push origin feature/add-metrics

# 5. Create PR on GitHub

# 6. Watch workflows run automatically
# - Tests run
# - Docker builds
# - Bot comments on PR

# 7. Address any failures
# Make fixes → commit → push
# Workflows run again automatically
```

## Summary

- ✅ **Release:** Push tag → Auto build → Push to GHCR (multi-arch)
- ✅ **PR:** Push commit → Auto build → Validate (amd64)
- ✅ **Tests:** Automatic on PR and push
- ✅ **Caching:** Fast subsequent builds
- ✅ **No secrets needed:** Uses GITHUB_TOKEN

**Quick Start:**
```bash
git tag v1.0.0
git push origin v1.0.0
# Done! Image at ghcr.io/OWNER/guardian-log:1.0.0
```
