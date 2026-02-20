package index

import (
	"testing"
	"time"

	"github.com/esakat/markdown-kb/internal/scanner"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func sampleDocs() []scanner.Document {
	now := time.Now()
	return []scanner.Document{
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
			Body:        "# 日本語ガイド\n\nこれは日本語のドキュメントです。全文検索のテスト。",
			ModTime:     now,
			Size:        300,
		},
	}
}

func indexSampleDocs(t *testing.T, store *Store) {
	t.Helper()
	for _, doc := range sampleDocs() {
		if err := store.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument(%q) error = %v", doc.RelPath, err)
		}
	}
}

func TestNew_CreatesStore(t *testing.T) {
	store := newTestStore(t)
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestIndexDocument_AndGetDocument(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	doc, err := store.GetDocument("guide.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
	if doc.Title != "Go Guide" {
		t.Errorf("Title = %q, want %q", doc.Title, "Go Guide")
	}
	if doc.Path != "guide.md" {
		t.Errorf("Path = %q, want %q", doc.Path, "guide.md")
	}
	if doc.Body == "" {
		t.Error("expected non-empty body")
	}
	if doc.Size != 100 {
		t.Errorf("Size = %d, want %d", doc.Size, 100)
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	store := newTestStore(t)

	doc, err := store.GetDocument("nonexistent.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if doc != nil {
		t.Errorf("expected nil for nonexistent document, got %v", doc)
	}
}

func TestSearch_English(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	results, total, err := store.Search("programming", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if total == 0 {
		t.Fatal("expected at least 1 result for 'programming'")
	}
	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}
	if results[0].Path != "guide.md" {
		t.Errorf("expected guide.md as top result, got %q", results[0].Path)
	}
}

func TestSearch_Japanese(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	results, total, err := store.Search("日本語", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if total == 0 {
		t.Fatal("expected at least 1 result for '日本語'")
	}
	if results[0].Path != "japanese.md" {
		t.Errorf("expected japanese.md as top result, got %q", results[0].Path)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	results, total, err := store.Search("", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if total != 0 || results != nil {
		t.Errorf("expected empty results for empty query, got %d results", total)
	}
}

func TestSearch_Pagination(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	// Search for "published" which appears in guide.md and japanese.md meta (trigram needs 3+ chars)
	results1, total, err := store.Search("published", 1, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if total < 2 {
		t.Fatalf("expected at least 2 total results, got %d", total)
	}
	if len(results1) != 1 {
		t.Fatalf("expected 1 result with limit=1, got %d", len(results1))
	}

	results2, _, err := store.Search("published", 1, 1)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results2) != 1 {
		t.Fatalf("expected 1 result on page 2, got %d", len(results2))
	}

	// The two pages should have different documents
	if results1[0].Path == results2[0].Path {
		t.Error("pagination returned same document on both pages")
	}
}

func TestSearchWithFilter(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	results, total, err := store.SearchWithFilter("Guide", map[string]string{"status": "published"}, 10, 0)
	if err != nil {
		t.Fatalf("SearchWithFilter() error = %v", err)
	}
	if total == 0 {
		t.Fatal("expected at least 1 result")
	}
	for _, r := range results {
		status, _ := r.Meta["status"].(string)
		if status != "published" {
			t.Errorf("expected status=published, got %q for %q", status, r.Path)
		}
	}
}

func TestUpsert_UpdatesExistingDocument(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	// Update the guide document
	updatedDoc := scanner.Document{
		RelPath:     "guide.md",
		Frontmatter: map[string]any{"title": "Updated Guide", "status": "published", "tags": []any{"go"}},
		Body:        "# Updated Guide\n\nNew content.",
		ModTime:     time.Now(),
		Size:        150,
	}
	if err := store.IndexDocument(updatedDoc); err != nil {
		t.Fatalf("IndexDocument() error = %v", err)
	}

	doc, err := store.GetDocument("guide.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if doc.Title != "Updated Guide" {
		t.Errorf("Title = %q, want %q", doc.Title, "Updated Guide")
	}
	if doc.Size != 150 {
		t.Errorf("Size = %d, want %d", doc.Size, 150)
	}

	// FTS should also be updated
	results, _, err := store.Search("New content", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("expected search to find updated content")
	}
}

func TestRemoveDocument(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	if err := store.RemoveDocument("guide.md"); err != nil {
		t.Fatalf("RemoveDocument() error = %v", err)
	}

	doc, err := store.GetDocument("guide.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if doc != nil {
		t.Error("expected nil after removal")
	}

	// Should not appear in search results
	results, _, err := store.Search("programming", 10, 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	for _, r := range results {
		if r.Path == "guide.md" {
			t.Error("removed document should not appear in search results")
		}
	}
}

func TestListDocuments(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocuments(10, 0)
	if err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(docs) != 3 {
		t.Errorf("got %d docs, want 3", len(docs))
	}
}

func TestListDocuments_Pagination(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocuments(2, 0)
	if err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(docs) != 2 {
		t.Errorf("got %d docs, want 2", len(docs))
	}

	docs2, _, err := store.ListDocuments(2, 2)
	if err != nil {
		t.Fatalf("ListDocuments() error = %v", err)
	}
	if len(docs2) != 1 {
		t.Errorf("got %d docs on page 2, want 1", len(docs2))
	}
}

func TestListTags(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	tags, err := store.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("expected non-empty tag list")
	}

	tagMap := make(map[string]int)
	for _, tc := range tags {
		tagMap[tc.Tag] = tc.Count
	}

	// "go" should appear in guide.md and api.md
	if tagMap["go"] != 2 {
		t.Errorf("tag 'go' count = %d, want 2", tagMap["go"])
	}
	if tagMap["tutorial"] != 1 {
		t.Errorf("tag 'tutorial' count = %d, want 1", tagMap["tutorial"])
	}
}

func TestListMetadataFields(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	fields, err := store.ListMetadataFields()
	if err != nil {
		t.Fatalf("ListMetadataFields() error = %v", err)
	}
	if len(fields) == 0 {
		t.Fatal("expected non-empty metadata fields")
	}

	fieldMap := make(map[string]MetadataField)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	// Check "status" field
	if sf, ok := fieldMap["status"]; ok {
		if sf.Type != "string" {
			t.Errorf("status type = %q, want %q", sf.Type, "string")
		}
		if len(sf.Values) < 2 {
			t.Errorf("expected at least 2 status values, got %d", len(sf.Values))
		}
	} else {
		t.Error("expected 'status' field in metadata")
	}

	// Check "tags" field
	if tf, ok := fieldMap["tags"]; ok {
		if tf.Type != "array" {
			t.Errorf("tags type = %q, want %q", tf.Type, "array")
		}
	} else {
		t.Error("expected 'tags' field in metadata")
	}
}

func TestIndexDocument_NilFrontmatter(t *testing.T) {
	store := newTestStore(t)

	doc := scanner.Document{
		RelPath: "plain.md",
		Body:    "Just plain text without any frontmatter.",
		ModTime: time.Now(),
		Size:    40,
	}
	if err := store.IndexDocument(doc); err != nil {
		t.Fatalf("IndexDocument() error = %v", err)
	}

	got, err := store.GetDocument("plain.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected document, got nil")
	}
	if got.Title != "" {
		t.Errorf("expected empty title, got %q", got.Title)
	}
}

func TestIndexDocument_WithNumberAndBoolMeta(t *testing.T) {
	store := newTestStore(t)

	doc := scanner.Document{
		RelPath:     "types.md",
		Frontmatter: map[string]any{"title": "Types", "priority": 1.0, "draft": true, "tags": "single-tag"},
		Body:        "Metadata type test",
		ModTime:     time.Now(),
		Size:        30,
	}
	if err := store.IndexDocument(doc); err != nil {
		t.Fatalf("IndexDocument() error = %v", err)
	}

	fields, err := store.ListMetadataFields()
	if err != nil {
		t.Fatalf("ListMetadataFields() error = %v", err)
	}
	fieldMap := make(map[string]MetadataField)
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	if f, ok := fieldMap["priority"]; ok {
		if f.Type != "number" {
			t.Errorf("priority type = %q, want number", f.Type)
		}
	} else {
		t.Error("expected 'priority' field")
	}

	if f, ok := fieldMap["draft"]; ok {
		if f.Type != "boolean" {
			t.Errorf("draft type = %q, want boolean", f.Type)
		}
	} else {
		t.Error("expected 'draft' field")
	}

	// Tags as a single string triggers the string branch in ListTags
	tags, err := store.ListTags()
	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}
	found := false
	for _, tc := range tags {
		if tc.Tag == "single-tag" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'single-tag' in tags")
	}
}

func TestListDocumentsWithFilter_StatusOnly(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocumentsWithFilter(map[string]string{"status": "published"}, 10, 0)
	if err != nil {
		t.Fatalf("ListDocumentsWithFilter() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2 (guide.md + japanese.md)", total)
	}
	for _, d := range docs {
		status, _ := d.Meta["status"].(string)
		if status != "published" {
			t.Errorf("expected status=published, got %q for %q", status, d.Path)
		}
	}
}

func TestListDocumentsWithFilter_TagOnly(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocumentsWithFilter(map[string]string{"tags": "go"}, 10, 0)
	if err != nil {
		t.Fatalf("ListDocumentsWithFilter() error = %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2 (guide.md + api.md)", total)
	}
	if len(docs) != 2 {
		t.Errorf("got %d docs, want 2", len(docs))
	}
}

func TestListDocumentsWithFilter_StatusAndTag(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocumentsWithFilter(map[string]string{"status": "published", "tags": "go"}, 10, 0)
	if err != nil {
		t.Fatalf("ListDocumentsWithFilter() error = %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1 (only guide.md is published+go)", total)
	}
	if len(docs) != 1 {
		t.Errorf("got %d docs, want 1", len(docs))
	}
	if len(docs) > 0 && docs[0].Path != "guide.md" {
		t.Errorf("expected guide.md, got %q", docs[0].Path)
	}
}

func TestListDocumentsWithFilter_EmptyFilters(t *testing.T) {
	store := newTestStore(t)
	indexSampleDocs(t, store)

	docs, total, err := store.ListDocumentsWithFilter(nil, 10, 0)
	if err != nil {
		t.Fatalf("ListDocumentsWithFilter() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3 (all docs)", total)
	}
	if len(docs) != 3 {
		t.Errorf("got %d docs, want 3", len(docs))
	}
}

func TestNewWithPath(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	store, err := NewWithPath(dbPath)
	if err != nil {
		t.Fatalf("NewWithPath() error = %v", err)
	}
	defer store.Close()

	doc := scanner.Document{
		RelPath:     "test.md",
		Frontmatter: map[string]any{"title": "Test"},
		Body:        "Test body",
		ModTime:     time.Now(),
		Size:        50,
	}
	if err := store.IndexDocument(doc); err != nil {
		t.Fatalf("IndexDocument() error = %v", err)
	}

	got, err := store.GetDocument("test.md")
	if err != nil {
		t.Fatalf("GetDocument() error = %v", err)
	}
	if got == nil || got.Title != "Test" {
		t.Errorf("expected document with title 'Test', got %v", got)
	}
}
