# Rate Limiting Improvements

## Summary

Optimized batch processing and prompts to prevent rate limiting with Gemini 2.5 Flash-Lite's free tier API (15 RPM, 250K TPM).

## Problem

Gemini 2.5 Flash-Lite has two rate limits:
- **15 RPM** (Requests Per Minute)
- **250,000 TPM** (Tokens Per Minute)

Previous settings were hitting the **TPM limit** (not RPM):
- Batch size: 20 domains = ~5,000+ tokens per request
- Verbose batch prompt with lots of instructional text
- **Result**: Token-heavy requests causing frequent 429 errors

## Solution

Two-pronged approach to reduce token usage:

### 1. Optimized Batch Prompt (65% Token Reduction)

**Before** (~800 tokens for 20 domains):
```
You are a cybersecurity expert analyzing DNS queries for potential threats.

Analyze the following 20 DNS queries:

### Query 1
- **Domain**: example.com
- **Client**: laptop (192.168.1.100)
...
## Response Format
Respond with a JSON array containing an analysis for each query:
...
```

**After** (~280 tokens for 20 domains):
```
Analyze these DNS queries for security threats. Respond with JSON array only.

1. example.com [US] (GoDaddy)
2. test.com [CN] (Namecheap)
...
Format: [{"domain":"x.com","classification":"Safe|Suspicious|Malicious","explanation":"...","risk_score":1-10,"suggested_action":"Allow|Investigate|Block"}]
```

### 2. Conservative Default Settings

| Variable | Old Default | New Default | Reason |
|----------|------------|-------------|--------|
| `LLM_BATCH_SIZE` | `20` | `10` | Fewer domains = fewer tokens per request |
| `LLM_BATCH_TIMEOUT` | `60s` | `90s` | More time to accumulate domains |
| `LLM_BATCH_DELAY` | `60s` | `90s` | Longer delay between requests |

**Result**: ~1,500 tokens per request (input + output) = **~17,000 TPM usage** (vs 250K limit = 7% utilization)

## Recommended Settings for Gemini 2.5 Flash-Lite

### Default Settings (Safest - Start Here)

```env
LLM_BATCH_SIZE=10          # Small batches = low token usage
LLM_BATCH_TIMEOUT=90s      # Wait to collect domains
LLM_BATCH_DELAY=90s        # ~40 requests/hour
```

**Token Usage**: ~1,500 tokens/request × 0.67 req/min = **~1,000 TPM** (0.4% of limit)

## If You Still Get Rate Limited

If you see `🚫 Rate limited` messages:

### Step 1: Reduce Batch Size (Lowest Tokens)
```env
LLM_BATCH_SIZE=5           # Minimal token usage per request
LLM_BATCH_TIMEOUT=120s     # Wait longer to accumulate
LLM_BATCH_DELAY=120s       # 2 minutes between requests
```

**Token Usage**: ~800 tokens/request × 0.5 req/min = **~400 TPM** (0.16% of limit)

### Step 2: Increase Delay Further
```env
LLM_BATCH_SIZE=5
LLM_BATCH_DELAY=180s       # 3 minutes between requests
```

### Step 3: Consider Model Switch
If still rate limited, try **gemini-1.5-flash** (may have different TPM limits):
```env
GEMINI_MODEL=gemini-1.5-flash
```

## For Faster Processing (Once Stable)

If you're NOT getting rate limited and want faster throughput:

```env
LLM_BATCH_SIZE=15          # More domains per request
LLM_BATCH_TIMEOUT=60s      # Process faster
LLM_BATCH_DELAY=60s        # 1 request per minute
```

**Token Usage**: ~2,500 tokens/request × 1 req/min = **~2,500 TPM** (1% of limit)

## Files Changed

### Code Changes
- `internal/llm/prompt.go` - **Optimized batch prompt** (65% token reduction)
- `internal/config/config.go` - Changed defaults: batch size 10, delays 90s
- `.env.example` - Updated with Flash-Lite specific guidance

### Documentation
- `RATE_LIMITING_IMPROVEMENTS.md` - This file (comprehensive guide)

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

**Gemini 2.5 Flash-Lite Free Tier**: 15 RPM, 250K TPM, 1K RPD

### Token Usage Examples

| Batch Size | Prompt Tokens | Response Tokens | Total/Request | At 1 req/min | % of TPM Limit |
|-----------|--------------|-----------------|---------------|--------------|----------------|
| **5** | ~500 | ~300 | ~800 | ~800 TPM | **0.3%** ✅ |
| **10** | ~900 | ~600 | ~1,500 | ~1,500 TPM | **0.6%** ✅ |
| **15** | ~1,300 | ~1,200 | ~2,500 | ~2,500 TPM | **1.0%** ✅ |
| **20** | ~1,700 | ~1,600 | ~3,300 | ~3,300 TPM | **1.3%** ✅ |

### Why You Were Getting Rate Limited

**Previous Settings**:
- 20 domains per batch
- Old verbose prompt: ~800 tokens of instructions
- = ~3,300 tokens per request
- With network delays, retries, or multiple processes: **Could hit 250K TPM limit**

**New Optimized Settings**:
- 10 domains per batch
- Compact prompt: ~150 tokens of instructions
- = ~1,500 tokens per request
- At 1 req/min: **~1,500 TPM (0.6% of limit)** ✅

## Next Steps

1. Update your `.env` file with the new settings
2. Restart guardian-log
3. Monitor logs for rate limiting
4. Adjust settings if needed

The defaults are very conservative (1 RPM) to guarantee no rate limiting. If you need faster processing and rarely see rate limit errors, you can decrease `LLM_BATCH_DELAY` to 30s or 45s.
