package parser

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMeta   map[string]any
		wantBody   string
		wantErr    bool
	}{
		{
			name: "valid frontmatter with body",
			input: `---
title: Test Document
status: spec
tags:
  - go
  - markdown
---

# Hello World

Body content here.
`,
			wantMeta: map[string]any{
				"title":  "Test Document",
				"status": "spec",
				"tags":   []any{"go", "markdown"},
			},
			wantBody: "\n# Hello World\n\nBody content here.\n",
		},
		{
			name:     "no frontmatter",
			input:    "# Just a heading\n\nSome content.\n",
			wantMeta: nil,
			wantBody: "# Just a heading\n\nSome content.\n",
		},
		{
			name: "empty frontmatter",
			input: `---
---

Body only.
`,
			wantMeta: nil,
			wantBody: "\nBody only.\n",
		},
		{
			name:     "empty file",
			input:    "",
			wantMeta: nil,
			wantBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, body, err := ParseFrontmatter(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantMeta == nil && meta != nil {
				t.Errorf("expected nil meta, got %v", meta)
			}
			if tt.wantMeta != nil {
				if meta == nil {
					t.Fatal("expected non-nil meta, got nil")
				}
				if meta["title"] != tt.wantMeta["title"] {
					t.Errorf("title = %v, want %v", meta["title"], tt.wantMeta["title"])
				}
				if meta["status"] != tt.wantMeta["status"] {
					t.Errorf("status = %v, want %v", meta["status"], tt.wantMeta["status"])
				}
			}

			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
