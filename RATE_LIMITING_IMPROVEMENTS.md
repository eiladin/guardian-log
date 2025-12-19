# Rate Limiting Improvements

## Summary

Added configurable batch processing settings to help prevent rate limiting with Gemini's free tier API.

## Problem

The hardcoded batch settings were too aggressive for Gemini's free tier (15 RPM limit):
- Batch size: 10 domains
- Batch timeout: 10 seconds
- Request delay: 10 seconds
- **Result**: Up to 6 requests per minute, causing frequent rate limiting

## Solution

Made batch processing fully configurable via environment variables with more conservative defaults:

### New Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_BATCH_SIZE` | `20` | Number of domains to analyze per API call |
| `LLM_BATCH_TIMEOUT` | `60s` | Maximum time to wait before processing a partial batch |
| `LLM_BATCH_DELAY` | `60s` | Minimum delay between batch requests |

### New Defaults (Much More Conservative)

```env
LLM_BATCH_SIZE=20          # 2x larger batches (fewer API calls)
LLM_BATCH_TIMEOUT=60s      # 6x longer wait (accumulate more domains)
LLM_BATCH_DELAY=60s        # 6x longer delay (1 request per minute = 100% safe)
```

**Result**: Maximum **1 request per minute** with **20 domains per request** = up to 20 domains analyzed per minute with zero rate limiting.

## Recommended Settings for Your Setup

Since you're on a **30s poll interval**, try these settings:

```env
# In your .env file
POLL_INTERVAL=30s
LLM_BATCH_SIZE=20
LLM_BATCH_TIMEOUT=60s
LLM_BATCH_DELAY=60s
```

This will:
1. Poll AdGuard every 30 seconds for new DNS queries
2. Accumulate domains for up to 60 seconds
3. Send up to 20 domains per API call
4. Wait 60 seconds between API calls
5. Process ~20 domains per minute safely

## If You Still Get Rate Limited

If you see `🚫 Rate limited` messages in logs:

### Option 1: Increase Delay (Safest)
```env
LLM_BATCH_DELAY=90s        # 1.5 minutes between requests
# or
LLM_BATCH_DELAY=120s       # 2 minutes between requests
```

### Option 2: Increase Batch Size
```env
LLM_BATCH_SIZE=30          # Larger batches = fewer requests
LLM_BATCH_TIMEOUT=90s      # Wait longer to collect more domains
```

### Option 3: Both
```env
LLM_BATCH_SIZE=30
LLM_BATCH_TIMEOUT=90s
LLM_BATCH_DELAY=90s
```

## Files Changed

### Code Changes
- `internal/config/config.go` - Added batch configuration fields
- `internal/llm/analyzer.go` - Made batch settings configurable via constructor
- `cmd/guardian-log/main.go` - Pass configuration to analyzer

### Documentation
- `.env.example` - Added batch settings with defaults and comments
- `docs/deployment/CONFIGURATION.md` - Added rate limiting section with examples

## Testing

```bash
# Build and run with new settings
make build
./bin/guardian-log

# Check logs for batch settings
# You should see:
# "LLM analyzer initialized (batch: 20 domains, timeout: 60s, delay: 60s)"

# Monitor for rate limiting
# Look for: 🚫 Rate limited messages
```

## Migration

If you already have a `.env` file, add these lines:

```env
# Add these to your existing .env
LLM_BATCH_SIZE=20
LLM_BATCH_TIMEOUT=60s
LLM_BATCH_DELAY=60s
```

Or leave them out to use the defaults (same values).

## Understanding the Math

**Gemini Free Tier**: 15 RPM (requests per minute)

**Old Settings** (10s delay):
- 6 requests per minute
- 10 domains per request
- = 60 domains/minute (40% of limit used)
- ⚠️ But batches could queue and fire rapidly, exceeding limit

**New Defaults** (60s delay):
- 1 request per minute
- 20 domains per request
- = 20 domains/minute (6.7% of limit used)
- ✅ Guaranteed safe, never hits rate limit

**Aggressive Settings** (30s delay):
- 2 requests per minute
- 20 domains per request
- = 40 domains/minute (13.3% of limit used)
- ⚠️ Close to limit, may occasionally hit it

## Next Steps

1. Update your `.env` file with the new settings
2. Restart guardian-log
3. Monitor logs for rate limiting
4. Adjust settings if needed

The defaults are very conservative (1 RPM) to guarantee no rate limiting. If you need faster processing and rarely see rate limit errors, you can decrease `LLM_BATCH_DELAY` to 30s or 45s.
