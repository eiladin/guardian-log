package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/eiladin/guardian-log/internal/config"
	"github.com/eiladin/guardian-log/internal/storage"
)

func main() {
	// Load configuration to get DB path
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Check if database file exists
	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", cfg.DBPath)
	}

	// Open database
	store, err := storage.NewBoltStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer store.Close()

	fmt.Printf("=== Guardian-Log Database Inspection ===\n")
	fmt.Printf("Database: %s\n\n", cfg.DBPath)

	// Get all baselines
	baselines, err := store.GetAllBaselines()
	if err != nil {
		log.Fatalf("Failed to get baselines: %v", err)
	}

	if len(baselines) == 0 {
		fmt.Println("No baselines found in database.")
		fmt.Println("The database is empty - run guardian-log to start building baselines.")
		return
	}

	// Display statistics
	totalDomains := 0
	for _, baseline := range baselines {
		totalDomains += len(baseline.Domains)
	}

	fmt.Printf("Total Clients: %d\n", len(baselines))
	fmt.Printf("Total Domains: %d\n\n", totalDomains)

	// Display each baseline
	fmt.Printf("=== Client Baselines ===\n\n")
	for i, baseline := range baselines {
		fmt.Printf("[%d] Client: %s (%s)\n", i+1, baseline.ClientName, baseline.ClientID)
		fmt.Printf("    Domains: %d\n", len(baseline.Domains))
		fmt.Printf("    Last Updated: %s\n", baseline.LastUpdated.Format("2006-01-02 15:04:05"))

		if len(baseline.Domains) > 0 {
			fmt.Printf("    Sample Domains (first 10):\n")
			maxDisplay := 10
			if len(baseline.Domains) < maxDisplay {
				maxDisplay = len(baseline.Domains)
			}
			for j := 0; j < maxDisplay; j++ {
				fmt.Printf("      - %s\n", baseline.Domains[j])
			}
			if len(baseline.Domains) > 10 {
				fmt.Printf("      ... and %d more\n", len(baseline.Domains)-10)
			}
		}
		fmt.Println()
	}

	// Optional: export to JSON
	if len(os.Args) > 1 && os.Args[1] == "--json" {
		jsonData, err := json.MarshalIndent(baselines, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println("=== JSON Export ===")
		fmt.Println(string(jsonData))
	}
}
