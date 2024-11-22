package internal

import (
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

func RunCommandFromDir(commandName string, preparedCommand *exec.Cmd, workDir string) {
	preparedCommand.Dir = workDir
	RunCommand(commandName, preparedCommand)
}
