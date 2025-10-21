package modulesvc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/registrysvc"
	"github.com/spf13/viper"
)

type ModuleSvc struct {
	Action       *action.Action
	HTTPClient   *httpclient.HTTPClient
	DockerClient *dockerclient.DockerClient
	RegistrySvc  *registrysvc.RegistrySvc
}

type SidecarRequest struct {
	Client           *client.Client
	Containers       *models.Containers
	RegistryModule   *models.RegistryModule
	BackendModule    models.BackendModule
	SidecarImage     string
	NetworkConfig    *network.NetworkingConfig
	SidecarResources *container.Resources
}

func New(
	action *action.Action,
	httpClient *httpclient.HTTPClient,
	dockerClient *dockerclient.DockerClient,
	registrySvc *registrysvc.RegistrySvc,
) *ModuleSvc {

	return &ModuleSvc{
		Action:       action,
		HTTPClient:   httpClient,
		DockerClient: dockerClient,
		RegistrySvc:  registrySvc,
	}
}

func (ms *ModuleSvc) GetVaultRootToken(client *client.Client) (string, error) {
	logStream, err := client.ContainerLogs(context.Background(), constant.VaultContainer, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return "", err
	}
	defer helpers.CloseReader(logStream)

	buffer := make([]byte, 8)
	for {
		_, err := logStream.Read(buffer)
		if err != nil {
			return "", err
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

			return vaultRootToken, nil
		}
	}
}

func (ms *ModuleSvc) PerformModuleReadinessCheck(wg *sync.WaitGroup, errCh chan<- error, moduleName string, port int) {
	defer wg.Done()

	slog.Info(ms.Action.Name, "text", "Waiting module on port", "module", moduleName, "port", port)

	requestURL := ms.Action.CreateURL(strconv.Itoa(port), "/admin/health")
	retries := constant.ModuleReadinessMaxRetries
	for {
		time.Sleep(constant.ModuleReadinessCheckWait)

		ready, err := ms.checkContainerStatusCode(requestURL)
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			slog.Debug(ms.Action.Name, "text", err)
		}

		if ready {
			slog.Info(ms.Action.Name, "text", "Module is ready", "module", moduleName)
			return
		}

		retries--

		if retries == 0 {
			err := fmt.Errorf("module %s is unready and out of retries", moduleName)
			slog.Error(ms.Action.Name, "error", err)
			select {
			case errCh <- err:
			default:
			}
			return
		}

		slog.Info(ms.Action.Name, "text", "Module is unready, retries remaining", "module", moduleName, "retries", retries, "maxRetries", constant.ModuleReadinessMaxRetries)
	}
}

func (ms *ModuleSvc) checkContainerStatusCode(requestURL string) (bool, error) {
	resp, err := ms.HTTPClient.GetReturnResponse(requestURL, map[string]string{})
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, nil
	}
	defer httpclient.CloseResponse(resp)

	return resp.StatusCode == http.StatusOK, nil
}

func (ms *ModuleSvc) DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, error) {
	deployedModules := make(map[string]int)
	networkConfig := helpers.NewModuleNetworkConfig()

	var sidecarWG sync.WaitGroup
	sidecarErrCh := make(chan error, 10)

	for registryName, registryModules := range containers.RegistryModules {
		if len(registryModules) > 0 {
			slog.Info(ms.Action.Name, "text", "Deploying modules", "registry", registryName)
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
			moduleEnv := ms.GetModuleEnv(containers, registryModule, backendModule)
			container := models.NewModuleContainer(registryModule.Name, moduleImage, moduleEnv, backendModule, networkConfig)

			err := ms.DeployModule(client, container)
			if err != nil {
				return nil, err
			}

			deployedModules[registryModule.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				sidecarWG.Add(1)

				go ms.deploySidecarAsync(&sidecarWG, sidecarErrCh, &SidecarRequest{
					Client:           client,
					Containers:       containers,
					RegistryModule:   registryModule,
					BackendModule:    backendModule,
					SidecarImage:     sidecarImage,
					NetworkConfig:    networkConfig,
					SidecarResources: sidecarResources,
				})
			}
		}
	}

	go func() {
		sidecarWG.Wait()
		close(sidecarErrCh)
	}()

	for err := range sidecarErrCh {
		return nil, err
	}

	return deployedModules, nil
}

func (ms *ModuleSvc) deploySidecarAsync(wg *sync.WaitGroup, errCh chan<- error, req *SidecarRequest) {
	defer wg.Done()

	sidecarEnv := ms.GetSidecarEnv(req.Containers, req.RegistryModule, req.BackendModule, nil, nil)
	sidecarContainer := models.NewSidecarContainer(req.RegistryModule.SidecarName, req.SidecarImage, sidecarEnv, req.BackendModule, req.NetworkConfig, req.SidecarResources)

	err := ms.DeployModule(req.Client, sidecarContainer)
	if err != nil {
		err := fmt.Errorf("failed to deploy %s sidecar with error %w", req.RegistryModule.SidecarName, err)
		slog.Error(ms.Action.Name, "error", err)

		select {
		case errCh <- err:
		default:
		}
	}
}

func (ms *ModuleSvc) GetBackendModule(containers *models.Containers, moduleName string) (*models.BackendModule, *models.RegistryModule) {
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

func (ms *ModuleSvc) GetModuleImageVersion(backendModule models.BackendModule, registryModule *models.RegistryModule) string {
	if backendModule.ModuleVersion != nil {
		return *backendModule.ModuleVersion
	}

	return *registryModule.Version
}

func (ms *ModuleSvc) GetSidecarImage(registryModules []*models.RegistryModule) (string, bool, error) {
	sidecarModule := viper.GetStringMap(field.SidecarModule)
	sidecarImageVersion, err := ms.getSidecarImageVersion(registryModules, sidecarModule[field.SidecarModuleVersionEntry])
	if err != nil {
		return "", false, err
	}

	localImage := sidecarModule[field.SidecarModuleLocalImageEntry]
	if localImage != nil && localImage.(string) != "" {
		return fmt.Sprintf("%s:%s", localImage.(string), sidecarImageVersion), false, nil
	}

	namespace := ms.RegistrySvc.GetNamespace(sidecarImageVersion)
	image := sidecarModule[field.SidecarModuleImageEntry]

	return fmt.Sprintf("%s/%s", namespace, fmt.Sprintf("%s:%s", image.(string), sidecarImageVersion)), true, nil
}

func (ms *ModuleSvc) getSidecarImageVersion(registryModules []*models.RegistryModule, sidecarConfigVersion any) (string, error) {
	ok, sidecarRegistryVersion := ms.findSidecarRegistryVersion(registryModules)
	if !ok && sidecarConfigVersion == nil {
		return "", fmt.Errorf("%s sidecar version is not found in registry or in the current config", sidecarConfigVersion)
	}

	if sidecarConfigVersion != nil {
		return sidecarConfigVersion.(string), nil
	}

	return sidecarRegistryVersion, nil
}

func (ms *ModuleSvc) findSidecarRegistryVersion(registryModules []*models.RegistryModule) (bool, string) {
	for _, registryModule := range registryModules {
		if registryModule.Name == constant.SidecarProjectName {
			return true, *registryModule.Version
		}
	}

	return false, ""
}

func (ms *ModuleSvc) GetModuleImage(moduleVersion string, registryModule *models.RegistryModule) string {
	return fmt.Sprintf("%s/%s:%s", ms.RegistrySvc.GetNamespace(moduleVersion), registryModule.Name, moduleVersion)
}

func (ms *ModuleSvc) GetModuleEnv(myContainer *models.Containers, module *models.RegistryModule, backendModule models.BackendModule) []string {
	var combinedEnv []string
	combinedEnv = append(combinedEnv, myContainer.GlobalEnv...)

	if backendModule.UseVault {
		combinedEnv = moduleenv.AppendVaultEnv(combinedEnv, myContainer.VaultRootToken)
	}
	if backendModule.UseOkapiURL {
		combinedEnv = moduleenv.AppendOkapiEnv(combinedEnv, module.SidecarName, backendModule.ModuleServerPort)
	}
	if backendModule.DisableSystemUser {
		combinedEnv = moduleenv.AppendDisableSystemUserEnv(combinedEnv, module.Name)
	}
	combinedEnv = moduleenv.AppendModuleEnv(combinedEnv, backendModule.ModuleEnv)

	return combinedEnv
}

func (ms *ModuleSvc) GetSidecarEnv(containers *models.Containers, module *models.RegistryModule, backendModule models.BackendModule, moduleURL *string, sidecarURL *string) []string {
	var combinedEnv []string
	combinedEnv = append(combinedEnv, containers.SidecarEnv...)
	combinedEnv = moduleenv.AppendVaultEnv(combinedEnv, containers.VaultRootToken)
	combinedEnv = moduleenv.AppendKeycloakEnv(combinedEnv)
	combinedEnv = moduleenv.AppendSidecarEnv(combinedEnv, module, backendModule.ModuleServerPort, moduleURL, sidecarURL)

	return combinedEnv
}

func (ms *ModuleSvc) DeployModule(client *client.Client, myContainer *models.Container) error {
	containerName := ms.getContainerName(myContainer)

	if myContainer.PullImage {
		err := ms.PullModule(client, myContainer.Image)
		if err != nil {
			return err
		}
	}

	cr, err := client.ContainerCreate(context.Background(), myContainer.Config, myContainer.HostConfig, myContainer.NetworkConfig, myContainer.Platform, containerName)
	if err != nil {
		return err
	}

	if len(cr.Warnings) > 0 {
		slog.Warn(ms.Action.Name, "text", "Caught module creation with warning", "container", containerName, "warnings", cr.Warnings)
	}

	err = client.ContainerStart(context.Background(), cr.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	slog.Info(ms.Action.Name, "text", "Deployed module", "module", containerName)

	return nil
}

func (ms *ModuleSvc) getContainerName(myContainer *models.Container) string {
	if strings.HasPrefix(myContainer.Name, constant.ManagementModulePattern) {
		return fmt.Sprintf("eureka-%s", myContainer.Name)
	}

	return fmt.Sprintf("eureka-%s-%s", viper.GetString(field.ProfileName), myContainer.Name)
}

func (ms *ModuleSvc) PullModule(client *client.Client, imageName string) error {
	registryAuthToken, err := ms.RegistrySvc.GetAuthTokenIfPresent()
	if err != nil {
		return err
	}

	reader, err := client.ImagePull(context.Background(), imageName, image.PullOptions{
		RegistryAuth: registryAuthToken,
	})
	if err != nil {
		return err
	}
	defer helpers.CloseReader(reader)

	decoder := json.NewDecoder(reader)

	var event *models.Event
	for {
		if err := decoder.Decode(&event); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if event.Error != "" {
			return fmt.Errorf("pulling module image with error %+v", event.Error)
		}
	}

	return nil
}

func (ms *ModuleSvc) GetDeployedModules(client *client.Client, filters filters.Args) ([]container.Summary, error) {
	deployedModules, err := client.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	return deployedModules, nil
}

func (ms *ModuleSvc) UndeployModuleByNamePattern(client *client.Client, pattern string) error {
	deployedModules, err := ms.GetDeployedModules(client, filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: pattern,
	}))
	if err != nil {
		return err
	}

	for _, deployedModule := range deployedModules {
		err = ms.undeployModule(client, deployedModule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *ModuleSvc) undeployModule(client *client.Client, deployedModule container.Summary) error {
	err := client.NetworkDisconnect(context.Background(), constant.NetworkID, deployedModule.ID, false)
	if err != nil {
		slog.Warn(ms.Action.Name, "text", "Module network is disconnected with warnings", "moduleId", deployedModule.ID, "error", err.Error())
	}

	err = client.ContainerStop(context.Background(), deployedModule.ID, container.StopOptions{
		Signal: "9",
	})
	if err != nil {
		return err
	}

	err = client.ContainerRemove(context.Background(), deployedModule.ID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	if err != nil {
		slog.Error(ms.Action.Name, "error", err, "module", deployedModule.ID, "operation", "container remove")
	}

	containerName := strings.ReplaceAll(deployedModule.Names[0], "/", "")
	slog.Info(ms.Action.Name, "text", "Undeployed module", "module", containerName)

	return nil
}
