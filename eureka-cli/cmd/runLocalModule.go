/*
Copyright © 2026 Open Library Foundation

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
	"strconv"

	"github.com/Masterminds/semver/v3"
	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
	"github.com/spf13/cobra"
)

const (
	defaultLocalApplicationName = "app-local"
	localApplicationBaseVersion = "1.0.0"
)

// runLocalModuleCmd represents the runLocalModule command
var runLocalModuleCmd = &cobra.Command{
	Use:   "runLocalModule",
	Short: "Run a local module",
	Long: `Build and run a private, not-yet-published backend module from its local source folder.

The module is layered into the environment through a dedicated child application (default app-local)
that depends on the currently deployed base application, so modules that are not registered in
FOLIO LSP/FAR can be developed, integration-tested and demoed without publishing them first.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.RunLocalModule)
		if err != nil {
			return err
		}

		return run.RunLocalModule()
	},
}

func (run *Run) RunLocalModule() error {
	if err := run.setKeycloakMasterAccessTokenIntoContext(constant.ClientCredentials); err != nil {
		return err
	}
	if err := run.validateModulePath(params.ModulePath); err != nil {
		return err
	}
	run.Config.UpgradeModuleSvc.SetDefaultNamespaceIntoContext()
	if err := run.resolveModuleIdentity(); err != nil {
		return err
	}

	// The --applicationName flag binds the same params field shared with undeployApplication (whose default
	// is empty), so its default cannot be relied on here; fall back to app-local when it is not set.
	appName := params.ApplicationName
	if appName == "" {
		appName = defaultLocalApplicationName
	}

	var (
		moduleName    = params.ModuleName
		moduleVersion = params.ModuleVersion
		modulePath    = params.ModulePath
		namespace     = params.Namespace
		shouldBuild   = !helpers.IsFolioNamespace(params.Namespace)
	)

	baseApp, err := run.Config.ManagementSvc.GetLatestApplication()
	if err != nil {
		return err
	}
	baseAppName := helpers.GetString(baseApp, "name")
	baseAppVersion := helpers.GetString(baseApp, "version")
	for _, value := range helpers.GetAnySlice(baseApp, "modules") {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if helpers.GetString(entry, "name") == moduleName {
			return errors.ModuleAlreadyInBaseApplication(moduleName, baseAppName)
		}
	}

	slog.Info(run.Config.Action.Name, "text", "RUNNING LOCAL MODULE", "module", moduleName, "version", moduleVersion, "application", appName, "build", shouldBuild)
	if shouldBuild {
		if !params.SkipModuleArtifact {
			if err := run.Config.UpgradeModuleSvc.BuildModuleArtifact(moduleName, moduleVersion, modulePath); err != nil {
				return err
			}
		}
		if !params.SkipModuleImage {
			if err := run.Config.UpgradeModuleSvc.BuildModuleImage(namespace, moduleName, moduleVersion, modulePath); err != nil {
				return err
			}
		}
	}

	var (
		newModuleDescriptor map[string]any
		descriptorPath      string
	)
	if shouldBuild {
		newModuleDescriptor, err = run.Config.UpgradeModuleSvc.ReadModuleDescriptor(moduleName, moduleVersion, modulePath)
		if err != nil {
			return err
		}
		descriptorPath, err = run.Config.UpgradeModuleSvc.GetModuleDescriptorPath(modulePath)
		if err != nil {
			return err
		}
	}

	if !params.SkipModuleDeployment {
		if err := run.deployLocalModuleAndSidecarPair(descriptorPath); err != nil {
			return err
		}
	}

	newAppID, isNew, keepAppID, discovery, err := run.buildOrMergeLocalApp(appName, baseAppName, baseAppVersion, shouldBuild, newModuleDescriptor)
	if err != nil {
		return err
	}

	if !params.SkipModuleDiscovery {
		if err := run.Config.ManagementSvc.CreateNewModuleDiscovery(discovery); err != nil {
			if cleanupErr := run.cleanupLocalAppOnFailure(appName, keepAppID); cleanupErr != nil {
				return cleanupErr
			}

			return err
		}
	}
	if !params.SkipTenantEntitlement {
		if err := run.entitleTenantsToLocalApp(isNew, newAppID); err != nil {
			if cleanupErr := run.cleanupLocalAppOnFailure(appName, keepAppID); cleanupErr != nil {
				return cleanupErr
			}

			return err
		}
	}

	slog.Info(run.Config.Action.Name, "text", "REMOVING SUPERSEDED LOCAL APPLICATIONS", "name", appName)
	if err := run.Config.ManagementSvc.RemoveApplications(appName, newAppID); err != nil {
		return err
	}
	if params.Cleanup {
		if err := run.Config.UpgradeModuleSvc.CleanModuleArtifact(moduleName, modulePath); err != nil {
			return err
		}
	}
	slog.Info(run.Config.Action.Name, "text", "Local module running", "module", moduleName, "application", newAppID)

	return nil
}

func (run *Run) resolveModuleIdentity() error {
	needName := params.ModuleName == ""
	needVersion := params.ModuleVersion == ""
	if needName || needVersion {
		resolvedName, resolvedVersion, err := run.Config.UpgradeModuleSvc.ResolveModuleIdentity(params.ModulePath)
		if err != nil {
			return err
		}
		if needName {
			params.ModuleName = resolvedName
		}
		if needVersion {
			nextVersion, err := nextLocalModuleVersion(resolvedVersion)
			if err != nil {
				return err
			}
			params.ModuleVersion = nextVersion
		}
	}
	params.ID = fmt.Sprintf("%s-%s", params.ModuleName, params.ModuleVersion)

	slog.Info(run.Config.Action.Name, "text", "RESOLVED MODULE IDENTITY", "module", params.ModuleName, "version", params.ModuleVersion, "id", params.ID)

	return nil
}

func nextLocalModuleVersion(currentVersion string) (string, error) {
	if helpers.IsSnapshot(currentVersion) {
		return helpers.IncrementSnapshotVersion(currentVersion)
	}

	parsed, err := semver.NewVersion(currentVersion)
	if err != nil {
		return "", err
	}

	return parsed.IncPatch().String(), nil
}

func (run *Run) reserveUsedHostPorts() error {
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return err
	}
	defer run.Config.DockerClient.Close(client)

	deployed, err := run.Config.ModuleSvc.GetDeployedModules(client, filters.NewArgs())
	if err != nil {
		return err
	}

	var reserved int
	for _, summary := range deployed {
		for _, port := range summary.Ports {
			if port.PublicPort != 0 {
				run.Config.Action.ReservedPorts = append(run.Config.Action.ReservedPorts, int(port.PublicPort))
				reserved++
			}
		}
	}
	slog.Info(run.Config.Action.Name, "text", "Reserved host ports already in use by running containers", "count", reserved)

	return nil
}

func (run *Run) deployLocalModuleAndSidecarPair(descriptorPath string) error {
	if err := run.reserveUsedHostPorts(); err != nil {
		return err
	}

	return run.deployModuleAndSidecarPair(func(modules *models.ProxyModulesByRegistry, backendModules map[string]models.BackendModule) error {
		modules.FolioModules = append(modules.FolioModules, &models.ProxyModule{ID: params.ID, Action: "enable"})

		localBackendModule, err := run.newLocalBackendModule(descriptorPath)
		if err != nil {
			return err
		}
		backendModules[params.ModuleName] = *localBackendModule

		return nil
	})
}

func (run *Run) newLocalBackendModule(descriptorPath string) (*models.BackendModule, error) {
	port, err := run.Config.Action.GetPreReservedPort()
	if err != nil {
		return nil, err
	}
	privatePort, err := strconv.Atoi(constant.PrivateServerPort)
	if err != nil {
		return nil, err
	}

	// A non-empty LocalDescriptorPath makes the deploy use the just-built foliolocal image (PullImage=false)
	return models.NewBackendModuleWithSidecar(run.Config.Action, models.BackendModuleProperties{
		DeployModule:        true,
		DeploySidecar:       helpers.BoolPtr(true),
		LocalDescriptorPath: descriptorPath,
		Name:                params.ModuleName,
		Version:             helpers.StringPtr(params.ModuleVersion),
		Port:                helpers.IntPtr(port),
		PrivatePort:         helpers.IntPtr(privatePort),
		Env:                 map[string]any{},
		SidecarEnv:          map[string]any{},
		Resources:           map[string]any{},
		Volumes:             []string{},
	})
}

func (run *Run) buildOrMergeLocalApp(applicationName, baseAppName, baseAppVersion string, shouldBuild bool, newModuleDescriptor map[string]any) (newAppID string, isNew bool, keepApplicationID string, discovery []map[string]string, err error) {
	existing, err := run.Config.ManagementSvc.GetLatestApplicationByName(applicationName)
	if err != nil {
		return "", false, "", nil, err
	}

	privatePort, err := strconv.Atoi(constant.PrivateServerPort)
	if err != nil {
		return "", false, "", nil, err
	}
	localModule := map[string]any{
		"id":      params.ID,
		"name":    params.ModuleName,
		"version": params.ModuleVersion,
	}
	discovery = []map[string]string{{
		"id":       params.ID,
		"name":     params.ModuleName,
		"version":  params.ModuleVersion,
		"location": helpers.GetSidecarURL(params.ModuleName, privatePort),
	}}

	var (
		newVersion                string
		dependencies              any
		backendModules            []map[string]any
		backendModuleDescriptors  []any
		frontendModules           []map[string]any
		frontendModuleDescriptors []any
	)
	if existing == nil {
		isNew = true
		// First deployment starts at 1.0.0, consistent with the base application and other child apps
		newVersion = localApplicationBaseVersion
		dependencies = map[string]any{"name": baseAppName, "version": baseAppVersion}
		backendModules = []map[string]any{localModule}
		if shouldBuild && newModuleDescriptor != nil {
			backendModuleDescriptors = []any{newModuleDescriptor}
		}
	} else {
		keepApplicationID = helpers.GetString(existing, "id")
		oldVersion, err := semver.NewVersion(helpers.GetString(existing, "version"))
		if err != nil {
			return "", false, "", nil, err
		}
		newVersion = oldVersion.IncPatch().String()
		dependencies = existing["dependencies"]
		backendModules, backendModuleDescriptors = mergeLocalBackendModules(existing, localModule, newModuleDescriptor)
		frontendModules = convertAnySliceToMapSlice(helpers.GetAnySlice(existing, "uiModules"))
		frontendModuleDescriptors = helpers.GetAnySlice(existing, "uiModuleDescriptors")
	}

	newAppID = fmt.Sprintf("%s-%s", applicationName, newVersion)
	if params.SkipApplication {
		return newAppID, isNew, keepApplicationID, discovery, nil
	}

	if err := run.Config.ManagementSvc.CreateNewApplication(&models.ApplicationUpgradeRequest{
		ApplicationName:              applicationName,
		NewApplicationID:             newAppID,
		NewApplicationVersion:        newVersion,
		NewDependencies:              dependencies,
		NewBackendModules:            backendModules,
		NewFrontendModules:           frontendModules,
		NewBackendModuleDescriptors:  backendModuleDescriptors,
		NewFrontendModuleDescriptors: frontendModuleDescriptors,
		ShouldBuild:                  shouldBuild,
	}); err != nil {
		return "", false, "", nil, err
	}

	return newAppID, isNew, keepApplicationID, discovery, nil
}

func mergeLocalBackendModules(existing, localModule, newModuleDescriptor map[string]any) ([]map[string]any, []any) {
	moduleName := helpers.GetString(localModule, "name")

	var (
		backendModules []map[string]any
		replaced       bool
	)
	for _, value := range helpers.GetAnySlice(existing, "modules") {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if helpers.GetString(entry, "name") == moduleName {
			backendModules = append(backendModules, localModule)
			replaced = true
			continue
		}
		backendModules = append(backendModules, map[string]any{
			"id":      helpers.GetString(entry, "id"),
			"name":    helpers.GetString(entry, "name"),
			"version": helpers.GetString(entry, "version"),
		})
	}
	if !replaced {
		backendModules = append(backendModules, localModule)
	}

	var backendModuleDescriptors []any
	for _, value := range helpers.GetAnySlice(existing, "moduleDescriptors") {
		if entry, ok := value.(map[string]any); ok && helpers.GetModuleNameFromID(helpers.GetString(entry, "id")) == moduleName {
			continue
		}
		backendModuleDescriptors = append(backendModuleDescriptors, value)
	}
	if newModuleDescriptor != nil {
		backendModuleDescriptors = append(backendModuleDescriptors, newModuleDescriptor)
	}

	return backendModules, backendModuleDescriptors
}

func convertAnySliceToMapSlice(values []any) []map[string]any {
	if len(values) == 0 {
		return nil
	}

	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if entry, ok := value.(map[string]any); ok {
			result = append(result, entry)
		}
	}

	return result
}

func (run *Run) entitleTenantsToLocalApp(isNew bool, newAppID string) error {
	if isNew {
		slog.Info(run.Config.Action.Name, "text", "ENTITLING TENANTS TO LOCAL APPLICATION", "application", newAppID)
		return run.Config.ManagementSvc.CreateTenantEntitlementForApplication(constant.NoneConsortium, constant.All, newAppID)
	}

	slog.Info(run.Config.Action.Name, "text", "UPGRADING TENANT ENTITLEMENT TO LOCAL APPLICATION", "application", newAppID)
	return run.Config.ManagementSvc.UpgradeTenantEntitlement(constant.NoneConsortium, constant.All, newAppID)
}

func (run *Run) cleanupLocalAppOnFailure(applicationName, keepApplicationID string) error {
	slog.Info(run.Config.Action.Name, "text", "REMOVING LOCAL APPLICATION ON FAILURE", "name", applicationName, "keep", keepApplicationID)

	return run.Config.ManagementSvc.RemoveApplications(applicationName, keepApplicationID)
}

func init() {
	rootCmd.AddCommand(runLocalModuleCmd)
	runLocalModuleCmd.PersistentFlags().StringVarP(&params.ModulePath, action.ModulePath.Long, action.ModulePath.Short, "", action.ModulePath.Description)
	runLocalModuleCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	runLocalModuleCmd.PersistentFlags().StringVarP(&params.ModuleVersion, action.ModuleVersion.Long, action.ModuleVersion.Short, "", action.ModuleVersion.Description)
	runLocalModuleCmd.PersistentFlags().StringVarP(&params.ApplicationName, action.ApplicationName.Long, action.ApplicationName.Short, defaultLocalApplicationName, action.ApplicationName.Description)
	runLocalModuleCmd.PersistentFlags().StringVarP(&params.Namespace, action.Namespace.Long, action.Namespace.Short, "", action.Namespace.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.Cleanup, action.Cleanup.Long, action.Cleanup.Short, false, action.Cleanup.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleArtifact, action.SkipModuleArtifact.Long, action.SkipModuleArtifact.Short, false, action.SkipModuleArtifact.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleImage, action.SkipModuleImage.Long, action.SkipModuleImage.Short, false, action.SkipModuleImage.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDeployment, action.SkipModuleDeployment.Long, action.SkipModuleDeployment.Short, false, action.SkipModuleDeployment.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipApplication, action.SkipApplication.Long, action.SkipApplication.Short, false, action.SkipApplication.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipModuleDiscovery, action.SkipModuleDiscovery.Long, action.SkipModuleDiscovery.Short, false, action.SkipModuleDiscovery.Description)
	runLocalModuleCmd.PersistentFlags().BoolVarP(&params.SkipTenantEntitlement, action.SkipTenantEntitlement.Long, action.SkipTenantEntitlement.Short, false, action.SkipTenantEntitlement.Description)

	if err := runLocalModuleCmd.MarkPersistentFlagRequired(action.ModulePath.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModulePath, err).Error())
		os.Exit(1)
	}

	if err := runLocalModuleCmd.RegisterFlagCompletionFunc(action.Namespace.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return constant.GetNamespaces(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
