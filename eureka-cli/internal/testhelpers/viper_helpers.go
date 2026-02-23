package testhelpers

import (
	"github.com/spf13/viper"
)

// ViperTestConfig helps manage viper configuration during tests
type ViperTestConfig struct {
	originalValues map[string]any
	keyExisted     map[string]bool
}

// NewViperTestConfig creates a new viper test configuration manager
func NewViperTestConfig() *ViperTestConfig {
	return &ViperTestConfig{
		originalValues: make(map[string]any),
		keyExisted:     make(map[string]bool),
	}
}

// Set sets a configuration value and stores the original value
func (v *ViperTestConfig) Set(key string, value any) {
	if _, tracked := v.keyExisted[key]; !tracked {
		v.keyExisted[key] = viper.IsSet(key)
		if v.keyExisted[key] {
			v.originalValues[key] = viper.Get(key)
		}
	}
	viper.Set(key, value)
}

// Reset restores all original values
func (v *ViperTestConfig) Reset() {
	for key, existed := range v.keyExisted {
		if existed {
			viper.Set(key, v.originalValues[key])
			continue
		}

		viper.Set(key, nil)
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
