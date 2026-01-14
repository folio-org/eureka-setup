package upgrademodulesvc

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
)

// UpgradeModuleBuildManager defines the interface for building operations to upgrade a module
type UpgradeModuleBuildManager interface {
	BuildModuleArtifact(moduleName, newModuleVersion, modulePath string) error
	CleanModuleArtifact(moduleName, modulePath string) error
	BuildModuleImage(namespace, moduleName, newModuleVersion, modulePath string) error
	ReadModuleDescriptor(moduleName, newModuleVersion, modulePath string) (newModuleDescriptor map[string]any, err error)
}

func (um *UpgradeModuleSvc) BuildModuleArtifact(moduleName, newModuleVersion, modulePath string) error {
	slog.Info(um.Action.Name, "text", "BUILDING MODULE ARTIFACT", "module", moduleName, "version", newModuleVersion)
	slog.Info(um.Action.Name, "text", "Cleaning target directory", "module", moduleName, "path", modulePath)
	if err := um.ExecSvc.ExecFromDir(exec.Command("mvn", "clean", "-DskipTests"), modulePath); err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "Setting new artifact version", "module", moduleName, "version", newModuleVersion)
	if err := um.ExecSvc.ExecFromDir(exec.Command("mvn", "versions:set", fmt.Sprintf("-DnewVersion=%s", newModuleVersion)), modulePath); err != nil {
		return err
	}
	slog.Info(um.Action.Name, "text", "Packaging new artifact", "module", moduleName, "version", newModuleVersion)

	return um.ExecSvc.ExecFromDir(exec.Command("mvn", "package", "-DskipTests"), modulePath)
}

func (um *UpgradeModuleSvc) CleanModuleArtifact(moduleName, modulePath string) error {
	slog.Info(um.Action.Name, "text", "CLEANING MODULE ARTIFACT", "module", moduleName)
	if err := um.ExecSvc.ExecFromDir(exec.Command("mvn", "versions:revert"), modulePath); err != nil {
		return err
	}

	return um.ExecSvc.ExecFromDir(exec.Command("mvn", "clean", "package", "-DskipTests"), modulePath)
}

func (um *UpgradeModuleSvc) BuildModuleImage(namespace, moduleName, newModuleVersion, modulePath string) error {
	imageName := fmt.Sprintf("%s/%s:%s", namespace, moduleName, newModuleVersion)
	slog.Info(um.Action.Name, "text", "BUILDING MODULE IMAGE", "module", moduleName, "image", imageName)
	return um.ExecSvc.ExecFromDir(exec.Command("docker", "build", "--tag", imageName,
		"--file", "./Dockerfile",
		"--progress", "plain",
		"--no-cache",
		".",
	), modulePath)
}

func (um *UpgradeModuleSvc) ReadModuleDescriptor(moduleName, newModuleVersion, modulePath string) (newModuleDescriptor map[string]any, err error) {
	slog.Info(um.Action.Name, "text", "READING NEW MODULE DESCRIPTOR", "module", moduleName, "path", modulePath)
	descriptorPath := filepath.Join(modulePath, "target", constant.ModuleDescriptor)
	if err := helpers.ReadJSONFromFile(descriptorPath, &newModuleDescriptor); err != nil {
		return nil, err
	}
	if len(newModuleDescriptor) == 0 {
		slog.Info(um.Action.Name, "text", "New module descriptor was not found", "module", moduleName, "version", newModuleVersion)
		return nil, nil
	}

	return newModuleDescriptor, nil
}
