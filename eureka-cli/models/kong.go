package models

// KongRoute represents a Kong API gateway route configuration
type KongRoute struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Protocols  []string `json:"protocols"`
	Methods    []string `json:"methods"`
	Paths      []string `json:"paths"`
	Expression string   `json:"expression"`
	Tags       []string `json:"tags"`
	Service    struct {
		ID string `json:"id"`
	} `json:"service"`
}

// KongRoutesResponse represents the response containing a list of Kong routes
type KongRoutesResponse struct {
	Data []KongRoute `json:"data"`
	Next string      `json:"next,omitempty"`
}
