package internal

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	WorkDir        string = ".eureka"
	ComposeFileDir string = "./misc"

	ContentTypeJson   string = "application/json"
	DockerInternalUrl string = "http://host.docker.internal:%d%s"
	HostIp            string = "0.0.0.0"
	ServerPort        string = "8081"
	DebugPort         string = "5005"

	VaultTokenPattern       string = ".*:"
	ModuleIdPattern         string = "([a-z-_]+)([\\d-_.]+)([a-zA-Z0-9-_.]+)"
	EnvNamePattern          string = "[.-]+"
	ManagementModulePattern string = "mgr-"
)

const (
	ProfileNameKey string = "profile.name"

	ApplicationsKey string = "applications"

	RegistryUrlKey                  string = "registry.registry-url"
	RegistrySidecarImageKey         string = "registry.sidecar-image"
	RegistryFolioInstallJsonUrlKey  string = "registry.folio-install-json-url"
	RegistryEurekaInstallJsonUrlKey string = "registry.eureka-install-json-url"

	EnvironmentKey string = "environment"

	FilesModuleEnvKey         string = "files.module-env"
	FilesModuleDescriptorsKey string = "files.module-descriptors"

	ResourcesVaultKey              string = "resources.vault"
	ResourcesKeycloakKey           string = "resources.keycloak"
	ResourcesMgrTenantsKey         string = "resources.mgr-tenants"
	ResourcesMgrApplicationsKey    string = "resources.mgr-applications"
	ResourcesMgrTenantEntitlements string = "resources.mgr-tenant-entitlements"

	TenantConfigKey string = "tenants"

	BackendModuleKey  string = "backend-modules"
	FrontendModuleKey string = "frontend-modules"

	PortKey      string = "port"
	SidecarKey   string = "sidecar"
	ModuleEnvKey string = "environment"
)

var (
	VaultTokenRegexp = regexp.MustCompile(VaultTokenPattern)
	ModuleIdRegexp   = regexp.MustCompile(ModuleIdPattern)
	EnvNameRegexp    = regexp.MustCompile(EnvNamePattern)

	PortIndex = 30000
)

func GetEnvironmentFromConfig(commandName string, sharedEnvMap map[string]string) []string {
	var sharedEnv []string

	for name, value := range sharedEnvMap {
		env := fmt.Sprintf("%s=%s", strings.ToUpper(name), value)

		slog.Info(commandName, "Found environment", env)

		sharedEnv = append(sharedEnv, env)
	}

	return sharedEnv
}

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any) map[string]BackendModule {
	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		var (
			port        *int
			sidecar     *bool
			environment map[string]interface{}
		)

		if value != nil {
			mapEntry := value.(map[string]interface{})

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
			backendModulesMap[name] = *NewBackendModuleAndSidecar(name, *port, *sidecar, environment)
		} else {
			backendModulesMap[name] = *NewBackendModule(name, *port, environment)
		}

		slog.Info(commandName, "Found backend module", name)
	}

	return backendModulesMap
}

func GetFrontendModulesFromConfig(commandName string, frontendModulesAnyMap map[string]any) map[string]FrontendModule {
	frontendModulesMap := make(map[string]FrontendModule)

	for name := range frontendModulesAnyMap {
		frontendModulesMap[name] = *NewFrontendModule(name)

		slog.Info(commandName, "Found frontend module", name)
	}

	return frontendModulesMap
}

func CreateModuleEnvFile(commandName string, fileModuleEnv string) *os.File {
	err := os.Remove(fileModuleEnv)
	if err != nil {
		slog.Warn(commandName, fmt.Sprintf("os.Remove warn=\"%s\"", err.Error()), "")
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
		slog.Warn(commandName, fmt.Sprintf("os.Remove warn=\"%s\"", err.Error()), "")
	}

	moduleDescriptorsFile, err := os.OpenFile(fileModuleDescriptors, os.O_CREATE, 0644)
	if err != nil {
		slog.Error(commandName, "os.OpenFile error", "")
		panic(err)
	}

	return moduleDescriptorsFile
}

func RunCommand(commandName string, preparedCommand *exec.Cmd, composeFileDir string) {
	preparedCommand.Dir = composeFileDir
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr
	if err := preparedCommand.Run(); err != nil {
		slog.Error(commandName, "systemCmd.Run() error", "")
		panic(err)
	}
}
