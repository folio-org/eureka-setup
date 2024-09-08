package internal

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const (
	DockerComposeWorkDir string = "./misc"
	NetworkName          string = "fpm-net"
	NetworkId            string = "eureka"
	DockerInternalUrl    string = "http://host.docker.internal:%d%s"
	HostIp               string = "0.0.0.0"
	ServerPort           string = "8081"
	DebugPort            string = "5005"

	PlatformCompleteUrl           string = "http://ui.eureka:80"
	PlatformCompleteRepositoryUrl string = "https://github.com/folio-org/platform-complete.git"

	VaultRootTokenPattern               string = ".*:"
	ModuleIdPattern                     string = "([a-z-_]+)([\\d-_.]+)([a-zA-Z0-9-_.]+)"
	EnvNamePattern                      string = "[.-]+"
	ManagementModulePattern             string = "mgr-"
	ModulesPattern                      string = "eureka-"
	ManagementModuleContainerPattern    string = "eureka-mgr-"
	MultipleModulesContainerPattern     string = "eureka-mod-"
	ManagementOrModulesContainerPattern string = "^(eureka-)(mod|mgr)-(.+)"
	SingleModuleContainerPattern        string = "^(eureka-)(%[1]s|%[1]s-sc)$"
	SingleUiContainerPattern            string = "eureka-platform-complete-ui-%s"
)

const (
	ProfileNameKey string = "profile.name"

	ApplicationKey       string = "application"
	ApplicationPortStart string = "application.port-start"

	RegistryUrlKey                  string = "registry.registry-url"
	RegistryFolioInstallJsonUrlKey  string = "registry.folio-install-json-url"
	RegistryEurekaInstallJsonUrlKey string = "registry.eureka-install-json-url"

	EnvironmentKey string = "environment"

	FilesModuleEnvKey         string = "files.module-env"
	FilesModuleDescriptorsKey string = "files.module-descriptors"

	ResourcesKongKey               string = "resources.kong"
	ResourcesVaultKey              string = "resources.vault"
	ResourcesKeycloakKey           string = "resources.keycloak"
	ResourcesMgrTenantsKey         string = "resources.mgr-tenants"
	ResourcesMgrApplicationsKey    string = "resources.mgr-applications"
	ResourcesMgrTenantEntitlements string = "resources.mgr-tenant-entitlements"

	TenantsKey string = "tenants"
	UsersKey   string = "users"
	RolesKey   string = "roles"

	SidecarModule               string = "sidecar-module"
	SidecarModuleEnvironmentKey string = "sidecar-module.environment"
	BackendModuleKey            string = "backend-modules"
	FrontendModuleKey           string = "frontend-modules"

	DeployModuleKey string = "deploy-module"
	VersionKey      string = "version"
	PortKey         string = "port"
	SidecarKey      string = "sidecar"
	ModuleEnvKey    string = "environment"
)

var (
	VaultRootTokenRegexp *regexp.Regexp = regexp.MustCompile(VaultRootTokenPattern)
	ModuleIdRegexp       *regexp.Regexp = regexp.MustCompile(ModuleIdPattern)
	EnvNameRegexp        *regexp.Regexp = regexp.MustCompile(EnvNamePattern)

	PortIndex int = 30000
)

func GetEnvironmentFromConfig(commandName string) []string {
	var environmentVariables []string

	for key, value := range viper.GetStringMapString(EnvironmentKey) {
		environment := fmt.Sprintf("%s=%s", strings.ToUpper(key), value)

		slog.Info(commandName, "Found environment", environment)

		environmentVariables = append(environmentVariables, environment)
	}

	return environmentVariables
}

func GetSidecarEnvironmentFromConfig(commandName string) []string {
	var environmentVariables []string

	for key, value := range viper.GetStringMapString(SidecarModuleEnvironmentKey) {
		environment := fmt.Sprintf("%s=%s", strings.ToUpper(key), value)

		slog.Info(commandName, "Found sidecar environment", environment)

		environmentVariables = append(environmentVariables, environment)
	}

	return environmentVariables
}

func GetEnvironmentFromMapByKey(requestKey string) string {
	for key, value := range viper.GetStringMapString(EnvironmentKey) {
		if strings.ToUpper(key) == requestKey {
			return value
		}
	}

	panic(fmt.Errorf("internal.GetEnvironmentFromMapByKey error - Cannnot find environment value key by %s", requestKey))
}

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any) map[string]BackendModule {
	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		var (
			deployModule bool = true
			version      *string
			port         *int
			sidecar      *bool
			environment  map[string]interface{}
		)

		if value != nil {
			mapEntry := value.(map[string]interface{})

			if mapEntry[DeployModuleKey] != nil {
				deployModule = mapEntry[DeployModuleKey].(bool)
			}

			if mapEntry[VersionKey] != nil {
				var versionValue string
				_, ok := mapEntry[VersionKey].(float64)
				if ok {
					versionValue = strconv.FormatFloat(mapEntry[VersionKey].(float64), 'f', -1, 64)
				} else {
					versionValue = mapEntry[VersionKey].(string)
				}
				version = &versionValue
			}

			if mapEntry[PortKey] != nil {
				portValue := mapEntry[PortKey].(int)
				port = &portValue
			}

			if mapEntry[SidecarKey] != nil {
				sidecarValue := mapEntry[SidecarKey].(bool)
				sidecar = &sidecarValue
			}

			if mapEntry[ModuleEnvKey] != nil {
				environment = mapEntry[ModuleEnvKey].(map[string]interface{})
			} else {
				environment = make(map[string]interface{})
			}
		} else {
			PortIndex++
			port = &PortIndex

			sidecarDefaultValue := true
			sidecar = &sidecarDefaultValue

			environment = make(map[string]interface{})
		}

		if sidecar != nil && *sidecar {
			backendModulesMap[name] = *NewBackendModuleAndSidecar(deployModule, name, version, *port, *sidecar, environment)
		} else {
			backendModulesMap[name] = *NewBackendModule(name, *port, environment)
		}

		moduleInfo := name
		if version != nil {
			moduleInfo = fmt.Sprintf("%s with fixed version %s", name, *version)
		}

		slog.Info(commandName, "Found backend module in config", moduleInfo)
	}

	return backendModulesMap
}

func GetFrontendModulesFromConfig(commandName string, frontendModulesAnyMap map[string]any) map[string]FrontendModule {
	frontendModulesMap := make(map[string]FrontendModule)

	for name, value := range frontendModulesAnyMap {
		var (
			deployModule bool = true
			version      *string
		)

		if value != nil {
			mapEntry := value.(map[string]interface{})

			if mapEntry[DeployModuleKey] != nil {
				deployModule = mapEntry[DeployModuleKey].(bool)
			}

			if mapEntry[VersionKey] != nil {
				versionValue := mapEntry[VersionKey].(string)
				version = &versionValue
			}
		}

		frontendModulesMap[name] = *NewFrontendModule(deployModule, name, version)

		moduleInfo := name
		if version != nil {
			moduleInfo = fmt.Sprintf("name %s with version %s", name, *version)
		}

		slog.Info(commandName, "Found frontend module in config", moduleInfo)
	}

	return frontendModulesMap
}

func CreateModuleEnvFile(commandName string, fileModuleEnv string) *os.File {
	err := os.Remove(fileModuleEnv)
	if err != nil {
		slog.Warn(commandName, fmt.Sprintf("os.Remove warning, '%s'", err.Error()), "")
	}

	fileModuleEnvPointer, err := os.OpenFile(fileModuleEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error(commandName, "os.OpenFile error", "")
		panic(err)
	}

	return fileModuleEnvPointer
}

func CreateModuleDescriptorsFile(commandName string, fileModuleDescriptors string) *os.File {
	err := os.Remove(fileModuleDescriptors)
	if err != nil {
		slog.Warn(commandName, fmt.Sprintf("os.Remove warning, '%s'", err.Error()), "")
	}

	moduleDescriptorsFile, err := os.OpenFile(fileModuleDescriptors, os.O_CREATE, 0644)
	if err != nil {
		slog.Error(commandName, "os.OpenFile error", "")
		panic(err)
	}

	return moduleDescriptorsFile
}

func PrepareStripesConfigJson(commandName string, configPath string, tenant string) {
	stripesConfigJsFilePath := fmt.Sprintf("%s/stripes.config.js", configPath)
	readFileBytes, err := os.ReadFile(stripesConfigJsFilePath)
	if err != nil {
		slog.Error(commandName, "os.ReadFile error", "")
		panic(err)
	}

	replaceMap := map[string]string{"${kongUrl}": viper.GetString(ResourcesKongKey),
		"${tenantUrl}":      PlatformCompleteUrl,
		"${keycloakUrl}":    viper.GetString(ResourcesKeycloakKey),
		"${hasAllPerms}":    `true`,
		"${isSingleTenant}": `true`,
		"${tenantOptions}":  fmt.Sprintf(`{%[1]s: {name: "%[1]s", clientId: "%[1]s%s"}}`, tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX")),
	}

	var newReadFileStr string = string(readFileBytes)
	for key, value := range replaceMap {
		newReadFileStr = strings.Replace(newReadFileStr, key, value, -1)
	}

	err = os.WriteFile(stripesConfigJsFilePath, []byte(newReadFileStr), 0)
	if err != nil {
		slog.Error(commandName, "os.WriteFile error", "")
		panic(err)
	}
}
