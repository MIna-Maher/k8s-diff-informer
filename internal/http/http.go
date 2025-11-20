package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

// Server represents the HTTP server for metrics and health endpoints
type Server struct {
	server *http.Server
	port   int
}

// NewServer creates a new HTTP server
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start starts the HTTP server with metrics and health endpoints
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoints
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/ready", s.readinessHandler)
	mux.HandleFunc("/readyz", s.readinessHandler)

	// Root endpoint with basic info
	mux.HandleFunc("/", s.rootHandler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
		// Security settings
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	klog.Infof("Starting HTTP server on port %d", s.port)
	klog.Infof("Metrics endpoint: http://localhost:%d/metrics", s.port)
	klog.Infof("Health endpoint: http://localhost:%d/health", s.port)
	klog.Infof("Readiness endpoint: http://localhost:%d/ready", s.port)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("HTTP server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	<-ctx.Done()
	return s.Shutdown()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown() error {
	klog.Info("Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		klog.Errorf("HTTP server shutdown error: %v", err)
		return err
	}

	klog.Info("HTTP server shutdown complete")
	return nil
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// readinessHandler handles readiness check requests
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// rootHandler handles requests to the root endpoint
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
  "service": "k8s-diff-informer",
  "endpoints": {
    "metrics": "/metrics",
    "health": "/health",
    "readiness": "/ready"
  },
  "timestamp": "%s"
}`, time.Now().UTC().Format(time.RFC3339))
}
