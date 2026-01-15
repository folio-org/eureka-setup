package moduleenv

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// ModuleEnvProcessor defines the interface for building module environment variable configurations
type ModuleEnvProcessor interface {
	VaultEnv(env []string, vaultRootToken string) []string
	OkapiEnv(env []string, sidecarName string, privatePort int) []string
	DisabledSystemUserEnv(env []string, moduleName string) []string
	KeycloakEnv(env []string) []string
	ModuleEnv(env []string, newEnv map[string]any) []string
	SidecarEnv(env []string, module *models.ProxyModule, privatePort int, moduleURL, sidecarURL string) []string
}

// ModuleEnv provides functionality for constructing environment variables for modules
type ModuleEnv struct {
	Action *action.Action
}

// New creates a new ModuleEnv instance
func New(action *action.Action) *ModuleEnv {
	return &ModuleEnv{Action: action}
}

func (mv *ModuleEnv) VaultEnv(env []string, vaultRootToken string) []string {
	newEnv := []string{
		"SECRET_STORE_TYPE=VAULT",
		fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultRootToken),
		fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", constant.VaultHTTP),
	}
	env = append(env, newEnv...)

	return env
}

func (mv *ModuleEnv) OkapiEnv(env []string, sidecarName string, privatePort int) []string {
	newEnv := []string{fmt.Sprintf(
		"OKAPI_HOST=%s.eureka", sidecarName),
		fmt.Sprintf("OKAPI_PORT=%d", privatePort),
		fmt.Sprintf("OKAPI_SERVICE_HOST=%s.eureka", sidecarName),
		fmt.Sprintf("OKAPI_SERVICE_URL=http://%s.eureka:%d", sidecarName, privatePort),
		fmt.Sprintf("OKAPI_URL=http://%s.eureka:%d", sidecarName, privatePort),
	}
	env = append(env, newEnv...)

	return env
}

func (mv *ModuleEnv) DisabledSystemUserEnv(env []string, moduleName string) []string {
	newEnv := []string{
		"FOLIO_SYSTEM_USER_ENABLED=false",
		"SYSTEM_USER_CREATE=false",
		"SYSTEM_USER_ENABLED=false",
		fmt.Sprintf("SYSTEM_USER_NAME=%s", moduleName),
		fmt.Sprintf("SYSTEM_USER_USERNAME=%s", moduleName),
	}
	env = append(env, newEnv...)

	return env
}

func (mv *ModuleEnv) KeycloakEnv(env []string) []string {
	newEnv := []string{
		fmt.Sprintf("KC_URL=%s", constant.KeycloakHTTP),
		fmt.Sprintf("KC_ADMIN_CLIENT_ID=%s", action.GetConfigEnv("KC_ADMIN_CLIENT_ID", mv.Action.ConfigGlobalEnv)),
		fmt.Sprintf("KC_SERVICE_CLIENT_ID=%s", action.GetConfigEnv("KC_SERVICE_CLIENT_ID", mv.Action.ConfigGlobalEnv)),
		fmt.Sprintf("KC_LOGIN_CLIENT_SUFFIX=%s", action.GetConfigEnv("KC_LOGIN_CLIENT_SUFFIX", mv.Action.ConfigGlobalEnv)),
	}
	env = append(env, newEnv...)

	return env
}

func (mv *ModuleEnv) ModuleEnv(env []string, newEnv map[string]any) []string {
	for key, value := range newEnv {
		if key == "" {
			continue
		}
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return env
}

func (mv *ModuleEnv) SidecarEnv(env []string, module *models.ProxyModule, privatePort int, moduleURL, sidecarURL string) []string {
	var newEnv []string
	if moduleURL == "" && sidecarURL == "" {
		newEnv = []string{
			fmt.Sprintf("MODULE_NAME=%s", module.Metadata.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Metadata.Version),
			fmt.Sprintf("MODULE_URL=http://%s.eureka:%d", module.Metadata.Name, privatePort),
			fmt.Sprintf("SIDECAR_NAME=%s", module.Metadata.SidecarName),
			fmt.Sprintf("SIDECAR_URL=http://%s.eureka:%d", module.Metadata.SidecarName, privatePort),
		}
	} else {
		newEnv = []string{
			fmt.Sprintf("MODULE_NAME=%s", module.Metadata.Name),
			fmt.Sprintf("MODULE_VERSION=%s", *module.Metadata.Version),
			fmt.Sprintf("MODULE_URL=%s", moduleURL),
			fmt.Sprintf("SIDECAR_NAME=%s", module.Metadata.SidecarName),
			fmt.Sprintf("SIDECAR_URL=%s", sidecarURL),
		}
	}
	// Change the default port on Quarkus netty server
	if strconv.Itoa(privatePort) != constant.PrivateServerPort {
		newEnv = append(newEnv, fmt.Sprintf("QUARKUS_HTTP_PORT=%d", privatePort))
	}
	env = append(env, newEnv...)

	return env
}
