package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/goccy/go-yaml"
)

// ParseFrontmatter reads YAML frontmatter from a Markdown file.
// It stops reading as soon as the closing --- is found (fast path).
func ParseFrontmatter(r io.Reader) (map[string]any, string, error) {
	scanner := bufio.NewScanner(r)

	// First line must be ---
	if !scanner.Scan() {
		return nil, "", nil
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		// No frontmatter, return everything as body
		var body strings.Builder
		body.WriteString(scanner.Text())
		body.WriteString("\n")
		for scanner.Scan() {
			body.WriteString(scanner.Text())
			body.WriteString("\n")
		}
		return nil, body.String(), scanner.Err()
	}

	// Read until closing ---
	var fmBuilder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		fmBuilder.WriteString(line)
		fmBuilder.WriteString("\n")
	}

	// Parse YAML
	var meta map[string]any
	if fmContent := fmBuilder.String(); fmContent != "" {
		if err := yaml.Unmarshal([]byte(fmContent), &meta); err != nil {
			return nil, "", err
		}
	}

	// Read remaining body
	var body strings.Builder
	for scanner.Scan() {
		body.WriteString(scanner.Text())
		body.WriteString("\n")
	}

	return meta, body.String(), scanner.Err()
}
