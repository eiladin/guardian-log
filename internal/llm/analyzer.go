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
	queryQueue chan storage.DNSQuery
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc

	// Batching
	batchSize    int
	batchTimeout time.Duration
	currentBatch []storage.DNSQuery
	batchMu      sync.Mutex
	batchTimer   *time.Timer

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
func NewAnalyzer(provider Provider, whoisService *enrichment.WHOISService, store *storage.BoltStore, batchSize int, batchTimeout, requestDelay time.Duration) *Analyzer {
	ctx, cancel := context.WithCancel(context.Background())

	// Validate batch settings
	if batchSize <= 0 {
		batchSize = 20 // Default
	}
	if batchTimeout <= 0 {
		batchTimeout = 60 * time.Second // Default
	}
	if requestDelay <= 0 {
		requestDelay = 60 * time.Second // Default
	}

	rateLimiterSize := 1 // Only allow 1 concurrent batch request

	analyzer := &Analyzer{
		provider:     provider,
		whoisService: whoisService,
		store:        store,
		queryQueue:   make(chan storage.DNSQuery, 100), // Buffer up to 100 queries
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		currentBatch: make([]storage.DNSQuery, 0, batchSize),
		rateLimiter:  make(chan struct{}, rateLimiterSize), // Semaphore for rate limiting
		requestDelay: requestDelay,
		ctx:          ctx,
		cancel:       cancel,
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
