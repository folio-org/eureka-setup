package uistep

import (
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/helpers"
)

func (us *UIStep) PreparePackageJSON(configPath string, tenant string) error {
	var packageJSON struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		License         string            `json:"license"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Resolutions     map[string]string `json:"resolutions"`
	}

	packageJSONPath := fmt.Sprintf("%s/package.json", configPath)
	err := helpers.ReadJsonFromFile(us.Action, packageJSONPath, &packageJSON)
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
		slog.Info(us.Action.Name, "text", fmt.Sprintf("Added %d extra modules to package.json", len(modules)))
		err = helpers.WriteJsonToFile(us.Action, packageJSONPath, packageJSON)
		if err != nil {
			return err
		}
	}

	return nil
}
