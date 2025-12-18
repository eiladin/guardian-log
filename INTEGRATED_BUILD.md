# Integrated Build - Single Binary Deployment ✅

## Overview
Guardian-Log now supports **integrated deployment** where the React frontend is embedded directly into the Go binary. This means you only need to run a single executable that serves both the API and the web dashboard!

## Benefits
- ✅ **Single Binary**: One file to deploy, no separate frontend server needed
- ✅ **Simplified Deployment**: Copy one binary instead of managing multiple services
- ✅ **Docker-Ready**: Perfect for containerization (coming in Milestone 4)
- ✅ **No CORS Issues**: Frontend and API served from same origin
- ✅ **Production Optimized**: Frontend is pre-built and optimized
- ✅ **Offline Capable**: All frontend assets embedded in binary

## How It Works

### Build Process
1. **Frontend Build**: Vite compiles React app to optimized static files in `web/dist/`
2. **Copy to webfs**: Build script copies `web/dist/` to `webfs/web/dist/`
3. **Go Embed**: Go's `embed` package includes `webfs/web/dist/` in the binary
4. **Single Binary**: Result is one executable with everything embedded

### Architecture
```
guardian-log binary
  ├─ Go Backend (Port 8080)
  │   ├─ API Routes (/api/*)
  │   └─ Static File Server (/* → embedded frontend)
  └─ Embedded Frontend
      ├─ index.html
      ├─ assets/*.js
      └─ assets/*.css
```

### Code Structure
- **`webfs/webfs.go`**: Embeds `web/dist` directory using `go:embed`
- **`internal/api/server.go`**: Serves embedded files with SPA fallback
- **`web/src/api.ts`**: Uses relative URLs (`/api/*`) for API calls
- **`web/vite.config.ts`**: Configured proxy for development

## Building

### Quick Build (Recommended)
```bash
./build.sh
```

This script:
1. Builds frontend (`npm run build` in `web/`)
2. Copies `web/dist` to `webfs/web/dist`
3. Builds Go binary with embedded frontend
4. Outputs `./bin/guardian-log`

### Manual Build
If you prefer to build manually:

```bash
# 1. Build frontend
cd web
npm install
npm run build
cd ..

# 2. Copy dist to webfs
rm -rf webfs/web
mkdir -p webfs/web
cp -r web/dist webfs/web/

# 3. Build Go binary
go build -o ./bin/guardian-log ./cmd/guardian-log
```

## Running

### Production Mode (Integrated)
```bash
./bin/guardian-log
```

Output:
```
Guardian-Log starting...
🌐 Server starting on :8080
   - API: http://localhost:8080/api
   - Dashboard: http://localhost:8080
📱 Serving frontend from embedded filesystem
```

Open your browser to: **http://localhost:8080**

The dashboard is served directly by the Go server!

### Development Mode (Separate)
For frontend development, use the dev server:

```bash
# Terminal 1: Run backend
go run ./cmd/guardian-log

# Terminal 2: Run frontend with hot reload
cd web
npm run dev
```

Open http://localhost:5173 - Vite proxy forwards `/api/*` to backend.

## Technical Details

### Frontend API Configuration
The frontend uses relative URLs:
```typescript
// web/src/api.ts
const API_BASE_URL = '/api';
```

This works in both modes:
- **Dev**: Vite proxy forwards `/api` → `http://localhost:8080/api`
- **Prod**: Same origin, no proxy needed

### SPA Routing
The server implements SPA fallback routing:
```go
// If file not found in embedded FS, serve index.html
// This allows client-side routing to work
r.URL.Path = "/"
spaHandler.ServeHTTP(w, r)
```

This means routes like `/anomalies`, `/settings` work even though they don't exist as files.

### Embedding Process
```go
//go:embed all:web/dist
var distFS embed.FS

func GetFS() (fs.FS, error) {
    return fs.Sub(distFS, "web/dist")
}
```

The `all:` prefix includes all files, even dotfiles and subdirectories.

## File Structure

### Before Build
```
guardian-log/
├── web/
│   ├── src/           # React source
│   └── dist/          # Built frontend (after npm run build)
├── webfs/
│   └── webfs.go       # Embed directive
└── internal/api/
    └── server.go      # Serves embedded files
```

### After Build
```
guardian-log/
├── webfs/
│   ├── web/
│   │   └── dist/      # Copied frontend (for embedding)
│   └── webfs.go
└── bin/
    └── guardian-log   # Single binary with embedded frontend
```

## Binary Size
- **Without frontend**: ~10 MB
- **With frontend embedded**: ~12 MB
- **Frontend assets**: ~200 KB (gzipped: ~63 KB)

The embedded frontend adds minimal size to the binary!

## Deployment

### Local Deployment
```bash
# Build
./build.sh

# Copy binary to server
scp ./bin/guardian-log user@server:/opt/guardian-log/

# Run on server
ssh user@server
/opt/guardian-log/guardian-log
```

### Systemd Service
```ini
[Unit]
Description=Guardian-Log DNS Anomaly Detection
After=network.target

[Service]
Type=simple
User=guardian
WorkingDirectory=/opt/guardian-log
ExecStart=/opt/guardian-log/guardian-log
Restart=always

[Install]
WantedBy=multi-user.target
```

### Docker (Milestone 4)
Coming soon! The integrated build makes Docker deployment trivial:
- Single `COPY` command to add binary
- Single `CMD` to run
- No nginx or separate frontend container needed

## Troubleshooting

### Frontend Not Loading
If you see `⚠️ Warning: Frontend not embedded`, the dist folder wasn't included during build.

**Solution**: Run `./build.sh` instead of `go build` directly.

### API Not Found (404)
If API calls return 404, check that:
1. Backend is running on port 8080
2. Frontend is using `/api` (not `http://localhost:8080/api`)

### Stale Frontend
If you see old frontend after rebuild:

**Solution**: The old binary is cached. Rebuild:
```bash
rm ./bin/guardian-log
./build.sh
```

## Development Workflow

### Frontend Changes
```bash
# Option 1: Use dev server (hot reload)
cd web && npm run dev

# Option 2: Rebuild everything
./build.sh
```

### Backend Changes
```bash
# If no frontend changes needed
go build -o ./bin/guardian-log ./cmd/guardian-log

# Or use the build script
./build.sh
```

## Next: Milestone 4 (Docker)

The integrated build makes Milestone 4 straightforward:

**Dockerfile**:
```dockerfile
FROM golang:1.21 AS builder
# Copy source
# Run ./build.sh
# Result: Single binary

FROM alpine:latest
# Copy binary
# Run it
```

That's it! No multi-stage builds for frontend/backend, no nginx, no complexity.

## Summary

✅ **Single binary deployment achieved!**
✅ **Frontend embedded in Go backend**
✅ **SPA routing works correctly**
✅ **Dev mode still available for development**
✅ **Ready for Docker containerization**

The integrated build significantly simplifies deployment and makes Guardian-Log much more portable and easy to run anywhere!

---

**Status**: Complete ✅
**Date**: December 18, 2024
**Binary**: `./bin/guardian-log` (API + Dashboard)
**Port**: 8080 (serves both API and frontend)
