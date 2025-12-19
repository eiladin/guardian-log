package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eiladin/guardian-log/internal/analyzer"
	"github.com/eiladin/guardian-log/internal/api"
	"github.com/eiladin/guardian-log/internal/config"
	"github.com/eiladin/guardian-log/internal/enrichment"
	"github.com/eiladin/guardian-log/internal/ingestor"
	"github.com/eiladin/guardian-log/internal/llm"
	"github.com/eiladin/guardian-log/internal/llm/providers/gemini"
	"github.com/eiladin/guardian-log/internal/storage"
	"github.com/eiladin/guardian-log/webfs"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Guardian-Log starting...")
	log.Printf("AdGuard Home URL: %s", cfg.AdGuardURL)
	log.Printf("Poll Interval: %s", cfg.PollInterval)
	log.Printf("Database Path: %s", cfg.DBPath)

	// Initialize BoltDB storage
	store, err := storage.NewBoltStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
		log.Println("Storage closed")
	}()

	// Initialize AdGuard client
	adguardClient := ingestor.NewAdGuardClient(cfg.AdGuardURL, cfg.AdGuardUser, cfg.AdGuardPassword)

	// Initialize baseline analyzer
	baselineAnalyzer := analyzer.NewBaselineAnalyzer(store)

	// Initialize poller
	poller := ingestor.NewPoller(adguardClient, baselineAnalyzer, cfg.PollInterval)

	// Initialize LLM analysis if enabled
	var llmAnalyzer *llm.Analyzer
	if cfg.LLMEnabled {
		log.Printf("🤖 LLM Analysis: Enabled (provider: %s)", cfg.LLMProvider)

		// Initialize WHOIS service
		whoisService := enrichment.NewWHOISService(store)

		// Initialize LLM provider based on configuration
		var provider llm.Provider
		switch cfg.LLMProvider {
		case "gemini":
			provider, err = gemini.NewProvider(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.LLMTimeout)
			if err != nil {
				log.Fatalf("Failed to initialize Gemini provider: %v", err)
			}
			log.Printf("Gemini provider initialized (model: %s)", cfg.GeminiModel)

		// Future providers can be added here
		// case "openai": ...
		// case "anthropic": ...
		// case "ollama": ...

		default:
			log.Fatalf("Unsupported LLM provider: %s", cfg.LLMProvider)
		}

		// Initialize LLM analyzer with configured batch settings
		llmAnalyzer = llm.NewAnalyzer(provider, whoisService, store, cfg.LLMBatchSize, cfg.LLMBatchTimeout, cfg.LLMBatchDelay)
		poller.SetLLMAnalyzer(llmAnalyzer)
		defer llmAnalyzer.Stop()

		log.Printf("LLM analyzer initialized (batch: %d domains, timeout: %s, delay: %s)",
			cfg.LLMBatchSize, cfg.LLMBatchTimeout, cfg.LLMBatchDelay)
	} else {
		log.Println("LLM Analysis: Disabled")
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get embedded web filesystem
	webFS, err := webfs.GetFS()
	if err != nil {
		log.Printf("⚠️  Warning: Could not load embedded frontend: %v", err)
		webFS = nil // API server will work without frontend
	}

	// Initialize and start API server
	apiServer := api.NewServer(store, cfg, adguardClient, llmAnalyzer, webFS)

	// Start API server in a goroutine
	go func() {
		apiAddr := ":8080" // Default API port
		if err := apiServer.Start(apiAddr); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start poller in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- poller.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("\nReceived shutdown signal, stopping gracefully...")
		cancel()

		// Shutdown API server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := apiServer.Stop(shutdownCtx); err != nil {
			log.Printf("Error stopping API server: %v", err)
		}

		// Give the poller time to finish current operation
		time.Sleep(2 * time.Second)

		log.Println("Shutdown complete")
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			log.Printf("Poller error: %v", err)
		}
	}
}
