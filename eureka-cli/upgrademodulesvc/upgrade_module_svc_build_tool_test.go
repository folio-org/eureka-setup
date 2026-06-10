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

func createDir(t *testing.T, dir string, name string) {
	t.Helper()
	require.NoError(t, os.Mkdir(filepath.Join(dir, name), 0o750))
}

func TestDetectBuildTool_Maven(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "pom.xml")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mavenBuild, tool)
}

func TestDetectBuildTool_Gradle(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, gradleBuild, tool)
}

func TestDetectBuildTool_GradleKotlinDSL(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle.kts")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, gradleBuild, tool)
}

func TestDetectBuildTool_Grails(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createDir(t, modulePath, "grails-app")
	createFile(t, modulePath, "build.gradle")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, grailsBuild, tool)
}

func TestDetectBuildTool_GradleTakesPrecedenceOverMaven(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")
	createFile(t, modulePath, "pom.xml")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, gradleBuild, tool)
}

func TestDetectBuildTool_GrailsAppAsFileIsNotGrails(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "grails-app")
	createFile(t, modulePath, "pom.xml")

	// Act
	tool, err := detectBuildTool(modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mavenBuild, tool)
}

func TestDetectBuildTool_NoBuildFiles(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()

	// Act
	_, err := detectBuildTool(modulePath)

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

func TestModuleDescriptorPath_Maven(t *testing.T) {
	// Act
	descriptorPath := moduleDescriptorPath("module", mavenBuild)

	// Assert
	assert.Equal(t, filepath.Join("module", "target", "ModuleDescriptor.json"), descriptorPath)
}

func TestModuleDescriptorPath_Gradle(t *testing.T) {
	// Act
	descriptorPath := moduleDescriptorPath("module", gradleBuild)

	// Assert
	assert.Equal(t, filepath.Join("module", "build", "resources", "main", "ModuleDescriptor.json"), descriptorPath)
}

func TestModuleDescriptorPath_Grails(t *testing.T) {
	// Act
	descriptorPath := moduleDescriptorPath("module", grailsBuild)

	// Assert
	assert.Equal(t, filepath.Join("module", "build", "resources", "main", "okapi", "ModuleDescriptor.json"), descriptorPath)
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
