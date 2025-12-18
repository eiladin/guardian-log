# Guardian-Log Implementation Plan

## Project Overview
Guardian-Log is a self-hosted LLM-powered DNS anomaly detection and explanation system that integrates with AdGuard Home. It monitors DNS queries, identifies anomalous patterns, and provides human-readable explanations using various LLM providers.

## Technology Stack
- **Backend:** Go 1.21+
- **Database:** BoltDB (embedded KV store)
- **Frontend:** React + TypeScript + Vite
- **Integration:** AdGuard Home API
- **LLM Providers:** Claude, GPT-4, Gemini, Ollama (local)
- **Deployment:** Docker + Docker Compose

---

## Project Structure

```
guardian-log/
├── cmd/
│   └── guardian-log/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── ingestor/
│   │   ├── adguard.go             # AdGuard Home API client
│   │   └── poller.go              # Polling logic
│   ├── storage/
│   │   ├── boltdb.go              # BoltDB wrapper
│   │   └── models.go              # Data models
│   ├── analyzer/
│   │   ├── baseline.go            # Baseline/anomaly detection
│   │   └── enrichment.go          # WHOIS/RDAP enrichment
│   ├── llm/
│   │   ├── provider.go            # LLM provider interface
│   │   ├── openai.go              # OpenAI implementation
│   │   ├── anthropic.go           # Anthropic implementation
│   │   ├── gemini.go              # Google Gemini implementation
│   │   └── ollama.go              # Ollama implementation
│   └── api/
│       ├── server.go              # HTTP server
│       └── handlers.go            # API endpoints
├── web/
│   ├── src/
│   │   ├── components/           # React components
│   │   ├── pages/                # Page components
│   │   ├── api/                  # API client
│   │   └── App.tsx               # Main app component
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── go.mod
├── go.sum
├── SPECIFICATION.md
└── PLAN.md
```

---

## Milestone 1: Core Ingestor

### Deliverable
AdGuard API integration + BoltDB storage logic with real-time "First Seen" event logging.

### Implementation Steps

#### 1.1 Project Initialization
- [x] Create Go module (`go mod init github.com/eiladin/guardian-log`)
- [ ] Set up basic project structure (cmd/, internal/)
- [ ] Create .env.example with required variables
- [ ] Add .gitignore for Go projects

#### 1.2 Configuration Management
- [ ] Create config package
- [ ] Implement environment variable loading
- [ ] Define configuration struct:
  - AdGuard Home URL, username, password
  - Polling interval (default: 10s)
  - BoltDB path
  - Log level
- [ ] Add validation for required configs

#### 1.3 AdGuard Home API Client
- [ ] Research AdGuard Home API `/control/querylog` endpoint
- [ ] Create AdGuard client struct with authentication
- [ ] Implement query log fetching:
  - Parse JSON response
  - Extract: Client ID, Client Name, Domain, Timestamp, Query Type
- [ ] Add error handling and retries
- [ ] Implement pagination/cursor handling (if API supports it)

#### 1.4 Data Models
- [ ] Define core data structures:
  ```go
  type DNSQuery struct {
      ClientID    string
      ClientName  string
      Domain      string
      Timestamp   time.Time
      QueryType   string
  }

  type Baseline struct {
      ClientID    string
      Domains     []string
      LastUpdated time.Time
  }
  ```

#### 1.5 BoltDB Storage Layer
- [ ] Initialize BoltDB connection
- [ ] Create buckets:
  - `baselines` - stores client -> domains mapping
  - `queries` - stores processed query IDs for deduplication
- [ ] Implement operations:
  - `GetClientBaseline(clientID) -> []string`
  - `AddDomainToBaseline(clientID, domain)`
  - `HasSeenQuery(queryID) -> bool`
  - `MarkQueryProcessed(queryID)`
- [ ] Add proper error handling and transaction management

#### 1.6 Baseline & Anomaly Detection
- [ ] Implement anomaly detection logic:
  - Check if domain exists in client's baseline
  - If not found, flag as "first-seen"
- [ ] Create event logging:
  - Format: `[FIRST-SEEN] Client: {name} ({id}) | Domain: {domain} | Time: {timestamp}`
  - Log to stdout with structured logging

#### 1.7 Polling Orchestrator
- [ ] Create polling loop with configurable interval
- [ ] Implement graceful shutdown (handle SIGTERM/SIGINT)
- [ ] Add deduplication logic:
  - Track processed query IDs
  - Skip already-processed queries
- [ ] Integrate all components:
  - Poll AdGuard -> Check baseline -> Log anomalies -> Store in DB

#### 1.8 Testing & Validation
- [ ] Create test AdGuard Home instance (or use mock data)
- [ ] Verify polling works correctly
- [ ] Test baseline persistence across restarts
- [ ] Confirm "first-seen" events only fire once per client/domain pair
- [ ] Test error scenarios (AdGuard down, DB corruption, etc.)

### Dependencies (M1)
```
go get go.etcd.io/bbolt         # BoltDB
go get github.com/joho/godotenv # Environment variables
go get go.uber.org/zap          # Structured logging (optional)
```

### Checkpoint Criteria
✅ Application successfully polls AdGuard Home API
✅ BoltDB stores and retrieves baselines correctly
✅ "First-seen" events log to terminal in real-time
✅ Application handles restarts without duplicate logging
✅ Graceful shutdown works properly

---

## Milestone 2: LLM Service

### Deliverable
Multi-provider LLM client with WHOIS enrichment returning structured JSON explanations for detected anomalies.

### Overview
In Milestone 2, we'll add intelligence to Guardian-Log by enriching anomaly detections with domain information (WHOIS/RDAP) and using LLMs to provide human-readable explanations with risk assessments. This transforms raw DNS anomalies into actionable security insights.

### Implementation Steps

#### 2.1 Update Configuration for LLM Settings
- [ ] Extend `internal/config/config.go` to include:
  - LLM provider selection (openai, anthropic, gemini, ollama)
  - API keys for cloud providers
  - Ollama endpoint and model configuration
  - LLM timeout settings
- [ ] Add validation for LLM configuration
- [ ] Update `.env.example` with LLM settings

#### 2.2 Define Core LLM Types and Interfaces
- [ ] Create `internal/llm/types.go` with:
  ```go
  type Provider interface {
      Analyze(ctx context.Context, query storage.DNSQuery, whois WHOISData) (*Analysis, error)
      Name() string
  }

  type Analysis struct {
      Domain          string    `json:"domain"`
      ClientID        string    `json:"client_id"`
      Classification  string    `json:"classification"`  // Safe, Suspicious, Malicious
      Explanation     string    `json:"explanation"`
      RiskScore       int       `json:"risk_score"`      // 1-10
      SuggestedAction string    `json:"suggested_action"` // Allow, Investigate, Block
      AnalyzedAt      time.Time `json:"analyzed_at"`
      Provider        string    `json:"provider"`
  }

  type WHOISData struct {
      Domain       string `json:"domain"`
      Registrar    string `json:"registrar"`
      Country      string `json:"country"`
      CreatedDate  string `json:"created_date"`
      UpdatedDate  string `json:"updated_date"`
      ExpiryDate   string `json:"expiry_date"`
      NameServers  []string `json:"name_servers"`
  }
  ```
- [ ] Create provider factory function based on configuration

#### 2.3 WHOIS/RDAP Enrichment Service
- [ ] Create `internal/enrichment/whois.go`
- [ ] Implement WHOIS client using `github.com/likexian/whois`
- [ ] Parse WHOIS responses using `github.com/likexian/whois-parser`
- [ ] Extract key fields: registrar, country, dates, nameservers
- [ ] Add caching layer to prevent duplicate lookups:
  - Use BoltDB bucket `whois_cache`
  - Cache TTL: 24 hours
- [ ] Handle lookup failures gracefully (return partial data)
- [ ] Add rate limiting to respect WHOIS server policies
- [ ] Log WHOIS lookup statistics

#### 2.4 System Prompt Design
- [ ] Create `internal/llm/prompt.go` with prompt builder
- [ ] Define system prompt template:
  ```
  You are a Network Security Analyst analyzing DNS queries from a home network.
  Provide concise, actionable security analysis.

  Context:
  - Device: {client_name} ({client_id})
  - Domain: {domain}
  - Query Type: {query_type}
  - WHOIS Data: {whois_json}

  Respond ONLY with valid JSON in this exact format:
  {
    "classification": "Safe|Suspicious|Malicious",
    "explanation": "1-2 sentence explanation of what this domain is and why the device is accessing it",
    "risk_score": <1-10>,
    "suggested_action": "Allow|Investigate|Block"
  }
  ```
- [ ] Implement JSON response validation and parsing
- [ ] Add retry logic for malformed responses (up to 2 retries)

#### 2.5 LLM Provider Implementations

##### 2.5.1 Gemini Provider (Start with this - generous free tier!)
- [ ] Create `internal/llm/gemini.go`
- [ ] Install SDK: `go get github.com/google/generative-ai-go/genai`
- [ ] Use Gemini 1.5 Flash (fast and free tier available)
- [ ] Implement GenerateContent API
- [ ] Configure JSON response mode using schema
- [ ] Handle safety filters and content blocking
- [ ] Add rate limiting (15 RPM free tier)
- [ ] Test with API key from https://aistudio.google.com/app/apikey
- [ ] Add request/response logging for debugging

**Why Gemini First:**
- Free tier: 15 requests/minute, 1500 requests/day
- Fast responses (< 2 seconds typically)
- Good at structured JSON output
- No local setup required

##### 2.5.2 Ollama Provider (Optional - for fully local setup)
- [ ] Create `internal/llm/ollama.go`
- [ ] Implement HTTP client for Ollama API
- [ ] Support configurable endpoint (default: http://localhost:11434)
- [ ] Support model selection (llama3, mistral, etc.)
- [ ] Test with local Ollama instance
- [ ] Handle connection errors gracefully

##### 2.5.3 OpenAI Provider (Optional - add if needed)
- [ ] Create `internal/llm/openai.go`
- [ ] Use `github.com/sashabaranov/go-openai`
- [ ] Implement ChatCompletion API with GPT-4o-mini (cost-effective)
- [ ] Configure response format as JSON
- [ ] Handle API errors and rate limits
- [ ] Add usage/cost tracking

##### 2.5.4 Anthropic Provider (Optional - add if needed)
- [ ] Create `internal/llm/anthropic.go`
- [ ] Use `github.com/anthropics/anthropic-sdk-go`
- [ ] Implement Messages API with Claude 3.5 Sonnet
- [ ] Use JSON mode for structured output
- [ ] Handle streaming responses
- [ ] Add token usage tracking

#### 2.6 Analysis Orchestrator
- [ ] Create `internal/analyzer/llm_analyzer.go`
- [ ] Implement analysis pipeline:
  1. Receive anomaly (DNSQuery) from baseline analyzer
  2. Perform WHOIS lookup (with caching)
  3. Build analysis prompt with context
  4. Call configured LLM provider
  5. Parse and validate JSON response
  6. Store analysis result in database
- [ ] Use goroutines + channels for async processing
- [ ] Implement analysis queue to handle bursts
- [ ] Add timeout handling (30s per analysis)
- [ ] Add circuit breaker pattern for LLM failures
- [ ] Log analysis statistics (success rate, avg time, etc.)

#### 2.7 Extend Storage for LLM Analysis
- [ ] Update `internal/storage/boltdb.go` with new buckets:
  - `anomalies` - Pending anomalies awaiting LLM analysis
  - `analyses` - Completed LLM analyses
  - `whois_cache` - Cached WHOIS lookups
- [ ] Implement storage operations:
  - `SaveAnomaly(anomaly)` - Store new anomaly
  - `GetPendingAnomalies()` - Get unanalyzed anomalies
  - `SaveAnalysis(analysis)` - Store LLM analysis result
  - `GetAnalyses(filters)` - Query analyses by date/risk/client
  - `CacheWHOIS(domain, data, ttl)` - Cache WHOIS data
  - `GetCachedWHOIS(domain)` - Retrieve cached WHOIS
- [ ] Add indexes for efficient queries

#### 2.8 Integration with Poller
- [ ] Modify `internal/ingestor/poller.go`:
  - When anomaly detected, save to `anomalies` bucket
  - Trigger async LLM analysis
  - Continue with normal flow (don't block polling)
- [ ] Add configuration flag to enable/disable LLM analysis
- [ ] Log when LLM analysis is queued vs completed

### Dependencies (M2)
```bash
# Core dependencies (required)
go get github.com/google/generative-ai-go/genai   # Gemini (primary LLM)
go get github.com/likexian/whois                  # WHOIS lookups
go get github.com/likexian/whois-parser           # WHOIS parsing

# Optional LLM providers (add as needed)
go get github.com/sashabaranov/go-openai          # OpenAI
go get github.com/anthropics/anthropic-sdk-go     # Anthropic
# Ollama uses HTTP client (no SDK needed)
```

### Testing Strategy (M2)

#### Unit Tests
- [ ] WHOIS client parsing
- [ ] LLM prompt generation
- [ ] JSON response validation
- [ ] Provider factory logic

#### Integration Tests
- [ ] End-to-end analysis flow (anomaly → WHOIS → LLM → storage)
- [ ] Each LLM provider integration
- [ ] WHOIS caching behavior
- [ ] Analysis queue under load

#### Manual Testing
- [ ] Get Gemini API key from https://aistudio.google.com/app/apikey
- [ ] Test with Gemini 1.5 Flash (primary provider)
- [ ] Verify analysis quality and relevance
- [ ] Test error handling (LLM timeout, invalid responses, safety blocks, etc.)
- [ ] Verify rate limiting respects free tier (15 RPM)
- [ ] Test with various domain types (safe, suspicious, malicious)

### Checkpoint Criteria (M2)
✅ Gemini provider working end-to-end
✅ WHOIS enrichment retrieving domain information
✅ Analyses stored in database with complete metadata
✅ LLM analysis doesn't block DNS query polling
✅ JSON responses validated and parsed correctly
✅ Rate limiting prevents API quota exhaustion
✅ Analysis viewable via CLI tool or logs
✅ Safety filters handled gracefully

### Output Example (M2)

```bash
2025/12/17 22:30:15 Fetched 500 queries from AdGuard Home
2025/12/17 22:30:15 [FIRST-SEEN] Client: 192.168.1.100 (192.168.1.100) | Domain: tracker.example.com | Type: A | Time: 2025-12-17 22:30:15
2025/12/17 22:30:15 [ANALYSIS] Queued anomaly for LLM analysis: tracker.example.com
2025/12/17 22:30:16 [WHOIS] Retrieved data for tracker.example.com (cached)
2025/12/17 22:30:18 [LLM] Analysis complete for tracker.example.com
  Classification: Suspicious
  Risk Score: 7/10
  Explanation: Third-party advertising tracker commonly used for behavioral targeting
  Suggested Action: Investigate
2025/12/17 22:30:18 Detected 1 new anomaly | Baseline: 15 clients, 523 domains
```

### Success Metrics (M2)
- LLM analysis success rate > 95%
- Average analysis time < 5 seconds
- WHOIS cache hit rate > 80%
- Zero polling disruptions from LLM analysis

---

## Milestone 3: Web Dashboard

### Deliverable
React UI with anomaly feed, LLM explanations, and Approve/Block actions.

### Implementation Steps

#### 3.1 REST API Backend
- [ ] Create HTTP server with routing
- [ ] Implement endpoints:
  - `GET /api/anomalies` - List pending anomalies
  - `GET /api/anomalies/:id` - Get specific anomaly
  - `POST /api/anomalies/:id/approve` - Approve (add to baseline)
  - `POST /api/anomalies/:id/block` - Block via AdGuard
  - `GET /api/stats` - Baseline statistics
  - `GET /api/settings` - Get current settings
  - `POST /api/settings` - Update settings
- [ ] Add CORS middleware
- [ ] Implement JSON serialization

#### 3.2 AdGuard Home Block API Integration
- [ ] Research AdGuard API for blocking domains
- [ ] Implement block functionality per client
- [ ] Handle API errors gracefully

#### 3.3 Frontend Setup
- [ ] Initialize Vite + React + TypeScript project
- [ ] Set up routing (React Router)
- [ ] Configure API client (axios/fetch)
- [ ] Add Tailwind CSS or Material-UI

#### 3.4 UI Components
- [ ] **Activity Feed**
  - Anomaly cards with device, domain, timestamp
  - LLM explanation display
  - Risk score badge (color-coded)
  - Classification badge

- [ ] **Action Buttons**
  - Approve button → calls `/approve` → removes from feed
  - Block button → calls `/block` → updates AdGuard
  - Confirmation dialogs

- [ ] **Settings Page**
  - AdGuard credentials form
  - LLM provider selection
  - API key inputs
  - Polling interval configuration

- [ ] **Stats Dashboard**
  - Total devices monitored
  - Baseline domain counts
  - Anomalies detected (today/week/all)

#### 3.5 State Management
- [ ] Set up React Context or Redux
- [ ] Implement real-time updates (polling or WebSocket)
- [ ] Handle loading and error states

### Dependencies (M3)
```bash
# Frontend
npm create vite@latest web -- --template react-ts
npm install react-router-dom axios
npm install -D tailwindcss postcss autoprefixer
```

### Checkpoint Criteria
✅ UI displays pending anomalies with LLM analysis
✅ "Approve" action persists to database
✅ "Block" action updates AdGuard Home
✅ Settings page saves configuration
✅ UI updates in real-time

---

## Milestone 4: Dockerization

### Deliverable
Complete Docker Compose setup for one-command deployment.

### Implementation Steps

#### 4.1 Backend Dockerfile
- [ ] Create multi-stage build:
  - Build stage: Go compilation
  - Runtime stage: Minimal Alpine image
- [ ] Configure health checks
- [ ] Set proper working directory

#### 4.2 Frontend Dockerfile
- [ ] Create multi-stage build:
  - Build stage: npm build
  - Runtime stage: nginx to serve static files
- [ ] Configure nginx for SPA routing

#### 4.3 Docker Compose Configuration
- [ ] Define services:
  - `guardian-log` (backend)
  - `guardian-web` (frontend)
  - Optional: `adguard-home` (for testing)
- [ ] Configure volumes:
  - BoltDB persistence
  - Configuration files
- [ ] Set up networking
- [ ] Define environment variables

#### 4.4 Documentation
- [ ] Update README with:
  - Quick start guide
  - Configuration instructions
  - Troubleshooting
- [ ] Create .env.example with all variables
- [ ] Add docker-compose.example.yml

### Checkpoint Criteria
✅ `docker-compose up` starts entire stack
✅ Data persists across container restarts
✅ Frontend accessible on configured port
✅ Backend API accessible from frontend

---

## Environment Variables Reference

```env
# AdGuard Home Configuration
AGH_URL=http://192.168.1.2:8080
AGH_USER=admin
AGH_PASS=password

# LLM Provider Selection
LLM_PROVIDER=ollama  # ollama | openai | anthropic | gemini

# LLM API Keys (only needed for cloud providers)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
GEMINI_API_KEY=...

# Ollama Configuration (for local)
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3

# Application Configuration
POLL_INTERVAL=10s
DB_PATH=/data/guardian.db
LOG_LEVEL=info
API_PORT=8080

# Frontend Configuration
VITE_API_URL=http://localhost:8080
```

---

## Development Workflow

### Phase 1: Milestone 1 (Current Focus)
1. Set up Go project structure
2. Implement AdGuard API client
3. Create BoltDB storage layer
4. Build polling orchestrator
5. Test with real AdGuard instance

### Phase 2: Milestone 2
1. Implement WHOIS enrichment
2. Create LLM provider interface
3. Implement all four providers
4. Test analysis pipeline

### Phase 3: Milestone 3
1. Build REST API
2. Create React frontend
3. Implement UI components
4. Integrate with backend

### Phase 4: Milestone 4
1. Create Dockerfiles
2. Configure Docker Compose
3. Test deployment
4. Write documentation

---

## Testing Strategy

### Unit Tests
- AdGuard API client
- BoltDB operations
- Baseline detection logic
- LLM response parsing

### Integration Tests
- End-to-end polling flow
- API endpoints
- Database persistence

### Manual Testing
- Real AdGuard Home integration
- LLM provider accuracy
- UI/UX functionality

---

## Future Enhancements (Post-M4)

- [ ] WebSocket support for real-time UI updates
- [ ] Advanced filtering and search
- [ ] Historical analysis and trends
- [ ] Email/webhook notifications
- [ ] Custom rule engine
- [ ] Multi-AdGuard instance support
- [ ] Export reports (PDF/CSV)
- [ ] Mobile-responsive design improvements

---

## Success Metrics

- **M1:** Application runs continuously without crashes
- **M2:** LLM accuracy > 90% for common domains
- **M3:** UI response time < 500ms for all operations
- **M4:** Deployment time < 5 minutes from clone to running

---

## Next Steps

Start with **Milestone 1** implementation. The first task is to create the basic project structure and AdGuard API client.
