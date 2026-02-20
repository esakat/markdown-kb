package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":     {Data: []byte("<html><body>SPA</body></html>")},
		"assets/main.js": {Data: []byte("console.log('hello')")},
		"assets/style.css": {Data: []byte("body { margin: 0 }")},
	}
}

func TestSPAHandler_ServesIndex(t *testing.T) {
	handler := spaHandler(testFS())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "SPA") {
		t.Errorf("body = %q, want to contain 'SPA'", string(body))
	}
}

func TestSPAHandler_ServesStaticFiles(t *testing.T) {
	handler := spaHandler(testFS())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/assets/main.js")
	if err != nil {
		t.Fatalf("GET /assets/main.js error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "console.log") {
		t.Errorf("expected JS content, got %q", string(body))
	}
}

func TestSPAHandler_FallbackToIndex(t *testing.T) {
	handler := spaHandler(testFS())
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Request a path that doesn't exist as a file — should serve index.html
	resp, err := http.Get(ts.URL + "/docs/some/path")
	if err != nil {
		t.Fatalf("GET /docs/some/path error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "SPA") {
		t.Errorf("fallback should serve index.html, got %q", string(body))
	}
}

func TestSPAHandler_EmptyFS(t *testing.T) {
	handler := spaHandler(fstest.MapFS{})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	defer resp.Body.Close()

	// Empty FS — should return 404
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
