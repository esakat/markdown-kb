package api

// Response is the standard API response envelope.
type Response[T any] struct {
	Data  T      `json:"data"`
	Error string `json:"error,omitempty"`
}

// PaginatedResponse wraps paginated API results.
type PaginatedResponse[T any] struct {
	Data  []T  `json:"data"`
	Total int  `json:"total"`
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
}
