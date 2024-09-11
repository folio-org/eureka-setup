package internal

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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

// TODO Move the hardcoded environment to config
func AppendSidecarEnvironment(environment []string, module *RegistryModule) []string {
	extraEnvironment := []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
		fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
		fmt.Sprintf("MODULE_URL=http://%s.eureka:%s", module.Name, ServerPort),
		fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
		fmt.Sprintf("SIDECAR_URL=http://%s.eureka:%s", module.SidecarName, ServerPort),
		"MOD_USERS_KEYCLOAK_URL=http://mod-users-keycloak.eureka:8081",
		"SIDECAR_FORWARD_UNKNOWN_REQUESTS=true",
		"SIDECAR_FORWARD_UNKNOWN_REQUESTS_DESTINATION=http://api-gateway.eureka:8000",
	}
	environment = append(environment, extraEnvironment...)

	return environment
}
