package internal

import (
	"log/slog"
	"os"
	"os/exec"
)

func RunCommand(commandName string, preparedCommand *exec.Cmd, composeFileDir string) {
	preparedCommand.Dir = composeFileDir
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr
	if err := preparedCommand.Run(); err != nil {
		slog.Error(commandName, "systemCmd.Run() error", "")
		panic(err)
	}
}
