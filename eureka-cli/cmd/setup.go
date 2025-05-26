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
	slog.Info(setupCommand, internal.GetFuncName(), "### CREATING SETUP CLI CONFIG IN HOME DIR ###")

	srcPath := getCurrentLocalConfig()
	dstPath := path.Join(createHomeConfigDirIfNeeded(), srcPath)
	internal.CopySingleFile(setupCommand, srcPath, dstPath)
}

func getCurrentLocalConfig() string {
	if withConfigFile == "" {
		return fmt.Sprintf("%s.%s", internal.ConfigCombined, internal.ConfigType)
	}
	return withConfigFile
}

func createHomeConfigDirIfNeeded() string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	dstConfigDir := path.Join(home, internal.ConfigDir)
	if err = os.MkdirAll(dstConfigDir, 0700); err != nil {
		slog.Error(setupCommand, internal.GetFuncName(), "os.MkdirAll error")
		panic(err)
	}

	return dstConfigDir
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
