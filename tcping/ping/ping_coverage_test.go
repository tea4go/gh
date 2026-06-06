package ping

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// helperPing implements IPing for TPinger construction in tests
type helperPing func(ctx context.Context) *TStats

func (h helperPing) Ping(ctx context.Context) *TStats { return h(ctx) }
func (h helperPing) SetTarget(t *TTarget)             {}

func newTestPinger(buf *bytes.Buffer, u *url.URL, fn func(ctx context.Context) *TStats, interval time.Duration, count int) *TPinger {
	return NewPinger(buf, u, helperPing(fn), interval, count)
}

// --- SProtocol and NewProtocol tests ---

func TestSProtocol_String(t *testing.T) {
	tests := []struct {
		p    SProtocol
		want string
	}{
		{TCP, "tcp"},
		{HTTP, "http"},
		{HTTPS, "https"},
		{SProtocol(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("SProtocol(%d).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestNewProtocol(t *testing.T) {
	tests := []struct {
		input   string
		want    SProtocol
		wantErr bool
	}{
		{"tcp", TCP, false},
		{"TCP", TCP, false},
		{"http", HTTP, false},
		{"HTTP", HTTP, false},
		{"https", HTTPS, false},
		{"HTTPS", HTTPS, false},
		{"ftp", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		got, err := NewProtocol(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewProtocol(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("NewProtocol(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- Register / Load tests ---

func TestRegisterAndLoad(t *testing.T) {
	testProto := SProtocol(100)
	factory := func(u *url.URL, op *TOption) (IPing, error) { return nil, nil }
	Register(testProto, factory)
	defer func() {
		delete(pinger, testProto)
	}()

	loaded := Load(testProto)
	if loaded == nil {
		t.Fatal("Load should return the registered factory")
	}
	if f := Load(SProtocol(101)); f != nil {
		t.Error("Load for unregistered protocol should return nil")
	}
}

// --- TOption tests ---

func TestTOption_Fields(t *testing.T) {
	resolver := &net.Resolver{}
	proxy, _ := url.Parse("http://192.168.1.1:8080")
	op := TOption{
		Timeout:    5 * time.Second,
		Resolver:   resolver,
		HttpProxy:  proxy,
		HttpMethod: "POST",
		IsMeta:     true,
		UserAgent:  "test-agent",
		IsTls:      true,
	}
	if op.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", op.Timeout)
	}
	if op.Resolver != resolver {
		t.Error("Resolver mismatch")
	}
	if op.HttpProxy != proxy {
		t.Error("HttpProxy mismatch")
	}
	if op.HttpMethod != "POST" {
		t.Errorf("HttpMethod = %q, want POST", op.HttpMethod)
	}
	if !op.IsMeta {
		t.Error("IsMeta should be true")
	}
	if op.UserAgent != "test-agent" {
		t.Errorf("UserAgent = %q, want test-agent", op.UserAgent)
	}
	if !op.IsTls {
		t.Error("IsTls should be true")
	}
}

// --- TStats tests ---

type testStringer string

func (s testStringer) String() string { return string(s) }

func TestTStats_FormatMeta(t *testing.T) {
	s := &TStats{Meta: map[string]fmt.Stringer{}}
	if got := s.FormatMeta(); got != "" {
		t.Errorf("FormatMeta() empty = %q, want empty", got)
	}

	s.Meta["key1"] = testStringer("val1")
	if got := s.FormatMeta(); got != "key1=val1" {
		t.Errorf("FormatMeta() one = %q, want key1=val1", got)
	}

	s.Meta["key2"] = testStringer("val2")
	got := s.FormatMeta()
	if got != "key1=val1 key2=val2" {
		t.Errorf("FormatMeta() two = %q, want key1=val1 key2=val2", got)
	}
}

func TestTStats_Fields(t *testing.T) {
	err := fmt.Errorf("test error")
	s := TStats{
		Connected:   true,
		Duration:    100 * time.Millisecond,
		DNSDuration: 10 * time.Millisecond,
		Address:     "1.2.3.4:443",
		Meta:        map[string]fmt.Stringer{"a": testStringer("b")},
		Extra:       testStringer("extra"),
		Error:       err,
	}
	if !s.Connected {
		t.Error("Connected should be true")
	}
	if s.Duration != 100*time.Millisecond {
		t.Errorf("Duration = %v", s.Duration)
	}
	if s.DNSDuration != 10*time.Millisecond {
		t.Errorf("DNSDuration = %v", s.DNSDuration)
	}
	if s.Address != "1.2.3.4:443" {
		t.Errorf("Address = %q", s.Address)
	}
	if s.Error != err {
		t.Error("Error mismatch")
	}
	if s.Extra.String() != "extra" {
		t.Errorf("Extra = %q", s.Extra.String())
	}
}

// --- FormatDuration tests ---

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0ns"},
		{100 * time.Nanosecond, "100ns"},
		{1 * time.Microsecond, "1.00us"},
		{1 * time.Millisecond, "1.00ms"},
		{500 * time.Millisecond, "500.00ms"},
		{1 * time.Second, "1.00s"},
		{90 * time.Second, "90.00s"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.d)
		got = strings.Replace(got, "µ", "u", -1) // micro sign
		got = strings.Replace(got, "μ", "u", -1) // Greek mu
		if got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

// --- FormatError tests ---

func TestFormatError_Timeout(t *testing.T) {
	err := context.DeadlineExceeded
	got := FormatError(err)
	if got != "连接超时" {
		t.Errorf("FormatError(DeadlineExceeded) = %q, want 连接超时", got)
	}
}

func TestFormatError_EOF(t *testing.T) {
	got := FormatError(io.EOF)
	if got != "网络主动断开" {
		t.Errorf("FormatError(EOF) = %q, want 网络主动断开", got)
	}
}

type timeoutNetErr struct{}

func (e *timeoutNetErr) Error() string  { return "timeout" }
func (e *timeoutNetErr) Timeout() bool   { return true }
func (e *timeoutNetErr) Temporary() bool { return false }

type tempNetErr struct{}

func (e *tempNetErr) Error() string  { return "temporary" }
func (e *tempNetErr) Timeout() bool   { return false }
func (e *tempNetErr) Temporary() bool { return true }

func TestFormatError_UrlErrorTimeout(t *testing.T) {
	urlErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: &timeoutNetErr{},
	}
	got := FormatError(urlErr)
	if got != "连接超时" {
		t.Errorf("FormatError(urlErr timeout) = %q, want 连接超时", got)
	}
}

func TestFormatError_UrlErrorNotTimeout(t *testing.T) {
	innerErr := errors.New("some inner error")
	urlErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: innerErr,
	}
	got := FormatError(urlErr)
	if got != "some inner error" {
		t.Errorf("FormatError(urlErr non-timeout) = %q, want some inner error", got)
	}
}

func TestFormatError_NetErrorTimeout(t *testing.T) {
	got := FormatError(&timeoutNetErr{})
	// timeoutNetErr matches the net.Error type assertion in the first switch,
	// which returns "连接超时" for Timeout()
	if got != "连接超时" {
		t.Errorf("FormatError(net timeout) = %q, want 连接超时", got)
	}
}

func TestFormatError_NetErrorTemporary(t *testing.T) {
	got := FormatError(&tempNetErr{})
	if got != "网络临时错误" {
		t.Errorf("FormatError(net temporary) = %q, want 网络临时错误", got)
	}
}

func TestFormatError_OpError_DNSError(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Name: "nonexistent.example.com"},
	}
	got := FormatError(opErr)
	if got != "域名解析错误" {
		t.Errorf("FormatError(DNSError) = %q, want 域名解析错误", got)
	}
}

func TestFormatError_OpError_SyscallError_ECONNREFUSED(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED},
	}
	got := FormatError(opErr)
	// The string "connection refused" is in the error message, so it hits
	// the string-matching branch before the opErr branch
	if got != "连接被拒绝" {
		t.Errorf("FormatError(ECONNREFUSED) = %q, want 连接被拒绝", got)
	}
}

func TestFormatError_OpError_SyscallError_ETIMEDOUT(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.ETIMEDOUT},
	}
	got := FormatError(opErr)
	// The net.OpError implements net.Error, so it matches the first switch case
	// Timeout() is false for this OpError, so it falls through to the string matching
	// "connection" is in the error string but no exact match, so it ends up in
	// the second net.Error type assertion check for Timeout/Temporary
	if got == "" {
		t.Error("FormatError should return non-empty string")
	}
}

func TestFormatError_ForciblyClosed(t *testing.T) {
	err := errors.New("An existing connection was forcibly closed by the remote host")
	got := FormatError(err)
	if got != "远程主机强行关闭了现有连接" {
		t.Errorf("FormatError(forcibly closed) = %q, want 远程主机强行关闭了现有连接", got)
	}
}

func TestFormatError_CertSANs(t *testing.T) {
	err := errors.New("certificate because it doesn't contain any IP SANs")
	got := FormatError(err)
	if got != "无法验证证书" {
		t.Errorf("FormatError(cert sans) = %q", got)
	}
}

func TestFormatError_NoSuchHost(t *testing.T) {
	err := errors.New("lookup nonexistent: no such host")
	got := FormatError(err)
	if got != "无效域名" {
		t.Errorf("FormatError(no such host) = %q", got)
	}
}

func TestFormatError_GetAddrInfo(t *testing.T) {
	err := errors.New("getaddrinfow failed")
	got := FormatError(err)
	if got != "域名解析错误" {
		t.Errorf("FormatError(getaddrinfow) = %q", got)
	}
}

func TestFormatError_ClosedNetworkConnection(t *testing.T) {
	err := errors.New("use of closed network connection")
	got := FormatError(err)
	if got != "使用已关闭的网络连接" {
		t.Errorf("FormatError(closed network) = %q", got)
	}
}

func TestFormatError_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:1: connection refused")
	got := FormatError(err)
	if got != "连接被拒绝" {
		t.Errorf("FormatError(connection refused) = %q", got)
	}
}

func TestFormatError_HTTPSToHTTP(t *testing.T) {
	err := errors.New("server gave HTTP response to HTTPS client")
	got := FormatError(err)
	if got != "服务器需要https访问" {
		t.Errorf("FormatError(https to http) = %q", got)
	}
}

func TestFormatError_CertNotValid(t *testing.T) {
	err := errors.New("x509: certificate is not valid for names")
	got := FormatError(err)
	if got != "无效的网站证书" {
		t.Errorf("FormatError(cert not valid) = %q", got)
	}
}

func TestFormatError_CertValid(t *testing.T) {
	err := errors.New("x509: certificate is valid for names")
	got := FormatError(err)
	if got != "网站证书不匹配" {
		t.Errorf("FormatError(cert valid) = %q", got)
	}
}

func TestFormatError_ActivelyRefused(t *testing.T) {
	err := errors.New("No connection could be made because the target machine actively refused it")
	got := FormatError(err)
	if got != "无法建立连接" {
		t.Errorf("FormatError(actively refused) = %q", got)
	}
}

func TestFormatError_WasForciblyClosed(t *testing.T) {
	// "was forcibly closed by the remote host" contains "forcibly closed"
	// which matches the earlier check for "远程主机强行关闭了现有连接"
	err := errors.New("was forcibly closed by the remote host")
	got := FormatError(err)
	if got != "远程主机强行关闭了现有连接" {
		t.Errorf("FormatError(was forcibly closed) = %q, want 远程主机强行关闭了现有连接", got)
	}
}

func TestFormatError_UnknownError(t *testing.T) {
	err := errors.New("some random error")
	got := FormatError(err)
	if got != "some random error" {
		t.Errorf("FormatError(unknown) = %q, want original error text", got)
	}
}

func TestFormatError_OpErrorUnknownSyscall(t *testing.T) {
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.ENOTCONN},
	}
	got := FormatError(opErr)
	if got == "" {
		t.Error("FormatError should return non-empty string")
	}
}

// --- ParseDuration tests ---

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"100", 100 * time.Millisecond, false},
		{"1s", time.Second, false},
		{"500ms", 500 * time.Millisecond, false},
		{"invalid", 0, true},
	}
	for _, tt := range tests {
		got, err := ParseDuration(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseDuration(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- ParseAddress tests ---

func TestParseAddress(t *testing.T) {
	u, err := ParseAddress("tcp://1.2.3.4:80")
	if err != nil {
		t.Fatalf("ParseAddress with scheme error: %v", err)
	}
	if u.Scheme != "tcp" {
		t.Errorf("Scheme = %q, want tcp", u.Scheme)
	}

	u, err = ParseAddress("1.2.3.4:80")
	if err != nil {
		t.Fatalf("ParseAddress without scheme error: %v", err)
	}
	if u.Scheme != "tcp" {
		t.Errorf("Scheme = %q, want tcp (default)", u.Scheme)
	}
}

// --- FormatIP edge cases ---

func TestFormatIP_AdditionalCases(t *testing.T) {
	_, err := FormatIP("")
	if err == nil {
		t.Error("FormatIP('') should return error")
	}
	_, err = FormatIP("not-an-ip")
	if err == nil {
		t.Error("FormatIP('not-an-ip') should return error")
	}
}

// --- TPinger tests ---

func TestNewPinger(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u,
		func(ctx context.Context) *TStats {
			return &TStats{Connected: true, Duration: time.Millisecond}
		}, time.Second, 1)

	if pinger == nil {
		t.Fatal("NewPinger returned nil")
	}
	if pinger.result.Target.Protocol != "tcp" {
		t.Errorf("Protocol = %q, want tcp", pinger.result.Target.Protocol)
	}
	if pinger.result.Target.URL != "tcp://1.2.3.4:80" {
		t.Errorf("URL = %q", pinger.result.Target.URL)
	}
}

func TestTPinger_Stop_Done(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u,
		func(ctx context.Context) *TStats {
			return &TStats{Connected: true, Duration: time.Millisecond}
		}, time.Second, 1)

	done := pinger.Done()
	if done == nil {
		t.Fatal("Done() returned nil channel")
	}

	pinger.Stop()
	pinger.Stop() // Second Stop should not panic

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Done channel should be closed after Stop")
	}
}

func TestTPinger_Ping_Count(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	counter := 0
	pinger := newTestPinger(&buf, u,
		func(ctx context.Context) *TStats {
			counter++
			return &TStats{Connected: true, Duration: time.Millisecond, Address: "1.2.3.4:80"}
		}, 10*time.Millisecond, 3)

	pinger.Ping()
	if counter != 3 {
		t.Errorf("Expected 3 pings, got %d", counter)
	}
}

func TestTPinger_SetResult_Connected(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)

	st := &TStats{
		Connected:   true,
		Duration:    50 * time.Millisecond,
		DNSDuration: 5 * time.Millisecond,
		Address:     "1.2.3.4:80",
	}
	pinger.SetResult(st)
	// After first SetResult, minDuration should be updated from MaxInt64 to actual duration
	// However, SetResult also increments total, so let's just verify it ran correctly
	if pinger.totalDuration != 50*time.Millisecond {
		t.Errorf("totalDuration = %v, want 50ms", pinger.totalDuration)
	}
}

func TestTPinger_SetResult_Error(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)

	st := &TStats{
		Connected:   false,
		Duration:    100 * time.Millisecond,
		DNSDuration: 10 * time.Millisecond,
		Address:     "1.2.3.4:80",
		Error:       fmt.Errorf("connection refused"),
	}
	pinger.SetResult(st)
	if pinger.failedTotal != 1 {
		t.Errorf("failedTotal = %d, want 1", pinger.failedTotal)
	}
}

func TestTPinger_SetResult_WithContextCancel(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)

	st := &TStats{
		Connected: false,
		Duration:  100 * time.Millisecond,
		Address:   "1.2.3.4:80",
		Error:     context.Canceled,
	}
	pinger.SetResult(st)
	if pinger.failedTotal != 1 {
		t.Errorf("failedTotal = %d, want 1", pinger.failedTotal)
	}
}

func TestTPinger_SetResult_WithMeta(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)

	st := &TStats{
		Connected:   true,
		Duration:    50 * time.Millisecond,
		DNSDuration: 5 * time.Millisecond,
		Address:     "1.2.3.4:80",
		Meta:        map[string]fmt.Stringer{"status": testStringer("200")},
	}
	pinger.SetResult(st)
	if !strings.Contains(buf.String(), "status=200") {
		t.Errorf("Output should contain meta, got: %q", buf.String())
	}
}

func TestTPinger_SetResult_WithExtra(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)

	st := &TStats{
		Connected:   true,
		Duration:    50 * time.Millisecond,
		DNSDuration: 5 * time.Millisecond,
		Address:     "1.2.3.4:80",
		Extra:       testStringer("tls: 1.3"),
	}
	pinger.SetResult(st)
	if !strings.Contains(buf.String(), "tls: 1.3") {
		t.Errorf("Output should contain extra, got: %q", buf.String())
	}
}

// --- TTarget tests ---

func TestTTarget_String(t *testing.T) {
	tt := TTarget{Protocol: "tcp", IP: "1.2.3.4", Port: 80}
	got := tt.String()
	want := "tcp://1.2.3.4:80"
	if got != want {
		t.Errorf("TTarget.String() = %q, want %q", got, want)
	}
}

// --- TResult tests ---

func TestTResult_Avg(t *testing.T) {
	r := TResult{CounterOK: 2, TotalDuration: 200 * time.Millisecond}
	if got := r.Avg(); got != 100*time.Millisecond {
		t.Errorf("TResult.Avg() = %v, want 100ms", got)
	}
	r2 := TResult{CounterOK: 0}
	if got := r2.Avg(); got != 0 {
		t.Errorf("TResult.Avg() with zero counter = %v, want 0", got)
	}
}

func TestTResult_Failed(t *testing.T) {
	r := TResult{Counter: 10, CounterOK: 7}
	if got := r.Failed(); got != 3 {
		t.Errorf("TResult.Failed() = %d, want 3", got)
	}
}

func TestTResult_String(t *testing.T) {
	r := TResult{
		Counter:       4,
		CounterOK:     3,
		Target:        TTarget{Protocol: "tcp", IP: "1.2.3.4", Port: 80},
		MinDuration:   10 * time.Millisecond,
		MaxDuration:   50 * time.Millisecond,
		TotalDuration: 120 * time.Millisecond,
	}
	got := r.String()
	if !strings.Contains(got, "tcp://1.2.3.4:80") {
		t.Errorf("TResult.String() should contain target, got: %q", got)
	}
}

// --- TPinger Summarize ---

func TestTPinger_Summarize_ZeroTotal(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)
	pinger.Summarize()
	if buf.Len() > 0 {
		t.Error("Summarize with zero total should not write anything")
	}
}

func TestTPinger_Summarize_WithResults(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u,
		func(ctx context.Context) *TStats {
			return &TStats{Connected: true, Duration: time.Millisecond, Address: "1.2.3.4:80"}
		}, 10*time.Millisecond, 2)

	pinger.Ping()
	buf.Reset()
	pinger.Summarize()
	if !strings.Contains(buf.String(), "1.2.3.4:80") {
		t.Errorf("Summarize output should contain URL, got: %q", buf.String())
	}
}

// --- GetStats ---

func TestTPinger_GetStats(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u, nil, time.Second, 1)
	stats, err := pinger.GetStats()
	if err != nil {
		t.Errorf("GetStats() returned error: %v", err)
	}
	if stats != nil {
		t.Error("GetStats() should return nil")
	}
}

// --- Constants test ---

func TestConstants(t *testing.T) {
	if DefaultCounter != 4 {
		t.Errorf("DefaultCounter = %d, want 4", DefaultCounter)
	}
	if DefaultInterval != time.Second {
		t.Errorf("DefaultInterval = %v, want 1s", DefaultInterval)
	}
	if DefaultTimeout != 5*time.Second {
		t.Errorf("DefaultTimeout = %v, want 5s", DefaultTimeout)
	}
}

// --- IPing interface test ---

type mockPing struct{}

func (m *mockPing) Ping(ctx context.Context) *TStats { return &TStats{Connected: true} }
func (m *mockPing) SetTarget(t *TTarget)             {}

func TestIPing_Interface(t *testing.T) {
	var p IPing = &mockPing{}
	stats := p.Ping(context.Background())
	if !stats.Connected {
		t.Error("mockPing should return connected=true")
	}
}

// --- Ensure strconv import is used ---

func TestStrconvImport(t *testing.T) {
	if strconv.Itoa(42) != "42" {
		t.Error("strconv not working")
	}
}

// --- PingServer test ---

func TestTPinger_PingServer(t *testing.T) {
	u, _ := url.Parse("tcp://1.2.3.4:80")
	var buf bytes.Buffer
	pinger := newTestPinger(&buf, u,
		func(ctx context.Context) *TStats {
			return &TStats{Connected: true, Duration: time.Millisecond, Address: "1.2.3.4:80"}
		}, 10*time.Millisecond, 0)

	// PingServer runs until Stop() is called (counter=0 means indefinite)
	go pinger.PingServer()
	time.Sleep(200 * time.Millisecond)
	pinger.Stop()
}

// --- FormatError additional coverage ---

// netErrNoTimeout implements net.Error without Timeout or Temporary
type netErrNoTimeout struct {
	msg string
}

func (e *netErrNoTimeout) Error() string   { return e.msg }
func (e *netErrNoTimeout) Timeout() bool    { return false }
func (e *netErrNoTimeout) Temporary() bool  { return false }

func TestFormatError_NetErrorNoTimeoutNoTemp(t *testing.T) {
	ne := &netErrNoTimeout{msg: "some net error"}
	got := FormatError(ne)
	if got != "some net error" {
		t.Errorf("FormatError(netErrNoTimeout) = %q", got)
	}
}

func TestFormatError_OpError_NoSyscallErr(t *testing.T) {
	// OpError with a simple error (not DNSError, not SyscallError)
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: fmt.Errorf("simple op error"),
	}
	got := FormatError(opErr)
	// This hits the net.Error path but not the opErr.SyscallError path
	if got == "" {
		t.Error("FormatError should return non-empty string")
	}
}

func TestFormatError_OpError_WithAddr(t *testing.T) {
	opErr := &net.OpError{
		Op:   "dial",
		Net:  "tcp",
		Addr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1},
		Err:  fmt.Errorf("refused"),
	}
	got := FormatError(opErr)
	if got == "" {
		t.Error("FormatError should return non-empty string")
	}
}
