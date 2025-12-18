# Guardian-Log Documentation

Complete documentation for Guardian-Log - AI-powered DNS anomaly detection for AdGuard Home.

## 📖 Table of Contents

### Getting Started
- **[Installation](deployment/INSTALL.md)** - Complete installation guide
- **[Configuration](deployment/CONFIGURATION.md)** - Environment variables and settings
- **[Quick Start](../README.md#quick-start)** - Get running in 5 minutes

### Deployment
- **[Docker Deployment](deployment/DOCKER.md)** - Container and Docker Compose guide
- **[Troubleshooting](deployment/TROUBLESHOOTING.md)** - Common issues and solutions

### Development
- **[Development Guide](development/GUIDE.md)** - Local development with hot reload
- **[VS Code Setup](development/VSCODE.md)** - IDE configuration and debugging
- **[Debug Guide](development/DEBUG_GUIDE.md)** - Debugging workflows
- **[Quick Debug](development/QUICK_DEBUG.md)** - Quick debugging reference

### CI/CD & Operations
- **[GitHub Actions](ci-cd/GITHUB_ACTIONS.md)** - Automated builds and releases
- **[Workflows](../.github/workflows/README.md)** - Workflow documentation

### Architecture & Design
- **[System Architecture](ARCHITECTURE.md)** - Overall system design
- **[API Reference](API.md)** - REST API documentation
- **[Design Docs](design/)** - Implementation plans and decisions

### Project History
- **[Milestone 3: Web Dashboard](milestones/M3_COMPLETE.md)** - React UI implementation
- **[Milestone 4: Dockerization](milestones/M4_COMPLETE.md)** - Container deployment

## Quick Links

### Common Tasks

**Deploy with Docker:**
```bash
docker-compose up -d
```

**Local Development:**
```bash
make dev-backend   # Terminal 1
make dev-frontend  # Terminal 2
```

**Build Docker Image:**
```bash
make docker-build
```

**Run Tests:**
```bash
make test
```

### Configuration Files

- `.env` - Environment variables (copy from `.env.example`)
- `docker-compose.yml` - Docker Compose configuration
- `.air.toml` - Hot reload configuration (backend)
- `web/vite.config.ts` - Frontend dev server configuration

### Important Directories

- `cmd/` - Application entry points
- `internal/` - Go packages (api, ingestor, storage, llm, etc.)
- `web/` - React frontend application
- `webfs/` - Embedded frontend filesystem
- `data/` - Runtime data (BoltDB, etc.)

## Documentation Structure

```
docs/
├── README.md (this file)
├── deployment/
│   ├── INSTALL.md              # Installation guide
│   ├── CONFIGURATION.md        # Configuration reference
│   ├── DOCKER.md               # Docker deployment
│   └── TROUBLESHOOTING.md      # Common issues
├── development/
│   ├── GUIDE.md                # Development guide
│   ├── VSCODE.md               # VS Code setup
│   ├── DEBUG_GUIDE.md          # Debugging workflows
│   ├── QUICK_DEBUG.md          # Quick debug reference
│   └── CHANGES.md              # Recent changes
├── ci-cd/
│   └── GITHUB_ACTIONS.md       # CI/CD documentation
├── design/
│   ├── PLAN.md                 # Implementation plan
│   ├── RATE_LIMITING.md        # Rate limiting design
│   └── TRUE_BATCH_PROCESSING.md # Batch processing design
├── milestones/
│   ├── M3_COMPLETE.md          # Milestone 3 summary
│   └── M4_COMPLETE.md          # Milestone 4 summary
├── ARCHITECTURE.md             # System architecture
└── API.md                      # API reference
```

## Contributing to Documentation

Documentation contributions are welcome! Please:

1. Keep docs concise and actionable
2. Use code examples where appropriate
3. Update links when moving/renaming files
4. Test all code examples before committing

## Need Help?

- Check [Troubleshooting Guide](deployment/TROUBLESHOOTING.md)
- Search [GitHub Issues](https://github.com/OWNER/guardian-log/issues)
- Start a [Discussion](https://github.com/OWNER/guardian-log/discussions)

---

**[← Back to Main README](../README.md)**
