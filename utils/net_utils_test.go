package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
)

// 模拟 net.Error
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// 模拟临时错误（非超时）
type temporaryError struct{}

func (e *temporaryError) Error() string   { return "temporary error" }
func (e *temporaryError) Timeout() bool   { return false }
func (e *temporaryError) Temporary() bool { return true }

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

	// Test with redirect response (resp.Request.Response is set)
	redirectResp := &http.Response{
		StatusCode: 200,
		Request: &http.Request{
			Response: &http.Response{
				StatusCode: 301,
			},
		},
	}
	if code := GetStatusCode(redirectResp); code != 301 {
		t.Errorf("GetStatusCode redirect 判断失败，期望 301，实际 %d", code)
	}
}

func TestIsAddrInUse(t *testing.T) {
	fmt.Println("=== 开始测试: 端口占用判断 (TestIsAddrInUse) ===")
	// 模拟 OpError with "address already in use"
	err := &net.OpError{
		Err: errors.New("address already in use"),
	}
	if !IsAddrInUse(err) {
		t.Error("IsAddrInUse 对标准错误判断失败")
	}

	// OpError with "Only one usage of each"
	err2 := &net.OpError{
		Err: errors.New("Only one usage of each socket address"),
	}
	if !IsAddrInUse(err2) {
		t.Error("IsAddrInUse 对 Windows 错误判断失败")
	}

	// Non-OpError should return false
	err3 := errors.New("address already in use")
	if IsAddrInUse(err3) {
		t.Error("IsAddrInUse 对普通错误应返回 false")
	}

	// OpError with different error
	err4 := &net.OpError{
		Err: errors.New("some other error"),
	}
	if IsAddrInUse(err4) {
		t.Error("IsAddrInUse 对无关 OpError 应返回 false")
	}
}

func TestIsNetClose(t *testing.T) {
	fmt.Println("=== 开始测试: IsNetClose ===")
	tests := []struct {
		errMsg string
		want   bool
	}{
		{"use of closed network connection", true},
		{"connection was forcibly closed by the remote host", true},
		{"broken pipe", true},
		{"some other error", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsNetClose(errors.New(tt.errMsg))
		if got != tt.want {
			t.Errorf("IsNetClose(%q) = %v, want %v", tt.errMsg, got, tt.want)
		}
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

func TestGetNetErrorTimeout(t *testing.T) {
	fmt.Println("=== 开始测试: 网络超时错误 (TestGetNetErrorTimeout) ===")
	err := &timeoutError{}
	if msg := GetNetError(err); msg != "网络连接超时" {
		t.Errorf("GetNetError 超时判断失败，期望 '网络连接超时'，实际 '%s'", msg)
	}
}

func TestGetNetErrorTemporary(t *testing.T) {
	fmt.Println("=== 开始测试: 网络临时错误 ===")
	err := &temporaryError{}
	if msg := GetNetError(err); msg != "网络临时错误" {
		t.Errorf("GetNetError 临时错误判断失败，期望 '网络临时错误'，实际 '%s'", msg)
	}
}

func TestGetNetErrorDNSError(t *testing.T) {
	fmt.Println("=== 开始测试: 域名解析错误 ===")
	opErr := &net.OpError{
		Err: &net.DNSError{
			Err:  "no such host",
			Name: "nonexistent.example.com",
		},
	}
	if msg := GetNetError(opErr); msg != "域名解析错误" {
		t.Errorf("GetNetError DNS 判断失败，期望 '域名解析错误'，实际 '%s'", msg)
	}
}

func TestGetNetErrorSyscallError(t *testing.T) {
	fmt.Println("=== 开始测试: 系统调用错误 ===")
	opErr := &net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ECONNREFUSED,
		},
	}
	if msg := GetNetError(opErr); msg != "连接被拒绝" {
		t.Errorf("GetNetError ECONNREFUSED 判断失败，期望 '连接被拒绝'，实际 '%s'", msg)
	}

	opErr2 := &net.OpError{
		Err: &os.SyscallError{
			Err: syscall.ETIMEDOUT,
		},
	}
	if msg := GetNetError(opErr2); msg != "网络连接超时" {
		t.Errorf("GetNetError ETIMEDOUT 判断失败，期望 '网络连接超时'，实际 '%s'", msg)
	}
}

func TestGetNetErrorStringMatches(t *testing.T) {
	fmt.Println("=== 开始测试: GetNetError 字符串匹配 ===")
	tests := []struct {
		errMsg string
		want   string
	}{
		{"use of closed network connection", "监听端口已关闭"},
		{"unable to authenticate", "无法用户密码验证"},
		{"closed network connection", "使用已关闭网络连接"},
		{"connection refused", "连接被拒绝"},
		{"server gave HTTP response to HTTPS client", "服务器需要https访问"},
		{"x509: certificate is not valid", "无效的网站证书"},
		{"x509: certificate is valid", "网站证书不匹配"},
		{"no such host", "网站域名不存在"},
		{"actively refused it", "无法建立连接"},
		{"was forcibly closed by the remote host", "远程主机强制关闭了现有连接"},
		{"way forbidden by its access permissions", "访问权限被拒绝，请检查 Windows 防火墙"},
		{"broken pipe", "对端已关闭连接"},
		{"i/o timeout", "网络连接超时"},
		{"some unknown error", "some unknown error"},
	}
	for _, tt := range tests {
		got := GetNetError(errors.New(tt.errMsg))
		if got != tt.want {
			t.Errorf("GetNetError(%q) = %q, want %q", tt.errMsg, got, tt.want)
		}
	}
}

func TestGetIPAdress(t *testing.T) {
	fmt.Println("=== 开始测试: 获取本机 IP (TestGetIPAdress) ===")
	ip := GetIPAdress()
	t.Logf("获取本机 IP: %s", ip)
}

func TestGetIPAdressByPrefix(t *testing.T) {
	fmt.Println("=== 开始测试: GetIPAdressByPrefix ===")
	// Test with a prefix that likely doesn't match
	ip := GetIPAdressByPrefix("999.999")
	t.Logf("GetIPAdressByPrefix(999.999): %s", ip)

	// Test with common prefix
	ip = GetIPAdressByPrefix("192.168")
	t.Logf("GetIPAdressByPrefix(192.168): %s", ip)

	// Test with empty prefix
	ip = GetIPAdressByPrefix("")
	t.Logf("GetIPAdressByPrefix(''): %s", ip)
}

func TestGetIPAdressByName(t *testing.T) {
	fmt.Println("=== 开始测试: GetIPAdressByName ===")
	// Test with non-existent interface
	ip := GetIPAdressByName("nonexistent_iface_xyz")
	if ip != "" {
		t.Errorf("Expected empty for non-existent interface, got %s", ip)
	}

	// Test with "lo" (loopback, usually exists on Linux)
	ip = GetIPAdressByName("lo")
	t.Logf("GetIPAdressByName(lo): %s", ip)
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
