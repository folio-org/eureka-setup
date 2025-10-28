package models

// ReindexJobResponse represents the response from a reindex job request
type ReindexJobResponse struct {
	ID        string            `json:"id"`
	JobStatus string            `json:"jobStatus"`
	Errors    []ReindexJobError `json:"errors,omitempty"`
}

// ReindexJobError represents an error in a reindex job response
type ReindexJobError struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}
