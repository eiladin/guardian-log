package ingestor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eiladin/guardian-log/internal/storage"
)

// AdGuardClient handles communication with AdGuard Home API
type AdGuardClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

// QueryLogResponse represents the response from /control/querylog
type QueryLogResponse struct {
	Data   []QueryLogEntry `json:"data"`
	Oldest string          `json:"oldest"`
}

// QueryLogEntry represents a single entry in the AdGuard query log
type QueryLogEntry struct {
	Answer           []Answer               `json:"answer"`
	OriginalAnswer   []Answer               `json:"original_answer,omitempty"`
	AnswerDNSSEC     bool                   `json:"answer_dnssec,omitempty"`
	Cached           bool                   `json:"cached,omitempty"`
	Client           string                 `json:"client"`
	ClientID         string                 `json:"client_id,omitempty"`
	ClientInfo       map[string]interface{} `json:"client_info,omitempty"`
	ClientProto      string                 `json:"client_proto,omitempty"`
	ECS              string                 `json:"ecs,omitempty"`
	ElapsedMs        string                 `json:"elapsedMs"`
	Question         Question               `json:"question"`
	OriginalQuestion Question               `json:"original_question,omitempty"`
	Reason           string                 `json:"reason"`
	Rules            []interface{}          `json:"rules,omitempty"`
	ServiceName      string                 `json:"service_name,omitempty"`
	Status           string                 `json:"status"`
	Time             string                 `json:"time"`
	Upstream         string                 `json:"upstream,omitempty"`
}

// Question represents the DNS question
type Question struct {
	Class       string `json:"class"`
	Name        string `json:"name"`         // The domain name
	UnicodeName string `json:"unicode_name"` // Unicode version of the domain
	Type        string `json:"type"`
}

// Answer represents a DNS answer
type Answer struct {
	TTL   int    `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// NewAdGuardClient creates a new AdGuard Home API client
func NewAdGuardClient(baseURL, username, password string) *AdGuardClient {
	return &AdGuardClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchQueryLog retrieves recent DNS queries from AdGuard Home
func (c *AdGuardClient) FetchQueryLog(limit int) ([]storage.DNSQuery, error) {
	url := fmt.Sprintf("%s/control/querylog", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth
	req.SetBasicAuth(c.username, c.password)

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var queryLogResp QueryLogResponse
	if err := json.Unmarshal(respBytes, &queryLogResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our internal model
	queries := make([]storage.DNSQuery, 0, len(queryLogResp.Data))
	for _, entry := range queryLogResp.Data {
		query, err := c.convertToQuery(entry)
		if err != nil {
			// Skip queries that fail to convert
			continue
		}
		queries = append(queries, query)
	}

	return queries, nil
}

// convertToQuery converts AdGuard's query log entry to our internal model
func (c *AdGuardClient) convertToQuery(entry QueryLogEntry) (storage.DNSQuery, error) {
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, entry.Time)
	if err != nil {
		return storage.DNSQuery{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Extract answer value if available
	answerValue := ""
	if len(entry.Answer) > 0 {
		answerValue = entry.Answer[0].Value
	}

	// Determine ClientID and ClientName
	// ClientID is used as the unique identifier (for database keys)
	// ClientName is used for display
	clientID := entry.ClientID
	if clientID == "" {
		clientID = entry.Client // Use IP address if ClientID is not provided
	}

	clientName := entry.Client
	if clientName == "" {
		clientName = entry.ClientID
	}

	// Extract domain - use question.name (the correct field)
	domain := entry.Question.Name
	if domain == "" && entry.OriginalQuestion.Name != "" {
		domain = entry.OriginalQuestion.Name
	}

	// Use unicode name as fallback if available
	if domain == "" && entry.Question.UnicodeName != "" {
		domain = entry.Question.UnicodeName
	}

	return storage.DNSQuery{
		ClientID:   clientID,
		ClientName: clientName,
		Domain:     domain,
		Timestamp:  timestamp,
		QueryType:  entry.Question.Type,
		Answer:     answerValue,
		Reason:     entry.Reason,
	}, nil
}

// TestConnection verifies connectivity to AdGuard Home
func (c *AdGuardClient) TestConnection() error {
	url := fmt.Sprintf("%s/control/status", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to AdGuard Home: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status code %d", resp.StatusCode)
	}

	return nil
}

// BlockDomain adds a domain to the AdGuard Home blocklist
func (c *AdGuardClient) BlockDomain(domain string) error {
	// AdGuard Home API endpoint for adding filtering rules
	url := fmt.Sprintf("%s/control/filtering/add_url", c.baseURL)

	// Create the request payload to add a custom blocking rule
	// We'll use AdBlock syntax: ||domain^ blocks the domain and all subdomains
	payload := map[string]interface{}{
		"name":    fmt.Sprintf("Guardian-Log Block: %s", domain),
		"url":     fmt.Sprintf("data:text/plain,||%s^", domain),
		"enabled": true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to block domain: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to block domain, status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
