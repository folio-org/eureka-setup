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
	gatewayURLTemplate, err := action.GetGatewayURLTemplate(name)
	if err != nil {
		return nil, err
	}

	action := action.New(name, gatewayURLTemplate, &actionParams)
	return &Run{Config: runconfig.New(action, logger)}, nil
}

func (run *Run) PingGateway() error {
	gatewayURL := fmt.Sprintf(run.Config.Action.GatewayURLTemplate, constant.KongPort)
	return run.Config.HTTPClient.Ping(gatewayURL)
}

func setConfig(params *actionparams.ActionParams) {
	if params.ConfigFile == "" {
		setConfigNameByProfile(params.Profile)
		return
	}

	viper.SetConfigFile(params.ConfigFile)
}

func setDefaultLogger() *slog.Logger {
	logLevel := slog.LevelInfo
	if actionParams.EnableDebug {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}))

	slog.SetDefault(logger)

	return logger
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
	}

	return fmt.Sprintf("config.%s", profile)
}

func createHomeDir(overwriteFiles bool, embeddedFs *embed.FS) {
	homeConfigDir, err := helpers.GetHomeDirPath()
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

func (run *Run) ConsortiumPartition(fn func(string, constant.TenantType) error) error {
	if !action.IsSet(field.Consortiums) {
		return fn(constant.NoneConsortium, constant.Default)
	}

	for consortiumName := range run.Config.Action.ConfigConsortiums {
		for _, tenantType := range constant.GetTenantTypes() {
			slog.Info(run.Config.Action.Name, "text", "Running partition with consortium and tenant type", "consortium", consortiumName, "tenantType", tenantType)
			if err := fn(consortiumName, tenantType); err != nil {
				return err
			}
		}
	}

	return nil
}

func (run *Run) TenantPartition(consortiumName string, tenantType constant.TenantType, fn func(string, string) error) error {
	if err := run.GetVaultRootToken(); err != nil {
		return err
	}

	tenants, err := run.Config.ManagementSvc.GetTenants(consortiumName, tenantType)
	if err != nil {
		return err
	}

	for _, value := range tenants {
		entry := value.(map[string]any)
		configTenant := entry["name"].(string)
		if !helpers.HasTenant(configTenant, run.Config.Action.ConfigTenants) {
			continue
		}
		tenantType := entry["description"].(string)

		err = fn(configTenant, tenantType)
		if err != nil {
			return err
		}
	}

	return nil
}
