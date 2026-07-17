package helpers

import (
	"fmt"
	"strings"
)

func ToGatewayVersion(pristineVersion string) string {
    return strings.ReplaceAll(pristineVersion, "+", "-")
}

// ToSyntheticID converts a raw semantic version into a gateway-compliant identifier
func ToSyntheticID(moduleName, pristineVersion string) string {
	gatewayVersion := ToGatewayVersion(pristineVersion)
	return fmt.Sprintf("%s-%s", moduleName, gatewayVersion)
}

// ShouldEmbedDescriptor determines if a module's full JSON should be packed inline
func ShouldEmbedDescriptor(globalFetch bool, isLocal bool, moduleID string, cachedDescriptors map[string]any) bool {
	if globalFetch || isLocal {
		return true
	}
	if _, exists := cachedDescriptors[moduleID]; exists {
		return true
	}
	return false
}

// IsStrictModuleID ensures module matching behaves strictly by verifying the remainder starts with a semantic digit
func IsStrictModuleID(id string, name string) bool {
	prefix := name + "-"
	if !strings.HasPrefix(id, prefix) {
		return false
	}
	remainder := strings.TrimPrefix(id, prefix)
	if len(remainder) == 0 {
		return false
	}
	return remainder[0] >= '0' && remainder[0] <= '9'
}

// GetDefaultTenant extracts the active tenant name from profile definitions or defaults to diku
func GetDefaultTenant(tenants map[string]any) string {
	for k := range tenants {
		if k != "" {
			return k
		}
	}
	return "diku"
}

// ResolveUIPlatformTag unifies custom workspace image tracking formats across all service bounds
func ResolveUIPlatformTag(namespace, tenantName string) string {
	if tenantName == "" {
		tenantName = "diku"
	}
	return fmt.Sprintf("%s/platform-lsp-ui-%s:latest", namespace, tenantName)
}