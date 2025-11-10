package testhelpers

import (
	"github.com/spf13/viper"
)

// ViperTestConfig helps manage viper configuration during tests
type ViperTestConfig struct {
	originalValues map[string]any
}

// NewViperTestConfig creates a new viper test configuration manager
func NewViperTestConfig() *ViperTestConfig {
	return &ViperTestConfig{
		originalValues: make(map[string]any),
	}
}

// Set sets a configuration value and stores the original value
func (v *ViperTestConfig) Set(key string, value any) {
	if viper.IsSet(key) {
		v.originalValues[key] = viper.Get(key)
	}
	viper.Set(key, value)
}

// Reset restores all original values
func (v *ViperTestConfig) Reset() {
	// Reset all keys that were set
	for key := range v.originalValues {
		if originalValue, exists := v.originalValues[key]; exists {
			viper.Set(key, originalValue)
		}
	}
}

// SetupViperForTest configures viper with common test values
func SetupViperForTest(config map[string]any) *ViperTestConfig {
	vc := NewViperTestConfig()
	for key, value := range config {
		vc.Set(key, value)
	}
	return vc
}
