package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/esakat/markdown-kb/internal/parser"
)

// Document represents a parsed Markdown file discovered by the scanner.
type Document struct {
	RelPath     string         // relative path from root directory
	AbsPath     string         // absolute path
	Frontmatter map[string]any // parsed YAML frontmatter (nil if none or invalid)
	Body        string         // content after frontmatter
	ModTime     time.Time      // file modification time
	Size        int64          // file size in bytes
}

// skipDirs contains directory names that should be skipped during scanning.
var skipDirs = map[string]bool{
	".git":         true,
	".svn":         true,
	".hg":          true,
	"node_modules": true,
}

// Scan recursively walks rootDir and returns all .md files as Documents.
// Results are sorted by RelPath. Symlinks are not followed.
func Scan(rootDir string) ([]Document, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolving root directory: %w", err)
	}

	info, err := os.Lstat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("accessing directory %q: %w", rootDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", rootDir)
	}

	var docs []Document

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		// Skip symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			// Skip known directories and dot-prefixed directories (except root)
			name := d.Name()
			if path != absRoot {
				if skipDirs[name] || strings.HasPrefix(name, ".") {
					return fs.SkipDir
				}
			}
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil // skip on error
		}

		fi, err := d.Info()
		if err != nil {
			return nil
		}

		doc := Document{
			RelPath: relPath,
			AbsPath: path,
			ModTime: fi.ModTime(),
			Size:    fi.Size(),
		}

		// Empty file
		if fi.Size() == 0 {
			docs = append(docs, doc)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		// Skip non-UTF-8 binary files
		if !utf8.Valid(content) {
			return nil
		}

		meta, body, parseErr := parser.ParseFrontmatter(strings.NewReader(string(content)))
		if parseErr != nil {
			// Bad frontmatter: put full content in body, leave frontmatter nil
			doc.Body = string(content)
		} else {
			doc.Frontmatter = meta
			doc.Body = body
		}

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].RelPath < docs[j].RelPath
	})

	return docs, nil
}
