package helpers

import (
	"fmt"
	"strings"
)

// ImageOverrideConfig encapsulates user profile configuration definitions
type ImageOverrideConfig struct {
	Image         string
	ImageRegistry string
}

// ==================== Gateway, Identity & Selection Logic ====================

// ToGatewayVersion normalizes versions by replacing illegal characters
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

// ==================== Tenant & UI Platform Tracking Logic ====================

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

// ==================== Internal Boilerplate Deduplication ====================

// Unified lookup and type-assertion to eliminate boilerplate duplication across delegators
func getModuleMap(modulesConfig map[string]any, moduleName string) (map[string]any, bool) {
	if modulesConfig == nil {
		return nil, false
	}
	rawModuleConfig, exists := modulesConfig[moduleName]
	if !exists {
		return nil, false
	}
	moduleMap, ok := rawModuleConfig.(map[string]any)
	return moduleMap, ok
}

// ==================== FromMap Context-Isolated Extractors ====================

// ExtractImageOverrideFromMap extracts properties directly from an isolated module map context
func ExtractImageOverrideFromMap(moduleMap map[string]any) ImageOverrideConfig {
	var override ImageOverrideConfig
	if moduleMap == nil {
		return override
	}
	if img, ok := moduleMap["image"].(string); ok {
		override.Image = strings.TrimSpace(img)
	}
	if imgReg, ok := moduleMap["image-registry"].(string); ok {
		override.ImageRegistry = strings.TrimSpace(imgReg)
	}
	return override
}

// ExtractDescriptorRegistryFromMap extracts the registry URL directly from an isolated module map context
func ExtractDescriptorRegistryFromMap(moduleMap map[string]any) string {
	if moduleMap == nil {
		return ""
	}
	if reg, ok := moduleMap["registry"].(string); ok {
		return strings.TrimSpace(reg)
	}
	return ""
}

// ExtractPrivatePortFromMap safely extracts private port targets from an isolated module map context
func ExtractPrivatePortFromMap(moduleMap map[string]any) *int {
	if moduleMap == nil {
		return nil
	}
	return GetConfiguredPrivatePort(moduleMap)
}

// ==================== Global Map Top-Level Delegators ====================

// ExtractImageOverride handles type-safe extraction of image properties from a global map
func ExtractImageOverride(modulesConfig map[string]any, moduleName string) ImageOverrideConfig {
	if moduleMap, ok := getModuleMap(modulesConfig, moduleName); ok {
		return ExtractImageOverrideFromMap(moduleMap)
	}
	return ImageOverrideConfig{}
}

// ExtractDescriptorRegistry handles type-safe extraction of the Okapi descriptor registry URL from a global map
func ExtractDescriptorRegistry(modulesConfig map[string]any, moduleName string) string {
	if moduleMap, ok := getModuleMap(modulesConfig, moduleName); ok {
		return ExtractDescriptorRegistryFromMap(moduleMap)
	}
	return ""
}

// ExtractPrivatePort safely looks up a module's custom configuration from a global map
func ExtractPrivatePort(modulesConfig map[string]any, moduleName string) *int {
	if moduleMap, ok := getModuleMap(modulesConfig, moduleName); ok {
		return ExtractPrivatePortFromMap(moduleMap)
	}
	return nil
}

// ==================== Structural Parsing Engines ====================

// ResolveRepoPath isolates the clean image repository path by stripping domain/port prefixes if present
func ResolveRepoPath(moduleName, customImage string) string {
	repoPath := moduleName
	if customImage != "" {
		repoPath = customImage
	}
	if idx := strings.Index(repoPath, "/"); idx != -1 {
		firstPart := repoPath[:idx]
		if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") {
			repoPath = repoPath[idx+1:]
		}
	}
	return repoPath
}

// SanitizeModuleImage reconciles framework version rules against custom profile overrides
func SanitizeModuleImage(fallbackImage, version string, backupVersion *string, override ImageOverrideConfig) (string, bool) {
	customImage := strings.TrimSpace(override.Image)
	if customImage == "" {
		return fallbackImage, false
	}
	if strings.Contains(customImage, ":") {
		return customImage, true
	}

	tagSource := version
	if tagSource == "" && backupVersion != nil {
		tagSource = *backupVersion
	}

	cleanTag := tagSource
	if strings.Contains(cleanTag, "+") {
		cleanTag = strings.Split(cleanTag, "+")[0]
	} else if strings.Contains(cleanTag, "-SNAPSHOT.") {
		parts := strings.Split(cleanTag, "-SNAPSHOT.")
		if len(parts) > 1 {
			cleanTag = parts[0] + "-SNAPSHOT." + strings.Split(parts[1], "-")[0]
		}
	}

	repoPath := ResolveRepoPath(fallbackImage, customImage)
	moduleSpecificImageRegistry := strings.TrimSpace(override.ImageRegistry)
	if moduleSpecificImageRegistry != "" {
		return fmt.Sprintf("%s/%s:%s", moduleSpecificImageRegistry, repoPath, cleanTag), true
	}
	return fmt.Sprintf("%s:%s", repoPath, cleanTag), true
}