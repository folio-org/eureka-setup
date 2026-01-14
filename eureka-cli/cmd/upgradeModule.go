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
	"fmt"
	"log/slog"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/field"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/folio-org/eureka-setup/eureka-cli/modulesvc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// upgradeModuleCmd represents the upgradeModule command
var upgradeModuleCmd = &cobra.Command{
	Use:   "upgradeModule",
	Short: "Upgrade module",
	Long:  `Upgrade a single backend module in the current profile.`,
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
	if err := run.setModuleDiscoveryDataIntoContext(); err != nil {
		return err
	}
	if err := run.Config.UpgradeModuleSvc.SetNewModuleVersionAndIDIntoContext(); err != nil {
		return err
	}
	run.Config.UpgradeModuleSvc.SetDefaultNamespaceIntoContext()

	var (
		moduleName       = params.ModuleName
		newModuleVersion = params.ModuleVersion
		modulePath       = params.ModulePath
		namespace        = params.Namespace
		shouldBuild      = !helpers.IsFolioNamespace(params.Namespace)
	)
	if err := run.validateModulePath(modulePath); err != nil {
		return err
	}

	slog.Info(run.Config.Action.Name, "text", "UPGRADING MODULE", "module", moduleName, "version", newModuleVersion, "build", shouldBuild)
	if shouldBuild {
		if !params.SkipModuleArtifact {
			if err := run.Config.UpgradeModuleSvc.BuildModuleArtifact(moduleName, newModuleVersion, modulePath); err != nil {
				return err
			}
		}
		if !params.SkipModuleImage {
			if err := run.Config.UpgradeModuleSvc.BuildModuleImage(namespace, moduleName, newModuleVersion, modulePath); err != nil {
				return err
			}
		}
	}

	var newModuleDescriptor map[string]any
	if shouldBuild {
		readModuleDescriptor, err := run.Config.UpgradeModuleSvc.ReadModuleDescriptor(moduleName, newModuleVersion, modulePath)
		if err != nil {
			return err
		}
		newModuleDescriptor = readModuleDescriptor
	}
	if !params.SkipModuleDeployment {
		if err := run.deployNewModuleAndSidecarPair(); err != nil {
			return err
		}
	}

	app, err := run.Config.ManagementSvc.GetLatestApplication()
	if err != nil {
		return err
	}
	oldBackendModules := helpers.GetAnySlice(app, "modules")
	newBackendModules, newDiscoveryModules, oldModuleID, err := run.Config.UpgradeModuleSvc.UpdateBackendModules(moduleName, newModuleVersion, shouldBuild, oldBackendModules)
	if err != nil {
		return err
	}
	oldFrontendModules := helpers.GetAnySlice(app, "uiModules")
	newFrontendModules := run.Config.UpgradeModuleSvc.UpdateFrontendModules(shouldBuild, oldFrontendModules)

	var newBackendModuleDescriptors []any
	if shouldBuild {
		oldBackendModuleDescriptors := helpers.GetAnySlice(app, "moduleDescriptors")
		newBackendModuleDescriptors = run.Config.UpgradeModuleSvc.UpdateBackendModuleDescriptors(moduleName, oldModuleID, newModuleDescriptor, oldBackendModuleDescriptors)
	}

	oldAppVersion := helpers.GetString(app, "version")
	appVersion, err := semver.NewVersion(oldAppVersion)
	if err != nil {
		return err
	}
	newAppVersion := appVersion.IncPatch().String()

	var (
		appName  = helpers.GetString(app, "name")
		newAppID = fmt.Sprintf("%s-%s", appName, newAppVersion)
	)
	if !params.SkipApplication {
		newDependencies := helpers.GetMapOrDefault(app, "dependencies", nil)
		newFrontendModuleDescriptors := helpers.GetAnySlice(app, "uiModuleDescriptors")
		if err := run.Config.ManagementSvc.CreateNewApplication(&models.ApplicationUpgradeRequest{
			ApplicationName:              appName,
			NewApplicationID:             newAppID,
			NewApplicationVersion:        newAppVersion,
			NewDependencies:              newDependencies,
			NewBackendModules:            newBackendModules,
			NewFrontendModules:           newFrontendModules,
			NewBackendModuleDescriptors:  newBackendModuleDescriptors,
			NewFrontendModuleDescriptors: newFrontendModuleDescriptors,
			ShouldBuild:                  shouldBuild,
		}); err != nil {
			return err
		}
	}
	if !params.SkipModuleDiscovery {
		if err := run.Config.ManagementSvc.CreateNewModuleDiscovery(newDiscoveryModules); err != nil {
			if downstreamErr := run.cleanupApplicationsOnFailure(constant.NoneConsortium, constant.All, appName, err); downstreamErr != nil {
				return downstreamErr
			}

			return err
		}
	}
	if !params.SkipTenantEntitlement {
		slog.Info(run.Config.Action.Name, "text", "UPGRADING TENANT ENTITLEMENT", "from", oldAppVersion, "to", newAppVersion)
		if err := run.Config.ManagementSvc.UpgradeTenantEntitlement(constant.NoneConsortium, constant.All, newAppID); err != nil {
			if downstreamErr := run.cleanupApplicationsOnFailure(constant.NoneConsortium, constant.All, appName, err); downstreamErr != nil {
				return downstreamErr
			}

			return err
		}
	}
	slog.Info(run.Config.Action.Name, "text", "REMOVING APPLICATIONS", "name", appName)
	if err := run.Config.ManagementSvc.RemoveApplications(appName, newAppID); err != nil {
		return err
	}
	if params.Cleanup {
		if err := run.Config.UpgradeModuleSvc.CleanModuleArtifact(moduleName, modulePath); err != nil {
			return err
		}
	}

	return nil
}

func (run *Run) deployNewModuleAndSidecarPair() error {
	slog.Info(run.Config.Action.Name, "text", "DEPLOYING NEW MODULE AND SIDECAR PAIR", "module", params.ModuleName, "id", params.ID)
	backendModules, err := run.Config.ModuleProps.ReadBackendModules(false, false)
	if err != nil {
		return err
	}

	installJsonURLs := run.Config.Action.GetCombinedInstallJsonURLs()
	modules, err := run.Config.RegistrySvc.GetModules(installJsonURLs, true, false)
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

func (run *Run) validateModulePath(modulePath string) error {
	if modulePath == "" {
		return nil
	}

	info, err := os.Stat(modulePath)
	if os.IsNotExist(err) {
		return errors.ModulePathNotFound(modulePath)
	}
	if err != nil {
		return errors.ModulePathAccessFailed(modulePath, err)
	}
	if !info.IsDir() {
		return errors.ModulePathNotDirectory(modulePath)
	}

	return nil
}

func (run *Run) cleanupApplicationsOnFailure(consortiumName string, tenantType constant.TenantType, appName string, upstreamErr error) error {
	tenants, err := run.Config.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return errors.Wrap(upstreamErr, "failed to cleanup apps on failure - cannot retrieve tenants")
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		tenantName := helpers.GetString(entry, "name")
		if !helpers.HasTenant(tenantName, run.Config.Action.ConfigTenants) {
			continue
		}

		entitlements, err := run.Config.ManagementSvc.GetTenantEntitlements(tenantName, false)
		if err != nil {
			return errors.Wrap(upstreamErr, "failed to cleanup apps on failure - cannot retrieve tenant entitlements")
		}
		entitlement := entitlements.Entitlements[0]

		slog.Info(run.Config.Action.Name, "text", "REMOVING APPLICATIONS ON FAILURE", "name", appName)
		if err := run.Config.ManagementSvc.RemoveApplications(appName, entitlement.ApplicationID); err != nil {
			return errors.Wrapf(err, "failed to cleanup apps on failure - cannot remove apps (%s app id is ignored)", entitlement.ApplicationID)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(upgradeModuleCmd)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModuleVersion, action.ModuleVersion.Long, action.ModuleVersion.Short, "", action.ModuleVersion.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.ModulePath, action.ModulePath.Long, action.ModulePath.Short, "", action.ModulePath.Description)
	upgradeModuleCmd.PersistentFlags().StringVarP(&params.Namespace, action.Namespace.Long, action.Namespace.Short, "", action.Namespace.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.Cleanup, action.Cleanup.Long, action.Cleanup.Short, false, action.Cleanup.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleArtifact, action.SkipModuleArtifact.Long, action.SkipModuleArtifact.Short, false, action.SkipModuleArtifact.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleImage, action.SkipModuleImage.Long, action.SkipModuleImage.Short, false, action.SkipModuleImage.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDeployment, action.SkipModuleDeployment.Long, action.SkipModuleDeployment.Short, false, action.SkipModuleDeployment.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipApplication, action.SkipApplication.Long, action.SkipApplication.Short, false, action.SkipApplication.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDiscovery, action.SkipModuleDiscovery.Long, action.SkipModuleDiscovery.Short, false, action.SkipModuleDiscovery.Description)
	upgradeModuleCmd.PersistentFlags().BoolVarP(&params.SkipTenantEntitlement, action.SkipTenantEntitlement.Long, action.SkipTenantEntitlement.Short, false, action.SkipTenantEntitlement.Description)

	if err := upgradeModuleCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}
	if err := upgradeModuleCmd.MarkPersistentFlagRequired(action.ModulePath.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModulePath, err).Error())
		os.Exit(1)
	}

	if err := upgradeModuleCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
	if err := upgradeModuleCmd.RegisterFlagCompletionFunc(action.Namespace.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return constant.GetNamespaces(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
