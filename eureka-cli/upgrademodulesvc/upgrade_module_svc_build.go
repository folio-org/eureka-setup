package upgrademodulesvc

import (
	"fmt"
	"log/slog"
	"os/exec"

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
	build, err := detectModuleBuild(modulePath)
	if err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "Detected build tool", "module", moduleName, "tool", build.tool.String(), "path", build.dir)
	if build.tool == mavenBuild {
		return um.buildMavenArtifact(moduleName, newModuleVersion, build.dir)
	}
	return um.buildGradleArtifact(moduleName, newModuleVersion, build)
}

func (um *UpgradeModuleSvc) buildMavenArtifact(moduleName, newModuleVersion, buildDir string) error {
	slog.Info(um.Action.Name, "text", "Cleaning target directory", "module", moduleName, "path", buildDir)
	if err := um.ExecSvc.ExecFromDir(mvnCommand("clean", "-DskipTests"), buildDir); err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "Setting new artifact version", "module", moduleName, "version", newModuleVersion)
	if err := um.ExecSvc.ExecFromDir(mvnCommand("versions:set", fmt.Sprintf("-DnewVersion=%s", newModuleVersion)), buildDir); err != nil {
		return err
	}
	slog.Info(um.Action.Name, "text", "Packaging new artifact", "module", moduleName, "version", newModuleVersion)

	return um.ExecSvc.ExecFromDir(mvnCommand("package", "-DskipTests"), buildDir)
}

func (um *UpgradeModuleSvc) buildGradleArtifact(moduleName, newModuleVersion string, build moduleBuild) error {
	if err := checkGradleWrapper(build.dir); err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "Cleaning build directory", "module", moduleName, "path", build.dir)
	if err := um.ExecSvc.ExecFromDir(gradlewCommand("clean"), build.dir); err != nil {
		return err
	}

	// Grails modules derive their version from the appVersion property instead
	versionFlag := fmt.Sprintf("-Pversion=%s", newModuleVersion)
	if build.tool == grailsBuild {
		versionFlag = fmt.Sprintf("-PappVersion=%s", newModuleVersion)
	}
	slog.Info(um.Action.Name, "text", "Packaging new artifact", "module", moduleName, "version", newModuleVersion)

	return um.ExecSvc.ExecFromDir(gradlewCommand("assemble", versionFlag), build.dir)
}

func (um *UpgradeModuleSvc) CleanModuleArtifact(moduleName, modulePath string) error {
	slog.Info(um.Action.Name, "text", "CLEANING MODULE ARTIFACT", "module", moduleName)
	build, err := detectModuleBuild(modulePath)
	if err != nil {
		return err
	}

	if build.tool == mavenBuild {
		return um.cleanMavenArtifact(build.dir)
	}
	return um.cleanGradleArtifact(build.dir)
}

func (um *UpgradeModuleSvc) cleanMavenArtifact(buildDir string) error {
	if err := um.ExecSvc.ExecFromDir(mvnCommand("versions:revert"), buildDir); err != nil {
		return err
	}

	return um.ExecSvc.ExecFromDir(mvnCommand("clean", "package", "-DskipTests"), buildDir)
}

func (um *UpgradeModuleSvc) cleanGradleArtifact(buildDir string) error {
	if err := checkGradleWrapper(buildDir); err != nil {
		return err
	}

	return um.ExecSvc.ExecFromDir(gradlewCommand("clean"), buildDir)
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
	build, err := detectModuleBuild(modulePath)
	if err != nil {
		return nil, err
	}

	descriptorPath := build.descriptorPath()
	slog.Info(um.Action.Name, "text", "READING NEW MODULE DESCRIPTOR", "module", moduleName, "path", descriptorPath)
	if err := helpers.ReadJSONFromFile(descriptorPath, &newModuleDescriptor); err != nil {
		return nil, err
	}
	if len(newModuleDescriptor) == 0 {
		slog.Info(um.Action.Name, "text", "New module descriptor was not found", "module", moduleName, "version", newModuleVersion)
		return nil, nil
	}

	return newModuleDescriptor, nil
}
