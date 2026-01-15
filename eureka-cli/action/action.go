package action

import (
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Action is a container that holds the state (context) of the current deployment (command being executed)
type Action struct {
	Name                               string
	GatewayURLTemplate                 string
	ReservedPorts                      []int
	Param                              *Param
	Caser                              cases.Caser
	VaultRootToken                     string
	KeycloakAccessToken                string
	KeycloakMasterAccessToken          string
	ConfigProfileName                  string
	ConfigRegistryURL                  string
	ConfigInstallFolio                 string
	ConfigInstallEureka                string
	ConfigPortStart                    int
	ConfigPortEnd                      int
	ConfigManagementTopicSharing       bool
	ConfigTopicSharingTenant           string
	ConfigApplication                  map[string]any
	ConfigApplicationName              string
	ConfigApplicationVersion           string
	ConfigApplicationID                string
	ConfigApplicationFetchDescriptors  bool
	ConfigApplicationPortStart         int
	ConfigApplicationPortEnd           int
	ConfigApplicationDependencies      map[string]any
	ConfigApplicationStripesBranch     string
	ConfigApplicationGatewayHostname   string
	ConfigNamespacePlatformCompleteUI  string
	ConfigGlobalEnv                    map[string]string
	ConfigEnvFolio                     string
	ConfigSidecarModule                map[string]any
	ConfigSidecarModuleResources       map[string]any
	ConfigSidecarModuleNativeBinaryCmd []string
	ConfigBackendModules               map[string]any
	ConfigFrontendModules              map[string]any
	ConfigCustomFrontendModules        map[string]any
	ConfigTenants                      map[string]any
	ConfigRoles                        map[string]any
	ConfigUsers                        map[string]any
	ConfigRolesCapabilitySets          map[string]any
	ConfigConsortiums                  map[string]any
}

func New(name string, gatewayURL string, actionParam *Param) *Action {
	applicationName := viper.GetString(field.ApplicationName)
	applicationVersion := viper.GetString(field.ApplicationVersion)
	return &Action{
		Name:                               name,
		GatewayURLTemplate:                 gatewayURL,
		ReservedPorts:                      []int{},
		Param:                              actionParam,
		Caser:                              cases.Lower(language.English),
		ConfigProfileName:                  viper.GetString(field.ProfileName),
		ConfigRegistryURL:                  viper.GetString(field.RegistryURL),
		ConfigInstallFolio:                 viper.GetString(field.InstallFolio),
		ConfigInstallEureka:                viper.GetString(field.InstallEureka),
		ConfigManagementTopicSharing:       viper.GetBool(field.BackendModulesManagementTopicSharing),
		ConfigTopicSharingTenant:           viper.GetString(field.EnvTopicSharingTenant),
		ConfigApplication:                  viper.GetStringMap(field.Application),
		ConfigApplicationName:              applicationName,
		ConfigApplicationVersion:           applicationVersion,
		ConfigApplicationID:                fmt.Sprintf("%s-%s", applicationName, applicationVersion),
		ConfigApplicationFetchDescriptors:  viper.GetBool(field.ApplicationFetchDescriptors),
		ConfigApplicationPortStart:         viper.GetInt(field.ApplicationPortStart),
		ConfigApplicationPortEnd:           viper.GetInt(field.ApplicationPortEnd),
		ConfigApplicationDependencies:      viper.GetStringMap(field.ApplicationDependencies),
		ConfigApplicationStripesBranch:     viper.GetString(field.ApplicationStripesBranch),
		ConfigApplicationGatewayHostname:   viper.GetString(field.ApplicationGatewayHostname),
		ConfigNamespacePlatformCompleteUI:  viper.GetString(field.NamespacesPlatformCompleteUI),
		ConfigGlobalEnv:                    viper.GetStringMapString(field.Env),
		ConfigEnvFolio:                     viper.GetString(field.EnvFolio),
		ConfigSidecarModule:                viper.GetStringMap(field.SidecarModule),
		ConfigSidecarModuleResources:       viper.GetStringMap(field.SidecarModuleResources),
		ConfigSidecarModuleNativeBinaryCmd: GetSidecarModuleCmd(),
		ConfigBackendModules:               viper.GetStringMap(field.BackendModules),
		ConfigFrontendModules:              viper.GetStringMap(field.FrontendModules),
		ConfigCustomFrontendModules:        viper.GetStringMap(field.CustomFrontendModules),
		ConfigTenants:                      viper.GetStringMap(field.Tenants),
		ConfigRoles:                        viper.GetStringMap(field.Roles),
		ConfigUsers:                        viper.GetStringMap(field.Users),
		ConfigRolesCapabilitySets:          viper.GetStringMap(field.RolesCapabilitySetsEntry),
		ConfigConsortiums:                  viper.GetStringMap(field.Consortiums),
	}
}

// ==================== Request URL ====================

func (a *Action) GetRequestURL(port string, route string) string {
	return fmt.Sprintf(a.GatewayURLTemplate, port) + route
}

// ==================== Application ====================

func (a *Action) IsChildApp() bool {
	return len(a.ConfigApplicationDependencies) > 0
}

// ==================== Environment ====================

func GetSidecarModuleCmd() []string {
	nativeBinaryCmd := viper.GetStringSlice(field.SidecarModuleNativeBinaryCmd)
	if len(nativeBinaryCmd) > 0 {
		return nativeBinaryCmd
	}

	return viper.GetStringSlice(field.SidecarModuleCmd)
}

func (a *Action) GetConfigEnvVars(key string) []string {
	var envVars []string
	for key, value := range viper.GetStringMapString(key) {
		envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return envVars
}

// ==================== Reserve Ports ====================

func (a *Action) GetPreReservedPortSet(n int) (ports []int, err error) {
	for range n {
		port, err := a.GetPreReservedPort()
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}

	return ports, nil
}

func (a *Action) GetPreReservedPort() (int, error) {
	var freePort int
	for port := a.ConfigApplicationPortStart; port <= a.ConfigApplicationPortEnd; port++ {
		if a.isPortFree(a.ConfigApplicationPortStart, a.ConfigApplicationPortEnd, port) && !slices.Contains(a.ReservedPorts, port) {
			freePort = port
			break
		}
	}
	if freePort == 0 {
		return 0, errors.NoFreeTCPPort(a.ConfigApplicationPortStart, a.ConfigApplicationPortEnd)
	}
	a.ReservedPorts = append(a.ReservedPorts, freePort)

	return freePort, nil
}

func (a *Action) isPortFree(portStart, portEnd int, port int) bool {
	tcpListen, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(port)))
	if err != nil {
		slog.Debug(a.Name, "text", "TCP port is reserved or already bound in range", "target", port, "start", portStart, "end", portEnd)
		return false
	}
	defer func() { _ = tcpListen.Close() }()

	return true
}

// ==================== Install Json URLs ====================

func (a *Action) GetCombinedInstallJsonURLs() map[string]string {
	return map[string]string{
		constant.FolioRegistry:  a.ConfigInstallFolio,
		constant.EurekaRegistry: a.ConfigInstallEureka,
	}
}

func (a *Action) GetEurekaInstallJsonURLs() map[string]string {
	return map[string]string{
		constant.EurekaRegistry: a.ConfigInstallEureka,
	}
}

// ==================== Module URL  ====================

func (a *Action) GetModuleURL(moduleID string) string {
	return fmt.Sprintf("%s/_/proxy/modules/%s", a.ConfigRegistryURL, moduleID)
}

// ==================== Viper ====================

func GetConfigEnv(key string, env map[string]string) string {
	return env[strings.ToLower(key)]
}

func IsSet(key string) bool {
	return viper.IsSet(key)
}

func (a *Action) GetKafkaTopicConfigTenant(configTenant string) string {
	if a.ConfigManagementTopicSharing {
		return a.ConfigTopicSharingTenant
	}

	return configTenant
}
