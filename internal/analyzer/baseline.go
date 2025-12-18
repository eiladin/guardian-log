package analyzer

import (
	"fmt"
	"log"
	"time"

	"github.com/eiladin/guardian-log/internal/storage"
)

// BaselineAnalyzer handles anomaly detection based on client baselines
type BaselineAnalyzer struct {
	store *storage.BoltStore
}

// NewBaselineAnalyzer creates a new baseline analyzer
func NewBaselineAnalyzer(store *storage.BoltStore) *BaselineAnalyzer {
	return &BaselineAnalyzer{
		store: store,
	}
}

// ProcessQuery analyzes a DNS query and determines if it's an anomaly
// Returns true if this is a first-seen (anomalous) query
func (a *BaselineAnalyzer) ProcessQuery(query storage.DNSQuery) (bool, error) {
	// Check if we've already processed this exact query
	queryID := query.QueryID()
	seen, err := a.store.HasSeenQuery(queryID)
	if err != nil {
		return false, fmt.Errorf("failed to check if query was seen: %w", err)
	}

	if seen {
		// Already processed this query, skip it
		return false, nil
	}

	// Mark query as processed
	if err := a.store.MarkQueryProcessed(queryID); err != nil {
		return false, fmt.Errorf("failed to mark query as processed: %w", err)
	}

	// Check if domain is in baseline
	inBaseline, err := a.store.HasDomainInBaseline(query.ClientID, query.Domain)
	if err != nil {
		return false, fmt.Errorf("failed to check baseline: %w", err)
	}

	if !inBaseline {
		// This is a first-seen domain for this client
		return true, nil
	}

	return false, nil
}

// LogAnomaly logs an anomaly event to stdout
func (a *BaselineAnalyzer) LogAnomaly(query storage.DNSQuery) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[FIRST-SEEN] Client: %s (%s) | Domain: %s | Type: %s | Time: %s",
		query.ClientName,
		query.ClientID,
		query.Domain,
		query.QueryType,
		timestamp,
	)
}

// ApproveAnomaly adds a domain to the client's baseline (for future use)
func (a *BaselineAnalyzer) ApproveAnomaly(clientID, clientName, domain string) error {
	return a.store.AddDomainToBaseline(clientID, clientName, domain)
}

// GetBaselineStats returns statistics about stored baselines
func (a *BaselineAnalyzer) GetBaselineStats() (map[string]interface{}, error) {
	baselines, err := a.store.GetAllBaselines()
	if err != nil {
		return nil, fmt.Errorf("failed to get baselines: %w", err)
	}

	totalDomains := 0
	for _, baseline := range baselines {
		totalDomains += len(baseline.Domains)
	}

	stats := map[string]interface{}{
		"total_clients": len(baselines),
		"total_domains": totalDomains,
		"clients":       baselines,
	}

	return stats, nil
}
