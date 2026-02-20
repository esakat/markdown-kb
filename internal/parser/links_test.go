package parser

import (
	"sort"
	"testing"
)

func TestExtractLinks_WikiLinks(t *testing.T) {
	body := "See [[guide]] for details and also [[setup/install]]."
	links := ExtractLinks(body)

	sort.Strings(links)
	expected := []string{"guide.md", "setup/install.md"}
	if len(links) != len(expected) {
		t.Fatalf("expected %d links, got %d: %v", len(expected), len(links), links)
	}
	for i, l := range links {
		if l != expected[i] {
			t.Errorf("link[%d] = %q, want %q", i, l, expected[i])
		}
	}
}

func TestExtractLinks_MarkdownLinks(t *testing.T) {
	body := "Read the [guide](guide.md) and [setup](docs/setup.md)."
	links := ExtractLinks(body)

	sort.Strings(links)
	expected := []string{"docs/setup.md", "guide.md"}
	if len(links) != len(expected) {
		t.Fatalf("expected %d links, got %d: %v", len(expected), len(links), links)
	}
	for i, l := range links {
		if l != expected[i] {
			t.Errorf("link[%d] = %q, want %q", i, l, expected[i])
		}
	}
}

func TestExtractLinks_Mixed(t *testing.T) {
	body := "See [[overview]] and [API docs](api/readme.md) for more."
	links := ExtractLinks(body)

	sort.Strings(links)
	expected := []string{"api/readme.md", "overview.md"}
	if len(links) != len(expected) {
		t.Fatalf("expected %d links, got %d: %v", len(expected), len(links), links)
	}
	for i, l := range links {
		if l != expected[i] {
			t.Errorf("link[%d] = %q, want %q", i, l, expected[i])
		}
	}
}

func TestExtractLinks_IgnoresExternalURLs(t *testing.T) {
	body := "Visit [Google](https://google.com) and [local](local.md)."
	links := ExtractLinks(body)

	if len(links) != 1 || links[0] != "local.md" {
		t.Errorf("expected [local.md], got %v", links)
	}
}

func TestExtractLinks_IgnoresAnchors(t *testing.T) {
	body := "See [section](#heading) and [[real-page]]."
	links := ExtractLinks(body)

	if len(links) != 1 || links[0] != "real-page.md" {
		t.Errorf("expected [real-page.md], got %v", links)
	}
}

func TestExtractLinks_NoDuplicates(t *testing.T) {
	body := "Link to [[guide]] and again [[guide]]."
	links := ExtractLinks(body)

	if len(links) != 1 || links[0] != "guide.md" {
		t.Errorf("expected [guide.md], got %v", links)
	}
}

func TestExtractLinks_Empty(t *testing.T) {
	links := ExtractLinks("No links here.")
	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}

func TestExtractLinks_WikiLinkWithExtension(t *testing.T) {
	body := "See [[guide.md]] for details."
	links := ExtractLinks(body)
	if len(links) != 1 || links[0] != "guide.md" {
		t.Errorf("expected [guide.md], got %v", links)
	}
}
