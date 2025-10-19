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
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/folio-org/eureka-cli/runconfig"
	"github.com/folio-org/eureka-cli/runparams"
	"github.com/folio-org/eureka-cli/tenanttype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Run is a container that holds the RunConfig instance
type Run struct {
	Config *runconfig.RunConfig
}

func NewRun(name string) *Run {
	action := action.New(name)

	return &Run{
		Config: runconfig.New(action),
	}
}

func NewCustomRun(name string, startPort, endPort int) *Run {
	action := action.NewCustom(name, startPort, endPort)

	return &Run{
		Config: runconfig.New(action),
	}
}

func (r *Run) PartitionByConsortiumAndTenantType(callback func(string, tenanttype.TenantType)) {
	if viper.IsSet(field.Consortiums) {
		for consortiumName := range viper.GetStringMap(field.Consortiums) {
			for _, tenantType := range tenanttype.Get() {
				slog.Info(r.Config.Action.Name, "text", fmt.Sprintf("Running sequentially for %s consortium and %s tenant type", consortiumName, tenantType))
				callback(consortiumName, tenantType)
			}
		}

		return
	}

	callback(constant.NoneConsortium, tenanttype.Default)
}

func setDefaultLogger() {
	logLevel := slog.LevelInfo
	if rp.EnableDebug {
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

func setConfig(params *runparams.RunParams) {
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

	homeConfigDir := helpers.GetHomeDirPath(action)
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
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
