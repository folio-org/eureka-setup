package moduleenv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/models"
)

// ModuleEnvProcessor defines the interface for building module environment variable configurations
type ModuleEnvProcessor interface {
	VaultEnv(envVars []string, vaultRootToken string) []string
	OkapiEnv(envVars []string, sidecarName string, portServer int) []string
	DisabledSystemUserEnv(envVars []string, moduleName string) []string
	KeycloakEnv(envVars []string) []string
	ModuleEnv(envVars []string, extraEnvVars map[string]any) []string
	SidecarEnv(envVars []string, module *models.RegistryModule, portServer int, moduleURL, sidecarURL *string) []string
}

// ModuleEnv provides functionality for constructing environment variables for modules
type ModuleEnv struct {
	Action *action.Action
}

// New creates a new ModuleEnv instance
func New(action *action.Action) *ModuleEnv {
	return &ModuleEnv{Action: action}
}

func (mv *ModuleEnv) VaultEnv(envVars []string, vaultRootToken string) []string {
	extraEnvVars := []string{"SECRET_STORE_TYPE=VAULT",
		fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultRootToken),
		fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", constant.VaultHTTP),
	}
	envVars = append(envVars, extraEnvVars...)

	return envVars
}

func (mv *ModuleEnv) OkapiEnv(envVars []string, sidecarName string, portServer int) []string {
	extraEnvVars := []string{fmt.Sprintf("OKAPI_HOST=%s.eureka", sidecarName),
		fmt.Sprintf("OKAPI_PORT=%d", portServer),
		fmt.Sprintf("OKAPI_SERVICE_HOST=%s.eureka", sidecarName),
		fmt.Sprintf("OKAPI_SERVICE_URL=http://%s.eureka:%d", sidecarName, portServer),
		fmt.Sprintf("OKAPI_URL=http://%s.eureka:%d", sidecarName, portServer),
	}
	envVars = append(envVars, extraEnvVars...)

	return envVars
}

func (mv *ModuleEnv) DisabledSystemUserEnv(envVars []string, moduleName string) []string {
	extraEnvVars := []string{"FOLIO_SYSTEM_USER_ENABLED=false",
		"SYSTEM_USER_CREATE=false",
		"SYSTEM_USER_ENABLED=false",
		fmt.Sprintf("SYSTEM_USER_NAME=%s", moduleName),
		fmt.Sprintf("SYSTEM_USER_USERNAME=%s", moduleName),
	}
	envVars = append(envVars, extraEnvVars...)

	return envVars
}

func (mv *ModuleEnv) KeycloakEnv(envVars []string) []string {
	extraEnvVars := []string{fmt.Sprintf("KC_URL=%s", constant.KeycloakHTTP),
		fmt.Sprintf("KC_ADMIN_CLIENT_ID=%s", action.GetConfigEnv("KC_ADMIN_CLIENT_ID", mv.Action.ConfigGlobalEnv)),
		fmt.Sprintf("KC_SERVICE_CLIENT_ID=%s", action.GetConfigEnv("KC_SERVICE_CLIENT_ID", mv.Action.ConfigGlobalEnv)),
		fmt.Sprintf("KC_LOGIN_CLIENT_SUFFIX=%s", action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", mv.Action.ConfigGlobalEnv)),
	}
	envVars = append(envVars, extraEnvVars...)

	return envVars
}

func (mv *ModuleEnv) ModuleEnv(envVars []string, extraEnvVars map[string]any) []string {
	for key, value := range extraEnvVars {
		if key == "" {
			continue
		}
		envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return envVars
}

func (mv *ModuleEnv) SidecarEnv(envVars []string, module *models.RegistryModule, portServer int, moduleURL, sidecarURL *string) []string {
	var extraEnvVars []string
	if moduleURL == nil && sidecarURL == nil {
		extraEnvVars = []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
			fmt.Sprintf("MODULE_URL=http://%s.eureka:%d", module.Name, portServer),
			fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
			fmt.Sprintf("SIDECAR_URL=http://%s.eureka:%d", module.SidecarName, portServer),
		}
	} else {
		extraEnvVars = []string{fmt.Sprintf("MODULE_NAME=%s", module.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Version),
			fmt.Sprintf("MODULE_URL=%s", *moduleURL),
			fmt.Sprintf("SIDECAR_NAME=%s", module.SidecarName),
			fmt.Sprintf("SIDECAR_URL=%s", *sidecarURL),
		}
	}
	// Change the default port on Quarkus netty server
	if strconv.Itoa(portServer) != constant.ServerPort {
		extraEnvVars = append(extraEnvVars, fmt.Sprintf("QUARKUS_HTTP_PORT=%d", portServer))
	}
	envVars = append(envVars, extraEnvVars...)

	return envVars
}
