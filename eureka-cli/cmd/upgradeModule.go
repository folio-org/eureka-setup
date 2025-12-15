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
	"github.com/folio-org/eureka-cli/modulesvc"
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
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}

	var (
		moduleName       = run.Config.Action.Param.ModuleName
		newModuleVersion = run.Config.Action.Param.ModuleVersion
		modulePath       = run.Config.Action.Param.ModulePath
		namespace        = run.Config.Action.Param.Namespace
	)
	params.ID = fmt.Sprintf("%s-%s", moduleName, newModuleVersion)

	slog.Info(run.Config.Action.Name, "text", "UPGRADING MODULE", "module", moduleName, "version", newModuleVersion)
	if !run.Config.Action.Param.SkipBuildModuleArtifact {
		if err := run.buildModuleArtifact(moduleName, newModuleVersion, modulePath); err != nil {
			return err
		}
	}
	if !run.Config.Action.Param.SkipBuildModuleImage {
		if err := run.buildModuleImage(namespace, moduleName, newModuleVersion, modulePath); err != nil {
			return err
		}
	}

	newModuleDescriptor, err := run.readModuleDescriptor(moduleName, newModuleVersion, modulePath)
	if err != nil {
		return err
	}
	if !run.Config.Action.Param.SkipModuleDeployment {
		if err := run.deployNewModuleAndSidecarPair(); err != nil {
			return err
		}
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
		oldAppVersion = appVersion.String()
		newAppVersion = appVersion.IncPatch().String()
		newAppID      = fmt.Sprintf("%s-%s", appName, newAppVersion)
	)
	headers, err := helpers.SecureApplicationJSONHeaders(run.Config.Action.KeycloakMasterAccessToken)
	if err != nil {
		return err
	}

	newBackendModules, newDiscoveryModules, oldModuleID, err := run.updateBackendModules(moduleName, newModuleVersion, app["modules"].([]any))
	if err != nil {
		return err
	}

	newFrontendModules := run.updateFrontendModules(app["uiModules"].([]any))
	newModuleDescriptors := run.updateBackendModuleDescriptors(oldModuleID, app["moduleDescriptors"].([]any), newModuleDescriptor)
	if !run.Config.Action.Param.SkipApplication {
		if err := run.createNewApplication(app, appName, newAppID, newAppVersion, newBackendModules, newFrontendModules, newModuleDescriptors, headers); err != nil {
			return err
		}
	}
	if !run.Config.Action.Param.SkipModuleDiscovery {
		if err := run.createNewModuleDiscovery(newDiscoveryModules, headers); err != nil {
			if downstreamErr := run.cleanupApplicationsOnFailure(appName, headers, err); downstreamErr != nil {
				return downstreamErr
			}
			return err
		}
	}
	if !run.Config.Action.Param.SkipTenantEntitlement {
		if err := run.upgradeTenantEntitlement(oldAppVersion, newAppID, headers); err != nil {
			if downstreamErr := run.cleanupApplicationsOnFailure(appName, headers, err); downstreamErr != nil {
				return downstreamErr
			}
			return err
		}
	}

	return run.removeApplications(appName, newAppID, headers)
}

func (run *Run) buildModuleArtifact(moduleName, newModuleVersion, modulePath string) error {
	slog.Info(run.Config.Action.Name, "text", "BUILDING MODULE ARTIFACT", "module", moduleName)
	slog.Info(run.Config.Action.Name, "text", "Cleaning target directory", "module", moduleName, "path", modulePath)
	if err := run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "clean", "-DskipTests"), modulePath); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "Setting new artifact version", "module", moduleName)
	if err := run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "versions:set", fmt.Sprintf("-DnewVersion=%s", newModuleVersion)), modulePath); err != nil {
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Packaging new artifact", "module", moduleName)

	return run.Config.ExecSvc.ExecFromDir(exec.Command("mvn", "package", "-DskipTests"), modulePath)
}

func (run *Run) buildModuleImage(namespace, moduleName, newModuleVersion, modulePath string) error {
	imageName := fmt.Sprintf("%s/%s:%s", namespace, moduleName, newModuleVersion)
	slog.Info(run.Config.Action.Name, "text", "BUILDING MODULE IMAGE", "module", moduleName, "image", imageName)
	return run.Config.ExecSvc.ExecFromDir(exec.Command("docker", "build", "--tag", imageName,
		"--file", "./Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), modulePath)
}

func (run *Run) readModuleDescriptor(moduleName, newModuleVersion, modulePath string) (newModuleDescriptor map[string]any, err error) {
	slog.Info(run.Config.Action.Name, "text", "READING NEW MODULE DESCRIPTOR", "module", moduleName, "path", modulePath)
	descriptorPath := filepath.Join(modulePath, "target", "ModuleDescriptor.json")
	if err := helpers.ReadJSONFromFile(descriptorPath, &newModuleDescriptor); err != nil {
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

func (run *Run) updateBackendModules(moduleName, newModuleVersion string, modules []any) (newBackendModules []map[string]any, newDiscoveryModules []map[string]string, oldModuleID string, err error) {
	for _, value := range modules {
		entry := value.(map[string]any)

		if entry["name"] == moduleName {
			oldModuleID = entry["id"].(string)
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

	return newBackendModules, newDiscoveryModules, oldModuleID, nil
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

func (run *Run) updateBackendModuleDescriptors(oldModuleID string, moduleDescriptors []any, newModuleVersion map[string]any) (newModuleDescriptors []any) {
	for _, value := range moduleDescriptors {
		entry := value.(map[string]any)
		if entry["id"] == oldModuleID {
			continue
		}

		newModuleDescriptors = append(newModuleDescriptors, value)
	}

	return append(newModuleDescriptors, newModuleVersion)
}

func (run *Run) deployNewModuleAndSidecarPair() error {
	slog.Info(run.Config.Action.Name, "text", "DEPLOYING NEW MODULE AND SIDECAR PAIR", "module", params.ModuleName, "id", params.ID)
	backendModules, err := run.Config.ModuleProps.ReadBackendModules(false, false)
	if err != nil {
		return err
	}

	installURLs := run.Config.Action.GetCombinedInstallURLs()
	modules, err := run.Config.RegistrySvc.GetModules(installURLs, true, false)
	if err != nil {
		return err
	}
	run.Config.RegistrySvc.ExtractModuleMetadata(modules)

	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)
	if err := run.setVaultRootTokenIntoContext(client); err != nil {
		return err
	}

	pair, err := modulesvc.NewModulePair(run.Config.Action, run.Config.Action.Param)
	if err != nil {
		return err
	}

	pair.Containers = &models.Containers{
		Modules:        modules,
		BackendModules: backendModules,
		IsManagement:   false,
	}
	return run.Config.UpgradeModuleSvc.DeployModuleAndSidecarPair(client, pair)
}

func (run *Run) createNewApplication(app map[string]any, appName, newAppID, newAppVersion string, newBackendModules, newFrontendModules []map[string]any, newModuleDescriptors []any, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "CREATING NEW APPLICATION", "name", appName, "version", newAppVersion)
	payload, err := json.Marshal(map[string]any{
		"id":                  newAppID,
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

func (run *Run) upgradeTenantEntitlement(oldAppID, newAppID string, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "UPGRADING TENANT ENTITLEMENT", "from", oldAppID, "to", newAppID)
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
			"applications": []string{newAppID},
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

func (run *Run) cleanupApplicationsOnFailure(appName string, headers map[string]string, upstreamErr error) error {
	tenants, err := run.Config.ManagementSvc.GetTenants(constant.NoneConsortium, constant.Default)
	if err != nil {
		return errors.Wrap(upstreamErr, "failed to cleanup apps on failure - cannot retrieve tenants")
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := entry["name"].(string)
		if !helpers.HasTenant(tenantName, run.Config.Action.ConfigTenants) {
			continue
		}

		entitlements, err := run.Config.ManagementSvc.GetTenantEntitlements(tenantName, false)
		if err != nil {
			return errors.Wrap(upstreamErr, "failed to cleanup apps on failure - cannot retrieve tenant entitlements")
		}
		entitlement := entitlements.Entitlements[0]

		if err := run.removeApplications(appName, entitlement.ApplicationID, headers); err != nil {
			return errors.Wrapf(err, "failed to cleanup apps on failure (%s app id is ignored)", entitlement.ApplicationID)
		}
	}

	return nil
}

func (run *Run) removeApplications(appName, ignoreAppID string, headers map[string]string) error {
	slog.Info(run.Config.Action.Name, "text", "REMOVING APPLICATIONS", "name", appName)
	apps, err := run.Config.ManagementSvc.GetApplications()
	if err != nil {
		return err
	}

	for _, entry := range apps.ApplicationDescriptors {
		id := entry["id"]
		if id == ignoreAppID {
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
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModulePath, action.ModulePath.Long, action.ModulePath.Short, "", action.ModulePath.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.Namespace, action.Namespace.Long, action.Namespace.Short, constant.LocalNamespace, action.Namespace.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipBuildModuleArtifact, action.SkipBuildModuleArtifact.Long, action.SkipBuildModuleArtifact.Short, false, action.SkipBuildModuleArtifact.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipBuildModuleImage, action.SkipBuildModuleImage.Long, action.SkipBuildModuleImage.Short, false, action.SkipBuildModuleImage.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDeployment, action.SkipModuleDeployment.Long, action.SkipModuleDeployment.Short, false, action.SkipModuleDeployment.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipApplication, action.SkipApplication.Long, action.SkipApplication.Short, false, action.SkipApplication.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDiscovery, action.SkipModuleDiscovery.Long, action.SkipModuleDiscovery.Short, false, action.SkipModuleDiscovery.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipTenantEntitlement, action.SkipTenantEntitlement.Long, action.SkipTenantEntitlement.Short, false, action.SkipTenantEntitlement.Description)

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
