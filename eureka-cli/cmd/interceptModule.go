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
	"sync"

	"github.com/docker/docker/client"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/models"
	"github.com/spf13/cobra"
)

// interceptModuleCmd represents the interceptModule command
var interceptModuleCmd = &cobra.Command{
	Use:   "interceptModule",
	Short: "Intercept module",
	Long:  `Intercept/redirect module traffic to IntelliJ.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.InterceptModule)
		if err != nil {
			return err
		}

		return r.InterceptModule()
	},
}

func (r *Run) InterceptModule() error {
	myModule, err := models.NewInterceptModule(r.RunConfig.Action, actionParams.ID, actionParams.DefaultGateway, actionParams.ModuleURL, actionParams.SidecarURL)
	if err != nil {
		return err
	}

	slog.Info(r.RunConfig.Action.Name, "text", "INTERCEPTING MODULE", "module", myModule.ModuleName)
	globalEnv := action.GetConfigEnvVars(field.Env)
	globalSidecarEnv := action.GetConfigEnvVars(field.SidecarModuleEnv)
	backendModules, err := r.RunConfig.ModuleParams.ReadBackendModulesFromConfig(false)
	if err != nil {
		return err
	}

	instalJsonURLs := map[string]string{
		constant.FolioRegistry:  r.RunConfig.Action.ConfigFolioRegistry,
		constant.EurekaRegistry: r.RunConfig.Action.ConfigEurekaRegistry,
	}
	registryModules, err := r.RunConfig.RegistrySvc.GetModules(instalJsonURLs, false)
	if err != nil {
		return err
	}
	r.RunConfig.RegistrySvc.ExtractModuleNameAndVersion(registryModules)

	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return err
	}
	defer r.RunConfig.DockerClient.Close(client)

	slog.Info(r.RunConfig.Action.Name, "text", "UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	pattern := fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, r.RunConfig.Action.ConfigProfile, myModule.ModuleName)
	err = r.RunConfig.ModuleSvc.UndeployModuleByNamePattern(client, pattern)
	if err != nil {
		return err
	}

	registryHosts := map[string]string{
		constant.FolioRegistry:  "",
		constant.EurekaRegistry: "",
	}
	myModule.Containers = models.NewCoreAndBusinessContainers(r.RunConfig.Action.VaultRootToken, registryHosts, registryModules, backendModules, globalEnv, globalSidecarEnv)

	err = r.UpdateModuleDiscovery(*myModule.SidecarURL)
	if err != nil {
		return err
	}

	if actionParams.Restore {
		return r.deployDefaultModuleAndSidecar(myModule, client)
	}

	return r.deployCustomSidecarForInterception(!actionParams.Restore, myModule, client)
}

func (r *Run) deployDefaultModuleAndSidecar(myModule *models.InterceptModule, client *client.Client) error {
	slog.Info(r.RunConfig.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	myModule.ClearURLs()

	err := r.prepareContainerNetwork(myModule, true)
	if err != nil {
		return err
	}

	err = r.deployModule(myModule, client)
	if err != nil {
		return err
	}

	err = r.deploySidecar(false, myModule, client)
	if err != nil {
		return err
	}

	slog.Info(r.RunConfig.Action.Name, "text", "WAITING FOR MODULE TO INITIALIZE")
	var deployModuleWG sync.WaitGroup
	errCh := make(chan error, 1)

	deployModuleWG.Add(1)
	go r.RunConfig.ModuleSvc.CheckModuleReadiness(&deployModuleWG, errCh, myModule.ModuleName, myModule.BackendModule.ModuleExposedServerPort)
	deployModuleWG.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Run) deployCustomSidecarForInterception(printModuleEnv bool, myModule *models.InterceptModule, client *client.Client) error {
	slog.Info(r.RunConfig.Action.Name, "text", "DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION")
	err := r.prepareContainerNetwork(myModule, false)
	if err != nil {
		return err
	}

	return r.deploySidecar(printModuleEnv, myModule, client)
}

func (r *Run) prepareContainerNetwork(myModule *models.InterceptModule, moduleAndSidecar bool) error {
	myModule.NetworkConfig = helpers.NewModuleNetworkConfig()

	if moduleAndSidecar {
		moduleServerPort, err := r.RunConfig.Action.GetPreReservedPort()
		if err != nil {
			return err
		}
		moduleDebugPort, err := r.RunConfig.Action.GetPreReservedPort()
		if err != nil {
			return err
		}
		sidecarServerPort, err := r.RunConfig.Action.GetPreReservedPort()
		if err != nil {
			return err
		}
		sidecarDebugPort, err := r.RunConfig.Action.GetPreReservedPort()
		if err != nil {
			return err
		}

		myModule.BackendModule, myModule.RegistryModule = r.RunConfig.ModuleSvc.GetBackendModule(myModule.Containers, myModule.ModuleName)
		myModule.BackendModule.ModulePortBindings = helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, myModule.BackendModule.ModuleServerPort)
		myModule.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, myModule.BackendModule.ModuleServerPort)
		myModule.BackendModule.ModuleExposedServerPort = moduleServerPort

		return nil
	}

	sidecarServerPort, err := helpers.ExtractPortFromURL(*myModule.SidecarURL)
	if err != nil {
		return err
	}
	sidecarDebugPort, err := r.RunConfig.Action.GetPreReservedPort()
	if err != nil {
		return err
	}

	myModule.SidecarServerPort = sidecarServerPort

	myModule.BackendModule, myModule.RegistryModule = r.RunConfig.ModuleSvc.GetBackendModule(myModule.Containers, myModule.ModuleName)
	myModule.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, myModule.BackendModule.ModuleServerPort)

	return nil
}

func (r *Run) deployModule(myModule *models.InterceptModule, client *client.Client) error {
	moduleVersion := r.RunConfig.ModuleSvc.GetModuleImageVersion(*myModule.BackendModule, myModule.RegistryModule)
	moduleImage := r.RunConfig.ModuleSvc.GetModuleImage(moduleVersion, myModule.RegistryModule)
	moduleEnv := r.RunConfig.ModuleSvc.GetModuleEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule)
	moduleContainer := models.NewModuleContainer(myModule.RegistryModule.Name, moduleImage, moduleEnv, *myModule.BackendModule, myModule.NetworkConfig)

	err := r.RunConfig.ModuleSvc.DeployModule(client, moduleContainer)
	if err != nil {
		return err
	}

	return nil
}

func (r *Run) deploySidecar(printModuleEnv bool, myModule *models.InterceptModule, client *client.Client) error {
	sidecarImage, pullSidecarImage, err := r.RunConfig.ModuleSvc.GetSidecarImage(myModule.Containers.RegistryModules[constant.EurekaRegistry])
	if err != nil {
		return err
	}

	sidecarResources := helpers.CreateResources(false, r.RunConfig.Action.ConfigSidecarResources)
	sidecarEnv := r.RunConfig.ModuleSvc.GetSidecarEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule, myModule.ModuleURL, myModule.SidecarURL)
	sidecarContainer := models.NewSidecarContainer(myModule.RegistryModule.SidecarName, sidecarImage, sidecarEnv, *myModule.BackendModule, myModule.NetworkConfig, sidecarResources)
	sidecarContainer.PullImage = pullSidecarImage

	err = r.RunConfig.ModuleSvc.DeployModule(client, sidecarContainer)
	if err != nil {
		return err
	}

	if printModuleEnv {
		moduleEnv := r.RunConfig.ModuleSvc.GetModuleEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule)

		if myModule.BackendModule.UseOkapiURL {
			moduleOkapiEnv := []string{"OKAPI_HOST=localhost",
				fmt.Sprintf("OKAPI_PORT=%d", myModule.SidecarServerPort),
				"OKAPI_SERVICE_HOST=localhost",
				fmt.Sprintf("OKAPI_SERVICE_PORT=%d", myModule.SidecarServerPort),
				fmt.Sprintf("OKAPI_SERVICE_URL=http://localhost:%d", myModule.SidecarServerPort),
				fmt.Sprintf("OKAPI_URL=http://localhost:%d", myModule.SidecarServerPort),
			}

			moduleEnv = append(moduleEnv, moduleOkapiEnv...)
		}

		fmt.Println()
		fmt.Printf("%s ###\n", "Can be embedded into IntelliJ Run/Debug Configuration")
		for _, value := range moduleEnv {
			fmt.Println(value)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.ModuleURL, "moduleUrl", "m", "", "Module URL, e.g. http://host.docker.internal:36002 or 36002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().StringVarP(&actionParams.SidecarURL, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002 or 37002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().BoolVarP(&actionParams.Restore, "restore", "r", false, "Restore module & sidecar")
	interceptModuleCmd.PersistentFlags().BoolVarP(&actionParams.DefaultGateway, "defaultGateway", "g", false, "Use default gateway in URLs, .e.g http://host.docker.internal:{{port}} will be set automatically")
	if err := interceptModuleCmd.MarkPersistentFlagRequired("id"); err != nil {
		slog.Error("failed to mark id flag as required", "error", err)
		os.Exit(1)
	}
}
