package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/viper"
)

const (
	WorkDir             string = ".eureka"
	ApplicationJson     string = "application/json"
	DockerInternalUrl   string = "http://host.docker.internal:%d%s"
	VaultTokenPattern   string = `.*:`
	ModuleIdPattern     string = `([a-z-_]+)([\d-_.]+)([a-zA-Z0-9-_.]+)`
	EnvNamePattern      string = `[.-]+`
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

	HostIpValueKey     string = "0.0.0.0"
	ServerPortValueKey string = "8081"
	DebugPortValueKey  string = "5005"
)

var (
	VaultTokenRegexp = regexp.MustCompile(VaultTokenPattern)
	ModuleIdRegexp   = regexp.MustCompile(ModuleIdPattern)
	EnvNameRegexp    = regexp.MustCompile(EnvNamePattern)
)

type RegistryModule struct {
	Id      string `json:"id"`
	Action  string `json:"action"`
	Name    string
	Version string
}

type RegistryModules []RegistryModule

type ModuleDescriptor struct {
}

type BackendModule struct {
	DeployModule       bool
	ModuleName         string
	ModuleExposedPorts *nat.PortSet
	ModulePortBindings *nat.PortMap
	ModuleEnv          map[string]interface{}

	DeploySidecar       bool
	SidecarExposedPorts *nat.PortSet
	SidecarPortBindings *nat.PortMap
}

type Applications struct {
	ApplicationDescriptors []map[string]interface{} `json:"applicationDescriptors"`
	TotalRecords           int                      `json:"totalRecords"`
}

type RegisterModuleDto struct {
	RegistryUrls              map[string]string
	RegistryModules           map[string][]RegistryModule
	BackendModulesMap         map[string]BackendModule
	ModuleDescriptorsMap      map[string]interface{}
	CacheFileModuleEnvPointer *os.File
	EnableDebug               bool
}

type DeployModuleDto struct {
	Name          string
	Image         string
	RegistryAuth  string
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
	Platform      *v1.Platform
	PullImage     bool
}

type DeployModulesDto struct {
	VaultToken              string
	RegistryHostname        map[string]string
	EurekaRegistryAuthToken string
	RegistryModules         map[string][]RegistryModule
	BackendModulesMap       map[string]BackendModule
	SharedEnv               []string
}

func NewBackendModuleWithSidecar(name string, port int, moduleEnv map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)
	sidecarPortBindings := CreatePortBindings(port+2000, port+3000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnv, true, exposedPorts, sidecarPortBindings}
}

func NewBackendModule(name string, port int, moduleEnv map[string]interface{}) *BackendModule {
	exposedPorts := CreateExposedPorts()
	modulePortBindings := CreatePortBindings(port, port+1000)

	return &BackendModule{true, name, exposedPorts, modulePortBindings, moduleEnv, false, nil, nil}
}

func CreateExposedPorts() *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})
	moduleExposedPorts[nat.Port(ServerPortValueKey)] = struct{}{}
	moduleExposedPorts[nat.Port(DebugPortValueKey)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int) *nat.PortMap {
	var serverPortBinding []nat.PortBinding
	var serverDebugPortBinding []nat.PortBinding

	serverPortBinding = append(serverPortBinding, nat.PortBinding{HostIP: HostIpValueKey, HostPort: strconv.Itoa(hostServerPort)})
	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{HostIP: HostIpValueKey, HostPort: strconv.Itoa(hostServerDebugPort)})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(ServerPortValueKey)] = serverPortBinding
	portBindings[nat.Port(DebugPortValueKey)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func GetSharedEnvFromConfig(commandName string, sharedEnvMap map[string]string) []string {
	var sharedEnv []string

	for name, value := range sharedEnvMap {
		env := fmt.Sprintf("%s=%s", strings.ToUpper(name), value)

		slog.Info(commandName, "Found shared ENV", env)

		sharedEnv = append(sharedEnv, env)
	}

	return sharedEnv
}

func AppendModuleEnv(envMap map[string]interface{}, moduleEnv []string) []string {
	if len(envMap) > 0 {
		for key, value := range envMap {
			moduleEnv = append(moduleEnv, fmt.Sprintf("%s=%s", strings.ToUpper(key), value))
		}
	}

	return moduleEnv
}

func AppendVaultEnv(moduleEnv []string, vaultToken string, vaultUrl string) []string {
	moduleEnv = append(moduleEnv, "SECRET_STORE_TYPE=VAULT")
	moduleEnv = append(moduleEnv, fmt.Sprintf("SECRET_STORE_VAULT_TOKEN=%s", vaultToken))
	moduleEnv = append(moduleEnv, fmt.Sprintf("SECRET_STORE_VAULT_ADDRESS=%s", vaultUrl))

	return moduleEnv
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

func GetModulesFromRegistries(commandName string, installJsonUrls map[string]string) map[string][]RegistryModule {
	registryModulesMap := make(map[string][]RegistryModule)

	for registryName, installJsonUrl := range installJsonUrls {
		var registryModules []RegistryModule

		installJsonResp, err := http.Get(installJsonUrl)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "http.Get error")
			panic(err)
		}
		defer installJsonResp.Body.Close()

		err = json.NewDecoder(installJsonResp.Body).Decode(&registryModules)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
			panic(err)
		}

		if len(registryModules) > 0 {
			slog.Info(commandName, fmt.Sprintf("Found %s registry modules", registryName), len(registryModules))

			sort.Slice(registryModules, func(i, j int) bool {
				switch strings.Compare(registryModules[i].Id, registryModules[j].Id) {
				case -1:
					return true
				case 1:
					return false
				}

				return registryModules[i].Id > registryModules[j].Id
			})
		}

		registryModulesMap[registryName] = registryModules
	}

	return registryModulesMap
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

func DumpHttpRequest(commandName string, req *http.Request) {
	dumpResp, err := httputil.DumpRequest(req, true)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "httputil.DumpRequest error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Request")
	fmt.Println(string(dumpResp))
}

func DumpHttpResponse(commandName string, resp *http.Response) {
	dumpResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "httputil.DumpResponse error")
		panic(err)
	}

	fmt.Println("### Dumping HTTP Response")
	fmt.Println(string(dumpResp))
}

func DeregisterModules(commandName string, moduleName string, enableDebug bool) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(DockerInternalUrl, 9901, "/applications"), nil)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
		panic(err)
	}

	if enableDebug {
		DumpHttpRequest(commandName, req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer resp.Body.Close()

	if enableDebug {
		DumpHttpResponse(commandName, resp)
	}

	var apps Applications

	err = json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
		panic(err)
	}

	if apps.TotalRecords == 0 {
		slog.Info(commandName, SecondaryMessageKey, "No deployed module applications were found")
	}

	if apps.TotalRecords > 0 {
		for _, v := range apps.ApplicationDescriptors {
			id := v["id"].(string)

			if moduleName != "" {
				moduleNameFiltered := ModuleIdRegexp.ReplaceAllString(id, `$1`)

				if moduleNameFiltered[strings.LastIndex(moduleNameFiltered, "-")] == 45 {
					moduleNameFiltered = moduleNameFiltered[:strings.LastIndex(moduleNameFiltered, "-")]
				}

				if moduleNameFiltered != moduleName {
					continue
				}
			}

			slog.Info(commandName, "Deregistering application", id)

			delAppReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(DockerInternalUrl, 9901, fmt.Sprintf("/applications/%s", id)), nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, delAppReq)
			}

			delAppResp, err := http.DefaultClient.Do(delAppReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer delAppResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, delAppResp)
			}
		}
	}
}

func RegisterModules(commandName string, enableDebug bool, dto *RegisterModuleDto) {
	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Registering %s registry modules", registryName))

		for moduleIndex, module := range registryModules {
			module.Name = ModuleIdRegexp.ReplaceAllString(module.Id, `$1`)
			module.Version = ModuleIdRegexp.ReplaceAllString(module.Id, `$2$3`)

			if module.Name[strings.LastIndex(module.Name, "-")] == 45 {
				module.Name = module.Name[:strings.LastIndex(module.Name, "-")]
			}

			registryModules[moduleIndex] = module

			_, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			moduleNameEnv := EnvNameRegexp.ReplaceAllString(strings.ToUpper(module.Name), `_`)

			_, err := dto.CacheFileModuleEnvPointer.WriteString(fmt.Sprintf("export %s_VERSION=%s\r\n", moduleNameEnv, module.Version))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "moduleEnvVarsFile.WriteString error")
				panic(err)
			}

			var moduleDescriptorsUrl string

			if registryName == "folio" {
				moduleDescriptorsUrl = fmt.Sprintf("%s/_/proxy/modules/%s", dto.RegistryUrls["folio"], module.Id)
			} else {
				moduleDescriptorsUrl = fmt.Sprintf("%s/descriptors/%s.json", dto.RegistryUrls["eureka"], module.Id)
			}

			moduleDescriptorsReq, err := http.NewRequest(http.MethodGet, moduleDescriptorsUrl, nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, moduleDescriptorsReq)
			}

			moduleDescriptorsResp, err := http.DefaultClient.Do(moduleDescriptorsReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer moduleDescriptorsResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, moduleDescriptorsResp)
			}

			var moduleDescriptors interface{}

			err = json.NewDecoder(moduleDescriptorsResp.Body).Decode(&moduleDescriptors)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
				panic(err)
			}

			dto.ModuleDescriptorsMap[module.Id] = moduleDescriptors

			var appModules []map[string]string
			appModules = append(appModules, map[string]string{"name": module.Name, "version": module.Version})

			appBytes, err := json.Marshal(map[string]interface{}{
				"id":                module.Id,
				"version":           module.Version,
				"name":              module.Name,
				"description":       "Deployed by Eureka CLI",
				"modules":           appModules,
				"moduleDescriptors": moduleDescriptors,
			})
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
				panic(err)
			}

			if enableDebug {
				fmt.Println("### Dumping HTTP Request Body")
				fmt.Println(string(appBytes))
			}

			postAppReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, 9901, "/applications"), bytes.NewBuffer(appBytes))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			postAppReq.Header.Add("Content-Type", ApplicationJson)

			if enableDebug {
				DumpHttpRequest(commandName, postAppReq)
			}

			postAppResp, err := http.DefaultClient.Do(postAppReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer postAppResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, postAppResp)
			}

			var appDiscModules []map[string]string
			appDiscModules = append(appDiscModules, map[string]string{
				"id":       module.Id,
				"name":     module.Name,
				"version":  module.Version,
				"location": fmt.Sprintf("http://%s.eureka:8081", module.Name),
			})

			appDiscInfoBytes, err := json.Marshal(map[string]interface{}{
				"discovery": appDiscModules,
			})
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
				panic(err)
			}

			if enableDebug {
				fmt.Println("### Dumping HTTP Request Body")
				fmt.Println(string(appDiscInfoBytes))
			}

			postAppDiscReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, 9901, "/modules/discovery"), bytes.NewBuffer(appDiscInfoBytes))
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			postAppDiscReq.Header.Add("Content-Type", ApplicationJson)

			if enableDebug {
				DumpHttpRequest(commandName, postAppDiscReq)
			}

			postAppDiscResp, err := http.DefaultClient.Do(postAppDiscReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer postAppDiscResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, postAppResp)
			}

			slog.Info(commandName, "Registered module", fmt.Sprintf("%s %d %d", module.Id, postAppResp.StatusCode, postAppDiscResp.StatusCode))
		}
	}
}

func CreateContainerdCli(commandName string) *client.Client {
	containerdCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "client.NewClientWithOpts error")
		panic(err)
	}

	return containerdCli
}

func GetEurekaRegistryAuthToken(commandName string) string {
	session, err := session.NewSession()
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "session.NewSession() error")
		panic(err)
	}

	ecrClient := ecr.New(session)

	authToken, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		panic(err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "base64.StdEncoding.DecodeString error")
		panic(err)
	}

	authCreds := strings.Split(string(decodedBytes), ":")

	jsonBytes, err := json.Marshal(map[string]string{"username": authCreds[0], "password": authCreds[1]})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
		panic(err)
	}

	encodedAuth := base64.StdEncoding.EncodeToString(jsonBytes)

	slog.Info(commandName, SecondaryMessageKey, "Created registry auth token")

	return encodedAuth
}

func GetVaultToken(commandName string, containerdCli *client.Client) string {
	logStream, err := containerdCli.ContainerLogs(context.Background(), "vault", container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerLogs error")
		panic(err)
	}
	defer logStream.Close()

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "logStream.Read(buffer) error")
			panic(err)
		}

		count := binary.BigEndian.Uint32(buffer[4:])
		rawLogLine := make([]byte, count)

		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Warn(commandName, SecondaryMessageKey, "logStream.Read(rawLogLine) encoutered an EOF")
		}

		parsedLogLine := string(rawLogLine)

		if strings.Contains(parsedLogLine, "init.sh: Root VAULT TOKEN is:") {
			vaultToken := strings.TrimSpace(VaultTokenRegexp.ReplaceAllString(parsedLogLine, `$1`))

			slog.Info(commandName, "Found Vault Token", vaultToken)

			return vaultToken
		}
	}
}

func GetDeployedModules(commandName string, containerdCli *client.Client, filters filters.Args) []types.Container {
	deployedModules, err := containerdCli.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerList error")
		panic(err)
	}

	if len(deployedModules) == 0 {
		slog.Info(commandName, SecondaryMessageKey, "No deployed module containers were found")
	}

	return deployedModules
}

func DeployModule(commandName string, containerdCli *client.Client, dto *DeployModuleDto) {
	containerName := fmt.Sprintf("eureka-%s", dto.Name)

	if dto.PullImage {
		reader, err := containerdCli.ImagePull(context.Background(), dto.Image, types.ImagePullOptions{RegistryAuth: dto.RegistryAuth})
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "cli.ImagePull error")
			panic(err)
		}
		defer reader.Close()

		decoder := json.NewDecoder(reader)

		type Event struct {
			Status         string `json:"status"`
			Error          string `json:"error"`
			Progress       string `json:"progress"`
			ProgressDetail struct {
				Current int `json:"current"`
				Total   int `json:"total"`
			} `json:"progressDetail"`
		}

		var event *Event
		for {
			if err := decoder.Decode(&event); err != nil {
				if err == io.EOF {
					break
				}

				panic(err)
			}

			if event.Error == "" {
				slog.Info(commandName, "Pulling module container image", fmt.Sprintf("%+v", event.Status))
			} else {
				slog.Error(commandName, "Pulling module container image", fmt.Sprintf("%+v", event.Error))
			}
		}
	}

	cr, err := containerdCli.ContainerCreate(context.Background(), dto.Config, dto.HostConfig, dto.NetworkConfig, dto.Platform, containerName)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerCreate error")
		panic(err)
	}

	if cr.Warnings != nil && len(cr.Warnings) > 0 {
		slog.Warn(commandName, SecondaryMessageKey, fmt.Sprintf("cli.ContainerCreate warnings=\"%s\"", cr.Warnings))
	}

	err = containerdCli.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerStart error")
		panic(err)
	}

	slog.Info(commandName, "Deployed module container", fmt.Sprintf("%s %s", containerName, cr.ID))
}

func UndeployModule(commandName string, containerdCli *client.Client, deployedModule types.Container) {
	err := containerdCli.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerStop error")
		panic(err)
	}

	err = containerdCli.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "cli.ContainerRemove error")
		panic(err)
	}

	slog.Info(commandName, "Undeployed module container", fmt.Sprintf("%s %s %s", deployedModule.ID, deployedModule.Image, deployedModule.Status))
}

func DeployModules(commandName string, containerdCli *client.Client, dto *DeployModulesDto) {
	eurekaSidecarImage := fmt.Sprintf("%s/%s", viper.GetString(EurekaRegistryHostnameKey), viper.GetString(EurekaSidecarImageKey))

	resourceUrlKeycloak := viper.GetString(ResourceUrlKeycloakKey)
	resourceUrlVault := viper.GetString(ResourceUrlVaultKey)
	resourceUrlMgrTenants := viper.GetString(ResourceUrlMgrTenantsKey)
	resourceUrlMgrApplications := viper.GetString(ResourceUrlMgrApplicationsKey)
	resourceUrlMgrTenantEntitlements := viper.GetString(ResourceUrlMgrTenantEntitlements)

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{"fpm-net": {NetworkID: "eureka"}},
	}

	pullSidecarImage := true

	for registryName, registryModules := range dto.RegistryModules {
		slog.Info(commandName, PrimaryMessageKey, fmt.Sprintf("Deploying %s registry modules", registryName))

		for _, module := range registryModules {
			backendModule, ok := dto.BackendModulesMap[module.Name]
			if !ok {
				continue
			}

			moduleNameEnv := EnvNameRegexp.ReplaceAllString(strings.ToUpper(module.Name), `_`)

			var image string

			if registryName == "folio" {
				image = fmt.Sprintf("%s/%s:%s", dto.RegistryHostname["folio"], module.Name, module.Version)
			} else {
				image = fmt.Sprintf("%s/%s:%s", dto.RegistryHostname["eureka"], module.Name, module.Version)
			}

			var moduleEnv []string

			// Shared ENV
			moduleEnv = append(moduleEnv, dto.SharedEnv...)

			// Module ENV
			moduleEnv = AppendModuleEnv(backendModule.ModuleEnv, moduleEnv)

			// Vault ENV
			moduleEnv = AppendVaultEnv(moduleEnv, dto.VaultToken, resourceUrlVault)

			var moduleRegistryAuthToken string
			if registryName == "folio" {
				moduleRegistryAuthToken = ""
			} else {
				moduleRegistryAuthToken = dto.EurekaRegistryAuthToken
			}

			deployModuleDto := &DeployModuleDto{
				Name:         module.Name,
				Image:        image,
				RegistryAuth: moduleRegistryAuthToken,
				Config: &container.Config{
					Image:        image,
					Hostname:     strings.ToLower(moduleNameEnv),
					Env:          moduleEnv,
					ExposedPorts: *backendModule.ModuleExposedPorts,
				},
				HostConfig: &container.HostConfig{
					PortBindings:  *backendModule.ModulePortBindings,
					RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyOnFailure, MaximumRetryCount: 3},
				},
				NetworkConfig: networkConfig,
				Platform:      &v1.Platform{},
				PullImage:     true,
			}

			// TODO Use mutex and make parallel
			DeployModule(commandName, containerdCli, deployModuleDto)

			if backendModule.DeploySidecar {
				var sidecarEnv []string

				// Keycloak ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("KC_URL=%s", resourceUrlKeycloak))
				sidecarEnv = append(sidecarEnv, "KC_ADMIN_CLIENT_ID=superAdmin")
				sidecarEnv = append(sidecarEnv, "KC_SERVICE_CLIENT_ID=m2m-client")
				sidecarEnv = append(sidecarEnv, "KC_LOGIN_CLIENT_SUFFIX=login-app")

				// Vault ENV
				sidecarEnv = AppendVaultEnv(sidecarEnv, dto.VaultToken, resourceUrlVault)

				// Management ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("TM_CLIENT_URL=%s", resourceUrlMgrTenants))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("AM_CLIENT_URL=%s", resourceUrlMgrApplications))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("TE_CLIENT_URL=%s", resourceUrlMgrTenantEntitlements))

				// Module ENV
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_NAME=%s", module.Name))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_VERSION=%s", module.Version))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("MODULE_URL=http://%s.eureka:8081", module.Name))
				sidecarEnv = append(sidecarEnv, fmt.Sprintf("SIDECAR_URL=http://%s-sc.eureka:8081", module.Name))

				// Java ENV
				sidecarEnv = append(sidecarEnv, "JAVA_OPTIONS=-XX:MaxRAMPercentage=85.0 -Xms50m -Xmx128m -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005")

				deploySidecarDto := &DeployModuleDto{
					Name:         fmt.Sprintf("%s-sc", module.Name),
					Image:        eurekaSidecarImage,
					RegistryAuth: dto.EurekaRegistryAuthToken,
					Config: &container.Config{
						Image:        eurekaSidecarImage,
						Hostname:     fmt.Sprintf("%s-sc", strings.ToLower(moduleNameEnv)),
						Env:          sidecarEnv,
						ExposedPorts: *backendModule.SidecarExposedPorts,
					},
					HostConfig: &container.HostConfig{
						PortBindings:  *backendModule.SidecarPortBindings,
						RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyOnFailure, MaximumRetryCount: 3},
					},
					NetworkConfig: networkConfig,
					Platform:      &v1.Platform{},
					PullImage:     pullSidecarImage,
				}

				DeployModule(commandName, containerdCli, deploySidecarDto)

				pullSidecarImage = false
			}
		}
	}
}

func CreateTenants(commandName string, enableDebug bool) {
	tenants := viper.GetStringSlice(TenantConfigKey)

	getTenantsReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf(DockerInternalUrl, 9902, "/tenants"), nil)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
		panic(err)
	}

	if enableDebug {
		DumpHttpRequest(commandName, getTenantsReq)
	}

	getTenantsResp, err := http.DefaultClient.Do(getTenantsReq)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
		panic(err)
	}
	defer getTenantsResp.Body.Close()

	if enableDebug {
		DumpHttpResponse(commandName, getTenantsResp)
	}

	var foundTenantsMap map[string]interface{}

	err = json.NewDecoder(getTenantsResp.Body).Decode(&foundTenantsMap)
	if err != nil {
		slog.Error(commandName, SecondaryMessageKey, "json.NewDecoder error")
		panic(err)
	}

	foundTenants := foundTenantsMap["tenants"].([]interface{})

	for _, value := range foundTenants {
		mapEntry := value.(map[string]interface{})

		name := mapEntry["name"].(string)

		if slices.Contains(tenants, name) {
			id := mapEntry["id"].(string)

			delTenantReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(DockerInternalUrl, 9902, fmt.Sprintf("/tenants/%s", id)), nil)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
				panic(err)
			}

			if enableDebug {
				DumpHttpRequest(commandName, delTenantReq)
			}

			delTenantResp, err := http.DefaultClient.Do(delTenantReq)
			if err != nil {
				slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
				panic(err)
			}
			defer delTenantResp.Body.Close()

			if enableDebug {
				DumpHttpResponse(commandName, delTenantResp)
			}

			slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Removed tenant %s", name))

			break
		}
	}

	for _, tenant := range tenants {
		tenantBytes, err := json.Marshal(map[string]string{
			"name":        tenant,
			"description": "Default_tenant",
		})
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "json.Marshal error")
			panic(err)
		}

		if enableDebug {
			fmt.Println("### Dumping HTTP Request Body")
			fmt.Println(string(tenantBytes))
		}

		tenantReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf(DockerInternalUrl, 9902, "/tenants"), bytes.NewBuffer(tenantBytes))
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "http.NewRequest error")
			panic(err)
		}

		tenantReq.Header.Add("Content-Type", ApplicationJson)

		if enableDebug {
			DumpHttpRequest(commandName, tenantReq)
		}

		tenantResp, err := http.DefaultClient.Do(tenantReq)
		if err != nil {
			slog.Error(commandName, SecondaryMessageKey, "http.DefaultClient.Do error")
			panic(err)
		}
		defer tenantResp.Body.Close()

		if enableDebug {
			DumpHttpResponse(commandName, tenantResp)
		}

		slog.Info(commandName, SecondaryMessageKey, fmt.Sprintf("Created tenant %s", tenant))
	}
}
