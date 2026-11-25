package modulesvc

import (
	"maps"
	"slices"
	"strings"

	appErrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/moby/moby/client"
)

// ModuleLocalImageChecker defines the interface for checking images that only a local build provides
type ModuleLocalImageChecker interface {
	HasLocalOnlyImages(backendModules map[string]models.BackendModule) bool
	LocalOnlyImagesNeedRegistry(backendModules map[string]models.BackendModule) bool
	CheckLocalOnlyImages(client *client.Client, containers *models.Containers) error
}

func (ms *ModuleSvc) HasLocalOnlyImages(backendModules map[string]models.BackendModule) bool {
	return ms.isLocalOnlySidecarImage() || hasLocalOnlyModules(backendModules)
}

// LocalOnlyImagesNeedRegistry reports whether the registry is needed: it decides which local-descriptor modules
// are deployed, and it provides the version of an unpinned local-only sidecar
func (ms *ModuleSvc) LocalOnlyImagesNeedRegistry(backendModules map[string]models.BackendModule) bool {
	return ms.isLocalOnlySidecarImage() && !ms.isSidecarVersionPinned() || hasLocalOnlyModules(backendModules)
}

func (ms *ModuleSvc) CheckLocalOnlyImages(dockerClient *client.Client, containers *models.Containers) error {
	images, err := ms.getLocalOnlyImages(containers)
	if err != nil {
		return err
	}

	var missing []string
	for _, image := range images {
		exists, err := imageExists(dockerClient, image)
		if err != nil {
			return err
		}
		if !exists {
			missing = append(missing, image)
		}
	}
	if len(missing) > 0 {
		return appErrors.LocalImagesNotFound(missing)
	}

	return nil
}

func (ms *ModuleSvc) getLocalOnlyImages(containers *models.Containers) ([]string, error) {
	var images []string
	if ms.isLocalOnlySidecarImage() {
		sidecarImage, _, err := ms.GetSidecarImage(containers.Modules.EurekaModules)
		if err != nil {
			return nil, err
		}
		images = append(images, sidecarImage)
	}

	for _, moduleName := range slices.Sorted(maps.Keys(containers.BackendModules)) {
		backendModule := containers.BackendModules[moduleName]
		if !isLocalOnlyModule(backendModule) {
			continue
		}
		_, module := ms.GetBackendModule(containers, moduleName)
		if module == nil {
			continue
		}
		version := ms.GetModuleImageVersion(backendModule, module)
		images = append(images, ms.GetModuleImage(&models.ProxyModule{Metadata: models.ProxyModuleMetadata{Name: moduleName, Version: &version}}))
	}

	return images, nil
}

func hasLocalOnlyModules(backendModules map[string]models.BackendModule) bool {
	for _, backendModule := range backendModules {
		if isLocalOnlyModule(backendModule) {
			return true
		}
	}

	return false
}

func isLocalOnlyModule(backendModule models.BackendModule) bool {
	return backendModule.DeployModule && backendModule.LocalDescriptorPath != ""
}

// isLocalOnlySidecarImage reports a custom-namespace image without a namespace: a pull could only
// resolve it to docker.io/library, where no sidecar image is published
func (ms *ModuleSvc) isLocalOnlySidecarImage() bool {
	customNamespace := helpers.GetBool(ms.Action.ConfigSidecarModule, field.SidecarModuleCustomNamespaceEntry)
	image := helpers.GetString(ms.Action.ConfigSidecarModule, field.SidecarModuleImageEntry)

	return customNamespace && !strings.Contains(image, "/")
}

func (ms *ModuleSvc) isSidecarVersionPinned() bool {
	_, pinned := ms.Action.ConfigSidecarModule[field.SidecarModuleVersionEntry].(string)
	return pinned
}
