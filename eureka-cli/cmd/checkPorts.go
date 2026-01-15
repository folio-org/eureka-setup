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
	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
)

// checkPortsCmd represents the checkPorts command
var checkPortsCmd = &cobra.Command{
	Use:   "checkPorts",
	Short: "Check ports",
	Long:  `Check container ports.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.CheckPorts)
		if err != nil {
			return err
		}

		return run.CheckPorts()
	},
}

func (run *Run) CheckPorts() error {
	slog.Info(run.Config.Action.Name, "text", "CHECKING CONTAINER PORTS")
	if err := run.deployNetcatContainer(); err != nil {
		return err
	}

	modules, err := run.getDeployedModules()
	if err != nil {
		return err
	}
	run.runNetcat(modules)

	return nil
}

func (run *Run) deployNetcatContainer() error {
	preparedCommand := exec.Command("docker", "compose", "--progress", "plain", "--ansi", "never", "--project-name", "eureka", "up", "--detach", "netcat")
	homeDir, err := helpers.GetHomeMiscDir()
	if err != nil {
		return err
	}

	return run.Config.ExecSvc.ExecFromDir(preparedCommand, homeDir)
}

func (run *Run) getDeployedModules() ([]container.Summary, error) {
	client, err := run.Config.DockerClient.Create()
	if err != nil {
		return nil, err
	}
	defer run.Config.DockerClient.Close(client)

	filters := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: fmt.Sprintf(constant.ProfileContainerPattern, run.Config.Action.ConfigProfileName),
	})
	containers, err := run.Config.ModuleSvc.GetDeployedModules(client, filters)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (run *Run) runNetcat(modules []container.Summary) {
	slog.Info(run.Config.Action.Name, "text", "Running netcat -zv [container] [private port]")
	for _, module := range modules {
		name := fmt.Sprintf("%s.eureka", strings.ReplaceAll(module.Names[0], "/", ""))
		for _, portPair := range module.Ports {
			privatePort := strconv.Itoa(int(portPair.PrivatePort))
			_ = run.Config.ExecSvc.Exec(exec.Command("docker", "exec", "-i", "netcat", "nc", "-zv", name, privatePort))
		}
	}
}

func init() {
	rootCmd.AddCommand(checkPortsCmd)
}
