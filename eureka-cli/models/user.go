package models

// UserResponse represents the response from the users API
type UserResponse struct {
	Users        []User `json:"users"`
	TotalRecords int    `json:"totalRecords"`
}

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Active   bool   `json:"active"`
	Type     string `json:"type"`
	Personal *struct {
		FirstName              string `json:"firstName"`
		LastName               string `json:"lastName"`
		Email                  string `json:"email"`
		PreferredContactTypeId string `json:"preferredContactTypeId"`
	} `json:"personal,omitempty"`
}
