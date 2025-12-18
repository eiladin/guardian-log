# Installation Guide

Complete installation guide for Guardian-Log.

## Quick Start

See [README Quick Start](../../README.md#quick-start) for the fastest way to get running.

## Installation Methods

### 1. Docker Compose (Recommended)

```bash
# Clone repository
git clone https://github.com/OWNER/guardian-log.git
cd guardian-log

# Configure
cp .env.example .env
vim .env

# Start
docker-compose up -d
```

### 2. Docker

```bash
docker run -d \
  --name guardian-log \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  ghcr.io/OWNER/guardian-log:latest
```

### 3. Pre-built Binary

Download from [GitHub Releases](https://github.com/OWNER/guardian-log/releases):

```bash
# Download latest release
curl -LO https://github.com/OWNER/guardian-log/releases/latest/download/guardian-log-linux-amd64

# Make executable
chmod +x guardian-log-linux-amd64

# Run
./guardian-log-linux-amd64
```

### 4. Build from Source

**Prerequisites:**
- Go 1.25+
- Node.js 20+
- Make

```bash
# Clone
git clone https://github.com/OWNER/guardian-log.git
cd guardian-log

# Build
make build

# Run
./bin/guardian-log
```

## Prerequisites

### Required

- **AdGuard Home** - Running and accessible
- **Gemini API Key** - Free at https://aistudio.google.com/app/apikey

### Optional

- Docker & Docker Compose (for container deployment)
- Go 1.25+ (for building from source)
- Node.js 20+ (for frontend development)

## Post-Installation

After installation:

1. **Configure** - See [Configuration Guide](CONFIGURATION.md)
2. **Start** - Run Guardian-Log
3. **Access** - Open http://localhost:8080
4. **Monitor** - Check logs for anomaly detection

## Next Steps

- [Configuration Guide](CONFIGURATION.md) - Configure all options
- [Docker Deployment](DOCKER.md) - Advanced Docker setup
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues

---

**[← Back to Docs](../README.md)**
