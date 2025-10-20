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
	"github.com/spf13/viper"
)

// interceptModuleCmd represents the interceptModule command
var interceptModuleCmd = &cobra.Command{
	Use:   "interceptModule",
	Short: "Intercept module",
	Long:  `Intercept/redirect module traffic to IntelliJ.`,
	Run: func(cmd *cobra.Command, args []string) {
		startPort := viper.GetInt(field.ApplicationPortStart)
		endPort := viper.GetInt(field.ApplicationPortEnd)
		NewCustomRun(action.InterceptModule, startPort, endPort).InterceptModule()
	},
}

func (r *Run) InterceptModule() {
	myModule := models.NewInterceptModule(r.Config.Action, rp.ID, rp.DefaultGateway, rp.ModuleURL, rp.SidecarURL, viper.GetInt(field.ApplicationPortStart), viper.GetInt(field.ApplicationPortEnd))

	slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("INTERCEPTING %s MODULE", myModule.ModuleName))
	globalEnv := helpers.GetConfigEnvVars(field.Env)
	globalSidecarEnv := helpers.GetConfigEnvVars(field.SidecarModuleEnv)
	backendModulesMap := r.Config.ModuleParams.GetBackendModulesFromConfig(false, false, viper.GetStringMap(field.BackendModules))

	instalJsonURLs := map[string]string{
		constant.FolioRegistry:  viper.GetString(field.InstallFolio),
		constant.EurekaRegistry: viper.GetString(field.InstallEureka),
	}
	registryModules := r.Config.RegistryStep.GetModules(instalJsonURLs, false)
	r.Config.RegistryStep.ExtractModuleNameAndVersion(registryModules, false)

	vaultRootToken, client := r.GetVaultRootTokenWithDockerClient()
	defer func() {
		_ = client.Close()
	}()

	slog.Info(r.Config.Action.Name, "text", "UNDEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	r.Config.ModuleStep.UndeployModuleByNamePattern(client, fmt.Sprintf(constant.SingleModuleOrSidecarContainerPattern, viper.GetString(field.ProfileName), myModule.ModuleName), false)

	registryHosts := map[string]string{constant.FolioRegistry: "", constant.EurekaRegistry: ""}
	myModule.Containers = models.NewCoreAndBusinessContainers(vaultRootToken, registryHosts, registryModules, backendModulesMap, globalEnv, globalSidecarEnv)

	r.UpdateModuleDiscovery(*myModule.SidecarUrl)
	if rp.Restore {
		r.deployDefaultModuleAndSidecar(myModule, client)
		return
	}
	r.deployCustomSidecarForInterception(!rp.Restore, myModule, client)
}

func (r *Run) deployDefaultModuleAndSidecar(myModule *models.InterceptModule, client *client.Client) {
	slog.Info(r.Config.Action.Name, "text", "DEPLOYING DEFAULT MODULE AND SIDECAR PAIR")
	myModule.ClearURLs()

	r.prepareContainerNetwork(myModule, true)
	r.deployModule(myModule, client)
	r.deploySidecar(false, myModule, client)

	slog.Info(r.Config.Action.Name, "text", "WAITING FOR MODULE TO INITIALIZE")
	var waitMutex sync.WaitGroup
	waitMutex.Add(1)
	go r.Config.ModuleStep.PerformModuleHealthCheck(&waitMutex, myModule.ModuleName, myModule.BackendModule.ModuleExposedServerPort)
	waitMutex.Wait()
}

func (r *Run) deployCustomSidecarForInterception(printModuleEnv bool, myModule *models.InterceptModule, client *client.Client) {
	slog.Info(r.Config.Action.Name, "text", "DEPLOYING CUSTOM SIDECAR FOR INTERCEPTION")
	r.prepareContainerNetwork(myModule, false)
	r.deploySidecar(printModuleEnv, myModule, client)
}

func (r *Run) prepareContainerNetwork(myModule *models.InterceptModule, moduleAndSidecar bool) {
	myModule.NetworkConfig = helpers.NewModuleNetworkConfig()

	if moduleAndSidecar {
		moduleServerPort := helpers.SetFreePortFromRange(r.Config.Action)
		moduleDebugPort := helpers.SetFreePortFromRange(r.Config.Action)
		sidecarServerPort := helpers.SetFreePortFromRange(r.Config.Action)
		sidecarDebugPort := helpers.SetFreePortFromRange(r.Config.Action)

		myModule.BackendModule, myModule.RegistryModule = r.Config.ModuleStep.GetBackendModule(myModule.Containers, myModule.ModuleName)
		myModule.BackendModule.ModulePortBindings = helpers.CreatePortBindings(moduleServerPort, moduleDebugPort, myModule.BackendModule.ModuleServerPort)
		myModule.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, myModule.BackendModule.ModuleServerPort)
		myModule.BackendModule.ModuleExposedServerPort = moduleServerPort
		return
	}

	sidecarServerPort := helpers.ExtractPortFromURL(r.Config.Action, *myModule.SidecarUrl)
	sidecarDebugPort := helpers.SetFreePortFromRange(r.Config.Action)

	myModule.SidecarServerPort = sidecarServerPort

	myModule.BackendModule, myModule.RegistryModule = r.Config.ModuleStep.GetBackendModule(myModule.Containers, myModule.ModuleName)
	myModule.BackendModule.SidecarPortBindings = helpers.CreatePortBindings(sidecarServerPort, sidecarDebugPort, myModule.BackendModule.ModuleServerPort)
}

func (r *Run) deployModule(myModule *models.InterceptModule, client *client.Client) {
	moduleVersion := r.Config.ModuleStep.GetModuleImageVersion(*myModule.BackendModule, myModule.RegistryModule)
	moduleImage := r.Config.ModuleStep.GetModuleImage(moduleVersion, myModule.RegistryModule)
	moduleEnv := r.Config.ModuleStep.GetModuleEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule)
	moduleContainer := models.NewModuleContainer(myModule.RegistryModule.Name, moduleImage, moduleEnv, *myModule.BackendModule, myModule.NetworkConfig)
	r.Config.ModuleStep.DeployModule(client, moduleContainer)
}

func (r *Run) deploySidecar(printModuleEnv bool, myModule *models.InterceptModule, client *client.Client) {
	sidecarImage, pullSidecarImage := r.Config.ModuleStep.GetSidecarImage(myModule.Containers.RegistryModules[constant.EurekaRegistry])
	sidecarResources := helpers.CreateResources(false, viper.GetStringMap(field.SidecarModuleResources))
	sidecarEnv := r.Config.ModuleStep.GetSidecarEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule, myModule.ModuleUrl, myModule.SidecarUrl)
	sidecarContainer := models.NewSidecarContainer(myModule.RegistryModule.SidecarName, sidecarImage, sidecarEnv, *myModule.BackendModule, myModule.NetworkConfig, sidecarResources)
	sidecarContainer.PullImage = pullSidecarImage

	r.Config.ModuleStep.DeployModule(client, sidecarContainer)

	if printModuleEnv {
		moduleEnv := r.Config.ModuleStep.GetModuleEnv(myModule.Containers, myModule.RegistryModule, *myModule.BackendModule)

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
}

func init() {
	rootCmd.AddCommand(interceptModuleCmd)
	interceptModuleCmd.PersistentFlags().StringVarP(&rp.ID, "id", "i", "", "Module id, e.g. mod-orders:13.1.0-SNAPSHOT.1021 (required)")
	interceptModuleCmd.PersistentFlags().StringVarP(&rp.ModuleURL, "moduleUrl", "m", "", "Module URL, e.g. http://host.docker.internal:36002 or 36002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().StringVarP(&rp.SidecarURL, "sidecarUrl", "s", "", "Sidecar URL e.g. http://host.docker.internal:37002 or 37002 (if -g is used)")
	interceptModuleCmd.PersistentFlags().BoolVarP(&rp.Restore, "restore", "r", false, "Restore module & sidecar")
	interceptModuleCmd.PersistentFlags().BoolVarP(&rp.DefaultGateway, "defaultGateway", "g", false, "Use default gateway in URLs, .e.g http://host.docker.internal:{{port}} will be set automatically")
	if err := interceptModuleCmd.MarkPersistentFlagRequired("id"); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
