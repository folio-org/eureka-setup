package helpers

import (
	"fmt"
	"strings"

	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func GetConfigEnvVars(key string) []string {
	var envVars []string
	for key, value := range viper.GetStringMapString(key) {
		envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}
	return envVars
}

func GetConfigEnv(key string) string {
	return viper.GetStringMapString(field.Env)[strings.ToLower(key)]
}
