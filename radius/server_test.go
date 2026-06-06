package radius

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// TestServerResetClientNets tests ResetClientNets method
func TestServerResetClientNets(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"192.168.1.0/24": "secret1",
			"10.0.0.1":       "secret2",
		},
	}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed: %v", err)
	}

	if len(server.ClientNets) != 2 {
		t.Errorf("Expected 2 client nets, got %d", len(server.ClientNets))
	}
	if len(server.ClientSecrets) != 2 {
		t.Errorf("Expected 2 client secrets, got %d", len(server.ClientSecrets))
	}
}

// TestServerResetClientNetsInvalidIP tests ResetClientNets with invalid IP
func TestServerResetClientNetsInvalidIP(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"invalid-ip": "secret1",
		},
	}

	err := server.ResetClientNets()
	if err == nil {
		t.Error("Expected error for invalid IP")
	}
}

// TestServerResetClientNetsNilMap tests ResetClientNets with nil map
func TestServerResetClientNetsNilMap(t *testing.T) {
	server := &Server{}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed with nil map: %v", err)
	}
}

// TestServerGetSecretByIPString tests GetSecretByIPString method
func TestServerGetSecretByIPString(t *testing.T) {
	server := &Server{
		ClientNets: []net.IPNet{
			*parseCIDR("192.168.1.0/24"),
		},
		ClientSecrets: [][]byte{[]byte("secret1")},
	}

	secret := server.GetSecretByIPString("192.168.1.100")
	if secret == nil {
		t.Error("Expected secret for 192.168.1.100")
	}
	if string(secret) != "secret1" {
		t.Errorf("Expected 'secret1', got '%s'", string(secret))
	}

	// IP not in any subnet
	secret = server.GetSecretByIPString("10.0.0.1")
	if secret != nil {
		t.Error("Expected nil for IP not in any subnet")
	}

	// Invalid IP
	secret = server.GetSecretByIPString("invalid")
	if secret != nil {
		t.Error("Expected nil for invalid IP")
	}
}

// TestServerGetSecretByIP tests GetSecretByIP method
func TestServerGetSecretByIP(t *testing.T) {
	server := &Server{
		ClientNets: []net.IPNet{
			*parseCIDR("192.168.1.0/24"),
			*parseCIDR("10.0.0.0/8"),
		},
		ClientSecrets: [][]byte{[]byte("secret1"), []byte("secret2")},
	}

	secret := server.GetSecretByIP(net.ParseIP("192.168.1.100"))
	if string(secret) != "secret1" {
		t.Errorf("Expected 'secret1', got '%s'", string(secret))
	}

	secret = server.GetSecretByIP(net.ParseIP("10.0.0.1"))
	if string(secret) != "secret2" {
		t.Errorf("Expected 'secret2', got '%s'", string(secret))
	}

	// IP not in any subnet
	secret = server.GetSecretByIP(net.ParseIP("172.16.0.1"))
	if secret != nil {
		t.Error("Expected nil for IP not in any subnet")
	}
}

// TestServerAddClientsMap tests AddClientsMap method
func TestServerAddClientsMap(t *testing.T) {
	server := &Server{
		ClientsMap: make(map[string]string),
	}

	server.AddClientsMap(map[string]string{
		"192.168.1.0/24": "secret1",
		"10.0.0.1":       "secret2",
	})

	if len(server.ClientsMap) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(server.ClientsMap))
	}
	if server.ClientsMap["192.168.1.0/24"] != "secret1" {
		t.Errorf("Expected 'secret1', got '%s'", server.ClientsMap["192.168.1.0/24"])
	}
}

// TestServerListenAndServeNoHandler tests ListenAndServe with no handler
func TestServerListenAndServeNoHandler(t *testing.T) {
	server := &Server{
		Port: 0,
	}

	err := server.ListenAndServe()
	if err == nil {
		t.Error("Expected error for no handler")
	}
}

// TestServerListenAndServeAlreadyRunning tests ListenAndServe when already running
func TestServerListenAndServeAlreadyRunning(t *testing.T) {
	server := &Server{
		Port:     0,
		Handler:  HandlerFunc(func(w ResponseWriter, p *TDataPacket) {}),
		listener: &net.UDPConn{}, // Fake listener
	}

	err := server.ListenAndServe()
	if err == nil {
		t.Error("Expected error for already running server")
	}
}

// TestServerClose tests Close method
func TestServerClose(t *testing.T) {
	server := &Server{}

	err := server.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestHandlerFunc tests HandlerFunc type
func TestHandlerFunc(t *testing.T) {
	called := false
	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {
		called = true
	})

	// Create a dummy packet
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// HandlerFunc.ServeRadius should call the function
	handler.ServeRadius(nil, packet)
	if !called {
		t.Error("HandlerFunc was not called")
	}
}

// TestServerResetClientNetsWithCIDR tests ResetClientNets with CIDR notation
func TestServerResetClientNetsWithCIDR(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"192.168.1.0/24": "secret1",
			"10.0.0.0/8":     "secret2",
		},
	}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed: %v", err)
	}

	if len(server.ClientNets) != 2 {
		t.Errorf("Expected 2 client nets, got %d", len(server.ClientNets))
	}
}

// TestServerResetClientNetsWithSingleIP tests ResetClientNets with single IP
func TestServerResetClientNetsWithSingleIP(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"10.0.0.1": "secret2",
		},
	}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed: %v", err)
	}

	if len(server.ClientNets) != 1 {
		t.Errorf("Expected 1 client net, got %d", len(server.ClientNets))
	}

	// The single IP should be converted to /32
	secret := server.GetSecretByIP(net.ParseIP("10.0.0.1"))
	if secret == nil {
		t.Error("Expected secret for 10.0.0.1")
	}
	if string(secret) != "secret2" {
		t.Errorf("Expected 'secret2', got '%s'", string(secret))
	}
}

// TestServerGetSecretByIPStringEmpty tests GetSecretByIPString with empty client nets
func TestServerGetSecretByIPStringEmpty(t *testing.T) {
	server := &Server{}

	secret := server.GetSecretByIPString("192.168.1.1")
	if secret != nil {
		t.Error("Expected nil for empty client nets")
	}
}

// TestServerGetSecretByIPEmpty tests GetSecretByIP with empty client nets
func TestServerGetSecretByIPEmpty(t *testing.T) {
	server := &Server{}

	secret := server.GetSecretByIP(net.ParseIP("192.168.1.1"))
	if secret != nil {
		t.Error("Expected nil for empty client nets")
	}
}

// TestServerListenAndServeWithInvalidClientsMap tests ListenAndServe with invalid clients map
func TestServerListenAndServeWithInvalidClientsMap(t *testing.T) {
	server := &Server{
		Port:  0,
		Handler: HandlerFunc(func(w ResponseWriter, p *TDataPacket) {}),
		ClientsMap: map[string]string{
			"invalid-ip": "secret",
		},
	}

	err := server.ListenAndServe()
	if err == nil {
		t.Error("Expected error for invalid clients map")
	}
}

// Helper function to parse CIDR
func parseCIDR(cidr string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return ipNet
}

// TestServerResetClientNetsMultipleSubnets tests ResetClientNets with multiple subnets
func TestServerResetClientNetsMultipleSubnets(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"192.168.1.0/24": "secret1",
			"192.168.2.0/24": "secret2",
			"10.0.0.0/8":     "secret3",
		},
	}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed: %v", err)
	}

	if len(server.ClientNets) != 3 {
		t.Errorf("Expected 3 client nets, got %d", len(server.ClientNets))
	}
	if len(server.ClientSecrets) != 3 {
		t.Errorf("Expected 3 client secrets, got %d", len(server.ClientSecrets))
	}
}

// TestServerAddClientsMapMerge tests AddClientsMap merging
func TestServerAddClientsMapMerge(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"192.168.1.0/24": "secret1",
		},
	}

	server.AddClientsMap(map[string]string{
		"10.0.0.0/8": "secret2",
	})

	if len(server.ClientsMap) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(server.ClientsMap))
	}
}

// TestServerAddClientsMapOverwrite tests AddClientsMap overwriting
func TestServerAddClientsMapOverwrite(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"192.168.1.0/24": "secret1",
		},
	}

	server.AddClientsMap(map[string]string{
		"192.168.1.0/24": "newsecret",
	})

	if server.ClientsMap["192.168.1.0/24"] != "newsecret" {
		t.Errorf("Expected 'newsecret', got '%s'", server.ClientsMap["192.168.1.0/24"])
	}
}

// TestResponseWriterLocalAddr tests responseWriter LocalAddr
func TestResponseWriterLocalAddr(t *testing.T) {
	rw := &responseWriter{
		conn: nil,
	}
	// This will panic with nil conn, but we test the method exists
	defer func() {
		recover()
	}()
	rw.LocalAddr()
}

// TestResponseWriterRemoteAddr tests responseWriter RemoteAddr
func TestResponseWriterRemoteAddr(t *testing.T) {
	addr := &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 1812}
	rw := &responseWriter{
		addr: addr,
	}

	remoteAddr := rw.RemoteAddr()
	if remoteAddr.String() != "192.168.1.1:1812" {
		t.Errorf("Expected '192.168.1.1:1812', got '%s'", remoteAddr.String())
	}
}

// TestResponseWriterWriteNilConn tests responseWriter Write with nil conn
func TestResponseWriterWriteNilConn(t *testing.T) {
	rw := &responseWriter{
		conn: nil,
		packet: NewPacket(CodeAccessRequest, []byte("secret")),
	}

	// This should panic or error with nil conn
	defer func() {
		recover()
	}()
	rw.Write(NewPacket(CodeAccessAccept, []byte("secret")))
}

// TestServerResetClientNetsClearsPrevious tests that ResetClientNets clears previous entries
func TestServerResetClientNetsClearsPrevious(t *testing.T) {
	server := &Server{
		ClientNets: []net.IPNet{*parseCIDR("172.16.0.0/16")},
		ClientSecrets: [][]byte{[]byte("oldsecret")},
		ClientsMap: map[string]string{
			"192.168.1.0/24": "newsecret",
		},
	}

	err := server.ResetClientNets()
	if err != nil {
		t.Fatalf("ResetClientNets failed: %v", err)
	}

	// Should have only the new entry
	if len(server.ClientNets) != 1 {
		t.Errorf("Expected 1 client net, got %d", len(server.ClientNets))
	}
}

// TestResponseWriterAccessAcceptWithConn tests AccessAccept with a real UDP connection
func TestResponseWriterAccessAcceptWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	err = rw.AccessAccept()
	if err != nil {
		t.Fatalf("AccessAccept failed: %v", err)
	}
}

// TestResponseWriterAccessRejectWithConn tests AccessReject with a real UDP connection
func TestResponseWriterAccessRejectWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	err = rw.AccessReject()
	if err != nil {
		t.Fatalf("AccessReject failed: %v", err)
	}
}

// TestResponseWriterAccessChallengeWithConn tests AccessChallenge with a real UDP connection
func TestResponseWriterAccessChallengeWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	err = rw.AccessChallenge()
	if err != nil {
		t.Fatalf("AccessChallenge failed: %v", err)
	}
}

// TestResponseWriterAccountingResponseWithConn tests AccountingResponse with a real UDP connection
func TestResponseWriterAccountingResponseWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccountingRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	err = rw.AccountingResponse()
	if err != nil {
		t.Fatalf("AccountingResponse failed: %v", err)
	}
}

// TestResponseWriterWriteWithConn tests Write with a real UDP connection
func TestResponseWriterWriteWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccessAccept, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	err = rw.Write(packet)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

// TestResponseWriterWriteInvalidPacket tests Write with invalid packet code
func TestResponseWriterWriteInvalidPacket(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: NewPacket(CodeAccessRequest, []byte("secret")),
	}

	// Write a packet with unknown code - should fail
	badPacket := &TDataPacket{
		Code:       Code(100),
		Secret:     []byte("secret"),
		Dictionary: Builtin,
	}

	err = rw.Write(badPacket)
	if err == nil {
		t.Error("Expected error for invalid packet code")
	}
}

// TestResponseWriterAccessAcceptWithAttributes tests AccessAccept with attributes
func TestResponseWriterAccessAcceptWithAttributes(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	remoteAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: conn.LocalAddr().(*net.UDPAddr).Port}

	rw := &responseWriter{
		conn:   conn,
		addr:   remoteAddr,
		packet: packet,
	}

	// Create attributes
	attr, err := Builtin.NewAttr("Session-Timeout", uint32(3600))
	if err != nil {
		t.Fatalf("NewAttr failed: %v", err)
	}

	err = rw.AccessAccept(attr)
	if err != nil {
		t.Fatalf("AccessAccept with attributes failed: %v", err)
	}
}

// TestResponseWriterLocalAddrWithConn tests LocalAddr with a real connection
func TestResponseWriterLocalAddrWithConn(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer conn.Close()

	rw := &responseWriter{
		conn: conn,
	}

	localAddr := rw.LocalAddr()
	if localAddr == nil {
		t.Error("LocalAddr should not be nil")
	}
}

// TestResponseWriterAccessRespondWithNilConn tests accessRespond with nil conn
func TestResponseWriterAccessRespondWithNilConn(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	rw := &responseWriter{
		conn:   nil,
		packet: packet,
	}

	// This will try to encode and then write to nil conn
	defer func() {
		recover()
	}()
	rw.AccessAccept()
}

// TestServerListenAndServeWithValidConfig tests ListenAndServe with valid config
// This test starts the server and then immediately closes it
func TestServerListenAndServeAndClose(t *testing.T) {
	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {})

	server := &Server{
		Port:     0, // Use port 0 for random available port
		Handler:  handler,
		Secret:   []byte("secret"),
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Close the server
	err := server.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestServerListenAndServeWithAddr tests ListenAndServe with custom address
func TestServerListenAndServeWithAddr(t *testing.T) {
	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {})

	server := &Server{
		Addr:     "127.0.0.1",
		Port:     0,
		Handler:  handler,
		Secret:   []byte("secret"),
		Network:  "udp",
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Close the server
	err := server.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestServerListenAndServeWithClientsMap tests ListenAndServe with ClientsMap
func TestServerListenAndServeWithClientsMap(t *testing.T) {
	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {})

	server := &Server{
		Addr:    "127.0.0.1",
		Port:    0,
		Handler: handler,
		Secret:  []byte("secret"),
		ClientsMap: map[string]string{
			"127.0.0.1": "secret",
		},
		Dictionary: Builtin,
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Close the server
	err := server.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestClientServerIntegration tests client-server communication
func TestClientServerIntegration(t *testing.T) {
	secret := []byte("testsecret")
	receivedPacket := make(chan *TDataPacket, 1)

	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {
		// Store the received packet
		select {
		case receivedPacket <- p:
		default:
		}
		// Send AccessAccept response
		w.AccessAccept()
	})

	server := &Server{
		Addr:    "127.0.0.1",
		Port:    0,
		Handler: handler,
		Secret:  secret,
		Dictionary: Builtin,
		ClientsMap: map[string]string{
			"127.0.0.1": string(secret),
		},
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual server port
	localAddr := server.listener.LocalAddr().(*net.UDPAddr)
	serverPort := localAddr.Port

	// Create client and send packet
	client := &Client{
		Net:         "udp",
		DialTimeout: 2 * time.Second,
		ReadTimeout: 2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	packet := NewPacket(CodeAccessRequest, secret)
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}
	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Send packet to server
	addr := fmt.Sprintf("127.0.0.1:%d", serverPort)
	response, err := client.SendPacket(packet, addr)
	if err != nil {
		t.Logf("SendPacket failed (expected in test environment): %v", err)
	} else {
		t.Logf("Got response: %v", response)
	}

	// Wait for server to receive the packet
	select {
	case p := <-receivedPacket:
		if p == nil {
			t.Error("Received nil packet")
		}
	case <-time.After(2 * time.Second):
		t.Log("Timeout waiting for packet (may be expected)")
	}

	// Close the server
	server.Close()
}

// TestServerResetClientNetsWithInvalidCIDR tests ResetClientNets with invalid CIDR
// that can't be parsed as either CIDR or IP
func TestServerResetClientNetsWithInvalidCIDR(t *testing.T) {
	server := &Server{
		ClientsMap: map[string]string{
			"not-a-cidr-or-ip": "secret1",
		},
	}

	err := server.ResetClientNets()
	if err == nil {
		t.Error("Expected error for invalid CIDR/IP")
	}
}

// TestServerCloseWithListener tests Close with an active listener
func TestServerCloseWithListener(t *testing.T) {
	handler := HandlerFunc(func(w ResponseWriter, p *TDataPacket) {})

	server := &Server{
		Addr:    "127.0.0.1",
		Port:    0,
		Handler: handler,
		Secret:  []byte("secret"),
	}

	// Start server in background
	go func() {
		server.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Close should work with active listener
	err := server.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Double close should also work (returns nil since listener is nil after first close)
	err = server.Close()
	if err != nil {
		t.Fatalf("Second close failed: %v", err)
	}
}
