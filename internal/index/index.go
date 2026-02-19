package index

import (
	"database/sql"
	"fmt"
)

// Store provides full-text search and metadata indexing using SQLite FTS5.
type Store struct {
	db *sql.DB
}

// New creates a new index store with an in-memory SQLite database.
func New() (*Store, error) {
	// TODO: implement SQLite FTS5 initialization
	return &Store{}, nil
}

// NewWithPath creates a new index store backed by a file.
func NewWithPath(path string) (*Store, error) {
	_ = path
	_ = fmt.Sprintf("placeholder")
	// TODO: implement file-backed SQLite
	return &Store{}, nil
}

// IndexDocument adds or updates a document in the index.
func (s *Store) IndexDocument(path string, meta map[string]any, body string) error {
	// TODO: implement
	_ = path
	_ = meta
	_ = body
	return nil
}

// Search performs a full-text search and returns matching documents.
func (s *Store) Search(query string) ([]SearchResult, error) {
	// TODO: implement FTS5 search with BM25 ranking
	_ = query
	return nil, nil
}

// SearchResult represents a single search hit.
type SearchResult struct {
	Path      string
	Title     string
	Snippet   string
	Score     float64
	Meta      map[string]any
}

// Close releases the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
