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

	viper.Set("existing-key", "original")
	vc := testhelpers.SetupViperForTest(map[string]any{"existing-key": "updated"})

	assert.Equal(t, "updated", viper.GetString("existing-key"))

	vc.Reset()

	assert.Equal(t, "original", viper.GetString("existing-key"))
}

func TestViperTestConfig_Reset_UnsetsNewKey(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	vc := testhelpers.SetupViperForTest(map[string]any{"new-key": "value"})

	assert.True(t, viper.IsSet("new-key"))

	vc.Reset()

	assert.False(t, viper.IsSet("new-key"))
}
