package internal

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/spf13/viper"
)

const (
	ConfigDir     string = ".eureka"
	ConfigMinimal string = "config.minimal"
	ConfigType    string = "yaml"

	FolioRegistry  string = "folio"
	EurekaRegistry string = "eureka"

	DockerComposeWorkDir string = "./misc"
	DefaultNetworkName   string = "eureka-net"
	DefaultNetworkId     string = "eureka"
	DefaultHostname      string = "http://host.docker.internal:%d%s"
	DefaultHostIp        string = "0.0.0.0"
	DefaultServerPort    string = "8081"
	DefaultDebugPort     string = "5005"

	FolioKeycloakRepositoryUrl    string = "https://github.com/folio-org/folio-keycloak"
	FolioKongRepositoryUrl        string = "https://github.com/folio-org/folio-kong"
	PlatformCompleteRepositoryUrl string = "https://github.com/folio-org/platform-complete.git"
)

const (
	ManagementModulePattern string = "mgr-"

	SingleModuleContainerPattern    string = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
	MultipleModulesContainerPattern string = "eureka-%s-mod-"
)

var PortStartIndex int = 30000

func GetGatewayHostname() string {
	if viper.IsSet(ApplicationGatewayHostnameKey) {
		return viper.GetString(ApplicationGatewayHostnameKey) + ":%d%s"
	}

	return DefaultHostname
}

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any, managementOnly bool) map[string]BackendModule {
	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		var dto BackendModuleDto
		if value == nil {
			dto = createDefaultBackendDto(name)
		} else {
			dto = createConfigurableBackendDto(value, name)
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

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found backend module in config: %s", moduleInfo))
	}

	return backendModulesMap
}

func createDefaultBackendDto(name string) (dto BackendModuleDto) {
	dto.deployModule = true
	if !strings.HasPrefix(name, ManagementModulePattern) {
		dto.deploySidecar = getDefaultDeploySidecar()
	}
	dto.version = nil

	PortStartIndex++
	dto.port = &PortStartIndex
	dto.portServer = getDefaultPortServer()
	dto.environment = make(map[string]any)
	dto.resources = make(map[string]any)

	return dto
}

func createConfigurableBackendDto(value any, name string) (dto BackendModuleDto) {
	mapEntry := value.(map[string]any)

	dto.deployModule = getDeployModule(mapEntry)
	dto.deploySidecar = getDeploySidecar(mapEntry, name)
	dto.version = getVersion(mapEntry)
	dto.port = getPort(mapEntry)
	dto.portServer = getPortServer(mapEntry)
	dto.environment = createEnvironment(mapEntry)
	dto.resources = createResources(mapEntry)

	return dto
}

func getDeployModule(mapEntry map[string]any) bool {
	if mapEntry["deploy-module"] != nil {
		return mapEntry["deploy-module"].(bool)
	}

	return true
}

func getDeploySidecar(mapEntry map[string]any, name string) *bool {
	var sidecar *bool
	if mapEntry["deploy-sidecar"] != nil {
		sidecarValue := mapEntry["deploy-sidecar"].(bool)
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

func getVersion(mapEntry map[string]any) *string {
	var version *string
	if mapEntry["version"] != nil {
		var versionValue string
		_, ok := mapEntry["version"].(float64)
		if ok {
			versionValue = strconv.FormatFloat(mapEntry["version"].(float64), 'f', -1, 64)
		} else {
			versionValue = mapEntry["version"].(string)
		}
		version = &versionValue
	}

	return version
}

func getPort(mapEntry map[string]any) *int {
	var port *int
	if mapEntry["port"] != nil {
		portValue := mapEntry["port"].(int)
		port = &portValue
	} else {
		PortStartIndex++
		port = &PortStartIndex
	}

	return port
}

func getPortServer(mapEntry map[string]any) *int {
	var portServer *int
	if mapEntry["port-server"] != nil {
		portServerValue := mapEntry["port-server"].(int)
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
	if mapEntry["environment"] != nil {
		return mapEntry["environment"].(map[string]any)
	}

	return make(map[string]any)
}

func createResources(mapEntry map[string]any) map[string]any {
	if mapEntry["resources"] != nil {
		return mapEntry["resources"].(map[string]any)
	}

	return make(map[string]any)
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

				if mapEntry["deploy-module"] != nil {
					deployModule = mapEntry["deploy-module"].(bool)
				}

				if mapEntry["version"] != nil {
					versionValue := mapEntry["version"].(string)
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

func DeployUi(tenant string) bool {
	mapEntry, ok := viper.GetStringMap(TenantsKey)[tenant].(map[string]any)
	if !ok {
		return false
	}

	deployUi, ok := mapEntry["deploy-ui"]
	if !ok || deployUi == nil || !deployUi.(bool) {
		return false
	}

	return true
}
