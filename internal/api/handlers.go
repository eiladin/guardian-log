package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/eiladin/guardian-log/internal/storage"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondError sends an error JSON response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}

// handleAnomalies handles GET /api/anomalies
func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get status filter from query params (optional)
	statusFilter := r.URL.Query().Get("status")

	// Get all anomalies from storage
	anomalies, err := s.store.GetAllAnomalies(statusFilter)
	if err != nil {
		log.Printf("Error retrieving anomalies: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve anomalies")
		return
	}

	// Convert to API response format
	response := make([]AnomalyResponse, 0, len(anomalies))
	for _, anomaly := range anomalies {
		response = append(response, AnomalyResponse{
			ID:              anomaly.ID,
			Domain:          anomaly.Domain,
			ClientID:        anomaly.ClientID,
			ClientName:      anomaly.ClientName,
			QueryType:       anomaly.QueryType,
			Classification:  anomaly.Classification,
			RiskScore:       anomaly.RiskScore,
			Explanation:     anomaly.Explanation,
			SuggestedAction: anomaly.SuggestedAction,
			DetectedAt:      anomaly.DetectedAt,
			Status:          anomaly.Status,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// handleAnomalyAction handles POST /api/anomalies/{id}/approve and /api/anomalies/{id}/block
func (s *Server) handleAnomalyAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract anomaly ID and action from URL path
	// Path format: /api/anomalies/{id}/{action}
	path := strings.TrimPrefix(r.URL.Path, "/api/anomalies/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		respondError(w, http.StatusBadRequest, "Invalid URL format. Expected: /api/anomalies/{id}/{action}")
		return
	}

	// URL decode the ID (in case it contains special characters)
	anomalyID, err := url.PathUnescape(parts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid anomaly ID")
		return
	}

	action := parts[1]

	// Validate action
	if action != "approve" && action != "block" {
		respondError(w, http.StatusBadRequest, "Invalid action. Must be 'approve' or 'block'")
		return
	}

	// Get the anomaly
	anomaly, err := s.store.GetAnomalyByID(anomalyID)
	if err != nil {
		log.Printf("Error retrieving anomaly %s: %v", anomalyID, err)
		respondError(w, http.StatusNotFound, "Anomaly not found")
		return
	}

	// Perform the action
	switch action {
	case "approve":
		if err := s.approveAnomaly(anomaly); err != nil {
			log.Printf("Error approving anomaly %s: %v", anomalyID, err)
			respondError(w, http.StatusInternalServerError, "Failed to approve anomaly")
			return
		}
		log.Printf("✅ Anomaly approved: %s (domain: %s, client: %s)", anomalyID, anomaly.Domain, anomaly.ClientID)

	case "block":
		if err := s.blockAnomaly(anomaly); err != nil {
			log.Printf("Error blocking anomaly %s: %v", anomalyID, err)
			respondError(w, http.StatusInternalServerError, "Failed to block anomaly")
			return
		}
		log.Printf("🚫 Anomaly blocked: %s (domain: %s, client: %s)", anomalyID, anomaly.Domain, anomaly.ClientID)
	}

	respondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Anomaly %s successfully", action+"d"),
	})
}

// approveAnomaly approves an anomaly by adding the domain to the baseline
func (s *Server) approveAnomaly(anomaly *storage.Anomaly) error {
	// Add domain to baseline
	if err := s.store.AddDomainToBaseline(anomaly.ClientID, anomaly.ClientName, anomaly.Domain); err != nil {
		return fmt.Errorf("failed to add domain to baseline: %w", err)
	}

	// Update anomaly status
	if err := s.store.UpdateAnomalyStatus(anomaly.ID, "approved"); err != nil {
		return fmt.Errorf("failed to update anomaly status: %w", err)
	}

	return nil
}

// blockAnomaly blocks an anomaly by calling AdGuard Home API
func (s *Server) blockAnomaly(anomaly *storage.Anomaly) error {
	// Call AdGuard Home API to add domain to blocklist
	if err := s.adguardClient.BlockDomain(anomaly.Domain); err != nil {
		return fmt.Errorf("failed to block domain in AdGuard Home: %w", err)
	}

	// Update anomaly status
	if err := s.store.UpdateAnomalyStatus(anomaly.ID, "blocked"); err != nil {
		return fmt.Errorf("failed to update anomaly status: %w", err)
	}

	return nil
}

// handleStats handles GET /api/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get stats from storage
	stats, err := s.store.GetStats()
	if err != nil {
		log.Printf("Error retrieving stats: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve statistics")
		return
	}

	// Get LLM analyzer stats if available
	var llmStats StatsResponse
	llmStats.TotalQueries = int64(stats["total_queries"].(int))
	llmStats.UniqueClients = stats["unique_clients"].(int)
	llmStats.TotalAnomalies = int64(stats["total_anomalies"].(int))
	llmStats.PendingAnomalies = stats["pending_anomalies"].(int)
	llmStats.ApprovedAnomalies = stats["approved_anomalies"].(int)
	llmStats.BlockedAnomalies = stats["blocked_anomalies"].(int)
	llmStats.MaliciousCount = stats["malicious_count"].(int)
	llmStats.SuspiciousCount = stats["suspicious_count"].(int)

	if s.llmAnalyzer != nil {
		analyzerStats := s.llmAnalyzer.GetStats()
		llmStats.LLMAnalysesTotal = int64(analyzerStats["total_analyses"].(int))
		llmStats.LLMAnalysesSuccess = int64(analyzerStats["successful_analyses"].(int))
		llmStats.LLMAnalysesFailed = int64(analyzerStats["failed_analyses"].(int))
	}

	respondJSON(w, http.StatusOK, llmStats)
}

// handleSettings handles GET and PUT /api/settings
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetSettings(w, r)
	case http.MethodPut:
		s.handleUpdateSettings(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetSettings handles GET /api/settings
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	response := SettingsResponse{
		AdGuardURL:   s.config.AdGuardURL,
		PollInterval: s.config.PollInterval.String(),
		LLMEnabled:   s.config.LLMEnabled,
		LLMProvider:  s.config.LLMProvider,
	}

	// Add provider-specific settings
	if s.config.LLMProvider == "gemini" {
		response.GeminiModel = s.config.GeminiModel
		response.HasGeminiAPIKey = s.config.GeminiAPIKey != ""
	}

	respondJSON(w, http.StatusOK, response)
}

// handleUpdateSettings handles PUT /api/settings
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// For now, return not implemented - actual implementation would require
	// reloading configuration and restarting components
	respondError(w, http.StatusNotImplemented, "Settings update not yet implemented")
}
