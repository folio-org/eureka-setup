package testhelpers_test

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestViperTestConfig_Reset_RestoresExistingValue(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Arrange
	viper.Set("existing-key", "original")
	vc := testhelpers.SetupViperForTest(map[string]any{"existing-key": "updated"})
	assert.Equal(t, "updated", viper.GetString("existing-key"))

	// Act
	vc.Reset()

	// Assert
	assert.Equal(t, "original", viper.GetString("existing-key"))
}

func TestViperTestConfig_Reset_UnsetsNewKey(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Arrange
	vc := testhelpers.SetupViperForTest(map[string]any{"new-key": "value"})
	assert.True(t, viper.IsSet("new-key"))

	// Act
	vc.Reset()

	// Assert
	assert.False(t, viper.IsSet("new-key"))
}
