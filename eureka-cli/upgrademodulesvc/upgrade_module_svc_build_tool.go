package upgrademodulesvc

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/errors"
)

type buildTool int

const (
	mavenBuild buildTool = iota
	gradleBuild
	grailsBuild
)

func (t buildTool) String() string {
	switch t {
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
