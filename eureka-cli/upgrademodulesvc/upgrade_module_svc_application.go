package upgrademodulesvc

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
)

// UpgradeModuleApplicationBuilder defines the interface for updating application configurations during the module upgrade
type UpgradeModuleApplicationBuilder interface {
	UpdateBackendModules(moduleName, newModuleVersion string, shouldBuild bool, modules []any) ([]map[string]any, []map[string]string, string, error)
	UpdateFrontendModules(shouldBuild bool, modules []any) (newFrontendModules []map[string]any)
	UpdateBackendModuleDescriptors(moduleName, oldModuleID string, newModuleDescriptor map[string]any, moduleDescriptors []any) []any
}

func (um *UpgradeModuleSvc) UpdateBackendModules(moduleName, newModuleVersion string, shouldBuild bool, modules []any) ([]map[string]any, []map[string]string, string, error) {
	var (
		newBackendModules   []map[string]any
		newDiscoveryModules []map[string]string
		oldModuleID         string
	)
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Printf("\nDUMPING backend module entries\n")
	}
	for _, value := range modules {
		entry := value.(map[string]any)
		if helpers.GetString(entry, "name") == moduleName {
			oldModuleID = helpers.GetString(entry, "id")
			moduleID := fmt.Sprintf("%s-%s", moduleName, newModuleVersion)
			if shouldBuild {
				entry = map[string]any{
					"id":      moduleID,
					"name":    moduleName,
					"version": newModuleVersion,
				}
			} else {
				moduleDescriptorURL := um.Action.GetModuleURL(moduleID)
				entry = map[string]any{
					"id":      moduleID,
					"name":    moduleName,
					"version": newModuleVersion,
					"url":     moduleDescriptorURL,
				}
			}

			privatePort, err := strconv.Atoi(constant.PrivateServerPort)
			if err != nil {
				return nil, nil, "", err
			}
			sidecarURL := helpers.GetSidecarURL(moduleName, privatePort)

			newDiscoveryModules = append(newDiscoveryModules, map[string]string{
				"id":       moduleID,
				"name":     moduleName,
				"version":  newModuleVersion,
				"location": sidecarURL,
			})
		} else {
			entry = um.getDefaultModuleEntry(shouldBuild, entry)
		}
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			fmt.Println("backend module entry =", entry)
		}
		newBackendModules = append(newBackendModules, entry)
	}

	return newBackendModules, newDiscoveryModules, oldModuleID, nil
}

func (um *UpgradeModuleSvc) UpdateFrontendModules(shouldBuild bool, modules []any) (newFrontendModules []map[string]any) {
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Printf("\nDUMPING frontend module entries\n")
	}
	for _, value := range modules {
		entry := um.getDefaultModuleEntry(shouldBuild, value.(map[string]any))
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			fmt.Println("frontend module entry =", entry)
		}
		newFrontendModules = append(newFrontendModules, entry)
	}

	return newFrontendModules
}

func (um *UpgradeModuleSvc) getDefaultModuleEntry(shouldBuild bool, entry map[string]any) map[string]any {
	var (
		moduleID      = helpers.GetString(entry, "id")
		moduleName    = helpers.GetString(entry, "name")
		moduleVersion = helpers.GetString(entry, "version")
	)
	if shouldBuild {
		return map[string]any{
			"id":      moduleID,
			"name":    moduleName,
			"version": moduleVersion,
		}
	} else {
		moduleDescriptorURL := um.Action.GetModuleURL(moduleID)
		return map[string]any{
			"id":      moduleID,
			"name":    moduleName,
			"version": moduleVersion,
			"url":     moduleDescriptorURL,
		}
	}
}

func (um *UpgradeModuleSvc) UpdateBackendModuleDescriptors(moduleName, oldModuleID string, newModuleDescriptor map[string]any, moduleDescriptors []any) []any {
	var newModuleDescriptors []any
	for _, value := range moduleDescriptors {
		entry := value.(map[string]any)
		if helpers.GetString(entry, "id") == oldModuleID {
			continue
		}
		newModuleDescriptors = append(newModuleDescriptors, value)
	}
	if len(newModuleDescriptors) > 0 {
		slog.Info(um.Action.Name, "text", "Added new module descriptor", "module", moduleName)
		newModuleDescriptors = append(newModuleDescriptors, newModuleDescriptor)
	}

	return newModuleDescriptors
}
