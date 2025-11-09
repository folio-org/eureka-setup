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
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	embedFS      *embed.FS
	logger       *slog.Logger
	actionParams actionparams.ActionParams
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eureka-cli",
	Short: "Eureka CLI",
	Long:  `Eureka CLI orchestrates the deployment of a local Eureka-based development environment.`,
}

func Execute(mainEmbedFS *embed.FS) {
	embedFS = mainEmbedFS

	err := rootCmd.Execute()
	if err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func initConfig() {
	setConfig(&actionParams)
	logger = setDefaultLogger()
	viper.AutomaticEnv()

	if actionParams.OverwriteFiles {
		createHomeDir(true, embedFS)
	} else {
		if err := viper.ReadInConfig(); err != nil {
			createHomeDir(false, embedFS)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}
}

func init() {
	profiles := constant.GetProfiles()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&actionParams.Profile, "profile", "p", "combined", fmt.Sprintf("Use a specific profile, options: %s", profiles))
	if err := rootCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return profiles, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error("failed to register profile flag completion function", "error", err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringVarP(&actionParams.ConfigFile, "configFile", "c", "", "Use a specific config file")
	rootCmd.PersistentFlags().BoolVarP(&actionParams.OverwriteFiles, "overwriteFiles", "o", false, fmt.Sprintf("Overwrite files in %s home directory", constant.ConfigDir))
	rootCmd.PersistentFlags().BoolVarP(&actionParams.EnableDebug, "enableDebug", "d", false, "Enable debug")
}
