package index

import (
	"testing"
	"time"

	"github.com/esakat/markdown-kb/internal/scanner"
)

func TestBuildGraph_NodesAndTagEdges(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer store.Close()

	now := time.Now()

	// Two docs sharing the "go" tag
	store.IndexDocument(scanner.Document{
		RelPath:     "guide.md",
		Frontmatter: map[string]any{"title": "Guide", "tags": []any{"go", "tutorial"}},
		Body:        "# Guide",
		ModTime:     now,
		Size:        100,
	})
	store.IndexDocument(scanner.Document{
		RelPath:     "setup.md",
		Frontmatter: map[string]any{"title": "Setup", "tags": []any{"go"}},
		Body:        "# Setup",
		ModTime:     now,
		Size:        80,
	})

	graph, err := store.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	// Should have a tag edge for shared "go" tag
	var tagEdges int
	for _, e := range graph.Edges {
		if e.Type == "tag" && e.Label == "go" {
			tagEdges++
		}
	}
	if tagEdges != 1 {
		t.Errorf("expected 1 tag edge for 'go', got %d", tagEdges)
	}
}

func TestBuildGraph_LinkEdges(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer store.Close()

	now := time.Now()

	store.IndexDocument(scanner.Document{
		RelPath:     "index.md",
		Frontmatter: map[string]any{"title": "Index"},
		Body:        "See [[guide]] for details.",
		ModTime:     now,
		Size:        50,
	})
	store.IndexDocument(scanner.Document{
		RelPath:     "guide.md",
		Frontmatter: map[string]any{"title": "Guide"},
		Body:        "# Guide",
		ModTime:     now,
		Size:        30,
	})

	graph, err := store.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	var linkEdges int
	for _, e := range graph.Edges {
		if e.Type == "link" {
			linkEdges++
			if e.Source != "index.md" || e.Target != "guide.md" {
				t.Errorf("unexpected link edge: %s -> %s", e.Source, e.Target)
			}
		}
	}
	if linkEdges != 1 {
		t.Errorf("expected 1 link edge, got %d", linkEdges)
	}
}

func TestBuildGraph_Empty(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer store.Close()

	graph, err := store.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(graph.Edges))
	}
}

func TestBuildGraph_LinkToNonexistentDoc(t *testing.T) {
	store, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer store.Close()

	now := time.Now()

	store.IndexDocument(scanner.Document{
		RelPath:     "index.md",
		Frontmatter: map[string]any{"title": "Index"},
		Body:        "Link to [[nonexistent]].",
		ModTime:     now,
		Size:        50,
	})

	graph, err := store.BuildGraph()
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}

	// Should have 1 node and no link edges (target doesn't exist)
	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges (link target doesn't exist), got %d", len(graph.Edges))
	}
}
