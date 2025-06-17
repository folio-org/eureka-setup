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
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const rootCommand string = "Root"

var (
	withConfigFile        string
	withProfile           string
	withOverwriteFiles    bool
	withModuleName        string
	withEnableDebug       bool
	withBuildImages       bool
	withUpdateCloned      bool
	withEnableEcsRequests bool
	withTenant            string
	withNamespace         string
	withAll               bool
	withId                string
	withModuleUrl         string
	withSidecarUrl        string
	withRestore           bool
	withDefaultGateway    bool
	withOnlyRequired      bool
	withUser              string
	withLength            int
	withModuleType        string
	withPurgeSchemas      bool
)

var embeddedFs embed.FS

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eureka-cli",
	Short: "Eureka CLI",
	Long:  `Eureka CLI to deploy a local development environment.`,
}

func Execute(mainEmbeddedFs embed.FS) {
	embeddedFs = mainEmbeddedFs
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	setConfig(withEnableDebug, withConfigFile, withProfile)

	viper.AutomaticEnv()

	if withOverwriteFiles {
		createHomeDir(withEnableDebug, true, embeddedFs)
		tryReadInConfig(func(configErr error) { cobra.CheckErr(configErr) })
	} else {
		tryReadInConfig(func(configErr error) { createHomeDir(withEnableDebug, false, embeddedFs) })
		tryReadInConfig(func(configErr error) { cobra.CheckErr(configErr) })
	}
}

func tryReadInConfig(afterReadCallback func(configErr error)) {
	if err := viper.ReadInConfig(); err != nil {
		afterReadCallback(err)
	}
}

func setConfig(enabledDebug bool, configFile string, profile string) {
	if configFile == "" {
		setConfigNameByProfile(enabledDebug, profile)
		return
	}

	viper.SetConfigFile(configFile)
}

func setConfigNameByProfile(enabledDebug bool, profile string) {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configPath := filepath.Join(home, internal.ConfigDir)
	viper.AddConfigPath(configPath)
	viper.SetConfigType(internal.ConfigType)

	configFile := getConfigFileByProfile(enabledDebug, profile)
	viper.SetConfigName(configFile)
}

func getConfigFileByProfile(enabledDebug bool, profile string) string {
	if profile == "" {
		profile = internal.AvailableProfiles[0]
	}

	if enabledDebug {
		fmt.Println("Using profile:", profile)
		fmt.Println()
	}

	return fmt.Sprintf("config.%s", profile)
}

func createHomeDir(enabledDebug bool, overwriteFiles bool, embeddedFs embed.FS) {
	homeConfigDir := internal.GetHomeDirPath(rootCommand)
	if enabledDebug {
		if overwriteFiles {
			fmt.Printf("Overwriting files in %s home directory\n\n", homeConfigDir)
		} else {
			fmt.Printf("Creating missing files in %s home directory\n\n", homeConfigDir)
		}
	}

	err := fs.WalkDir(embeddedFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(homeConfigDir, path)
		if d.IsDir() {
			err := os.MkdirAll(dstPath, 0755)
			if err != nil {
				return err
			}
		} else {
			content, err := fs.ReadFile(embeddedFs, path)
			if err != nil {
				return err
			}

			err = os.WriteFile(dstPath, content, 0644)
			if err != nil {
				return err
			}

			if enabledDebug {
				fmt.Println("Created file:", dstPath)
			}
		}

		return nil
	})
	cobra.CheckErr(err)

	if withEnableDebug {
		fmt.Println()
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&withProfile, "profile", "p", "combined", fmt.Sprintf("Use a specific profile, options: %s", internal.AvailableProfiles))
	rootCmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return internal.AvailableProfiles, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.PersistentFlags().StringVarP(&withConfigFile, "configFile", "c", "", "Use a specific config file")
	rootCmd.PersistentFlags().BoolVarP(&withOverwriteFiles, "overwriteFiles", "o", false, fmt.Sprintf("Overwrite files in %s home directory", internal.ConfigDir))
	rootCmd.PersistentFlags().BoolVarP(&withEnableDebug, "enableDebug", "d", false, "Enable debug")
}
