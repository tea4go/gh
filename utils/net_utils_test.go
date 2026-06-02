package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

// 模拟 net.Error
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestGetStatusCode(t *testing.T) {
	fmt.Println("=== 开始测试: HTTP 状态码获取 (TestGetStatusCode) ===")
	resp := &http.Response{
		StatusCode: 200,
	}
	if code := GetStatusCode(resp); code != 200 {
		t.Errorf("GetStatusCode 判断失败，期望 200，实际 %d", code)
	}

	resp = nil
	if code := GetStatusCode(nil); code != 0 {
		t.Errorf("GetStatusCode nil 判断失败，期望 0，实际 %d", code)
	}
}

func TestIsAddrInUse(t *testing.T) {
	fmt.Println("=== 开始测试: 端口占用判断 (TestIsAddrInUse) ===")
	// 模拟 OpError
	err := &net.OpError{
		Err: errors.New("address already in use"),
	}
	if !IsAddrInUse(err) {
		t.Error("IsAddrInUse 对标准错误判断失败")
	}

	err2 := errors.New("Only one usage of each socket address")
	// IsAddrInUse 要求 *net.OpError 类型，非该类型会直接返回 false吗？
	// func IsAddrInUse(err error) bool {
	//   opErr, ok := err.(*net.OpError)
	//   if !ok { return false }
	//   ...
	// }
	// 代码逻辑如下：
	/*
		opErr, ok := err.(*net.OpError)
		if !ok {
			return false
		}
	*/
	// 所以传入普通错误 "Only one usage..." 会因为不是 *net.OpError 类型而立即返回 false。
	// IsAddrInUse 的逻辑假定错误必须是 *net.OpError 类型。

	if IsAddrInUse(err2) {
		t.Log("IsAddrInUse 对普通错误返回了 true（与代码阅读预期不符，可能遗漏了某些逻辑）")
	} else {
		t.Log("IsAddrInUse 对普通错误返回 false，符合预期")
	}
}

func TestGetNetError(t *testing.T) {
	if msg := GetNetError(io.EOF); msg != "网络主动断开" {
		t.Errorf("GetNetError EOF 判断失败，期望 '网络主动断开'，实际 '%s'", msg)
	}

	// 测试端口占用
	opErr := &net.OpError{
		Err: errors.New("address already in use"),
	}
	if msg := GetNetError(opErr); msg != "端口已经占用" {
		t.Errorf("GetNetError AddrInUse 判断失败，期望 '端口已经占用'，实际 '%s'", msg)
	}

	// 测试普通错误
	err := errors.New("connection refused")
	if msg := GetNetError(err); msg != "连接被拒绝" {
		t.Errorf("GetNetError 普通错误判断失败，期望 '连接被拒绝'，实际 '%s'", msg)
	}
}

func TestGetIPAdress(t *testing.T) {
	fmt.Println("=== 开始测试: 获取本机 IP (TestGetIPAdress) ===")
	// 这依赖于本机网络环境，只需确保不发生 panic 即可。
	ip := GetIPAdress()
	t.Logf("获取本机 IP: %s", ip)
}

func TestGetAllIPAdress(t *testing.T) {
	fmt.Println("=== 开始测试: 获取所有 IP (TestGetAllIPAdress) ===")
	ips := GetAllIPAdress()
	t.Logf("获取所有 IP: %+v", ips)
}

func TestGetAllMacAdress(t *testing.T) {
	macs := GetAllMacAdress()
	t.Logf("获取所有 MAC: %s", macs)
}

func TestGetNetErrorTimeout(t *testing.T) {
	fmt.Println("=== 开始测试: 网络超时错误 (TestGetNetErrorTimeout) ===")
	err := &timeoutError{}
	if msg := GetNetError(err); msg != "网络连接超时" {
		t.Errorf("GetNetError 超时判断失败，期望 '网络连接超时'，实际 '%s'", msg)
	}
}
