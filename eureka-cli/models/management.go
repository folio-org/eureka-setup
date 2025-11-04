package models

// ==================== Tenant Management ====================

// TenantCreateRequest represents the payload for creating a new tenant
type TenantCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TenantsResponse represents the response containing a list of tenants
type TenantsResponse struct {
	Tenants []Tenant `json:"tenants"`
}

// Tenant represents a tenant entity in the management system
type Tenant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ==================== Tenant Entitlement Management ====================

// TenantEntitlementRequest represents the payload for creating or removing tenant entitlements
type TenantEntitlementRequest struct {
	TenantID     string   `json:"tenantId"`
	Applications []string `json:"applications"`
}

// ==================== Application Management ====================

// ApplicationCreateRequest represents the payload for creating a new application with modules and descriptors
type ApplicationCreateRequest struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Version             string              `json:"version"`
	Description         string              `json:"description"`
	Platform            string              `json:"platform"`
	Dependencies        map[string]any      `json:"dependencies"`
	Modules             []ApplicationModule `json:"modules"`
	UIModules           []ApplicationModule `json:"uiModules"`
	ModuleDescriptors   []any               `json:"moduleDescriptors"`
	UIModuleDescriptors []any               `json:"uiModuleDescriptors"`
}

// ApplicationModule represents a module within an application
type ApplicationModule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url,omitempty"`
}

// ApplicationsResponse represents the response containing a list of application descriptors
type ApplicationsResponse struct {
	ApplicationDescriptors []map[string]any `json:"applicationDescriptors"`
	TotalRecords           int              `json:"totalRecords"`
}

// ModuleDiscoveryRequest represents the payload for registering module discovery information
type ModuleDiscoveryRequest struct {
	Discovery []ModuleDiscovery `json:"discovery"`
}

// ModuleDiscoveryResponse represents the response containing a list of module discovery information
type ModuleDiscoveryResponse struct {
	Discovery    []ModuleDiscovery `json:"discovery"`
	TotalRecords int               `json:"totalRecords"`
}

// ModuleDiscovery represents discovery information for a deployed module
type ModuleDiscovery struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Location string `json:"location"`
}
