package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/eiladin/guardian-log/internal/llm"
	"github.com/eiladin/guardian-log/internal/storage"
)

// AnalyzeBatch performs batch LLM analysis on multiple DNS queries using Gemini
func (p *Provider) AnalyzeBatch(ctx context.Context, queries []storage.DNSQuery, whoisData map[string]*storage.WHOISData) ([]*llm.Analysis, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("no queries to analyze")
	}

	// Create context with timeout (longer for batches)
	analyzeCtx, cancel := context.WithTimeout(ctx, p.timeout*2)
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

	// Build the batch prompt
	prompt := llm.BuildBatchPrompt(queries, whoisData)

	log.Printf("🚀 [Gemini] Analyzing batch of %d domains in single request", len(queries))

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
				log.Printf("⚠️  [Gemini] Batch rate limit exceeded after %d retries", MaxRetries)
				return nil, llm.ErrRateLimited
			}

			// Calculate backoff with exponential increase
			sleepDuration := backoff
			if sleepDuration > MaxBackoff {
				sleepDuration = MaxBackoff
			}

			log.Printf("⏳ [Gemini] Batch rate limited, retry %d/%d after %v",
				attempt+1, MaxRetries, sleepDuration)

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
		return nil, fmt.Errorf("Gemini batch API request failed: %w", err)
	}

	// Extract response text
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini for batch")
	}

	// Get the text from the first part
	var responseText string
	switch part := resp.Candidates[0].Content.Parts[0].(type) {
	case genai.Text:
		responseText = string(part)
	default:
		return nil, fmt.Errorf("unexpected response type from Gemini batch")
	}

	log.Printf("📥 [Gemini] Received batch response (%d bytes)", len(responseText))

	// Parse JSON array response
	var batchResponses []llm.BatchAnalysisResponse
	if err := json.Unmarshal([]byte(responseText), &batchResponses); err != nil {
		log.Printf("❌ [Gemini] Failed to parse batch JSON: %v", err)
		log.Printf("Response was: %s", responseText)
		return nil, fmt.Errorf("%w: %v", llm.ErrInvalidJSON, err)
	}

	// Validate we got responses for all queries
	if len(batchResponses) != len(queries) {
		log.Printf("⚠️  [Gemini] Expected %d responses, got %d", len(queries), len(batchResponses))
		return nil, fmt.Errorf("batch response count mismatch: expected %d, got %d",
			len(queries), len(batchResponses))
	}

	// Convert batch responses to Analysis objects
	analyses := make([]*llm.Analysis, len(queries))
	for i, batchResp := range batchResponses {
		// Validate the response
		if err := batchResp.Validate(); err != nil {
			log.Printf("⚠️  [Gemini] Batch response[%d] validation failed: %v", i, err)
			return nil, fmt.Errorf("batch response[%d] validation failed: %w", i, err)
		}

		// Ensure domain matches (responses should be in order)
		if batchResp.Domain != queries[i].Domain {
			log.Printf("⚠️  [Gemini] Domain mismatch at index %d: expected %s, got %s",
				i, queries[i].Domain, batchResp.Domain)
			return nil, fmt.Errorf("domain mismatch at index %d", i)
		}

		// Build Analysis object
		analyses[i] = &llm.Analysis{
			Domain:          queries[i].Domain,
			ClientID:        queries[i].ClientID,
			ClientName:      queries[i].ClientName,
			Classification:  batchResp.Classification,
			Explanation:     batchResp.Explanation,
			RiskScore:       batchResp.RiskScore,
			SuggestedAction: batchResp.SuggestedAction,
			AnalyzedAt:      time.Now(),
			Provider:        p.Name(),
			QueryType:       queries[i].QueryType,
		}

		log.Printf("  [%d/%d] %s -> %s (risk: %d/10)",
			i+1, len(queries), analyses[i].Domain, analyses[i].Classification, analyses[i].RiskScore)
	}

	log.Printf("✅ [Gemini] Batch analysis complete: %d domains analyzed in single request", len(analyses))

	return analyses, nil
}
