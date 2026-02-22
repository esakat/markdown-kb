package index

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
)

// TreeNode represents a node in the directory tree.
type TreeNode struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "dir" or "file"
	Path     string      `json:"path,omitempty"`
	Title    string      `json:"title,omitempty"`
	Tags     []string    `json:"tags,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
}

// PathEntry is a lightweight path+title pair for tree building.
type PathEntry struct {
	Path  string
	Title string
	Tags  []string
}

// ListPaths returns all document paths, titles, and tags, ordered by path.
func (s *Store) ListPaths() ([]PathEntry, error) {
	rows, err := s.db.Query("SELECT path, title, meta FROM documents ORDER BY path")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []PathEntry
	for rows.Next() {
		var e PathEntry
		var metaJSON *string
		if err := rows.Scan(&e.Path, &e.Title, &metaJSON); err != nil {
			return nil, err
		}
		if metaJSON != nil {
			e.Tags = extractTags(*metaJSON)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// extractTags parses tags from a JSON meta string.
func extractTags(metaJSON string) []string {
	var meta map[string]any
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		return nil
	}
	raw, ok := meta["tags"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	tags := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			tags = append(tags, s)
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

// BuildTree constructs a nested TreeNode from a flat list of PathEntry.
func BuildTree(entries []PathEntry) *TreeNode {
	root := &TreeNode{Name: "", Type: "dir"}

	for _, e := range entries {
		parts := strings.Split(filepath.ToSlash(e.Path), "/")
		insertNode(root, parts, e)
	}

	sortTree(root)
	return root
}

func insertNode(parent *TreeNode, parts []string, entry PathEntry) {
	if len(parts) == 1 {
		// Leaf file node
		parent.Children = append(parent.Children, &TreeNode{
			Name:  parts[0],
			Type:  "file",
			Path:  entry.Path,
			Title: entry.Title,
			Tags:  entry.Tags,
		})
		return
	}

	// Find or create directory node
	dirName := parts[0]
	var dir *TreeNode
	for _, child := range parent.Children {
		if child.Type == "dir" && child.Name == dirName {
			dir = child
			break
		}
	}
	if dir == nil {
		dir = &TreeNode{Name: dirName, Type: "dir"}
		parent.Children = append(parent.Children, dir)
	}

	insertNode(dir, parts[1:], entry)
}

func sortTree(node *TreeNode) {
	if len(node.Children) == 0 {
		return
	}

	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		// Directories first, then files
		if a.Type != b.Type {
			return a.Type == "dir"
		}
		return a.Name < b.Name
	})

	for _, child := range node.Children {
		if child.Type == "dir" {
			sortTree(child)
		}
	}
}
