package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/esakat/markdown-kb/internal/scanner"

	_ "modernc.org/sqlite"
)

// Store provides full-text search and metadata indexing using SQLite FTS5.
type Store struct {
	db *sql.DB
}

// SearchResult represents a single search hit.
type SearchResult struct {
	Path    string         `json:"path"`
	Title   string         `json:"title"`
	Snippet string         `json:"snippet"`
	Score   float64        `json:"score"`
	Meta    map[string]any `json:"meta"`
}

// DocumentSummary represents a document's metadata without body.
type DocumentSummary struct {
	Path    string         `json:"path"`
	Title   string         `json:"title"`
	Meta    map[string]any `json:"meta"`
	ModTime time.Time      `json:"mod_time"`
	Size    int64          `json:"size"`
}

// DocumentDetail represents a full document including body.
type DocumentDetail struct {
	DocumentSummary
	Body string `json:"body"`
}

// TagCount represents a tag with its occurrence count.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// MetadataField represents a discovered frontmatter field.
type MetadataField struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

const schema = `
CREATE TABLE IF NOT EXISTS documents (
    path     TEXT PRIMARY KEY,
    title    TEXT,
    meta     TEXT,
    body     TEXT,
    mod_time TEXT,
    size     INTEGER
);

CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
    path,
    title,
    body,
    meta,
    tokenize='trigram'
);
`

func openDB(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// For in-memory databases, restrict to a single connection so all
	// operations share the same database instance. Without this,
	// database/sql may open multiple connections to :memory:, each
	// with its own empty database.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return &Store{db: db}, nil
}

// New creates a new index store with an in-memory SQLite database.
func New() (*Store, error) {
	return openDB(":memory:")
}

// NewWithPath creates a new index store backed by a file.
func NewWithPath(path string) (*Store, error) {
	return openDB(path)
}

// IndexDocument adds or updates a document in the index (UPSERT).
func (s *Store) IndexDocument(doc scanner.Document) error {
	title, _ := doc.Frontmatter["title"].(string)

	metaJSON, err := json.Marshal(doc.Frontmatter)
	if err != nil {
		metaJSON = []byte("{}")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing FTS entry
	tx.Exec("DELETE FROM documents_fts WHERE path = ?", doc.RelPath)

	// UPSERT into documents table
	_, err = tx.Exec(`
		INSERT INTO documents (path, title, meta, body, mod_time, size)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			title = excluded.title,
			meta = excluded.meta,
			body = excluded.body,
			mod_time = excluded.mod_time,
			size = excluded.size
	`, doc.RelPath, title, string(metaJSON), doc.Body, doc.ModTime.Format(time.RFC3339), doc.Size)
	if err != nil {
		return fmt.Errorf("upserting document: %w", err)
	}

	// Insert into FTS index
	_, err = tx.Exec(`
		INSERT INTO documents_fts (path, title, body, meta)
		VALUES (?, ?, ?, ?)
	`, doc.RelPath, title, doc.Body, string(metaJSON))
	if err != nil {
		return fmt.Errorf("inserting FTS entry: %w", err)
	}

	return tx.Commit()
}

// RemoveDocument deletes a document from the index.
func (s *Store) RemoveDocument(path string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM documents WHERE path = ?", path); err != nil {
		return fmt.Errorf("deleting document: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM documents_fts WHERE path = ?", path); err != nil {
		return fmt.Errorf("deleting FTS entry: %w", err)
	}

	return tx.Commit()
}

// Search performs a full-text search and returns matching documents ordered by BM25 score.
func (s *Store) Search(query string, limit, offset int) ([]SearchResult, int, error) {
	return s.SearchWithFilter(query, nil, limit, offset)
}

// SearchWithFilter performs a full-text search with optional metadata filters.
func (s *Store) SearchWithFilter(query string, filters map[string]string, limit, offset int) ([]SearchResult, int, error) {
	if query == "" {
		return nil, 0, nil
	}

	// Build the WHERE clause for metadata filters
	var filterClauses []string
	var filterArgs []any

	for key, val := range filters {
		filterClauses = append(filterClauses, "json_extract(d.meta, ?) LIKE ?")
		filterArgs = append(filterArgs, "$."+key, "%"+val+"%")
	}

	filterSQL := ""
	if len(filterClauses) > 0 {
		filterSQL = " AND " + strings.Join(filterClauses, " AND ")
	}

	// Count total matches
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM documents_fts f
		JOIN documents d ON d.path = f.path
		WHERE documents_fts MATCH ?%s
	`, filterSQL)

	countArgs := append([]any{query}, filterArgs...)
	var total int
	if err := s.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting search results: %w", err)
	}

	// Fetch results with BM25 ranking
	searchQuery := fmt.Sprintf(`
		SELECT d.path, d.title, snippet(documents_fts, 2, '<b>', '</b>', '...', 32) as snippet,
			   bm25(documents_fts) as score, d.meta
		FROM documents_fts f
		JOIN documents d ON d.path = f.path
		WHERE documents_fts MATCH ?%s
		ORDER BY bm25(documents_fts)
		LIMIT ? OFFSET ?
	`, filterSQL)

	searchArgs := append([]any{query}, filterArgs...)
	searchArgs = append(searchArgs, limit, offset)

	rows, err := s.db.Query(searchQuery, searchArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("searching: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var metaJSON string
		if err := rows.Scan(&r.Path, &r.Title, &r.Snippet, &r.Score, &metaJSON); err != nil {
			return nil, 0, fmt.Errorf("scanning result: %w", err)
		}
		json.Unmarshal([]byte(metaJSON), &r.Meta)
		results = append(results, r)
	}

	return results, total, rows.Err()
}

// ListDocumentsWithFilter returns a paginated list of documents filtered by metadata.
// When filters is nil or empty, it returns all documents (same as ListDocuments).
func (s *Store) ListDocumentsWithFilter(filters map[string]string, limit, offset int) ([]DocumentSummary, int, error) {
	var filterClauses []string
	var filterArgs []any

	for key, val := range filters {
		filterClauses = append(filterClauses, "json_extract(meta, ?) LIKE ?")
		filterArgs = append(filterArgs, "$."+key, "%"+val+"%")
	}

	whereSQL := ""
	if len(filterClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(filterClauses, " AND ")
	}

	// Count total matching documents
	var total int
	countQuery := "SELECT COUNT(*) FROM documents" + whereSQL
	if err := s.db.QueryRow(countQuery, filterArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting filtered documents: %w", err)
	}

	// Fetch paginated results
	query := fmt.Sprintf(`
		SELECT path, title, meta, mod_time, size
		FROM documents%s
		ORDER BY path
		LIMIT ? OFFSET ?
	`, whereSQL)

	args := append(filterArgs, limit, offset)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing filtered documents: %w", err)
	}
	defer rows.Close()

	var docs []DocumentSummary
	for rows.Next() {
		var d DocumentSummary
		var metaJSON, modTimeStr string
		if err := rows.Scan(&d.Path, &d.Title, &metaJSON, &modTimeStr, &d.Size); err != nil {
			return nil, 0, fmt.Errorf("scanning document: %w", err)
		}
		json.Unmarshal([]byte(metaJSON), &d.Meta)
		d.ModTime, _ = time.Parse(time.RFC3339, modTimeStr)
		docs = append(docs, d)
	}

	return docs, total, rows.Err()
}

// ListDocuments returns a paginated list of documents.
func (s *Store) ListDocuments(limit, offset int) ([]DocumentSummary, int, error) {
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting documents: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT path, title, meta, mod_time, size
		FROM documents
		ORDER BY path
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing documents: %w", err)
	}
	defer rows.Close()

	var docs []DocumentSummary
	for rows.Next() {
		var d DocumentSummary
		var metaJSON, modTimeStr string
		if err := rows.Scan(&d.Path, &d.Title, &metaJSON, &modTimeStr, &d.Size); err != nil {
			return nil, 0, fmt.Errorf("scanning document: %w", err)
		}
		json.Unmarshal([]byte(metaJSON), &d.Meta)
		d.ModTime, _ = time.Parse(time.RFC3339, modTimeStr)
		docs = append(docs, d)
	}

	return docs, total, rows.Err()
}

// GetDocument retrieves a single document by path.
func (s *Store) GetDocument(path string) (*DocumentDetail, error) {
	var d DocumentDetail
	var metaJSON, modTimeStr string

	err := s.db.QueryRow(`
		SELECT path, title, meta, body, mod_time, size
		FROM documents
		WHERE path = ?
	`, path).Scan(&d.Path, &d.Title, &metaJSON, &d.Body, &modTimeStr, &d.Size)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting document: %w", err)
	}

	json.Unmarshal([]byte(metaJSON), &d.Meta)
	d.ModTime, _ = time.Parse(time.RFC3339, modTimeStr)

	return &d, nil
}

// ListTags returns all tags with their occurrence counts.
func (s *Store) ListTags() ([]TagCount, error) {
	rows, err := s.db.Query("SELECT meta FROM documents WHERE meta IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("querying documents: %w", err)
	}
	defer rows.Close()

	tagCounts := make(map[string]int)
	for rows.Next() {
		var metaJSON string
		if err := rows.Scan(&metaJSON); err != nil {
			continue
		}
		var meta map[string]any
		if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
			continue
		}

		tags, ok := meta["tags"]
		if !ok {
			continue
		}

		switch v := tags.(type) {
		case []any:
			for _, tag := range v {
				if s, ok := tag.(string); ok {
					tagCounts[s]++
				}
			}
		case string:
			tagCounts[v]++
		}
	}

	var result []TagCount
	for tag, count := range tagCounts {
		result = append(result, TagCount{Tag: tag, Count: count})
	}

	return result, rows.Err()
}

// ListMetadataFields returns auto-detected frontmatter fields with their types and sample values.
func (s *Store) ListMetadataFields() ([]MetadataField, error) {
	rows, err := s.db.Query("SELECT meta FROM documents WHERE meta IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("querying documents: %w", err)
	}
	defer rows.Close()

	// field name -> type and unique values
	type fieldInfo struct {
		typ    string
		values map[string]bool
	}
	fields := make(map[string]*fieldInfo)

	for rows.Next() {
		var metaJSON string
		if err := rows.Scan(&metaJSON); err != nil {
			continue
		}
		var meta map[string]any
		if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
			continue
		}

		for key, val := range meta {
			fi, ok := fields[key]
			if !ok {
				fi = &fieldInfo{values: make(map[string]bool)}
				fields[key] = fi
			}

			switch v := val.(type) {
			case []any:
				fi.typ = "array"
				for _, item := range v {
					if s, ok := item.(string); ok && len(fi.values) < 10 {
						fi.values[s] = true
					}
				}
			case string:
				fi.typ = "string"
				if len(fi.values) < 10 {
					fi.values[v] = true
				}
			case float64:
				fi.typ = "number"
				if len(fi.values) < 10 {
					fi.values[fmt.Sprintf("%v", v)] = true
				}
			case bool:
				fi.typ = "boolean"
				if len(fi.values) < 10 {
					fi.values[fmt.Sprintf("%v", v)] = true
				}
			}
		}
	}

	var result []MetadataField
	for name, fi := range fields {
		var values []string
		for v := range fi.values {
			values = append(values, v)
		}
		result = append(result, MetadataField{Name: name, Type: fi.typ, Values: values})
	}

	return result, rows.Err()
}

// Close releases the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
