package internal

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
)

func RunCommand(commandName string, preparedCommand *exec.Cmd) {
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr

	if err := preparedCommand.Run(); err != nil {
		slog.Error(commandName, GetFuncName(), "systemCmd.Run() error")
		panic(err)
	}
}

func RunCommandIgnoreError(commandName string, preparedCommand *exec.Cmd) {
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr

	_ = preparedCommand.Run()
}

func RunCommandReturnOutput(commandName string, preparedCommand *exec.Cmd) (stdout, stderr bytes.Buffer) {
	preparedCommand.Stdout = &stdout
	preparedCommand.Stderr = &stderr

	if err := preparedCommand.Run(); err != nil {
		slog.Error(commandName, GetFuncName(), "systemCmd.Run() error")
		panic(err)
	}

	return stdout, stderr
}

func RunCommandFromDir(commandName string, preparedCommand *exec.Cmd, workDir string) {
	preparedCommand.Dir = workDir

	RunCommand(commandName, preparedCommand)
}
