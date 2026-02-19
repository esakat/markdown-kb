package server

import (
	"fmt"
	"net/http"

	"github.com/esakat/markdown-kb/internal/config"
	"github.com/esakat/markdown-kb/internal/index"
)

// Server is the HTTP server that serves the web UI and API.
type Server struct {
	cfg   config.ServeConfig
	store *index.Store
	mux   *http.ServeMux
}

// New creates a new server instance.
func New(cfg config.ServeConfig, store *index.Store) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	// API v1
	s.mux.HandleFunc("GET /api/v1/documents", s.handleListDocuments)
	s.mux.HandleFunc("GET /api/v1/documents/{path...}", s.handleGetDocument)
	s.mux.HandleFunc("GET /api/v1/search", s.handleSearch)
	s.mux.HandleFunc("GET /api/v1/tags", s.handleListTags)
	s.mux.HandleFunc("GET /api/v1/metadata/fields", s.handleMetadataFields)

	// Health check
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	// SPA fallback (TODO: embed static files)
	// s.mux.Handle("/", http.FileServer(http.FS(webFS)))
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data":[],"total":0}`)
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"error":"not implemented"}`)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data":[],"total":0}`)
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data":[]}`)
}

func (s *Server) handleMetadataFields(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data":[]}`)
}
