package llm

import (
	"context"
	"time"

	"github.com/eiladin/guardian-log/internal/storage"
)

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Analyze takes a DNS query and WHOIS data and returns security analysis
	Analyze(ctx context.Context, query storage.DNSQuery, whois *storage.WHOISData) (*Analysis, error)

	// AnalyzeBatch takes multiple queries and returns analyses for all in a single API call
	// Returns analyses in the same order as queries
	AnalyzeBatch(ctx context.Context, queries []storage.DNSQuery, whoisData map[string]*storage.WHOISData) ([]*Analysis, error)

	// Name returns the provider name (e.g., "gemini", "openai")
	Name() string

	// SupportsBatch returns true if the provider supports batch analysis
	SupportsBatch() bool
}

// Analysis represents the LLM's security analysis of a DNS query
type Analysis struct {
	// Domain being analyzed
	Domain string `json:"domain"`

	// Client that made the query
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`

	// LLM analysis results
	Classification  string `json:"classification"`   // Safe, Suspicious, Malicious
	Explanation     string `json:"explanation"`      // Human-readable explanation
	RiskScore       int    `json:"risk_score"`       // 1-10 (1=safe, 10=malicious)
	SuggestedAction string `json:"suggested_action"` // Allow, Investigate, Block

	// Metadata
	AnalyzedAt time.Time `json:"analyzed_at"`
	Provider   string    `json:"provider"` // Which LLM provider was used
	QueryType  string    `json:"query_type,omitempty"`
}

// LLMResponse represents the raw JSON response from an LLM
// This is what we expect the LLM to return
type LLMResponse struct {
	Classification  string `json:"classification"`
	Explanation     string `json:"explanation"`
	RiskScore       int    `json:"risk_score"`
	SuggestedAction string `json:"suggested_action"`
}

// Validate checks if the LLM response is valid
func (r *LLMResponse) Validate() error {
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
