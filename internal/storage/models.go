package storage

import (
	"time"
)

// DNSQuery represents a single DNS query from AdGuard Home
type DNSQuery struct {
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	Domain     string    `json:"domain"`
	Timestamp  time.Time `json:"timestamp"`
	QueryType  string    `json:"query_type"`
	Answer     string    `json:"answer,omitempty"`
	Reason     string    `json:"reason,omitempty"` // AdGuard's filtering reason
	Response   string    `json:"response,omitempty"`
	Upstream   string    `json:"upstream,omitempty"`
}

// Baseline represents the known domains for a specific client
type Baseline struct {
	ClientID    string    `json:"client_id"`
	ClientName  string    `json:"client_name"`
	Domains     []string  `json:"domains"`
	LastUpdated time.Time `json:"last_updated"`
}

// AnomalyEvent represents a first-seen domain for a client
type AnomalyEvent struct {
	Query      DNSQuery  `json:"query"`
	DetectedAt time.Time `json:"detected_at"`
}

// Anomaly represents a detected security threat from LLM analysis
type Anomaly struct {
	ID              string    `json:"id,omitempty"`
	Domain          string    `json:"domain"`
	ClientID        string    `json:"client_id"`
	ClientName      string    `json:"client_name"`
	QueryType       string    `json:"query_type"`
	Classification  string    `json:"classification"` // Suspicious or Malicious
	RiskScore       int       `json:"risk_score"`     // 1-10
	Explanation     string    `json:"explanation"`
	SuggestedAction string    `json:"suggested_action"` // Investigate or Block
	DetectedAt      time.Time `json:"detected_at"`
	Status          string    `json:"status"` // pending, approved, blocked
}

// WHOISData contains enrichment information about a domain
type WHOISData struct {
	Domain      string    `json:"domain"`
	Registrar   string    `json:"registrar,omitempty"`
	Country     string    `json:"country,omitempty"`
	CreatedDate string    `json:"created_date,omitempty"`
	UpdatedDate string    `json:"updated_date,omitempty"`
	ExpiryDate  string    `json:"expiry_date,omitempty"`
	NameServers []string  `json:"name_servers,omitempty"`
	LookedUpAt  time.Time `json:"looked_up_at"`
}

// QueryID generates a unique ID for deduplication
func (q *DNSQuery) QueryID() string {
	return q.ClientID + "|" + q.Domain + "|" + q.Timestamp.Format(time.RFC3339)
}
