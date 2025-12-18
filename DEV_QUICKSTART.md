# Development Quick Start 🚀

## Start Development Servers

### Option 1: Two Terminals (Recommended)
```bash
# Terminal 1: Backend (auto-reload on .go changes)
make dev-backend

# Terminal 2: Frontend (instant hot reload)
make dev-frontend

# Access: http://localhost:5173
```

### Option 2: Tmux (Single Terminal)
```bash
make dev-both
# Ctrl+B then D to detach
```

## URLs

| Service | URL | Description |
|---------|-----|-------------|
| **Frontend Dev** | http://localhost:5173 | Use this for development! |
| Backend API | http://localhost:8080 | Backend only |
| Health Check | http://localhost:8080/api/health | Test backend |

## Development Workflow

1. **Edit frontend** (`web/src/*`)
   - Save → Instant hot reload
   - No browser refresh needed

2. **Edit backend** (`internal/*`, `cmd/*`)
   - Save → Auto rebuild (2-3 seconds)
   - Refresh browser to see changes

3. **Edit API**
   - Update `internal/api/handlers.go`
   - Update `web/src/api.ts`
   - Update `web/src/types.ts`

## Debugging

### Frontend (Browser)
```bash
# Open http://localhost:5173
# Press F12 → DevTools
# Console, Network, React DevTools
```

### Backend (Logs)
```bash
# Set LOG_LEVEL=debug in .env
# Watch backend terminal for logs
```

### API Testing
```bash
# Test endpoints
curl http://localhost:8080/api/anomalies
curl http://localhost:8080/api/stats
```

## Common Commands

```bash
make help              # Show all commands
make dev-backend       # Start backend with hot-reload
make dev-frontend      # Start frontend dev server
make dev-both          # Start both in tmux
make test              # Run tests
make lint              # Run linters
make fmt               # Format code
make build             # Build production binary
make docker-build      # Build Docker image
```

## Troubleshooting

### Port in use
```bash
# Kill process on port 8080 (backend)
kill $(lsof -ti:8080)

# Kill process on port 5173 (frontend)
kill $(lsof -ti:5173)
```

### Hot reload not working
```bash
# Restart servers
pkill -f "air"
pkill -f "vite"
make dev-backend  # Terminal 1
make dev-frontend # Terminal 2
```

### Build errors
```bash
# Backend
go mod tidy
make clean
go build ./...

# Frontend
cd web
rm -rf node_modules
npm install
```

## File Structure

```
guardian-log/
├── cmd/guardian-log/     # Main application entry
├── internal/
│   ├── api/              # REST API handlers
│   ├── ingestor/         # AdGuard integration
│   ├── llm/              # LLM analysis
│   └── storage/          # Database layer
├── web/
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── api.ts        # API client
│   │   ├── types.ts      # TypeScript types
│   │   └── App.tsx       # Main app
│   └── vite.config.ts    # Dev server config
├── webfs/                # Embedded frontend
├── .air.toml             # Hot-reload config
├── .env                  # Environment variables
└── Makefile              # Build automation
```

## Before Committing

```bash
make test    # ✓ Tests pass
make lint    # ✓ No lint errors
make fmt     # ✓ Code formatted
make build   # ✓ Production build works
```

## Need More Help?

📖 **Full documentation:** [DEVELOPMENT.md](./DEVELOPMENT.md)

**Quick links:**
- [Hot Reload Setup](./DEVELOPMENT.md#hot-reload-setup)
- [Debugging Guide](./DEVELOPMENT.md#debugging)
- [IDE Setup](./DEVELOPMENT.md#ide-setup)
- [Troubleshooting](./DEVELOPMENT.md#troubleshooting)
