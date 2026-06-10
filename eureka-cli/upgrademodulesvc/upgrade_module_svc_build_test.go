package upgrademodulesvc

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func expectedGradlew() string {
	if runtime.GOOS == "windows" {
		return ".\\gradlew.bat"
	}
	return "./gradlew"
}

func newSvcWithRecordedCommands(t *testing.T, buildDir string) (*UpgradeModuleSvc, *[][]string) {
	t.Helper()
	commands := &[][]string{}
	mockExec := new(testhelpers.MockCommandExecutor)
	mockExec.On("ExecFromDir", mock.Anything, buildDir).Run(func(args mock.Arguments) {
		cmd := args.Get(0).(*exec.Cmd)
		*commands = append(*commands, cmd.Args)
	}).Return(nil)
	return &UpgradeModuleSvc{Action: testhelpers.NewMockAction(), ExecSvc: mockExec}, commands
}

func TestBuildModuleArtifact_GradleUsesVersionProperty(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")
	svc, commands := newSvcWithRecordedCommands(t, modulePath)

	// Act
	err := svc.BuildModuleArtifact("mod-test", "1.1.0", modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, [][]string{
		{expectedGradlew(), "clean"},
		{expectedGradlew(), "build", "-x", "test", "-Pversion=1.1.0"},
	}, *commands)
}

func TestBuildModuleArtifact_GrailsUsesAppVersionProperty(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createDir(t, modulePath, "grails-app")
	createFile(t, modulePath, "build.gradle")
	svc, commands := newSvcWithRecordedCommands(t, modulePath)

	// Act
	err := svc.BuildModuleArtifact("mod-test", "1.1.0", modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, [][]string{
		{expectedGradlew(), "clean"},
		{expectedGradlew(), "build", "-x", "test", "-PappVersion=1.1.0"},
	}, *commands)
}

func TestCleanModuleArtifact_GradleCleansWithoutRebuild(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "build.gradle")
	svc, commands := newSvcWithRecordedCommands(t, modulePath)

	// Act
	err := svc.CleanModuleArtifact("mod-test", modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, [][]string{
		{expectedGradlew(), "clean"},
	}, *commands)
}

func TestCleanModuleArtifact_MavenRevertsVersionAndRebuilds(t *testing.T) {
	// Arrange
	modulePath := t.TempDir()
	createFile(t, modulePath, "pom.xml")
	svc, commands := newSvcWithRecordedCommands(t, modulePath)

	// Act
	err := svc.CleanModuleArtifact("mod-test", modulePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, [][]string{
		{"mvn", "versions:revert"},
		{"mvn", "clean", "package", "-DskipTests"},
	}, *commands)
}
