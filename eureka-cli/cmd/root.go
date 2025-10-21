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

	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rf embed.FS
	rp runparams.RunParams
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eureka-cli",
	Short: "Eureka CLI",
	Long:  `Eureka CLI to deploy a local development environment.`,
}

func Execute(embedFS embed.FS) {
	rf = embedFS

	err := rootCmd.Execute()
	if err != nil {
		slog.Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func initConfig() {
	setConfig(&rp)
	setDefaultLogger()
	viper.AutomaticEnv()
	if rp.OverwriteFiles {
		createHomeDir(true, rf)
		tryReadInConfig(func(configErr error) { cobra.CheckErr(configErr) })
	} else {
		tryReadInConfig(func(configErr error) { createHomeDir(false, rf) })
		tryReadInConfig(func(configErr error) { cobra.CheckErr(configErr) })
	}
}

func init() {
	profiles := constant.GetProfiles()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&rp.Profile, "profile", "p", "combined", fmt.Sprintf("Use a specific profile, options: %s", profiles))
	if err := rootCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return profiles, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error("failed to register profile flag completion function", "error", err)
		os.Exit(1)
	}
	rootCmd.PersistentFlags().StringVarP(&rp.ConfigFile, "configFile", "c", "", "Use a specific config file")
	rootCmd.PersistentFlags().BoolVarP(&rp.OverwriteFiles, "overwriteFiles", "o", false, fmt.Sprintf("Overwrite files in %s home directory", constant.ConfigDir))
	rootCmd.PersistentFlags().BoolVarP(&rp.EnableDebug, "enableDebug", "d", false, "Enable debug")
}
