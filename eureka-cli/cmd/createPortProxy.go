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
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/field"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createPortProxyCmd represents the createPortProxy command
var createPortProxyCmd = &cobra.Command{
	Use:   "createPortProxy",
	Short: "Create port proxy",
	Long:  `Create a platform-specific port proxy to reroute module traffic.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		run, err := New(action.CreatePortProxy)
		if err != nil {
			return err
		}

		return run.CreatePortProxy()
	},
}

func (run *Run) CreatePortProxy() error {
	if runtime.GOOS == "windows" {
		if err := run.createPortProxyForWindows(); err != nil {
			return err
		}
	} else {
		slog.Warn(run.Config.Action.Name, "text", "Command not implemented for platform", "platform", runtime.GOOS)
	}

	return nil
}

func (run *Run) createPortProxyForWindows() error {
	var (
		sidecarHostname = fmt.Sprintf("%s.eureka", helpers.GetSidecarName(params.ModuleName))
		listenAddress   = fmt.Sprintf("listenaddress=%s", sidecarHostname)
		listenPort      = fmt.Sprintf("listenport=%d", params.PrivatePort)
		from            = fmt.Sprintf("%s:%d", sidecarHostname, params.PrivatePort)
	)
	if params.Restore {
		if err := run.showCurrentPortProxiesForWindows(); err != nil {
			return err
		}

		if err := run.execWithArgs("netsh", "interface", "portproxy", "delete", "v4tov4", listenAddress, listenPort); err != nil {
			slog.Error(run.Config.Action.Name, "text", "Failed to remove port proxy", "error", err)
			return err
		}
		slog.Info(run.Config.Action.Name, "text", "Deleted port proxy", "from", from)

		return nil
	}

	sidecarPort, err := helpers.GetPortFromURL(params.SidecarURL)
	if err != nil {
		return err
	}

	var (
		connectAddress = fmt.Sprintf("connectaddress=%s", params.GatewayHostname)
		connectPort    = fmt.Sprintf("connectport=%d", sidecarPort)
		to             = fmt.Sprintf("%s:%d", params.GatewayHostname, sidecarPort)
	)
	if err := run.execWithArgs("netsh", "interface", "portproxy", "add", "v4tov4", listenAddress, listenPort, connectAddress, connectPort); err != nil {
		slog.Error(run.Config.Action.Name, "text", "Failed to add port proxy", "error", err)
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Added a port proxy", "from", from, "to", to)

	if err := run.showCurrentPortProxiesForWindows(); err != nil {
		return err
	}

	if err := run.execWithArgs("tracert", sidecarHostname); err != nil {
		slog.Error(run.Config.Action.Name, "text", "Failed to trace hostname", "from", from, "error", err)
		return fmt.Errorf("%w: Check if hostname exists in /etc/hosts: %s", err, sidecarHostname)
	}
	slog.Info(run.Config.Action.Name, "text", "Traced hostname", "from", sidecarHostname)

	return nil
}

func (run *Run) showCurrentPortProxiesForWindows() error {
	if err := run.execWithArgs("netsh", "interface", "portproxy", "show", "all"); err != nil {
		slog.Error(run.Config.Action.Name, "text", "Failed to show port proxies", "error", err)
		return err
	}
	slog.Info(run.Config.Action.Name, "text", "Showed current port proxies")

	return nil
}

func (run *Run) execWithArgs(args ...string) error {
	if params.EnableDebug {
		fmt.Println(args)
	}

	stdout, stderr, err := run.Config.ExecSvc.ExecReturnOutput(exec.Command(args[0], args[1:]...))
	for _, b := range []bytes.Buffer{stdout, stderr} {
		if b.Len() > 0 {
			output := helpers.FilterEmptyLines(b.String())
			if output != "" {
				fmt.Println()
				fmt.Println(output)
				fmt.Println()
			}
		}
	}

	return err
}

func init() {
	rootCmd.AddCommand(createPortProxyCmd)
	createPortProxyCmd.PersistentFlags().StringVarP(&params.ModuleName, action.ModuleName.Long, action.ModuleName.Short, "", action.ModuleName.Description)
	createPortProxyCmd.PersistentFlags().StringVarP(&params.SidecarURL, action.SidecarURL.Long, action.SidecarURL.Short, "", action.SidecarURL.Description)
	createPortProxyCmd.PersistentFlags().IntVarP(&params.PrivatePort, action.PrivatePort.Long, action.PrivatePort.Short, 8081, action.PrivatePort.Description)
	createPortProxyCmd.PersistentFlags().StringVarP(&params.GatewayHostname, action.GatewayHostname.Long, action.GatewayHostname.Short, "host.docker.internal", action.GatewayHostname.Description)
	createPortProxyCmd.PersistentFlags().BoolVarP(&params.Restore, action.Restore.Long, action.Restore.Short, false, action.Restore.Description)

	if err := createPortProxyCmd.MarkPersistentFlagRequired(action.SidecarURL.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.SidecarURL, err).Error())
		os.Exit(1)
	}
	if err := createPortProxyCmd.MarkPersistentFlagRequired(action.ModuleName.Long); err != nil {
		slog.Error(errors.MarkFlagRequiredFailed(action.ModuleName, err).Error())
		os.Exit(1)
	}

	if err := createPortProxyCmd.RegisterFlagCompletionFunc(action.ModuleName.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return helpers.GetBackendModuleNames(viper.GetStringMap(field.BackendModules)), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
