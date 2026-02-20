package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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

func TestHandleTree(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/tree")
	if err != nil {
		t.Fatalf("GET /api/v1/tree error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected 'data' to be an object (tree root)")
	}

	if data["type"] != "dir" {
		t.Errorf("root type = %v, want 'dir'", data["type"])
	}

	children, ok := data["children"].([]any)
	if !ok {
		t.Fatal("expected 'children' to be an array")
	}
	if len(children) != 3 {
		t.Errorf("expected 3 root children, got %d", len(children))
	}
}

func TestHandleRawFile(t *testing.T) {
	// Create a temporary directory with a test file
	tmpDir := t.TempDir()
	testContent := []byte("test image content")
	os.MkdirAll(filepath.Join(tmpDir, "images"), 0o755)
	os.WriteFile(filepath.Join(tmpDir, "images", "test.png"), testContent, 0o644)

	store, err := index.New()
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	cfg := config.ServeConfig{Port: 0, RootDir: tmpDir}
	srv := New(cfg, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/api/v1/raw/images/test.png")
	if err != nil {
		t.Fatalf("GET /api/v1/raw/images/test.png error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHandleRawFile_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := index.New()
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	cfg := config.ServeConfig{Port: 0, RootDir: tmpDir}
	srv := New(cfg, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	// Use URL-encoded ".." to bypass client-side path normalization
	resp, err := http.Get(ts.URL + "/api/v1/raw/%2e%2e/%2e%2e/etc/passwd")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("path traversal should be rejected")
	}
}

func TestHandleListDocuments_WithStatusFilter(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?status=published")
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
	total, _ := body["total"].(float64)

	if int(total) != 2 {
		t.Errorf("total = %v, want 2 (guide.md + japanese.md are published)", total)
	}
	if len(data) != 2 {
		t.Errorf("got %d docs, want 2", len(data))
	}

	for _, item := range data {
		doc, _ := item.(map[string]any)
		meta, _ := doc["meta"].(map[string]any)
		if meta["status"] != "published" {
			t.Errorf("expected status=published, got %v", meta["status"])
		}
	}
}

func TestHandleListDocuments_WithTagFilter(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?tag=go")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	total, _ := body["total"].(float64)
	if int(total) != 2 {
		t.Errorf("total = %v, want 2 (guide.md + api.md have tag 'go')", total)
	}
}

func TestHandleListDocuments_WithStatusAndTagFilter(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/documents?status=published&tag=go")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	total, _ := body["total"].(float64)
	if int(total) != 1 {
		t.Errorf("total = %v, want 1 (only guide.md is published+go)", total)
	}

	data, _ := body["data"].([]any)
	if len(data) == 1 {
		doc, _ := data[0].(map[string]any)
		if doc["path"] != "guide.md" {
			t.Errorf("expected guide.md, got %v", doc["path"])
		}
	}
}

func newTestServerWithGitRepo(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()

	// Create a temp git repo with test documents
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	os.WriteFile(filepath.Join(dir, "guide.md"), []byte("---\ntitle: Go Guide\nstatus: published\ntags:\n  - go\n  - tutorial\n---\n# Go Guide\n\nLearn Go programming.\n"), 0o644)
	run("add", "guide.md")
	run("commit", "-m", "initial: add guide.md")

	os.WriteFile(filepath.Join(dir, "guide.md"), []byte("---\ntitle: Go Guide\nstatus: published\ntags:\n  - go\n  - tutorial\n---\n# Go Guide\n\nLearn Go programming language.\nNew content added.\n"), 0o644)
	run("add", "guide.md")
	run("commit", "-m", "update: expand guide.md")

	store, err := index.New()
	if err != nil {
		t.Fatalf("index.New() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Index docs from the git repo
	docs, err := scanner.Scan(dir)
	if err != nil {
		t.Fatalf("scanner.Scan() error = %v", err)
	}
	for _, doc := range docs {
		if err := store.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument(%q) error = %v", doc.RelPath, err)
		}
	}

	cfg := config.ServeConfig{Port: 0, RootDir: dir}
	srv := New(cfg, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return srv, ts
}

func TestHandleHistory(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/git/history/guide.md")
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
	if len(data) < 2 {
		t.Errorf("expected at least 2 commits, got %d", len(data))
	}

	first, _ := data[0].(map[string]any)
	if first["hash"] == nil || first["hash"] == "" {
		t.Error("expected non-empty hash")
	}
	if first["author"] == nil || first["author"] == "" {
		t.Error("expected non-empty author")
	}
}

func TestHandleHistory_NotFound(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/git/history/nonexistent.md")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestHandleDiff(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	// First get the history to get commit hashes
	resp, err := http.Get(ts.URL + "/api/v1/git/history/guide.md")
	if err != nil {
		t.Fatalf("GET history error = %v", err)
	}
	var histBody map[string]any
	json.NewDecoder(resp.Body).Decode(&histBody)
	resp.Body.Close()

	commits, _ := histBody["data"].([]any)
	if len(commits) < 2 {
		t.Fatal("need at least 2 commits for diff test")
	}

	newCommit, _ := commits[0].(map[string]any)
	oldCommit, _ := commits[1].(map[string]any)
	from := oldCommit["hash"].(string)
	to := newCommit["hash"].(string)

	resp2, err := http.Get(ts.URL + "/api/v1/git/diff/guide.md?from=" + from + "&to=" + to)
	if err != nil {
		t.Fatalf("GET diff error = %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}

	var diffBody map[string]any
	json.NewDecoder(resp2.Body).Decode(&diffBody)

	diff, _ := diffBody["data"].(string)
	if diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestHandleDiff_MissingParams(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/git/diff/guide.md")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestHandleBlame(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/git/blame/guide.md")
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
		t.Error("expected non-empty blame data")
	}
}

func TestHandleBlame_WithLineRange(t *testing.T) {
	_, ts := newTestServerWithGitRepo(t)

	resp, err := http.Get(ts.URL + "/api/v1/git/blame/guide.md?start=1&end=3")
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
	if len(data) != 3 {
		t.Errorf("expected 3 blame lines, got %d", len(data))
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

func TestHandleGraph(t *testing.T) {
	_, ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/v1/graph")
	if err != nil {
		t.Fatalf("GET /api/v1/graph error = %v", err)
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

	nodes, ok := data["nodes"].([]any)
	if !ok {
		t.Fatal("expected 'nodes' to be an array")
	}
	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(nodes))
	}

	edges, ok := data["edges"].([]any)
	if !ok {
		t.Fatal("expected 'edges' to be an array")
	}
	// guide.md and api.md share "go" tag
	if len(edges) < 1 {
		t.Error("expected at least 1 edge for shared 'go' tag")
	}
}
