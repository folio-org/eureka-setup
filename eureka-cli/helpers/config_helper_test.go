package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestIsModuleEnabled_EnabledModule(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": map[string]any{
			field.ModuleDeployModuleEntry: true,
		},
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.True(t, result)
}

func TestIsModuleEnabled_DisabledModule(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": map[string]any{
			field.ModuleDeployModuleEntry: false,
		},
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.False(t, result)
}

func TestIsModuleEnabled_ModuleNotExists(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.False(t, result)
}

func TestIsModuleEnabled_NilValue(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": nil,
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.False(t, result)
}

func TestIsModuleEnabled_InvalidEntryType(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": "not-a-map",
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.False(t, result)
}

func TestIsModuleEnabled_NoDeployEntry(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": map[string]any{
			"other-field": "value",
		},
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.True(t, result) // Default is true when deploy entry doesn't exist
}

func TestIsModuleEnabled_InvalidDeployType(t *testing.T) {
	// Arrange
	configBackendModules := map[string]any{
		"mod-users": map[string]any{
			field.ModuleDeployModuleEntry: "true",
		},
	}

	// Act
	result := helpers.IsModuleEnabled("mod-users", configBackendModules)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_EnabledUI(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": map[string]any{
			field.TenantsDeployUIEntry: true,
		},
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.True(t, result)
}

func TestIsUIEnabled_DisabledUI(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": map[string]any{
			field.TenantsDeployUIEntry: false,
		},
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_TenantNotExists(t *testing.T) {
	// Arrange
	configTenants := map[string]any{}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_NilValue(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": nil,
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_InvalidEntryType(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": "not-a-map",
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_NoDeployUIEntry(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": map[string]any{
			"other-field": "value",
		},
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestIsUIEnabled_InvalidDeployUIType(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku": map[string]any{
			field.TenantsDeployUIEntry: "true",
		},
	}

	// Act
	result := helpers.IsUIEnabled("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestHasTenant_TenantExists(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"diku":    map[string]any{},
		"tenant2": map[string]any{},
	}

	// Act
	result := helpers.HasTenant("diku", configTenants)

	// Assert
	assert.True(t, result)
}

func TestHasTenant_TenantNotExists(t *testing.T) {
	// Arrange
	configTenants := map[string]any{
		"tenant1": map[string]any{},
	}

	// Act
	result := helpers.HasTenant("diku", configTenants)

	// Assert
	assert.False(t, result)
}

func TestHasTenant_EmptyMap(t *testing.T) {
	// Arrange
	configTenants := map[string]any{}

	// Act
	result := helpers.HasTenant("diku", configTenants)

	// Assert
	assert.False(t, result)
}
