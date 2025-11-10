package models

// UserResponse represents the response containing a list of users
type UserResponse struct {
	Users        []User `json:"users"`
	TotalRecords int    `json:"totalRecords"`
}

// User represents a user entity with personal information
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
