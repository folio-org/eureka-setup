package models

// KongRoute represents a Kong route structure
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

// KongRoutesResponse represents the response from Kong routes API
type KongRoutesResponse struct {
	Data []KongRoute `json:"data"`
	Next string      `json:"next,omitempty"`
}
