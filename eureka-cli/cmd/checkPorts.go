/*
Copyright © 2024 EPAM_Systems/Thunderjet/Boburbek_Kadirkhodjaev

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
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const checkPortsCommand = "Undeploy Modules"

// checkPortsCmd represents the checkPorts command
var checkPortsCmd = &cobra.Command{
	Use:   "checkPorts",
	Short: "Check ports",
	Long:  `Check container ports.`,
	Run: func(cmd *cobra.Command, args []string) {
		CheckPorts()
	},
}

func CheckPorts() {
	slog.Info(checkPortsCommand, internal.GetFuncName(), "### CHECKING CONTAINER PORTS ###")
	client := internal.CreateDockerClient(checkPortsCommand)
	defer client.Close()

	filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf(internal.MultipleModulesContainerPattern, viper.GetString(internal.ProfileNameKey))})
	deployedModules := internal.GetDeployedModules(checkPortsCommand, client, filters)
	runNetcat(deployedModules)
}

func runNetcat(modules []types.Container) {
	slog.Info(checkPortsCommand, internal.GetFuncName(), "Running netcat -zv [container] [private port]")
	for _, module := range modules {
		moduleName := fmt.Sprintf("%s.eureka", strings.ReplaceAll(module.Names[0], "/", ""))
		for _, portPair := range module.Ports {
			modulePrivatePort := strconv.Itoa(int(portPair.PrivatePort))
			internal.RunCommandIgnoreError(listSystemCommand, exec.Command("docker", "exec", "-i", "netcat", "nc", "-zv", moduleName, modulePrivatePort))
		}
	}
}

func init() {
	rootCmd.AddCommand(checkPortsCmd)
}
