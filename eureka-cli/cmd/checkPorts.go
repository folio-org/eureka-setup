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
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkPortsCmd represents the checkPorts command
var checkPortsCmd = &cobra.Command{
	Use:   "checkPorts",
	Short: "Check ports",
	Long:  `Check container ports.`,
	Run: func(cmd *cobra.Command, args []string) {
		NewRun(action.CheckPorts).CheckPorts()
	},
}

func (r *Run) CheckPorts() {
	slog.Info(r.Config.Action.Name, "text", "CHECKING CONTAINER PORTS")
	r.deployNetcatContainer()

	modules := r.getDeployedModules()
	r.runNetcat(modules)
}

func (r *Run) deployNetcatContainer() {
	preparedCommand := exec.Command("docker", "compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach", "netcat")
	helpers.ExecFromDir(preparedCommand, helpers.GetHomeMiscDir(r.Config.Action))
}

func (r *Run) getDeployedModules() []container.Summary {
	client := r.Config.DockerClient.Create()
	defer func() {
		_ = client.Close()
	}()

	filters := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: fmt.Sprintf(constant.ProfileContainerPattern, viper.GetString(field.ProfileName))})
	containers := r.Config.ModuleStep.GetDeployedModules(client, filters)

	return containers
}

func (r *Run) runNetcat(modules []container.Summary) {
	slog.Info(r.Config.Action.Name, "text", "Running netcat -zv [container] [private port]")
	for _, module := range modules {
		moduleName := fmt.Sprintf("%s.eureka", strings.ReplaceAll(module.Names[0], "/", ""))

		for _, portPair := range module.Ports {
			modulePrivatePort := strconv.Itoa(int(portPair.PrivatePort))
			helpers.ExecIgnoreError(exec.Command("docker", "exec", "-i", "netcat", "nc", "-zv", moduleName, modulePrivatePort))
		}
	}
}

func init() {
	rootCmd.AddCommand(checkPortsCmd)
}
