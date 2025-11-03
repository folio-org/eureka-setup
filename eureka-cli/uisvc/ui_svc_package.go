package uisvc

import (
	"log/slog"
	"path/filepath"

	"github.com/folio-org/eureka-cli/helpers"
)

// UIPackageJSONProcessor defines the interface for UI package.json operations
type UIPackageJSONProcessor interface {
	PreparePackageJSON(configPath string) error
}

func (us *UISvc) PreparePackageJSON(configPath string) error {
	var packageJSON struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		License         string            `json:"license"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Resolutions     map[string]string `json:"resolutions"`
	}
	packageJSONPath := filepath.Join(configPath, "package.json")
	err := helpers.ReadJsonFromFile(us.Action.Name, packageJSONPath, &packageJSON)
	if err != nil {
		return err
	}

	packageJSON.Scripts["build"] = "export DEBUG=stripes*; export NODE_OPTIONS=\"--max-old-space-size=8000 $NODE_OPTIONS\"; stripes build stripes.config.js --languages en --sourcemap=false --no-minify"
	updates := 0
	modules := []string{
		"@folio/consortia-settings",
		"@folio/authorization-policies",
		"@folio/authorization-roles",
		"@folio/plugin-select-application",
	}
	for _, module := range modules {
		if packageJSON.Dependencies[module] == "" {
			packageJSON.Dependencies[module] = ">=1.0.0"
			updates++
		}
	}
	if updates > 0 {
		slog.Info(us.Action.Name, "text", "Added extra modules to package.json", "moduleCount", len(modules))
		err = helpers.WriteJsonToFile(packageJSONPath, packageJSON)
		if err != nil {
			return err
		}
	}

	return nil
}
