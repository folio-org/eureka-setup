package helpers

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
	}{
		{
			name:        "successful command",
			command:     getEchoCommand(),
			args:        getEchoArgs("hello"),
			expectError: false,
		},
		{
			name:        "failing command",
			command:     getNonExistentCommand(),
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args...)
			err := Exec(cmd)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestExecIgnoreError(t *testing.T) {
	tests := []struct {
		name    string
		command string
		args    []string
	}{
		{
			name:    "successful command - should not panic",
			command: getEchoCommand(),
			args:    getEchoArgs("test"),
		},
		{
			name:    "failing command - should not panic",
			command: getNonExistentCommand(),
			args:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args...)
			
			// This should not panic regardless of command success/failure
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ExecIgnoreError panicked: %v", r)
				}
			}()
			
			ExecIgnoreError(cmd)
		})
	}
}

func TestExecReturnOutput(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "successful command with output",
			command:        getEchoCommand(),
			args:           getEchoArgs("hello world"),
			expectError:    false,
			expectedOutput: "hello world",
		},
		{
			name:        "failing command",
			command:     getNonExistentCommand(),
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args...)
			stdout, stderr, err := ExecReturnOutput(cmd)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				if tt.expectedOutput != "" {
					output := strings.TrimSpace(stdout.String())
					if !strings.Contains(output, tt.expectedOutput) {
						t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
					}
				}
			}

			// Check that stdout and stderr are proper Buffer types
			if stdout.String() == "" && stderr.String() == "" && !tt.expectError {
				// For successful commands, we should get some output
				t.Log("No output captured, but command succeeded")
			}
		})
	}
}

func TestExecFromDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		command     string
		args        []string
		workDir     string
		expectError bool
	}{
		{
			name:        "successful command in specific directory",
			command:     getPwdCommand(),
			args:        []string{},
			workDir:     tempDir,
			expectError: false,
		},
		{
			name:        "failing command in specific directory",
			command:     getNonExistentCommand(),
			args:        []string{},
			workDir:     tempDir,
			expectError: true,
		},
		{
			name:        "command in non-existent directory",
			command:     getEchoCommand(),
			args:        getEchoArgs("test"),
			workDir:     "/non/existent/directory",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(tt.command, tt.args...)
			err := ExecFromDir(cmd, tt.workDir)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				// Verify the working directory was set
				if cmd.Dir != tt.workDir {
					t.Errorf("Working directory was not set correctly: got %q, want %q", cmd.Dir, tt.workDir)
				}
			}
		})
	}
}

func TestLogCompletion(t *testing.T) {
	// Test that LogCompletion doesn't panic and handles different durations
	tests := []struct {
		name       string
		actionName string
		delay      time.Duration
	}{
		{
			name:       "immediate completion",
			actionName: "test-action",
			delay:      0,
		},
		{
			name:       "delayed completion",
			actionName: "slow-action",
			delay:      10 * time.Millisecond,
		},
		{
			name:       "empty action name",
			actionName: "",
			delay:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			if tt.delay > 0 {
				time.Sleep(tt.delay)
			}

			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("LogCompletion panicked: %v", r)
				}
			}()

			LogCompletion(tt.actionName, start)

			// Verify that some time has passed
			duration := time.Since(start)
			if duration < tt.delay {
				t.Errorf("Duration should be at least %v, got %v", tt.delay, duration)
			}
		})
	}
}

// Helper functions for cross-platform command testing
func getEchoCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "echo"
}

func getEchoArgs(message string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", "echo", message}
	}
	return []string{message}
}

func getPwdCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "pwd"
}

func getPwdArgs() []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", "cd"}
	}
	return []string{}
}

func getNonExistentCommand() string {
	return "definitely-not-a-real-command-" + fmt.Sprint(time.Now().UnixNano())
}