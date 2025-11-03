package models

// ReindexJobResponse represents the response from a search reindex job operation
type ReindexJobResponse struct {
	ID        string            `json:"id"`
	JobStatus string            `json:"jobStatus"`
	Errors    []ReindexJobError `json:"errors,omitempty"`
}

// ReindexJobError represents an error that occurred during a reindex job
type ReindexJobError struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
