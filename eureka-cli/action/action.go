package action

import (
	"fmt"
	"log/slog"
	"net"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Action is a container that holds the state (context) of current deployment
type Action struct {
	Name                              string
	GatewayURLTemplate                string
	ReservedPorts                     []int
	Params                            *actionparams.ActionParams
	Caser                             cases.Caser
	VaultRootToken                    string
	KeycloakAccessToken               string
	KeycloakMasterAccessToken         string
	ConfigProfile                     string
	ConfigRegistryURL                 string
	ConfigFolioRegistry               string
	ConfigEurekaRegistry              string
	ConfigPortStart                   int
	ConfigPortEnd                     int
	ConfigApplication                 map[string]any
	ConfigApplicationName             string
	ConfigApplicationVersion          string
	ConfigApplicationID               string
	ConfigApplicationPlatform         string
	ConfigApplicationFetchDescriptors bool
	ConfigApplicationPortStart        int
	ConfigApplicationPortEnd          int
	ConfigApplicationDependencies     map[string]any
	ConfigApplicationStripesBranch    string
	ConfigApplicationGatewayHostname  string
	ConfigNamespacePlatformCompleteUI string
	ConfigGlobalEnv                   map[string]string
	ConfigEnvFolio                    string
	ConfigSidecarModule               map[string]any
	ConfigSidecarResources            map[string]any
	ConfigBackendModules              map[string]any
	ConfigFrontendModules             map[string]any
	ConfigCustomFrontendModules       map[string]any
	ConfigTenants                     map[string]any
	ConfigRoles                       map[string]any
	ConfigUsers                       map[string]any
	ConfigRolesCapabilitySets         map[string]any
	ConfigConsortiums                 map[string]any
}

func New(name, gatewayURL string, actionParams *actionparams.ActionParams) *Action {
	return newGeneric(name, gatewayURL, actionParams, "", "", "")
}

func NewWithCredentials(name, gatewayURL string, actionParams *actionparams.ActionParams, vaultRootToken, keycloakAccessToken, keycloakMasterAccessToken string) *Action {
	return newGeneric(name, gatewayURL, actionParams, vaultRootToken, keycloakAccessToken, keycloakMasterAccessToken)
}

func newGeneric(name, gatewayURL string, actionParams *actionparams.ActionParams, vaultRootToken, keycloakAccessToken, keycloakMasterAccessToken string) *Action {
	applicationName := viper.GetString(field.ApplicationName)
	applicationVersion := viper.GetString(field.ApplicationVersion)

	return &Action{
		Name:                              name,
		GatewayURLTemplate:                gatewayURL,
		ReservedPorts:                     []int{},
		Params:                            actionParams,
		VaultRootToken:                    vaultRootToken,
		KeycloakAccessToken:               keycloakAccessToken,
		KeycloakMasterAccessToken:         keycloakMasterAccessToken,
		Caser:                             cases.Lower(language.English),
		ConfigProfile:                     viper.GetString(field.ProfileName),
		ConfigRegistryURL:                 viper.GetString(field.RegistryURL),
		ConfigFolioRegistry:               viper.GetString(field.InstallFolio),
		ConfigEurekaRegistry:              viper.GetString(field.InstallEureka),
		ConfigApplication:                 viper.GetStringMap(field.Application),
		ConfigApplicationName:             applicationName,
		ConfigApplicationVersion:          applicationVersion,
		ConfigApplicationID:               fmt.Sprintf("%s-%s", applicationName, applicationVersion),
		ConfigApplicationPlatform:         viper.GetString(field.ApplicationPlatform),
		ConfigApplicationFetchDescriptors: viper.GetBool(field.ApplicationFetchDescriptors),
		ConfigApplicationPortStart:        viper.GetInt(field.ApplicationPortStart),
		ConfigApplicationPortEnd:          viper.GetInt(field.ApplicationPortEnd),
		ConfigApplicationDependencies:     viper.GetStringMap(field.ApplicationDependencies),
		ConfigApplicationStripesBranch:    viper.GetString(field.ApplicationStripesBranch),
		ConfigApplicationGatewayHostname:  viper.GetString(field.ApplicationGatewayHostname),
		ConfigNamespacePlatformCompleteUI: viper.GetString(field.NamespacesPlatformCompleteUI),
		ConfigGlobalEnv:                   viper.GetStringMapString(field.Env),
		ConfigEnvFolio:                    viper.GetString(field.EnvFolio),
		ConfigSidecarModule:               viper.GetStringMap(field.SidecarModule),
		ConfigSidecarResources:            viper.GetStringMap(field.SidecarModuleResources),
		ConfigBackendModules:              viper.GetStringMap(field.BackendModules),
		ConfigFrontendModules:             viper.GetStringMap(field.FrontendModules),
		ConfigCustomFrontendModules:       viper.GetStringMap(field.CustomFrontendModules),
		ConfigTenants:                     viper.GetStringMap(field.Tenants),
		ConfigRoles:                       viper.GetStringMap(field.Roles),
		ConfigUsers:                       viper.GetStringMap(field.Users),
		ConfigRolesCapabilitySets:         viper.GetStringMap(field.RolesCapabilitySetsEntry),
		ConfigConsortiums:                 viper.GetStringMap(field.Consortiums),
	}
}

func GetGatewayURLTemplate(actionName string) (string, error) {
	gatewayURL, err := GetGatewayURL(actionName)
	if err != nil {
		return "", err
	}

	return gatewayURL + ":%s", nil
}

func GetGatewayURL(actionName string) (string, error) {
	slog.Debug(actionName, "text", "RETRIEVING GATEWAY URL")
	gatewayURL, err := getConfigGatewayURL(actionName)
	if gatewayURL == "" {
		gatewayURL, err = getDefaultGatewayURL(actionName)
	}
	if gatewayURL == "" {
		gatewayURL, err = getOtherGatewayURL(actionName)
	}
	if gatewayURL == "" || err != nil {
		return "", errors.GatewayURLConstructFailed(runtime.GOOS, err)
	}
	slog.Debug(actionName, "text", "Retrieved gateway URL", "url", gatewayURL)

	return gatewayURL, nil
}

func getConfigGatewayURL(actionName string) (gatewayURL string, err error) {
	if !viper.IsSet(field.ApplicationGatewayHostname) {
		return "", nil
	}

	hostname := viper.GetString(field.ApplicationGatewayHostname)
	err = helpers.IsHostnameReachable(actionName, hostname)
	if err != nil {
		slog.Warn(actionName, "text", "Retrieving config gateway hostname was unsuccessful", "error", err)
		return "", err
	}

	if !strings.HasPrefix(hostname, "http://") {
		gatewayURL = fmt.Sprintf("http://%s", hostname)
	} else {
		gatewayURL = hostname
	}

	return gatewayURL, nil
}

func getDefaultGatewayURL(actionName string) (gatewayURL string, err error) {
	err = helpers.IsHostnameReachable(actionName, constant.DockerHostname)
	if err != nil {
		slog.Warn(actionName, "text", "Retrieving default gateway URL was unsuccessful", "error", err)
		return "", nil
	}

	return fmt.Sprintf("http://%s", constant.DockerHostname), nil
}

func getOtherGatewayURL(actionName string) (gatewayURL string, err error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		err = errors.UnsupportedPlatform(runtime.GOOS, constant.DockerGatewayIP)
		slog.Warn(actionName, "text", "Retrieving other gateway URL was unsuccessful", "error", err)
		return "", err
	}

	return fmt.Sprintf("http://%s", constant.DockerGatewayIP), nil
}

func (a *Action) GetPreReserverPortSet(fns []func() (int, error)) (ports []int, err error) {
	for _, fn := range fns {
		port, err := fn()
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
		slog.Debug(a.Name, "text", "TCP port is reserved or already bound in range", "port", port, "portStart", portStart, "portEnd", portEnd)
		return false
	}
	defer a.closeListener(tcpListen)

	return true
}

func (a *Action) closeListener(listener net.Listener) {
	_ = listener.Close()
}

func (a *Action) GetRequestURL(port string, route string) string {
	return fmt.Sprintf(a.GatewayURLTemplate, port) + route
}

func (a *Action) GetConfigEnvVars(key string) []string {
	var envVars []string
	for key, value := range viper.GetStringMapString(key) {
		envVars = append(envVars, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
	}

	return envVars
}

func GetConfigEnv(key string, configGlobalEnv map[string]string) string {
	return configGlobalEnv[strings.ToLower(key)]
}

func IsSet(key string) bool {
	return viper.IsSet(key)
}

func (a *Action) GetCombinedInstallJsonURLs() map[string]string {
	return map[string]string{
		constant.FolioRegistry:  a.ConfigFolioRegistry,
		constant.EurekaRegistry: a.ConfigEurekaRegistry,
	}
}

func (a *Action) GetEurekaInstallJsonURLs() map[string]string {
	return map[string]string{
		constant.EurekaRegistry: a.ConfigEurekaRegistry,
	}
}

func (a *Action) GetCombinedRegistryURLs() map[string]string {
	return map[string]string{
		constant.FolioRegistry:  a.ConfigRegistryURL,
		constant.EurekaRegistry: a.ConfigRegistryURL,
	}
}
