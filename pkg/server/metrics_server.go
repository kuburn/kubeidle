package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer represents the HTTP server for exposing metrics
type MetricsServer struct {
	port int
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	return &MetricsServer{
		port: port,
	}
}

// Start starts the metrics server
func (s *MetricsServer) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting metrics server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
