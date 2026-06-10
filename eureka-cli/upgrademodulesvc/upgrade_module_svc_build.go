package upgrademodulesvc

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
)

// UpgradeModuleBuildManager defines the interface for building operations to upgrade a module
type UpgradeModuleBuildManager interface {
	BuildModuleArtifact(moduleName, newModuleVersion, modulePath string) error
	CleanModuleArtifact(moduleName, modulePath string) error
	BuildModuleImage(namespace, moduleName, newModuleVersion, modulePath string) error
	ReadModuleDescriptor(moduleName, newModuleVersion, modulePath string) (newModuleDescriptor map[string]any, err error)
}

type buildTool int

const (
	mavenBuild buildTool = iota
	gradleBuild
	grailsBuild
)

func (tool buildTool) String() string {
	switch tool {
	case gradleBuild:
		return "gradle"
	case grailsBuild:
		return "grails"
	default:
		return "maven"
	}
}

// Grails is detected via the grails-app directory, not grailsw: some Grails modules, e.g. mod-agreements, ship without the Grails wrapper
func detectBuildTool(modulePath string) (buildTool, error) {
	if fileInfo, err := os.Stat(filepath.Join(modulePath, "grails-app")); err == nil && fileInfo.IsDir() {
		return grailsBuild, nil
	}
	for _, buildFile := range []string{"build.gradle", "build.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(modulePath, buildFile)); err == nil {
			return gradleBuild, nil
		}
	}
	if _, err := os.Stat(filepath.Join(modulePath, "pom.xml")); err == nil {
		return mavenBuild, nil
	}

	return mavenBuild, errors.ModuleBuildToolNotFound(modulePath)
}

func moduleDescriptorPath(modulePath string, tool buildTool) string {
	switch tool {
	case grailsBuild:
		return filepath.Join(modulePath, "build", "resources", "main", "okapi", constant.ModuleDescriptor)
	case gradleBuild:
		return filepath.Join(modulePath, "build", "resources", "main", constant.ModuleDescriptor)
	default:
		return filepath.Join(modulePath, "target", constant.ModuleDescriptor)
	}
}

func gradlewCommand(args ...string) *exec.Cmd {
	gradlew := "./gradlew"
	if runtime.GOOS == "windows" {
		gradlew = ".\\gradlew.bat"
	}

	return exec.Command(gradlew, args...)
}

func (um *UpgradeModuleSvc) BuildModuleArtifact(moduleName, newModuleVersion, modulePath string) error {
	slog.Info(um.Action.Name, "text", "BUILDING MODULE ARTIFACT", "module", moduleName, "version", newModuleVersion)
	tool, err := detectBuildTool(modulePath)
	if err != nil {
		return err
	}

	slog.Info(um.Action.Name, "text", "Detected build tool", "module", moduleName, "tool", tool.String())
	if tool == mavenBuild {
		return um.buildMavenArtifact(moduleName, newModuleVersion, modulePath)
	}
	return um.buildGradleArtifact(moduleName, newModuleVersion, modulePath, tool)
}

func (um *UpgradeModuleSvc) buildMavenArtifact(moduleName, newModuleVersion, modulePath string) error {
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

func (um *UpgradeModuleSvc) buildGradleArtifact(moduleName, newModuleVersion, modulePath string, tool buildTool) error {
	slog.Info(um.Action.Name, "text", "Cleaning build directory", "module", moduleName, "path", modulePath)
	if err := um.ExecSvc.ExecFromDir(gradlewCommand("clean"), modulePath); err != nil {
		return err
	}

	// Grails modules derive their version from the appVersion property instead
	versionFlag := fmt.Sprintf("-Pversion=%s", newModuleVersion)
	if tool == grailsBuild {
		versionFlag = fmt.Sprintf("-PappVersion=%s", newModuleVersion)
	}
	slog.Info(um.Action.Name, "text", "Packaging new artifact", "module", moduleName, "version", newModuleVersion)

	return um.ExecSvc.ExecFromDir(gradlewCommand("build", "-x", "test", versionFlag), modulePath)
}

func (um *UpgradeModuleSvc) CleanModuleArtifact(moduleName, modulePath string) error {
	slog.Info(um.Action.Name, "text", "CLEANING MODULE ARTIFACT", "module", moduleName)
	tool, err := detectBuildTool(modulePath)
	if err != nil {
		return err
	}

	if tool == mavenBuild {
		return um.cleanMavenArtifact(modulePath)
	}
	return um.cleanGradleArtifact(modulePath)
}

func (um *UpgradeModuleSvc) cleanMavenArtifact(modulePath string) error {
	if err := um.ExecSvc.ExecFromDir(exec.Command("mvn", "versions:revert"), modulePath); err != nil {
		return err
	}

	return um.ExecSvc.ExecFromDir(exec.Command("mvn", "clean", "package", "-DskipTests"), modulePath)
}

func (um *UpgradeModuleSvc) cleanGradleArtifact(modulePath string) error {
	return um.ExecSvc.ExecFromDir(gradlewCommand("clean"), modulePath)
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
	tool, err := detectBuildTool(modulePath)
	if err != nil {
		return nil, err
	}
	if err := helpers.ReadJSONFromFile(moduleDescriptorPath(modulePath, tool), &newModuleDescriptor); err != nil {
		return nil, err
	}
	if len(newModuleDescriptor) == 0 {
		slog.Info(um.Action.Name, "text", "New module descriptor was not found", "module", moduleName, "version", newModuleVersion)
		return nil, nil
	}

	return newModuleDescriptor, nil
}
