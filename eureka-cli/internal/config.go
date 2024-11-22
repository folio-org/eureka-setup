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

	DockerComposeWorkDir  string = "./misc"
	NetworkName           string = "fpm-net"
	NetworkId             string = "eureka"
	HostDockerInternalUrl string = "http://host.docker.internal:%d%s"
	HostIp                string = "0.0.0.0"
	DefaultServerPort     string = "8081"
	DefaultDebugPort      string = "5005"

	FolioKeycloakRepositoryUrl    string = "https://github.com/folio-org/folio-keycloak"
	FolioKongRepositoryUrl        string = "https://github.com/folio-org/folio-kong"
	PlatformCompleteRepositoryUrl string = "https://github.com/folio-org/platform-complete.git"
)

const (
	ManagementModulePattern string = "mgr-"

	SingleModuleContainerPattern string = "^(eureka-%s-)(%[2]s|%[2]s-sc)$"
)

var PortStartIndex int = 30000

func GetBackendModulesFromConfig(commandName string, backendModulesAnyMap map[string]any, managementOnly bool) map[string]BackendModule {
	const (
		DeployModuleKey = "deploy-module"
		VersionKey      = "version"
		PortKey         = "port"
		PortKeyServer   = "port-server"
		SidecarKey      = "sidecar"
		ModuleEnvKey    = "environment"
	)

	backendModulesMap := make(map[string]BackendModule)

	for name, value := range backendModulesAnyMap {
		var (
			deployModule bool = true
			version      *string
			port         *int
			portServer   *int
			sidecar      *bool
			environment  map[string]interface{}
		)

		getDefaultPortServer := func() *int {
			defaultServerPort, _ := strconv.Atoi(DefaultServerPort)
			portServerValue := defaultServerPort
			return &portServerValue
		}

		getDefaultPort := func() *int {
			PortStartIndex++
			return &PortStartIndex
		}

		getDefaultSidecar := func() *bool {
			sidecarDefaultValue := true
			return &sidecarDefaultValue
		}

		if value == nil {
			port = getDefaultPort()
			portServer = getDefaultPortServer()
			// Avoid deploying a sidecar for mgr-* modules
			if !strings.HasPrefix(name, ManagementModulePattern) {
				sidecar = getDefaultSidecar()
			}
			environment = make(map[string]interface{})
		} else {
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
			} else {
				port = getDefaultPort()
			}

			if mapEntry[PortKeyServer] != nil {
				portServerValue := mapEntry[PortKeyServer].(int)
				portServer = &portServerValue
			} else {
				portServer = getDefaultPortServer()
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
		}

		if sidecar != nil && *sidecar {
			backendModulesMap[name] = *NewBackendModuleAndSidecar(deployModule, name, version, *port, *portServer, *sidecar, environment)
		} else {
			backendModulesMap[name] = *NewBackendModule(name, version, *port, *portServer, environment)
		}

		moduleInfo := name
		if version != nil {
			moduleInfo = fmt.Sprintf("%s with fixed version %s", name, *version)
		}

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found backend module in config: %s", moduleInfo))
	}

	return backendModulesMap
}

func GetFrontendModulesFromConfig(commandName string, frontendModulesAnyMap map[string]any) map[string]FrontendModule {
	const (
		DeployModuleKey = "deploy-module"
		VersionKey      = "version"
	)

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

		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Found frontend module in config: %s", moduleInfo))
	}

	return frontendModulesMap
}

func PrepareStripesConfigJs(commandName string, configPath string, tenant string, kongUrl string, keycloakUrl string, platformCompleteUrl string) {
	stripesConfigJsFilePath := fmt.Sprintf("%s/stripes.config.js", configPath)
	readFileBytes, err := os.ReadFile(stripesConfigJsFilePath)
	if err != nil {
		slog.Error(commandName, GetFuncName(), "os.ReadFile error")
		panic(err)
	}

	replaceMap := map[string]string{"${kongUrl}": kongUrl,
		"${tenantUrl}":      platformCompleteUrl,
		"${keycloakUrl}":    keycloakUrl,
		"${hasAllPerms}":    `false`,
		"${isSingleTenant}": `true`,
		"${tenantOptions}":  fmt.Sprintf(`{%[1]s: {name: "%[1]s", clientId: "%[1]s%s"}}`, tenant, GetEnvironmentFromMapByKey("KC_LOGIN_CLIENT_SUFFIX")),
	}

	var newReadFileStr string = string(readFileBytes)
	for key, value := range replaceMap {
		newReadFileStr = strings.Replace(newReadFileStr, key, value, -1)
	}

	fmt.Println(newReadFileStr)

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

	updates := 0
	modules := []string{"@folio/authorization-policies", "@folio/authorization-roles", "@folio/plugin-select-application"}
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
	var stripesBranch plumbing.ReferenceName

	if viper.IsSet(ApplicationStripesBranch) {
		branchStr := viper.GetString(ApplicationStripesBranch)
		stripesBranch = plumbing.ReferenceName(branchStr)
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("Got stripes branch from config: %s", stripesBranch))
	} else {
		stripesBranch = defaultBranch
		slog.Info(commandName, GetFuncName(), fmt.Sprintf("No stripes branch in config. Using default branch: %s", stripesBranch))
	}

	return stripesBranch
}

func HasTenant(tenant string) bool {
	return slices.Contains(ConvertMapKeysToSlice(viper.GetStringMap(TenantsKey)), tenant)
}

func DeployUi(tenant string) bool {
	deployUi := viper.GetStringMap(TenantsKey)[tenant].(map[string]interface{})["deploy-ui"]

	if deployUi == nil || !deployUi.(bool) {
		return false
	}

	return true
}
