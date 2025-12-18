package llm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eiladin/guardian-log/internal/enrichment"
	"github.com/eiladin/guardian-log/internal/storage"
)

// Analyzer orchestrates LLM analysis of DNS queries
type Analyzer struct {
	provider     Provider
	whoisService *enrichment.WHOISService
	store        *storage.BoltStore

	// Async processing
	queryQueue   chan storage.DNSQuery
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc

	// Batching
	batchSize     int
	batchTimeout  time.Duration
	currentBatch  []storage.DNSQuery
	batchMu       sync.Mutex
	batchTimer    *time.Timer

	// Rate limiting
	rateLimiter  chan struct{} // Semaphore for rate limiting
	requestDelay time.Duration // Delay between requests

	// Statistics
	mu                 sync.Mutex
	totalAnalyses      int
	successfulAnalyses int
	failedAnalyses     int
	rateLimitedCount   int
	batchCount         int
}

// NewAnalyzer creates a new LLM analyzer
func NewAnalyzer(provider Provider, whoisService *enrichment.WHOISService, store *storage.BoltStore) *Analyzer {
	ctx, cancel := context.WithCancel(context.Background())

	// Configure batching
	// Process up to 10 domains per batch, or flush after 10 seconds
	batchSize := 10
	batchTimeout := 10 * time.Second

	// Configure rate limiting based on provider
	// With batching: 1 request per 10 seconds (6 RPM for batches of 10 = 60 domains/min)
	requestDelay := 10 * time.Second
	rateLimiterSize := 1 // Only allow 1 concurrent request

	analyzer := &Analyzer{
		provider:      provider,
		whoisService:  whoisService,
		store:         store,
		queryQueue:    make(chan storage.DNSQuery, 100),      // Buffer up to 100 queries
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
		currentBatch:  make([]storage.DNSQuery, 0, batchSize),
		rateLimiter:   make(chan struct{}, rateLimiterSize), // Semaphore for rate limiting
		requestDelay:  requestDelay,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start background worker
	analyzer.wg.Add(1)
	go analyzer.worker()

	log.Printf("📦 [Analyzer] Batch processing enabled: %d domains per batch, %v timeout",
		batchSize, batchTimeout)
	log.Printf("🚦 [Analyzer] Rate limiting: %v delay between batches, max %d concurrent",
		requestDelay, rateLimiterSize)

	return analyzer
}

// AnalyzeAsync queues a DNS query for asynchronous analysis
func (a *Analyzer) AnalyzeAsync(query interface{}) {
	// Type assert to DNSQuery
	dnsQuery, ok := query.(storage.DNSQuery)
	if !ok {
		log.Printf("[Analyzer] Invalid query type, expected storage.DNSQuery")
		return
	}

	select {
	case a.queryQueue <- dnsQuery:
		log.Printf("🤖 [Analyzer] Query queued successfully: %s (queue depth: %d)", dnsQuery.Domain, len(a.queryQueue))
	default:
		log.Printf("❌ [Analyzer] Queue full, dropping analysis for %s", dnsQuery.Domain)
	}
}

// worker processes queued queries in batches
func (a *Analyzer) worker() {
	defer a.wg.Done()

	// Initialize batch timer
	a.batchTimer = time.NewTimer(a.batchTimeout)
	defer a.batchTimer.Stop()

	for {
		select {
		case <-a.ctx.Done():
			// Flush any remaining batch before shutdown
			a.batchMu.Lock()
			if len(a.currentBatch) > 0 {
				log.Printf("📦 [Analyzer] Flushing final batch of %d queries on shutdown", len(a.currentBatch))
				a.processBatch(a.currentBatch)
				a.currentBatch = a.currentBatch[:0]
			}
			a.batchMu.Unlock()
			log.Println("[Analyzer] Shutting down worker")
			return

		case query := <-a.queryQueue:
			a.batchMu.Lock()
			a.currentBatch = append(a.currentBatch, query)
			batchFull := len(a.currentBatch) >= a.batchSize
			a.batchMu.Unlock()

			// If batch is full, process it immediately
			if batchFull {
				a.batchMu.Lock()
				batch := make([]storage.DNSQuery, len(a.currentBatch))
				copy(batch, a.currentBatch)
				a.currentBatch = a.currentBatch[:0]
				a.batchMu.Unlock()

				// Reset timer
				if !a.batchTimer.Stop() {
					select {
					case <-a.batchTimer.C:
					default:
					}
				}
				a.batchTimer.Reset(a.batchTimeout)

				log.Printf("📦 [Analyzer] Batch full (%d queries), processing now", len(batch))
				go a.processBatch(batch)
			}

		case <-a.batchTimer.C:
			// Timeout expired, process whatever we have
			a.batchMu.Lock()
			if len(a.currentBatch) > 0 {
				batch := make([]storage.DNSQuery, len(a.currentBatch))
				copy(batch, a.currentBatch)
				a.currentBatch = a.currentBatch[:0]
				a.batchMu.Unlock()

				log.Printf("⏰ [Analyzer] Batch timeout, processing %d queries", len(batch))
				go a.processBatch(batch)
			} else {
				a.batchMu.Unlock()
			}

			// Reset timer
			a.batchTimer.Reset(a.batchTimeout)
		}
	}
}

// analyzeQuery performs the full analysis workflow for a single query
func (a *Analyzer) analyzeQuery(query storage.DNSQuery) error {
	a.mu.Lock()
	a.totalAnalyses++
	a.mu.Unlock()

	log.Printf("🔍 [Analyzer] Starting analysis for %s (client: %s)", query.Domain, query.ClientID)

	// Step 1: WHOIS enrichment
	log.Printf("🌐 [Analyzer] Looking up WHOIS for %s", query.Domain)
	whoisData, err := a.whoisService.Lookup(query.Domain)
	if err != nil {
		log.Printf("⚠️  [Analyzer] WHOIS lookup failed for %s: %v", query.Domain, err)
		// Continue with partial data
	} else {
		log.Printf("✅ [Analyzer] WHOIS lookup complete for %s (registrar: %s, country: %s)",
			query.Domain, whoisData.Registrar, whoisData.Country)
	}

	// Step 2: LLM analysis with rate limiting
	// Acquire rate limiter token
	select {
	case a.rateLimiter <- struct{}{}:
		// Got token, proceed
		defer func() {
			// Release token after delay
			time.Sleep(a.requestDelay)
			<-a.rateLimiter
		}()
	case <-a.ctx.Done():
		return a.ctx.Err()
	}

	log.Printf("🤖 [Analyzer] Sending %s to %s for analysis", query.Domain, a.provider.Name())
	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	analysis, err := a.provider.Analyze(ctx, query, whoisData)
	if err != nil {
		// Check if it's a rate limit error
		if err == ErrRateLimited {
			log.Printf("🚫 [Analyzer] Rate limited for %s, will retry later", query.Domain)
			a.mu.Lock()
			a.rateLimitedCount++
			a.failedAnalyses++
			a.mu.Unlock()

			// Requeue for retry after a delay
			go func() {
				time.Sleep(30 * time.Second)
				a.AnalyzeAsync(query)
			}()
			return err
		}

		log.Printf("❌ [Analyzer] LLM analysis failed for %s: %v", query.Domain, err)
		a.mu.Lock()
		a.failedAnalyses++
		a.mu.Unlock()
		return err
	}

	log.Printf("✅ [Analyzer] LLM analysis complete for %s: %s (risk: %d/10)",
		query.Domain, analysis.Classification, analysis.RiskScore)

	// Step 3: Store analysis results
	if err := a.store.SaveAnalysis(analysis); err != nil {
		log.Printf("[Analyzer] Failed to save analysis for %s: %v", query.Domain, err)
		a.mu.Lock()
		a.failedAnalyses++
		a.mu.Unlock()
		return err
	}

	// Step 4: If suspicious or malicious, also store as anomaly
	if analysis.Classification == "Suspicious" || analysis.Classification == "Malicious" {
		anomaly := storage.Anomaly{
			Domain:          query.Domain,
			ClientID:        query.ClientID,
			ClientName:      query.ClientName,
			QueryType:       query.QueryType,
			Classification:  analysis.Classification,
			RiskScore:       analysis.RiskScore,
			Explanation:     analysis.Explanation,
			SuggestedAction: analysis.SuggestedAction,
			DetectedAt:      analysis.AnalyzedAt,
		}

		if err := a.store.SaveAnomaly(&anomaly); err != nil {
			log.Printf("[Analyzer] Failed to save anomaly for %s: %v", query.Domain, err)
		} else {
			log.Printf("[Analyzer] 🚨 ANOMALY DETECTED: %s -> %s (risk: %d/10, action: %s)",
				query.Domain, analysis.Classification, analysis.RiskScore, analysis.SuggestedAction)
		}
	}

	a.mu.Lock()
	a.successfulAnalyses++
	a.mu.Unlock()

	return nil
}

// GetStats returns statistics about the analyzer
func (a *Analyzer) GetStats() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()

	successRate := 0.0
	if a.totalAnalyses > 0 {
		successRate = float64(a.successfulAnalyses) / float64(a.totalAnalyses) * 100
	}

	avgBatchSize := 0.0
	if a.batchCount > 0 {
		avgBatchSize = float64(a.totalAnalyses) / float64(a.batchCount)
	}

	stats := map[string]interface{}{
		"total_analyses":      a.totalAnalyses,
		"successful_analyses": a.successfulAnalyses,
		"failed_analyses":     a.failedAnalyses,
		"rate_limited_count":  a.rateLimitedCount,
		"batches_processed":   a.batchCount,
		"avg_batch_size":      fmt.Sprintf("%.1f", avgBatchSize),
		"success_rate":        successRate,
		"queue_depth":         len(a.queryQueue),
		"provider":            a.provider.Name(),
	}

	// Add WHOIS stats
	whoisStats := a.whoisService.GetStats()
	for k, v := range whoisStats {
		stats["whois_"+k] = v
	}

	return stats
}

// Stop gracefully shuts down the analyzer
func (a *Analyzer) Stop() {
	log.Println("[Analyzer] Stopping...")
	a.cancel()
	a.wg.Wait()
	log.Println("[Analyzer] Stopped")
}
