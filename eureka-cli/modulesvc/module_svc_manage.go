package modulesvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
)

// ModuleManager defines the interface for managing module deployment and lifecycle
type ModuleManager interface {
	GetDeployedModules(client *client.Client, filters filters.Args) ([]container.Summary, error)
	GetModule(client *client.Client, moduleName string) ([]container.Summary, error)
	PullModule(client *client.Client, imageName string) error
	DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, int, error)
	DeployModule(client *client.Client, container *models.Container) error
	UndeployModuleByNamePattern(client *client.Client, pattern string) error
}

func (ms *ModuleSvc) GetDeployedModules(client *client.Client, filters filters.Args) ([]container.Summary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerList)
	defer cancel()

	deployedModules, err := client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	return deployedModules, nil
}

func (ms *ModuleSvc) GetModule(client *client.Client, moduleName string) ([]container.Summary, error) {
	containerName := fmt.Sprintf("eureka-%s-%s", ms.Action.ConfigProfileName, moduleName)
	if strings.HasPrefix(moduleName, constant.ManagementModulePattern) {
		containerName = fmt.Sprintf("eureka-%s", moduleName)
	}

	return ms.GetDeployedModules(client, filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: fmt.Sprintf("^%s$", containerName),
	}))
}

func (ms *ModuleSvc) PullModule(client *client.Client, imageName string) error {
	_, err := client.ImageInspect(context.Background(), imageName)
	if err == nil {
		slog.Info(ms.Action.Name, "text", "Image already exists locally", "image", imageName)
		return nil
	}
	if !errdefs.IsNotFound(err) {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerImagePull)
	defer cancel()

	authorizationToken, err := ms.RegistrySvc.GetAuthorizationToken()
	if err != nil {
		return err
	}

	reader, err := client.ImagePull(ctx, imageName, image.PullOptions{
		RegistryAuth: authorizationToken,
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
		if event.Error == "" {
			current := helpers.ConvertMemory(helpers.BytesToMib, int64(event.ProgressDetail.Current))
			total := helpers.ConvertMemory(helpers.BytesToMib, int64(event.ProgressDetail.Total))
			slog.Debug(ms.Action.Name, "text", "Pulling module", "imageName", imageName, "status", event.Status, "progressCurrent", current, "progressTotal", total)
		} else {
			return appErrors.ModulePullFailed(imageName, errors.New(event.Error))
		}
	}

	return nil
}

func (ms *ModuleSvc) DeployModules(client *client.Client, containers *models.Containers, sidecarImage string, sidecarResources *container.Resources) (map[string]int, int, error) {
	newlyDeployed := make(map[string]int)
	totalMatched := 0

	var sidecarWG sync.WaitGroup
	sidecarErrCh := make(chan error, 10)
	allModules := [][]*models.ProxyModule{containers.Modules.FolioModules, containers.Modules.EurekaModules}
	for _, modules := range allModules {
		for _, module := range modules {
			if ms.shouldSkipModule(module, containers.IsManagement) {
				continue
			}
			if !ms.shouldDeployModule(module, containers.BackendModules) {
				continue
			}

			backendModule := containers.BackendModules[module.Metadata.Name]
			totalMatched++

			existingContainers, err := ms.GetModule(client, module.Metadata.Name)
			if err != nil {
				return nil, 0, err
			}
			if len(existingContainers) > 0 {
				var portParts []string
				for _, p := range existingContainers[0].Ports {
					portParts = append(portParts, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
				}
				slog.Info(ms.Action.Name, "text", "Container already deployed, skipping", "module", module.Metadata.Name, "ports", strings.Join(portParts, ", "))
				continue
			}

			version := ms.GetModuleImageVersion(backendModule, module)
			module.Metadata.Version = &version

			// --- AUTOMATED IMAGE & TIMESTAMPS SANITIZATION ENGINE ---
			imageName := ms.GetModuleImage(module)
			pullImage := backendModule.LocalDescriptorPath == ""

			slog.Debug(ms.Action.Name, "text", "Sanitization Engine trace init", "module", module.Metadata.Name, "framework_regex_resolved_version", version)

			if rawModuleConfig, exists := ms.Action.ConfigBackendModules[module.Metadata.Name]; exists {
				if moduleMap, ok := rawModuleConfig.(map[string]any); ok {
					if img, imgExists := moduleMap["image"]; imgExists {
						if imgStr, isStr := img.(string); isStr {
							customImage := strings.TrimSpace(imgStr)
							if customImage != "" {
								if strings.Contains(customImage, ":") {
									imageName = customImage
									slog.Debug(ms.Action.Name, "text", "Sanitization Engine: using exact static custom image reference", "module", module.Metadata.Name, "imageName", imageName)
								} else {
									tagSource := version
									if tagSource == "" && module.Metadata.Version != nil {
										tagSource = *module.Metadata.Version
										slog.Debug(ms.Action.Name, "text", "Sanitization Engine: framework regex version failed; recovered version from injection metadata", "module", module.Metadata.Name, "tagSource", tagSource)
									}

									cleanTag := strings.Split(tagSource, "+")[0]
									repoPath := customImage

									if idx := strings.Index(repoPath, "/"); idx != -1 {
										firstPart := repoPath[:idx]
										if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") {
											repoPath = repoPath[idx+1:]
										}
									}

									var paths []string
									if strings.Contains(repoPath, "/") {
										paths = []string{repoPath}
									} else {
										paths = []string{repoPath}
										for _, ns := range constant.GetNamespaces() {
											paths = append(paths, ns+"/"+repoPath)
										}
									}

									registries := ms.Action.ConfigDockerRegistries
									if len(registries) == 0 {
										registries = []string{"docker.io"}
									}

									chosenRegistry := registries[0]
									chosenPath := paths[0]
									found := false

									for _, path := range paths {
										for _, registry := range registries {
											var testURL string
											if registry == "docker.io" || registry == "registry-1.docker.io" {
												if !strings.Contains(path, "/") {
													continue
												}
												testURL = fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", path, cleanTag)
											} else {
												testURL = fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, path, cleanTag)
											}

											req, _ := http.NewRequest(http.MethodHead, testURL, nil)
											req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

											ctxClient := &http.Client{Timeout: 3 * time.Second}
											resp, err := ctxClient.Do(req)
											if err == nil {
												status := resp.StatusCode
												resp.Body.Close()
												if status == http.StatusOK {
													chosenRegistry = registry
													chosenPath = path
													found = true
													break
												}
											}
										}
										if found {
											break
										}
									}

									if chosenRegistry == "docker.io" || chosenRegistry == "registry-1.docker.io" {
										imageName = fmt.Sprintf("%s:%s", chosenPath, cleanTag)
									} else {
										imageName = fmt.Sprintf("%s/%s:%s", chosenRegistry, chosenPath, cleanTag)
									}
									slog.Debug(ms.Action.Name, "text", "Sanitization Engine output resolved details", "module", module.Metadata.Name, "selected_registry", chosenRegistry, "final_image_string", imageName)
								}
								pullImage = true
							}
						}
					}
				}
			}

			slog.Info(ms.Action.Name, "text", "Deploying module", "module", module.Metadata.Name,
				"port1", backendModule.ModuleExposedServerPort,
				"port2", backendModule.ModuleExposedDebugPort,
				"port3", backendModule.SidecarExposedServerPort,
				"port4", backendModule.SidecarExposedDebugPort)

			if err := ms.DeployModule(client, &models.Container{
				Name: module.Metadata.Name,
				Config: &container.Config{
					Image:        imageName,
					Hostname:     module.Metadata.Name,
					Env:          ms.GetModuleEnv(containers, module, backendModule),
					ExposedPorts: *backendModule.ModuleExposedPorts,
				},
				HostConfig: &container.HostConfig{
					PortBindings:  *backendModule.ModulePortBindings,
					RestartPolicy: *helpers.GetRestartPolicy(),
					Resources:     backendModule.ModuleResources,
					Binds:         backendModule.ModuleVolumes,
				},
				NetworkConfig: helpers.GetModuleNetworkConfig(),
				Platform:      helpers.GetPlatform(),
				PullImage:     pullImage,
			}); err != nil {
				return nil, 0, err
			}
			newlyDeployed[module.Metadata.Name] = backendModule.ModuleExposedServerPort

			if backendModule.DeploySidecar && sidecarImage != "" {
				sidecarWG.Add(1)
				go ms.deploySidecarAsync(&sidecarWG, sidecarErrCh, &models.SidecarRequest{
					Client:           client,
					Containers:       containers,
					Module:           module,
					BackendModule:    backendModule,
					SidecarImage:     sidecarImage,
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
		return nil, 0, err
	}

	return newlyDeployed, totalMatched, nil
}

func (ms *ModuleSvc) shouldSkipModule(module *models.ProxyModule, managementOnly bool) bool {
	isManagementModule := strings.Contains(module.Metadata.Name, constant.ManagementModulePattern)
	return (managementOnly && !isManagementModule) || (!managementOnly && isManagementModule)
}

func (ms *ModuleSvc) shouldDeployModule(module *models.ProxyModule, backendModules map[string]models.BackendModule) bool {
	backendModule, exists := backendModules[module.Metadata.Name]
	return exists && backendModule.DeployModule
}

func (ms *ModuleSvc) deploySidecarAsync(wg *sync.WaitGroup, errCh chan<- error, r *models.SidecarRequest) {
	defer wg.Done()

	container := &models.Container{
		Name: r.Module.Metadata.SidecarName,
		Config: &container.Config{
			Image:        r.SidecarImage,
			Hostname:     r.Module.Metadata.SidecarName,
			Env:          ms.GetSidecarEnv(r.Containers, r.Module, r.BackendModule, "", ""),
			ExposedPorts: *r.BackendModule.SidecarExposedPorts,
			Cmd:          helpers.GetConfigSidecarCmd(ms.Action.ConfigSidecarModuleNativeBinaryCmd),
		},
		HostConfig: &container.HostConfig{
			PortBindings:  *r.BackendModule.SidecarPortBindings,
			RestartPolicy: *helpers.GetRestartPolicy(),
			Resources:     *r.SidecarResources,
		},
		NetworkConfig: helpers.GetModuleNetworkConfig(),
		Platform:      helpers.GetPlatform(),
		PullImage:     false,
	}
	if err := ms.DeployModule(r.Client, container); err != nil {
		err := appErrors.SidecarDeployFailed(r.Module.Metadata.SidecarName, err)
		select {
		case errCh <- err:
		default:
		}
	}
}

func (ms *ModuleSvc) DeployModule(client *client.Client, c *models.Container) error {
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerDeploy)
	defer cancel()

	if c.PullImage {
		err := ms.PullModule(client, c.Config.Image)
		if err != nil {
			return err
		}
	}

	containerName := ms.getContainerName(c)
	createResponse, err := client.ContainerCreate(ctx, c.Config, c.HostConfig, c.NetworkConfig, c.Platform, containerName)
	if err != nil {
		return err
	}
	if len(createResponse.Warnings) > 0 {
		slog.Warn(ms.Action.Name, "text", "Module created with warning", "container", containerName, "warnings", createResponse.Warnings)
	}

	err = client.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	slog.Info(ms.Action.Name, "text", "Deployed module", "module", containerName)

	return nil
}

func (ms *ModuleSvc) getContainerName(container *models.Container) string {
	if strings.HasPrefix(container.Name, constant.ManagementModulePattern) {
		return fmt.Sprintf("eureka-%s", container.Name)
	}

	return fmt.Sprintf("eureka-%s-%s", ms.Action.ConfigProfileName, container.Name)
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
	ctx, cancel := context.WithTimeout(context.Background(), constant.ContextTimeoutDockerUndeploy)
	defer cancel()

	err := client.NetworkDisconnect(ctx, constant.NetworkID, deployedModule.ID, false)
	if err != nil {
		slog.Warn(ms.Action.Name, "text", "Module network is disconnected with warnings", "moduleId", deployedModule.ID, "error", err.Error())
	}

	err = client.ContainerStop(ctx, deployedModule.ID, container.StopOptions{
		Signal: "9",
	})
	if err != nil {
		return err
	}

	err = client.ContainerRemove(ctx, deployedModule.ID, container.RemoveOptions{
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