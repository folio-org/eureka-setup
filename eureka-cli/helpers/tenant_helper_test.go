package helpers

import (
	"testing"

	"github.com/spf13/viper"
)

func TestHasTenant(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name         string
		tenant       string
		tenantConfig map[string]interface{}
		expected     bool
	}{
		{
			name:   "tenant exists",
			tenant: "test-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"name": "Test Tenant",
				},
				"another-tenant": map[string]interface{}{
					"name": "Another Tenant",
				},
			},
			expected: true,
		},
		{
			name:   "tenant does not exist",
			tenant: "non-existent-tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"name": "Test Tenant",
				},
			},
			expected: false,
		},
		{
			name:         "empty tenant config",
			tenant:       "any-tenant",
			tenantConfig: map[string]interface{}{},
			expected:     false,
		},
		{
			name:   "tenant exists with nil value",
			tenant: "nil-tenant",
			tenantConfig: map[string]interface{}{
				"nil-tenant": nil,
				"valid-tenant": map[string]interface{}{
					"name": "Valid Tenant",
				},
			},
			expected: true, // Should still return true as the key exists
		},
		{
			name:   "tenant exists with empty string value",
			tenant: "empty-tenant",
			tenantConfig: map[string]interface{}{
				"empty-tenant": "",
				"valid-tenant": map[string]interface{}{
					"name": "Valid Tenant",
				},
			},
			expected: true, // Should still return true as the key exists
		},
		{
			name:   "multiple tenants check",
			tenant: "tenant2",
			tenantConfig: map[string]interface{}{
				"tenant1": map[string]interface{}{"name": "Tenant 1"},
				"tenant2": map[string]interface{}{"name": "Tenant 2"},
				"tenant3": map[string]interface{}{"name": "Tenant 3"},
			},
			expected: true,
		},
		{
			name:   "case sensitive check",
			tenant: "Test-Tenant",
			tenantConfig: map[string]interface{}{
				"test-tenant": map[string]interface{}{
					"name": "Test Tenant",
				},
			},
			expected: false, // Should be case sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper and set up test configuration
			viper.Reset()
			viper.Set("tenants", tt.tenantConfig)

			result := HasTenant(tt.tenant)
			if result != tt.expected {
				t.Errorf("HasTenant(%q) = %v, want %v", tt.tenant, result, tt.expected)
			}
		})
	}
}

func TestHasTenantIntegration(t *testing.T) {
	// Integration test to verify HasTenant works with actual tenant configurations
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	// Set up a realistic tenant configuration
	tenantConfig := map[string]interface{}{
		"diku": map[string]interface{}{
			"name":      "Datalogisk Institut",
			"deploy_ui": true,
		},
		"testlib": map[string]interface{}{
			"name":      "Test Library",
			"deploy_ui": false,
		},
		"college": map[string]interface{}{
			"name": "College Library",
		},
	}

	viper.Reset()
	viper.Set("tenants", tenantConfig)

	// Test existing tenants
	existingTenants := []string{"diku", "testlib", "college"}
	for _, tenant := range existingTenants {
		if !HasTenant(tenant) {
			t.Errorf("Expected tenant %q to exist", tenant)
		}
	}

	// Test non-existing tenants
	nonExistingTenants := []string{"nonexistent", "DIKU", "TestLib", ""}
	for _, tenant := range nonExistingTenants {
		if HasTenant(tenant) {
			t.Errorf("Expected tenant %q to not exist", tenant)
		}
	}
}
