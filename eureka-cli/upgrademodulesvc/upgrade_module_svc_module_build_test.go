package upgrademodulesvc

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	apperrors "github.com/folio-org/eureka-setup/eureka-cli/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createFile(t *testing.T, dir string, name string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte{}, 0o600))
}

func createDir(t *testing.T, dir string, name string) string {
	t.Helper()
	subDir := filepath.Join(dir, name)
	require.NoError(t, os.Mkdir(subDir, 0o750))
	return subDir
}

func TestDetectModuleBuild_Maven(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "pom.xml")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: mavenBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_Gradle(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: gradleBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_GradleKotlinDSL(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle.kts")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: gradleBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_Grails(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createDir(t, modulePath, "grails-app")
	createFile(t, modulePath, "build.gradle")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: grailsBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_GrailsInServiceSubdirectory(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	serviceDir := createDir(t, modulePath, "service")
	createDir(t, serviceDir, "grails-app")
	createFile(t, serviceDir, "build.gradle")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: grailsBuild, dir: serviceDir}, build)
}

func TestDetectModuleBuild_RootTakesPrecedenceOverServiceSubdirectory(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "pom.xml")
	serviceDir := createDir(t, modulePath, "service")
	createDir(t, serviceDir, "grails-app")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: mavenBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_GradleTakesPrecedenceOverMaven(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")
	createFile(t, modulePath, "pom.xml")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: gradleBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_GrailsAppAsFileIsNotGrails(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "grails-app")
	createFile(t, modulePath, "pom.xml")

	// Act
	build, err := detectModuleBuild(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, moduleBuild{tool: mavenBuild, dir: modulePath}, build)
}

func TestDetectModuleBuild_NoBuildFiles(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()

	// Act
	_, err := detectModuleBuild(modulePath)

	// Assert
	assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	assert.Contains(t, err.Error(), modulePath)
}

func TestBuildToolString_AllValues(t *testing.T) {
	// Assert
	assert.Equal(t, "maven", mavenBuild.String())
	assert.Equal(t, "gradle", gradleBuild.String())
	assert.Equal(t, "grails", grailsBuild.String())
}

func TestDescriptorPath_Maven(t *testing.T) {
	// Act
	descriptorPath := moduleBuild{tool: mavenBuild, dir: "module"}.descriptorPath()

	// Assert
	assert.Equal(t, filepath.Join("module", "target", "ModuleDescriptor.json"), descriptorPath)
}

func TestDescriptorPath_Gradle(t *testing.T) {
	// Act
	descriptorPath := moduleBuild{tool: gradleBuild, dir: "module"}.descriptorPath()

	// Assert
	assert.Equal(t, filepath.Join("module", "build", "resources", "main", "ModuleDescriptor.json"), descriptorPath)
}

func TestDescriptorPath_Grails(t *testing.T) {
	// Act
	descriptorPath := moduleBuild{tool: grailsBuild, dir: filepath.Join("module", "service")}.descriptorPath()

	// Assert
	assert.Equal(t, filepath.Join("module", "service", "build", "resources", "main", "okapi", "ModuleDescriptor.json"), descriptorPath)
}

func TestGradlewCommand_WrapsArgsWithWrapperScript(t *testing.T) {
	// Act
	cmd := gradlewCommand("build", "-x", "test")

	// Assert
	expectedGradlew := "./gradlew"
	if runtime.GOOS == "windows" {
		expectedGradlew = ".\\gradlew.bat"
	}
	assert.Equal(t, []string{expectedGradlew, "build", "-x", "test"}, cmd.Args)
}
