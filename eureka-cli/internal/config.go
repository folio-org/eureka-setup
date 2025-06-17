package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/spf13/viper"
)

const (
	ConfigDir  string = ".eureka"
	ConfigType string = "yaml"

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
	EdgeModulePattern       string = "edge-"

	AllContainerPattern        string = "^eureka-"
	ProfileContainerPattern    string = "^eureka-%s"
	ManagementContainerPattern string = "^eureka-mgr-"
	ModuleContainerPattern     string = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?$"
	SidecarContainerPattern    string = "^eureka-%s-[a-z]+-[a-z]+(-[a-z]{3,})?-sc$"

	SingleModuleOrSidecarContainerPattern string = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
)

const (
	ModSearchModuleName           string = "mod-search"
	ModDataExportWorkerModuleName string = "mod-data-export-worker"

	ElasticsearchContainerName string = "elasticsearch"
	MinioContainerName         string = "minio"
	CreateBucketsContainerName string = "createbuckets"
	FtpServerContainerName     string = "ftp-server"
)

var (
	PortStartIndex int   = 30000
	PortEndIndex   int   = 30999
	ReservedPorts  []int = []int{}

	AvailableProfiles = []string{"combined", "export", "search", "edge"}
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

func GetHomeMiscDir(commandName string) string {
	return filepath.Join(GetHomeDirPath(commandName), DockerComposeWorkDir)
}

func GetBackendModulesFromConfig(commandName string, managementOnly bool, printOutput bool, backendModulesAnyMap map[string]any) map[string]BackendModule {
	if len(backendModulesAnyMap) == 0 {
		if printOutput {
			slog.Info(commandName, GetFuncName(), "No backend modules were found in config")
		}

		return make(map[string]BackendModule)
	}

	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		if managementOnly && !IsManagementModule(name) || !managementOnly && IsManagementModule(name) {
			continue
		}

		backendDto := createBackendDto(commandName, name, value)
		backendModulesMap[name] = *createBackendModule(commandName, backendDto)

		printModuleInfo(commandName, name, backendDto, backendModulesMap, printOutput)
	}

	return backendModulesMap
}

func createBackendDto(commandName string, name string, value any) BackendModuleDto {
	if value == nil {
		return createDefaultBackendDto(commandName, name)
	}

	return createConfigurableBackendDto(commandName, value, name)
}

func createBackendModule(commandName string, dto BackendModuleDto) *BackendModule {
	if dto.deploySidecar != nil && *dto.deploySidecar {
		return NewBackendModuleWithSidecar(commandName, dto)
	}

	return NewBackendModule(commandName, dto)
}

func printModuleInfo(commandName string, name string, dto BackendModuleDto, backendModulesMap map[string]BackendModule, printOutput bool) {
	if !printOutput {
		return
	}

	moduleInfo := name
	if dto.version != nil {
		moduleInfo = fmt.Sprintf("%s with fixed version %s", name, *dto.version)
	}

	moduleServerPort := backendModulesMap[name].ModuleExposedServerPort
	moduleDebugPort := backendModulesMap[name].ModuleExposedDebugPort
	sidecarServerPort := backendModulesMap[name].SidecarExposedServerPort
	sidecarDebugPort := backendModulesMap[name].SidecarExposedDebugPort

	slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found backend module in config: %s, reserved ports: [%d|%d|%d|%d]", moduleInfo, moduleServerPort, moduleDebugPort, sidecarServerPort, sidecarDebugPort))
}

func IsManagementModule(name string) bool {
	return strings.HasPrefix(name, ManagementModulePattern)
}

func IsEdgeModule(name string) bool {
	return strings.HasPrefix(name, EdgeModulePattern)
}

func createDefaultBackendDto(commandName string, name string) (dto BackendModuleDto) {
	dto.deployModule = true

	if !IsManagementModule(name) && !IsEdgeModule(name) {
		dto.deploySidecar = getDefaultDeploySidecar()
	}

	dto.port = getDefaultPort(commandName)
	dto.portServer = getDefaultPortServer()
	dto.environment = make(map[string]any)
	dto.resources = make(map[string]any)
	dto.volumes = []string{}

	return dto
}

func createConfigurableBackendDto(commandName string, value any, name string) (dto BackendModuleDto) {
	mapEntry := value.(map[string]any)

	dto.deployModule = getKeyOrDefault(mapEntry, ModuleDeployModuleEntryKey, true).(bool)

	if !strings.HasPrefix(name, ManagementModulePattern) && !strings.HasPrefix(name, EdgeModulePattern) {
		dto.deploySidecar = getDeploySidecar(mapEntry)
	}

	dto.useVault = getKeyOrDefault(mapEntry, ModuleUseVaultEntryKey, false).(bool)
	dto.disableSystemUser = getKeyOrDefault(mapEntry, ModuleDisableSystemUserEntryKey, false).(bool)
	dto.useOkapiUrl = getKeyOrDefault(mapEntry, ModuleUseOkapiUrlEntryKey, false).(bool)
	dto.version = getVersion(mapEntry)
	dto.port = getPort(commandName, dto.deployModule, mapEntry)
	dto.portServer = getPortServer(mapEntry)
	dto.environment = getKeyOrDefault(mapEntry, ModuleEnvironmentEntryKey, make(map[string]any)).(map[string]any)
	dto.resources = getKeyOrDefault(mapEntry, ModuleResourceEntryKey, make(map[string]any)).(map[string]any)
	dto.volumes = getVolumes(commandName, mapEntry)

	return dto
}

func getKeyOrDefault(mapEntry map[string]any, key string, defaultValue any) any {
	if mapEntry[key] == nil {
		return defaultValue
	}

	return mapEntry[key]
}

func getDeploySidecar(mapEntry map[string]any) *bool {
	if mapEntry[ModuleDeploySidecarEntryKey] == nil {
		return getDefaultDeploySidecar()
	}
	sidecarValue := mapEntry[ModuleDeploySidecarEntryKey].(bool)

	return &sidecarValue
}

func getDefaultDeploySidecar() *bool {
	deploySidecarDefaultValue := true

	return &deploySidecarDefaultValue
}

func getVersion(mapEntry map[string]any) *string {
	if mapEntry[ModuleVersionEntryKey] == nil {
		return nil
	}

	var versionValue string
	_, ok := mapEntry[ModuleVersionEntryKey].(float64)
	if ok {
		versionValue = strconv.FormatFloat(mapEntry[ModuleVersionEntryKey].(float64), 'f', -1, 64)
	} else {
		versionValue = mapEntry[ModuleVersionEntryKey].(string)
	}

	return &versionValue
}

func getPort(commandName string, deployModule bool, mapEntry map[string]any) *int {
	if !deployModule {
		noPort := 0
		return &noPort
	}
	if mapEntry[ModulePortEntryKey] == nil {
		return getDefaultPort(commandName)
	}
	portValue := mapEntry[ModulePortEntryKey].(int)

	return &portValue
}

func getDefaultPort(commandName string) *int {
	freePort := GetAndSetFreePortFromRange(commandName, PortStartIndex, PortEndIndex, &ReservedPorts)

	return &freePort
}

func getPortServer(mapEntry map[string]any) *int {
	if mapEntry[ModulePortServerEntryKey] == nil {
		return getDefaultPortServer()
	}
	portServerValue := mapEntry[ModulePortServerEntryKey].(int)

	return &portServerValue
}

func getDefaultPortServer() *int {
	defaultServerPort, _ := strconv.Atoi(DefaultServerPort)
	portServerValue := defaultServerPort

	return &portServerValue
}

func getVolumes(commandName string, mapEntry map[string]any) []string {
	if mapEntry[ModuleVolumesEntryKey] == nil {
		return []string{}
	}

	var volumes []string
	for _, value := range mapEntry[ModuleVolumesEntryKey].([]any) {
		var volume string = value.(string)
		if runtime.GOOS == "windows" && strings.Contains(volume, "$EUREKA") {
			homeConfigDir := GetHomeDirPath(commandName)
			volume = strings.ReplaceAll(volume, "$EUREKA", homeConfigDir)
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

func GetFrontendModulesFromConfig(commandName string, printOutput bool, frontendModulesAnyMaps ...map[string]any) map[string]FrontendModule {
	if len(frontendModulesAnyMaps) == 0 {
		if printOutput {
			slog.Info(commandName, GetFuncName(), "No frontend modules were found in config")
		}

		return make(map[string]FrontendModule)
	}

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

func GetRequiredContainers(commandName string, requiredContainers []string) []string {
	if CanDeployModule(ModSearchModuleName) {
		requiredContainers = append(requiredContainers, ElasticsearchContainerName)
	}
	if CanDeployModule(ModDataExportWorkerModuleName) {
		requiredContainers = append(requiredContainers, []string{MinioContainerName, CreateBucketsContainerName, FtpServerContainerName}...)
	}
	if len(requiredContainers) > 0 {
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Retrieved required containers: %s", requiredContainers))
	}

	return requiredContainers
}

func CanDeployModule(module string) bool {
	for name, value := range viper.GetStringMap(BackendModulesKey) {
		if name == module {
			if value == nil {
				return true
			}

			deployModule, ok := value.(map[string]any)[ModuleDeployModuleEntryKey]
			if !ok {
				return true
			}

			return deployModule != nil && deployModule.(bool)
		}
	}

	return false
}

func ExtractPortFromUrl(commandName string, url string) int {
	sidecarServer, err := strconv.Atoi(strings.TrimSpace(regexp.MustCompile(ColonDelimitedPattern).ReplaceAllString(url, `$1`)))
	if err != nil {
		slog.Error(commandName, GetFuncName(), "strconv.Atoi error")
		panic(err)
	}

	return sidecarServer
}
