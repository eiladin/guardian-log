package llm

import "errors"

var (
	// ErrInvalidClassification is returned when the LLM returns an invalid classification
	ErrInvalidClassification = errors.New("invalid classification: must be Safe, Suspicious, or Malicious")

	// ErrMissingExplanation is returned when the LLM doesn't provide an explanation
	ErrMissingExplanation = errors.New("missing explanation in LLM response")

	// ErrInvalidRiskScore is returned when the risk score is out of range
	ErrInvalidRiskScore = errors.New("invalid risk score: must be between 1 and 10")

	// ErrInvalidAction is returned when the suggested action is invalid
	ErrInvalidAction = errors.New("invalid suggested action: must be Allow, Investigate, or Block")

	// ErrProviderNotFound is returned when the specified provider doesn't exist
	ErrProviderNotFound = errors.New("LLM provider not found")

	// ErrInvalidJSON is returned when the LLM response is not valid JSON
	ErrInvalidJSON = errors.New("LLM response is not valid JSON")

	// ErrTimeout is returned when an LLM request times out
	ErrTimeout = errors.New("LLM request timed out")

	// ErrRateLimited is returned when the LLM API rate limit is exceeded
	ErrRateLimited = errors.New("LLM API rate limit exceeded")
)
