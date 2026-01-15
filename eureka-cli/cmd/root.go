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
	"context"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runFs  *embed.FS
	logger *slog.Logger
	params action.Param
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "eureka-cli",
	Short:   "Eureka CLI",
	Long:    `Eureka CLI orchestrates the deployment of a local Eureka-based development environment.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate),
}

func Execute(fs *embed.FS) {
	runFs = fs
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func initConfig() {
	setConfig(&params)
	viper.AutomaticEnv()

	if params.OverwriteFiles {
		createHomeDir(true)
	} else {
		if err := viper.ReadInConfig(); err != nil {
			createHomeDir(false)
		}
	}

	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	logger, err = setDefaultLogger()
	cobra.CheckErr(err)
}

func setConfig(params *action.Param) {
	if params.ConfigFile == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, constant.ConfigDir))
		viper.SetConfigType(constant.ConfigType)
		if params.Profile == "" {
			params.Profile = constant.GetDefaultProfile()
		}
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			fmt.Println("Using profile:", params.Profile)
		}
		viper.SetConfigName(fmt.Sprintf("%s.%s", constant.ConfigPrefix, params.Profile))
	} else {
		viper.SetConfigFile(params.ConfigFile)
	}
}

func setDefaultLogger() (*slog.Logger, error) {
	logLevel := slog.LevelInfo
	if params.EnableDebug {
		logLevel = slog.LevelDebug
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(home, constant.ConfigDir, constant.LogDir)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	timestamp := time.Now().Format(constant.LogTimestampFormat)
	logFilePath := filepath.Join(logDir, fmt.Sprintf("%s-%s.log", params.Profile, timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Printf("Logging to: %s\n", logFilePath)
	}

	return logger, nil
}

func createHomeDir(overwriteFiles bool) {
	homeDir, err := helpers.GetHomeDirPath()
	cobra.CheckErr(err)

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		if overwriteFiles {
			fmt.Printf("Overwriting files in %s home directory\n\n", homeDir)
		} else {
			fmt.Printf("Creating missing files in %s home directory\n\n", homeDir)
		}
	}
	err = helpers.CopyMultipleFiles(homeDir, runFs)
	cobra.CheckErr(err)

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Println()
	}
}

func init() {
	profiles := constant.GetProfiles()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&params.Profile, action.Profile.Long, action.Profile.Short, "combined", fmt.Sprintf(action.Profile.Description, profiles))
	rootCmd.PersistentFlags().StringVarP(&params.ConfigFile, action.ConfigFile.Long, action.ConfigFile.Short, "", action.ConfigFile.Description)
	rootCmd.PersistentFlags().BoolVarP(&params.OverwriteFiles, action.OverwriteFiles.Long, action.OverwriteFiles.Short, false, fmt.Sprintf(action.OverwriteFiles.Description, constant.ConfigDir))
	rootCmd.PersistentFlags().BoolVarP(&params.EnableDebug, action.EnableDebug.Long, action.EnableDebug.Short, false, action.EnableDebug.Description)

	if err := rootCmd.RegisterFlagCompletionFunc(action.Profile.Long, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return profiles, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		slog.Error(errors.RegisterFlagCompletionFailed(err).Error())
		os.Exit(1)
	}
}
