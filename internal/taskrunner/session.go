package taskrunner

import (
	"os/exec"

	"github.com/chalkan3/sloth-runner/internal/types"
)

// NewSharedSession creates and starts a new persistent shell session.
func NewSharedSession() (*types.SharedSession, error) {
	cmd := exec.Command("bash", "-i") // Interactive shell

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &types.SharedSession{
			Cmd:    cmd,
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		nil
}
