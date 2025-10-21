package helpers

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestGetConfigEnvVars(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name     string
		key      string
		config   map[string]string
		expected []string
	}{
		{
			name: "basic environment variables",
			key:  "test_env",
			config: map[string]string{
				"var1": "value1",
				"var2": "value2",
			},
			expected: []string{"VAR1=value1", "VAR2=value2"},
		},
		{
			name:     "empty configuration",
			key:      "empty_env",
			config:   map[string]string{},
			expected: []string{},
		},
		{
			name: "single environment variable",
			key:  "single_env",
			config: map[string]string{
				"database_url": "postgres://localhost:5432/db",
			},
			expected: []string{"DATABASE_URL=postgres://localhost:5432/db"},
		},
		{
			name: "mixed case keys",
			key:  "mixed_env",
			config: map[string]string{
				"MixedCase": "value1",
				"lowercase": "value2",
				"UPPERCASE": "value3",
			},
			expected: []string{"MIXEDCASE=value1", "LOWERCASE=value2", "UPPERCASE=value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper
			viper.Reset()

			// Set up test configuration
			viper.Set(tt.key, tt.config)

			result := GetConfigEnvVars(tt.key)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("GetConfigEnvVars() returned %d items, want %d", len(result), len(tt.expected))
			}

			// Convert to map for easier comparison (order doesn't matter)
			resultMap := make(map[string]bool)
			for _, env := range result {
				resultMap[env] = true
			}

			expectedMap := make(map[string]bool)
			for _, env := range tt.expected {
				expectedMap[env] = true
			}

			// Check all expected items are present
			for expected := range expectedMap {
				if !resultMap[expected] {
					t.Errorf("Expected environment variable %q not found in result", expected)
				}
			}

			// Check no unexpected items are present
			for result := range resultMap {
				if !expectedMap[result] {
					t.Errorf("Unexpected environment variable %q found in result", result)
				}
			}
		})
	}
}

func TestGetConfigEnv(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	tests := []struct {
		name     string
		key      string
		envData  map[string]string
		expected string
	}{
		{
			name: "existing environment variable",
			key:  "database_url",
			envData: map[string]string{
				"database_url": "postgres://localhost:5432/db",
				"redis_url":    "redis://localhost:6379",
			},
			expected: "postgres://localhost:5432/db",
		},
		{
			name: "non-existent environment variable",
			key:  "non_existent",
			envData: map[string]string{
				"database_url": "postgres://localhost:5432/db",
			},
			expected: "",
		},
		{
			name:     "empty env configuration",
			key:      "any_key",
			envData:  map[string]string{},
			expected: "",
		},
		{
			name: "case insensitive key lookup",
			key:  "DATABASE_URL",
			envData: map[string]string{
				"database_url": "postgres://localhost:5432/db",
			},
			expected: "postgres://localhost:5432/db",
		},
		{
			name: "mixed case in env data",
			key:  "MixedCase",
			envData: map[string]string{
				"mixedcase": "test_value",
			},
			expected: "test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper
			viper.Reset()

			// Set up test environment configuration
			// Note: We need to set this under the "environment" key as that's what the function expects
			viper.Set("environment", tt.envData)

			result := GetConfigEnv(tt.key)

			if result != tt.expected {
				t.Errorf("GetConfigEnv(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetConfigEnvIntegration(t *testing.T) {
	// Integration test to ensure both functions work together
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	// Set up environment configuration
	envConfig := map[string]string{
		"db_host":    "localhost",
		"db_port":    "5432",
		"db_name":    "testdb",
		"redis_host": "redis-server",
		"redis_port": "6379",
	}

	viper.Reset()
	viper.Set("environment", envConfig)
	viper.Set("test_env", envConfig)

	// Test GetConfigEnv
	dbHost := GetConfigEnv("db_host")
	if dbHost != "localhost" {
		t.Errorf("GetConfigEnv failed: got %q, want %q", dbHost, "localhost")
	}

	// Test GetConfigEnvVars
	envVars := GetConfigEnvVars("test_env")
	if len(envVars) != len(envConfig) {
		t.Errorf("GetConfigEnvVars returned %d variables, want %d", len(envVars), len(envConfig))
	}

	// Check that all environment variables are properly formatted
	for _, envVar := range envVars {
		if !strings.Contains(envVar, "=") {
			t.Errorf("Environment variable %q is not properly formatted", envVar)
		}

		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			t.Errorf("Environment variable %q should have exactly one '=' sign", envVar)
		}

		key := parts[0]
		value := parts[1]

		// Key should be uppercase
		if key != strings.ToUpper(key) {
			t.Errorf("Environment variable key %q should be uppercase", key)
		}

		// Value should match original config (case insensitive key lookup)
		originalValue := envConfig[strings.ToLower(key)]
		if value != originalValue {
			t.Errorf("Environment variable %q has value %q, want %q", key, value, originalValue)
		}
	}
}
