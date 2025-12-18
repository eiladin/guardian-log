# Guardian-Log

A self-hosted LLM-powered DNS anomaly detection and explanation system that integrates with AdGuard Home.

## Overview

Guardian-Log monitors DNS queries from AdGuard Home, identifies first-seen (anomalous) domains for each device, and provides intelligent explanations using various LLM providers. It helps you understand what your devices are communicating with and why.

## Current Status: Milestone 3 ✅

**Milestone 3: Web Dashboard** - Complete!

Guardian-Log now includes a full-featured web dashboard:
- ✅ REST API server (Go) with CORS support
- ✅ React + TypeScript frontend with Vite
- ✅ Real-time anomaly feed with 10-second polling
- ✅ Approve/Block workflows via UI
- ✅ Comprehensive statistics dashboard
- ✅ Professional dark theme UI
- ✅ Mobile-responsive design
- ✅ Connection status monitoring

**Previous Milestones:**
- ✅ Milestone 1: Core Ingestor (AdGuard integration, baseline learning)
- ✅ Milestone 2: LLM Service (WHOIS, Gemini, async analysis)
- ✅ Milestone 3: Web Dashboard (React UI, API server)

## Features

### Core (Milestone 1)
- **AdGuard Home Integration**: Polls the AdGuard Home query log API
- **Baseline Learning**: Automatically builds a baseline of known domains per device
- **Anomaly Detection**: Detects and logs first-seen domains for each client
- **Persistent Storage**: Uses BoltDB for fast, embedded storage
- **Deduplication**: Prevents duplicate processing of the same query
- **Graceful Shutdown**: Handles SIGTERM/SIGINT signals properly

### LLM Analysis (Milestone 2)
- **WHOIS Enrichment**: Automatic WHOIS lookups with 24-hour caching and rate limiting
- **Multi-Provider Support**: Gemini (free tier!), OpenAI, Anthropic, Ollama
- **Intelligent Analysis**: LLM analyzes domain reputation, registration patterns, infrastructure
- **Risk Scoring**: 1-10 risk score with detailed explanations
- **Action Recommendations**: Allow, Investigate, or Block suggestions
- **Asynchronous Processing**: Non-blocking LLM analysis doesn't slow down polling
- **Anomaly Tracking**: Stores suspicious/malicious detections for review

### Web Dashboard (Milestone 3)
- **REST API Server**: Full HTTP API for anomaly management and statistics
- **React Frontend**: Modern TypeScript-based UI with Vite
- **Real-time Updates**: Auto-refresh anomaly feed every 10 seconds
- **Anomaly Cards**: Beautiful cards showing domain, client, risk, and LLM explanations
- **Approve Workflow**: One-click approval adds domain to client baseline
- **Block Workflow**: One-click blocking adds domain to AdGuard blocklist
- **Filter System**: View All, Pending, Approved, or Blocked anomalies
- **Statistics Dashboard**: 10 key metrics with color-coded indicators
- **Professional UI**: Dark theme optimized for security monitoring
- **Mobile Responsive**: Works seamlessly on all device sizes
- **Connection Monitoring**: Online/offline status with automatic retry

## Prerequisites

- Go 1.21 or higher
- Node.js 20+ (for web dashboard)
- AdGuard Home instance (running and accessible)
- AdGuard Home admin credentials
- Google Gemini API key (free at https://aistudio.google.com/app/apikey)

## Quick Start

Want to see the dashboard in action? Follow these steps:

### 1. Configure & Build
```bash
# Clone and configure
cp .env.example .env
# Edit .env with your AdGuard Home URL, credentials, and Gemini API key

# Build everything (frontend + backend with embedded UI)
./build.sh
```

### 2. Start Guardian-Log
```bash
./bin/guardian-log
```

Expected output:
```
🌐 Server starting on :8080
   - API: http://localhost:8080/api
   - Dashboard: http://localhost:8080
📱 Serving frontend from embedded filesystem
```

### 3. Open Dashboard
Open your browser to: **http://localhost:8080**

You'll see:
- 📊 Real-time statistics dashboard
- 🚨 Anomaly feed with LLM explanations
- ✓ Approve button (adds to baseline)
- 🚫 Block button (blocks in AdGuard)
- 🔄 Auto-refresh every 10 seconds

**That's it!** Single binary serves both API and dashboard.

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/eiladin/guardian-log.git
cd guardian-log
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment

Copy the example environment file and edit it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# AdGuard Home Configuration
AGH_URL=http://192.168.1.2:8080
AGH_USER=admin
AGH_PASS=your_password

# Application Configuration
POLL_INTERVAL=10s
DB_PATH=./data/guardian.db
LOG_LEVEL=info

# LLM Configuration (Milestone 2)
LLM_ENABLE=true
LLM_PROVIDER=gemini
LLM_TIMEOUT=30s

# Gemini Configuration (free tier - get key at https://aistudio.google.com/app/apikey)
GEMINI_API_KEY=your-gemini-api-key-here
GEMINI_MODEL=gemini-1.5-flash
```

**Note**: To disable LLM analysis and only use baseline detection, set `LLM_ENABLE=false`.

### 4. Build the application

```bash
go build -o guardian-log ./cmd/guardian-log
```

## Usage

### Running Guardian-Log

```bash
./guardian-log
```

The application will:
1. Connect to your AdGuard Home instance
2. Start polling the query log every 10 seconds (or your configured interval)
3. Log first-seen domains to the console in real-time

### Example Output

```
2025/12/17 22:00:00 Guardian-Log starting...
2025/12/17 22:00:00 AdGuard Home URL: http://192.168.1.2:8080
2025/12/17 22:00:00 Poll Interval: 10s
2025/12/17 22:00:00 Database Path: ./data/guardian.db
2025/12/17 22:00:01 Successfully connected to AdGuard Home
2025/12/17 22:00:01 Starting poller with interval: 10s
2025/12/17 22:00:01 Fetched 500 queries from AdGuard Home
2025/12/17 22:00:01 [FIRST-SEEN] Client: 192.168.1.100 (192.168.1.100) | Domain: api.example.com | Type: A | Time: 2025-12-17 22:00:01
2025/12/17 22:00:01 [FIRST-SEEN] Client: 192.168.1.101 (192.168.1.101) | Domain: tracking.ads.com | Type: A | Time: 2025-12-17 22:00:01
2025/12/17 22:00:01 Detected 2 new anomalies | Baseline: 2 clients, 2 domains
2025/12/17 22:00:11 Fetched 500 queries from AdGuard Home
2025/12/17 22:00:11 No anomalies detected (500 queries processed, 0 skipped)
```

### Stopping Guardian-Log

Press `Ctrl+C` or send SIGTERM:

```bash
# The application will shut down gracefully
^C
Received shutdown signal, stopping gracefully...
Poller stopped by context
Storage closed
Shutdown complete
```

## Architecture

```
┌─────────────────┐
│  AdGuard Home   │
│   Query Log     │
└────────┬────────┘
         │ HTTP API
         │ (Polling)
         ▼
┌─────────────────┐
│   Guardian-Log  │
│                 │
│  ┌───────────┐  │
│  │  Poller   │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼──────┐ │
│  │ Analyzer   │ │
│  └─────┬──────┘ │
│        │        │
│  ┌─────▼──────┐ │
│  │  BoltDB    │ │
│  │ (Baseline) │ │
│  └────────────┘ │
└─────────────────┘
         │
         ▼
  Terminal Output
  (First-Seen Events)
```

## Configuration

### Environment Variables

#### Core Settings
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `AGH_URL` | AdGuard Home URL | - | Yes |
| `AGH_USER` | AdGuard Home username | - | Yes |
| `AGH_PASS` | AdGuard Home password | - | Yes |
| `POLL_INTERVAL` | How often to poll (e.g., 5s, 10s, 1m) | 10s | No |
| `DB_PATH` | Path to BoltDB database file | ./data/guardian.db | No |
| `LOG_LEVEL` | Log verbosity (info, debug, warn, error) | info | No |

#### LLM Settings (Milestone 2)
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LLM_ENABLE` | Enable/disable LLM analysis | false | No |
| `LLM_PROVIDER` | LLM provider (gemini, openai, anthropic, ollama) | gemini | If LLM enabled |
| `LLM_TIMEOUT` | Timeout for LLM requests | 30s | No |
| `GEMINI_API_KEY` | Google Gemini API key | - | If using Gemini |
| `GEMINI_MODEL` | Gemini model name | gemini-1.5-flash | No |
| `OPENAI_API_KEY` | OpenAI API key | - | If using OpenAI |
| `OPENAI_MODEL` | OpenAI model name | gpt-4o-mini | No |
| `ANTHROPIC_API_KEY` | Anthropic API key | - | If using Anthropic |
| `ANTHROPIC_MODEL` | Anthropic model name | claude-3-5-sonnet-20241022 | No |
| `OLLAMA_URL` | Ollama server URL | http://localhost:11434 | If using Ollama |
| `OLLAMA_MODEL` | Ollama model name | llama3 | No |

## Database

Guardian-Log uses BoltDB (embedded key-value store) to persist:

- **Baselines**: Mapping of `ClientID -> [Domains]`
- **Processed Queries**: Deduplication tracking
- **WHOIS Cache**: Domain WHOIS data (24-hour TTL)
- **Analyses**: LLM analysis results
- **Anomalies**: Detected threats (suspicious/malicious)

The database file is created automatically at the path specified in `DB_PATH`.

### Database Schema

**Buckets:**
- `baselines`: Stores client baselines as JSON
- `processed_queries`: Stores query IDs with timestamps
- `whois_cache`: Cached WHOIS lookups (Milestone 2)
- `analyses`: LLM analysis results (Milestone 2)
- `anomalies`: Detected threats (Milestone 2)

### Inspecting the Database

A database inspection utility is included to view baseline contents:

```bash
# Build the inspection tool
go build -o inspect-db ./cmd/inspect-db

# View database contents
./inspect-db

# Export to JSON
./inspect-db --json
```

**Example Output:**
```
=== Guardian-Log Database Inspection ===
Database: ./data/guardian.db

Total Clients: 3
Total Domains: 147

=== Client Baselines ===

[1] Client: iPhone (192.168.1.100)
    Domains: 52
    Last Updated: 2024-12-17 21:30:15
    Sample Domains (first 10):
      - api.apple.com
      - configuration.apple.com
      - xp.apple.com
      ...
```

**How Baseline Learning Works:**
1. When a domain is seen for the first time for a client, it's logged as `[FIRST-SEEN]`
2. The domain is automatically added to that client's baseline
3. Future queries to the same domain from the same client are not flagged as anomalies
4. The baseline grows over time as devices make new connections

## Troubleshooting

### Connection Issues

If you see authentication errors:
```
Failed to connect to AdGuard Home: authentication failed with status code 401
```

1. Verify your `AGH_USER` and `AGH_PASS` are correct
2. Check that AdGuard Home is accessible at `AGH_URL`
3. Ensure the user has admin privileges

### No Anomalies Detected

If no first-seen events appear:
1. All domains are already in the baseline
2. No new DNS queries are happening
3. Check if AdGuard Home is receiving queries

To reset the baseline (start fresh):
```bash
rm -rf ./data/guardian.db
```

### High Memory Usage

If processing many queries, you can:
1. Increase `POLL_INTERVAL` to reduce frequency
2. Reduce the number of queries fetched (currently hardcoded to 100)

## Development

### Project Structure

```
guardian-log/
├── cmd/
│   └── guardian-log/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── ingestor/
│   │   ├── adguard.go          # AdGuard API client
│   │   └── poller.go           # Polling orchestrator
│   ├── storage/
│   │   ├── boltdb.go           # BoltDB implementation
│   │   ├── models.go           # Data models
│   │   └── utils.go            # Helper functions
│   └── analyzer/
│       └── baseline.go         # Anomaly detection logic
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── PLAN.md
├── SPECIFICATION.md
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o guardian-log ./cmd/guardian-log

# macOS
GOOS=darwin GOARCH=amd64 go build -o guardian-log ./cmd/guardian-log

# Windows
GOOS=windows GOARCH=amd64 go build -o guardian-log.exe ./cmd/guardian-log
```

## Roadmap

### ✅ Milestone 1: Core Ingestor (COMPLETE)
- ✅ AdGuard Home API integration
- ✅ BoltDB persistent storage
- ✅ Real-time anomaly detection
- ✅ Baseline learning system
- ✅ Database inspection utility

### ✅ Milestone 2: LLM Service (COMPLETE)
- ✅ WHOIS/RDAP enrichment with caching
- ✅ Multi-provider LLM architecture
- ✅ Google Gemini provider implementation
- ✅ Intelligent prompt generation
- ✅ Asynchronous analysis pipeline
- ✅ Risk scoring and action recommendations
- ✅ Anomaly storage and tracking

### 📋 Milestone 3: Web Dashboard (PLANNED)
- React-based UI with real-time updates
- Anomaly feed with LLM explanations
- Approve/Block workflow actions
- Settings and configuration management
- WHOIS data visualization
- Risk score trending

### 📋 Milestone 4: Dockerization (PLANNED)
- Multi-stage Docker builds
- Docker Compose orchestration
- One-command deployment
- Health checks and monitoring

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Your chosen license]

## Acknowledgments

- [AdGuard Home](https://github.com/AdguardTeam/AdGuardHome) - DNS filtering
- [BoltDB](https://github.com/etcd-io/bbolt) - Embedded database

---

**Status**: Milestone 2 Complete ✅ | Milestone 3 Next 📋
**Next**: Web dashboard with real-time anomaly feed and LLM analysis results
