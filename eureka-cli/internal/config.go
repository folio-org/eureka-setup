package internal

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/spf13/viper"
)

const (
	ConfigDir      string = ".eureka"
	ConfigCombined string = "config.combined"
	ConfigType     string = "yaml"

	FolioRegistry  string = "folio"
	EurekaRegistry string = "eureka"

	DockerComposeWorkDir   string = "./misc"
	DefaultNetworkId       string = "eureka"
	DefaultNetworkAlias    string = "eureka-net"
	DefaultDockerHostname  string = "host.docker.internal"
	DefaultDockerGatewayIp string = "172.17.0.1"

	DefaultHostIp     string = "0.0.0.0"
	DefaultServerPort string = "8081"
	DefaultDebugPort  string = "5005"

	FolioKeycloakRepositoryUrl    string = "https://github.com/folio-org/folio-keycloak"
	FolioKongRepositoryUrl        string = "https://github.com/folio-org/folio-kong"
	PlatformCompleteRepositoryUrl string = "https://github.com/folio-org/platform-complete.git"
)

const (
	ManagementModulePattern string = "mgr-"

	SingleModuleContainerPattern    string = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
	MultipleModulesContainerPattern string = "eureka-%s-*-"
)

var (
	PortStartIndex int = 30000
	PortEndIndex   int = 32000
)

func GetGatewayUrlTemplate(commandName string) string {
	schemaAndUrl := GetGatewaySchemaAndUrl(commandName)
	if schemaAndUrl == "" {
		LogErrorPanic(commandName, fmt.Sprintf("internal.GetGatewayUrlTemplate error - cannot construct geteway url template for %s platform", runtime.GOOS))
		return ""
	}

	return schemaAndUrl + ":%d%s"
}

func GetGatewaySchemaAndUrl(commandName string) string {
	var schemaAndUrl string
	if viper.IsSet(ApplicationGatewayHostnameKey) {
		schemaAndUrl = viper.GetString(ApplicationGatewayHostnameKey)
	} else if HostnameExists(commandName, DefaultDockerHostname) {
		schemaAndUrl = fmt.Sprintf("http://%s", DefaultDockerHostname)
	} else if runtime.GOOS == "linux" {
		schemaAndUrl = fmt.Sprintf("http://%s", DefaultDockerGatewayIp)
	}

	return schemaAndUrl
}

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any, managementOnly bool, printOutput bool) map[string]BackendModule {
	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		var dto BackendModuleDto
		if value == nil {
			dto = createDefaultBackendDto(commandName, name)
		} else {
			dto = createConfigurableBackendDto(commandName, value, name)
		}

		if dto.deploySidecar != nil && *dto.deploySidecar {
			backendModulesMap[name] = *NewBackendModuleWithSidecar(dto)
		} else {
			backendModulesMap[name] = *NewBackendModule(dto)
		}

		moduleInfo := name
		if dto.version != nil {
			moduleInfo = fmt.Sprintf("%s with fixed version %s", name, *dto.version)
		}

		if printOutput {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found backend module in config: %s", moduleInfo))
		}
	}

	return backendModulesMap
}

func createDefaultBackendDto(commandName string, name string) (dto BackendModuleDto) {
	dto.deployModule = true
	if !strings.HasPrefix(name, ManagementModulePattern) {
		dto.deploySidecar = getDefaultDeploySidecar()
	}
	dto.useVault = false
	dto.disableSystemUser = false
	dto.version = nil

	if PortStartIndex+1 >= PortEndIndex {
		LogErrorPanic(commandName, "internal.createDefaultBackendDto error - incremented PortStartIndex is exceeding PortEndIndex limit")
		return
	}

	PortStartIndex++
	dto.port = &PortStartIndex
	dto.portServer = getDefaultPortServer()
	dto.environment = make(map[string]any)
	dto.resources = make(map[string]any)
	dto.volumes = []string{}

	return dto
}

func createConfigurableBackendDto(commandName string, value any, name string) (dto BackendModuleDto) {
	mapEntry := value.(map[string]any)

	dto.deployModule = getDeployModule(mapEntry)
	dto.deploySidecar = getDeploySidecar(mapEntry, name)
	dto.useVault = getUseVault(mapEntry)
	dto.disableSystemUser = getDisableSystemUser(mapEntry)
	dto.version = getVersion(mapEntry)
	dto.port = getPort(mapEntry)
	dto.portServer = getPortServer(mapEntry)
	dto.environment = createEnvironment(mapEntry)
	dto.resources = createResources(mapEntry)
	dto.volumes = createVolumes(commandName, mapEntry)

	return dto
}

func getDeployModule(mapEntry map[string]any) bool {
	if mapEntry[ModuleDeployModuleEntryKey] == nil {
		return true
	}

	return mapEntry[ModuleDeployModuleEntryKey].(bool)
}

func getDeploySidecar(mapEntry map[string]any, name string) *bool {
	var sidecar *bool
	if mapEntry[ModuleDeploySidecarEntryKey] != nil {
		sidecarValue := mapEntry[ModuleDeploySidecarEntryKey].(bool)
		sidecar = &sidecarValue
	} else if !strings.HasPrefix(name, ManagementModulePattern) {
		sidecar = getDefaultDeploySidecar()
	}

	return sidecar
}

func getDefaultDeploySidecar() *bool {
	deploySidecarDefaultValue := true
	return &deploySidecarDefaultValue
}

func getUseVault(mapEntry map[string]any) bool {
	if mapEntry[ModuleUseVaultEntryKey] == nil {
		return false
	}

	return mapEntry[ModuleUseVaultEntryKey].(bool)
}

func getDisableSystemUser(mapEntry map[string]any) bool {
	if mapEntry[ModuleDisableSystemUserEntryKey] == nil {
		return false
	}

	return mapEntry[ModuleDisableSystemUserEntryKey].(bool)
}

func getVersion(mapEntry map[string]any) *string {
	var version *string
	if mapEntry[ModuleVersionEntryKey] != nil {
		var versionValue string
		_, ok := mapEntry[ModuleVersionEntryKey].(float64)
		if ok {
			versionValue = strconv.FormatFloat(mapEntry[ModuleVersionEntryKey].(float64), 'f', -1, 64)
		} else {
			versionValue = mapEntry[ModuleVersionEntryKey].(string)
		}
		version = &versionValue
	}

	return version
}

func getPort(mapEntry map[string]any) *int {
	var port *int
	if mapEntry[ModulePortEntryKey] != nil {
		portValue := mapEntry[ModulePortEntryKey].(int)
		port = &portValue
	} else {
		PortStartIndex++
		port = &PortStartIndex
	}

	return port
}

func getPortServer(mapEntry map[string]any) *int {
	var portServer *int
	if mapEntry[ModulePortServerEntryKey] != nil {
		portServerValue := mapEntry[ModulePortServerEntryKey].(int)
		portServer = &portServerValue
	} else {
		portServer = getDefaultPortServer()
	}

	return portServer
}

func getDefaultPortServer() *int {
	defaultServerPort, _ := strconv.Atoi(DefaultServerPort)
	portServerValue := defaultServerPort
	return &portServerValue
}

func createEnvironment(mapEntry map[string]any) map[string]any {
	if mapEntry[ModuleEnvironmentEntryKey] == nil {
		return make(map[string]any)
	}

	return mapEntry[ModuleEnvironmentEntryKey].(map[string]any)
}

func createResources(mapEntry map[string]any) map[string]any {
	if mapEntry[ModuleResourceEntryKey] == nil {
		return make(map[string]any)
	}

	return mapEntry[ModuleResourceEntryKey].(map[string]any)
}

func createVolumes(commandName string, mapEntry map[string]any) []string {
	if mapEntry[ModuleVolumesEntryKey] == nil {
		return []string{}
	}

	var volumes []string
	for _, value := range mapEntry[ModuleVolumesEntryKey].([]any) {
		var volume string = value.(string)
		if runtime.GOOS == "windows" && strings.Contains(volume, "$CWD") {
			cwd := GetCurrentWorkDirPath(commandName)
			volume = strings.ReplaceAll(volume, "$CWD", cwd)
		}

		if _, err := os.Stat(volume); os.IsNotExist(err) {
			if err != nil {
				slog.Error(commandName, GetFuncName(), "os.IsNotExist error")
				panic(err)
			}
		}

		volumes = append(volumes, volume)
	}

	return volumes
}

func GetFrontendModulesFromConfig(commandName string, frontendModulesAnyMaps ...map[string]any) map[string]FrontendModule {
	frontendModulesMap := make(map[string]FrontendModule)

	for _, frontendModulesAnyMap := range frontendModulesAnyMaps {
		for name, value := range frontendModulesAnyMap {
			var (
				deployModule bool = true
				version      *string
			)

			if value != nil {
				mapEntry := value.(map[string]any)

				if mapEntry[ModuleDeployModuleEntryKey] != nil {
					deployModule = mapEntry[ModuleDeployModuleEntryKey].(bool)
				}

				if mapEntry[ModuleVersionEntryKey] != nil {
					versionValue := mapEntry[ModuleVersionEntryKey].(string)
					version = &versionValue
				}
			}

			frontendModulesMap[name] = *NewFrontendModule(deployModule, name, version)

			moduleInfo := name
			if version != nil {
				moduleInfo = fmt.Sprintf("name %s with version %s", name, *version)
			}

			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found frontend module in config: %s", moduleInfo))
		}
	}

	return frontendModulesMap
}

func PrepareStripesConfigJs(commandName string, configPath string, tenant string, kongUrl string, keycloakUrl string, platformCompleteUrl string, enableEcsRequests bool) {
	stripesConfigJsFilePath := fmt.Sprintf("%s/stripes.config.js", configPath)
	readFileBytes, err := os.ReadFile(stripesConfigJsFilePath)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.ReadFile error")
		panic(err)
	}

	replaceMap := map[string]string{"${kongUrl}": kongUrl,
		"${tenantUrl}":         platformCompleteUrl,
		"${keycloakUrl}":       keycloakUrl,
		"${hasAllPerms}":       `false`,
		"${isSingleTenant}":    `true`,
		"${tenantOptions}":     fmt.Sprintf(`{%[1]s: {name: "%[1]s", clientId: "%[1]s%s"}}`, tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX")),
		"${enableEcsRequests}": strconv.FormatBool(enableEcsRequests),
	}

	var newReadFileStr string = string(readFileBytes)
	for key, value := range replaceMap {
		if !strings.Contains(newReadFileStr, key) {
			slog.Info(commandName, GetFuncName(), fmt.Sprintf("Key not found in stripes.config.js: %s", key))
			continue
		}

		newReadFileStr = strings.Replace(newReadFileStr, key, value, -1)
	}

	fmt.Println()
	fmt.Println("###### Dumping stripes.config.js ######")
	fmt.Println(newReadFileStr)
	fmt.Println()

	err = os.WriteFile(stripesConfigJsFilePath, []byte(newReadFileStr), 0)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.WriteFile error")
		panic(err)
	}
}

func PreparePackageJson(commandName string, configPath string, tenant string) {
	packageJsonPath := fmt.Sprintf("%s/package.json", configPath)

	var packageJson struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		License         string            `json:"license"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Resolutions     map[string]string `json:"resolutions"`
	}

	ReadJsonFromFile(commandName, packageJsonPath, &packageJson)

	packageJson.Scripts["build"] = "export DEBUG=stripes*; export NODE_OPTIONS=\"--max-old-space-size=8000 $NODE_OPTIONS\"; stripes build stripes.config.js --languages en --sourcemap=false --no-minify"

	updates := 0
	modules := []string{
		"@folio/consortia-settings",
		"@folio/authorization-policies",
		"@folio/authorization-roles",
		"@folio/plugin-select-application",
	}
	for _, module := range modules {
		if packageJson.Dependencies[module] == "" {
			packageJson.Dependencies[module] = ">=1.0.0"
			updates++
		}
	}

	if updates > 0 {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Added %d extra modules to package.json", len(modules)))
		WriteJsonToFile(commandName, packageJsonPath, packageJson)
	}
}

func GetStripesBranch(commandName string, defaultBranch plumbing.ReferenceName) plumbing.ReferenceName {
	if viper.IsSet(ApplicationStripesBranchKey) {
		branchStr := viper.GetString(ApplicationStripesBranchKey)
		stripesBranch := plumbing.ReferenceName(branchStr)
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Got stripes branch from config: %s", stripesBranch))

		return stripesBranch
	}

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("No stripes branch in config. Using default branch: %s", defaultBranch))

	return defaultBranch
}

func HasTenant(tenant string) bool {
	return slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant)
}

func CanDeployUi(tenant string) bool {
	mapEntry, ok := viper.GetStringMap(TenantsKey)[tenant].(map[string]any)
	if !ok {
		return false
	}

	deployUi, ok := mapEntry[TenantsDeployUiEntryKey]
	if !ok || deployUi == nil || !deployUi.(bool) {
		return false
	}

	return true
}

func CanDeployModule(module string) bool {
	for name, value := range viper.GetStringMap(BackendModulesKey) {
		if name == module {
			if value == nil {
				return true
			}

			deployModule, ok := value.(map[string]any)[ModuleDeployModuleEntryKey].(bool)
			if !ok {
				return false
			}

			return deployModule
		}
	}

	return false
}
