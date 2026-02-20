package parser

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// [[wiki-link]] pattern
	wikiLinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	// [text](path.md) pattern — captures the URL part
	mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
)

// ExtractLinks parses a Markdown body and returns unique local document paths.
// It recognizes both [[wiki-link]] and [text](path.md) formats.
// External URLs (http/https), anchors (#), and non-.md links are excluded.
// Wiki-links without .md extension get it appended automatically.
func ExtractLinks(body string) []string {
	seen := make(map[string]bool)
	var result []string

	add := func(path string) {
		path = filepath.ToSlash(path)
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}

	// Extract [[wiki-links]]
	for _, match := range wikiLinkRe.FindAllStringSubmatch(body, -1) {
		link := strings.TrimSpace(match[1])
		if link == "" {
			continue
		}
		if !strings.HasSuffix(link, ".md") {
			link += ".md"
		}
		add(link)
	}

	// Extract [text](path.md) links
	for _, match := range mdLinkRe.FindAllStringSubmatch(body, -1) {
		href := strings.TrimSpace(match[2])
		if href == "" {
			continue
		}
		// Skip external URLs
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			continue
		}
		// Skip anchors
		if strings.HasPrefix(href, "#") {
			continue
		}
		// Only include .md files
		if !strings.HasSuffix(href, ".md") {
			continue
		}
		add(href)
	}

	return result
}
