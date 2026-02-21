package index

import (
	"encoding/json"

	"github.com/esakat/markdown-kb/internal/parser"
)

// GraphNode represents a document node in the graph.
type GraphNode struct {
	Path  string   `json:"path"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

// GraphEdge represents a relationship between two documents.
type GraphEdge struct {
	Source string `json:"source"` // path
	Target string `json:"target"` // path
	Type   string `json:"type"`   // "link" or "tag"
	Label  string `json:"label"`  // tag name (for tag edges)
}

// GraphData holds the complete graph representation.
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// BuildGraph builds a graph of documents connected by shared tags and links.
func (s *Store) BuildGraph() (*GraphData, error) {
	rows, err := s.db.Query("SELECT path, title, meta, body FROM documents ORDER BY path")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type docInfo struct {
		path  string
		title string
		tags  []string
		links []string
	}

	var docs []docInfo
	pathSet := make(map[string]bool)

	for rows.Next() {
		var path, title, metaJSON, body string
		if err := rows.Scan(&path, &title, &metaJSON, &body); err != nil {
			continue
		}

		var meta map[string]any
		json.Unmarshal([]byte(metaJSON), &meta)

		var tags []string
		if rawTags, ok := meta["tags"]; ok {
			switch v := rawTags.(type) {
			case []any:
				for _, t := range v {
					if s, ok := t.(string); ok {
						tags = append(tags, s)
					}
				}
			case string:
				tags = []string{v}
			}
		}

		links := parser.ExtractLinks(body)

		docs = append(docs, docInfo{
			path:  path,
			title: title,
			tags:  tags,
			links: links,
		})
		pathSet[path] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build nodes
	nodes := make([]GraphNode, len(docs))
	for i, d := range docs {
		tags := d.tags
		if tags == nil {
			tags = []string{}
		}
		nodes[i] = GraphNode{
			Path:  d.path,
			Title: d.title,
			Tags:  tags,
		}
	}

	// Build edges
	var edges []GraphEdge
	edgeSeen := make(map[string]bool)

	addEdge := func(e GraphEdge) {
		// Normalize edge key (undirected for tags)
		key := e.Source + "|" + e.Target + "|" + e.Type + "|" + e.Label
		if !edgeSeen[key] {
			edgeSeen[key] = true
			edges = append(edges, e)
		}
	}

	// Link edges: doc A links to doc B
	for _, d := range docs {
		for _, link := range d.links {
			if pathSet[link] && link != d.path {
				addEdge(GraphEdge{
					Source: d.path,
					Target: link,
					Type:   "link",
				})
			}
		}
	}

	// Tag edges: docs that share the same tag
	tagToDocs := make(map[string][]string)
	for _, d := range docs {
		for _, tag := range d.tags {
			tagToDocs[tag] = append(tagToDocs[tag], d.path)
		}
	}
	for tag, paths := range tagToDocs {
		for i := 0; i < len(paths); i++ {
			for j := i + 1; j < len(paths); j++ {
				src, tgt := paths[i], paths[j]
				if src > tgt {
					src, tgt = tgt, src
				}
				addEdge(GraphEdge{
					Source: src,
					Target: tgt,
					Type:   "tag",
					Label:  tag,
				})
			}
		}
	}

	if nodes == nil {
		nodes = []GraphNode{}
	}
	if edges == nil {
		edges = []GraphEdge{}
	}

	return &GraphData{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
