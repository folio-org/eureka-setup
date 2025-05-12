package internal

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func GetEnvironmentFromConfig(commandName string, keyType string) []string {
	var environmentVariables []string
	for key, value := range viper.GetStringMapString(keyType) {
		environmentVariables = append(environmentVariables, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return environmentVariables
}

func GetEnvironmentFromMapByKey(requestKey string) string {
	return viper.GetStringMapString(EnvironmentKey)[strings.ToLower(requestKey)]
}

func AppendVaultEnvironment(environment []string, vaultRootToken string) []string {
	extraEnvironment := []string{"SECRET_STORE_TYPE=VAULT",
		fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultRootToken),
		fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", VaultUrl),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendKeycloakEnvironment(environment []string) []string {
	extraEnvironment := []string{fmt.Sprintf("KC_URL=%s", KeycloakUrl),
		fmt.Sprintf("KC_ADMIN_CLIENT_ID=%s", GetEnvironmentFromMapByKey("KC_ADMIN_CLIENT_ID")),
		fmt.Sprintf("KC_SERVICE_CLIENT_ID=%s", GetEnvironmentFromMapByKey("KC_SERVICE_CLIENT_ID")),
		fmt.Sprintf("KC_LOGIN_CLIENT_SUFFIX=%s", GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX")),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendModuleEnvironment(environment []string, extraEnvironmentMap map[string]any) []string {
	for key, value := range extraEnvironmentMap {
		if key == "" {
			continue
		}
		environment = append(environment, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return environment
}

func AppendSidecarEnvironment(environment []string, module *RegistryModule, portServer string, moduleUrl *string, sidecarUrl *string) []string {
	var extraEnvironment []string
	if moduleUrl == nil && sidecarUrl == nil {
		extraEnvironment = []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
			fmt.Sprintf("MODULE_URL=http://%s.eureka:%s", module.Name, portServer),
			fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
			fmt.Sprintf("SIDECAR_URL=http://%s.eureka:%s", module.SidecarName, portServer),
		}
	} else {
		extraEnvironment = []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
			fmt.Sprintf("MODULE_URL=%s", *moduleUrl),
			fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
			fmt.Sprintf("SIDECAR_URL=%s", *sidecarUrl),
		}
	}
	// Change the default port on netty server in Quarkus
	if portServer != DefaultServerPort {
		extraEnvironment = append(extraEnvironment, fmt.Sprintf("QUARKUS_HTTP_PORT=%s", portServer))
	}
	environment = append(environment, extraEnvironment...)

	return environment
}

func AppendDisableSystemUserEnvironment(environment []string, module *RegistryModule) []string {
	extraEnvironment := []string{"FOLIO_SYSTEM_USER_ENABLED=false",
		"SYSTEM_USER_CREATE=false",
		"SYSTEM_USER_ENABLED=false",
		fmt.Sprintf("SYSTEM_USER_NAME=%s", module.Name),
		fmt.Sprintf("SYSTEM_USER_USERNAME=%s", module.Name),
	}
	environment = append(environment, extraEnvironment...)

	return environment
}
