package execplus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 源码来源于：github.com/codeskyblue/kexec
// 增加两个功能：
// 1、增加退出Terminate函数（通过kill进程来实现）
// 2、两次调用Wait函数不会出错，原来exec调用两次会报错。
type CmdPlus struct {
	*exec.Cmd
	cancelFunc context.CancelFunc
	errChs     []chan error
	err        error
	finished   bool
	once       sync.Once
	mu         sync.Mutex
}

// SetShellName 设置 Shell 名称
func SetShellName(name string) {
	shell_name = name
}

// ConvertByte2String 将字节转换为字符串
func ConvertByte2String(byte []byte, charset string) string {
	var str string
	switch charset {
	case "GB18030":
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case "UTF-8":
		fallthrough
	default:
		str = string(byte)
	}

	return str
}

// Terminate 终止进程
func (Self *CmdPlus) Terminate() {
	if Self.cancelFunc != nil {
		Self.cancelFunc()
	}
}

// Wait 等待进程结束
func (Self *CmdPlus) Wait() error {
	if Self.Process == nil {
		return errors.New("程序没有启动！")
	}
	Self.once.Do(func() {
		if Self.errChs == nil {
			Self.errChs = make([]chan error, 0)
		}
		go func() {
			Self.err = Self.Cmd.Wait()
			Self.mu.Lock()
			Self.finished = true
			for _, errC := range Self.errChs {
				errC <- Self.err
			}
			Self.mu.Unlock()
		}()
	})

	//如果进程已经结束，则不需要等待。
	Self.mu.Lock()
	if Self.finished {
		Self.mu.Unlock()
		return Self.err
	}

	//进程没有结束，则需要创建一个管道，加到errChs数组里。
	//这样once里定义的函数，在进程结束后会通知errChs数组里所有的管道。
	errCh := make(chan error, 1)
	Self.errChs = append(Self.errChs, errCh)
	Self.mu.Unlock()

	return <-errCh
}

// ShowConsole 显示控制台
func (Self *CmdPlus) ShowConsole(flag bool) {
	if flag {
		Self.Stdout = os.Stdout
		Self.Stderr = os.Stderr
	} else {
		Self.Stdout = nil
		Self.Stderr = nil
	}
}

// SetEnv 设置环境变量
func (Self *CmdPlus) SetEnv(key, value string) bool {
	find := false
	for i, kv := range Self.Env {
		eq := strings.Index(kv, "=")
		if eq < 0 {
			continue
		}
		k := kv[:eq]
		if runtime.GOOS == "windows" {
			find = strings.ToLower(k) == strings.ToLower(key)
		} else {
			find = k == key
		}
		if find {
			Self.Env[i] = fmt.Sprintf("%s=%s", key, value)
			break
		}
	}
	if !find {
		Self.Env = append(Self.Env, fmt.Sprintf("%s=%s", key, value))
	}
	return !find
}
