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
| `LLM_BATCH_SIZE` | Domains per batch request | No | `20` |
| `LLM_BATCH_TIMEOUT` | Max wait before flushing batch | No | `60s` |
| `LLM_BATCH_DELAY` | Minimum delay between batches | No | `60s` |

### Gemini (Recommended)

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `GEMINI_API_KEY` | API key from Google | Yes | - |
| `GEMINI_MODEL` | Model name | No | `gemini-1.5-flash` |

Get free API key: https://aistudio.google.com/app/apikey

### Rate Limiting Configuration

**Important**: Adjust batch settings to avoid rate limiting with Gemini free tier.

Gemini Free Tier Limits (gemini-1.5-flash):
- **15 RPM** (requests per minute)
- **1 million TPM** (tokens per minute)
- **1,500 RPD** (requests per day)

#### Recommended Settings for Gemini Free Tier

```env
# Conservative settings (recommended for free tier)
LLM_BATCH_SIZE=20          # Process 20 domains per API call
LLM_BATCH_TIMEOUT=60s      # Wait up to 60s to collect domains
LLM_BATCH_DELAY=60s        # Wait 60s between batches (1 RPM = safe)
POLL_INTERVAL=30s          # Check for new DNS queries every 30s
```

This configuration processes up to **20 domains per minute** with minimal rate limiting.

#### Aggressive Settings (for testing or paid tiers)

```env
# Faster processing (may hit rate limits on free tier)
LLM_BATCH_SIZE=10          # Smaller batches
LLM_BATCH_TIMEOUT=30s      # Flush more frequently
LLM_BATCH_DELAY=30s        # 2 RPM (risky for free tier)
```

#### Very Conservative Settings (for heavy traffic)

```env
# Maximum safety (for high DNS query volume)
LLM_BATCH_SIZE=30          # Larger batches
LLM_BATCH_TIMEOUT=120s     # Wait longer to collect more domains
LLM_BATCH_DELAY=90s        # 40 seconds per batch minimum
```

**Tip**: Monitor the logs for `🚫 Rate limited` messages. If you see these frequently:
1. Increase `LLM_BATCH_DELAY` (e.g., from 60s to 90s)
2. Increase `LLM_BATCH_SIZE` (e.g., from 20 to 30)
3. Increase `LLM_BATCH_TIMEOUT` to allow more domains to accumulate

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
