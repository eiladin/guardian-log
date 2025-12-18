package llm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eiladin/guardian-log/internal/storage"
)

// BatchAnalysisResponse represents a single domain analysis in a batch response
type BatchAnalysisResponse struct {
	Domain          string `json:"domain"`
	Classification  string `json:"classification"`
	Explanation     string `json:"explanation"`
	RiskScore       int    `json:"risk_score"`
	SuggestedAction string `json:"suggested_action"`
}

// Validate checks if a batch analysis response is valid
func (r *BatchAnalysisResponse) Validate() error {
	if r.Domain == "" {
		return fmt.Errorf("missing domain in batch response")
	}
	if r.Classification != "Safe" && r.Classification != "Suspicious" && r.Classification != "Malicious" {
		return ErrInvalidClassification
	}
	if r.Explanation == "" {
		return ErrMissingExplanation
	}
	if r.RiskScore < 1 || r.RiskScore > 10 {
		return ErrInvalidRiskScore
	}
	if r.SuggestedAction != "Allow" && r.SuggestedAction != "Investigate" && r.SuggestedAction != "Block" {
		return ErrInvalidAction
	}
	return nil
}

// processBatch processes a batch of queries in a single LLM request
func (a *Analyzer) processBatch(queries []storage.DNSQuery) {
	if len(queries) == 0 {
		return
	}

	a.mu.Lock()
	a.batchCount++
	batchNum := a.batchCount
	a.mu.Unlock()

	log.Printf("📦 [Batch #%d] Processing %d domains", batchNum, len(queries))

	// Step 1: Gather WHOIS data for all domains
	whoisData := make(map[string]*storage.WHOISData)
	for _, query := range queries {
		whois, err := a.whoisService.Lookup(query.Domain)
		if err != nil {
			log.Printf("⚠️  [Batch #%d] WHOIS lookup failed for %s: %v", batchNum, query.Domain, err)
		}
		whoisData[query.Domain] = whois
	}

	// Step 2: Rate limiting for the entire batch
	select {
	case a.rateLimiter <- struct{}{}:
		defer func() {
			time.Sleep(a.requestDelay)
			<-a.rateLimiter
		}()
	case <-a.ctx.Done():
		return
	}

	ctx, cancel := context.WithTimeout(a.ctx, 90*time.Second) // Longer timeout for batches
	defer cancel()

	// Step 3: Check if provider supports true batch processing
	if a.provider.SupportsBatch() {
		log.Printf("🚀 [Batch #%d] Using true batch API call for %d domains", batchNum, len(queries))
		a.processBatchWithAPI(ctx, batchNum, queries, whoisData)
		return
	}

	// Fallback: Process batch sequentially
	log.Printf("📋 [Batch #%d] Provider doesn't support batching, processing %d domains sequentially", batchNum, len(queries))

	successCount := 0
	failCount := 0

	for i, query := range queries {
		log.Printf("  [%d/%d] Analyzing %s", i+1, len(queries), query.Domain)

		a.mu.Lock()
		a.totalAnalyses++
		a.mu.Unlock()

		// Use existing single-query analysis
		whois := whoisData[query.Domain]
		analysis, err := a.provider.Analyze(ctx, query, whois)

		if err != nil {
			if err == ErrRateLimited {
				log.Printf("🚫 [Batch #%d] Rate limited, stopping batch processing", batchNum)
				a.mu.Lock()
				a.rateLimitedCount++
				a.failedAnalyses += (len(queries) - i)
				a.mu.Unlock()

				// Requeue remaining queries
				for j := i; j < len(queries); j++ {
					go func(q storage.DNSQuery) {
						time.Sleep(30 * time.Second)
						a.AnalyzeAsync(q)
					}(queries[j])
				}
				return
			}

			log.Printf("❌ Failed to analyze %s: %v", query.Domain, err)
			a.mu.Lock()
			a.failedAnalyses++
			a.mu.Unlock()
			failCount++
			continue
		}

		// Save analysis
		if err := a.store.SaveAnalysis(analysis); err != nil {
			log.Printf("⚠️  Failed to save analysis for %s: %v", query.Domain, err)
		}

		// Save as anomaly if suspicious/malicious
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
				log.Printf("⚠️  Failed to save anomaly for %s: %v", query.Domain, err)
			} else {
				log.Printf("🚨 ANOMALY: %s -> %s (risk: %d/10)",
					query.Domain, analysis.Classification, analysis.RiskScore)
			}
		}

		a.mu.Lock()
		a.successfulAnalyses++
		a.mu.Unlock()
		successCount++
	}

	log.Printf("✅ [Batch #%d] Complete: %d succeeded, %d failed", batchNum, successCount, failCount)
}

// processBatchWithAPI processes a batch using the provider's native batch API
func (a *Analyzer) processBatchWithAPI(ctx context.Context, batchNum int, queries []storage.DNSQuery, whoisData map[string]*storage.WHOISData) {
	// Update analysis count
	a.mu.Lock()
	a.totalAnalyses += len(queries)
	a.mu.Unlock()

	// Call provider's batch method
	analyses, err := a.provider.AnalyzeBatch(ctx, queries, whoisData)
	if err != nil {
		// Handle rate limiting
		if err == ErrRateLimited {
			log.Printf("🚫 [Batch #%d] Rate limited, will retry all %d domains later", batchNum, len(queries))
			a.mu.Lock()
			a.rateLimitedCount++
			a.failedAnalyses += len(queries)
			a.mu.Unlock()

			// Requeue all queries
			for _, query := range queries {
				go func(q storage.DNSQuery) {
					time.Sleep(30 * time.Second)
					a.AnalyzeAsync(q)
				}(query)
			}
			return
		}

		// Other errors - log and count as failed
		log.Printf("❌ [Batch #%d] Batch analysis failed: %v", batchNum, err)
		a.mu.Lock()
		a.failedAnalyses += len(queries)
		a.mu.Unlock()
		return
	}

	// Process successful analyses
	successCount := 0
	failCount := 0

	for i, analysis := range analyses {
		if analysis == nil {
			log.Printf("⚠️  [Batch #%d] Nil analysis for query %d", batchNum, i)
			failCount++
			continue
		}

		// Save analysis
		if err := a.store.SaveAnalysis(analysis); err != nil {
			log.Printf("⚠️  [Batch #%d] Failed to save analysis for %s: %v", batchNum, analysis.Domain, err)
			failCount++
			continue
		}

		// Save as anomaly if suspicious/malicious
		if analysis.Classification == "Suspicious" || analysis.Classification == "Malicious" {
			anomaly := storage.Anomaly{
				Domain:          analysis.Domain,
				ClientID:        analysis.ClientID,
				ClientName:      analysis.ClientName,
				QueryType:       analysis.QueryType,
				Classification:  analysis.Classification,
				RiskScore:       analysis.RiskScore,
				Explanation:     analysis.Explanation,
				SuggestedAction: analysis.SuggestedAction,
				DetectedAt:      analysis.AnalyzedAt,
			}

			if err := a.store.SaveAnomaly(&anomaly); err != nil {
				log.Printf("⚠️  [Batch #%d] Failed to save anomaly for %s: %v", batchNum, analysis.Domain, err)
			} else {
				log.Printf("🚨 ANOMALY: %s -> %s (risk: %d/10)",
					analysis.Domain, analysis.Classification, analysis.RiskScore)
			}
		}

		successCount++
	}

	// Update statistics
	a.mu.Lock()
	a.successfulAnalyses += successCount
	a.failedAnalyses += failCount
	a.mu.Unlock()

	log.Printf("✅ [Batch #%d] Complete: %d succeeded, %d failed (single API call)", batchNum, successCount, failCount)
}
