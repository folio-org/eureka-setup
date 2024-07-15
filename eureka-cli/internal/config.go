package internal

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

const (
	WorkDir           string = ".eureka"
	ApplicationJson   string = "application/json"
	DockerInternalUrl string = "http://host.docker.internal:%d%s"
	HostIp            string = "0.0.0.0"
	ServerPort        string = "8081"
	DebugPort         string = "5005"

	VaultTokenPattern string = `.*:`
	ModuleIdPattern   string = `([a-z-_]+)([\d-_.]+)([a-zA-Z0-9-_.]+)`
	EnvNamePattern    string = `[.-]+`

	PrimaryMessageKey   string = "Primary Message"
	SecondaryMessageKey string = "Secondary Message"
)

const (
	FolioRegistryHostnameKey string = "registries.folio.registry-hostname"
	FolioRegistryUrlKey      string = "registries.folio.registry-url"
	FolioInstallJsonUrlKey   string = "registries.folio.install-json-url"

	EurekaRegistryHostnameKey string = "registries.eureka.registry-hostname"
	EurekaRegistryUrlKey      string = "registries.eureka.registry-url"
	EurekaInstallJsonUrlKey   string = "registries.eureka.install-json-url"
	EurekaSidecarImageKey     string = "registries.eureka.sidecar-image"
	EurekaUsernameKey         string = "registries.eureka.username"
	EurekaPasswordKey         string = "registries.eureka.password"

	SharedEnvKey string = "shared-env"

	CacheFileModuleEnvKey         string = "cache-files.module-env"
	CacheFileModuleDescriptorsKey string = "cache-files.module-descriptors"

	ResourceUrlVaultKey              string = "resource-urls.vault"
	ResourceUrlKeycloakKey           string = "resource-urls.keycloak"
	ResourceUrlMgrTenantsKey         string = "resource-urls.mgr-tenants"
	ResourceUrlMgrApplicationsKey    string = "resource-urls.mgr-applications"
	ResourceUrlMgrTenantEntitlements string = "resource-urls.mgr-tenant-entitlements"

	TenantConfigKey string = "tenant-config"

	BackendModuleKey string = "backend-modules"

	PortKey          string = "port"
	DeployModuleKey  string = "deploy-module"
	DeploySidecarKey string = "deploy-sidecar"
	ModuleEnvKey     string = "module-env"
)

var (
	VaultTokenRegexp = regexp.MustCompile(VaultTokenPattern)
	ModuleIdRegexp   = regexp.MustCompile(ModuleIdPattern)
	EnvNameRegexp    = regexp.MustCompile(EnvNamePattern)
)

func GetSharedEnvFromConfig(commandName string, sharedEnvMap map[string]string) []string {
	var sharedEnv []string

	for name, value := range sharedEnvMap {
		env := fmt.Sprintf("%s=%s", strings.ToUpper(name), value)

		slog.Info(commandName, "Found shared ENV", env)

		sharedEnv = append(sharedEnv, env)
	}

	return sharedEnv
}

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any) map[string]BackendModule {
	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		mapEntry := value.((map[string]interface{}))

		if !mapEntry[DeployModuleKey].(bool) {
			continue
		}

		port := mapEntry[PortKey].(int)
		var moduleEnv map[string]interface{}

		if mapEntry[ModuleEnvKey] != nil {
			moduleEnv = mapEntry[ModuleEnvKey].(map[string]interface{})
		} else {
			moduleEnv = make(map[string]interface{})
		}

		if mapEntry[DeployModuleKey].(bool) {
			backendModulesMap[name] = *NewBackendModuleWithSidecar(name, port, moduleEnv)
		} else {
			backendModulesMap[name] = *NewBackendModule(name, port, moduleEnv)
		}

		slog.Info(commandName, "Found backend module", name)
	}

	return backendModulesMap
}

func CreateModuleEnvCacheFile(commandName string, cacheFileModuleEnv string) *os.File {
	err := os.Remove(cacheFileModuleEnv)
	if err != nil {
		slog.Warn(commandName, SecondaryMessageKey, fmt.Sprintf("os.Remove warn=\"%s\"", err.Error()))
	}

	fileModuleEnvPointer, err := os.OpenFile(cacheFileModuleEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "os.OpenFile error")
		panic(err)
	}

	return fileModuleEnvPointer
}

func CreateModuleDescriptorsCacheFile(commandName string, cacheFileModuleDescriptors string) *os.File {
	err := os.Remove(cacheFileModuleDescriptors)
	if err != nil {
		slog.Warn(commandName, SecondaryMessageKey, fmt.Sprintf("os.Remove warn=\"%s\"", err.Error()))
	}

	moduleDescriptorsFile, err := os.OpenFile(cacheFileModuleDescriptors, os.O_CREATE, 0644)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "os.OpenFile error")
		panic(err)
	}

	return moduleDescriptorsFile
}
