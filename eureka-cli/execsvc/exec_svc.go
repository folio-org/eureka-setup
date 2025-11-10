package execsvc

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/folio-org/eureka-cli/action"
)

// CommandRunner defines the interface for executing system commands
type CommandRunner interface {
	Exec(cmd *exec.Cmd) error
	ExecIgnoreError(cmd *exec.Cmd)
	ExecReturnOutput(cmd *exec.Cmd) (stdout, stderr bytes.Buffer, err error)
	ExecFromDir(cmd *exec.Cmd, workDir string) error
}

// ExecSvc implements CommandRunner for production use
type ExecSvc struct {
	Action *action.Action
}

// New creates a new ExecSvc instance
func New(action *action.Action) *ExecSvc {
	return &ExecSvc{Action: action}
}

func (es *ExecSvc) Exec(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (es *ExecSvc) ExecIgnoreError(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func (es *ExecSvc) ExecReturnOutput(cmd *exec.Cmd) (bytes.Buffer, bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}

func (es *ExecSvc) ExecFromDir(cmd *exec.Cmd, workDir string) error {
	cmd.Dir = workDir
	return es.Exec(cmd)
}
