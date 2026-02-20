package index

import (
	"testing"
	"time"

	"github.com/esakat/markdown-kb/internal/scanner"
)

func TestBuildTree_Empty(t *testing.T) {
	tree := BuildTree(nil)
	if tree.Type != "dir" {
		t.Errorf("root type = %q, want 'dir'", tree.Type)
	}
	if len(tree.Children) != 0 {
		t.Errorf("expected 0 children, got %d", len(tree.Children))
	}
}

func TestBuildTree_SingleFile(t *testing.T) {
	entries := []PathEntry{{Path: "readme.md", Title: "README"}}
	tree := BuildTree(entries)

	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}
	child := tree.Children[0]
	if child.Name != "readme.md" {
		t.Errorf("name = %q, want 'readme.md'", child.Name)
	}
	if child.Type != "file" {
		t.Errorf("type = %q, want 'file'", child.Type)
	}
	if child.Path != "readme.md" {
		t.Errorf("path = %q, want 'readme.md'", child.Path)
	}
	if child.Title != "README" {
		t.Errorf("title = %q, want 'README'", child.Title)
	}
}

func TestBuildTree_NestedDirs(t *testing.T) {
	entries := []PathEntry{
		{Path: "docs/api/rest.md", Title: "REST API"},
		{Path: "docs/guide.md", Title: "Guide"},
	}
	tree := BuildTree(entries)

	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 root child (docs), got %d", len(tree.Children))
	}

	docs := tree.Children[0]
	if docs.Name != "docs" || docs.Type != "dir" {
		t.Fatalf("expected docs dir, got %q (%s)", docs.Name, docs.Type)
	}

	// Should have api/ dir first, then guide.md file (dirs before files)
	if len(docs.Children) != 2 {
		t.Fatalf("expected 2 children in docs, got %d", len(docs.Children))
	}

	if docs.Children[0].Name != "api" || docs.Children[0].Type != "dir" {
		t.Errorf("first child should be api dir, got %q (%s)", docs.Children[0].Name, docs.Children[0].Type)
	}
	if docs.Children[1].Name != "guide.md" || docs.Children[1].Type != "file" {
		t.Errorf("second child should be guide.md file, got %q (%s)", docs.Children[1].Name, docs.Children[1].Type)
	}

	// Check nested file
	apiDir := docs.Children[0]
	if len(apiDir.Children) != 1 {
		t.Fatalf("expected 1 child in api, got %d", len(apiDir.Children))
	}
	if apiDir.Children[0].Path != "docs/api/rest.md" {
		t.Errorf("path = %q, want 'docs/api/rest.md'", apiDir.Children[0].Path)
	}
}

func TestBuildTree_MixedDepths(t *testing.T) {
	entries := []PathEntry{
		{Path: "a.md", Title: "A"},
		{Path: "b/c.md", Title: "C"},
		{Path: "b/d.md", Title: "D"},
		{Path: "z.md", Title: "Z"},
	}
	tree := BuildTree(entries)

	// Root: b/ dir first, then a.md, z.md (dirs before files, alphabetical)
	if len(tree.Children) != 3 {
		t.Fatalf("expected 3 root children, got %d", len(tree.Children))
	}

	if tree.Children[0].Name != "b" || tree.Children[0].Type != "dir" {
		t.Errorf("first should be dir 'b', got %q (%s)", tree.Children[0].Name, tree.Children[0].Type)
	}
	if tree.Children[1].Name != "a.md" || tree.Children[1].Type != "file" {
		t.Errorf("second should be file 'a.md', got %q (%s)", tree.Children[1].Name, tree.Children[1].Type)
	}
	if tree.Children[2].Name != "z.md" || tree.Children[2].Type != "file" {
		t.Errorf("third should be file 'z.md', got %q (%s)", tree.Children[2].Name, tree.Children[2].Type)
	}
}

func TestBuildTree_Sort(t *testing.T) {
	entries := []PathEntry{
		{Path: "z.md", Title: "Z"},
		{Path: "a.md", Title: "A"},
		{Path: "m/b.md", Title: "B"},
		{Path: "d/e.md", Title: "E"},
	}
	tree := BuildTree(entries)

	// Dirs first (d, m), then files (a.md, z.md)
	if tree.Children[0].Name != "d" {
		t.Errorf("first = %q, want 'd'", tree.Children[0].Name)
	}
	if tree.Children[1].Name != "m" {
		t.Errorf("second = %q, want 'm'", tree.Children[1].Name)
	}
	if tree.Children[2].Name != "a.md" {
		t.Errorf("third = %q, want 'a.md'", tree.Children[2].Name)
	}
	if tree.Children[3].Name != "z.md" {
		t.Errorf("fourth = %q, want 'z.md'", tree.Children[3].Name)
	}
}

func TestListPaths(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer store.Close()

	docs := []scanner.Document{
		{RelPath: "guide.md", Frontmatter: map[string]any{"title": "Guide"}, Body: "content", ModTime: time.Now(), Size: 100},
		{RelPath: "api/rest.md", Frontmatter: map[string]any{"title": "REST"}, Body: "content", ModTime: time.Now(), Size: 200},
	}
	for _, doc := range docs {
		if err := store.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument error = %v", err)
		}
	}

	entries, err := store.ListPaths()
	if err != nil {
		t.Fatalf("ListPaths() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Should be ordered by path
	if entries[0].Path != "api/rest.md" {
		t.Errorf("first entry path = %q, want 'api/rest.md'", entries[0].Path)
	}
	if entries[0].Title != "REST" {
		t.Errorf("first entry title = %q, want 'REST'", entries[0].Title)
	}
	if entries[1].Path != "guide.md" {
		t.Errorf("second entry path = %q, want 'guide.md'", entries[1].Path)
	}
}
