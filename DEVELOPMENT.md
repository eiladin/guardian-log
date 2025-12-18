# Development Guide

This guide covers live-reload debugging and development workflows for Guardian-Log.

## Table of Contents
- [Quick Start](#quick-start)
- [Development Modes](#development-modes)
- [Hot Reload Setup](#hot-reload-setup)
- [Debugging](#debugging)
- [IDE Setup](#ide-setup)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Option 1: Separate Terminals (Recommended for Beginners)

**Terminal 1 - Backend with Hot Reload:**
```bash
make dev-backend
# Backend runs on http://localhost:8080
# Auto-reloads on any .go file changes
```

**Terminal 2 - Frontend with Hot Reload:**
```bash
make dev-frontend
# Frontend dev server on http://localhost:5173
# Hot Module Replacement (HMR) enabled
```

**Access:** Open http://localhost:5173 in your browser
- Frontend changes = instant hot reload (no page refresh)
- Backend changes = auto-restart (page refresh needed)
- API calls from frontend automatically proxy to backend

### Option 2: Single Terminal with Tmux

```bash
make dev-both
# Starts both servers in split tmux windows
# Ctrl+B then D to detach
# tmux attach -t guardian-dev to reattach
```

## Development Modes

### 🔥 Hot Reload Development (Separate Servers)

**Architecture:**
```
Browser
  └─→ http://localhost:5173 (Vite Dev Server)
        └─→ /api/* → http://localhost:8080 (Go Backend)
```

**How it works:**
1. **Frontend (Vite)**: Runs on port 5173 with Hot Module Replacement
   - Changes to React components = instant update (no refresh)
   - Changes to CSS/styles = instant update
   - TypeScript compilation happens in real-time

2. **Backend (Air)**: Runs on port 8080 with auto-restart
   - Changes to .go files trigger rebuild
   - Binary restarts automatically (takes ~2-3 seconds)
   - API endpoints available immediately after restart

3. **Proxy**: Vite forwards `/api/*` requests to backend
   - Configured in `web/vite.config.ts`
   - No CORS issues
   - Same-origin cookies work

**When to use:**
- Active frontend/UI development
- Rapid iteration on React components
- Debugging API interactions
- Best developer experience

### 🏗️ Production Mode (Single Binary)

```bash
make build
./bin/guardian-log
# Access: http://localhost:8080
```

**How it works:**
- Frontend built into `web/dist`
- Static files embedded in Go binary
- Single server on port 8080
- No hot reload (requires rebuild)

**When to use:**
- Testing production build
- Verifying embedded frontend works
- Performance testing
- Before creating Docker image

## Hot Reload Setup

### Backend Hot Reload (Air)

**Installation:**
```bash
# If you don't have air installed:
go install github.com/air-verse/air@latest
```

**Configuration:**
File: `.air.toml` (already configured)

**Features:**
- ✅ Watches all `.go` files
- ✅ Excludes test files
- ✅ 1 second delay to batch changes
- ✅ Build errors shown in `build-errors.log`
- ✅ Colored output

**Customize watching:**
Edit `.air.toml` to change:
- `exclude_dir`: Directories to ignore
- `include_ext`: File extensions to watch
- `delay`: Time to wait before rebuilding

### Frontend Hot Reload (Vite)

**Built-in features:**
- ✅ Hot Module Replacement (HMR)
- ✅ Fast refresh for React
- ✅ CSS hot reload
- ✅ TypeScript type checking
- ✅ Source maps for debugging

**Configuration:**
File: `web/vite.config.ts`

**Customize:**
```typescript
export default defineConfig({
  server: {
    port: 5173,        // Change dev server port
    host: true,        // Expose on network
    open: true,        // Auto-open browser
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

## Debugging

### Backend Debugging (Go)

#### Using Delve (Go Debugger)

**Installation:**
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

**Debug mode:**
```bash
# Start backend with debugger
dlv debug ./cmd/guardian-log --headless --listen=:2345 --api-version=2

# In another terminal, connect:
dlv connect :2345
```

**VS Code Launch Configuration:**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Guardian-Log Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/guardian-log",
      "env": {
        "AGH_URL": "http://192.168.1.2:8080",
        "GEMINI_API_KEY": "your-key-here"
      },
      "args": []
    }
  ]
}
```

#### Logging

**Enable verbose logging:**
```bash
# Set LOG_LEVEL in .env
LOG_LEVEL=debug

# Or export before running:
export LOG_LEVEL=debug
make dev-backend
```

**Log levels:**
- `debug`: Verbose, includes all details
- `info`: Normal operations (default)
- `warn`: Warnings and errors
- `error`: Errors only

### Frontend Debugging (React)

#### Browser DevTools

**Chrome/Edge DevTools:**
1. Open http://localhost:5173
2. Press F12
3. Source maps enabled automatically
4. Set breakpoints in TypeScript files

**React DevTools Extension:**
- Install: [Chrome](https://chrome.google.com/webstore/detail/react-developer-tools/fmkadmapgofadopljbjfkapdkoienihi) | [Firefox](https://addons.mozilla.org/en-US/firefox/addon/react-devtools/)
- Component tree inspection
- Props/State inspection
- Performance profiling

#### VS Code Debugging

**Install:** Debugger for Chrome extension

**Launch configuration:**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Frontend in Chrome",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/web",
      "sourceMapPathOverrides": {
        "webpack:///src/*": "${webRoot}/src/*"
      }
    }
  ]
}
```

#### Console Logging

**Temporary debug logs:**
```typescript
// In React components
console.log('Anomalies:', anomalies);
console.table(stats);

// Network requests
console.log('API Response:', await GuardianAPI.getAnomalies());
```

**Production logging:**
```typescript
// Use proper error boundaries
if (process.env.NODE_ENV === 'development') {
  console.log('Debug info:', data);
}
```

### API Debugging

#### Network Tab

**Browser DevTools → Network:**
- Filter: `api`
- See all API requests/responses
- Timing information
- Request/response headers

#### cURL Testing

**Test API endpoints directly:**
```bash
# Get anomalies
curl http://localhost:8080/api/anomalies

# Get stats
curl http://localhost:8080/api/stats

# Approve anomaly
curl -X POST http://localhost:8080/api/anomalies/abc-123/approve

# Block anomaly
curl -X POST http://localhost:8080/api/anomalies/abc-123/block

# Health check
curl http://localhost:8080/api/health
```

#### Postman/Thunder Client

Import these endpoints:
- `GET http://localhost:8080/api/anomalies`
- `GET http://localhost:8080/api/anomalies?status=pending`
- `GET http://localhost:8080/api/stats`
- `POST http://localhost:8080/api/anomalies/{id}/approve`
- `POST http://localhost:8080/api/anomalies/{id}/block`

## IDE Setup

### VS Code (Recommended)

**Extensions:**
```bash
# Go development
code --install-extension golang.go

# React/TypeScript
code --install-extension dbaeumer.vscode-eslint
code --install-extension esbenp.prettier-vscode

# Debugging
code --install-extension ms-vscode.js-debug
```

**Workspace Settings:**
`.vscode/settings.json`:
```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.lintOnSave": "workspace",
  "go.formatTool": "gofmt",
  "editor.formatOnSave": true,
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

**Complete Launch Configuration:**
`.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/guardian-log",
      "envFile": "${workspaceFolder}/.env"
    },
    {
      "name": "Debug Frontend",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/web"
    }
  ],
  "compounds": [
    {
      "name": "Debug Full Stack",
      "configurations": ["Debug Backend", "Debug Frontend"],
      "stopAll": true
    }
  ]
}
```

### GoLand / WebStorm

**Run Configuration - Backend:**
- **Type:** Go Build
- **Package:** `./cmd/guardian-log`
- **Working directory:** Project root
- **Environment:** Load from `.env`

**Run Configuration - Frontend:**
- **Type:** npm
- **Command:** `run dev`
- **Working directory:** `web/`

## Troubleshooting

### Frontend doesn't connect to backend

**Symptoms:** API calls fail with CORS errors or connection refused

**Solutions:**
```bash
# 1. Check backend is running
curl http://localhost:8080/api/health

# 2. Check Vite proxy config
cat web/vite.config.ts | grep -A 5 "proxy"

# 3. Restart both servers
# Kill all processes
pkill -f "air"
pkill -f "vite"

# Restart
make dev-backend  # Terminal 1
make dev-frontend # Terminal 2
```

### Hot reload not working

**Backend:**
```bash
# Check air is running
ps aux | grep air

# Check .air.toml exists
ls -la .air.toml

# Check build errors
tail -f build-errors.log

# Force rebuild
touch internal/api/server.go
```

**Frontend:**
```bash
# Check Vite is running
ps aux | grep vite

# Clear Vite cache
rm -rf web/node_modules/.vite
cd web && npm run dev
```

### Port already in use

**Backend (8080):**
```bash
# Find process using port 8080
lsof -ti:8080

# Kill it
kill $(lsof -ti:8080)

# Or with sudo
sudo kill $(sudo lsof -ti:8080)
```

**Frontend (5173):**
```bash
# Find process using port 5173
lsof -ti:5173

# Kill it
kill $(lsof -ti:5173)
```

### Environment variables not loading

**Check .env file:**
```bash
# Ensure .env exists
ls -la .env

# Check contents
cat .env

# Load manually for testing
source .env
make dev-backend
```

### Build errors

**Backend:**
```bash
# Check Go version
go version  # Should be >= 1.25

# Update dependencies
go mod tidy

# Clean and rebuild
make clean
go mod download
go build ./...
```

**Frontend:**
```bash
# Check Node version
node --version  # Should be >= 18

# Clean and reinstall
cd web
rm -rf node_modules package-lock.json
npm install

# Check for errors
npm run build
```

## Development Workflow Tips

### Making Changes

**Frontend changes:**
1. Edit files in `web/src/`
2. Save → instant hot reload in browser
3. No restart needed

**Backend changes:**
1. Edit files in `internal/`, `cmd/`, etc.
2. Save → air rebuilds automatically
3. Wait 2-3 seconds for restart
4. Refresh browser page to see changes

**API changes:**
1. Update backend handler in `internal/api/handlers.go`
2. Update frontend API client in `web/src/api.ts`
3. Update TypeScript types in `web/src/types.ts`
4. Test with browser DevTools → Network tab

### Testing Changes

**Quick iteration:**
```bash
# 1. Start dev servers
make dev-both  # or separate terminals

# 2. Make changes
# Edit web/src/App.tsx

# 3. See instant results in browser
# No manual refresh needed

# 4. Check backend logs
# Watch Terminal 1 for backend output

# 5. Test API calls
# Browser DevTools → Network → api
```

**Before committing:**
```bash
# 1. Run tests
make test

# 2. Run linters
make lint

# 3. Format code
make fmt

# 4. Test production build
make build
./bin/guardian-log

# 5. Test Docker build
make docker-build
```

### Productivity Tips

1. **Use two monitors:**
   - Monitor 1: Code editor
   - Monitor 2: Browser with DevTools open

2. **Keep terminals visible:**
   - Use tmux or terminal tabs
   - Watch for build errors immediately

3. **Enable auto-save:**
   - VS Code: `"files.autoSave": "afterDelay"`
   - Triggers hot reload automatically

4. **Use React DevTools:**
   - Inspect component state
   - Profile performance
   - Debug re-renders

5. **Bookmark dev URLs:**
   - Frontend: http://localhost:5173
   - Backend: http://localhost:8080/api/health
   - Quick access during development

## Advanced Topics

### Custom Environment Variables

**Add to `.env`:**
```bash
MY_CUSTOM_VAR=value
```

**Use in backend:**
```go
customVar := os.Getenv("MY_CUSTOM_VAR")
```

**Use in frontend:**
```typescript
// Must be prefixed with VITE_
// In .env:
VITE_MY_VAR=value

// In code:
const myVar = import.meta.env.VITE_MY_VAR
```

### Debugging Production Issues

**Enable source maps in production:**
```bash
# Build with source maps
go build -o ./bin/guardian-log ./cmd/guardian-log

# Frontend already includes source maps in dist/
cd web && npm run build
```

**Test production build locally:**
```bash
make build
./bin/guardian-log
# Access: http://localhost:8080 (no port 5173!)
```

### Performance Profiling

**Backend (pprof):**
```go
// Add to cmd/guardian-log/main.go
import _ "net/http/pprof"

// Add endpoint:
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Access profiles:
// http://localhost:6060/debug/pprof/
```

**Frontend (React DevTools Profiler):**
1. Wrap component in `<Profiler>`
2. Record session
3. Analyze render times

## Summary

**Quick reference:**
```bash
# Start development
make dev-backend  # Terminal 1
make dev-frontend # Terminal 2

# Or
make dev-both     # Tmux split screen

# Access
http://localhost:5173  # Frontend (use this!)
http://localhost:8080  # Backend API

# Test production
make build
./bin/guardian-log

# Build Docker
make docker-build
```

Happy coding! 🚀
