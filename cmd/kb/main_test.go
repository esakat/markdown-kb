package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/esakat/markdown-kb/internal/scanner"
)

func createTestDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	os.WriteFile(filepath.Join(tmp, "hello.md"), []byte("---\ntitle: Hello\nstatus: published\ntags:\n  - test\n---\n\n# Hello\n"), 0o644)
	os.WriteFile(filepath.Join(tmp, "world.md"), []byte("# World\n\nNo frontmatter.\n"), 0o644)

	os.MkdirAll(filepath.Join(tmp, "sub"), 0o755)
	os.WriteFile(filepath.Join(tmp, "sub", "nested.md"), []byte("---\ntitle: Nested\nstatus: draft\n---\n\nNested.\n"), 0o644)

	return tmp
}

func TestValidateRootDir_Exists(t *testing.T) {
	tmp := t.TempDir()
	if err := validateRootDir(tmp); err != nil {
		t.Errorf("expected no error for existing dir, got %v", err)
	}
}

func TestValidateRootDir_NotExists(t *testing.T) {
	err := validateRootDir("/nonexistent/dir/abc123")
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}

func TestValidateRootDir_NotADir(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "file.txt")
	os.WriteFile(f, []byte("hello"), 0o644)
	err := validateRootDir(f)
	if err == nil {
		t.Error("expected error for file (not dir)")
	}
}

func TestScanAndIndex(t *testing.T) {
	tmp := createTestDir(t)
	store, docs, err := scanAndIndex(tmp)
	if err != nil {
		t.Fatalf("scanAndIndex() error = %v", err)
	}
	defer store.Close()

	if len(docs) != 3 {
		t.Errorf("expected 3 docs, got %d", len(docs))
	}

	doc, err := store.GetDocument("hello.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if doc == nil {
		t.Fatal("expected hello.md in index")
	}
	if doc.Title != "Hello" {
		t.Errorf("Title = %q, want %q", doc.Title, "Hello")
	}
}

func TestScanAndIndex_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	store, docs, err := scanAndIndex(tmp)
	if err != nil {
		t.Fatalf("scanAndIndex() error = %v", err)
	}
	defer store.Close()

	if len(docs) != 0 {
		t.Errorf("expected 0 docs, got %d", len(docs))
	}
}

func TestDocToEntry_WithFrontmatter(t *testing.T) {
	doc := scanner.Document{
		RelPath:     "test.md",
		Frontmatter: map[string]any{"title": "Test", "status": "draft", "tags": []any{"go", "test"}},
		Body:        "body",
		ModTime:     time.Now(),
		Size:        100,
	}

	entry := docToEntry(doc)
	if entry.Title != "Test" {
		t.Errorf("Title = %q, want %q", entry.Title, "Test")
	}
	if entry.Status != "draft" {
		t.Errorf("Status = %q, want %q", entry.Status, "draft")
	}
	if len(entry.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(entry.Tags))
	}
}

func TestDocToEntry_NilFrontmatter(t *testing.T) {
	doc := scanner.Document{
		RelPath: "plain.md",
		Body:    "body",
		ModTime: time.Now(),
		Size:    50,
	}

	entry := docToEntry(doc)
	if entry.Title != "" {
		t.Errorf("Title = %q, want empty", entry.Title)
	}
	if entry.Tags != nil {
		t.Errorf("Tags should be nil, got %v", entry.Tags)
	}
}

func TestOutputJSON(t *testing.T) {
	docs := []scanner.Document{
		{
			RelPath:     "test.md",
			Frontmatter: map[string]any{"title": "Test"},
			Size:        50,
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(docs)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("outputJSON() error = %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if output == "" {
		t.Error("expected non-empty JSON output")
	}
}

func TestOutputText(t *testing.T) {
	docs := []scanner.Document{
		{
			RelPath:     "test.md",
			Frontmatter: map[string]any{"title": "Test", "status": "published", "tags": []any{"go"}},
			Size:        50,
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputText(docs)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("outputText() error = %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if output == "" {
		t.Error("expected non-empty text output")
	}
}
