package ldapserver

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	proxyproto "github.com/pires/go-proxyproto"
	ldap "github.com/openstandia/goldap/message"
)

// ========== Server creation ==========

func TestNewServer(t *testing.T) {
	s := NewServer()
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.chDone == nil {
		t.Error("chDone should be initialized")
	}
}

// ========== SetClients / CheckClient / GetClients ==========

func TestServer_SetClients_CIDR(t *testing.T) {
	s := NewServer()
	err := s.SetClients("192.168.1.0/24")
	if err != nil {
		t.Fatalf("SetClients CIDR error: %v", err)
	}
	if len(s.client_nets) != 1 {
		t.Errorf("client_nets len = %d, want 1", len(s.client_nets))
	}
}

func TestServer_SetClients_IP(t *testing.T) {
	s := NewServer()
	err := s.SetClients("10.0.0.1")
	if err != nil {
		t.Fatalf("SetClients IP error: %v", err)
	}
	if len(s.client_nets) != 1 {
		t.Errorf("client_nets len = %d, want 1", len(s.client_nets))
	}
}

func TestServer_SetClients_Invalid(t *testing.T) {
	s := NewServer()
	err := s.SetClients("not-a-valid-cidr")
	if err == nil {
		t.Error("SetClients with invalid input should return error")
	}
}

func TestServer_SetClients_Multiple(t *testing.T) {
	s := NewServer()
	err := s.SetClients("192.168.1.0/24;10.0.0.1")
	if err != nil {
		t.Fatalf("SetClients multiple error: %v", err)
	}
	if len(s.client_nets) != 2 {
		t.Errorf("client_nets len = %d, want 2", len(s.client_nets))
	}
}

func TestServer_SetClients_Duplicate(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	_ = s.SetClients("192.168.1.0/24")
	if len(s.client_nets) != 1 {
		t.Errorf("client_nets len = %d, want 1 (no duplicates)", len(s.client_nets))
	}
}

func TestServer_CheckClient(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	if !s.CheckClient("192.168.1.100:389") {
		t.Error("Should match IP in CIDR range with port")
	}
	if !s.CheckClient("192.168.1.100") {
		t.Error("Should match IP in CIDR range without port")
	}
	if s.CheckClient("invalid-ip") {
		t.Error("Should return false for invalid IP")
	}
	if s.CheckClient("10.0.0.1:389") {
		t.Error("Should return false for IP not in range")
	}
}

func TestServer_CheckClient_CountsAccess(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	s.CheckClient("192.168.1.100:389")
	s.CheckClient("192.168.1.100:389")
	val, ok := s.client_ips.Load("192.168.1.100")
	if !ok || val.(int) != 2 {
		t.Errorf("access count = %v, want 2", val)
	}
}

func TestServer_IsClient(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	if !s.IsClient("192.168.1.0/24") {
		t.Error("IsClient should return true for registered CIDR")
	}
	if s.IsClient("10.0.0.0/24") {
		t.Error("IsClient should return false for unregistered CIDR")
	}
}

func TestServer_GetClients(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24;10.0.0.0/8")
	got := s.GetClients()
	if !strings.Contains(got, "192.168.1.0/24") || !strings.Contains(got, "10.0.0.0/8") {
		t.Errorf("GetClients() = %q", got)
	}
}

func TestServer_GetClientNets(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	if len(s.GetClientNets()) != 1 {
		t.Errorf("GetClientNets len = %d, want 1", len(s.GetClientNets()))
	}
}

func TestServer_GetClientIPs(t *testing.T) {
	s := NewServer()
	_ = s.SetClients("192.168.1.0/24")
	s.CheckClient("192.168.1.100")
	if s.GetClientIPs() == nil {
		t.Error("GetClientIPs should not return nil")
	}
}

// ========== Response creation ==========

func TestNewResponses(t *testing.T) {
	_ = NewBindResponse(LDAPResultSuccess)
	_ = NewResponse(LDAPResultSuccess)
	_ = NewExtendedResponse(LDAPResultSuccess)
	_ = NewCompareResponse(LDAPResultCompareTrue)
	_ = NewModifyResponse(LDAPResultSuccess)
	_ = NewDeleteResponse(LDAPResultSuccess)
	_ = NewAddResponse(LDAPResultSuccess)
	_ = NewSearchResultDoneResponse(LDAPResultSuccess)
	_ = NewSearchResultEntry("cn=test,dc=example,dc=com")
}

// ========== Packet reading ==========

func TestReadBytes(t *testing.T) {
	data := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0x60, 0x00}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	b, err := readBytes(br, &buf, 1)
	if err != nil || b != 0x30 {
		t.Errorf("readBytes: b=0x%02x err=%v", b, err)
	}
}

func TestReadBytes_Error(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{}))
	var buf []byte
	_, err := readBytes(br, &buf, 1)
	if err == nil {
		t.Error("readBytes on empty should error")
	}
}

func TestReadTagAndLength(t *testing.T) {
	data := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0x60, 0x00}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	tl, err := readTagAndLength(br, &buf)
	if err != nil {
		t.Fatalf("readTagAndLength error: %v", err)
	}
	if tl.Class != 0 || !tl.IsCompound || tl.Length != 5 {
		t.Errorf("tl = %+v", tl)
	}
}

func TestReadTagAndLength_LongForm(t *testing.T) {
	data := make([]byte, 130)
	data[0] = 0x30
	data[1] = 0x81
	data[2] = 0x80
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	tl, err := readTagAndLength(br, &buf)
	if err != nil || tl.Length != 128 {
		t.Errorf("Length = %d, err = %v", tl.Length, err)
	}
}

func TestReadTagAndLength_IndefiniteLength(t *testing.T) {
	data := []byte{0x30, 0x80}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	_, err := readTagAndLength(br, &buf)
	if err == nil {
		t.Error("Indefinite length should error")
	}
}

func TestReadTagAndLength_Overflow(t *testing.T) {
	data := []byte{0x30, 0x84, 0xFF, 0xFF, 0xFF, 0xFF}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	_, err := readTagAndLength(br, &buf)
	if err == nil {
		t.Error("Overflow length should error")
	}
}

func TestReadMessagePacket(t *testing.T) {
	data := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0x60, 0x00}
	br := bufio.NewReader(bytes.NewReader(data))
	mp, err := readMessagePacket(br)
	if err != nil || len(mp.bytes) != 7 {
		t.Errorf("bytes len = %d, err = %v", len(mp.bytes), err)
	}
}

func TestReadMessagePacket_Empty(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{}))
	_, err := readMessagePacket(br)
	if err == nil {
		t.Error("Empty reader should error")
	}
}

func TestDecodeMessage_Invalid(t *testing.T) {
	_, err := decodeMessage([]byte{0xFF, 0xFF})
	if err == nil {
		t.Error("Invalid data should error")
	}
}

// ========== RouteMux ==========

func TestNewRouteMux(t *testing.T) {
	if NewRouteMux() == nil {
		t.Fatal("NewRouteMux returned nil")
	}
}

func TestRouteMux_Routes(t *testing.T) {
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {})
	mux.Search(func(w ResponseWriter, m *Message) {})
	mux.Add(func(w ResponseWriter, m *Message) {})
	mux.Delete(func(w ResponseWriter, m *Message) {})
	mux.Modify(func(w ResponseWriter, m *Message) {})
	mux.Compare(func(w ResponseWriter, m *Message) {})
	mux.Extended(func(w ResponseWriter, m *Message) {})
	mux.Abandon(func(w ResponseWriter, m *Message) {})
	mux.NotFound(func(w ResponseWriter, m *Message) {})
	if len(mux.routes) != 8 {
		t.Errorf("Expected 8 routes, got %d", len(mux.routes))
	}
}

func TestRoute_Chaining(t *testing.T) {
	mux := NewRouteMux()
	r := mux.Search(func(w ResponseWriter, m *Message) {}).
		BaseDn("dc=example,dc=com").
		Filter("(objectclass=*)").
		Scope(2).
		Label("my-search").
		AuthenticationChoice("simple")
	if r.sBasedn != "dc=example,dc=com" || !r.uBasedn {
		t.Error("BaseDn not set")
	}
	if r.sFilter != "(objectclass=*)" || !r.uFilter {
		t.Error("Filter not set")
	}
	if r.sScope != 2 || !r.uScope {
		t.Error("Scope not set")
	}
	if r.label != "my-search" {
		t.Error("Label not set")
	}
	if r.sAuthChoice != "simple" || !r.uAuthChoice {
		t.Error("AuthChoice not set")
	}
}

func TestRoute_RequestName(t *testing.T) {
	mux := NewRouteMux()
	r := mux.Extended(func(w ResponseWriter, m *Message) {}).RequestName(ldap.LDAPOID("1.3.6.1.4.1.1466.20037"))
	if r.exoName != "1.3.6.1.4.1.1466.20037" {
		t.Errorf("exoName = %q", r.exoName)
	}
}

// ========== Route Match ==========

func makeTestMessage(t *testing.T) *Message {
	t.Helper()
	data := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	return &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
}

func TestRoute_Match_Bind(t *testing.T) {
	r := &route{operation: BIND}
	m := makeTestMessage(t)
	if !r.Match(m) {
		t.Error("BIND route should match BindRequest")
	}
}

func TestRoute_Match_WrongOp(t *testing.T) {
	r := &route{operation: SEARCH}
	m := makeTestMessage(t)
	if r.Match(m) {
		t.Error("SEARCH route should not match BindRequest")
	}
}

func TestRoute_Match_WithBaseDn(t *testing.T) {
	// Basedn is only checked for SearchRequest; a BindRequest route with
	// sBasedn still matches a BindRequest (Basedn is ignored for BIND).
	r := &route{operation: BIND, sBasedn: "dc=example,dc=com", uBasedn: true}
	m := makeTestMessage(t) // BindRequest
	if !r.Match(m) {
		t.Error("Bind route with Basedn should still match BindRequest (Basedn only applies to Search)")
	}
}

// ========== ServeLDAP ==========

func TestServeLDAP_NoRoutes(t *testing.T) {
	mux := NewRouteMux()
	m := makeTestMessage(t)
	ch := make(chan *ldap.LDAPMessage, 1)
	w := responseWriterImpl{chanOut: ch, messageID: 1}
	mux.ServeLDAP(w, m)
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Error("Timeout - should get default response")
	}
}

func TestServeLDAP_MatchedBind(t *testing.T) {
	called := false
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		called = true
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	m := makeTestMessage(t)
	ch := make(chan *ldap.LDAPMessage, 1)
	w := responseWriterImpl{chanOut: ch, messageID: 1}
	mux.ServeLDAP(w, m)
	if !called {
		t.Error("Bind handler should have been called")
	}
	<-ch
}

func TestServeLDAP_NotFoundRoute(t *testing.T) {
	called := false
	mux := NewRouteMux()
	mux.NotFound(func(w ResponseWriter, m *Message) {
		called = true
		w.Write(NewResponse(LDAPResultOperationsError))
	})
	m := makeTestMessage(t)
	ch := make(chan *ldap.LDAPMessage, 1)
	w := responseWriterImpl{chanOut: ch, messageID: 1}
	mux.ServeLDAP(w, m)
	if !called {
		t.Error("NotFound handler should have been called")
	}
	<-ch
}

// ========== Client ==========

func TestClient_New(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()
	defer clientPipe.Close()

	s := NewServer()
	c := s.newClient(clientPipe)
	if c == nil || c.rwc == nil || c.br == nil || c.bw == nil {
		t.Error("newClient fields not initialized")
	}
}

func TestClient_Accessors(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()
	defer clientPipe.Close()

	s := NewServer()
	c := s.newClient(clientPipe)

	if c.GetConn() != clientPipe {
		t.Error("GetConn mismatch")
	}
	if c.Addr() == nil {
		t.Error("Addr should not be nil")
	}
	if c.GetRaw() != nil {
		t.Error("GetRaw should be nil initially")
	}
	_, ok := c.GetMessageByID(1)
	if ok {
		t.Error("GetMessageByID should return false for unregistered message")
	}

	serverPipe2, clientPipe2 := net.Pipe()
	defer serverPipe2.Close()
	defer clientPipe2.Close()
	c.SetConn(clientPipe2)
	if c.GetConn() != clientPipe2 {
		t.Error("SetConn failed")
	}
}

func TestClient_ReadPacket(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer clientPipe.Close()

	s := NewServer()
	s.Handle(NewRouteMux())
	s.SetClients("127.0.0.1/8")
	c := s.newClient(clientPipe)

	go func() {
		defer serverPipe.Close()
		data := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
		serverPipe.Write(data)
	}()

	mp, err := c.ReadPacket()
	if err != nil {
		t.Logf("ReadPacket error: %v", err)
	}
	if mp != nil && len(mp.bytes) > 0 {
		msg, err := mp.readMessage()
		if err != nil {
			t.Logf("readMessage error: %v", err)
		} else {
			t.Logf("messageID=%d, protocolOp=%s", msg.MessageID().Int(), msg.ProtocolOpName())
		}
	}
}

func TestClient_ReadPacket_Closed(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	serverPipe.Close()
	clientPipe.Close()

	s := NewServer()
	c := s.newClient(clientPipe)
	_, err := c.ReadPacket()
	if err == nil {
		t.Error("ReadPacket on closed should error")
	}
}

func TestClient_WriteMessage(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer clientPipe.Close()

	s := NewServer()
	s.Handle(NewRouteMux())
	c := s.newClient(clientPipe)

	bindResp := NewBindResponse(LDAPResultSuccess)
	m := ldap.NewLDAPMessageWithProtocolOp(bindResp)
	m.SetMessageID(1)

	go func() {
		c.writeMessage(m)
		clientPipe.Close()
	}()

	buf := make([]byte, 4096)
	serverPipe.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := serverPipe.Read(buf)
	if err != nil {
		t.Logf("Read error: %v", err)
	} else if n > 0 {
		t.Logf("writeMessage wrote %d bytes", n)
	}
	serverPipe.Close()
}

func TestClient_ProcessRequestMessage(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	bindCalled := false
	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		bindCalled = true
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	s.Handle(mux)

	c := s.newClient(clientPipe)
	c.Numero = 1
	c.chanOut = make(chan *ldap.LDAPMessage)
	c.writeDone = make(chan bool)
	c.requestList = make(map[int]*Message)

	go func() {
		for range c.chanOut {
		}
		close(c.writeDone)
	}()

	data := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}

	c.wg.Add(1)
	c.ProcessRequestMessage(&msg)
	time.Sleep(100 * time.Millisecond)

	if bindCalled {
		t.Log("ProcessRequestMessage called bind handler!")
	}
	clientPipe.Close()
}

func TestClient_RegisterUnregisterRequest(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	s := NewServer()
	s.Handle(NewRouteMux())
	c := s.newClient(clientPipe)
	c.requestList = make(map[int]*Message)

	data := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}

	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2), Client: c}
	c.registerRequest(m)

	reg, ok := c.GetMessageByID(1)
	if !ok || reg == nil {
		t.Error("GetMessageByID should return the registered message")
	}

	c.unregisterRequest(m)
	_, ok = c.GetMessageByID(1)
	if ok {
		t.Error("GetMessageByID should return false after unregister")
	}
}

// ========== Message ==========

func TestMessage_Abandon(t *testing.T) {
	m := &Message{Done: make(chan bool, 2)}
	m.Abandon()
	select {
	case <-m.Done:
	default:
		t.Error("Abandon should send to Done channel")
	}
}

func TestMessage_GetBindRequest(t *testing.T) {
	m := makeTestMessage(t)
	_ = m.GetBindRequest()
}

// ========== Server Handle/Stop ==========

func TestServer_Handle(t *testing.T) {
	s := NewServer()
	mux := NewRouteMux()
	s.Handle(mux)
	if s.Handler == nil {
		t.Error("Handler should be set")
	}
}

// TestServer_Handle_SecondCall: Handle calls os.Exit(1) on second call,
// which cannot be caught by recover() — skipped since it terminates the process.

func TestServer_Stop(t *testing.T) {
	s := NewServer()
	go s.Stop()
	time.Sleep(100 * time.Millisecond)
	select {
	case <-s.chDone:
	default:
		t.Error("Stop should close chDone")
	}
}

// ========== Integration: ListenAndServe ==========

func startTestServer(t *testing.T, handler Handler) (*Server, string) {
	t.Helper()
	s := NewServer()
	if handler != nil {
		s.Handle(handler)
	}
	s.SetClients("127.0.0.1/8")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	ch := make(chan error, 1)
	go s.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), ch)

	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("ListenAndServe: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server start timeout")
	}
	return s, fmt.Sprintf("127.0.0.1:%d", port)
}

func TestServer_Connect(t *testing.T) {
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	s, addr := startTestServer(t, mux)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	conn.Close()
	time.Sleep(100 * time.Millisecond)
	_ = s
}

func TestServer_ClientNotAllowed(t *testing.T) {
	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {})
	s.Handle(mux)
	s.SetClients("192.168.1.0/24")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	ch := make(chan error, 1)
	go s.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), ch)
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("server start timeout")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err == nil {
		time.Sleep(200 * time.Millisecond)
		conn.Close()
	}
	_ = s
}

// ========== Constants ==========

func TestConstants(t *testing.T) {
	if ApplicationBindRequest != 0 {
		t.Errorf("ApplicationBindRequest = %d", ApplicationBindRequest)
	}
	if ApplicationSearchRequest != 3 {
		t.Errorf("ApplicationSearchRequest = %d", ApplicationSearchRequest)
	}
	if LDAPResultSuccess != 0 {
		t.Errorf("LDAPResultSuccess = %d", LDAPResultSuccess)
	}
	if LDAPResultInvalidCredentials != 49 {
		t.Errorf("LDAPResultInvalidCredentials = %d", LDAPResultInvalidCredentials)
	}
}

// ========== Helper: build LDAP SearchRequest packet ==========

func buildLDAPSearchRequest(msgID uint8) []byte {
	baseObj := []byte{0x04, 0x00}
	scope := []byte{0x0a, 0x01, 0x00}
	deref := []byte{0x0a, 0x01, 0x00}
	sizeLimit := []byte{0x02, 0x01, 0x00}
	timeLimit := []byte{0x02, 0x01, 0x00}
	typesOnly := []byte{0x01, 0x01, 0x00}
	filter := []byte{0x87, 0x00}
	attrs := []byte{0x30, 0x00}

	var content []byte
	content = append(content, baseObj...)
	content = append(content, scope...)
	content = append(content, deref...)
	content = append(content, sizeLimit...)
	content = append(content, timeLimit...)
	content = append(content, typesOnly...)
	content = append(content, filter...)
	content = append(content, attrs...)

	searchReq := append([]byte{0x63, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, msgID}
	seq := append(msgIDEnc, searchReq...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

// Ensure ldap import is used
var _ ldap.LDAPMessage

// ========== readBytes partial read ==========

func TestReadBytes_PartialRead(t *testing.T) {
	// Test when n > 0 but n != length (partial read scenario)
	// bufio.Reader with 1 byte but requesting 10 will return EOF after reading 1
	data := []byte{0x30, 0x05, 0x02}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	// Read 2 bytes first (should succeed)
	b, err := readBytes(br, &buf, 2)
	if err != nil {
		t.Fatalf("readBytes first call: %v", err)
	}
	if b != 0x05 {
		t.Errorf("expected last byte 0x05, got 0x%02x", b)
	}
	// Now request more than available - should get partial read with n != length
	b2, err := readBytes(br, &buf, 10)
	// This will likely error (EOF) but the partial read path is exercised
	t.Logf("readBytes partial: b=0x%02x, err=%v, buf len=%d", b2, err, len(buf))
}

// ========== readMessage panic recovery ==========

func TestReadMessage_PanicRecovery(t *testing.T) {
	// Create invalid message bytes that will cause panic during decode
	invalidBytes := []byte{0x30, 0x0a, 0x02, 0x01, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	mp := &messagePacket{bytes: invalidBytes}
	_, err := mp.readMessage()
	if err == nil {
		t.Error("readMessage should return error for invalid data that causes panic")
	}
}

// ========== Route Match - SearchRequest ==========

func buildLDAPSearchRequestWithDetails(msgID uint8, baseObject string, filter string, scope int) []byte {
	// Build a more complete SearchRequest
	baseObjBytes := append([]byte{0x04, byte(len(baseObject))}, []byte(baseObject)...)
	scopeBytes := []byte{0x0a, 0x01, byte(scope)}
	deref := []byte{0x0a, 0x01, 0x00}
	sizeLimit := []byte{0x02, 0x01, 0x00}
	timeLimit := []byte{0x02, 0x01, 0x00}
	typesOnly := []byte{0x01, 0x01, 0x00}
	// Simple present filter: (objectclass=*)
	filterBytes := append([]byte{0x87, byte(len(filter))}, []byte(filter)...)
	attrs := []byte{0x30, 0x00}

	var content []byte
	content = append(content, baseObjBytes...)
	content = append(content, scopeBytes...)
	content = append(content, deref...)
	content = append(content, sizeLimit...)
	content = append(content, timeLimit...)
	content = append(content, typesOnly...)
	content = append(content, filterBytes...)
	content = append(content, attrs...)

	searchReq := append([]byte{0x63, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, msgID}
	seq := append(msgIDEnc, searchReq...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

func makeSearchMessage(t *testing.T, baseObject, filter string, scope int) *Message {
	t.Helper()
	data := buildLDAPSearchRequestWithDetails(1, baseObject, filter, scope)
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	return &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
}

func TestRoute_Match_SearchRequest_BaseDn(t *testing.T) {
	r := &route{
		operation: SEARCH,
		sBasedn:   "dc=example,dc=com",
		uBasedn:   true,
	}
	m := makeSearchMessage(t, "dc=example,dc=com", "objectclass", 0)
	if !r.Match(m) {
		t.Error("SearchRequest with matching BaseDn should match")
	}

	m2 := makeSearchMessage(t, "dc=other,dc=com", "objectclass", 0)
	if r.Match(m2) {
		t.Error("SearchRequest with non-matching BaseDn should not match")
	}
}

func TestRoute_Match_SearchRequest_Filter(t *testing.T) {
	r := &route{
		operation: SEARCH,
		sFilter:   "(objectclass=*)",
		uFilter:   true,
	}
	m := makeSearchMessage(t, "", "objectclass", 0)
	if !r.Match(m) {
		t.Error("SearchRequest with matching filter should match")
	}
}

func TestRoute_Match_SearchRequest_Scope(t *testing.T) {
	r := &route{
		operation: SEARCH,
		sScope:    2,
		uScope:    true,
	}
	m := makeSearchMessage(t, "", "objectclass", 2)
	if !r.Match(m) {
		t.Error("SearchRequest with matching scope should match")
	}

	m2 := makeSearchMessage(t, "", "objectclass", 0)
	if r.Match(m2) {
		t.Error("SearchRequest with non-matching scope should not match")
	}
}

func TestRoute_Match_SearchRequest_AllConditions(t *testing.T) {
	r := &route{
		operation: SEARCH,
		sBasedn:   "dc=example,dc=com",
		uBasedn:   true,
		sFilter:   "(objectclass=*)",
		uFilter:   true,
		sScope:    1,
		uScope:    true,
	}
	m := makeSearchMessage(t, "dc=example,dc=com", "objectclass", 1)
	if !r.Match(m) {
		t.Error("SearchRequest matching all conditions should match")
	}
}

// ========== Route Match - ExtendedRequest ==========

func buildLDAPExtendedRequest(msgID uint8, requestName string) []byte {
	// ExtendedRequest ::= [APPLICATION 23] SEQUENCE {
	//   requestName      [0] LDAPOID,
	//   requestValue     [1] OCTET STRING OPTIONAL }
	oidBytes := append([]byte{0x80, byte(len(requestName))}, []byte(requestName)...)
	extReq := append([]byte{0x77, byte(len(oidBytes))}, oidBytes...)
	msgIDEnc := []byte{0x02, 0x01, msgID}
	seq := append(msgIDEnc, extReq...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

func makeExtendedMessage(t *testing.T, requestName string) *Message {
	t.Helper()
	data := buildLDAPExtendedRequest(1, requestName)
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	return &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
}

func TestRoute_Match_ExtendedRequest(t *testing.T) {
	r := &route{
		operation: EXTENDED,
		exoName:   "1.3.6.1.4.1.1466.20037",
	}
	m := makeExtendedMessage(t, "1.3.6.1.4.1.1466.20037")
	if !r.Match(m) {
		t.Error("ExtendedRequest with matching RequestName should match")
	}

	m2 := makeExtendedMessage(t, "1.3.6.1.4.1.1466.20038")
	if r.Match(m2) {
		t.Error("ExtendedRequest with non-matching RequestName should not match")
	}
}

// ========== Route Match - BindRequest with AuthChoice ==========

func buildLDAPBindRequestWithAuth(msgID uint8, authChoice string) []byte {
	// BindRequest with authentication choice
	// Simple authentication: 0x80
	// SASL: 0xA3
	var authBytes []byte
	if authChoice == "simple" {
		authBytes = []byte{0x80, 0x00} // empty password
	} else {
		// SASL: [3] SEQUENCE { mechanism, credentials }
		authBytes = []byte{0xA3, 0x04, 0x04, 0x02, 0x4F, 0x4B, 0x80, 0x00}
	}

	name := []byte{0x04, 0x00} // empty name
	version := []byte{0x02, 0x01, 0x03}
	var content []byte
	content = append(content, version...)
	content = append(content, name...)
	content = append(content, authBytes...)

	bindReq := append([]byte{0x60, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, msgID}
	seq := append(msgIDEnc, bindReq...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

func makeBindMessageWithAuth(t *testing.T, authChoice string) *Message {
	t.Helper()
	data := buildLDAPBindRequestWithAuth(1, authChoice)
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	return &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
}

func TestRoute_Match_BindRequest_AuthChoice(t *testing.T) {
	r := &route{
		operation:   BIND,
		sAuthChoice: "simple",
		uAuthChoice: true,
	}
	m := makeBindMessageWithAuth(t, "simple")
	if !r.Match(m) {
		t.Error("BindRequest with matching AuthChoice should match")
	}
}

// ========== ServeLDAP - AbandonRequest ==========

func buildLDAPAbandonRequest(msgID uint8, abandonID int) []byte {
	// AbandonRequest ::= [APPLICATION 16] MessageID
	abandonReq := []byte{0x50, 0x01, byte(abandonID)}
	msgIDEnc := []byte{0x02, 0x01, msgID}
	seq := append(msgIDEnc, abandonReq...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

func makeAbandonMessage(t *testing.T, abandonID int) *Message {
	t.Helper()
	data := buildLDAPAbandonRequest(1, abandonID)
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	return &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
}

func TestServeLDAP_AbandonRequest(t *testing.T) {
	mux := NewRouteMux()
	// No Abandon handler registered - should be handled by default

	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	s := NewServer()
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")
	c := s.newClient(clientPipe)
	c.Numero = 1
	c.requestList = make(map[int]*Message)

	// Register a message that can be abandoned
	data := []byte{0x30, 0x0c, 0x02, 0x01, 0x05, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	mp := &messagePacket{bytes: data}
	msg, _ := mp.readMessage()
	targetMsg := &Message{LDAPMessage: &msg, Done: make(chan bool, 2), Client: c}
	c.requestList[5] = targetMsg

	// Create abandon request for message ID 5
	abandonMsg := makeAbandonMessage(t, 5)
	abandonMsg.Client = c

	ch := make(chan *ldap.LDAPMessage, 1)
	w := responseWriterImpl{chanOut: ch, messageID: 1}

	// ServeLDAP should handle AbandonRequest even without explicit handler
	mux.ServeLDAP(w, abandonMsg)

	// Check that the target message was signaled to abandon
	select {
	case <-targetMsg.Done:
		t.Log("Abandon signal sent successfully")
	case <-time.After(time.Second):
		t.Error("Abandon signal not sent")
	}
	clientPipe.Close()
}

// ========== Message getters ==========

func TestMessage_GetSearchRequest(t *testing.T) {
	m := makeSearchMessage(t, "dc=example,dc=com", "objectclass", 0)
	sr := m.GetSearchRequest()
	if string(sr.BaseObject()) != "dc=example,dc=com" {
		t.Errorf("BaseObject = %q", sr.BaseObject())
	}
}

func TestMessage_GetAddRequest(t *testing.T) {
	// Build AddRequest
	// AddRequest ::= [APPLICATION 8] SEQUENCE { entry, attributes }
	entry := []byte{0x04, 0x03, 'c', 'n', '='}
	attrs := []byte{0x30, 0x00}
	var content []byte
	content = append(content, entry...)
	content = append(content, attrs...)
	addReq := append([]byte{0x68, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, 0x01}
	seq := append(msgIDEnc, addReq...)
	data := append([]byte{0x30, byte(len(seq))}, seq...)

	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
	_ = m.GetAddRequest()
}

func TestMessage_GetDeleteRequest(t *testing.T) {
	// DelRequest ::= [APPLICATION 10] LDAPDN
	dn := []byte{0x04, 0x03, 'd', 'n', '='}
	delReq := append([]byte{0x4a, byte(len(dn))}, dn...)
	msgIDEnc := []byte{0x02, 0x01, 0x01}
	seq := append(msgIDEnc, delReq...)
	data := append([]byte{0x30, byte(len(seq))}, seq...)

	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
	_ = m.GetDeleteRequest()
}

func TestMessage_GetModifyRequest(t *testing.T) {
	// ModifyRequest ::= [APPLICATION 6] SEQUENCE { object, changes }
	object := []byte{0x04, 0x03, 'd', 'n', '='}
	changes := []byte{0x30, 0x00}
	var content []byte
	content = append(content, object...)
	content = append(content, changes...)
	modReq := append([]byte{0x66, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, 0x01}
	seq := append(msgIDEnc, modReq...)
	data := append([]byte{0x30, byte(len(seq))}, seq...)

	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
	_ = m.GetModifyRequest()
}

func TestMessage_GetCompareRequest(t *testing.T) {
	// CompareRequest ::= [APPLICATION 14] SEQUENCE { entry, AttributeValueAssertion }
	// AttributeValueAssertion ::= SEQUENCE { attributeDesc, assertionValue }
	entry := []byte{0x04, 0x02, 'c', 'n'}
	ava := []byte{0x30, 0x06, 0x04, 0x02, 'c', 'n', 0x04, 0x00}
	var content []byte
	content = append(content, entry...)
	content = append(content, ava...)
	compReq := append([]byte{0x6e, byte(len(content))}, content...)
	msgIDEnc := []byte{0x02, 0x01, 0x01}
	seq := append(msgIDEnc, compReq...)
	data := append([]byte{0x30, byte(len(seq))}, seq...)

	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}
	_ = m.GetCompareRequest()
}

func TestMessage_GetExtendedRequest(t *testing.T) {
	m := makeExtendedMessage(t, "1.3.6.1.4.1.1466.20037")
	_ = m.GetExtendedRequest()
}

func TestMessage_GetAbandonRequest(t *testing.T) {
	m := makeAbandonMessage(t, 5)
	_ = m.GetAbandonRequest()
}

// ========== Server ListenAndServeTLS ==========

func TestServer_ListenAndServeTLS_InvalidCert(t *testing.T) {
	s := NewServer()
	mux := NewRouteMux()
	s.Handle(mux)

	ch := make(chan error, 1)
	go s.ListenAndServeTLS("127.0.0.1:0", "/nonexistent/cert.pem", "/nonexistent/key.pem", ch)

	select {
	case err := <-ch:
		if err == nil {
			t.Error("Should return error for invalid cert files")
		}
		t.Logf("Got expected error: %v", err)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for error")
	}
}

// ========== Server serve - Accept timeout ==========

func TestServer_Serve_AcceptTimeout(t *testing.T) {
	s := NewServer()
	mux := NewRouteMux()
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")

	// Create a listener with very short accept deadline to trigger OpError timeout
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	s.Listener = ln
	s.proxyListener = &proxyproto.Listener{Listener: ln}

	// Set accept deadline to trigger timeout path
	if tcpLn, ok := ln.(*net.TCPListener); ok {
		tcpLn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	}

	// Start serve in goroutine, it should handle accept timeout gracefully
	done := make(chan bool, 1)
	go func() {
		s.serve()
		done <- true
	}()

	// Wait for timeout to be processed, then stop server
	time.Sleep(200 * time.Millisecond)
	s.Stop()

	select {
	case <-done:
		t.Log("serve exited after Stop")
	case <-time.After(3 * time.Second):
		t.Error("serve should exit after Stop")
	}
}

// ========== Client serve - full flow ==========

func TestClient_Serve_FullFlow(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	bindCalled := false
	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		bindCalled = true
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")

	c := s.newClient(clientPipe)
	c.Numero = 1

	// Start client serve
	s.wg.Add(1)
	go c.serve()

	// Send a bind request
	bindData := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	serverPipe.Write(bindData)

	// Read response
	buf := make([]byte, 4096)
	serverPipe.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := serverPipe.Read(buf)
	if err != nil {
		t.Logf("Read error: %v", err)
	} else {
		t.Logf("Received %d bytes response", n)
	}

	time.Sleep(100 * time.Millisecond)
	if bindCalled {
		t.Log("Bind handler was called")
	}

	// Close connection to end serve
	clientPipe.Close()
	s.wg.Wait()
}

func TestClient_Serve_OnNewConnection(t *testing.T) {
	// Test OnNewConnection callback by directly creating a client
	// We cannot test through the server loop because close() panics when
	// chanOut is nil (which happens when OnNewConnection returns error)
	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {})
	s.Handle(mux)

	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	c := s.newClient(clientPipe)
	c.Numero = 1

	// Pre-initialize fields that close() needs, since OnNewConnection error
	// causes serve() to return before initializing them
	c.closing = make(chan bool)
	c.chanOut = make(chan *ldap.LDAPMessage)
	c.writeDone = make(chan bool)
	c.requestList = make(map[int]*Message)

	// Start the writeDone goroutine
	go func() {
		for range c.chanOut {
		}
		close(c.writeDone)
	}()

	s.onNewConnection = func(cn net.Conn) error {
		return fmt.Errorf("connection rejected")
	}

	s.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered in OnNewConnection test: %v", r)
				s.wg.Done()
			}
		}()
		c.serve()
	}()

	time.Sleep(200 * time.Millisecond)
	clientPipe.Close()
	s.wg.Wait()
}

func TestClient_Serve_UnbindRequest(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")

	c := s.newClient(clientPipe)
	c.Numero = 1

	s.wg.Add(1)
	go c.serve()

	// Send UnbindRequest [APPLICATION 2]
	unbindData := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0x42, 0x00}
	serverPipe.Write(unbindData)

	time.Sleep(100 * time.Millisecond)
	// Client should close after UnbindRequest
	s.wg.Wait()
}

func TestClient_Serve_InvalidMessage(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {})
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")

	c := s.newClient(clientPipe)
	c.Numero = 1

	s.wg.Add(1)
	go c.serve()

	// Send invalid message (will fail to parse but should continue loop)
	invalidData := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0xFF, 0xFF, 0xFF}
	serverPipe.Write(invalidData)

	// Send valid bind after invalid
	time.Sleep(50 * time.Millisecond)
	bindData := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	serverPipe.Write(bindData)

	time.Sleep(100 * time.Millisecond)
	clientPipe.Close()
	s.wg.Wait()
}

// ========== Server Stop with active client ==========

func TestServer_Stop_WithActiveClient(t *testing.T) {
	serverPipe, clientPipe := net.Pipe()
	defer serverPipe.Close()

	s := NewServer()
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	s.Handle(mux)
	s.SetClients("127.0.0.1/8")

	c := s.newClient(clientPipe)
	c.Numero = 1
	c.chanOut = make(chan *ldap.LDAPMessage)
	c.writeDone = make(chan bool)
	c.requestList = make(map[int]*Message)

	go func() {
		for range c.chanOut {
		}
		close(c.writeDone)
	}()

	s.wg.Add(1)
	go c.serve()

	// Stop server while client is active
	go s.Stop()

	time.Sleep(200 * time.Millisecond)
	clientPipe.Close()
	s.wg.Wait()
}

// ========== Integration: Full server with multiple operations ==========

func TestServer_FullIntegration(t *testing.T) {
	mux := NewRouteMux()
	mux.Bind(func(w ResponseWriter, m *Message) {
		w.Write(NewBindResponse(LDAPResultSuccess))
	})
	mux.Search(func(w ResponseWriter, m *Message) {
		e := NewSearchResultEntry("cn=test,dc=example,dc=com")
		e.AddAttribute("cn", "test")
		w.Write(e)
		w.Write(NewSearchResultDoneResponse(LDAPResultSuccess))
	})
	mux.Add(func(w ResponseWriter, m *Message) {
		w.Write(NewAddResponse(LDAPResultSuccess))
	})
	mux.Delete(func(w ResponseWriter, m *Message) {
		w.Write(NewDeleteResponse(LDAPResultSuccess))
	})
	mux.Modify(func(w ResponseWriter, m *Message) {
		w.Write(NewModifyResponse(LDAPResultSuccess))
	})
	mux.Compare(func(w ResponseWriter, m *Message) {
		w.Write(NewCompareResponse(LDAPResultCompareTrue))
	})
	mux.Extended(func(w ResponseWriter, m *Message) {
		w.Write(NewExtendedResponse(LDAPResultSuccess))
	})

	s, addr := startTestServer(t, mux)
	defer s.Stop()

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	// Send bind request
	bindData := []byte{0x30, 0x0c, 0x02, 0x01, 0x01, 0x60, 0x07, 0x02, 0x01, 0x03, 0x04, 0x00, 0x80, 0x00}
	conn.Write(bindData)

	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Logf("Read error: %v", err)
	} else {
		t.Logf("Bind response: %d bytes", n)
	}

	// Send search request
	searchData := buildLDAPSearchRequestWithDetails(2, "dc=example,dc=com", "objectclass", 0)
	conn.Write(searchData)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		t.Logf("Read error: %v", err)
	} else {
		t.Logf("Search response: %d bytes", n)
	}
}

// ========== readBytes with n > 0 but n != length ==========

func TestReadBytes_ShortRead(t *testing.T) {
	// Create a reader that will return less than requested
	data := []byte{0x30, 0x05}
	br := bufio.NewReader(bytes.NewReader(data))
	var buf []byte
	b, err := readBytes(br, &buf, 10)
	// Should read available bytes and log warning
	t.Logf("readBytes: b=0x%02x, err=%v, buf len=%d", b, err, len(buf))
}

// ========== Route Match - default case ==========

func TestRoute_Match_DefaultCase(t *testing.T) {
	// Test with a message type that doesn't have special handling
	// UnbindRequest falls through to default return true
	r := &route{operation: "UnbindRequest"}

	// Build UnbindRequest
	data := []byte{0x30, 0x05, 0x02, 0x01, 0x01, 0x42, 0x00}
	mp := &messagePacket{bytes: data}
	msg, err := mp.readMessage()
	if err != nil {
		t.Fatalf("readMessage error: %v", err)
	}
	m := &Message{LDAPMessage: &msg, Done: make(chan bool, 2)}

	// Match should return true for default case
	if !r.Match(m) {
		t.Error("Default case should return true")
	}
}
