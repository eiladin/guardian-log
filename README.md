# Guardian-Log

Self-hosted AI-powered DNS anomaly detection for AdGuard Home.

## What It Does

Guardian-Log monitors your AdGuard Home DNS queries in real-time, using AI to detect suspicious domains and potential security threats. When an anomaly is detected, you can approve (whitelist) or block it directly from the web dashboard.

**Key Features:**
- 🔍 **Real-time monitoring** - Polls AdGuard Home for new DNS queries
- 🤖 **AI-powered analysis** - Uses Google Gemini to detect anomalies
- 🎯 **Smart baseline learning** - Learns normal DNS patterns per device
- 📊 **Web dashboard** - Modern React UI for managing alerts
- 🐳 **Easy deployment** - Docker Compose for one-command setup
- 🏗️ **Production ready** - Multi-arch images, CI/CD, security best practices

## Quick Start

### Docker Compose (Recommended)

```bash
# 1. Clone and configure
git clone https://github.com/OWNER/guardian-log.git
cd guardian-log
cp .env.example .env
vim .env  # Add AdGuard URL and Gemini API key

# 2. Start Guardian-Log
docker-compose up -d

# 3. Open dashboard
open http://localhost:8080
```

### From Binary

```bash
# Build
make build

# Configure
cp .env.example .env
vim .env

# Run
./bin/guardian-log
```

## Screenshots

*Dashboard showing real-time anomaly detection with AI analysis*

## Configuration

Required environment variables:

```env
# AdGuard Home
AGH_URL=http://192.168.1.1:8080
AGH_USER=admin
AGH_PASS=your_password

# AI Analysis (get free key at https://aistudio.google.com/app/apikey)
GEMINI_API_KEY=your_gemini_key
```

See [Configuration Guide](docs/deployment/CONFIGURATION.md) for all options.

## Documentation

### 📚 Getting Started
- **[Installation](docs/deployment/INSTALL.md)** - Detailed setup instructions
- **[Configuration](docs/deployment/CONFIGURATION.md)** - Environment variables and options
- **[Docker Deployment](docs/deployment/DOCKER.md)** - Container and compose guide
- **[Troubleshooting](docs/deployment/TROUBLESHOOTING.md)** - Common issues

### 💻 Development
- **[Development Guide](docs/development/GUIDE.md)** - Local development setup with hot reload
- **[VS Code Setup](docs/development/VSCODE.md)** - IDE configuration and debugging

### 🚀 Operations
- **[CI/CD](docs/ci-cd/GITHUB_ACTIONS.md)** - Automated builds and releases
- **[API Reference](docs/API.md)** - REST API documentation

### 🏛️ Architecture
- **[System Architecture](docs/ARCHITECTURE.md)** - Design and components
- **[Milestones](docs/milestones/)** - Development progress

## Features

### DNS Monitoring
- Polls AdGuard Home every 10 seconds (configurable)
- Builds baseline of normal DNS patterns per client
- Detects deviations from established patterns
- WHOIS enrichment for domain context

### AI Analysis
- Google Gemini integration (free tier available)
- Analyzes domain patterns, TLDs, registration info
- Risk scoring (1-10) with explanations
- Actionable recommendations

### Web Dashboard
- Real-time anomaly feed
- Approve (whitelist) or Block workflows
- Statistics and insights
- Filter by status
- Auto-refresh every 10 seconds

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────┐
│ AdGuard     │────▶│ Guardian-Log │────▶│ Gemini   │
│ Home        │     │              │     │ AI       │
└─────────────┘     └──────┬───────┘     └──────────┘
                           │
                    ┌──────▼───────┐
                    │ Web Dashboard│
                    └──────────────┘
```

**Components:**
- **Ingestor** - Polls AdGuard Home API
- **Analyzer** - AI-powered anomaly detection
- **API Server** - REST API + embedded web UI
- **Storage** - BoltDB for persistence

## Technology Stack

**Backend:** Go 1.25+ • BoltDB • Google Gemini API • AdGuard Home API

**Frontend:** React 18 • TypeScript • Vite • Tailwind CSS

**Deployment:** Docker • Docker Compose • Multi-arch (amd64, arm64) • GitHub Actions

## Commands

```bash
# Development
make dev-backend          # Start backend (hot reload)
make dev-frontend         # Start frontend (hot reload)

# Building
make build                # Build single binary
make docker-build         # Build Docker image
make docker-build-multi   # Multi-arch build

# Testing
make test                 # Run tests
make lint                 # Run linters
```

See [Development Guide](docs/development/GUIDE.md) for details.

## Roadmap

- [x] AdGuard Home integration
- [x] AI-powered analysis
- [x] Web dashboard
- [x] Docker deployment
- [x] Multi-arch builds
- [x] CI/CD pipeline
- [ ] Prometheus metrics
- [ ] Email/webhook notifications
- [ ] Multi-user support
- [ ] Historical trends

See [SPECIFICATION.md](SPECIFICATION.md) for the complete roadmap.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Development setup:**
```bash
make dev-backend   # Terminal 1
make dev-frontend  # Terminal 2
# Access at http://localhost:5173
```

## License

[License Type] - See [LICENSE](LICENSE)

## Support

- **Issues:** [GitHub Issues](https://github.com/OWNER/guardian-log/issues)
- **Discussions:** [GitHub Discussions](https://github.com/OWNER/guardian-log/discussions)

## Acknowledgments

- AdGuard Home for excellent DNS filtering
- Google Gemini for AI capabilities
- The Go and React communities

---

**Made with ❤️ for privacy-focused security monitoring**
