package api

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eiladin/guardian-log/internal/config"
	"github.com/eiladin/guardian-log/internal/ingestor"
	"github.com/eiladin/guardian-log/internal/llm"
	"github.com/eiladin/guardian-log/internal/storage"
)

// Server represents the HTTP API server
type Server struct {
	store         *storage.BoltStore
	config        *config.Config
	adguardClient *ingestor.AdGuardClient
	llmAnalyzer   *llm.Analyzer
	httpServer    *http.Server
	webFS         fs.FS // Optional embedded frontend filesystem
}

// NewServer creates a new API server
func NewServer(
	store *storage.BoltStore,
	cfg *config.Config,
	adguardClient *ingestor.AdGuardClient,
	llmAnalyzer *llm.Analyzer,
	webFS fs.FS,
) *Server {
	return &Server{
		store:         store,
		config:        cfg,
		adguardClient: adguardClient,
		llmAnalyzer:   llmAnalyzer,
		webFS:         webFS,
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// API routes (registered first to take precedence)
	mux.HandleFunc("/api/anomalies", s.handleAnomalies)
	mux.HandleFunc("/api/anomalies/", s.handleAnomalyAction)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/health", s.handleHealth)

	// Serve static files from embedded dist folder if available
	if s.webFS != nil {
		// Create SPA handler for frontend
		spaHandler := http.FileServer(http.FS(s.webFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// If path starts with /api, it's handled by API routes above
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.NotFound(w, r)
				return
			}

			// Serve static files or index.html for SPA routes
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			// Check if file exists in embedded FS
			if _, err := fs.Stat(s.webFS, path); err == nil {
				spaHandler.ServeHTTP(w, r)
				return
			}

			// File not found, serve index.html for SPA client-side routing
			r.URL.Path = "/"
			spaHandler.ServeHTTP(w, r)
		})
		log.Printf("📱 Serving frontend from embedded filesystem")
	} else {
		log.Printf("⚠️  Warning: Frontend not embedded. API routes only.")
	}

	// CORS middleware
	handler := enableCORS(mux)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🌐 Server starting on %s", addr)
	log.Printf("   - API: http://localhost%s/api", addr)
	log.Printf("   - Dashboard: http://localhost%s", addr)

	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping API server...")
	return s.httpServer.Shutdown(ctx)
}

// enableCORS adds CORS headers to responses
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth responds with server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
