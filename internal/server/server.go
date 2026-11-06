package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/crawler"
	"github.com/alexandrmotologa/brokenlens/internal/model"
	"github.com/alexandrmotologa/brokenlens/web"
)

// Server handles embedded HTTP serving and API requests for visual inspection.
type Server struct {
	cfg         *config.Config
	reportMu    sync.RWMutex
	report      *model.AuditReport
	clientsMu   sync.Mutex
	clients     map[chan []byte]bool
	isScanning  bool
	scanningMu  sync.Mutex
}

// NewServer creates a new web studio server.
func NewServer(cfg *config.Config, initialReport *model.AuditReport) *Server {
	return &Server{
		cfg:     cfg,
		report:  initialReport,
		clients: make(map[chan []byte]bool),
	}
}

// LoadReportFromFile loads an existing report JSON file into the server.
func (s *Server) LoadReportFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var rep model.AuditReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return err
	}
	s.reportMu.Lock()
	s.report = &rep
	s.reportMu.Unlock()
	return nil
}

// Start launches the HTTP server and blocks until context cancellation or error.
func (s *Server) Start(addr string) error {
	distFS, err := web.GetFS()
	if err != nil {
		return fmt.Errorf("failed to load embedded web assets: %w", err)
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/report", s.handleGetReport)
	mux.HandleFunc("POST /api/scan", s.handlePostScan)
	mux.HandleFunc("GET /api/events", s.handleSSEEvents)

	// Static SPA handler with fallback to index.html
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded filesystem
		f, err := distFS.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA routing
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusNotFound)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.Copy(w, indexFile)
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // Keep open for SSE
	}

	fmt.Printf("[brokenlens] Report Studio running at: http://%s\n", addr)
	return server.ListenAndServe()
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	s.reportMu.RLock()
	rep := s.report
	s.reportMu.RUnlock()

	if rep == nil {
		http.Error(w, `{"error":"No audit report generated yet"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rep)
}

type scanRequest struct {
	URL string `json:"url"`
}

func (s *Server) handlePostScan(w http.ResponseWriter, r *http.Request) {
	s.scanningMu.Lock()
	if s.isScanning {
		s.scanningMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "An audit is already running"})
		return
	}
	s.isScanning = true
	s.scanningMu.Unlock()

	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		s.scanningMu.Lock()
		s.isScanning = false
		s.scanningMu.Unlock()
		http.Error(w, `{"error":"Valid 'url' parameter required"}`, http.StatusBadRequest)
		return
	}

	// Trigger asynchronous scan
	go s.runAsyncScan(req.URL)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "Audit initiated", "url": req.URL})
}

func (s *Server) runAsyncScan(targetURL string) {
	defer func() {
		s.scanningMu.Lock()
		s.isScanning = false
		s.scanningMu.Unlock()
	}()

	scanCfg := *s.cfg
	scanCfg.TargetURL = targetURL
	_ = scanCfg.Validate()

	engine, err := crawler.NewEngine(&scanCfg, func(stats crawler.ProgressStats) {
		payload, err := json.Marshal(map[string]interface{}{
			"type":           "progress",
			"totalScanned":   stats.TotalScanned,
			"totalQueued":    stats.TotalQueued,
			"totalBroken":    stats.TotalBroken,
			"activeWorkers":  stats.ActiveWorkers,
			"currentUrl":     stats.CurrentURL,
			"durationMs":     stats.Duration.Milliseconds(),
			"requestsPerSec": stats.RequestsPerSec,
		})
		if err == nil {
			s.broadcastEvent(payload)
		}
	})

	if err != nil {
		return
	}

	rep, err := engine.Run(context.Background())
	if err == nil && rep != nil {
		s.reportMu.Lock()
		s.report = rep
		s.reportMu.Unlock()

		donePayload, err := json.Marshal(map[string]interface{}{
			"type":   "done",
			"report": rep,
		})
		if err == nil {
			s.broadcastEvent(donePayload)
		}
	}
}

func (s *Server) handleSSEEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msgCh := make(chan []byte, 100)

	s.clientsMu.Lock()
	s.clients[msgCh] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, msgCh)
		close(msgCh)
		s.clientsMu.Unlock()
	}()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *Server) broadcastEvent(data []byte) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for ch := range s.clients {
		select {
		case ch <- data:
		default:
		}
	}
}
