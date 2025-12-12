/*
Copyright © 2025 Open Library Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/errors"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// upgradeModuleCmd represents the upgradeModule command
var upgradeModuleCmd = &cobra.Command{
	Use:   "upgradeModule",
	Short: "Upgrade module",
	Long:  `Upgrade a single module for the current profile.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.UpgradeModule)
		if err != nil {
			return err
		}

		return run.UpgradeModule()
	},
}

func (run *Run) UpgradeModule() error {
	var (
		moduleName       = run.Config.Action.Param.ModuleName
		newModuleVersion = run.Config.Action.Param.ModuleVersion
		moduleTargetPath = run.Config.Action.Param.ModuleTargetPath
	)
	slog.Info(run.Config.Action.Name, "text", "UPGRADING MODULE", "module", moduleName, "version", newModuleVersion)
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	if run.Config.Action.Param.BuildModule {
		if err := run.buildModuleArtifact(moduleName, newModuleVersion, moduleTargetPath); err != nil {
			return err
		}
	}

	newModuleDescriptor, err := run.readModuleDescriptor(moduleName, newModuleVersion, moduleTargetPath)
	if err != nil {
		return err
	}

	app, err := run.getLatestApplication()
	if err != nil {
		return err
	}

	appVersion, err := semver.NewVersion(app["version"].(string))
	if err != nil {
		return err
	}

	var (
		appName       = app["name"].(string)
		newAppVersion = appVersion.IncPatch()
		newAppId      = fmt.Sprintf("%s-%s", appName, newAppVersion)
	)
	newBackendModules, newDiscoveryModules, oldModuleId, err := run.updateBackendModules(moduleName, newModuleVersion, app["modules"].([]any))
	if err != nil {
		return err
	}
	newFrontendModules := run.updateFrontendModules(app["uiModules"].([]any))
	newModuleDescriptors := run.updateBackendModuleDescriptors(oldModuleId, app["moduleDescriptors"].([]any), newModuleDescriptor)

	headers, err := helpers.SecureApplicationJSONHeaders(run.Config.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}
	if err := run.createNewApplication(app, appName, newAppId, newAppVersion.String(), newBackendModules, newFrontendModules, newModuleDescriptors, headers); err != nil {
		return err
	}
	if err := run.createNewModuleDiscovery(newDiscoveryModules, headers); err != nil {
		return err
	}
	if err := run.upgradeTenantEntitlement(newAppId, headers); err != nil {
		return err
	}
	if err := run.removeOldApplications(appName, newAppId, headers); err != nil {
		return err
	}

	return nil
}

func (run *Run) buildModuleArtifact(moduleName, newModuleVersion, moduleTargetPath string) error {
	if run.Config.Action.Param.BuildModule {
		slog.Info(run.Config.Action.Name, "text", "BUILDING MODULE ARTIFACT", "module", moduleName, "version", newModuleVersion)

		slog.Info(run.Config.Action.Name, "text", "Cleaning target directory")
		if err := run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "clean", "-DskipTests"), moduleTargetPath); err != nil {
			return err
		}

		slog.Info(run.Config.Action.Name, "text", "Setting new artifact version", "version", newModuleVersion)
		if err := run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "versions:set", fmt.Sprintf("-DnewVersion=%s", newModuleVersion)), moduleTargetPath); err != nil {
			return err
		}

		slog.Info(run.Config.Action.Name, "text", "Packaging the new artifact")
		if err := run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "package", "-DskipTests"), moduleTargetPath); err != nil {
			return err
		}
	}

	return nil
}

func (run *Run) readModuleDescriptor(moduleName, newModuleVersion, moduleTargetPath string) (newModuleDescriptor map[string]any, err error) {
	slog.Info(run.Config.Action.Name, "text", "READING NEW MODULE DESCRIPTOR", "module", moduleName, "path", moduleTargetPath)
	localPath := filepath.Join(moduleTargetPath, "target", "ModuleDescriptor.json")
	if err := helpers.ReadJSONFromFile(localPath, &newModuleDescriptor); err != nil {
		return nil, err
	}
	if len(newModuleDescriptor) == 0 {
		slog.Info(run.Config.Action.Name, "text", "New module descriptor was not found", "module", moduleName, "version", newModuleVersion)
		return nil, nil
	}

	return newModuleDescriptor, nil
}

func (run *Run) getLatestApplication() (map[string]any, error) {
	app, err := run.Config.ManagementSvc.GetLatestApplication()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (run *Run) updateBackendModules(moduleName, newModuleVersion string, modules []any) (newBackendModules []map[string]any, newDiscoveryModules []map[string]string, oldModuleId string, err error) {
	for _, value := range modules {
		entry := value.(map[string]any)

		if entry["name"] == moduleName {
			oldModuleId = entry["id"].(string)
			moduleID := fmt.Sprintf("%s-%s", moduleName, newModuleVersion)
			entry = map[string]any{
				"id":      moduleID,
				"name":    moduleName,
				"version": newModuleVersion,
			}

			privatePort, err := strconv.Atoi(constant.PrivateServerPort)
			if err != nil {
				return nil, nil, "", err
			}

			var sidecarURL string
			if strings.HasPrefix(moduleName, "edge") {
				sidecarURL = fmt.Sprintf("http://%s.eureka:%d", moduleName, privatePort)
			} else {
				sidecarURL = fmt.Sprintf("http://%s-sc.eureka:%d", moduleName, privatePort)
			}

			newDiscoveryModules = append(newDiscoveryModules, map[string]string{
				"id":       moduleID,
				"name":     moduleName,
				"version":  newModuleVersion,
				"location": sidecarURL,
			})
		} else {
			entry = map[string]any{
				"id":      entry["id"],
				"name":    entry["name"],
				"version": entry["version"],
			}
		}
		newBackendModules = append(newBackendModules, entry)
	}

	return newBackendModules, newDiscoveryModules, oldModuleId, nil
}

func (run *Run) updateFrontendModules(modules []any) (newFrontendModules []map[string]any) {
	for _, value := range modules {
		entry := value.(map[string]any)
		newFrontendModules = append(newFrontendModules, map[string]any{
			"id":      entry["id"],
			"name":    entry["name"],
			"version": entry["version"],
		})
	}

	return newFrontendModules
}

func (run *Run) updateBackendModuleDescriptors(oldModuleId string, moduleDescriptors []any, newModuleVersion map[string]any) (newModuleDescriptors []any) {
	for _, value := range moduleDescriptors {
		entry := value.(map[string]any)
		if entry["id"] == oldModuleId {
			continue
		}

		newModuleDescriptors = append(newModuleDescriptors, value)
	}

	return append(newModuleDescriptors, newModuleVersion)
}

func (run *Run) createNewApplication(app map[string]any, appName, newAppId, newAppVersion string, newBackendModules, newFrontendModules []map[string]any, newModuleDescriptors []any, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "CREATING NEW APPLICATION", "name", appName, "version", newAppVersion)
	payload, err := json.Marshal(map[string]any{
		"id":                  newAppId,
		"name":                appName,
		"version":             newAppVersion,
		"description":         "Default",
		"dependencies":        app["dependencies"],
		"modules":             newBackendModules,
		"uiModules":           newFrontendModules,
		"moduleDescriptors":   newModuleDescriptors,
		"uiModuleDescriptors": app["uiModuleDescriptors"],
	})
	if err != nil {
		return err
	}
	requestURL := run.Config.Action.GetRequestURL(constant.KongPort, "/applications?check=true")

	var appResponse models.ApplicationDescriptor
	if err := run.Config.HTTPClient.PostReturnStruct(requestURL, payload, headers, &appResponse); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Created application", "id", appResponse.ID, "backendModules", len(newBackendModules), "frontendModules", len(newFrontendModules))

	return nil
}

func (run *Run) createNewModuleDiscovery(newDiscoveryModules []map[string]string, headers map[string]string) error {
	payload, err := json.Marshal(map[string]any{
		"discovery": newDiscoveryModules,
	})
	if err != nil {
		return err
	}
	requestURL := run.Config.Action.GetRequestURL(constant.KongPort, "/modules/discovery")

	var discoveryResponse models.ModuleDiscoveryResponse
	if err := run.Config.HTTPClient.PostReturnStruct(requestURL, payload, headers, &discoveryResponse); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Created module discovery", "count", len(newDiscoveryModules), "totalRecords", discoveryResponse.TotalRecords)

	return nil
}

func (run *Run) upgradeTenantEntitlement(newAppId string, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "UPGRADING TENANT ENTITLEMENT", "id", newAppId)
	tenantParameters, err := run.Config.TenantSvc.GetEntitlementTenantParameters(constant.NoneConsortium)
	if err != nil {
		return err
	}

	tenants, err := run.Config.ManagementSvc.GetTenants(constant.NoneConsortium, constant.Default)
	if err != nil {
		return nil
	}

	requestURL := run.Config.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/entitlements?async=false&tenantParameters=%s", tenantParameters))
	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := entry["name"].(string)
		if !helpers.HasTenant(tenantName, run.Config.Action.ConfigTenants) {
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"tenantId":     entry["id"],
			"applications": []string{newAppId},
		})
		if err != nil {
			return err
		}

		var decodedResponse models.TenantEntitlementResponse
		if err := run.Config.HTTPClient.PutReturnStruct(requestURL, payload, headers, &decodedResponse); err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Upgraded tenant entitlement", "tenant", tenantName, "flowId", decodedResponse.FlowID)
	}

	return nil
}

func (run *Run) removeOldApplications(appName, newAppId string, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "REMOVING OLD APPLICATIONS", "name", appName)
	apps, err := run.Config.ManagementSvc.GetApplications()
	if err != nil {
		return err
	}

	for _, entry := range apps.ApplicationDescriptors {
		id := entry["id"]
		if id == newAppId {
			continue
		}
		requestURL := run.Config.Action.GetRequestURL(constant.KongPort, fmt.Sprintf("/applications/%s", id))

		var decodedResponse models.ApplicationDescriptor
		if err := run.Config.HTTPClient.DeleteReturnStruct(requestURL, headers, &decodedResponse); err != nil {
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Removed application", "id", id)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upgradeModuleCmd)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModuleVersion, action.ModuleVersion.Long, action.ModuleVersion.Short, "", action.ModuleVersion.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModuleTargetPath, action.ModuleTargetPath.Long, action.ModuleTargetPath.Short, "", action.ModuleTargetPath.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.BuildModule, action.BuildModule.Long, action.BuildModule.Short, false, action.BuildModule.Description)

	if err := upgradeModuleCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}
	if err := upgradeModuleCmd.MarkPersistentFlagRequired(action.ModuleVersion.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleVersion, err).Error())
		os.Exit(1)
	}

	if err := upgradeModuleCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
