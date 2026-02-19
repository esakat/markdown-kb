package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("failed to resolve testdata: %v", err)
	}
	return abs
}

func TestScan_NormalFiles(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Collect relative paths
	paths := make(map[string]Document)
	for _, d := range docs {
		paths[d.RelPath] = d
	}

	// Should find files in root and nested docs/
	wantPaths := []string{
		"with_frontmatter.md",
		"no_frontmatter.md",
		"empty.md",
		"bad_frontmatter.md",
		filepath.Join("docs", "nested.md"),
	}
	for _, p := range wantPaths {
		if _, ok := paths[p]; !ok {
			t.Errorf("expected to find %q in scan results", p)
		}
	}
}

func TestScan_SkipDirectories(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	for _, d := range docs {
		dir := filepath.Dir(d.RelPath)
		if dir == ".git" || dir == "node_modules" || dir == ".obsidian" {
			t.Errorf("should have skipped %q, but found in results", d.RelPath)
		}
	}
}

func TestScan_FrontmatterParsing(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	byPath := make(map[string]Document)
	for _, d := range docs {
		byPath[d.RelPath] = d
	}

	// with_frontmatter.md should have parsed frontmatter
	t.Run("with frontmatter", func(t *testing.T) {
		d, ok := byPath["with_frontmatter.md"]
		if !ok {
			t.Fatal("with_frontmatter.md not found")
		}
		if d.Frontmatter == nil {
			t.Fatal("expected non-nil frontmatter")
		}
		if d.Frontmatter["title"] != "Test Document" {
			t.Errorf("title = %v, want %q", d.Frontmatter["title"], "Test Document")
		}
		if d.Frontmatter["status"] != "published" {
			t.Errorf("status = %v, want %q", d.Frontmatter["status"], "published")
		}
		if d.Body == "" {
			t.Error("expected non-empty body")
		}
	})

	// no_frontmatter.md should have nil frontmatter and full content as body
	t.Run("no frontmatter", func(t *testing.T) {
		d, ok := byPath["no_frontmatter.md"]
		if !ok {
			t.Fatal("no_frontmatter.md not found")
		}
		if d.Frontmatter != nil {
			t.Errorf("expected nil frontmatter, got %v", d.Frontmatter)
		}
		if d.Body == "" {
			t.Error("expected non-empty body")
		}
	})

	// bad_frontmatter.md should have nil frontmatter, full content as body
	t.Run("bad frontmatter", func(t *testing.T) {
		d, ok := byPath["bad_frontmatter.md"]
		if !ok {
			t.Fatal("bad_frontmatter.md not found")
		}
		if d.Frontmatter != nil {
			t.Errorf("expected nil frontmatter for bad YAML, got %v", d.Frontmatter)
		}
		if d.Body == "" {
			t.Error("expected non-empty body for bad frontmatter file")
		}
	})
}

func TestScan_EmptyFile(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	byPath := make(map[string]Document)
	for _, d := range docs {
		byPath[d.RelPath] = d
	}

	d, ok := byPath["empty.md"]
	if !ok {
		t.Fatal("empty.md not found")
	}
	if d.Frontmatter != nil {
		t.Errorf("expected nil frontmatter for empty file, got %v", d.Frontmatter)
	}
	if d.Body != "" {
		t.Errorf("expected empty body, got %q", d.Body)
	}
	if d.Size != 0 {
		t.Errorf("expected size 0, got %d", d.Size)
	}
}

func TestScan_NonExistentDirectory(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestScan_DocumentMetadata(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	for _, d := range docs {
		// AbsPath should be absolute
		if !filepath.IsAbs(d.AbsPath) {
			t.Errorf("AbsPath %q should be absolute", d.AbsPath)
		}

		// RelPath should be relative
		if filepath.IsAbs(d.RelPath) {
			t.Errorf("RelPath %q should be relative", d.RelPath)
		}

		// ModTime should not be zero (file exists)
		if d.ModTime.IsZero() {
			t.Errorf("ModTime should not be zero for %q", d.RelPath)
		}
	}
}

func TestScan_NotADirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "notadir.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Scan(tmpFile)
	if err == nil {
		t.Error("expected error when scanning a file (not directory), got nil")
	}
}

func TestScan_BinaryFileSkipped(t *testing.T) {
	tmp := t.TempDir()
	// Write a .md file with invalid UTF-8 bytes
	if err := os.WriteFile(filepath.Join(tmp, "binary.md"), []byte{0xff, 0xfe, 0x00, 0x01}, 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a valid .md file
	if err := os.WriteFile(filepath.Join(tmp, "valid.md"), []byte("# Valid"), 0o644); err != nil {
		t.Fatal(err)
	}

	docs, err := Scan(tmp)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc (binary skipped), got %d", len(docs))
	}
	if docs[0].RelPath != "valid.md" {
		t.Errorf("expected valid.md, got %q", docs[0].RelPath)
	}
}

func TestScan_SymlinkNotFollowed(t *testing.T) {
	tmp := t.TempDir()
	realDir := filepath.Join(tmp, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "doc.md"), []byte("# Doc"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to the directory
	linkDir := filepath.Join(tmp, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skip("symlinks not supported on this system")
	}

	docs, err := Scan(tmp)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only find doc.md under real/, not under link/
	for _, d := range docs {
		if filepath.Dir(d.RelPath) == "link" {
			t.Errorf("should not follow symlink, found %q", d.RelPath)
		}
	}
}

func TestScan_NonMdFilesIgnored(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("# Read"), 0o644)
	os.WriteFile(filepath.Join(tmp, "notes.txt"), []byte("text"), 0o644)
	os.WriteFile(filepath.Join(tmp, "image.png"), []byte("png"), 0o644)

	docs, err := Scan(tmp)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 .md file, got %d", len(docs))
	}
}

func TestScan_ResultsAreSorted(t *testing.T) {
	docs, err := Scan(testdataDir(t))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	paths := make([]string, len(docs))
	for i, d := range docs {
		paths[i] = d.RelPath
	}

	if !sort.StringsAreSorted(paths) {
		t.Errorf("results should be sorted by RelPath, got %v", paths)
	}
}
