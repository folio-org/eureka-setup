package helpers

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

func Exec(preparedCommand *exec.Cmd) error {
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr
	if err := preparedCommand.Run(); err != nil {
		return err
	}

	return nil
}

func ExecIgnoreError(preparedCommand *exec.Cmd) {
	preparedCommand.Stdout = os.Stdout
	preparedCommand.Stderr = os.Stderr
	_ = preparedCommand.Run()
}

func ExecReturnOutput(preparedCommand *exec.Cmd) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	preparedCommand.Stdout = &stdout
	preparedCommand.Stderr = &stderr
	if err := preparedCommand.Run(); err != nil {
		return stdout, stderr, err
	}

	return stdout, stderr, nil
}

func ExecFromDir(preparedCommand *exec.Cmd, workDir string) error {
	preparedCommand.Dir = workDir
	err := Exec(preparedCommand)
	if err != nil {
		return err
	}

	return nil
}

func LogCompletion(actionName string, start time.Time) {
	duration := time.Since(start)
	slog.Info(actionName, "text", "Command completed", "duration", duration)
}
