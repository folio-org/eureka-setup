package uisvc

import (
	"log/slog"
	"path/filepath"

	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/j011195/eureka-setup/eureka-cli/models"
)

// UIPackageJSONProcessor defines the interface for UI package.json operations
type UIPackageJSONProcessor interface {
	PreparePackageJSON(configPath string) error
}

func (us *UISvc) PreparePackageJSON(configPath string) error {
	var packageJSON models.PackageJSON
	packageJSONPath := filepath.Join(configPath, "package.json")

	err := helpers.ReadJSONFromFile(packageJSONPath, &packageJSON)
	if err != nil {
		return err
	}
	packageJSON.Scripts["build"] = "export DEBUG=stripes*; export NODE_OPTIONS=\"--max-old-space-size=8000 $NODE_OPTIONS\"; stripes build stripes.config.js --languages en --sourcemap=false --no-minify"

	modules := []string{
		"@folio/authorization-policies",
		"@folio/authorization-roles",
		"@folio/plugin-select-application",
	}
	if !us.Action.Param.SingleTenant {
		modules = append(modules, "@folio/consortia-settings")
	}

	updates := 0
	for _, module := range modules {
		if packageJSON.Dependencies[module] == "" {
			packageJSON.Dependencies[module] = ">=1.0.0"
			updates++
		}
	}
	if updates > 0 {
		slog.Info(us.Action.Name, "text", "Added extra modules to package.json", "count", len(modules))
		err = helpers.WriteJSONToFile(packageJSONPath, packageJSON)
		if err != nil {
			return err
		}
	}

	return nil
}
