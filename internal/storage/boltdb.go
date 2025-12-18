package storage

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	baselineBucket         = []byte("baselines")
	processedQueriesBucket = []byte("processed_queries")
	whoisCacheBucket       = []byte("whois_cache")
	anomaliesBucket        = []byte("anomalies")
	analysesBucket         = []byte("analyses")
)

// BoltStore provides persistent storage using BoltDB
type BoltStore struct {
	db *bolt.DB
}

// NewBoltStore creates a new BoltDB storage instance
func NewBoltStore(dbPath string) (*BoltStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := ensureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			baselineBucket,
			processedQueriesBucket,
			whoisCacheBucket,
			anomaliesBucket,
			analysesBucket,
		}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", string(bucket), err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &BoltStore{db: db}, nil
}

// Close closes the database connection
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// GetClientBaseline retrieves the baseline for a specific client
func (s *BoltStore) GetClientBaseline(clientID string) (*Baseline, error) {
	var baseline *Baseline

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(baselineBucket)
		data := b.Get([]byte(clientID))

		if data == nil {
			// No baseline exists yet for this client
			baseline = &Baseline{
				ClientID:    clientID,
				Domains:     []string{},
				LastUpdated: time.Now(),
			}
			return nil
		}

		baseline = &Baseline{}
		if err := json.Unmarshal(data, baseline); err != nil {
			return fmt.Errorf("failed to unmarshal baseline: %w", err)
		}
		return nil
	})

	return baseline, err
}

// AddDomainToBaseline adds a domain to a client's baseline
func (s *BoltStore) AddDomainToBaseline(clientID, clientName, domain string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(baselineBucket)

		// Get existing baseline
		var baseline Baseline
		data := b.Get([]byte(clientID))

		if data != nil {
			if err := json.Unmarshal(data, &baseline); err != nil {
				return fmt.Errorf("failed to unmarshal baseline: %w", err)
			}
		} else {
			baseline = Baseline{
				ClientID:   clientID,
				ClientName: clientName,
				Domains:    []string{},
			}
		}

		// Check if domain already exists
		for _, d := range baseline.Domains {
			if d == domain {
				return nil // Already exists
			}
		}

		// Add domain
		baseline.Domains = append(baseline.Domains, domain)
		baseline.LastUpdated = time.Now()

		// Save back to database
		encoded, err := json.Marshal(baseline)
		if err != nil {
			return fmt.Errorf("failed to marshal baseline: %w", err)
		}

		return b.Put([]byte(clientID), encoded)
	})
}

// HasSeenQuery checks if a query has been processed before
func (s *BoltStore) HasSeenQuery(queryID string) (bool, error) {
	var exists bool

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(processedQueriesBucket)
		data := b.Get([]byte(queryID))
		exists = data != nil
		return nil
	})

	return exists, err
}

// MarkQueryProcessed marks a query as processed
func (s *BoltStore) MarkQueryProcessed(queryID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(processedQueriesBucket)
		// Store timestamp when query was processed
		timestamp := []byte(time.Now().Format(time.RFC3339))
		return b.Put([]byte(queryID), timestamp)
	})
}

// HasDomainInBaseline checks if a domain exists in a client's baseline
func (s *BoltStore) HasDomainInBaseline(clientID, domain string) (bool, error) {
	baseline, err := s.GetClientBaseline(clientID)
	if err != nil {
		return false, err
	}

	for _, d := range baseline.Domains {
		if d == domain {
			return true, nil
		}
	}

	return false, nil
}

// GetAllBaselines retrieves all client baselines
func (s *BoltStore) GetAllBaselines() ([]Baseline, error) {
	var baselines []Baseline

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(baselineBucket)

		return b.ForEach(func(k, v []byte) error {
			var baseline Baseline
			if err := json.Unmarshal(v, &baseline); err != nil {
				return fmt.Errorf("failed to unmarshal baseline: %w", err)
			}
			baselines = append(baselines, baseline)
			return nil
		})
	})

	return baselines, err
}

// CleanOldProcessedQueries removes processed query records older than the specified duration
func (s *BoltStore) CleanOldProcessedQueries(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(processedQueriesBucket)

		var keysToDelete [][]byte

		err := b.ForEach(func(k, v []byte) error {
			timestamp, err := time.Parse(time.RFC3339, string(v))
			if err != nil {
				// If we can't parse, mark for deletion
				keysToDelete = append(keysToDelete, k)
				return nil
			}

			if timestamp.Before(cutoff) {
				keysToDelete = append(keysToDelete, k)
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Delete old entries
		for _, key := range keysToDelete {
			if err := b.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
}

// CacheWHOIS stores WHOIS data in the cache
func (s *BoltStore) CacheWHOIS(domain string, data interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(whoisCacheBucket)

		encoded, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal WHOIS data: %w", err)
		}

		return b.Put([]byte(domain), encoded)
	})
}

// GetCachedWHOIS retrieves cached WHOIS data for a domain
func (s *BoltStore) GetCachedWHOIS(domain string) (interface{}, error) {
	var data map[string]interface{}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(whoisCacheBucket)
		cached := b.Get([]byte(domain))

		if cached == nil {
			return nil
		}

		return json.Unmarshal(cached, &data)
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	// Import cycle workaround: return raw data and let caller unmarshal
	return data, nil
}

// SaveAnalysis stores an LLM analysis result
// Note: This function takes interface{} to avoid import cycles with internal/llm
func (s *BoltStore) SaveAnalysis(analysis interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(analysesBucket)

		// Generate a unique ID based on domain and timestamp
		encoded, err := json.Marshal(analysis)
		if err != nil {
			return fmt.Errorf("failed to marshal analysis: %w", err)
		}

		// Use timestamp-based key for chronological ordering
		key := []byte(time.Now().Format(time.RFC3339Nano))
		return b.Put(key, encoded)
	})
}

// SaveAnomaly stores a detected anomaly
func (s *BoltStore) SaveAnomaly(anomaly *Anomaly) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(anomaliesBucket)

		// Generate ID if not set
		if anomaly.ID == "" {
			anomaly.ID = fmt.Sprintf("%s|%s|%s",
				anomaly.ClientID,
				anomaly.Domain,
				anomaly.DetectedAt.Format(time.RFC3339))
		}

		// Set default status if not set
		if anomaly.Status == "" {
			anomaly.Status = "pending"
		}

		encoded, err := json.Marshal(anomaly)
		if err != nil {
			return fmt.Errorf("failed to marshal anomaly: %w", err)
		}

		return b.Put([]byte(anomaly.ID), encoded)
	})
}

// GetRecentAnomalies retrieves anomalies detected within the specified duration
func (s *BoltStore) GetRecentAnomalies(since time.Duration) ([]Anomaly, error) {
	var anomalies []Anomaly
	cutoff := time.Now().Add(-since)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(anomaliesBucket)

		return b.ForEach(func(k, v []byte) error {
			var anomaly Anomaly
			if err := json.Unmarshal(v, &anomaly); err != nil {
				return fmt.Errorf("failed to unmarshal anomaly: %w", err)
			}

			if anomaly.DetectedAt.After(cutoff) {
				anomalies = append(anomalies, anomaly)
			}
			return nil
		})
	})

	return anomalies, err
}

// GetAllAnomalies retrieves all anomalies, optionally filtered by status
func (s *BoltStore) GetAllAnomalies(statusFilter string) ([]Anomaly, error) {
	var anomalies []Anomaly

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(anomaliesBucket)

		return b.ForEach(func(k, v []byte) error {
			var anomaly Anomaly
			if err := json.Unmarshal(v, &anomaly); err != nil {
				return fmt.Errorf("failed to unmarshal anomaly: %w", err)
			}

			// Apply status filter if specified
			if statusFilter == "" || anomaly.Status == statusFilter {
				anomalies = append(anomalies, anomaly)
			}
			return nil
		})
	})

	return anomalies, err
}

// GetAnomalyByID retrieves a specific anomaly by its ID
func (s *BoltStore) GetAnomalyByID(id string) (*Anomaly, error) {
	var anomaly *Anomaly

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(anomaliesBucket)
		data := b.Get([]byte(id))

		if data == nil {
			return fmt.Errorf("anomaly not found: %s", id)
		}

		anomaly = &Anomaly{}
		if err := json.Unmarshal(data, anomaly); err != nil {
			return fmt.Errorf("failed to unmarshal anomaly: %w", err)
		}
		return nil
	})

	return anomaly, err
}

// UpdateAnomalyStatus updates the status of an anomaly
func (s *BoltStore) UpdateAnomalyStatus(id, status string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(anomaliesBucket)

		// Get existing anomaly
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("anomaly not found: %s", id)
		}

		var anomaly Anomaly
		if err := json.Unmarshal(data, &anomaly); err != nil {
			return fmt.Errorf("failed to unmarshal anomaly: %w", err)
		}

		// Update status
		anomaly.Status = status

		// Save back
		encoded, err := json.Marshal(anomaly)
		if err != nil {
			return fmt.Errorf("failed to marshal anomaly: %w", err)
		}

		return b.Put([]byte(id), encoded)
	})
}

// GetStats returns statistics about the stored data
func (s *BoltStore) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	err := s.db.View(func(tx *bolt.Tx) error {
		// Count baselines (unique clients)
		baselineCount := tx.Bucket(baselineBucket).Stats().KeyN
		stats["unique_clients"] = baselineCount

		// Count processed queries
		queryCount := tx.Bucket(processedQueriesBucket).Stats().KeyN
		stats["total_queries"] = queryCount

		// Count anomalies by status
		anomaliesBkt := tx.Bucket(anomaliesBucket)
		totalAnomalies := 0
		pendingCount := 0
		approvedCount := 0
		blockedCount := 0
		maliciousCount := 0
		suspiciousCount := 0

		anomaliesBkt.ForEach(func(k, v []byte) error {
			var anomaly Anomaly
			if err := json.Unmarshal(v, &anomaly); err != nil {
				return nil // Skip malformed entries
			}

			totalAnomalies++

			// Count by status (default to "pending" if not set)
			status := anomaly.Status
			if status == "" {
				status = "pending"
			}
			switch status {
			case "pending":
				pendingCount++
			case "approved":
				approvedCount++
			case "blocked":
				blockedCount++
			}

			// Count by classification
			switch anomaly.Classification {
			case "Malicious":
				maliciousCount++
			case "Suspicious":
				suspiciousCount++
			}

			return nil
		})

		stats["total_anomalies"] = totalAnomalies
		stats["pending_anomalies"] = pendingCount
		stats["approved_anomalies"] = approvedCount
		stats["blocked_anomalies"] = blockedCount
		stats["malicious_count"] = maliciousCount
		stats["suspicious_count"] = suspiciousCount

		// Count analyses
		analysisCount := tx.Bucket(analysesBucket).Stats().KeyN
		stats["total_analyses"] = analysisCount

		return nil
	})

	return stats, err
}
