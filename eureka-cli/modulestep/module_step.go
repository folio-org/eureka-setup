package modulestep

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/registrystep"
	"github.com/spf13/viper"
)

type ModuleStep struct {
	Action       *action.Action
	HTTPClient   *httpclient.HTTPClient
	DockerClient *dockerclient.DockerClient
	RegistryStep *registrystep.RegistryStep
}

func New(
	action *action.Action,
	httpClient *httpclient.HTTPClient,
	dockerClient *dockerclient.DockerClient,
	registryStep *registrystep.RegistryStep,
) *ModuleStep {

	return &ModuleStep{
		Action:       action,
		HTTPClient:   httpClient,
		DockerClient: dockerClient,
		RegistryStep: registryStep,
	}
}

func (ms *ModuleStep) GetVaultRootToken(client *client.Client) string {
	logStream, err := client.ContainerLogs(context.Background(), constant.VaultContainerName, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = logStream.Close()
	}()

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		count := binary.BigEndian.Uint32(buffer[4:])
		rawLogLine := make([]byte, count)

		_, err = logStream.Read(rawLogLine)
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
		}

		parsedLogLine := string(rawLogLine)

		if strings.Contains(parsedLogLine, constant.VaultRootTokenPattern) {
			vaultRootToken := helpers.GetVaultRootTokenFromLogs(parsedLogLine)

			return vaultRootToken
		}
	}
}

func (ms *ModuleStep) PerformModuleHealthCheck(waitMutex *sync.WaitGroup, moduleName string, port int) {
	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Waiting for module container %s on port %d to initialize", moduleName, port))

	healthCheckWait := 10 * time.Second

	requestURL := fmt.Sprintf(ms.HTTPClient.GetGatewayURL(), port, "/admin/health")
	healthCheckAttempts := constant.HealthCheckMaxAttempts
	for {
		time.Sleep(healthCheckWait)

		if ms.checkContainerStatusCode(requestURL) {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Module container %s is healthy", moduleName))
			waitMutex.Done()
			break
		}

		healthCheckAttempts--

		if healthCheckAttempts == 0 {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Module container %s is unhealthy, out of attempts", moduleName))
			waitMutex.Done()

			helpers.LogErrorPanic(ms.Action, fmt.Errorf("module container %s did not initialize, cannot continue", moduleName))
			return
		}

		slog.Info(ms.Action.Name, "text", fmt.Sprintf("Module container %s is unhealthy, %d/%d attempts left", moduleName, healthCheckAttempts, constant.HealthCheckMaxAttempts))
	}
}

func (ms *ModuleStep) checkContainerStatusCode(requestURL string) bool {
	response := ms.HTTPClient.DoGetReturnResponse(requestURL, false, map[string]string{})
	if response == nil {
		return false
	}
	defer func() {
		_ = response.Body.Close()
	}()

	return response.StatusCode == http.StatusOK
}

func (ms *ModuleStep) DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) map[string]int {
	deployedModules := make(map[string]int)
	networkConfig := helpers.NewModuleNetworkConfig()

	for registryName, registryModules := range containers.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(ms.Action.Name, "text", fmt.Sprintf("Deploying %s modules", registryName))
		}

		for _, registryModule := range registryModules {
			managementModule := strings.Contains(registryModule.Name, constant.ManagementModulePattern)
			if (containers.ManagementOnly && !managementModule) || (!containers.ManagementOnly && managementModule) {
				continue
			}

			backendModule, ok := containers.BackendModulesMap[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			moduleVersion := ms.GetModuleImageVersion(backendModule, registryModule)
			moduleImage := ms.GetModuleImage(moduleVersion, registryModule)
			moduleEnvironment := ms.GetModuleEnvironment(containers, registryModule, backendModule)
			container := models.NewModuleContainer(registryModule.Name, moduleImage, moduleEnvironment, backendModule, networkConfig)

			ms.DeployModule(client, container)

			deployedModules[registryModule.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				go func() {
					sidecarEnvironment := ms.GetSidecarEnvironment(containers, registryModule, backendModule, nil, nil)
					sidecarContainer := models.NewSidecarContainer(registryModule.SidecarName, sidecarImage, sidecarEnvironment, backendModule, networkConfig, sidecarResources)

					ms.DeployModule(client, sidecarContainer)
				}()
			}
		}
	}

	return deployedModules
}

func (ms *ModuleStep) GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.RegistryModule) {
	for _, registryModules := range containers.RegistryModules {
		for _, registryModule := range registryModules {
			backendModule, ok := containers.BackendModulesMap[registryModule.Name]
			if !ok || !backendModule.DeployModule {
				continue
			}

			if registryModule.Name == moduleName {
				return &backendModule, registryModule
			}
		}
	}

	return nil, nil
}

func (ms *ModuleStep) GetModuleImageVersion(backendModule models.BackendModule, registryModule *models.RegistryModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *registryModule.Version
}

func (ms *ModuleStep) GetSidecarImage(registryModules []*models.RegistryModule) (string, bool) {
	sidecarModule := viper.GetStringMap(field.SidecarModule)
	sidecarImageVersion := ms.getSidecarImageVersion(registryModules, sidecarModule[field.SidecarModuleVersionEntry])

	localImage := sidecarModule[field.SidecarModuleLocalImageEntry]
	if localImage != nil && localImage.(string) != "" {
		return fmt.Sprintf("%s:%s", localImage.(string), sidecarImageVersion), false
	}

	namespace := ms.RegistryStep.GetNamespace(sidecarImageVersion)
	image := sidecarModule[field.SidecarModuleImageEntry]
	return fmt.Sprintf("%s/%s", namespace, fmt.Sprintf("%s:%s", image.(string), sidecarImageVersion)), true
}

func (ms *ModuleStep) getSidecarImageVersion(registryModules []*models.RegistryModule, sidecarConfigVersion any) string {
	ok, sidecarRegistryVersion := ms.findSidecarRegistryVersion(registryModules)
	if !ok && sidecarConfigVersion == nil {
		helpers.LogErrorPanic(ms.Action, fmt.Errorf("%s sidecar version is not found in registry or in the current config", sidecarConfigVersion))
		return ""
	}

	if sidecarConfigVersion != nil {
		return sidecarConfigVersion.(string)
	}

	return sidecarRegistryVersion
}

func (ms *ModuleStep) findSidecarRegistryVersion(registryModules []*models.RegistryModule) (bool, string) {
	for _, registryModule := range registryModules {
		if registryModule.Name == constant.SidecarProjectName {
			return true, *registryModule.Version
		}
	}

	return false, ""
}

func (ms *ModuleStep) GetModuleImage(moduleVersion string, registryModule *models.RegistryModule) string {
	return fmt.Sprintf("%s/%s:%s", ms.RegistryStep.GetNamespace(moduleVersion), registryModule.Name, moduleVersion)
}

func (ms *ModuleStep) GetModuleEnvironment(myContainer *models.Containers, module *models.RegistryModule, backendModule models.BackendModule) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, myContainer.GlobalEnvironment...)

	if backendModule.UseVault {
		combinedEnvironment = moduleenv.AppendVaultEnv(combinedEnvironment, myContainer.VaultRootToken)
	}
	if backendModule.UseOkapiURL {
		combinedEnvironment = moduleenv.AppendOkapiEnv(combinedEnvironment, module.SidecarName, backendModule.ModuleServerPort)
	}
	if backendModule.DisableSystemUser {
		combinedEnvironment = moduleenv.AppendDisableSystemUserEnv(combinedEnvironment, module.Name)
	}
	combinedEnvironment = moduleenv.AppendModuleEnvironment(combinedEnvironment, backendModule.ModuleEnvironment)

	return combinedEnvironment
}

func (ms *ModuleStep) GetSidecarEnvironment(containers *models.Containers, module *models.RegistryModule, backendModule models.BackendModule, moduleURL *string, sidecarURL *string) []string {
	var combinedEnvironment []string
	combinedEnvironment = append(combinedEnvironment, containers.SidecarEnvironment...)
	combinedEnvironment = moduleenv.AppendVaultEnv(combinedEnvironment, containers.VaultRootToken)
	combinedEnvironment = moduleenv.AppendKeycloakEnv(combinedEnvironment)
	combinedEnvironment = moduleenv.AppendSidecarEnvironment(combinedEnvironment, module, backendModule.ModuleServerPort, moduleURL, sidecarURL)

	return combinedEnvironment
}

func (ms *ModuleStep) DeployModule(client *client.Client, myContainer *models.Container) {
	containerName := ms.getContainerName(myContainer)

	if myContainer.PullImage {
		ms.PullModule(client, myContainer.Image)
	}

	cr, err := client.ContainerCreate(context.Background(), myContainer.Config, myContainer.HostConfig, myContainer.NetworkConfig, myContainer.Platform, containerName)
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	if len(cr.Warnings) > 0 {
		slog.Warn(ms.Action.Name, "text", fmt.Sprintf("container creation warnings: %s", cr.Warnings))
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Deployed module container, name: %s", containerName))
}

func (ms *ModuleStep) getContainerName(myContainer *models.Container) string {
	if strings.HasPrefix(myContainer.Name, constant.ManagementModulePattern) {
		return fmt.Sprintf("eureka-%s", myContainer.Name)
	}

	return fmt.Sprintf("eureka-%s-%s", viper.GetString(field.ProfileName), myContainer.Name)
}

func (ms *ModuleStep) PullModule(client *client.Client, imageName string) {
	registryAuthToken := ms.RegistryStep.GetAuthTokenIfPresent()

	reader, err := client.ImagePull(context.Background(), imageName, image.PullOptions{RegistryAuth: registryAuthToken})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}
	defer func() {
		_ = reader.Close()
	}()

	decoder := json.NewDecoder(reader)

	var event *models.Event
	for {
		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}

		if event.Error != "" {
			helpers.LogErrorPanic(ms.Action, fmt.Errorf("pulling module container image %+v", event.Error))
			return
		}
	}
}

func (ms *ModuleStep) GetDeployedModules(client *client.Client, filters filters.Args) []container.Summary {
	deployedModules, err := client.ContainerList(context.Background(), container.ListOptions{All: true, Filters: filters})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	return deployedModules
}

func (ms *ModuleStep) UndeployModuleByNamePattern(client *client.Client, value string, removeAsync bool) {
	deployedModules := ms.GetDeployedModules(client, filters.NewArgs(filters.KeyValuePair{Key: "name", Value: value}))
	for _, deployedModule := range deployedModules {
		ms.undeployModule(client, deployedModule, removeAsync)
	}
}

func (ms *ModuleStep) undeployModule(client *client.Client, deployedModule container.Summary, removeAsync bool) {
	err := client.NetworkDisconnect(context.Background(), constant.DefaultNetworkID, deployedModule.ID, false)
	if err != nil {
		slog.Warn(ms.Action.Name, "text", fmt.Sprintf("network disconnected: %s", err.Error()))
	}

	err = client.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{Signal: "9"})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err)
		panic(err)
	}

	var callback = func() {
		err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{Force: true, RemoveVolumes: true})
		if err != nil {
			slog.Error(ms.Action.Name, "error", err)
			panic(err)
		}
	}

	if removeAsync {
		go callback()
	} else {
		callback()
	}

	containerName := strings.ReplaceAll(deployedModule.Names[0], "/", "")

	slog.Info(ms.Action.Name, "text", fmt.Sprintf("Undeployed module container, name: %s", containerName))
}
