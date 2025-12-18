# Docker Deployment Guide

## Overview
Guardian-Log is now fully containerized with support for:
- ✅ Multi-stage builds (optimized image size)
- ✅ Multi-architecture support (amd64, arm64)
- ✅ Docker Compose for easy deployment
- ✅ Comprehensive Makefile for build automation
- ✅ Health checks and resource limits
- ✅ Non-root user execution

## Quick Start with Docker Compose

### 1. Configure Environment
```bash
cp .env.example .env
# Edit .env with your settings
```

### 2. Start Guardian-Log
```bash
make docker-compose-up
```

Or directly:
```bash
docker-compose up -d
```

### 3. Access Dashboard
Open http://localhost:8080

### 4. View Logs
```bash
make docker-compose-logs
```

Or:
```bash
docker-compose logs -f guardian-log
```

### 5. Stop Services
```bash
make docker-compose-down
```

## Makefile Commands

### Building

```bash
# Build complete application (frontend + backend)
make build

# Quick build (skip npm install)
make build-quick

# Build Docker image for current platform
make docker-build

# Build multi-architecture images (amd64 + arm64)
make docker-build-multi

# Build multi-arch locally (no push)
make docker-build-local
```

### Running

```bash
# Build and run locally
make run

# Run Docker container
make docker-run

# Start with docker-compose
make docker-compose-up

# Stop docker-compose
make docker-compose-down

# View logs
make docker-compose-logs
```

### Development

```bash
# Show all available commands
make help

# Install dependencies
make install

# Run in development mode (separate servers)
make dev

# Run backend only (dev mode)
make dev-backend

# Run frontend only (dev mode)
make dev-frontend
```

### Maintenance

```bash
# Clean build artifacts
make clean

# Run tests
make test

# Run linters
make lint

# Format code
make fmt
```

## Docker Architecture

### Multi-Stage Build

**Stage 1: Frontend Builder**
- Base: `node:20-alpine`
- Builds React app with Vite
- Output: `web/dist/`

**Stage 2: Backend Builder**
- Base: `golang:1.21-alpine`
- Copies frontend dist
- Builds Go binary with embedded frontend
- Output: Single optimized binary

**Stage 3: Runtime**
- Base: `alpine:latest`
- Minimal runtime environment
- Non-root user execution
- Health checks enabled

### Image Size
- **Frontend builder**: ~200 MB (discarded)
- **Backend builder**: ~800 MB (discarded)
- **Final image**: ~25 MB (runtime only)

### Multi-Architecture Support

Guardian-Log supports multiple CPU architectures:
- **linux/amd64**: x86_64 (Intel/AMD)
- **linux/arm64**: ARM 64-bit (Raspberry Pi 4, M1 Macs, etc.)

Build for multiple architectures:
```bash
make docker-build-multi PLATFORMS=linux/amd64,linux/arm64
```

## Deployment Scenarios

### Scenario 1: Docker Compose (Recommended)

**Best for**: Local deployments, home labs, simple setups

```bash
# Start
docker-compose up -d

# Update
docker-compose pull
docker-compose up -d

# Stop
docker-compose down
```

### Scenario 2: Docker Run

**Best for**: Quick tests, CI/CD

```bash
docker run -d \
  --name guardian-log \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --env-file .env \
  guardian-log:latest
```

### Scenario 3: Kubernetes

**Best for**: Production clusters

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guardian-log
spec:
  replicas: 1
  selector:
    matchLabels:
      app: guardian-log
  template:
    metadata:
      labels:
        app: guardian-log
    spec:
      containers:
      - name: guardian-log
        image: guardian-log:latest
        ports:
        - containerPort: 8080
        env:
        - name: AGH_URL
          valueFrom:
            secretKeyRef:
              name: guardian-secrets
              key: agh-url
        - name: GEMINI_API_KEY
          valueFrom:
            secretKeyRef:
              name: guardian-secrets
              key: gemini-key
        volumeMounts:
        - name: data
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: guardian-data
```

### Scenario 4: Docker Swarm

**Best for**: Simple multi-node setups

```bash
docker stack deploy -c docker-compose.yml guardian
```

## Configuration

### Environment Variables

All configuration via environment variables:

```env
# AdGuard Home
AGH_URL=http://192.168.1.1:8080
AGH_USER=admin
AGH_PASS=your_password

# Application
POLL_INTERVAL=10s
DB_PATH=./data/guardian.db
LOG_LEVEL=info

# LLM
LLM_ENABLE=true
LLM_PROVIDER=gemini
LLM_TIMEOUT=30s

# Gemini
GEMINI_API_KEY=your_key_here
GEMINI_MODEL=gemini-1.5-flash
```

### Volume Mounts

**Data Persistence**:
```bash
-v $(pwd)/data:/app/data
```

Mounts local `./data` directory to persist:
- BoltDB database
- Baseline data
- Anomaly records

**Custom Config** (optional):
```bash
-v $(pwd)/.env:/app/.env:ro
```

Mount custom environment file (read-only).

## Advanced Usage

### Build for Specific Architecture

```bash
# Build for ARM64 only (Raspberry Pi, etc.)
docker buildx build --platform linux/arm64 -t guardian-log:arm64 .

# Build for AMD64 only
docker buildx build --platform linux/amd64 -t guardian-log:amd64 .
```

### Push to Registry

```bash
# Using Makefile
make docker-push DOCKER_REGISTRY=your-registry.com

# Using docker directly
docker tag guardian-log:latest your-registry.com/guardian-log:latest
docker push your-registry.com/guardian-log:latest
```

### Resource Limits

Edit `docker-compose.yml`:
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 1G
    reservations:
      cpus: '1'
      memory: 512M
```

### Health Checks

Container includes health check:
```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' guardian-log

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' guardian-log
```

Health check hits: `http://localhost:8080/api/health`

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs guardian-log

# Common issues:
# 1. Missing .env file
# 2. Invalid GEMINI_API_KEY
# 3. Can't reach AdGuard Home (AGH_URL)
```

### Database Locked

```bash
# If you see "database is locked"
docker-compose down
rm -f data/guardian.db-lock
docker-compose up -d
```

### Permission Errors

```bash
# Ensure data directory is writable
chmod 755 data/
```

Container runs as non-root user (UID 1000).

### Can't Connect to AdGuard Home

```bash
# Test from container
docker-compose exec guardian-log wget -O- $AGH_URL/control/status
```

If AGH is on host machine, use:
- Linux: `http://172.17.0.1:8080` (Docker bridge IP)
- macOS: `http://host.docker.internal:8080`
- Windows: `http://host.docker.internal:8080`

### Multi-Arch Build Issues

Ensure buildx is set up:
```bash
# Create builder
docker buildx create --name guardian-builder --use

# Verify
docker buildx ls
```

## Development Workflow

### Local Development

```bash
# Terminal 1: Backend
make dev-backend

# Terminal 2: Frontend (hot reload)
make dev-frontend
```

### Test Docker Build Locally

```bash
# Build
make docker-build

# Run
make docker-run

# Test
curl http://localhost:8080/api/health
```

### Iterate and Rebuild

```bash
# Make changes
# ...

# Rebuild
make docker-build

# Restart
docker-compose down
docker-compose up -d
```

## Production Checklist

- [ ] Use specific version tags (not `latest`)
- [ ] Set resource limits in compose file
- [ ] Use secrets management for API keys
- [ ] Enable log rotation
- [ ] Set up monitoring/alerts
- [ ] Configure backups for `data/` volume
- [ ] Use TLS/HTTPS (reverse proxy)
- [ ] Restrict network access (firewall)
- [ ] Regular updates (`docker-compose pull`)

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Build multi-arch
        run: make docker-build-multi VERSION=${{ github.ref_name }}
```

### GitLab CI Example

```yaml
build:
  stage: build
  script:
    - make docker-build
    - make docker-push DOCKER_REGISTRY=$CI_REGISTRY_IMAGE
  only:
    - main
    - tags
```

## Monitoring

### Prometheus Metrics (Future)

Guardian-Log can export metrics:
```yaml
# docker-compose.yml
ports:
  - "8080:8080"  # API + Dashboard
  - "9090:9090"  # Metrics (future)
```

### Log Aggregation

Send logs to external system:
```bash
docker run -d \
  --log-driver=json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  guardian-log:latest
```

Or use Docker logging driver (syslog, gelf, etc.).

## Summary

✅ **Makefile**: Comprehensive build automation
✅ **Dockerfile**: Multi-stage, optimized, multi-arch
✅ **Docker Compose**: One-command deployment
✅ **Health Checks**: Container monitoring
✅ **Non-root User**: Security best practices
✅ **Volume Mounts**: Data persistence

Guardian-Log is now production-ready for container deployments!

---

**Quick Commands Cheat Sheet**:
```bash
make help                  # Show all commands
make build                 # Build locally
make docker-build          # Build Docker image
make docker-compose-up     # Start services
make docker-compose-logs   # View logs
make docker-compose-down   # Stop services
```
