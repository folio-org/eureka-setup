package internal

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

func GetEnvironmentFromConfig(commandName string, keyType string) []string {
	var environmentVariables []string
	for key, value := range viper.GetStringMapString(keyType) {
		environment := fmt.Sprintf("%s=%s", strings.ToUpper(key), value)
		slog.Info(commandName, "Found environment", environment)
		environmentVariables = append(environmentVariables, environment)
	}

	return environmentVariables
}

func GetEnvironmentFromMapByKey(requestKey string) string {
	return viper.GetStringMapString(EnvironmentKey)[strings.ToLower(requestKey)]
}

func AppendVaultEnvironment(environment []string, vaultRootToken string, vaultUrl string) []string {
	extraEnvironment := []string{"SECRET_STORE_TYPE=VAULT",
		fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultRootToken),
		fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", vaultUrl),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendKeycloakEnvironment(commandName string, environment []string) []string {
	extraEnvironment := []string{fmt.Sprintf("KC_URL=%s", viper.GetString(ResourcesKeycloakKey)),
		fmt.Sprintf("KC_ADMIN_CLIENT_ID=%s", GetEnvironmentFromMapByKey("KC_ADMIN_CLIENT_ID")),
		fmt.Sprintf("KC_SERVICE_CLIENT_ID=%s", GetEnvironmentFromMapByKey("KC_SERVICE_CLIENT_ID")),
		fmt.Sprintf("KC_LOGIN_CLIENT_SUFFIX=%s", GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX")),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendManagementEnvironment(environment []string) []string {
	extraEnvironment := []string{fmt.Sprintf("TM_CLIENT_URL=%s", viper.GetString(ResourcesMgrTenantsKey)),
		fmt.Sprintf("AM_CLIENT_URL=%s", viper.GetString(ResourcesMgrApplicationsKey)),
		fmt.Sprintf("TE_CLIENT_URL=%s", viper.GetString(ResourcesMgrTenantEntitlements)),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendModuleEnvironment(extraEnvironmentMap map[string]interface{}, environment []string) []string {
	for key, value := range extraEnvironmentMap {
		if key == "" {
			continue
		}

		environment = append(environment, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return environment
}

func AppendSidecarEnvironment(environment []string, module *RegistryModule, portServer string) []string {
	extraEnvironment := []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
		fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
		fmt.Sprintf("MODULE_URL=http://%s.eureka:%s", module.Name, portServer),
		fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
		fmt.Sprintf("SIDECAR_URL=http://%s.eureka:%s", module.SidecarName, portServer),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}
