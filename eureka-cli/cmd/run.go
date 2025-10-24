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
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/runconfig"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run is a container that holds the RunConfig instance
type Run struct {
	Config *runconfig.RunConfig
}

func New(name string) (*Run, error) {
	gatewayURL, err := helpers.GetGatewayURL(name)
	if err != nil {
		return nil, err
	}

	action := action.New(name, gatewayURL, &ap)
	return &Run{
		Config: runconfig.New(action),
	}, nil
}

func NewCustom(name string, startPort, endPort int) (*Run, error) {
	gatewayURL, err := helpers.GetGatewayURL(name)
	if err != nil {
		return nil, err
	}

	action := action.NewCustom(name, gatewayURL, startPort, endPort, &ap)
	return &Run{
		Config: runconfig.New(action),
	}, nil
}

func (r *Run) Partition(callback func(string, constant.TenantType)) {
	if viper.IsSet(field.Consortiums) {
		for consortiumName := range viper.GetStringMap(field.Consortiums) {
			for _, tenantType := range constant.Get() {
				slog.Info(r.Config.Action.Name, "text", "Running partition with consortium and tenant type", "consortium", consortiumName, "tenantType", tenantType)
				callback(consortiumName, tenantType)
			}
		}
		return
	}

	callback(constant.NoneConsortium, constant.Default)
}

func (r *Run) PartitionErr(callback func(string, constant.TenantType) error) error {
	if viper.IsSet(field.Consortiums) {
		for consortiumName := range viper.GetStringMap(field.Consortiums) {
			for _, tenantType := range constant.Get() {
				slog.Info(r.Config.Action.Name, "text", "Running partition with consortium and tenant type", "consortium", consortiumName, "tenantType", tenantType)
				err := callback(consortiumName, tenantType)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	err := callback(constant.NoneConsortium, constant.Default)
	if err != nil {
		return err
	}

	return nil
}

func setDefaultLogger() {
	logLevel := slog.LevelInfo
	if ap.EnableDebug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	})))
}

func tryReadInConfig(afterReadCallback func(configErr error)) {
	if err := viper.ReadInConfig(); err != nil {
		afterReadCallback(err)
	}
}

func setConfig(params *actionparams.ActionParams) {
	if params.ConfigFile == "" {
		setConfigNameByProfile(params.Profile)
		return
	}

	viper.SetConfigFile(params.ConfigFile)
}

func setConfigNameByProfile(profile string) {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	configPath := filepath.Join(home, constant.ConfigDir)
	viper.AddConfigPath(configPath)
	viper.SetConfigType(constant.ConfigType)

	configFile := getConfigFileByProfile(profile)
	viper.SetConfigName(configFile)
}

func getConfigFileByProfile(profile string) string {
	if profile == "" {
		profile = constant.GetDefaultProfile()
	}

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Println("Using profile:", profile)
		fmt.Println()
	}

	return fmt.Sprintf("config.%s", profile)
}

func createHomeDir(overwriteFiles bool, embeddedFs embed.FS) {
	action := &action.Action{Name: action.Root}

	homeConfigDir, err := helpers.GetHomeDirPath(action)
	cobra.CheckErr(err)

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		if overwriteFiles {
			fmt.Printf("Overwriting files in %s home directory\n\n", homeConfigDir)
		} else {
			fmt.Printf("Creating missing files in %s home directory\n\n", homeConfigDir)
		}
	}

	err = fs.WalkDir(embeddedFs, ".", func(path string, d fs.DirEntry, err error) error {
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

			if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
				fmt.Println("Created file:", dstPath)
			}
		}

		return nil
	})
	cobra.CheckErr(err)
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		fmt.Println()
	}
}
