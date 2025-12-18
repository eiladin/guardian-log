package ingestor

import (
	"context"
	"log"
	"time"

	"github.com/eiladin/guardian-log/internal/analyzer"
)

// LLMAnalyzer defines the interface for LLM analysis
type LLMAnalyzer interface {
	AnalyzeAsync(query interface{})
	GetStats() map[string]interface{}
	Stop()
}

// Poller orchestrates the polling and analysis of AdGuard Home query logs
type Poller struct {
	client      *AdGuardClient
	analyzer    *analyzer.BaselineAnalyzer
	llmAnalyzer LLMAnalyzer // Optional LLM analyzer
	interval    time.Duration
}

// NewPoller creates a new poller instance
func NewPoller(client *AdGuardClient, analyzer *analyzer.BaselineAnalyzer, interval time.Duration) *Poller {
	return &Poller{
		client:   client,
		analyzer: analyzer,
		interval: interval,
	}
}

// SetLLMAnalyzer sets the optional LLM analyzer
func (p *Poller) SetLLMAnalyzer(llmAnalyzer LLMAnalyzer) {
	p.llmAnalyzer = llmAnalyzer
}

// Start begins the polling loop
func (p *Poller) Start(ctx context.Context) error {
	log.Printf("Starting poller with interval: %s", p.interval)

	// Test connection first
	if err := p.client.TestConnection(); err != nil {
		return err
	}
	log.Println("Successfully connected to AdGuard Home")

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Run once immediately
	if err := p.poll(ctx); err != nil {
		log.Printf("Error during initial poll: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Poller stopped by context")
			return ctx.Err()
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				log.Printf("Error during poll: %v", err)
				// Continue polling even if there's an error
			}
		}
	}
}

// poll fetches and processes queries
func (p *Poller) poll(ctx context.Context) error {
	// Check context before processing
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Fetch recent queries from AdGuard Home
	queries, err := p.client.FetchQueryLog(100)
	if err != nil {
		return err
	}

	log.Printf("Fetched %d queries from AdGuard Home", len(queries))

	// Process each query
	anomalyCount := 0
	processedCount := 0
	skippedEmpty := 0

	for _, query := range queries {
		// Skip empty domains
		if query.Domain == "" {
			skippedEmpty++
			continue
		}

		processedCount++

		// Process the query
		isAnomaly, err := p.analyzer.ProcessQuery(query)
		if err != nil {
			log.Printf("Error processing query: %v", err)
			continue
		}

		// If it's an anomaly, log it and add to baseline
		if isAnomaly {
			p.analyzer.LogAnomaly(query)

			// If LLM analysis is enabled, queue for analysis
			if p.llmAnalyzer != nil {
				log.Printf("🤖 [LLM] Queuing domain for analysis: %s", query.Domain)
				p.llmAnalyzer.AnalyzeAsync(query)
			} else {
				log.Printf("⚠️  [LLM] Analyzer not initialized, skipping LLM analysis for: %s", query.Domain)
			}

			// Automatically add to baseline so it won't be flagged again
			if err := p.analyzer.ApproveAnomaly(query.ClientID, query.ClientName, query.Domain); err != nil {
				log.Printf("Error adding domain to baseline: %v", err)
			}

			anomalyCount++
		}
	}

	// Log summary if there were anomalies or skipped queries
	if anomalyCount > 0 {
		// Get updated baseline stats
		stats, err := p.GetStats()
		if err == nil {
			log.Printf("Detected %d new anomalies | Baseline: %d clients, %d domains",
				anomalyCount, stats["total_clients"], stats["total_domains"])
		}
	} else if skippedEmpty > 0 {
		log.Printf("No anomalies detected (%d queries processed, %d skipped)", processedCount, skippedEmpty)
	}

	return nil
}

// GetStats returns current baseline statistics
func (p *Poller) GetStats() (map[string]interface{}, error) {
	return p.analyzer.GetBaselineStats()
}
