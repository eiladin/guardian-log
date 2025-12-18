# System Architecture

Guardian-Log system design and components.

## High-Level Architecture

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

## Components

### 1. Ingestor
- **Location:** `internal/ingestor/`
- **Purpose:** Polls AdGuard Home for DNS queries
- **Tech:** HTTP client with authentication
- **Frequency:** Configurable (default 10s)

### 2. Analyzer
- **Location:** `internal/analyzer/`
- **Purpose:** Detects anomalies using baseline learning
- **Algorithm:** First-seen domain detection per client
- **Storage:** BoltDB for persistent baselines

### 3. LLM Service
- **Location:** `internal/llm/`
- **Purpose:** AI-powered threat analysis
- **Providers:** Gemini, OpenAI, Anthropic, Ollama
- **Features:** Risk scoring, recommendations

### 4. API Server
- **Location:** `internal/api/`
- **Purpose:** REST API + embedded web UI
- **Tech:** Go HTTP server with CORS
- **Endpoints:** /api/anomalies, /api/stats

### 5. Web Dashboard
- **Location:** `web/`
- **Purpose:** User interface for managing alerts
- **Tech:** React 18 + TypeScript + Vite
- **Features:** Real-time updates, approve/block workflows

### 6. Storage
- **Location:** `internal/storage/`
- **Purpose:** Data persistence
- **Tech:** BoltDB (embedded key-value store)
- **Buckets:** baselines, anomalies, whois_cache

## Data Flow

```
1. AdGuard Query
   ↓
2. Poller Fetches
   ↓
3. Baseline Check
   ↓
4. Anomaly Detected?
   ├─ No → Skip
   └─ Yes → 5. LLM Analysis
            ↓
         6. Store Result
            ↓
         7. Display in UI
            ↓
         8. User Action
            ├─ Approve → Add to Baseline
            └─ Block → Add to AdGuard Blocklist
```

## Technology Stack

**Backend:**
- Language: Go 1.25+
- Database: BoltDB
- HTTP Server: net/http
- LLM: Google Gemini API

**Frontend:**
- Framework: React 18
- Language: TypeScript
- Build: Vite
- Styling: Tailwind CSS

**Deployment:**
- Containers: Docker
- Orchestration: Docker Compose
- Multi-arch: amd64, arm64
- CI/CD: GitHub Actions

## Design Principles

1. **Self-hosted** - No cloud dependencies except LLM API
2. **Embedded** - Single binary deployment
3. **Real-time** - Continuous monitoring with fast feedback
4. **AI-powered** - Intelligent threat detection
5. **User-friendly** - Simple web UI for management

## Scalability

**Current:** Single instance, single database
**Future:** Multi-instance, distributed storage

## Security

- Non-root container execution
- No secrets in codebase
- CORS-protected API
- Environment-based configuration

---

**[← Back to Docs](../README.md)**
