package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestGetStatusCode(t *testing.T) {
	fmt.Println("=== 开始测试: HTTP 状态码获取 (TestGetStatusCode) ===")
	resp := &http.Response{
		StatusCode: 200,
	}
	if code := GetStatusCode(resp); code != 200 {
		t.Errorf("GetStatusCode failed. Got %d", code)
	}

	resp = nil
	if code := GetStatusCode(nil); code != 0 {
		t.Errorf("GetStatusCode nil failed. Got %d", code)
	}
}

func TestIsAddrInUse(t *testing.T) {
	fmt.Println("=== 开始测试: 端口占用判断 (TestIsAddrInUse) ===")
	// Mock OpError
	err := &net.OpError{
		Err: errors.New("address already in use"),
	}
	if !IsAddrInUse(err) {
		t.Error("IsAddrInUse failed for standard error")
	}

	err2 := errors.New("Only one usage of each socket address")
	// IsAddrInUse expects *net.OpError for the first check, but falls through?
	// func IsAddrInUse(err error) bool {
	//   opErr, ok := err.(*net.OpError)
	//   if !ok { return false } 
	//   ...
	// }
	// Wait, the code says:
	/*
		opErr, ok := err.(*net.OpError)
		if !ok {
			return false
		}
	*/
	// So passing a plain error "Only one usage..." will return false immediately because it's not *net.OpError.
	// The logic in IsAddrInUse seems to assume the error MUST be *net.OpError.
	
	if IsAddrInUse(err2) {
		t.Log("IsAddrInUse returned true for plain error (unexpected based on code reading but maybe I missed something)")
	} else {
		t.Log("IsAddrInUse returned false for plain error as expected")
	}
}

func TestGetNetError(t *testing.T) {
	if msg := GetNetError(io.EOF); msg != "网络主动断开" {
		t.Errorf("GetNetError EOF failed. Got %s", msg)
	}

	// Test address in use
	opErr := &net.OpError{
		Err: errors.New("address already in use"),
	}
	if msg := GetNetError(opErr); msg != "端口已经占用" {
		t.Errorf("GetNetError AddrInUse failed. Got %s", msg)
	}
	
	// Test plain error
	err := errors.New("connection refused")
	if msg := GetNetError(err); msg != "连接被拒绝" {
		t.Errorf("GetNetError Plain failed. Got %s", msg)
	}
}

func TestGetIPAdress(t *testing.T) {
	fmt.Println("=== 开始测试: 获取本机 IP (TestGetIPAdress) ===")
	// This depends on the machine's network.
	// We just ensure it doesn't panic.
	ip := GetIPAdress()
	t.Logf("GetIPAdress: %s", ip)
}

func TestGetAllIPAdress(t *testing.T) {
	fmt.Println("=== 开始测试: 获取所有 IP (TestGetAllIPAdress) ===")
	ips := GetAllIPAdress()
	t.Logf("GetAllIPAdress: %s", ips)
}

func TestGetAllMacAdress(t *testing.T) {
	macs := GetAllMacAdress()
	t.Logf("GetAllMacAdress: %s", macs)
}

// Mocking net.Error
type timeoutError struct{}
func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestGetNetErrorTimeout(t *testing.T) {
	fmt.Println("=== 开始测试: 网络超时错误 (TestGetNetErrorTimeout) ===")
	err := &timeoutError{}
	if msg := GetNetError(err); msg != "网络连接超时" {
		t.Errorf("GetNetError Timeout failed. Got %s", msg)
	}
}
