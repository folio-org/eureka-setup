package helpers

import (
	"fmt"
	"strings"
)

// ImageOverrideConfig encapsulates user profile configuration definitions
type ImageOverrideConfig struct {
	Image    string
	Registry string
}

// ExtractImageOverride handles type-safe extraction of properties from unstructured configuration maps
func ExtractImageOverride(backendConfigs map[string]any, moduleName string) ImageOverrideConfig {
	var override ImageOverrideConfig
	rawModuleConfig, exists := backendConfigs[moduleName]
	if !exists {
		return override
	}

	moduleMap, ok := rawModuleConfig.(map[string]any)
	if !ok {
		return override
	}

	if img, ok := moduleMap["image"].(string); ok {
		override.Image = strings.TrimSpace(img)
	}
	if reg, ok := moduleMap["registry"].(string); ok {
		override.Registry = strings.TrimSpace(reg)
	}
	return override
}

// ExtractPrivatePort safely looks up a module's custom configuration map to retrieve its port
func ExtractPrivatePort(backendConfigs map[string]any, moduleName string) *int {
	rawModuleConfig, exists := backendConfigs[moduleName]
	if !exists {
		return nil
	}

	moduleMap, ok := rawModuleConfig.(map[string]any)
	if !ok {
		return nil
	}

	// Natively calls your package-level helper function
	return GetConfiguredPrivatePort(moduleMap)
}

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
	moduleSpecificRegistry := strings.TrimSpace(override.Registry)

	if moduleSpecificRegistry != "" {
		return fmt.Sprintf("%s/%s:%s", moduleSpecificRegistry, repoPath, cleanTag), true
	}

	return fmt.Sprintf("%s:%s", repoPath, cleanTag), true
}