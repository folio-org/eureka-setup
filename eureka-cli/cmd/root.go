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

const (
	rootCommand string = "Root"
	configDir   string = ".eureka"
)

var (
	configFile  string
	moduleName  string
	enableDebug bool
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", fmt.Sprintf("Config file (default is $HOME/%s/config.yaml)", configDir))
	rootCmd.PersistentFlags().BoolVarP(&enableDebug, "debug", "d", false, "Enable debug")
}

func initConfig() {
	slog.Info(rootCommand, "### READING CONFIG ###", "")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		configPath := path.Join(home, configDir)
		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		profile := viper.GetString(internal.ProfileNameKey)
		applicationsMap := viper.GetStringMap(internal.ApplicationKey)
		slog.Info(rootCommand, "Using config file", profile)
		slog.Info(rootCommand, "Using config profile", viper.ConfigFileUsed())
		slog.Info(rootCommand, "Using config application", fmt.Sprintf("%s-%s", applicationsMap["name"], applicationsMap["version"]))
	}
}
