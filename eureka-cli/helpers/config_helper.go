package helpers

import (
	"slices"

	"github.com/folio-org/eureka-setup/eureka-cli/field"
)

func IsModuleEnabled(module string, configBackendModules map[string]any) bool {
	value, exists := configBackendModules[module]
	if !exists || value == nil {
		return false
	}

	entry, ok := value.(map[string]any)
	if !ok {
		return false
	}

	deploy, ok := entry[field.ModuleDeployModuleEntry]
	if !ok {
		return true
	}
	enabled, ok := deploy.(bool)

	return ok && enabled
}

func IsUIEnabled(tenantName string, configTenants map[string]any) bool {
	value, exists := configTenants[tenantName]
	if !exists || value == nil {
		return false
	}

	entry, ok := value.(map[string]any)
	if !ok {
		return false
	}
	deploy, ok := entry[field.TenantsDeployUIEntry]
	enabled, isBool := deploy.(bool)

	return ok && isBool && enabled
}

func HasTenant(tenantName string, configTenants map[string]any) bool {
	tenants := ConvertMapKeyToSlice(configTenants)
	return slices.Contains(tenants, tenantName)
}

// GetConfiguredPrivatePort resolves a module's private port from its config entry, preferring the
// port-server compatibility alias over private-port; returns nil when no private port is configured
func GetConfiguredPrivatePort(entry map[string]any) *int {
	// Check if private-port key exists (even if nil)
	rawPrivatePort, privatePortExists := entry[field.ModulePrivatePortEntry]
	if !privatePortExists || rawPrivatePort == nil {
		return nil
	}

	// Check for port-server (compatibility alias) first
	if portServerPtr := GetIntPtr(entry, field.ModulePortServerEntry); portServerPtr != nil {
		return portServerPtr
	}

	return GetIntPtr(entry, field.ModulePrivatePortEntry)
}

func GetBackendModuleNames(configBackendModules map[string]any) []string {
	var names []string
	for name := range configBackendModules {
		names = append(names, name)
	}

	return names
}
