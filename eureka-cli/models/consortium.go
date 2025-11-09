package models

import (
	"fmt"
	"strings"
)

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
	IsCentral bool   `json:"isCentral"`
}

// ConsortiumTenantStatus represents the status of a consortium tenant
type ConsortiumTenantStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsCentral   bool   `json:"isCentral"`
	SetupStatus string `json:"setupStatus"`
}

// SortedConsortiumTenant represents a tenant within a consortium sorted by IsCentral
type SortedConsortiumTenant struct {
	Consortium string
	Name       string
	IsCentral  int
}

// String returns a formatted string representation of the consortium tenant.
// If the tenant is central (IsCentral == 1), it appends "(central)" to the name.
func (c SortedConsortiumTenant) String() string {
	if c.IsCentral == 1 {
		return fmt.Sprintf("%s (central)", c.Name)
	}

	return c.Name
}

// SortedConsortiumTenants represents a collection of consortium tenants sorted by IsCentral
type SortedConsortiumTenants []*SortedConsortiumTenant

// String returns a comma-separated list of tenant names.
func (c SortedConsortiumTenants) String() string {
	var builder strings.Builder
	for i, value := range c {
		builder.WriteString(value.Name)
		if i+1 < len(c) {
			builder.WriteString(", ")
		}
	}

	return builder.String()
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

// ==================== Settings Management ====================

// SettingsResponse represents the response containing a list of settings
type SettingsResponse struct {
	Settings     []Setting `json:"settings"`
	TotalRecords int       `json:"totalRecords,omitempty"`
}

// Setting represents a key-value configuration setting
type Setting struct {
	ID    string `json:"id,omitempty"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
