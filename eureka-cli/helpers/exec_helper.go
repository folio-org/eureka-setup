package helpers

import (
	"bytes"
	"os"
	"os/exec"
)

func Exec(c *exec.Cmd) error {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return err
	}

	return nil
}

func ExecIgnoreError(c *exec.Cmd) {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	_ = c.Run()
}

func ExecReturnOutput(c *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		return stdout, stderr, err
	}

	return stdout, stderr, nil
}

func ExecFromDir(c *exec.Cmd, workDir string) error {
	c.Dir = workDir

	err := Exec(c)
	if err != nil {
		return err
	}

	return nil
}
