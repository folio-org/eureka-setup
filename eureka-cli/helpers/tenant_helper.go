package helpers

import "slices"

func HasTenant(tenantName string, configTenants map[string]any) bool {
	return slices.Contains(ConvertMapToSlice(configTenants), tenantName)
}
