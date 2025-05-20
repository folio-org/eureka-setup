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
	"os"
	"path"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

const rootCommand string = "Root"

var (
	withConfigFile        string
	withModuleName        string
	withEnableDebug       bool
	withBuildImages       bool
	withUpdateCloned      bool
	withEnableEcsRequests bool
	withTenant            string
	withNamespace         string
	withShowAll           bool
	withId                string
	withModuleUrl         string
	withSidecarUrl        string
	withRestore           bool
	withDefaultGateway    bool
	withRequired          bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eureka-cli",
	Short: "Eureka CLI",
	Long:  `Eureka CLI to deploy a local development environment.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&withConfigFile, "config", "c", "", fmt.Sprintf("Config file (default is $HOME/%s/%s.%s)", internal.ConfigDir, internal.ConfigCombined, internal.ConfigType))
	rootCmd.PersistentFlags().BoolVarP(&withEnableDebug, "debug", "d", false, "Enable debug")
}

func initConfig() {
	if withConfigFile != "" {
		viper.SetConfigFile(withConfigFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		configPath := path.Join(home, internal.ConfigDir)
		viper.AddConfigPath(configPath)
		viper.SetConfigType(internal.ConfigType)
		viper.SetConfigName(internal.ConfigCombined)
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error(rootCommand, internal.GetFuncName(), fmt.Sprintf("Cannot find or parse configuration file. Check that file exists and does not contain errors.: %s", withConfigFile))
		panic(err)
	}
}
