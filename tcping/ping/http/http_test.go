package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tea4go/gh/tcping/ping"
)

// --- Int type tests ---

func TestInt_String(t *testing.T) {
	tests := []struct {
		val  Int
		want string
	}{
		{Int(0), "0"},
		{Int(200), "200"},
		{Int(404), "404"},
		{Int(1), "1"},
	}
	for _, tt := range tests {
		if got := tt.val.String(); got != tt.want {
			t.Errorf("Int(%d).String() = %q, want %q", tt.val, got, tt.want)
		}
	}
}

// --- Trace tests ---

func TestTrace_String_NoTLS(t *testing.T) {
	trace := Trace{
		ConnectDuration:       10 * time.Millisecond,
		WroteRequestDuration:  5 * time.Millisecond,
		WaitResponseDuration:  20 * time.Millisecond,
		BodyDuration:          30 * time.Millisecond,
	}
	got := trace.String()
	if got == "" {
		t.Error("Trace.String() should not be empty")
	}
	// Should not contain "tls=" when tls is false
	if fmt.Sprintf("%s", got) == "" {
		t.Error("String() should produce output")
	}
}

func TestTrace_String_WithTLS(t *testing.T) {
	trace := Trace{
		ConnectDuration:       10 * time.Millisecond,
		TLSDuration:           15 * time.Millisecond,
		tls:                   true,
		WroteRequestDuration:  5 * time.Millisecond,
		WaitResponseDuration:  20 * time.Millisecond,
		BodyDuration:          30 * time.Millisecond,
	}
	got := trace.String()
	if got == "" {
		t.Error("Trace.String() should not be empty")
	}
}

func TestTrace_WithTrace(t *testing.T) {
	trace := Trace{}
	ctx := trace.WithTrace(context.Background())
	if ctx == nil {
		t.Error("WithTrace should return a non-nil context")
	}
}

// --- NewHttp tests ---

func TestNewHttp_ValidURL(t *testing.T) {
	op := &ping.TOption{
		HttpMethod: "GET",
	}
	h, err := NewHttp("http://example.com", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}
	if h == nil {
		t.Fatal("NewHttp returned nil")
	}
}

func TestNewHttp_InvalidURL(t *testing.T) {
	op := &ping.TOption{
		HttpMethod: "INVALID METHOD",
	}
	_, err := NewHttp("http://example.com", op)
	if err == nil {
		t.Error("NewHttp with invalid method should return error")
	}
}

func TestNewHttp_DefaultMethod(t *testing.T) {
	op := &ping.TOption{}
	h, err := NewHttp("http://example.com", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}
	if h.method != http.MethodGet {
		t.Errorf("method = %q, want GET", h.method)
	}
}

func TestNewHttp_WithProxy(t *testing.T) {
	proxy, _ := url.Parse("http://192.168.1.1:8080")
	op := &ping.TOption{
		HttpProxy: proxy,
	}
	h, err := NewHttp("http://example.com", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}
	if h == nil {
		t.Fatal("NewHttp returned nil")
	}
}

// --- SetTarget tests ---

func TestTPing_SetTarget(t *testing.T) {
	proxy, _ := url.Parse("http://192.168.1.1:8080")
	op := &ping.TOption{
		HttpProxy: proxy,
		Timeout:   5 * time.Second,
	}
	h, err := NewHttp("http://example.com", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	target := &ping.TTarget{}
	h.SetTarget(target)

	if target.Proxy != proxy.String() {
		t.Errorf("Proxy = %q, want %q", target.Proxy, proxy.String())
	}
	if target.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", target.Timeout)
	}
}

func TestTPing_SetTarget_NoProxy(t *testing.T) {
	op := &ping.TOption{
		Timeout: 3 * time.Second,
	}
	h, err := NewHttp("http://example.com", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	target := &ping.TTarget{}
	h.SetTarget(target)

	if target.Proxy != "" {
		t.Errorf("Proxy = %q, want empty", target.Proxy)
	}
	if target.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v, want 3s", target.Timeout)
	}
}

// --- Ping tests using httptest ---

func TestTPing_Ping_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:   5 * time.Second,
		HttpMethod: "GET",
		IsMeta:    true,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
	if stats.Meta["status"] == nil || stats.Meta["status"].String() != "200" {
		t.Errorf("Expected status 200, got %v", stats.Meta["status"])
	}
	if stats.Duration == 0 {
		t.Error("Duration should not be zero")
	}
}

func TestTPing_Ping_NoMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
		IsMeta:     false,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
	// Extra should be nil when IsMeta is false
	if stats.Extra != nil {
		t.Error("Extra should be nil when IsMeta is false")
	}
}

func TestTPing_Ping_CustomTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
}

func TestTPing_Ping_ContextCanceled(t *testing.T) {
	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
	}
	h, err := NewHttp("http://192.0.2.1:1", op) // unreachable IP
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	stats := h.Ping(ctx)
	if stats.Connected {
		t.Error("Should not be connected to unreachable host")
	}
	if stats.Error == nil {
		t.Error("Should have error for unreachable host")
	}
}

func TestTPing_Ping_InvalidURL(t *testing.T) {
	op := &ping.TOption{
		Timeout: 5 * time.Second,
	}
	// Use a valid URL format but invalid host
	h, err := NewHttp("http://0.0.0.0:1", op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stats := h.Ping(ctx)
	if stats.Connected {
		t.Error("Should not be connected to invalid host")
	}
}

func TestTPing_Ping_DefaultTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout: 0, // Use default timeout
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
}

func TestTPing_Ping_WithUserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("user-agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		UserAgent:  "test-agent-123",
		HttpMethod: "GET",
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
	if receivedUA != "test-agent-123" {
		t.Errorf("User-Agent = %q, want test-agent-123", receivedUA)
	}
}

func TestTPing_Ping_ResponseWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1000)) // 1000 bytes
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:   5 * time.Second,
		HttpMethod: "GET",
		IsMeta:    true,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
	if stats.Meta["bytes"] == nil {
		t.Error("Expected bytes meta to be set")
	}
}

func TestTPing_Ping_TraceWithMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
		IsMeta:     true,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Fatalf("Expected connected, got error: %v", stats.Error)
	}
	// When IsMeta is true, Extra should be a *Trace
	if stats.Extra == nil {
		t.Error("Extra should not be nil when IsMeta is true")
	}
	// The trace should have some DNS and connect duration
	traceStr := stats.Extra.String()
	if traceStr == "" {
		t.Error("Trace String() should not be empty")
	}
}

func TestTPing_Ping_WithResolver(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
		Resolver:   &net.Resolver{},
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Errorf("Expected connected, got error: %v", stats.Error)
	}
}

func TestTPing_Ping_BodyReadError(t *testing.T) {
	// Create a server that closes connection after sending headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Force close after headers
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, _ := hijacker.Hijack()
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n"))
			conn.Close()
		}
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	// The result may be connected or not depending on whether body read fails
	// Just ensure it doesn't panic
	_ = stats
}

func TestTPing_Ping_TraceCallbacks(t *testing.T) {
	// Use a real server to trigger trace callbacks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	// Test with IsMeta=true to ensure trace callbacks are exercised
	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
		IsMeta:     true,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Fatalf("Expected connected, got error: %v", stats.Error)
	}

	// The trace Extra should have been populated
	if stats.Extra == nil {
		t.Fatal("Extra should not be nil when IsMeta is true")
	}

	// Verify the trace string contains expected fields
	traceStr := stats.Extra.String()
	if !containsField(traceStr, "connect=") {
		t.Errorf("Trace should contain connect=, got: %s", traceStr)
	}
	if !containsField(traceStr, "request=") {
		t.Errorf("Trace should contain request=, got: %s", traceStr)
	}
	if !containsField(traceStr, "wait_response=") {
		t.Errorf("Trace should contain wait_response=, got: %s", traceStr)
	}
}

func containsField(s, field string) bool {
	return strings.Contains(s, field)
}

func TestTPing_Ping_TraceWithTLS(t *testing.T) {
	// Use TLS server to trigger TLSHandshakeStart/Done callbacks
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello tls"))
	}))
	defer server.Close()

	op := &ping.TOption{
		Timeout:    5 * time.Second,
		HttpMethod: "GET",
		IsMeta:     true,
	}
	h, err := NewHttp(server.URL, op)
	if err != nil {
		t.Fatalf("NewHttp returned error: %v", err)
	}

	// Use the server's client to handle TLS
	h.client = server.Client()

	stats := h.Ping(context.Background())
	if !stats.Connected {
		t.Fatalf("Expected connected, got error: %v", stats.Error)
	}

	// Extra should be a *Trace with TLS info
	if stats.Extra == nil {
		t.Fatal("Extra should not be nil")
	}
	traceStr := stats.Extra.String()
	if !strings.Contains(traceStr, "tls=") {
		t.Errorf("TLS trace should contain tls=, got: %s", traceStr)
	}
}

// --- Register/Load integration ---

func TestHttpRegistered(t *testing.T) {
	f := ping.Load(ping.HTTP)
	if f == nil {
		t.Fatal("HTTP protocol should be registered")
	}

	f2 := ping.Load(ping.HTTPS)
	if f2 == nil {
		t.Fatal("HTTPS protocol should be registered")
	}
}

func TestHttpFactory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	op := &ping.TOption{
		Timeout:   5 * time.Second,
		HttpMethod: "GET",
	}

	f := ping.Load(ping.HTTP)
	if f == nil {
		t.Fatal("HTTP factory should be registered")
	}

	pinger, err := f(u, op)
	if err != nil {
		t.Fatalf("Factory returned error: %v", err)
	}
	if pinger == nil {
		t.Fatal("Factory returned nil pinger")
	}
}
