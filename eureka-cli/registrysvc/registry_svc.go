package registrysvc

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/awssvc"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
)

// RegistryProcessor defines the interface for registry-related operations
type RegistryProcessor interface {
	GetNamespace(version string) string
	GetModules(installJsonURLs map[string]string, verbose bool) (*models.ProxyModulesByRegistry, error)
	ExtractModuleMetadata(modules *models.ProxyModulesByRegistry)
	GetAuthorizationToken() (string, error)
}

// RegistrySvc provides functionality for interacting with module registries
type RegistrySvc struct {
	Action     *action.Action
	HTTPClient httpclient.HTTPClientRunner
	AWSSvc     awssvc.AWSProcessor
}

// New creates a new RegistrySvc instance
func New(action *action.Action, httpClient httpclient.HTTPClientRunner, awsSvc awssvc.AWSProcessor) *RegistrySvc {
	return &RegistrySvc{Action: action, HTTPClient: httpClient, AWSSvc: awsSvc}
}

func (rs *RegistrySvc) GetAuthorizationToken() (string, error) {
	return rs.AWSSvc.GetAuthorizationToken()
}

func (rs *RegistrySvc) GetNamespace(version string) string {
	ecrNamespace := rs.AWSSvc.GetECRNamespace()
	if ecrNamespace != "" {
		return ecrNamespace
	}

	if strings.Contains(version, "SNAPSHOT") {
		return constant.SnapshotRegistry
	} else {
		return constant.ReleaseRegistry
	}
}

func (rs *RegistrySvc) GetModules(installJsonURLs map[string]string, verbose bool) (*models.ProxyModulesByRegistry, error) {
	var folioModules, eurekaModules []*models.ProxyModule
	for registryName, installJsonURL := range installJsonURLs {
		var decodedResponse []*models.ProxyModule
		if err := rs.HTTPClient.GetRetryReturnStruct(installJsonURL, map[string]string{}, &decodedResponse); err != nil {
			return nil, err
		}

		if registryName == constant.FolioRegistry {
			for name, value := range rs.Action.ConfigCustomFrontendModules {
				if value == nil {
					continue
				}

				entry := value.(map[string]any)
				if entry[field.ModuleVersionEntry] == nil {
					continue
				}

				module := &models.ProxyModule{
					ID:     fmt.Sprintf("%s-%s", name, entry[field.ModuleVersionEntry].(string)),
					Action: "enable",
				}
				decodedResponse = append(decodedResponse, module)
			}
		}

		if len(decodedResponse) > 0 {
			if verbose {
				slog.Info(rs.Action.Name, "text", "Read registry", "registry", registryName, "count", len(decodedResponse))
			}
			sort.Slice(decodedResponse, func(i, j int) bool {
				switch strings.Compare(decodedResponse[i].ID, decodedResponse[j].ID) {
				case -1:
					return true
				case 1:
					return false
				}

				return decodedResponse[i].ID > decodedResponse[j].ID
			})
		}

		switch registryName {
		case constant.FolioRegistry:
			folioModules = decodedResponse
		case constant.EurekaRegistry:
			eurekaModules = decodedResponse
		}
	}

	return models.NewProxyModulesByRegistry(folioModules, eurekaModules), nil
}

func (rs *RegistrySvc) ExtractModuleMetadata(modules *models.ProxyModulesByRegistry) {
	allModules := [][]*models.ProxyModule{modules.FolioModules, modules.EurekaModules}
	for _, moduleByRegistry := range allModules {
		for moduleIndex, module := range moduleByRegistry {
			if module.ID == "okapi" {
				continue
			}

			module.Name = helpers.GetModuleNameFromID(module.ID)
			module.Version = helpers.GetOptionalModuleVersion(module.ID)
			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = helpers.GetSidecarName(module.Name)
			}
			moduleByRegistry[moduleIndex] = module
		}
	}
}
