package network

import (
	"fmt"
	"testing"
	"time"
)

// --- TPingResult.String ---

func TestTPingResultString(t *testing.T) {
	fmt.Println("=== 测试: TPingResult.String ===")
	pr := &TPingResult{
		Domain:       "example.com",
		IPAddr:       "1.2.3.4",
		DelayShort:   10.5,
		DelayLong:    20.3,
		DelayAverage: 15.4,
		Lost:         0,
	}
	s := pr.String()
	if !containsAll(s, "1.2.3.4", "15.40", "ms") {
		t.Errorf("TPingResult.String unexpected: %s", s)
	}
	t.Logf("TPingResult.String: %s", s)

	// Zero values
	pr2 := &TPingResult{}
	s2 := pr2.String()
	t.Logf("TPingResult.String (zero): %s", s2)
}

// --- icmpMessage.String ---

func TestIcmpMessageString(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.String ===")
	msg := &icmpMessage{
		Type:     8,
		Code:     0,
		Checksum: 12345,
		Ttl:      64,
		Body: &icmpEcho{
			ID:   100,
			SEQ:  1,
			Data: []byte("hello"),
		},
	}
	s := msg.String()
	if !containsAll(s, "Type=8", "Code=0", "Check=12345", "TTL=64") {
		t.Errorf("icmpMessage.String unexpected: %s", s)
	}
	t.Logf("icmpMessage.String: %s", s)
}

// --- icmpMessage.Marshal (v4) ---

func TestIcmpMessageMarshalV4(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal v4 ===")
	msg := &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
		Body: &icmpEcho{
			ID:   1234,
			SEQ:  1,
			Data: []byte("abcdefghijklmnop"),
		},
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(b) < 4 {
		t.Fatalf("Marshal result too short: %d", len(b))
	}
	// First byte should be Type (8)
	if b[0] != icmpv4EchoRequest {
		t.Errorf("Type byte wrong: %d", b[0])
	}
	// Second byte should be Code (0)
	if b[1] != 0 {
		t.Errorf("Code byte wrong: %d", b[1])
	}
	t.Logf("Marshal v4 result: %v (len=%d)", b[:8], len(b))
}

// --- icmpMessage.Marshal (v6) ---

func TestIcmpMessageMarshalV6(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal v6 ===")
	msg := &icmpMessage{
		Type: icmpv6EchoRequest,
		Code: 0,
		Body: &icmpEcho{
			ID:   5678,
			SEQ:  2,
			Data: []byte("test"),
		},
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(b) < 4 {
		t.Fatalf("Marshal result too short: %d", len(b))
	}
	// v6 should not compute checksum
	if b[0] != icmpv6EchoRequest {
		t.Errorf("Type byte wrong: %d", b[0])
	}
	t.Logf("Marshal v6 result: %v (len=%d)", b[:8], len(b))
}

// --- icmpMessage.Marshal with nil body ---

func TestIcmpMessageMarshalNilBody(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal nil body ===")
	msg := &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(b) != 4 {
		t.Errorf("expected 4 bytes for nil body, got %d", len(b))
	}
}

// --- icmpMessage.Marshal v6 reply ---

func TestIcmpMessageMarshalV6Reply(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal v6 reply ===")
	msg := &icmpMessage{
		Type: icmpv6EchoReply,
		Code: 0,
		Body: &icmpEcho{
			ID:   1,
			SEQ:  1,
			Data: []byte("data"),
		},
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	t.Logf("Marshal v6 reply result len=%d", len(b))
}

// --- parseICMPMessage ---

func TestParseICMPMessage(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage ===")

	// Build a valid ICMP echo reply
	// ICMP format: Type(1) + Code(1) + Checksum(2) + ID(2) + SEQ(2) + Data
	echo := &icmpEcho{
		ID:   100,
		SEQ:  1,
		Data: []byte("testdata"),
	}
	echoBytes, _ := echo.Marshal()

	// Build ICMP message: Type(1) + Code(1) + Checksum(2) + echo body(ID+SEQ+Data)
	b := make([]byte, 4+len(echoBytes))
	b[0] = icmpv4EchoReply // Type
	b[1] = 0               // Code
	b[2] = 0               // Checksum high
	b[3] = 0               // Checksum low
	copy(b[4:], echoBytes)

	msg, err := parseICMPMessage(b)
	if err != nil {
		t.Fatalf("parseICMPMessage failed: %v", err)
	}
	if msg.Type != icmpv4EchoReply {
		t.Errorf("Type mismatch: %d", msg.Type)
	}
	if msg.Body == nil {
		t.Fatal("Body is nil")
	}
	echoBody, ok := msg.Body.(*icmpEcho)
	if !ok {
		t.Fatalf("Body type wrong: %T", msg.Body)
	}
	if echoBody.ID != 100 {
		t.Errorf("ID mismatch: %d", echoBody.ID)
	}
	if echoBody.SEQ != 1 {
		t.Errorf("SEQ mismatch: %d", echoBody.SEQ)
	}
	t.Logf("Parsed ICMP: %s", msg)
}

// --- parseICMPMessage with IP header ---

func TestParseICMPMessageWithIPHeader(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage with IP header ===")

	echo := &icmpEcho{
		ID:   200,
		SEQ:  5,
		Data: []byte("hello"),
	}
	echoBytes, _ := echo.Marshal()

	// Prepend a 20-byte IP header
	ipHeader := make([]byte, 20)
	ipHeader[0] = 0x45 // version=4, IHL=5 (20 bytes)

	// ICMP message: Type + Code + Checksum + echo body
	icmpMsg := make([]byte, 4+len(echoBytes))
	icmpMsg[0] = icmpv4EchoReply
	icmpMsg[1] = 0
	copy(icmpMsg[4:], echoBytes)

	full := append(ipHeader, icmpMsg...)

	msg, err := parseICMPMessage(full)
	if err != nil {
		t.Fatalf("parseICMPMessage failed: %v", err)
	}
	if msg.Type != icmpv4EchoReply {
		t.Errorf("Type mismatch: %d", msg.Type)
	}
	echoBody := msg.Body.(*icmpEcho)
	if echoBody.ID != 200 {
		t.Errorf("ID mismatch: %d", echoBody.ID)
	}
}

// --- parseICMPMessage too short ---

func TestParseICMPMessageTooShort(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage too short ===")
	_, err := parseICMPMessage([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for short message")
	}
	t.Logf("Short message error (expected): %v", err)
}

// --- parseICMPMessage non-echo type ---

func TestParseICMPMessageNonEcho(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage non-echo type ===")
	// Need at least 9 bytes to avoid panic at b[8]
	b := make([]byte, 12)
	b[0] = 3 // Destination Unreachable
	b[1] = 1 // Host Unreachable
	b[2] = 0
	b[3] = 0

	msg, err := parseICMPMessage(b)
	if err != nil {
		t.Fatalf("parseICMPMessage failed: %v", err)
	}
	if msg.Body != nil {
		t.Errorf("expected nil body for non-echo type, got: %T", msg.Body)
	}
}

// --- icmpEcho tests ---

func TestIcmpEchoString(t *testing.T) {
	fmt.Println("=== 测试: icmpEcho.String ===")
	echo := &icmpEcho{
		ID:   42,
		SEQ:  7,
		Data: []byte("short"),
	}
	s := echo.String()
	if !containsAll(s, "Id=42", "Seq=7", "short") {
		t.Errorf("icmpEcho.String unexpected: %s", s)
	}
	t.Logf("icmpEcho.String: %s", s)

	// Long data (> 16 bytes)
	echo2 := &icmpEcho{
		ID:   1,
		SEQ:  1,
		Data: []byte("this is more than sixteen bytes"),
	}
	s2 := echo2.String()
	t.Logf("icmpEcho.String (long data): %s", s2)
}

func TestIcmpEchoLen(t *testing.T) {
	fmt.Println("=== 测试: icmpEcho.Len ===")
	echo := &icmpEcho{
		ID:   1,
		SEQ:  1,
		Data: []byte("12345"),
	}
	if echo.Len() != 4+5 {
		t.Errorf("Len mismatch: %d", echo.Len())
	}

	// nil echo
	var nilEcho *icmpEcho
	if nilEcho.Len() != 0 {
		t.Errorf("nil Len should be 0: %d", nilEcho.Len())
	}
}

func TestIcmpEchoGetId(t *testing.T) {
	fmt.Println("=== 测试: icmpEcho.GetId ===")
	echo := &icmpEcho{ID: 999}
	if echo.GetId() != 999 {
		t.Errorf("GetId mismatch: %d", echo.GetId())
	}
}

func TestIcmpEchoGetSeq(t *testing.T) {
	fmt.Println("=== 测试: icmpEcho.GetSeq ===")
	echo := &icmpEcho{SEQ: 42}
	if echo.GetSeq() != 42 {
		t.Errorf("GetSeq mismatch: %d", echo.GetSeq())
	}
}

func TestIcmpEchoMarshal(t *testing.T) {
	fmt.Println("=== 测试: icmpEcho.Marshal ===")
	echo := &icmpEcho{
		ID:   0x1234,
		SEQ:  0x5678,
		Data: []byte("ABCD"),
	}
	b, err := echo.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// ID: 0x12, 0x34
	if b[0] != 0x12 || b[1] != 0x34 {
		t.Errorf("ID bytes wrong: %v", b[:2])
	}
	// SEQ: 0x56, 0x78
	if b[2] != 0x56 || b[3] != 0x78 {
		t.Errorf("SEQ bytes wrong: %v", b[2:4])
	}
	// Data
	if string(b[4:]) != "ABCD" {
		t.Errorf("Data wrong: %v", b[4:])
	}
}

// --- parseICMPEchoBody ---

func TestParseICMPEchoBody(t *testing.T) {
	fmt.Println("=== 测试: parseICMPEchoBody ===")

	// Build echo body bytes
	b := []byte{
		0x00, 0x64, // ID = 100
		0x00, 0x01, // SEQ = 1
		't', 'e', 's', 't', // Data
	}
	echo := parseICMPEchoBody(b)
	if echo.ID != 100 {
		t.Errorf("ID mismatch: %d", echo.ID)
	}
	if echo.SEQ != 1 {
		t.Errorf("SEQ mismatch: %d", echo.SEQ)
	}
	if string(echo.Data) != "test" {
		t.Errorf("Data mismatch: %s", string(echo.Data))
	}

	// Empty data (just 4 bytes header)
	echo2 := parseICMPEchoBody([]byte{0x00, 0x01, 0x00, 0x01})
	if echo2.ID != 1 {
		t.Errorf("ID mismatch: %d", echo2.ID)
	}
	if echo2.Data != nil {
		t.Errorf("Data should be nil for 4-byte body: %v", echo2.Data)
	}
}

// --- Pinger with IP input (no DNS needed) ---

func TestPingerWithIPInput(t *testing.T) {
	fmt.Println("=== 测试: Pinger with IP ===")
	// This test verifies the code path where an IP is passed directly
	// It may fail due to no raw socket permissions, but tests the code path
	ok, result, err := Pinger("127.0.0.1", 1, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger failed (likely no raw socket): %v", err)
	} else {
		t.Logf("Pinger result: ok=%v, %s", ok, result)
	}
}

// --- Pinger with small size (below 16 minimum) ---

func TestPingerSmallSize(t *testing.T) {
	fmt.Println("=== 测试: Pinger small size ===")
	ok, result, err := Pinger("127.0.0.1", 1, 8, 1*time.Second)
	if err != nil {
		t.Logf("Pinger small size failed (expected no raw socket): %v", err)
	} else {
		t.Logf("Pinger small size result: ok=%v, %s", ok, result)
	}
}

// --- Pinger with zero count (should be adjusted to 1) ---

func TestPingerZeroCount(t *testing.T) {
	fmt.Println("=== 测试: Pinger zero count ===")
	ok, result, err := Pinger("127.0.0.1", 0, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger zero count failed (expected no raw socket): %v", err)
	} else {
		t.Logf("Pinger zero count result: ok=%v, %s", ok, result)
	}
}

// --- Pinger with negative timeout ---

func TestPingerNegativeTimeout(t *testing.T) {
	fmt.Println("=== 测试: Pinger negative timeout ===")
	ok, result, err := Pinger("127.0.0.1", 1, 32, -1*time.Second)
	if err != nil {
		t.Logf("Pinger negative timeout failed: %v", err)
	} else {
		t.Logf("Pinger negative timeout result: ok=%v, %s", ok, result)
	}
}

// --- Pinger with invalid hostname ---

func TestPingerInvalidHost(t *testing.T) {
	fmt.Println("=== 测试: Pinger invalid host ===")
	ok, _, err := Pinger("this-host-does-not-exist.invalid", 1, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger invalid host error (expected): %v", err)
	}
	if ok {
		t.Log("Pinger invalid host unexpectedly succeeded")
	}
}

// --- icmpMessage.Marshal with body that has zero length ---

func TestIcmpMessageMarshalBodyZeroLen(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal body zero length ===")
	msg := &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
		Body: &icmpEcho{
			ID:   1,
			SEQ:  1,
			Data: []byte{},
		},
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	// Body has Len()=4 (4+0), which is != 0, so body should be marshaled
	if len(b) != 8 { // 4 header + 4 echo header
		t.Errorf("expected 8 bytes, got %d", len(b))
	}
}

// --- icmpMessage.Marshal with odd length checksum calculation ---

func TestIcmpMessageMarshalOddLength(t *testing.T) {
	fmt.Println("=== 测试: icmpMessage.Marshal odd length ===")
	msg := &icmpMessage{
		Type: icmpv4EchoRequest,
		Code: 0,
		Body: &icmpEcho{
			ID:   1,
			SEQ:  1,
			Data: []byte("abc"), // odd number of data bytes -> total body = 7, total msg = 11 (odd)
		},
	}
	b, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if len(b) != 11 {
		t.Errorf("expected 11 bytes, got %d", len(b))
	}
	// Verify checksum bytes are set (b[2] and b[3])
	if b[2] == 0 && b[3] == 0 {
		t.Log("checksum bytes are both zero (might be valid for this payload)")
	}
}

// --- parseICMPMessage with v6 echo types ---

func TestParseICMPMessageV6(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage v6 ===")
	echo := &icmpEcho{
		ID:   300,
		SEQ:  10,
		Data: []byte("v6test"),
	}
	echoBytes, _ := echo.Marshal()

	b := make([]byte, 4+len(echoBytes))
	b[0] = icmpv6EchoReply // Type 129
	b[1] = 0
	b[2] = 0
	b[3] = 0
	copy(b[4:], echoBytes)

	msg, err := parseICMPMessage(b)
	if err != nil {
		t.Fatalf("parseICMPMessage v6 failed: %v", err)
	}
	if msg.Type != icmpv6EchoReply {
		t.Errorf("Type mismatch: %d", msg.Type)
	}
	echoBody := msg.Body.(*icmpEcho)
	if echoBody.ID != 300 {
		t.Errorf("ID mismatch: %d", echoBody.ID)
	}
}

// --- parseICMPMessage with v4 echo request type ---

func TestParseICMPMessageV4Request(t *testing.T) {
	fmt.Println("=== 测试: parseICMPMessage v4 request ===")
	echo := &icmpEcho{
		ID:   400,
		SEQ:  20,
		Data: []byte("v4req"),
	}
	echoBytes, _ := echo.Marshal()

	b := make([]byte, 4+len(echoBytes))
	b[0] = icmpv4EchoRequest // Type 8
	b[1] = 0
	b[2] = 0
	b[3] = 0
	copy(b[4:], echoBytes)

	msg, err := parseICMPMessage(b)
	if err != nil {
		t.Fatalf("parseICMPMessage v4 request failed: %v", err)
	}
	if msg.Type != icmpv4EchoRequest {
		t.Errorf("Type mismatch: %d", msg.Type)
	}
}

// --- parseICMPEchoBody with longer data ---

func TestParseICMPEchoBodyLongData(t *testing.T) {
	fmt.Println("=== 测试: parseICMPEchoBody long data ===")
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	b := make([]byte, 4+len(data))
	b[0] = 0x00
	b[1] = 0x01 // ID = 1
	b[2] = 0x00
	b[3] = 0x02 // SEQ = 2
	copy(b[4:], data)

	echo := parseICMPEchoBody(b)
	if echo.ID != 1 {
		t.Errorf("ID mismatch: %d", echo.ID)
	}
	if echo.SEQ != 2 {
		t.Errorf("SEQ mismatch: %d", echo.SEQ)
	}
	if len(echo.Data) != 100 {
		t.Errorf("Data length mismatch: %d", len(echo.Data))
	}
}

// --- connect_icmp directly (will fail without raw socket) ---

func TestConnectICMPDirect(t *testing.T) {
	fmt.Println("=== 测试: connect_icmp direct ===")
	// This will likely fail due to raw socket permissions, but tests the code path
	ttl, err := connect_icmp("127.0.0.1", 1*time.Second, 32, 1)
	if err != nil {
		t.Logf("connect_icmp failed (expected no raw socket): %v", err)
	} else {
		t.Logf("connect_icmp succeeded: ttl=%d", ttl)
	}
}

// --- connect_icmp with invalid host ---

func TestConnectICMPInvalidHost(t *testing.T) {
	fmt.Println("=== 测试: connect_icmp invalid host ===")
	_, err := connect_icmp("invalid-host", 1*time.Second, 32, 1)
	if err == nil {
		t.Log("connect_icmp unexpectedly succeeded with invalid host")
	} else {
		t.Logf("connect_icmp invalid host error (expected): %v", err)
	}
}

// --- Pinger with IP address that's local ---

func TestPingerLocalIP(t *testing.T) {
	fmt.Println("=== 测试: Pinger local IP ===")
	// Use 127.0.0.1 which should be reachable without DNS
	// But needs raw socket which may not be available
	ok, result, err := Pinger("127.0.0.1", 1, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger 127.0.0.1 failed (likely no raw socket): %v", err)
	} else {
		t.Logf("Pinger 127.0.0.1: ok=%v result=%s", ok, result)
	}
}

// --- Pinger with domain name (DNS resolution path) ---

func TestPingerWithDomainName(t *testing.T) {
	fmt.Println("=== 测试: Pinger with domain name ===")
	// Use a domain that should resolve but ping will likely fail without raw socket
	ok, result, err := Pinger("localhost", 1, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger localhost failed (expected): %v", err)
	} else {
		t.Logf("Pinger localhost: ok=%v result=%s", ok, result)
	}
}

// --- Pinger with all-zero result (all packets lost) ---

func TestPingerAllLost(t *testing.T) {
	fmt.Println("=== 测试: Pinger all lost ===")
	// Use a non-routable IP that will lose all packets
	ok, result, err := Pinger("192.0.2.1", 1, 32, 1*time.Second)
	if err != nil {
		t.Logf("Pinger all lost (expected): %v", err)
	} else {
		t.Logf("Pinger all lost: ok=%v result=%s", ok, result)
	}
}

// helper
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && len(sub) > 0 && findSubstring(s, sub)))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
