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
	"io"
	"log/slog"
	"os"
	"path"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
)

const setupCommand = "Setup"

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup CLI",
	Long:  `Setup CLI config.`,
	Run: func(cmd *cobra.Command, args []string) {
		Setup()
	},
}

func Setup() {
	slog.Info(setupCommand, "### CREATING SETUP CONFIG ###", "")
	srcConfigFile := fmt.Sprintf("%s.%s", configMinimal, configType)
	sourceFileStat, err := os.Stat(srcConfigFile)
	if err != nil {
		slog.Error(setupCommand, "os.Stat error", "")
		panic(err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		internal.LogErrorPanic(setupCommand, "sourceFileStat.Mode().IsRegular error")
	}

	source, err := os.Open(srcConfigFile)
	if err != nil {
		internal.LogErrorPanic(setupCommand, "os.Open error")
	}
	defer source.Close()

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	dstConfigDir := path.Join(home, configDir)
	err = os.MkdirAll(dstConfigDir, 0700)
	if err != nil {
		slog.Error(setupCommand, "os.MkdirAll error", "")
		panic(err)
	}

	dstConfigFile := path.Join(dstConfigDir, srcConfigFile)
	destination, err := os.Create(dstConfigFile)
	if err != nil {
		slog.Error(setupCommand, "os.Create error", "")
		panic(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		slog.Error(setupCommand, "io.Copy error", "")
		panic(err)
	}

	slog.Info(setupCommand, fmt.Sprintf("Created setup in %s", dstConfigFile), "")
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
