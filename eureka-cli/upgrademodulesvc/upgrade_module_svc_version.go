package upgrademodulesvc

import (
	"fmt"
	"log/slog"

	"github.com/Masterminds/semver/v3"
	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
)

// UpgradeModuleVersionManager defines the interface for version-related operations during the module upgrade
type UpgradeModuleVersionManager interface {
	SetNewModuleVersionAndIDIntoContext() error
	SetDefaultNamespaceIntoContext()
}

func (um *UpgradeModuleSvc) SetNewModuleVersionAndIDIntoContext() error {
	slog.Info(um.Action.Name, "text", "SETTING NEW MODULE VERSION AND ID", "module", um.Action.Param.ModuleName)
	slog.Info(um.Action.Name, "text", "Old id", "id", um.Action.Param.ID)
	if um.Action.Param.ModuleVersion == "" {
		oldModuleVersion, err := semver.NewVersion(helpers.GetModuleVersionFromID(um.Action.Param.ID))
		if err != nil {
			return err
		}
		if helpers.IsSnapshot(oldModuleVersion.String()) {
			um.Action.Param.ModuleVersion, err = helpers.IncrementSnapshotVersion(oldModuleVersion.String())
			if err != nil {
				return err
			}
		} else {
			um.Action.Param.ModuleVersion = oldModuleVersion.IncPatch().String()
		}
	}
	um.Action.Param.ID = fmt.Sprintf("%s-%s", um.Action.Param.ModuleName, um.Action.Param.ModuleVersion)

	slog.Info(um.Action.Name, "text", "New id", "newId", um.Action.Param.ID)

	return nil
}

func (um *UpgradeModuleSvc) SetDefaultNamespaceIntoContext() {
	if um.Action.Param.Namespace == "" {
		um.Action.Param.Namespace = constant.LocalNamespace
	}
}
