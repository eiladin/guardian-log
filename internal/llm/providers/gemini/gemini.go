package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/eiladin/guardian-log/internal/llm"
	"github.com/eiladin/guardian-log/internal/storage"
)

const (
	// MaxRetries is the maximum number of retry attempts for rate-limited requests
	MaxRetries = 3

	// InitialBackoff is the initial backoff duration for retries
	InitialBackoff = 1 * time.Second

	// MaxBackoff is the maximum backoff duration
	MaxBackoff = 30 * time.Second
)

// Provider implements the LLM Provider interface for Google Gemini
type Provider struct {
	apiKey  string
	model   string
	timeout time.Duration
}

// NewProvider creates a new Gemini provider
func NewProvider(apiKey, model string, timeout time.Duration) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if model == "" {
		model = "gemini-1.5-flash" // Default to Flash model
	}

	return &Provider{
		apiKey:  apiKey,
		model:   model,
		timeout: timeout,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "gemini"
}

// SupportsBatch returns true indicating Gemini supports batch analysis
func (p *Provider) SupportsBatch() bool {
	return true
}

// isRateLimitError checks if an error is a rate limit (429) error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check for gRPC status code RESOURCE_EXHAUSTED
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.ResourceExhausted {
			return true
		}
	}

	// Check error message for common rate limit indicators
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "429") ||
		strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "quota exceeded") ||
		strings.Contains(errMsg, "resource_exhausted")
}

// Analyze performs LLM analysis on a DNS query using Gemini with retry logic
func (p *Provider) Analyze(ctx context.Context, query storage.DNSQuery, whois *storage.WHOISData) (*llm.Analysis, error) {
	// Create context with timeout
	analyzeCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Initialize Gemini client
	client, err := genai.NewClient(analyzeCtx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Get the model
	model := client.GenerativeModel(p.model)

	// Configure model for JSON output
	model.ResponseMIMEType = "application/json"

	// Set temperature to 0 for consistent, deterministic responses
	temp := float32(0.0)
	model.Temperature = &temp

	// Build the prompt
	prompt := llm.BuildPrompt(query, whois)

	log.Printf("[Gemini] Analyzing domain: %s (client: %s)", query.Domain, query.ClientID)

	// Retry loop with exponential backoff
	var resp *genai.GenerateContentResponse
	backoff := InitialBackoff

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		// Generate content
		resp, err = model.GenerateContent(analyzeCtx, genai.Text(prompt))

		// Success - break out of retry loop
		if err == nil {
			break
		}

		// Check for timeout
		if analyzeCtx.Err() == context.DeadlineExceeded {
			return nil, llm.ErrTimeout
		}

		// Check if this is a rate limit error
		if isRateLimitError(err) {
			// If we've exhausted retries, return rate limit error
			if attempt == MaxRetries {
				log.Printf("⚠️  [Gemini] Rate limit exceeded after %d retries for %s", MaxRetries, query.Domain)
				return nil, llm.ErrRateLimited
			}

			// Calculate backoff with exponential increase
			sleepDuration := backoff
			if sleepDuration > MaxBackoff {
				sleepDuration = MaxBackoff
			}

			log.Printf("⏳ [Gemini] Rate limited, retry %d/%d after %v for %s",
				attempt+1, MaxRetries, sleepDuration, query.Domain)

			// Sleep for backoff duration
			select {
			case <-time.After(sleepDuration):
				// Continue to next retry
			case <-analyzeCtx.Done():
				return nil, llm.ErrTimeout
			}

			// Double the backoff for next retry (exponential backoff)
			backoff *= 2

			continue
		}

		// Non-rate-limit error - fail immediately
		return nil, fmt.Errorf("Gemini API request failed: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Get the text from the first part
	var responseText string
	switch part := resp.Candidates[0].Content.Parts[0].(type) {
	case genai.Text:
		responseText = string(part)
	default:
		return nil, fmt.Errorf("unexpected response type from Gemini")
	}

	log.Printf("[Gemini] Raw response: %s", responseText)

	// Parse JSON response
	var llmResp llm.LLMResponse
	if err := json.Unmarshal([]byte(responseText), &llmResp); err != nil {
		log.Printf("[Gemini] Failed to parse JSON: %v", err)
		return nil, fmt.Errorf("%w: %v", llm.ErrInvalidJSON, err)
	}

	// Validate response
	if err := llmResp.Validate(); err != nil {
		log.Printf("[Gemini] Response validation failed: %v", err)
		return nil, err
	}

	// Build Analysis result
	analysis := &llm.Analysis{
		Domain:          query.Domain,
		ClientID:        query.ClientID,
		ClientName:      query.ClientName,
		Classification:  llmResp.Classification,
		Explanation:     llmResp.Explanation,
		RiskScore:       llmResp.RiskScore,
		SuggestedAction: llmResp.SuggestedAction,
		AnalyzedAt:      time.Now(),
		Provider:        p.Name(),
		QueryType:       query.QueryType,
	}

	log.Printf("[Gemini] Analysis complete: %s -> %s (risk: %d/10, action: %s)",
		query.Domain, analysis.Classification, analysis.RiskScore, analysis.SuggestedAction)

	return analysis, nil
}
