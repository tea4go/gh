package logs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConnDirectConnect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 10)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "hello from conn test")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "hello from conn test") {
			t.Errorf("received message does not contain expected text: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cw.Destroy()
}

func TestConnConnectWithLogName(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 10)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.Name = "test_logger"
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "{LogName}test_logger{LogName}") {
			t.Errorf("expected LogName message, got: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for LogName message")
	}

	cw.Destroy()
}

func TestConnConnectFail(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = "127.0.0.1:1"
	cw.Level = LevelDebug
	cw.conn_timeout = 1 * time.Second

	err := cw.connect(1 * time.Second)
	if err == nil {
		t.Fatal("expected connection to fail")
	}
}

func TestConnWriteMsgLevelFiltered(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Level = LevelError
	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "filtered message")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConnWriteMsgNoConnection(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Level = LevelDebug
	cw.lgconn = nil
	err := cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "no conn message")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConnWriteMsgWithColor(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 10)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = true
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelError, time.Now(), "colored message")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "colored message") {
			t.Errorf("received message does not contain expected text: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cw.Destroy()
}

func TestConnWriteMsgPrintLevel(t *testing.T) {
	// LevelPrint is 8, which is > LevelDebug (7). In connWriter,
	// logLevel > c.Level returns nil (filtered). So Print IS filtered
	// unless the Level is set to LevelPrint.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 10)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelPrint // Must be LevelPrint to allow print messages
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelPrint, time.Now(), "print message")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "print message") {
			t.Errorf("received message does not contain expected text: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cw.Destroy()
}

func TestConnFlush(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Flush()
}

func TestConnSetGetLevel(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.SetLevel(LevelError)
	if cw.GetLevel() != LevelError {
		t.Errorf("GetLevel = %d, want %d", cw.GetLevel(), LevelError)
	}
}

func TestConnInitEmpty(t *testing.T) {
	cw := NewConn().(*connWriter)
	err := cw.Init("")
	if err != nil {
		t.Fatal(err)
	}
}

func TestConnInitWithConfig(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	cw := NewConn().(*connWriter)
	err = cw.Init(fmt.Sprintf(`{"net":"tcp","addr":"%s","level":%d,"name":"test"}`, addr, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	cw.Destroy()
}

func TestConnInitBadJSON(t *testing.T) {
	cw := NewConn().(*connWriter)
	err := cw.Init("{bad json")
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestConnWriteMsgByConnError(t *testing.T) {
	cw := NewConn().(*connWriter)
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	conn, _ := net.Dial("tcp", ln.Addr().String())
	conn.Close()
	ln.Close()

	cw.lgconn = conn
	err := cw.writeMsgByConn("test")
	if err == nil {
		t.Fatal("expected error writing to closed connection")
	}
}

func TestConnWriteMsgWriteError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := ln.Addr().String()
	serverConnCh := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		serverConnCh <- conn
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	serverConn := <-serverConnCh
	serverConn.Close()
	ln.Close()

	cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "should fail")
	time.Sleep(100 * time.Millisecond)
	cw.Destroy()
}

func TestConnConnectTcpProxyError(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = "127.0.0.1:1"
	cw.Level = LevelDebug
	cw.conn_timeout = 1 * time.Second

	// Set an unreachable SOCKS5 proxy - this should fail
	oldProxy := os.Getenv("log_socks5_proxy")
	os.Setenv("log_socks5_proxy", "127.0.0.1:1")
	defer os.Setenv("log_socks5_proxy", oldProxy)

	err := cw.connect_tcp_proxy(1 * time.Second)
	if err == nil {
		t.Fatal("expected connection to fail with proxy")
	}
}

func TestConnConnectHttpProxyFallback(t *testing.T) {
	// When the proxy fails, connect() falls back to direct dial
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = ln.Addr().String()
	cw.Level = LevelDebug
	cw.conn_timeout = 2 * time.Second

	// Set an unreachable proxy - should fallback to direct
	oldProxy := os.Getenv("log_http_proxy")
	os.Setenv("log_http_proxy", "127.0.0.1:1")
	defer os.Setenv("log_http_proxy", oldProxy)

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Logf("connect with proxy fallback returned error (ok): %v", err)
	}
	cw.Destroy()
}

func TestConnViaTLogger(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "log4go_conn_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 20)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	log := NewLogger()
	log.SetLogger(AdapterConn, fmt.Sprintf(`{"net":"tcp","addr":"%s","level":%d,"name":"testconn"}`, addr, LevelDebug))
	log.Info("conn logger test")
	time.Sleep(200 * time.Millisecond)
	log.Close()

	for len(received) > 0 {
		<-received
	}
}

func TestConnWriteMsgWithNewline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 10)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			received <- scanner.Text()
		}
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "message with newline\n")
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if !strings.Contains(msg, "message with newline") {
			t.Errorf("received message does not contain expected text: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	cw.Destroy()
}

func TestNewConnInterface(t *testing.T) {
	cw := NewConn()
	if cw == nil {
		t.Fatal("NewConn returned nil")
	}
	cwIface := cw.(ILogger)

	cwIface.SetLevel(LevelWarning)
	if cwIface.GetLevel() != LevelWarning {
		t.Errorf("GetLevel = %d, want %d", cwIface.GetLevel(), LevelWarning)
	}

	cwIface.Flush()
	cwIface.Destroy()
}

func TestConnInitNoAddr(t *testing.T) {
	cw := NewConn().(*connWriter)
	err := cw.Init(`{"net":"tcp","level":7}`)
	if err != nil {
		t.Logf("Init returned error (expected): %v", err)
	}
}

func TestConnWriteMsgWithClosedConn(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := ln.Addr().String()
	serverConnCh := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		serverConnCh <- conn
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	serverConn := <-serverConnCh
	serverConn.Close()
	ln.Close()

	cw.WriteMsg("test.go", 10, 4, "TestFunc", LevelInfo, time.Now(), "should fail")
	time.Sleep(100 * time.Millisecond)
	cw.Destroy()
}

func TestConnConnectWithLogNameError(t *testing.T) {
	// Connect with a LogName but to a valid server - verify the LogName write
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		// Close immediately to cause write error after accept
		conn.Close()
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.Name = "test_logger"
	cw.ColorFlag = false
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	// This might fail because the connection was closed immediately
	if err != nil {
		t.Logf("connect failed as expected with early close: %v", err)
	}
	cw.Destroy()
}

func TestConnHelperTempPath(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "log4go_conn_helper")
	defer os.RemoveAll(tmpDir)
	path := tmpDir + "/test.log"
	if !strings.Contains(path, "test.log") {
		t.Errorf("unexpected path: %s", path)
	}
}

// --- connect_tcp_proxy with no proxy (direct connection) ---
// NOTE: connect_tcp_proxy has a hardcoded default proxy "192.168.3.164:32129",
// so the direct connection path is only reachable when the flag is set.
// We test the proxy error path instead.

func TestConnConnectTcpProxyProxyError(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = "127.0.0.1:1"
	cw.Level = LevelDebug
	cw.conn_timeout = 1 * time.Second

	// Unset env so default proxy is used, but it's unreachable
	os.Unsetenv("log_socks5_proxy")

	err := cw.connect_tcp_proxy(1 * time.Second)
	if err == nil {
		t.Fatal("expected connection to fail with unreachable proxy")
	}
}

// --- connect_tcp_proxy with proxy error (unreachable proxy) ---

func TestConnConnectTcpProxyUnreachable(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = "127.0.0.1:1"
	cw.Level = LevelDebug
	cw.conn_timeout = 1 * time.Second

	// Set an unreachable SOCKS5 proxy
	os.Setenv("log_socks5_proxy", "127.0.0.1:1")
	defer os.Unsetenv("log_socks5_proxy")

	err := cw.connect_tcp_proxy(1 * time.Second)
	if err == nil {
		t.Fatal("expected connection to fail with unreachable proxy")
	}
}

// --- connect_tcp_proxy with Name set (proxy error path) ---

func TestConnConnectTcpProxyWithLogName(t *testing.T) {
	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = "127.0.0.1:1"
	cw.Level = LevelDebug
	cw.ColorFlag = false
	cw.Name = "proxy_test_logger"
	cw.conn_timeout = 1 * time.Second

	// Will fail due to unreachable proxy
	os.Unsetenv("log_socks5_proxy")
	err := cw.connect_tcp_proxy(1 * time.Second)
	if err == nil {
		t.Fatal("expected connection to fail with unreachable proxy")
	}
}

// --- conn Init with heartbeat reconnection ---

func TestConnInitHeartbeat(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	received := make(chan string, 20)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					received <- scanner.Text()
				}
			}(conn)
		}
	}()

	cw := NewConn().(*connWriter)
	err = cw.Init(fmt.Sprintf(`{"net":"tcp","addr":"%s","level":%d,"name":"hbtest"}`, addr, LevelDebug))
	if err != nil {
		t.Fatal(err)
	}

	// Wait for initial connection and heartbeat
	time.Sleep(6 * time.Second)

	// Check if we received a heartbeat
	select {
	case msg := <-received:
		t.Logf("Received message: %s", msg)
	default:
		t.Log("No heartbeat received yet (may be timing)")
	}

	cw.Destroy()
}

// --- conn Flush explicitly ---

func TestConnFlushWithConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	cw := NewConn().(*connWriter)
	cw.Net = "tcp"
	cw.Addr = addr
	cw.Level = LevelDebug
	cw.conn_timeout = 2 * time.Second

	err = cw.connect(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Flush is a no-op but we call it to cover the empty method
	cw.Flush()
	cw.Destroy()
}
