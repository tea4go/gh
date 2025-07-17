package execplus

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

var shell_name string = "cmd"

func CommandContext(ctx context.Context, name string, arg ...string) *CmdPlus {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd: cmd,
	}
}

func Command(name string, arg ...string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

func CommandStringContext(ctx context.Context, command string) *CmdPlus {
	cmd := exec.CommandContext(ctx, shell_name, "/c", command)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd: cmd,
	}
}

func CommandString(command string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, shell_name, "/c", command)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

// SetUser not support on windws
func (Self *CmdPlus) SetUser(name string) (err error) {
	return nil
}

func (Self *CmdPlus) HideWindow() {
	// 在windows下不显示cmd窗口
	if Self.Cmd.SysProcAttr == nil {
		Self.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	Self.Cmd.SysProcAttr.HideWindow = true
}
