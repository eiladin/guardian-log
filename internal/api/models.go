package api

import "time"

// AnomalyResponse represents an anomaly in API responses
type AnomalyResponse struct {
	ID              string    `json:"id"`
	Domain          string    `json:"domain"`
	ClientID        string    `json:"client_id"`
	ClientName      string    `json:"client_name"`
	QueryType       string    `json:"query_type"`
	Classification  string    `json:"classification"`
	RiskScore       int       `json:"risk_score"`
	Explanation     string    `json:"explanation"`
	SuggestedAction string    `json:"suggested_action"`
	DetectedAt      time.Time `json:"detected_at"`
	Status          string    `json:"status"` // pending, approved, blocked
}

// StatsResponse represents system statistics
type StatsResponse struct {
	TotalQueries       int64 `json:"total_queries"`
	UniqueClients      int   `json:"unique_clients"`
	TotalAnomalies     int64 `json:"total_anomalies"`
	PendingAnomalies   int   `json:"pending_anomalies"`
	ApprovedAnomalies  int   `json:"approved_anomalies"`
	BlockedAnomalies   int   `json:"blocked_anomalies"`
	MaliciousCount     int   `json:"malicious_count"`
	SuspiciousCount    int   `json:"suspicious_count"`
	LLMAnalysesTotal   int64 `json:"llm_analyses_total"`
	LLMAnalysesSuccess int64 `json:"llm_analyses_success"`
	LLMAnalysesFailed  int64 `json:"llm_analyses_failed"`
}

// SettingsResponse represents current settings (with sensitive data redacted)
type SettingsResponse struct {
	AdGuardURL      string `json:"adguard_url"`
	PollInterval    string `json:"poll_interval"`
	LLMEnabled      bool   `json:"llm_enabled"`
	LLMProvider     string `json:"llm_provider"`
	GeminiModel     string `json:"gemini_model,omitempty"`
	HasGeminiAPIKey bool   `json:"has_gemini_api_key,omitempty"`
}

// UpdateSettingsRequest represents a settings update request
type UpdateSettingsRequest struct {
	AdGuardURL      *string `json:"adguard_url,omitempty"`
	AdGuardUser     *string `json:"adguard_user,omitempty"`
	AdGuardPassword *string `json:"adguard_password,omitempty"`
	LLMEnabled      *bool   `json:"llm_enabled,omitempty"`
	LLMProvider     *string `json:"llm_provider,omitempty"`
	GeminiAPIKey    *string `json:"gemini_api_key,omitempty"`
	GeminiModel     *string `json:"gemini_model,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
