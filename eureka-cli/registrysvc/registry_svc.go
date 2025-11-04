package registrysvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	GetModules(installJsonURLs map[string]string, printModules bool) (map[string][]*models.RegistryModule, error)
	ExtractModuleNameAndVersion(registryModules map[string][]*models.RegistryModule)
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

func (rs *RegistrySvc) ExtractModuleNameAndVersion(registryModules map[string][]*models.RegistryModule) {
	for _, modules := range registryModules {
		for moduleIndex, module := range modules {
			if module.ID == "okapi" {
				continue
			}

			module.Name = helpers.GetModuleNameFromID(module.ID)
			module.Version = helpers.GetModuleVersionPFromID(module.ID)
			if strings.HasPrefix(module.Name, "edge") {
				module.SidecarName = module.Name
			} else {
				module.SidecarName = fmt.Sprintf("%s-sc", module.Name)
			}
			modules[moduleIndex] = module
		}
	}
}

func (rs *RegistrySvc) GetModules(installJsonURLs map[string]string, printModules bool) (map[string][]*models.RegistryModule, error) {
	registryModules := make(map[string][]*models.RegistryModule)
	for registryName, installJsonURL := range installJsonURLs {
		installJsonResp, err := rs.HTTPClient.GetReturnResponse(installJsonURL, map[string]string{})
		if err != nil {
			return nil, err
		}
		defer helpers.CloseReader(installJsonResp.Body)

		var decodedRegistryModules []*models.RegistryModule
		err = json.NewDecoder(installJsonResp.Body).Decode(&decodedRegistryModules)
		if err != nil && !errors.Is(err, io.EOF) {
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

				registryModule := &models.RegistryModule{
					ID:     fmt.Sprintf("%s-%s", name, entry[field.ModuleVersionEntry].(string)),
					Action: "enable",
				}
				decodedRegistryModules = append(decodedRegistryModules, registryModule)
			}
		}

		if len(decodedRegistryModules) > 0 {
			if printModules {
				slog.Info(rs.Action.Name, "text", "Read registry with modules", "registry", registryName, "moduleCount", len(decodedRegistryModules))
			}
			sort.Slice(decodedRegistryModules, func(i, j int) bool {
				switch strings.Compare(decodedRegistryModules[i].ID, decodedRegistryModules[j].ID) {
				case -1:
					return true
				case 1:
					return false
				}

				return decodedRegistryModules[i].ID > decodedRegistryModules[j].ID
			})
		}
		registryModules[registryName] = decodedRegistryModules
	}

	return registryModules, nil
}
