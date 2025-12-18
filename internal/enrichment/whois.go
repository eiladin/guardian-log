package enrichment

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"

	"github.com/eiladin/guardian-log/internal/storage"
)

const (
	// WhoisCacheTTL is how long to cache WHOIS results (24 hours)
	WhoisCacheTTL = 24 * time.Hour

	// RateLimitDelay is the delay between WHOIS lookups to respect server policies
	RateLimitDelay = 2 * time.Second
)

// WHOISService handles domain enrichment via WHOIS lookups
type WHOISService struct {
	store       *storage.BoltStore
	lastLookup  time.Time
	lookupCount int
	cacheHits   int
	cacheMisses int
}

// NewWHOISService creates a new WHOIS enrichment service
func NewWHOISService(store *storage.BoltStore) *WHOISService {
	return &WHOISService{
		store: store,
	}
}

// Lookup performs a WHOIS lookup for the given domain, using cache when available
func (s *WHOISService) Lookup(domain string) (*storage.WHOISData, error) {
	// Normalize domain (remove any trailing dots, convert to lowercase)
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	// Check cache first
	if cached := s.getFromCache(domain); cached != nil {
		s.cacheHits++
		log.Printf("[WHOIS] Cache hit for %s (age: %s)", domain, time.Since(cached.LookedUpAt).Round(time.Minute))
		return cached, nil
	}

	s.cacheMisses++

	// Rate limiting: ensure we don't make requests too quickly
	if !s.lastLookup.IsZero() {
		elapsed := time.Since(s.lastLookup)
		if elapsed < RateLimitDelay {
			sleepTime := RateLimitDelay - elapsed
			log.Printf("[WHOIS] Rate limiting: sleeping for %s", sleepTime)
			time.Sleep(sleepTime)
		}
	}

	// Perform WHOIS lookup
	log.Printf("[WHOIS] Looking up domain: %s", domain)
	s.lastLookup = time.Now()
	s.lookupCount++

	rawWhois, err := whois.Whois(domain)
	if err != nil {
		// Return partial data on error
		log.Printf("[WHOIS] Lookup failed for %s: %v", domain, err)
		return &storage.WHOISData{
			Domain:     domain,
			LookedUpAt: time.Now(),
		}, nil
	}

	// Parse WHOIS response
	parsed, err := whoisparser.Parse(rawWhois)
	if err != nil {
		// Return partial data on parse error
		log.Printf("[WHOIS] Parse failed for %s: %v", domain, err)
		return &storage.WHOISData{
			Domain:     domain,
			LookedUpAt: time.Now(),
		}, nil
	}

	// Build WHOISData from parsed result
	data := &storage.WHOISData{
		Domain:     domain,
		Registrar:  parsed.Registrar.Name,
		LookedUpAt: time.Now(),
	}

	// Extract country from registrant or admin contact
	if parsed.Registrant.Country != "" {
		data.Country = parsed.Registrant.Country
	} else if parsed.Administrative.Country != "" {
		data.Country = parsed.Administrative.Country
	}

	// Extract dates
	if parsed.Domain.CreatedDate != "" {
		data.CreatedDate = parsed.Domain.CreatedDate
	}
	if parsed.Domain.UpdatedDate != "" {
		data.UpdatedDate = parsed.Domain.UpdatedDate
	}
	if parsed.Domain.ExpirationDate != "" {
		data.ExpiryDate = parsed.Domain.ExpirationDate
	}

	// Extract name servers
	data.NameServers = parsed.Domain.NameServers

	// Cache the result
	if err := s.saveToCache(data); err != nil {
		log.Printf("[WHOIS] Failed to cache result for %s: %v", domain, err)
	}

	log.Printf("[WHOIS] Successfully looked up %s (registrar: %s, country: %s)",
		domain, data.Registrar, data.Country)

	return data, nil
}

// getFromCache retrieves cached WHOIS data if available and not expired
func (s *WHOISService) getFromCache(domain string) *storage.WHOISData {
	rawData, err := s.store.GetCachedWHOIS(domain)
	if err != nil || rawData == nil {
		return nil
	}

	// Convert raw map to WHOISData
	jsonBytes, err := json.Marshal(rawData)
	if err != nil {
		return nil
	}

	var cached storage.WHOISData
	if err := json.Unmarshal(jsonBytes, &cached); err != nil {
		return nil
	}

	// Check if cache entry is expired
	if time.Since(cached.LookedUpAt) > WhoisCacheTTL {
		return nil
	}

	return &cached
}

// saveToCache stores WHOIS data in the cache
func (s *WHOISService) saveToCache(data *storage.WHOISData) error {
	return s.store.CacheWHOIS(data.Domain, data)
}

// GetStats returns statistics about WHOIS lookups
func (s *WHOISService) GetStats() map[string]interface{} {
	cacheHitRate := 0.0
	totalRequests := s.cacheHits + s.cacheMisses
	if totalRequests > 0 {
		cacheHitRate = float64(s.cacheHits) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"total_lookups":  s.lookupCount,
		"cache_hits":     s.cacheHits,
		"cache_misses":   s.cacheMisses,
		"cache_hit_rate": fmt.Sprintf("%.1f%%", cacheHitRate),
	}
}
