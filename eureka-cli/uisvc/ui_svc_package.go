package uisvc

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/folio-org/eureka-setup/eureka-cli/models"
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

	var modulesToRemove []string
	if us.Action.Param.SingleTenant {
		modulesToRemove = append(modulesToRemove, "@folio/consortia-settings")
	}
	if !us.Action.Param.LinkedData {
		modulesToRemove = append(modulesToRemove, "@folio/ld-folio-wrapper")
	}

	removed := 0
	for _, mod := range modulesToRemove {
		if _, exists := packageJSON.Dependencies[mod]; exists {
			delete(packageJSON.Dependencies, mod)
			removed++
		}
	}
	if removed > 0 {
		slog.Info(us.Action.Name, "text", "Removed optional modules from package.json", "count", removed)
		if err = helpers.WriteJSONToFile(packageJSONPath, packageJSON); err != nil {
			return err
		}
	}

	dumpBytes, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("DUMPING package.json")
	fmt.Println(string(dumpBytes))
	fmt.Println()

	return nil
}
