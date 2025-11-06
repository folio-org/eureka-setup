package helpers

import (
	"slices"

	"github.com/folio-org/eureka-cli/field"
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
	return slices.Contains(ConvertMapToSlice(configTenants), tenantName)
}
