package models

// ==================== Consortium Management ====================

// ConsortiumCreateRequest represents the payload for creating a new consortium
type ConsortiumCreateRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Consortium represents a consortium entity
type Consortium struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ConsortiumResponse represents the response containing a list of consortia
type ConsortiumResponse struct {
	Consortia []Consortium `json:"consortia"`
}

// ==================== Consortium Tenant Management ====================

// ConsortiumTenantCreateRequest represents the payload for creating a tenant within a consortium
type ConsortiumTenantCreateRequest struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsCentral int    `json:"isCentral"`
}

// ConsortiumTenantsResponse represents the response containing a list of consortium tenants
type ConsortiumTenantsResponse struct {
	Tenants []ConsortiumTenant `json:"tenants"`
}

// ConsortiumTenant represents a tenant within a consortium
type ConsortiumTenant struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsCentral int    `json:"isCentral"`
}

// ConsortiumTenantStatus represents the status of a consortium tenant
type ConsortiumTenantStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsCentral   int    `json:"isCentral"`
	SetupStatus string `json:"setupStatus"`
}

// ==================== Central Ordering Settings ====================

// CentralOrderingSettingRequest represents the payload for creating or updating a central ordering setting
type CentralOrderingSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CentralOrderingSettingsResponse represents the response containing a list of central ordering settings
type CentralOrderingSettingsResponse struct {
	Settings []CentralOrderingSetting `json:"settings"`
}

// CentralOrderingSetting represents a central ordering configuration setting
type CentralOrderingSetting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
