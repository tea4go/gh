package network

import (
	"encoding/gob"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
)

// --- Types for gobconn tests ---

type GobTestMsg struct {
	ID   int
	Data string
}

type GobTestMsg2 struct {
	Code  int
	Label string
	Flag  bool
}

type GobTestPtrMsg struct {
	Value int
}

func init() {
	RegisterType(reflect.TypeOf(GobTestMsg{}))
	RegisterType(reflect.TypeOf(GobTestMsg2{}))
	RegisterType(reflect.TypeOf(GobTestPtrMsg{}))
}

// TestNewGobConnection tests creating a new GobConnection from a net.Pipe
func TestNewGobConnection(t *testing.T) {
	fmt.Println("=== 测试: NewGobConnection ===")
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	conn := NewGobConnection(client)
	if conn == nil {
		t.Fatal("NewGobConnection returned nil")
	}
}

// TestNewGobConnectionFromPool tests that after Close, a new connection can be obtained from the pool
func TestNewGobConnectionFromPool(t *testing.T) {
	fmt.Println("=== 测试: NewGobConnection (pool reuse) ===")
	server, client := net.Pipe()

	conn1 := NewGobConnection(client)
	conn1.Close()
	server.Close()

	server2, client2 := net.Pipe()
	defer server2.Close()
	defer client2.Close()

	conn2 := NewGobConnection(client2)
	if conn2 == nil {
		t.Fatal("NewGobConnection from pool returned nil")
	}
	conn2.Close()
}

// helper: send one message from client to server, verify it
func roundTrip(t *testing.T, msg interface{}) message {
	t.Helper()
	server, client := net.Pipe()
	sconn := NewGobConnection(server)
	cconn := NewGobConnection(client)

	var readMsg message
	var readErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		readMsg, readErr = sconn.Read()
	}()

	err := cconn.Write(msg)
	if err != nil {
		client.Close()
		server.Close()
		t.Fatalf("Write failed: %v", err)
	}

	wg.Wait()

	if readErr != nil {
		sconn.Close()
		cconn.Close()
		server.Close()
		client.Close()
		t.Fatalf("Read failed: %v", readErr)
	}

	sconn.Close()
	cconn.Close()
	server.Close()
	client.Close()
	return readMsg
}

// TestGobConnectionReadWrite tests bidirectional send/receive using net.Pipe
func TestGobConnectionReadWrite(t *testing.T) {
	fmt.Println("=== 测试: GobConnection Read/Write ===")

	msg := roundTrip(t, GobTestMsg{ID: 42, Data: "hello"})
	defer msg.Recovery()

	val := msg.Interface()
	tm, ok := val.(GobTestMsg)
	if !ok {
		t.Fatalf("wrong message type: %T", val)
	}
	if tm.ID != 42 || tm.Data != "hello" {
		t.Errorf("message mismatch: ID=%d Data=%s", tm.ID, tm.Data)
	}
}

// TestGobConnectionReadWriteMultiple tests multiple messages one at a time
func TestGobConnectionReadWriteMultiple(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 多消息收发 ===")

	// Message 1
	m1 := roundTrip(t, GobTestMsg{ID: 1, Data: "first"})
	v1 := m1.Interface()
	if tm, ok := v1.(GobTestMsg); !ok || tm.ID != 1 || tm.Data != "first" {
		t.Errorf("msg1 mismatch: %+v", v1)
	}
	m1.Recovery()

	// Message 2
	m2 := roundTrip(t, GobTestMsg2{Code: 2, Label: "second", Flag: true})
	v2 := m2.Interface()
	if tm, ok := v2.(GobTestMsg2); !ok || tm.Code != 2 || tm.Label != "second" || !tm.Flag {
		t.Errorf("msg2 mismatch: %+v", v2)
	}
	m2.Recovery()

	// Message 3
	m3 := roundTrip(t, GobTestPtrMsg{Value: 99})
	v3 := m3.Interface()
	if tm, ok := v3.(GobTestPtrMsg); !ok || tm.Value != 99 {
		t.Errorf("msg3 mismatch: %+v", v3)
	}
	m3.Recovery()
}

// TestGobConnectionWritePointer tests writing a pointer type
// Note: gobconn.Write uses value.Elem().Type().String() for pointers,
// so the receiver decodes as the non-pointer value type.
func TestGobConnectionWritePointer(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 写入指针类型 ===")

	RegisterType(reflect.TypeOf(&GobTestPtrMsg{}))

	msg := roundTrip(t, &GobTestPtrMsg{Value: 777})
	defer msg.Recovery()

	val := msg.Interface()
	// Write detects ptr/interface kind and uses Elem type, so Read returns non-pointer
	tm, ok := val.(GobTestPtrMsg)
	if !ok {
		t.Fatalf("wrong type: %T", val)
	}
	if tm.Value != 777 {
		t.Errorf("value mismatch: got %d, want 777", tm.Value)
	}
}

// TestGobConnectionReadUnregisteredType tests Read with an unregistered type
func TestGobConnectionReadUnregisteredType(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 读未注册类型 ===")

	server, client := net.Pipe()
	sconn := NewGobConnection(server)
	cconn := NewGobConnection(client)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		enc := gob.NewEncoder(client)
		err := enc.Encode(msgStruct{StructName: "nonexistent.Type"})
		if err != nil {
			t.Logf("encode header error: %v", err)
		}
	}()

	_, err := sconn.Read()
	if err == nil {
		t.Error("expected error reading unregistered type, got nil")
	}
	t.Logf("Read unregistered type error (expected): %v", err)

	wg.Wait()
	sconn.Close()
	cconn.Close()
	server.Close()
	client.Close()
}

// TestGobConnectionReadDecodeError tests Read when the connection is closed prematurely
func TestGobConnectionReadDecodeError(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 读解码错误 ===")

	server, client := net.Pipe()
	sconn := NewGobConnection(server)

	client.Close()

	_, err := sconn.Read()
	if err == nil {
		t.Error("expected error from Read on closed connection, got nil")
	}
	t.Logf("Read on closed conn error (expected): %v", err)

	sconn.Close()
	server.Close()
}

// TestGobConnectionWriteError tests Write when the connection is closed
func TestGobConnectionWriteError(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 写关闭连接 ===")

	server, client := net.Pipe()
	cconn := NewGobConnection(client)

	server.Close()
	client.Close()

	err := cconn.Write(GobTestMsg{ID: 1, Data: "fail"})
	if err == nil {
		t.Error("expected error writing to closed connection, got nil")
	}
	t.Logf("Write to closed conn error (expected): %v", err)
}

// TestGobConnectionLocalRemoteAddr tests LocalAddr and RemoteAddr
func TestGobConnectionLocalRemoteAddr(t *testing.T) {
	fmt.Println("=== 测试: GobConnection LocalAddr/RemoteAddr ===")

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	conn := NewGobConnection(client)

	local := conn.LocalAddr()
	remote := conn.RemoteAddr()

	if local == nil {
		t.Error("LocalAddr returned nil")
	}
	if remote == nil {
		t.Error("RemoteAddr returned nil")
	}
	t.Logf("LocalAddr: %v, RemoteAddr: %v", local, remote)
}

// TestGobConnectionClose tests Close method
func TestGobConnectionClose(t *testing.T) {
	fmt.Println("=== 测试: GobConnection Close ===")

	server, client := net.Pipe()
	conn := NewGobConnection(client)

	err := conn.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
	server.Close()
}

// TestGetMsgTypeAll tests GetMsgType, GetMsgAllType, DeleteType
func TestGetMsgTypeAll(t *testing.T) {
	fmt.Println("=== 测试: GetMsgType/GetMsgAllType/DeleteType ===")

	typ := reflect.TypeOf(GobTestMsg{})
	RegisterType(typ)

	// GetMsgType - existing
	got, err := GetMsgType(typ.String())
	if err != nil {
		t.Errorf("GetMsgType failed: %v", err)
	}
	if got != typ {
		t.Error("GetMsgType type mismatch")
	}

	// GetMsgAllType
	all := GetMsgAllType()
	found := false
	for _, name := range all {
		if name == typ.String() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetMsgAllType did not contain %s", typ.String())
	}

	// DeleteType
	DeleteType(typ.String())
	_, err = GetMsgType(typ.String())
	if err == nil {
		t.Error("GetMsgType should fail after DeleteType")
	}

	// Re-register for other tests
	RegisterType(typ)
}

// TestGetMsgTypeNotRegistered tests GetMsgType for an unregistered name
func TestGetMsgTypeNotRegistered(t *testing.T) {
	fmt.Println("=== 测试: GetMsgType 未注册 ===")

	_, err := GetMsgType("nonexistent.Type")
	if err == nil {
		t.Error("expected error for unregistered type")
	}
	if err != Errortype {
		t.Errorf("expected Errortype, got: %v", err)
	}
}

// TestGetPointerPutPointer tests the pointer pool get/put cycle
func TestGetPointerPutPointer(t *testing.T) {
	fmt.Println("=== 测试: getPointer/putPointer ===")

	typ := reflect.TypeOf(GobTestMsg{})

	v1 := getPointer(typ)
	if v1.IsNil() {
		t.Error("getPointer returned nil value")
	}
	v1.Elem().Set(reflect.Zero(typ))
	putPointer(v1)

	v2 := getPointer(typ)
	if v2.IsNil() {
		t.Error("getPointer returned nil value on second call")
	}
	putPointer(v2)
}

// TestGetMsgPutMsg tests the message pool get/put cycle
func TestGetMsgPutMsg(t *testing.T) {
	fmt.Println("=== 测试: getMsg/putMsg ===")

	m1 := getMsg()
	// Pool may return a reused message; just verify it's a valid message struct
	putMsg(m1)

	m2 := getMsg()
	putMsg(m2)
}

// TestMessageInterface tests message.Interface()
func TestMessageInterface(t *testing.T) {
	fmt.Println("=== 测试: message.Interface ===")

	typ := reflect.TypeOf(GobTestMsg{})
	v := reflect.New(typ)
	v.Elem().Set(reflect.ValueOf(GobTestMsg{ID: 7, Data: "test"}))

	msg := message{Type: typ.String(), value: v}
	iface := msg.Interface()
	tm, ok := iface.(GobTestMsg)
	if !ok {
		t.Fatalf("wrong type: %T", iface)
	}
	if tm.ID != 7 || tm.Data != "test" {
		t.Errorf("Interface() mismatch: %+v", tm)
	}
	msg.Recovery()
}

// TestMessageRecovery tests message.Recovery()
func TestMessageRecovery(t *testing.T) {
	fmt.Println("=== 测试: message.Recovery ===")

	typ := reflect.TypeOf(GobTestMsg{})
	v := reflect.New(typ)
	v.Elem().Set(reflect.ValueOf(GobTestMsg{ID: 8, Data: "recover"}))

	msg := message{Type: typ.String(), value: v}
	msg.Recovery() // Should not panic
}

// TestGobConnectionConcurrent tests concurrent write + read
func TestGobConnectionConcurrent(t *testing.T) {
	fmt.Println("=== 测试: GobConnection 并发 ===")

	server, client := net.Pipe()
	sconn := NewGobConnection(server)
	cconn := NewGobConnection(client)

	const count = 10
	var wg sync.WaitGroup
	wg.Add(1)

	received := 0
	go func() {
		defer wg.Done()
		for i := 0; i < count; i++ {
			msg, err := sconn.Read()
			if err != nil {
				t.Errorf("concurrent Read failed at %d: %v", i, err)
				return
			}
			msg.Recovery()
			received++
		}
	}()

	for i := 0; i < count; i++ {
		err := cconn.Write(GobTestMsg{ID: i, Data: fmt.Sprintf("msg%d", i)})
		if err != nil {
			t.Fatalf("concurrent Write failed at %d: %v", i, err)
		}
	}

	wg.Wait()
	if received != count {
		t.Errorf("received %d messages, expected %d", received, count)
	}

	sconn.Close()
	cconn.Close()
	server.Close()
	client.Close()
}
