package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/esakat/markdown-kb/internal/config"
	"github.com/esakat/markdown-kb/internal/index"
	"github.com/esakat/markdown-kb/web"
)

var version = "dev"

// Server is the HTTP server that serves the web UI and API.
type Server struct {
	cfg    config.ServeConfig
	store  *index.Store
	mux    *http.ServeMux
	server *http.Server
}

// New creates a new server instance.
func New(cfg config.ServeConfig, store *index.Store) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: s.Handler(),
	}
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/v1/documents", s.handleListDocuments)
	s.mux.HandleFunc("GET /api/v1/documents/{path...}", s.handleGetDocument)
	s.mux.HandleFunc("GET /api/v1/search", s.handleSearch)
	s.mux.HandleFunc("GET /api/v1/tags", s.handleListTags)
	s.mux.HandleFunc("GET /api/v1/metadata/fields", s.handleMetadataFields)
	s.mux.HandleFunc("GET /api/v1/tree", s.handleTree)
	s.mux.HandleFunc("GET /api/v1/raw/{path...}", s.handleRawFile)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)

	// SPA catch-all (lowest priority in ServeMux)
	sub, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		sub = web.DistFS
	}
	s.mux.Handle("GET /", spaHandler(sub))
}

// Handler returns the HTTP handler with CORS middleware.
func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server with a timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	docs, total, _ := s.store.ListDocuments(0, 0)
	_ = docs
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"version":   version,
		"documents": total,
	})
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	docs, total, err := s.store.ListDocuments(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}

	if docs == nil {
		docs = []index.DocumentSummary{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  docs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	doc, err := s.store.GetDocument(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get document")
		return
	}
	if doc == nil {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": doc})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Build filters from query params
	filters := make(map[string]string)
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		filters["tags"] = tag
	}

	var results []index.SearchResult
	var total int
	var err error

	if len(filters) > 0 {
		results, total, err = s.store.SearchWithFilter(q, filters, limit, offset)
	} else {
		results, total, err = s.store.Search(q, limit, offset)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	if results == nil {
		results = []index.SearchResult{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  results,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.store.ListTags()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}

	if tags == nil {
		tags = []index.TagCount{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": tags})
}

func (s *Server) handleMetadataFields(w http.ResponseWriter, r *http.Request) {
	fields, err := s.store.ListMetadataFields()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list metadata fields")
		return
	}

	if fields == nil {
		fields = []index.MetadataField{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": fields})
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	entries, err := s.store.ListPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build tree")
		return
	}

	tree := index.BuildTree(entries)
	writeJSON(w, http.StatusOK, map[string]any{"data": tree})
}

func (s *Server) handleRawFile(w http.ResponseWriter, r *http.Request) {
	filePath := r.PathValue("path")
	if filePath == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	// Prevent path traversal
	if strings.Contains(filePath, "..") {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	fullPath := filepath.Join(s.cfg.RootDir, filepath.FromSlash(filePath))

	// Verify the resolved path is within RootDir
	absRoot, err := filepath.Abs(s.cfg.RootDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil || !strings.HasPrefix(absPath, absRoot) {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	http.ServeFile(w, r, fullPath)
}
