//go:build !windows
// +build !windows

package execplus

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

var shell_name string = "/bin/bash"

func setupCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setsid = true
}

func CommandContext(ctx context.Context, name string, arg ...string) *CmdPlus {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	setupCmd(cmd)
	return &CmdPlus{
		Cmd: cmd,
	}
}

func Command(name string, arg ...string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	setupCmd(cmd)
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

func CommandStringContext(ctx context.Context, command string) *CmdPlus {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, shell_name, "-c", command)
	cmd.Env = os.Environ()
	setupCmd(cmd)
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

func CommandString(command string) *CmdPlus {
	return CommandStringContext(context.Background(), command)
}

// Ref: http://stackoverflow.com/questions/21705950/running-external-commands-through-os-exec-under-another-user
func (k *CmdPlus) SetUser(name string) (err error) {
	u, err := user.Lookup(name)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	if k.SysProcAttr == nil {
		k.SysProcAttr = &syscall.SysProcAttr{}
	}
	k.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	return nil
}

func (Self *CmdPlus) HideWindow() {
}
