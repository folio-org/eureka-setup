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

// moduleBuild describes which build tool a module uses and from which directory it is built.
// The build directory is the module repository root or its service subdirectory, where Grails-based modules keep their build files;
// the Dockerfile used by BuildModuleImage always resides in the repository root.
type moduleBuild struct {
	tool buildTool
	dir  string
}

func (b moduleBuild) descriptorPath() string {
	switch b.tool {
	case grailsBuild:
		return filepath.Join(b.dir, "build", "resources", "main", "okapi", constant.ModuleDescriptor)
	case gradleBuild:
		return filepath.Join(b.dir, "build", "resources", "main", constant.ModuleDescriptor)
	default:
		return filepath.Join(b.dir, "target", constant.ModuleDescriptor)
	}
}

func detectModuleBuild(modulePath string) (moduleBuild, error) {
	for _, dir := range []string{modulePath, filepath.Join(modulePath, "service")} {
		if tool, found := detectBuildToolIn(dir); found {
			return moduleBuild{tool: tool, dir: dir}, nil
		}
	}

	return moduleBuild{}, errors.ModuleBuildToolNotFound(modulePath)
}

// Grails is detected via the grails-app directory, not grailsw: some Grails modules, e.g. mod-agreements, ship without the Grails wrapper
func detectBuildToolIn(dir string) (tool buildTool, found bool) {
	if fileInfo, err := os.Stat(filepath.Join(dir, "grails-app")); err == nil && fileInfo.IsDir() {
		return grailsBuild, true
	}
	for _, buildFile := range []string{"build.gradle", "build.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(dir, buildFile)); err == nil {
			return gradleBuild, true
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
		return mavenBuild, true
	}

	return mavenBuild, false
}

func mvnCommand(args ...string) *exec.Cmd {
	return exec.Command("mvn", args...)
}

func gradlewScriptName() string {
	if runtime.GOOS == "windows" {
		return "gradlew.bat"
	}
	return "gradlew"
}

func checkGradleWrapper(dir string) error {
	gradlewPath := filepath.Join(dir, gradlewScriptName())
	if _, err := os.Stat(gradlewPath); err != nil {
		return errors.GradleWrapperNotFound(gradlewPath)
	}

	return nil
}

func gradlewCommand(args ...string) *exec.Cmd {
	return exec.Command("."+string(filepath.Separator)+gradlewScriptName(), args...)
}
