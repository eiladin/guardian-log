# Configuration Guide

Complete configuration reference for Guardian-Log.

## Quick Configuration

Copy and edit `.env`:

```bash
cp .env.example .env
vim .env
```

## Required Settings

```env
# AdGuard Home
AGH_URL=http://192.168.1.1:8080
AGH_USER=admin
AGH_PASS=your_password

# AI Analysis
GEMINI_API_KEY=your_gemini_api_key
```

## All Environment Variables

See `.env.example` for complete list with defaults.

### AdGuard Home

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `AGH_URL` | AdGuard Home URL | Yes | - |
| `AGH_USER` | Admin username | Yes | - |
| `AGH_PASS` | Admin password | Yes | - |

### Application

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `POLL_INTERVAL` | Polling frequency | No | `10s` |
| `DB_PATH` | Database file path | No | `./data/guardian.db` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | No | `info` |

### LLM Provider

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `LLM_ENABLE` | Enable AI analysis | No | `true` |
| `LLM_PROVIDER` | Provider (gemini/openai/anthropic/ollama) | No | `gemini` |
| `LLM_TIMEOUT` | Request timeout | No | `30s` |

### Gemini (Recommended)

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GEMINI_API_KEY` | API key from Google | Yes | - |
| `GEMINI_MODEL` | Model name | No | `gemini-1.5-flash` |

Get free API key: https://aistudio.google.com/app/apikey

## Advanced Configuration

### Polling Interval

```env
POLL_INTERVAL=30s  # Poll every 30 seconds
POLL_INTERVAL=1m   # Poll every minute
```

### Log Levels

```env
LOG_LEVEL=debug  # Verbose logging
LOG_LEVEL=info   # Normal logging (default)
LOG_LEVEL=warn   # Warnings only
LOG_LEVEL=error  # Errors only
```

### Database Path

```env
DB_PATH=./data/guardian.db      # Relative path
DB_PATH=/var/lib/guardian.db    # Absolute path
```

## Docker Configuration

When using Docker, mount data directory:

```yaml
volumes:
  - ./data:/app/data
```

## Kubernetes Configuration

Use ConfigMaps and Secrets:

```yaml
env:
  - name: AGH_URL
    valueFrom:
      configMapKeyRef:
        name: guardian-config
        key: agh-url
  - name: GEMINI_API_KEY
    valueFrom:
      secretKeyRef:
        name: guardian-secrets
        key: gemini-key
```

## Validation

Check configuration:

```bash
# View current config
./guardian-log --config

# Test AdGuard connection
curl -u $AGH_USER:$AGH_PASS $AGH_URL/control/status
```

---

**[← Back to Docs](../README.md)**
