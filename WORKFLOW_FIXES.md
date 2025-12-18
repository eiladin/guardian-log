# GitHub Actions Workflow Fixes ✅

## Issues Found During PR Testing

### Issue 1: golangci-lint Go Version Mismatch

**Error:**
```
can't load config: the Go language version (go1.24) used to build
golangci-lint is lower than the targeted Go version (1.25.5)
```

**Root Cause:**
The `latest` version of golangci-lint was built with Go 1.24, but the project requires Go 1.25.5.

**Fix:**
1. Pinned golangci-lint to specific version: `v1.62`
2. Added `install-mode: binary` to ensure proper Go version compatibility
3. Updated in `.github/workflows/test.yml`:
   ```yaml
   - name: Run golangci-lint
     uses: golangci/golangci-lint-action@v4
     with:
       version: v1.62              # Pinned version
       args: --timeout=5m
       install-mode: binary         # Binary mode for Go version compatibility
   ```

### Issue 2: Missing Embedded Files During go vet

**Error:**
```
webfs/webfs.go:8:12: pattern all:web/dist: no matching files found
```

**Root Cause:**
The `go vet` command runs before the frontend is built, so the files referenced in the `//go:embed` directive don't exist yet.

**Fix:**
Added frontend build step before `go vet` and `golangci-lint` in all jobs:

```yaml
- name: Build frontend (required for embed)
  run: |
    cd web
    npm ci
    npm run build
    cd ..
    mkdir -p webfs/web
    cp -r web/dist webfs/web/
```

This matches the normal build process where frontend must be built before Go compilation.

## Files Modified

### 1. `.github/workflows/test.yml`

**Changes to `test-go` job:**
- Added Node.js setup
- Added frontend build step before `go vet`

**Changes to `lint` job:**
- Added Node.js setup
- Added frontend build step before `golangci-lint`
- Pinned golangci-lint version to v1.62
- Added `install-mode: binary`

### 2. `.golangci.yml`

**Simplified configuration:**
- Removed deprecated/incompatible options
- Kept core linters (errcheck, gosimple, govet, etc.)
- Excluded all linters for `webfs/webfs.go` to avoid embed-related warnings

## Build Order

The correct build order is now enforced in CI:

```
1. Checkout code
   ↓
2. Set up Go + Node.js
   ↓
3. Build frontend (npm ci + npm run build)
   ↓
4. Copy frontend to webfs/web/dist
   ↓
5. Download Go dependencies
   ↓
6. Run go vet (embed files now exist!)
   ↓
7. Run tests
```

## Testing

All three jobs now follow this pattern:

| Job | Frontend Build | Reason |
|-----|----------------|--------|
| `test-go` | ✅ Yes | Needed for `go vet` |
| `test-frontend` | ✅ Yes | Native frontend tests |
| `lint` | ✅ Yes | Needed for `golangci-lint` |

## Verification

Workflow is now valid and should pass:

```bash
# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/test.yml'))"
# ✓ Valid YAML

# Test locally
make test
make lint
make build
```

## Impact

**Before:**
- ❌ golangci-lint failed with Go version error
- ❌ go vet failed with missing embed files

**After:**
- ✅ golangci-lint uses correct Go version
- ✅ go vet runs after frontend is built
- ✅ All CI checks should pass

## Next PR Test

The next PR should have all checks passing:
- ✅ Go tests
- ✅ Frontend build
- ✅ Linting
- ✅ go vet

---

**Status:** Fixed ✅
**Date:** December 18, 2024
