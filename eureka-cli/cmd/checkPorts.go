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
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// checkPortsCmd represents the checkPorts command
var checkPortsCmd = &cobra.Command{
	Use:   "checkPorts",
	Short: "Check ports",
	Long:  `Check container ports.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := New(action.CheckPorts)
		if err != nil {
			return err
		}

		return r.CheckPorts()
	},
}

func (r *Run) CheckPorts() error {
	slog.Info(r.RunConfig.Action.Name, "text", "CHECKING CONTAINER PORTS")
	err := r.deployNetcatContainer()
	if err != nil {
		return err
	}

	modules, err := r.getDeployedModules()
	if err != nil {
		return err
	}
	r.runNetcat(modules)

	return nil
}

func (r *Run) deployNetcatContainer() error {
	preparedCommand := exec.Command("docker", "compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach", "netcat")

	dir, err := helpers.GetHomeMiscDir(r.RunConfig.Action.Name)
	if err != nil {
		return err
	}

	return helpers.ExecFromDir(preparedCommand, dir)
}

func (r *Run) getDeployedModules() ([]container.Summary, error) {
	client, err := r.RunConfig.DockerClient.Create()
	if err != nil {
		return nil, err
	}
	defer r.RunConfig.DockerClient.Close(client)

	filters := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: fmt.Sprintf(constant.ProfileContainerPattern, r.RunConfig.Action.ConfigProfile),
	})
	containers, err := r.RunConfig.ModuleSvc.GetDeployedModules(client, filters)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (r *Run) runNetcat(modules []container.Summary) {
	slog.Info(r.RunConfig.Action.Name, "text", "Running netcat -zv [container] [private port]")
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
