package helpers

import (
	"slices"

	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func HasTenant(tenant string) bool {
	return slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(field.Tenants)), tenant)
}
