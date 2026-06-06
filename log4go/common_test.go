package logs

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestRegisterNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register with nil should panic")
		}
	}()
	Register("test_nil_adapter", nil)
}

func TestRegisterDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Register with duplicate name should panic")
		}
	}()
	Register(AdapterConsole, NewConsole)
}

func TestGetLevelNameValid(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{LevelEmergency, "事故[M]"},
		{LevelAlert, "警报[A]"},
		{LevelCritical, "危险[C]"},
		{LevelError, "错误[E]"},
		{LevelWarning, "警告[W]"},
		{LevelNotice, "通知[N]"},
		{LevelInfo, "信息[I]"},
		{LevelDebug, "调试[D]"},
	}
	for _, tc := range tests {
		result := GetLevelName(tc.level)
		if result != tc.expected {
			t.Errorf("GetLevelName(%d) = %s, want %s", tc.level, result, tc.expected)
		}
	}
}

func TestGetLevelNameInvalid(t *testing.T) {
	result := GetLevelName(100)
	if result != "无效" {
		t.Errorf("GetLevelName(100) = %s, want 无效", result)
	}
	result = GetLevelName(-5)
	if result != "无效" {
		t.Errorf("GetLevelName(-5) = %s, want 无效", result)
	}
}

func TestGetNow(t *testing.T) {
	now := GetNow()
	if now.IsZero() {
		t.Fatal("GetNow should not return zero time")
	}
}

func TestCheckError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("CheckError with error should panic")
		}
	}()
	CheckError("test_pos", errors.New("test error"))
}

func TestCheckErrorNil(t *testing.T) {
	// Should not panic
	CheckError("test_pos", nil)
}

func TestGetParamStringEnv(t *testing.T) {
	os.Setenv("TEST_PARAM_STRING", "env_value")
	defer os.Unsetenv("TEST_PARAM_STRING")

	result := GetParamString("TEST_PARAM_STRING", "flag_value", "default_value")
	if result != "env_value" {
		t.Errorf("GetParamString = %s, want env_value", result)
	}
}

func TestGetParamStringFlag(t *testing.T) {
	os.Unsetenv("TEST_PARAM_STRING_2")
	result := GetParamString("TEST_PARAM_STRING_2", "flag_value", "default_value")
	if result != "flag_value" {
		t.Errorf("GetParamString = %s, want flag_value", result)
	}
}

func TestGetParamStringDefault(t *testing.T) {
	os.Unsetenv("TEST_PARAM_STRING_3")
	result := GetParamString("TEST_PARAM_STRING_3", "", "default_value")
	if result != "default_value" {
		t.Errorf("GetParamString = %s, want default_value", result)
	}
}

func TestGetParamIntEnv(t *testing.T) {
	os.Setenv("TEST_PARAM_INT", "42")
	defer os.Unsetenv("TEST_PARAM_INT")

	result := GetParamInt("TEST_PARAM_INT", 10)
	if result != 42 {
		t.Errorf("GetParamInt = %d, want 42", result)
	}
}

func TestGetParamIntFlag(t *testing.T) {
	os.Unsetenv("TEST_PARAM_INT_2")
	result := GetParamInt("TEST_PARAM_INT_2", 10)
	if result != 10 {
		t.Errorf("GetParamInt = %d, want 10", result)
	}
}

func TestGetParamBoolEnvTrue(t *testing.T) {
	os.Setenv("TEST_PARAM_BOOL", "true")
	defer os.Unsetenv("TEST_PARAM_BOOL")

	result := GetParamBool("TEST_PARAM_BOOL", false)
	if !result {
		t.Errorf("GetParamBool = %v, want true", result)
	}
}

func TestGetParamBoolEnvFalse(t *testing.T) {
	os.Setenv("TEST_PARAM_BOOL_2", "false")
	defer os.Unsetenv("TEST_PARAM_BOOL_2")

	result := GetParamBool("TEST_PARAM_BOOL_2", true)
	if result {
		t.Errorf("GetParamBool = %v, want false", result)
	}
}

func TestGetParamBoolFlag(t *testing.T) {
	os.Unsetenv("TEST_PARAM_BOOL_3")
	result := GetParamBool("TEST_PARAM_BOOL_3", true)
	if !result {
		t.Errorf("GetParamBool = %v, want true", result)
	}
}

func TestFDebugEnabled(t *testing.T) {
	old := IsDebug
	IsDebug = true
	FDebug("test message %s", "arg")
	IsDebug = old
}

func TestFDebugDisabled(t *testing.T) {
	old := IsDebug
	IsDebug = false
	FDebug("test message %s", "arg")
	IsDebug = old
}

func TestFDebugWithNewline(t *testing.T) {
	old := IsDebug
	IsDebug = true
	FDebug("test message\n")
	IsDebug = old
}

func TestFormatLogStringNoArgs(t *testing.T) {
	result := formatLog("test message")
	if result != "test message" {
		t.Errorf("formatLog = %s, want test message", result)
	}
}

func TestFormatLogStringWithFormat(t *testing.T) {
	result := formatLog("hello %s", "world")
	if result != "hello world" {
		t.Errorf("formatLog = %s, want hello world", result)
	}
}

func TestFormatLogStringNoFormat(t *testing.T) {
	result := formatLog("hello", "world")
	if !strings.Contains(result, "hello") || !strings.Contains(result, "world") {
		t.Errorf("formatLog = %s, should contain hello and world", result)
	}
}

func TestFormatLogNonString(t *testing.T) {
	result := formatLog(123)
	if result != "123" {
		t.Errorf("formatLog = %s, want 123", result)
	}
}

func TestFormatLogNonStringWithArgs(t *testing.T) {
	result := formatLog(123, "extra")
	if !strings.Contains(result, "123") || !strings.Contains(result, "extra") {
		t.Errorf("formatLog = %s, should contain 123 and extra", result)
	}
}

func TestGetFileLines(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_test_lines")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test_lines.txt")
	content := "line1\nline2\nline3\n"
	err = ioutil.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	lines, err := GetFileLines(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines != 3 {
		t.Errorf("GetFileLines = %d, want 3", lines)
	}
}

func TestGetFileLinesEmpty(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_test_lines_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test_empty.txt")
	err = ioutil.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	lines, err := GetFileLines(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if lines != 0 {
		t.Errorf("GetFileLines = %d, want 0", lines)
	}
}

func TestGetFileLinesNotExist(t *testing.T) {
	_, err := GetFileLines("/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestGetNetErrorEOF(t *testing.T) {
	result := GetNetError(io.EOF)
	if result != "网络主动断开" {
		t.Errorf("GetNetError(EOF) = %s, want 网络主动断开", result)
	}
}

func TestGetNetErrorTimeout(t *testing.T) {
	err := &timeoutError{}
	result := GetNetError(err)
	if result != "网络连接超时" {
		t.Errorf("GetNetError(timeout) = %s, want 网络连接超时", result)
	}
}

func TestGetNetErrorTemporary(t *testing.T) {
	err := &temporaryError{}
	result := GetNetError(err)
	if result != "网络临时错误" {
		t.Errorf("GetNetError(temporary) = %s, want 网络临时错误", result)
	}
}

func TestGetNetErrorDNSError(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Name: "nonexistent.invalid"},
	}
	result := GetNetError(opErr)
	if result != "域名解析错误" {
		t.Errorf("GetNetError(DNSError) = %s, want 域名解析错误", result)
	}
}

func TestGetNetErrorConnectionRefused(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Err: syscall.ECONNREFUSED},
	}
	result := GetNetError(opErr)
	if result != "连接被拒绝" {
		t.Errorf("GetNetError(ECONNREFUSED) = %s, want 连接被拒绝", result)
	}
}

func TestGetNetErrorETIMEDOUT(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Err: syscall.ETIMEDOUT},
	}
	result := GetNetError(opErr)
	if result != "网络连接超时" {
		t.Errorf("GetNetError(ETIMEDOUT) = %s, want 网络连接超时", result)
	}
}

func TestGetNetErrorAddressInUse(t *testing.T) {
	opErr := &net.OpError{
		Op:  "listen",
		Net: "tcp",
		Err: errors.New("address already in use"),
	}
	result := GetNetError(opErr)
	if result != "端口已经占用" {
		t.Errorf("GetNetError(address in use) = %s, want 端口已经占用", result)
	}
}

func TestGetNetErrorClosedNetwork(t *testing.T) {
	err := errors.New("use of closed network connection")
	result := GetNetError(err)
	if result != "监听端口已关闭" {
		t.Errorf("GetNetError(closed network) = %s, want 监听端口已关闭", result)
	}
}

func TestGetNetErrorUnableToAuthenticate(t *testing.T) {
	err := errors.New("unable to authenticate")
	result := GetNetError(err)
	if result != "无法用户密码验证" {
		t.Errorf("GetNetError(auth) = %s, want 无法用户密码验证", result)
	}
}

func TestGetNetErrorHTTPResponse(t *testing.T) {
	err := errors.New("server gave HTTP response to HTTPS client")
	result := GetNetError(err)
	if result != "服务器需要https访问" {
		t.Errorf("GetNetError(http response) = %s, want 服务器需要https访问", result)
	}
}

func TestGetNetErrorX509NotValid(t *testing.T) {
	err := errors.New("x509: certificate is not valid")
	result := GetNetError(err)
	if result != "无效的网站证书" {
		t.Errorf("GetNetError(x509 not valid) = %s, want 无效的网站证书", result)
	}
}

func TestGetNetErrorX509Valid(t *testing.T) {
	err := errors.New("x509: certificate is valid")
	result := GetNetError(err)
	if result != "网站证书不匹配" {
		t.Errorf("GetNetError(x509 valid) = %s, want 网站证书不匹配", result)
	}
}

func TestGetNetErrorNoSuchHost(t *testing.T) {
	err := errors.New("no such host")
	result := GetNetError(err)
	if result != "网站域名不存在" {
		t.Errorf("GetNetError(no such host) = %s, want 网站域名不存在", result)
	}
}

func TestGetNetErrorActivelyRefused(t *testing.T) {
	err := errors.New("actively refused it")
	result := GetNetError(err)
	if result != "无法建立连接" {
		t.Errorf("GetNetError(actively refused) = %s, want 无法建立连接", result)
	}
}

func TestGetNetErrorForciblyClosed(t *testing.T) {
	err := errors.New("was forcibly closed by the remote host")
	result := GetNetError(err)
	if result != "远程主机强制关闭了现有连接" {
		t.Errorf("GetNetError(forcibly closed) = %s, want 远程主机强制关闭了现有连接", result)
	}
}

func TestGetNetErrorBrokenPipe(t *testing.T) {
	err := errors.New("broken pipe")
	result := GetNetError(err)
	if result != "对端已关闭连接" {
		t.Errorf("GetNetError(broken pipe) = %s, want 对端已关闭连接", result)
	}
}

func TestGetNetErrorIOTimeout(t *testing.T) {
	err := errors.New("i/o timeout")
	result := GetNetError(err)
	if result != "网络连接超时" {
		t.Errorf("GetNetError(i/o timeout) = %s, want 网络连接超时", result)
	}
}

func TestGetNetErrorConnectionRefusedString(t *testing.T) {
	err := errors.New("connection refused")
	result := GetNetError(err)
	if result != "连接被拒绝" {
		t.Errorf("GetNetError(connection refused) = %s, want 连接被拒绝", result)
	}
}

func TestGetNetErrorClosedConnection(t *testing.T) {
	err := errors.New("closed network connection")
	result := GetNetError(err)
	if result != "使用已关闭网络连接" {
		t.Errorf("GetNetError(closed connection) = %s, want 使用已关闭网络连接", result)
	}
}

func TestGetNetErrorSocketPermission(t *testing.T) {
	err := errors.New("An attempt was made to access a socket in a way forbidden by its access permissions.")
	result := GetNetError(err)
	if result != "服务不可用" {
		t.Errorf("GetNetError(socket permission) = %s, want 服务不可用", result)
	}
}

func TestGetNetErrorUnknown(t *testing.T) {
	err := errors.New("some unknown error")
	result := GetNetError(err)
	if result != "some unknown error" {
		t.Errorf("GetNetError(unknown) = %s, want some unknown error", result)
	}
}

// Mock net.Error implementations
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

type temporaryError struct{}

func (e *temporaryError) Error() string   { return "temporary" }
func (e *temporaryError) Timeout() bool   { return false }
func (e *temporaryError) Temporary() bool { return true }

func TestShowArgs(t *testing.T) {
	// Just verify it doesn't panic
	ShowArgs()
}

func TestStartLogger(t *testing.T) {
	// Save and restore env vars
	oldLevel := os.Getenv("log_level")
	oldShort := os.Getenv("log_short")
	oldServer := os.Getenv("log_server")
	oldName := os.Getenv("log_name")
	defer func() {
		os.Setenv("log_level", oldLevel)
		os.Setenv("log_short", oldShort)
		os.Setenv("log_server", oldServer)
		os.Setenv("log_name", oldName)
	}()

	os.Setenv("log_level", "7")
	os.Setenv("log_short", "false")
	os.Unsetenv("log_server")
	os.Unsetenv("log_name")

	StartLogger("test_log_name")
	Reset()
}

func TestStartLoggerWithServer(t *testing.T) {
	oldLevel := os.Getenv("log_level")
	oldShort := os.Getenv("log_short")
	oldServer := os.Getenv("log_server")
	oldName := os.Getenv("log_name")
	defer func() {
		os.Setenv("log_level", oldLevel)
		os.Setenv("log_short", oldShort)
		os.Setenv("log_server", oldServer)
		os.Setenv("log_name", oldName)
	}()

	os.Setenv("log_level", "7")
	os.Setenv("log_short", "false")
	os.Setenv("log_server", "127.0.0.1:9999") // Non-existent server
	os.Unsetenv("log_name")

	// This will try to connect but should not panic
	StartLogger("test_log_name")
	Reset()
}

func TestStartLoggerInvalidLevel(t *testing.T) {
	oldLevel := os.Getenv("log_level")
	defer os.Setenv("log_level", oldLevel)

	os.Setenv("log_level", "invalid")
	StartLogger()
	Reset()
}

func TestStartLoggerOutOfRangeLevel(t *testing.T) {
	oldLevel := os.Getenv("log_level")
	defer os.Setenv("log_level", oldLevel)

	os.Setenv("log_level", "100")
	StartLogger()
	Reset()
}
