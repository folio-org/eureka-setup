package registrysvc

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/awssvc"
	"github.com/folio-org/eureka-cli/constant"
	appErrors "github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/httpclient"
	"github.com/folio-org/eureka-cli/models"
)

// RegistryProcessor defines the interface for registry-related operations
type RegistryProcessor interface {
	GetNamespace(version string) string
	// TODO Add tests with useRemote false
	GetModules(installJsonURLs map[string]string, useRemote, verbose bool) (*models.ProxyModulesByRegistry, error)
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

func (rs *RegistrySvc) GetModules(installJsonURLs map[string]string, useRemote, verbose bool) (*models.ProxyModulesByRegistry, error) {
	homeDir, err := helpers.GetHomeDirPath()
	if err != nil {
		return nil, err
	}

	var folioModules, eurekaModules []*models.ProxyModule
	for registryName, installJsonURL := range installJsonURLs {
		decodedResponse, err := rs.processInstallJsonURL(&models.RegistryRequest{
			RegistryName:   registryName,
			InstallJsonURL: installJsonURL,
			HomeDir:        homeDir,
			UseRemote:      useRemote,
		})
		if err != nil {
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
				decodedResponse = append(decodedResponse, &models.ProxyModule{
					ID:     fmt.Sprintf("%s-%s", name, entry[field.ModuleVersionEntry].(string)),
					Action: "enable",
				})
			}
		}

		if len(decodedResponse) > 0 {
			if verbose {
				slog.Info(rs.Action.Name, "text", "Read registry", "name", registryName, "count", len(decodedResponse))
			}
			sort.Slice(decodedResponse, func(i, j int) bool {
				return decodedResponse[i].ID < decodedResponse[j].ID
			})
		}

		switch registryName {
		case constant.FolioRegistry:
			folioModules = decodedResponse
		case constant.EurekaRegistry:
			eurekaModules = decodedResponse
		}
	}

	return &models.ProxyModulesByRegistry{
		FolioModules:  folioModules,
		EurekaModules: eurekaModules,
	}, nil
}

func (rs *RegistrySvc) processInstallJsonURL(r *models.RegistryRequest) ([]*models.ProxyModule, error) {
	r.Metadata.FileName = fmt.Sprintf("install_%s.json", r.RegistryName)
	r.Metadata.Path = filepath.Join(r.HomeDir, r.Metadata.FileName)
	if !r.UseRemote || rs.Action.Param.SkipRegistry {
		decodedResponse, err := rs.readLocalFile(r)
		if err != nil {
			return nil, err
		}
		if len(decodedResponse) == 0 {
			return nil, appErrors.LocalInstallFileNotFound(err)
		}
		return decodedResponse, nil
	}

	decodedResponse, err := rs.readRemoteAndCreateLocalFile(r)
	if err != nil {
		return nil, err
	}

	return decodedResponse, nil
}

func (rs *RegistrySvc) readLocalFile(r *models.RegistryRequest) ([]*models.ProxyModule, error) {
	if err := helpers.IsRegularFile(r.Metadata.Path); err != nil {
		return nil, err
	}

	var decodedResponse []*models.ProxyModule
	if err := helpers.ReadJSONFromFile(r.Metadata.Path, &decodedResponse); err != nil {
		return nil, err
	}
	slog.Info(rs.Action.Name, "text", "Read registry", "name", r.RegistryName, "local", true, "file", r.Metadata.FileName)

	return decodedResponse, nil
}

func (rs *RegistrySvc) readRemoteAndCreateLocalFile(r *models.RegistryRequest) ([]*models.ProxyModule, error) {
	var decodedResponse []*models.ProxyModule
	if err := rs.HTTPClient.GetRetryReturnStruct(r.InstallJsonURL, map[string]string{}, &decodedResponse); err != nil {
		return nil, err
	}
	if err := helpers.WriteJSONToFile(r.Metadata.Path, decodedResponse); err != nil {
		return nil, err
	}
	slog.Info(rs.Action.Name, "text", "Read registry", "name", r.RegistryName, "local", false, "file", r.Metadata.FileName)

	return decodedResponse, nil
}

func (rs *RegistrySvc) ExtractModuleMetadata(modules *models.ProxyModulesByRegistry) {
	allModules := [][]*models.ProxyModule{modules.FolioModules, modules.EurekaModules}
	for _, moduleSet := range allModules {
		for i, module := range moduleSet {
			if module.ID == "okapi" {
				continue
			}
			module.Metadata.Name = helpers.GetModuleNameFromID(module.ID)
			module.Metadata.Version = helpers.GetOptionalModuleVersion(module.ID)
			module.Metadata.SidecarName = rs.getSidecarName(module)
			moduleSet[i] = module
		}
	}
}

// TODO Add tests
func (rs *RegistrySvc) getSidecarName(module *models.ProxyModule) string {
	if strings.HasPrefix(module.Metadata.Name, "edge") {
		return module.Metadata.Name
	} else {
		return helpers.GetSidecarName(module.Metadata.Name)
	}
}
