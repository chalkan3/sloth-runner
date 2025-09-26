//go:build !linux
// +build !linux

package main

import (
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// No-op on non-linux systems
}
