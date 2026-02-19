package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/esakat/markdown-kb/internal/config"
	"github.com/esakat/markdown-kb/internal/index"
	"github.com/esakat/markdown-kb/internal/scanner"
)

func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()

	store, err := index.New()
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Index sample documents
	now := time.Now()
	docs := []scanner.Document{
		{
			RelPath:     "guide.md",
			Frontmatter: map[string]any{"title": "Go Guide", "status": "published", "tags": []any{"go", "tutorial"}},
			Body:        "# Go Guide\n\nLearn Go programming language.",
			ModTime:     now,
			Size:        100,
		},
		{
			RelPath:     "api.md",
			Frontmatter: map[string]any{"title": "API Reference", "status": "draft", "tags": []any{"go", "api"}},
			Body:        "# API Reference\n\nREST API documentation.",
			ModTime:     now,
			Size:        200,
		},
		{
			RelPath:     "japanese.md",
			Frontmatter: map[string]any{"title": "日本語ガイド", "status": "published", "tags": []any{"japanese"}},
			Body:        "# 日本語ガイド\n\nこれは日本語のドキュメントです。",
			ModTime:     now,
			Size:        300,
		},
	}
	for _, doc := range docs {
		if err := store.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument(%q) error = %v", doc.RelPath, err)
		}
	}

	cfg := config.ServeConfig{Port: 0}
	srv := New(cfg, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return srv, ts
}

func TestHandleHealth(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET /api/health error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("status = %v, want %q", body["status"], "ok")
	}
	if _, ok := body["documents"]; !ok {
		t.Error("expected 'documents' field in health response")
	}
}

func TestHandleListDocuments(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents")
	if err != nil {
		t.Fatalf("GET /api/v1/documents error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].([]any)
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) != 3 {
		t.Errorf("expected 3 documents, got %d", len(data))
	}

	total, _ := body["total"].(float64)
	if int(total) != 3 {
		t.Errorf("total = %v, want 3", total)
	}
}

func TestHandleListDocuments_Pagination(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?page=1&limit=2")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, _ := body["data"].([]any)
	if len(data) != 2 {
		t.Errorf("expected 2 documents with limit=2, got %d", len(data))
	}

	page, _ := body["page"].(float64)
	if int(page) != 1 {
		t.Errorf("page = %v, want 1", page)
	}
	limit, _ := body["limit"].(float64)
	if int(limit) != 2 {
		t.Errorf("limit = %v, want 2", limit)
	}
}

func TestHandleListDocuments_InvalidLimit(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?limit=999")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	// Should cap at 100
	limit, _ := body["limit"].(float64)
	if int(limit) > 100 {
		t.Errorf("limit should be capped at 100, got %v", limit)
	}
}

func TestHandleGetDocument(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents/guide.md")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected 'data' to be an object")
	}
	if data["title"] != "Go Guide" {
		t.Errorf("title = %v, want %q", data["title"], "Go Guide")
	}
}

func TestHandleGetDocument_NotFound(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents/nonexistent.md")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleSearch(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/search?q=programming")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].([]any)
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) == 0 {
		t.Error("expected at least 1 search result")
	}
}

func TestHandleSearch_MissingQuery(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/search")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleListTags(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/tags")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].([]any)
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) == 0 {
		t.Error("expected at least 1 tag")
	}
}

func TestHandleMetadataFields(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/metadata/fields")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].([]any)
	if !ok {
		t.Fatal("expected 'data' to be an array")
	}
	if len(data) == 0 {
		t.Error("expected at least 1 metadata field")
	}
}

func TestCORSHeaders(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	cors := resp.Header.Get("Access-Control-Allow-Origin")
	if cors != "*" {
		t.Errorf("CORS header = %q, want %q", cors, "*")
	}
}

func TestStartAndShutdown(t *testing.T) {
	store, err := index.New()
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}
	defer store.Close()

	cfg := config.ServeConfig{Port: 0} // port 0 = random free port
	srv := New(cfg, store)

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	// Start should return ErrServerClosed
	if startErr := <-errCh; startErr != nil && startErr != http.ErrServerClosed {
		t.Fatalf("Start() returned unexpected error: %v", startErr)
	}
}

func TestShutdown_NilServer(t *testing.T) {
	store, _ := index.New()
	defer store.Close()
	srv := New(config.ServeConfig{}, store)
	// Shutdown without Start should be a no-op
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown() without Start should not error, got %v", err)
	}
}

func TestCORSPreflight(t *testing.T) {
	_, ts := newTestServer(t)

	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/api/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestHandleSearch_WithFilters(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/search?q=published&status=published")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, _ := body["data"].([]any)
	for _, item := range data {
		doc, _ := item.(map[string]any)
		meta, _ := doc["meta"].(map[string]any)
		if meta["status"] != "published" {
			t.Errorf("expected status=published, got %v", meta["status"])
		}
	}
}

func TestHandleSearch_WithTagFilter(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/search?q=Reference&tag=api")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHandleSearch_Pagination(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/search?q=published&page=1&limit=1")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	limit, _ := body["limit"].(float64)
	if int(limit) != 1 {
		t.Errorf("limit = %v, want 1", limit)
	}
}

func TestHandleListDocuments_Page2(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?page=2&limit=2")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, _ := body["data"].([]any)
	if len(data) != 1 {
		t.Errorf("expected 1 document on page 2 (3 total, limit 2), got %d", len(data))
	}
}

func TestQueryInt_InvalidValue(t *testing.T) {
	_, ts := newTestServer(t)

	// Negative page should fallback to default
	resp, err := http.Get(ts.URL + "/api/v1/documents?page=-1&limit=abc")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
