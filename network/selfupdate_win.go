//go:build windows

package network

import (
	"os/exec"
	"syscall"
)

func SetProcAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
}
