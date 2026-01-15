package execsvc_test

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/execsvc"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExec_DateCommand tests executing the date command
func TestExec_DateCommand(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "date", "/T")
	} else {
		cmd = exec.Command("date")
	}

	// Act
	err := svc.Exec(cmd)

	// Assert
	assert.NoError(t, err)
}

// TestExec_EchoCommand tests executing a simple echo command
func TestExec_EchoCommand(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "echo", "test")
	} else {
		cmd = exec.Command("echo", "test")
	}

	// Act
	err := svc.Exec(cmd)

	// Assert
	assert.NoError(t, err)
}

// TestExec_InvalidCommand tests executing an invalid command
func TestExec_InvalidCommand(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	cmd := exec.Command("nonexistentcommand12345")

	// Act
	err := svc.Exec(cmd)

	// Assert
	assert.Error(t, err)
}

// TestExecReturnOutput_ListDirectory tests listing directory contents and capturing output
func TestExecReturnOutput_ListDirectory(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "dir")
	} else {
		cmd = exec.Command("ls", "-al")
	}

	// Act
	stdout, stderr, err := svc.ExecReturnOutput(cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String(), "stdout should contain directory listing")
	assert.Empty(t, stderr.String(), "stderr should be empty for successful command")
}

// TestExecReturnOutput_EchoCommand tests echo command with output capture
func TestExecReturnOutput_EchoCommand(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	expectedOutput := "Hello, World!"
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "echo", expectedOutput)
	} else {
		cmd = exec.Command("echo", expectedOutput)
	}

	// Act
	stdout, stderr, err := svc.ExecReturnOutput(cmd)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), expectedOutput)
	assert.Empty(t, stderr.String())
}

// TestExecReturnOutput_CommandWithError tests command that writes to stderr
func TestExecReturnOutput_CommandWithError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: try to list a non-existent directory
		cmd = exec.Command("cmd", "/C", "dir", "C:\\nonexistentdir12345")
	} else {
		// Unix: try to list a non-existent directory
		cmd = exec.Command("ls", "/nonexistentdir12345")
	}

	// Act
	stdout, stderr, err := svc.ExecReturnOutput(cmd)

	// Assert
	assert.Error(t, err)
	// Either stdout or stderr should have error message depending on the OS
	output := stdout.String() + stderr.String()
	assert.NotEmpty(t, output, "should have error output")
}

// TestExecFromDir_ValidDirectory tests executing command from a specific directory
func TestExecFromDir_ValidDirectory(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	// Use the system temp directory which should always exist
	tempDir := os.TempDir()
	require.DirExists(t, tempDir, "temp directory should exist")

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "cd")
	} else {
		cmd = exec.Command("pwd")
	}

	// Act
	err := svc.ExecFromDir(cmd, tempDir)

	// Assert
	assert.NoError(t, err)
}

// TestExecFromDir_InvalidDirectory tests executing command from a non-existent directory
func TestExecFromDir_InvalidDirectory(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	invalidDir := "/nonexistent/directory/path/12345"
	if runtime.GOOS == "windows" {
		invalidDir = "C:\\nonexistent\\directory\\path\\12345"
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "echo", "test")
	} else {
		cmd = exec.Command("echo", "test")
	}

	// Act
	err := svc.ExecFromDir(cmd, invalidDir)

	// Assert
	assert.Error(t, err)
}

// TestExecFromDir_WithOutput tests executing command from directory and capturing output
func TestExecFromDir_WithOutput(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := execsvc.New(action)

	tempDir := os.TempDir()
	require.DirExists(t, tempDir)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "cd")
	} else {
		cmd = exec.Command("pwd")
	}

	cmd.Dir = tempDir

	// Act
	stdout, stderr, err := svc.ExecReturnOutput(cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, stdout.String())

	// Normalize paths for comparison
	outputPath := strings.TrimSpace(stdout.String())
	if runtime.GOOS == "windows" {
		// Windows paths might differ in case and format
		outputPath = strings.ToLower(outputPath)
		expectedPath := strings.ToLower(tempDir)
		assert.Contains(t, outputPath, expectedPath)
	} else {
		assert.Contains(t, outputPath, tempDir)
	}
	assert.Empty(t, stderr.String())
}

// TestNew_CreatesInstance tests that New creates a valid instance
func TestNew_CreatesInstance(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()

	// Act
	svc := execsvc.New(action)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
}
