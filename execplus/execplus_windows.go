package execplus

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

var shell_name string = "cmd"

// CommandContext 创建命令上下文
func CommandContext(ctx context.Context, name string, arg ...string) *CmdPlus {
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd: cmd,
	}
}

// Command 创建命令
func Command(name string, arg ...string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

// CommandStringContext 创建字符串命令上下文
func CommandStringContext(ctx context.Context, command string) *CmdPlus {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, shell_name, "/c", command)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

// CommandString 创建字符串命令
func CommandString(command string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, shell_name, "/c", command)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

// PowerShellFile 执行 PowerShell 文件
func PowerShellFile(psexe, psname string) *CmdPlus {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, psexe, "-ExecutionPolicy", "Bypass", "-File", psname)
	cmd.Env = os.Environ()
	return &CmdPlus{
		Cmd:        cmd,
		cancelFunc: cancel,
	}
}

// SetUser not support on windws
// SetUser 设置用户
func (Self *CmdPlus) SetUser(name string) (err error) {
	return nil
}

// HideWindow 隐藏窗口
func (Self *CmdPlus) HideWindow() {
	// 在windows下不显示cmd窗口
	if Self.Cmd.SysProcAttr == nil {
		Self.Cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	Self.Cmd.SysProcAttr.HideWindow = true
}
