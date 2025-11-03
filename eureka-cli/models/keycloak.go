package models

// ==================== User Management ====================

// KeycloakUserCreateRequest represents the payload for creating a new Keycloak user
type KeycloakUserCreateRequest struct {
	Username string                   `json:"username"`
	Active   bool                     `json:"active"`
	Type     string                   `json:"type"`
	Personal KeycloakUserPersonalInfo `json:"personal"`
}

// KeycloakUserPersonalInfo represents personal information for a Keycloak user
type KeycloakUserPersonalInfo struct {
	FirstName              string `json:"firstName"`
	LastName               string `json:"lastName"`
	Email                  string `json:"email"`
	PreferredContactTypeID string `json:"preferredContactTypeId"`
}

// KeycloakUserPasswordRequest represents the payload for setting a user's password
type KeycloakUserPasswordRequest struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// KeycloakUserRoleRequest represents the payload for assigning roles to a user
type KeycloakUserRoleRequest struct {
	UserID  string   `json:"userId"`
	RoleIDs []string `json:"roleIds"`
}

// KeycloakUserCreateResponse represents the response when creating a new Keycloak user
type KeycloakUserCreateResponse struct {
	ID string `json:"id"`
}

// KeycloakUsersResponse represents the response containing a list of Keycloak users
type KeycloakUsersResponse struct {
	Users        []KeycloakUser `json:"users"`
	TotalRecords int            `json:"totalRecords,omitempty"`
}

// KeycloakUser represents a Keycloak user entity
type KeycloakUser struct {
	ID       string         `json:"id"`
	Username string         `json:"username"`
	Active   bool           `json:"active"`
	Type     string         `json:"type,omitempty"`
	Personal map[string]any `json:"personal,omitempty"`
}

// ==================== Role Management ====================

// KeycloakRoleCreateRequest represents the payload for creating a new Keycloak role
type KeycloakRoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// KeycloakRolesResponse represents the response containing a list of Keycloak roles
type KeycloakRolesResponse struct {
	Roles      []KeycloakRole `json:"roles"`
	TotalCount int            `json:"totalRecords,omitempty"`
}

// KeycloakRole represents a Keycloak role entity
type KeycloakRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ==================== Capability Set Management ====================

// KeycloakCapabilitySetRequest represents the payload for assigning capability sets to a role
type KeycloakCapabilitySetRequest struct {
	RoleID           string   `json:"roleId"`
	CapabilitySetIDs []string `json:"capabilitySetIds"`
}

// KeycloakCapabilitySetsResponse represents the response containing a list of capability sets
type KeycloakCapabilitySetsResponse struct {
	CapabilitySets []KeycloakCapabilitySet `json:"capabilitySets"`
	TotalCount     int                     `json:"totalRecords,omitempty"`
}

// KeycloakCapabilitySet represents a capability set entity
type KeycloakCapabilitySet struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ApplicationID string `json:"applicationId,omitempty"`
}

// ==================== Client Configuration ====================

// KeycloakClientUpdateRequest represents the payload for updating a Keycloak client configuration
type KeycloakClientUpdateRequest struct {
	RootURL                      string            `json:"rootUrl"`
	BaseURL                      string            `json:"baseUrl"`
	AdminURL                     string            `json:"adminUrl"`
	RedirectURIs                 []string          `json:"redirectUris"`
	WebOrigins                   []string          `json:"webOrigins"`
	AuthorizationServicesEnabled bool              `json:"authorizationServicesEnabled"`
	ServiceAccountsEnabled       bool              `json:"serviceAccountsEnabled"`
	Attributes                   map[string]string `json:"attributes"`
}

// ==================== OAuth Token Responses ====================

// KeycloakTokenResponse represents the OAuth token response from Keycloak
type KeycloakTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
}
