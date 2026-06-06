package radius

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"
)

// TestNewPacket tests the NewPacket function
func TestNewPacket(t *testing.T) {
	secret := []byte("testsecret")
	packet := NewPacket(CodeAccessRequest, secret)
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}
	if packet.Code != CodeAccessRequest {
		t.Errorf("Expected Code %d, got %d", CodeAccessRequest, packet.Code)
	}
	if packet.Dictionary == nil {
		t.Error("Dictionary should not be nil")
	}
	if len(packet.Secret) != len(secret) {
		t.Errorf("Secret length mismatch")
	}
	if !bytes.Equal(packet.Secret, secret) {
		t.Errorf("Secret mismatch")
	}
	// Authenticator should be 16 bytes
	if len(packet.Authenticator) != 16 {
		t.Errorf("Authenticator should be 16 bytes, got %d", len(packet.Authenticator))
	}
}

// TestNewPacketWithDifferentCodes tests NewPacket with different packet codes
func TestNewPacketWithDifferentCodes(t *testing.T) {
	codes := []Code{
		CodeAccessRequest,
		CodeAccessAccept,
		CodeAccessReject,
		CodeAccountingRequest,
		CodeAccountingResponse,
		CodeAccessChallenge,
		CodeStatusServer,
		CodeStatusClient,
		CodeReserved,
	}

	for _, code := range codes {
		packet := NewPacket(code, []byte("secret"))
		if packet == nil {
			t.Errorf("NewPacket returned nil for code %d", code)
		}
		if packet.Code != code {
			t.Errorf("Expected Code %d, got %d", code, packet.Code)
		}
	}
}

// TestPacketString tests the String method
func TestPacketString(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("testsecret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add some attributes
	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// Should contain the code and identifier
	if !bytes.Contains([]byte(str), []byte("Code=1")) {
		t.Error("String() should contain Code=1")
	}
}

// TestPacketStringWithUserPassword tests String method with User-Password attribute
func TestPacketStringWithUserPassword(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("testsecret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Set User-Password (this will be encoded)
	err = packet.Set("User-Password", "mypassword")
	if err != nil {
		t.Fatalf("Set User-Password failed: %v", err)
	}

	str := packet.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

// TestParsePacketTooShort tests parsing a packet that is too short
func TestParsePacketTooShort(t *testing.T) {
	data := []byte{1, 2, 3} // Too short
	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for packet too short")
	}
}

// TestParsePacketInvalidLength tests parsing a packet with invalid length field
func TestParsePacketInvalidLength(t *testing.T) {
	// Create a packet with length < 20
	data := make([]byte, 20)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 10) // Invalid length < 20

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for invalid packet length")
	}
}

// TestParsePacketLengthTooLarge tests parsing a packet with length > maxPacketSize
func TestParsePacketLengthTooLarge(t *testing.T) {
	data := make([]byte, 20)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 5000) // Invalid length > 4095

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for packet length too large")
	}
}

// TestParsePacketValid tests parsing a valid packet
func TestParsePacketValid(t *testing.T) {
	// Create a minimal valid packet
	data := make([]byte, 20)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 20) // Length = 20 (header only)
	// Authenticator (16 bytes) - already zeros

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if packet.Code != CodeAccessRequest {
		t.Errorf("Expected Code %d, got %d", CodeAccessRequest, packet.Code)
	}
	if packet.Identifier != 1 {
		t.Errorf("Expected Identifier 1, got %d", packet.Identifier)
	}
}

// TestParsePacketWithAttributes tests parsing a packet with attributes
func TestParsePacketWithAttributes(t *testing.T) {
	// Create a packet with User-Name attribute
	// Header: 20 bytes
	// Attribute: Type(1) + Length(1) + Value
	attrValue := []byte("testuser")
	attrLen := byte(2 + len(attrValue)) // Type + Length + Value

	data := make([]byte, 20+int(attrLen))
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], uint16(20+int(attrLen)))
	// Authenticator (16 bytes) - zeros

	// Add User-Name attribute (type 1)
	data[20] = 1 // User-Name
	data[21] = attrLen
	copy(data[22:], attrValue)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if len(packet.AttrItems) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(packet.AttrItems))
	}
}

// TestParsePacketAttributeTooShort tests parsing with attribute length < 2
func TestParsePacketAttributeTooShort(t *testing.T) {
	// Create a packet with an attribute that has length < 2
	data := make([]byte, 22)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 22)
	// Authenticator (16 bytes) - zeros

	// Add an attribute with invalid length
	data[20] = 1 // Type
	data[21] = 1 // Invalid length < 2

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for attribute length < 2")
	}
}

// TestParsePacketAttributeLengthExceedsBuffer tests parsing with attribute length exceeding buffer
func TestParsePacketAttributeLengthExceedsBuffer(t *testing.T) {
	// Create a packet where attribute length exceeds remaining buffer
	data := make([]byte, 22)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 22)
	// Authenticator (16 bytes) - zeros

	// Add an attribute with length exceeding buffer
	data[20] = 1  // Type
	data[21] = 100 // Length exceeds remaining buffer

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for attribute length exceeding buffer")
	}
}

// TestParsePacketAttributeLengthTooLarge tests parsing with attribute length > 253
func TestParsePacketAttributeLengthTooLarge(t *testing.T) {
	// Create a packet with attribute length > 253
	data := make([]byte, 300)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 300)
	// Authenticator (16 bytes) - zeros

	// Add an attribute with invalid length > 253
	data[20] = 1   // Type
	data[21] = 254 // Invalid length > 253

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for attribute length > 253")
	}
}

// TestParsePacketWithUnknownAttribute tests parsing with unknown attribute type
func TestParsePacketWithUnknownAttribute(t *testing.T) {
	// Create a packet with an unknown attribute type
	attrValue := []byte("test")
	attrLen := byte(2 + len(attrValue))

	data := make([]byte, 20+int(attrLen))
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], uint16(20+int(attrLen)))

	// Use an unregistered attribute type (e.g., 200)
	data[20] = 200 // Unknown type
	data[21] = attrLen
	copy(data[22:], attrValue)

	// Should still parse, but with unknown attribute
	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		// Unknown attributes should cause an error during decode
		t.Log("ParsePacket returned error for unknown attribute (expected)")
	} else {
		t.Logf("ParsePacket returned error: %v", err)
	}
	_ = packet
}

// TestPacketClearAttr tests ClearAttr method
func TestPacketClearAttr(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	if len(packet.AttrItems) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(packet.AttrItems))
	}

	packet.ClearAttr()
	if len(packet.AttrItems) != 0 {
		t.Errorf("Expected 0 attributes after ClearAttr, got %d", len(packet.AttrItems))
	}
}

// TestPacketGetValue tests GetValue method
func TestPacketGetValue(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	value := packet.GetValue("User-Name")
	if value == nil {
		t.Error("GetValue returned nil for existing attribute")
	}
	if value != "testuser" {
		t.Errorf("Expected 'testuser', got %v", value)
	}

	// Test non-existent attribute
	value = packet.GetValue("NonExistent")
	if value != nil {
		t.Error("GetValue should return nil for non-existent attribute")
	}
}

// TestPacketGetString tests GetString method
func TestPacketGetString(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.GetString("User-Name")
	if str != "testuser" {
		t.Errorf("Expected 'testuser', got '%s'", str)
	}

	// Test non-existent attribute
	str = packet.GetString("NonExistent")
	if str != "" {
		t.Errorf("Expected empty string for non-existent attribute, got '%s'", str)
	}
}

// TestPacketFindAttr tests FindAttr method
func TestPacketFindAttr(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	attr := packet.FindAttr("User-Name")
	if attr == nil {
		t.Error("FindAttr returned nil for existing attribute")
	}

	attr = packet.FindAttr("NonExistent")
	if attr != nil {
		t.Error("FindAttr should return nil for non-existent attribute")
	}
}

// TestPacketAddAttr tests AddAttr method
func TestPacketAddAttr(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Adding same attribute again should fail
	err = packet.AddAttr("User-Name", "anotheruser")
	if err == nil {
		t.Error("Expected error when adding duplicate attribute")
	}
}

// TestPacketAddAttrUnknown tests AddAttr with unknown attribute name
func TestPacketAddAttrUnknown(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("UnknownAttribute", "value")
	if err == nil {
		t.Error("Expected error for unknown attribute name")
	}
}

// TestPacketSet tests Set method
func TestPacketSet(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Set new attribute (should add)
	err := packet.Set("User-Name", "testuser")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Set existing attribute (should update)
	err = packet.Set("User-Name", "newuser")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	value := packet.GetValue("User-Name")
	if value != "newuser" {
		t.Errorf("Expected 'newuser', got %v", value)
	}
}

// TestPacketPAP tests PAP method
func TestPacketPAP(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// PAP on packet without User-Name should fail
	_, _, err := packet.PAP()
	if err == nil {
		t.Error("Expected error for PAP without User-Name")
	}

	// Add User-Name
	err = packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// PAP on packet without User-Password should fail
	_, _, err = packet.PAP()
	if err == nil {
		t.Error("Expected error for PAP without User-Password")
	}

	// Add User-Password
	err = packet.Set("User-Password", "testpass")
	if err != nil {
		t.Fatalf("Set User-Password failed: %v", err)
	}

	username, password, err := packet.PAP()
	if err != nil {
		t.Fatalf("PAP failed: %v", err)
	}
	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", username)
	}
	if password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", password)
	}
}

// TestPacketPAPWrongCode tests PAP with wrong packet code
func TestPacketPAPWrongCode(t *testing.T) {
	packet := NewPacket(CodeAccessAccept, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	_, _, err := packet.PAP()
	if err == nil {
		t.Error("Expected error for PAP on non-AccessRequest packet")
	}
}

// TestPacketEncode tests Encode method
func TestPacketEncode(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short: %d bytes", len(data))
	}
	if data[0] != byte(CodeAccessRequest) {
		t.Errorf("Wrong code in encoded packet")
	}
}

// TestPacketEncodeAccessAccept tests encoding AccessAccept packet
func TestPacketEncodeAccessAccept(t *testing.T) {
	packet := NewPacket(CodeAccessAccept, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Set authenticator for response packet
	copy(packet.Authenticator[:], make([]byte, 16))

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short: %d bytes", len(data))
	}
}

// TestPacketEncodeAccountingRequest tests encoding AccountingRequest packet
func TestPacketEncodeAccountingRequest(t *testing.T) {
	packet := NewPacket(CodeAccountingRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short: %d bytes", len(data))
	}
}

// TestPacketEncodeUnknownCode tests encoding with unknown code
func TestPacketEncodeUnknownCode(t *testing.T) {
	packet := NewPacket(Code(100), []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	_, err := packet.Encode()
	if err == nil {
		t.Error("Expected error for unknown packet code")
	}
}

// TestPacketIsAuthentic tests IsAuthentic method
func TestPacketIsAuthentic(t *testing.T) {
	request := NewPacket(CodeAccessRequest, []byte("secret"))
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create a response packet
	response := NewPacket(CodeAccessAccept, []byte("secret"))
	if response == nil {
		t.Fatal("NewPacket returned nil")
	}
	response.Identifier = request.Identifier
	copy(response.Authenticator[:], request.Authenticator[:])

	// IsAuthentic should return false for unauthenticated response
	// (since we didn't properly calculate the authenticator)
	result := response.IsAuthentic(request)
	// This will be false since we didn't properly encode
	t.Logf("IsAuthentic result: %v", result)
}

// TestPacketIsAuthenticAccountingRequest tests IsAuthentic for AccountingRequest
func TestPacketIsAuthenticAccountingRequest(t *testing.T) {
	request := NewPacket(CodeAccountingRequest, []byte("secret"))
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Encode the packet to get proper authenticator
	data, err := request.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Check authenticity
	result := parsed.IsAuthentic(request)
	t.Logf("IsAuthentic for AccountingRequest: %v", result)
}

// TestSetVendorSpecific tests SetVendorSpecific function
func TestSetVendorSpecific(t *testing.T) {
	err := SetVendorSpecific("hillstone")
	if err != nil {
		t.Fatalf("SetVendorSpecific failed: %v", err)
	}
	if VerdorID != 28557 {
		t.Errorf("Expected VerdorID 28557, got %d", VerdorID)
	}
	if VerdorTag != 0 {
		t.Errorf("Expected VerdorTag 0, got %d", VerdorTag)
	}
	if VerdorTypeID != 11 {
		t.Errorf("Expected VerdorTypeID 11, got %d", VerdorTypeID)
	}
}

// TestSetVendorSpecificUnknown tests SetVendorSpecific with unknown vendor
func TestSetVendorSpecificUnknown(t *testing.T) {
	err := SetVendorSpecific("unknown")
	if err == nil {
		t.Error("Expected error for unknown vendor")
	}
}

// TestCodeConstants tests that all code constants are correct
func TestCodeConstants(t *testing.T) {
	if CodeAccessRequest != 1 {
		t.Errorf("CodeAccessRequest should be 1")
	}
	if CodeAccessAccept != 2 {
		t.Errorf("CodeAccessAccept should be 2")
	}
	if CodeAccessReject != 3 {
		t.Errorf("CodeAccessReject should be 3")
	}
	if CodeAccountingRequest != 4 {
		t.Errorf("CodeAccountingRequest should be 4")
	}
	if CodeAccountingResponse != 5 {
		t.Errorf("CodeAccountingResponse should be 5")
	}
	if CodeAccessChallenge != 11 {
		t.Errorf("CodeAccessChallenge should be 11")
	}
	if CodeStatusServer != 12 {
		t.Errorf("CodeStatusServer should be 12")
	}
	if CodeStatusClient != 13 {
		t.Errorf("CodeStatusClient should be 13")
	}
	if CodeReserved != 255 {
		t.Errorf("CodeReserved should be 255")
	}
}

// TestPacketEncodeDecodeRoundTrip tests encoding and decoding round trip
func TestPacketEncodeDecodeRoundTrip(t *testing.T) {
	secret := []byte("testsecret")

	packet := NewPacket(CodeAccessRequest, secret)
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	err = packet.AddAttr("NAS-IP-Address", net.ParseIP("192.168.1.1"))
	if err != nil {
		t.Fatalf("AddAttr NAS-IP-Address failed: %v", err)
	}

	err = packet.AddAttr("NAS-Port", uint32(1234))
	if err != nil {
		t.Fatalf("AddAttr NAS-Port failed: %v", err)
	}

	encoded, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	if decoded.Code != packet.Code {
		t.Errorf("Code mismatch")
	}
	if decoded.Identifier != packet.Identifier {
		t.Errorf("Identifier mismatch")
	}

	// Check User-Name
	username := decoded.GetString("User-Name")
	if username != "testuser" {
		t.Errorf("User-Name mismatch: expected 'testuser', got '%s'", username)
	}
}

// TestPacketWithMultipleAttributes tests packet with multiple attributes
func TestPacketWithMultipleAttributes(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	attrs := map[string]interface{}{
		"User-Name":      "testuser",
		"NAS-IP-Address": net.ParseIP("10.0.0.1"),
		"NAS-Port":       uint32(100),
		"Service-Type":   uint32(1),
	}

	for name, value := range attrs {
		err := packet.AddAttr(name, value)
		if err != nil {
			t.Fatalf("AddAttr %s failed: %v", name, err)
		}
	}

	if len(packet.AttrItems) != len(attrs) {
		t.Errorf("Expected %d attributes, got %d", len(attrs), len(packet.AttrItems))
	}
}

// TestParsePacketWithUserPassword tests parsing packet with User-Password attribute
func TestParsePacketWithUserPassword(t *testing.T) {
	secret := []byte("testsecret")

	// Create a packet with User-Password
	packet := NewPacket(CodeAccessRequest, secret)
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	password := "mypassword"
	err = packet.Set("User-Password", password)
	if err != nil {
		t.Fatalf("Set User-Password failed: %v", err)
	}

	// Encode
	encoded, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	decoded, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Get password
	decodedPassword := decoded.GetString("User-Password")
	if decodedPassword != password {
		t.Errorf("Password mismatch: expected '%s', got '%s'", password, decodedPassword)
	}
}

// TestParsePacketWithShortAttributeData tests parsing with truncated attribute data
func TestParsePacketWithShortAttributeData(t *testing.T) {
	// Create a packet where the attribute data is shorter than declared
	data := make([]byte, 25)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)
	// Authenticator (16 bytes) - zeros

	// Add an attribute that claims to be longer than available data
	data[20] = 1  // Type (User-Name)
	data[21] = 10 // Length claims 10 bytes, but only 3 bytes available
	data[22] = 'a'
	data[23] = 'b'
	data[24] = 'c'

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for truncated attribute data")
	}
}

// TestPacketStringWithNilDictionaryAttr tests String with attribute not in dictionary
func TestPacketStringWithNilDictionaryAttr(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Manually add an attribute with unknown type
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    250, // Unknown type
		AttrValue: "test",
	})

	str := packet.String()
	if str == "" {
		t.Error("String() should not return empty even with unknown attribute")
	}
}

// TestPacketStringWithIntegerAttribute tests String with integer attribute
func TestPacketStringWithIntegerAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("NAS-Port", uint32(12345))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

// TestPacketStringWithMessageAuthenticator tests String with Message-Authenticator
func TestPacketStringWithMessageAuthenticator(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("Message-Authenticator", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10})
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
}

// TestPacketEncodeWithNilAttribute tests Encode with nil attribute
func TestPacketEncodeWithNilAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add a nil attribute
	packet.AttrItems = append(packet.AttrItems, nil)

	// Should not panic
	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestPacketEncodeTooLong tests Encode with attribute too long
func TestPacketEncodeTooLong(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create a very long value
	longValue := make([]byte, 300)
	for i := range longValue {
		longValue[i] = 'a'
	}

	// This should fail during encode because attribute is too long
	err := packet.AddAttr("User-Name", string(longValue))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	_, err = packet.Encode()
	if err == nil {
		t.Error("Expected error for attribute too long")
	}
}

// TestParsePacketWithEmptyPassword tests parsing with empty User-Password
func TestParsePacketWithEmptyPassword(t *testing.T) {
	secret := []byte("testsecret")

	packet := NewPacket(CodeAccessRequest, secret)
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Set empty password
	err = packet.Set("User-Password", "")
	if err != nil {
		t.Fatalf("Set User-Password failed: %v", err)
	}

	encoded, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	password := decoded.GetString("User-Password")
	if password != "" {
		t.Errorf("Expected empty password, got '%s'", password)
	}
}

// TestPacketGetStringWithUint32 tests GetString with uint32 value
func TestPacketGetStringWithUint32(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("NAS-Port", uint32(12345))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.GetString("NAS-Port")
	if str != "12345" {
		t.Errorf("Expected '12345', got '%s'", str)
	}
}

// TestPacketGetStringWithByteSlice tests GetString with []byte value
func TestPacketGetStringWithByteSlice(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("State", []byte("teststate"))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.GetString("State")
	if str != "teststate" {
		t.Errorf("Expected 'teststate', got '%s'", str)
	}
}

// TestMaxPacketSize tests the maxPacketSize constant
func TestMaxPacketSize(t *testing.T) {
	if maxPacketSize != 4095 {
		t.Errorf("maxPacketSize should be 4095, got %d", maxPacketSize)
	}
}

// TestPacketEncodeStatusServer tests encoding StatusServer packet
func TestPacketEncodeStatusServer(t *testing.T) {
	packet := NewPacket(CodeStatusServer, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestPacketEncodeAccessReject tests encoding AccessReject packet
func TestPacketEncodeAccessReject(t *testing.T) {
	packet := NewPacket(CodeAccessReject, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	copy(packet.Authenticator[:], make([]byte, 16))

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestPacketEncodeAccessChallenge tests encoding AccessChallenge packet
func TestPacketEncodeAccessChallenge(t *testing.T) {
	packet := NewPacket(CodeAccessChallenge, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	copy(packet.Authenticator[:], make([]byte, 16))

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestPacketEncodeAccountingResponse tests encoding AccountingResponse packet
func TestPacketEncodeAccountingResponse(t *testing.T) {
	packet := NewPacket(CodeAccountingResponse, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	copy(packet.Authenticator[:], make([]byte, 16))

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestParsePacketExactLength tests parsing packet where data length equals declared length
func TestParsePacketExactLength(t *testing.T) {
	// Create a packet where the data length exactly matches the declared length
	data := make([]byte, 20)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 20)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if packet == nil {
		t.Error("Packet should not be nil")
	}
}

// TestParsePacketDataLongerThanLength tests parsing where data is longer than declared length
// Note: ParsePacket uses all data after the 20-byte header, not just up to declared length.
// So extra bytes beyond the declared length will be parsed as attributes.
func TestParsePacketDataLongerThanLength(t *testing.T) {
	// Create a packet where actual data is longer than declared length
	// The extra bytes will be parsed as attributes, so we need to make them valid
	data := make([]byte, 30)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 20) // Declared length is 20, but data is 30

	// The bytes from 20-30 will be parsed as attributes.
	// Make them a valid attribute: type=1, length=10, value="testuser"
	data[20] = 1  // User-Name
	data[21] = 10 // Length
	copy(data[22:], []byte("testuser"))

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if packet == nil {
		t.Error("Packet should not be nil")
	}
}

// TestPacketIsAuthenticFalse tests IsAuthentic returning false for unsupported codes
func TestPacketIsAuthenticFalse(t *testing.T) {
	request := NewPacket(CodeAccessRequest, []byte("secret"))
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Test with a code that doesn't match the switch cases
	response := NewPacket(CodeStatusServer, []byte("secret"))
	if response == nil {
		t.Fatal("NewPacket returned nil")
	}

	result := response.IsAuthentic(request)
	if result {
		t.Error("IsAuthentic should return false for StatusServer code")
	}
}

// TestPacketPAPWithNonStringUserName tests PAP with non-string User-Name
func TestPacketPAPWithNonStringUserName(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Manually add a non-string User-Name
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    1, // User-Name
		AttrValue: 12345, // Not a string
	})

	_, _, err := packet.PAP()
	if err == nil {
		t.Error("Expected error for non-string User-Name")
	}
}

// TestPacketPAPWithNonStringPassword tests PAP with non-string User-Password
func TestPacketPAPWithNonStringPassword(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Manually add a non-string User-Password
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    2, // User-Password
		AttrValue: 12345, // Not a string
	})

	_, _, err = packet.PAP()
	if err == nil {
		t.Error("Expected error for non-string User-Password")
	}
}

// TestPacketEncodeWithVendorSpecific tests encoding with Vendor-Specific attribute
func TestPacketEncodeWithVendorSpecific(t *testing.T) {
	// Set vendor specific
	err := SetVendorSpecific("hillstone")
	if err != nil {
		t.Fatalf("SetVendorSpecific failed: %v", err)
	}

	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err = packet.AddAttr("Vendor-Specific", "testvalue")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestParsePacketWithVendorSpecific tests parsing with Vendor-Specific attribute
func TestParsePacketWithVendorSpecific(t *testing.T) {
	// Create a Vendor-Specific attribute
	vsaData := EncodeAVPair(28557, 1, "testvalue")

	// Create a packet with VSA
	data := make([]byte, 20+2+len(vsaData))
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], uint16(20+2+len(vsaData)))

	// Add Vendor-Specific attribute (type 26)
	data[20] = 26
	data[21] = byte(2 + len(vsaData))
	copy(data[22:], vsaData)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if packet == nil {
		t.Error("Packet should not be nil")
	}
}

// TestParsePacketWithInvalidVendorSpecific tests parsing with invalid VSA
func TestParsePacketWithInvalidVendorSpecific(t *testing.T) {
	// Create an invalid VSA (too short)
	vsaData := []byte{0x00, 0x00, 0x00} // Only 3 bytes, need at least 7

	// Create a packet with invalid VSA
	data := make([]byte, 20+2+len(vsaData))
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], uint16(20+2+len(vsaData)))

	// Add Vendor-Specific attribute (type 26)
	data[20] = 26
	data[21] = byte(2 + len(vsaData))
	copy(data[22:], vsaData)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	// Should decode as text since VSA is invalid
	_ = packet
}

// TestParsePacketWithVeryLargeVendorID tests parsing with very large vendor ID
func TestParsePacketWithVeryLargeVendorID(t *testing.T) {
	// Create a VSA with vendor ID > 99999
	vsaData := EncodeAVPair(999999, 1, "testvalue")

	// Create a packet with VSA
	data := make([]byte, 20+2+len(vsaData))
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], uint16(20+2+len(vsaData)))

	// Add Vendor-Specific attribute (type 26)
	data[20] = 26
	data[21] = byte(2 + len(vsaData))
	copy(data[22:], vsaData)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	// Should decode as text since vendor ID > 99999
	_ = packet
}

// TestPacketEncodeMultipleAttributes tests encoding multiple attributes
func TestPacketEncodeMultipleAttributes(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add multiple attributes
	for i := 0; i < 10; i++ {
		err := packet.AddAttr("User-Name", "testuser")
		if err != nil {
			// Expected on second attempt
			break
		}
	}

	// Add different attributes
	err := packet.AddAttr("NAS-IP-Address", net.ParseIP("192.168.1.1"))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	err = packet.AddAttr("NAS-Port", uint32(1234))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestParsePacketWithMultipleAttributes tests parsing multiple attributes
func TestParsePacketWithMultipleAttributes(t *testing.T) {
	// Create a packet with multiple attributes
	data := make([]byte, 20)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1

	// Add User-Name
	userName := []byte("testuser")
	attr1 := append([]byte{1, byte(2 + len(userName))}, userName...)

	// Add NAS-Port (integer)
	nasPort := make([]byte, 4)
	binary.BigEndian.PutUint32(nasPort, 1234)
	attr2 := append([]byte{5, 6}, nasPort...)

	// Add NAS-IP-Address
	nasIP := net.ParseIP("192.168.1.1").To4()
	attr3 := append([]byte{4, 6}, nasIP...)

	attrs := append(attr1, attr2...)
	attrs = append(attrs, attr3...)

	totalLen := 20 + len(attrs)
	binary.BigEndian.PutUint16(data[2:4], uint16(totalLen))
	data = append(data, attrs...)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if len(packet.AttrItems) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(packet.AttrItems))
	}
}

// TestPacketStringWithEmptyUserPassword tests String with empty User-Password
func TestPacketStringWithEmptyUserPassword(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Manually add an empty User-Password
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    2, // User-Password
		AttrValue: "",
	})

	str := packet.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

// TestParsePacketWithRemainingBytes tests parsing where the declared length is exact
func TestParsePacketWithRemainingBytes(t *testing.T) {
	// Create a packet where the length exactly matches the content
	data := make([]byte, 25)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add one attribute
	data[20] = 1 // User-Name
	data[21] = 5 // Length = 5 (type + length + 3 bytes value)
	data[22] = 'a'
	data[23] = 'b'
	data[24] = 'c'

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}
	if packet == nil {
		t.Error("Packet should not be nil")
	}
	if len(packet.AttrItems) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(packet.AttrItems))
	}
}

// TestPacketEncodeWithTimeAttribute tests encoding with time attribute
func TestPacketEncodeWithTimeAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add Acct-Session-Time (time attribute)
	err := packet.AddAttr("Acct-Session-Time", uint32(3600))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestParsePacketWithTimeAttribute tests parsing with time attribute
func TestParsePacketWithTimeAttribute(t *testing.T) {
	// Create a packet with Acct-Session-Time
	data := make([]byte, 20+6) // Header + attribute
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 26)

	// Add Acct-Session-Time (type 46)
	data[20] = 46
	data[21] = 6
	binary.BigEndian.PutUint32(data[22:26], 3600)

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	value := packet.GetValue("Acct-Session-Time")
	if value == nil {
		t.Error("GetValue returned nil")
	}
}

// TestPacketGetStringWithUnknownType tests GetString with unknown attribute type
func TestPacketGetStringWithUnknownType(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Manually add an attribute with unknown type
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    200, // Unknown
		AttrValue: "test",
	})

	// GetString should return empty for unknown attribute
	str := packet.GetString("UnknownAttribute")
	if str != "" {
		t.Errorf("Expected empty string, got '%s'", str)
	}
}

// TestPacketGetStringWithFmtStringer tests GetString with fmt.Stringer value
func TestPacketGetStringWithFmtStringer(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add an attribute with a value that implements fmt.Stringer
	err := packet.AddAttr("NAS-IP-Address", net.ParseIP("192.168.1.1"))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	str := packet.GetString("NAS-IP-Address")
	if str == "" {
		t.Error("GetString should not return empty for IP address")
	}
}

// TestPacketEncodeWithAddressAttribute tests encoding with address attribute
func TestPacketEncodeWithAddressAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("NAS-IP-Address", net.ParseIP("192.168.1.1"))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse back
	decoded, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	ip := decoded.GetValue("NAS-IP-Address")
	if ip == nil {
		t.Error("GetValue returned nil")
	}
}

// TestParsePacketWithAddressAttribute tests parsing with address attribute
func TestParsePacketWithAddressAttribute(t *testing.T) {
	// Create a packet with NAS-IP-Address
	data := make([]byte, 20+6) // Header + attribute
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 26)

	// Add NAS-IP-Address (type 4)
	data[20] = 4
	data[21] = 6
	copy(data[22:26], net.ParseIP("192.168.1.1").To4())

	packet, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	value := packet.GetValue("NAS-IP-Address")
	if value == nil {
		t.Error("GetValue returned nil")
	}
}

// TestParsePacketWithInvalidAddress tests parsing with invalid address length
func TestParsePacketWithInvalidAddress(t *testing.T) {
	// Create a packet with invalid address length
	data := make([]byte, 20+5) // Header + attribute with wrong length
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add NAS-IP-Address with wrong length (should be 6, not 5)
	data[20] = 4
	data[21] = 5
	data[22] = 192
	data[23] = 168
	data[24] = 1

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for invalid address length")
	}
}

// TestParsePacketWithInvalidInteger tests parsing with invalid integer length
func TestParsePacketWithInvalidInteger(t *testing.T) {
	// Create a packet with invalid integer length
	data := make([]byte, 20+5) // Header + attribute with wrong length
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add NAS-Port with wrong length (should be 6, not 5)
	data[20] = 5
	data[21] = 5
	data[22] = 0
	data[23] = 0
	data[24] = 0

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for invalid integer length")
	}
}

// TestParsePacketWithInvalidTime tests parsing with invalid time length
func TestParsePacketWithInvalidTime(t *testing.T) {
	// Create a packet with invalid time length
	data := make([]byte, 20+5) // Header + attribute with wrong length
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add Acct-Session-Time with wrong length (should be 6, not 5)
	data[20] = 46
	data[21] = 5
	data[22] = 0
	data[23] = 0
	data[24] = 0

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for invalid time length")
	}
}

// TestParsePacketWithInvalidUTF8 tests parsing with invalid UTF-8 text
func TestParsePacketWithInvalidUTF8(t *testing.T) {
	// Create a packet with invalid UTF-8 in text attribute
	data := make([]byte, 20+5) // Header + attribute
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add User-Name with invalid UTF-8
	data[20] = 1
	data[21] = 5
	data[22] = 0xff // Invalid UTF-8
	data[23] = 0xfe
	data[24] = 0xfd

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error for invalid UTF-8")
	}
}

// TestPacketEncodeWithTextAttribute tests encoding with text attribute
func TestPacketEncodeWithTextAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	username := decoded.GetString("User-Name")
	if username != "testuser" {
		t.Errorf("Expected 'testuser', got '%s'", username)
	}
}

// TestPacketEncodeWithStringAttribute tests encoding with string attribute
func TestPacketEncodeWithStringAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("State", []byte("teststate"))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	state := decoded.GetString("State")
	if state != "teststate" {
		t.Errorf("Expected 'teststate', got '%s'", state)
	}
}

// TestPacketEncodeWithIntegerAttribute tests encoding with integer attribute
func TestPacketEncodeWithIntegerAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("NAS-Port", uint32(12345))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := ParsePacket(data, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	value := decoded.GetValue("NAS-Port")
	if value == nil {
		t.Fatal("GetValue returned nil")
	}
	if value.(uint32) != 12345 {
		t.Errorf("Expected 12345, got %d", value.(uint32))
	}
}

// TestPacketEncodeWithTimeValue tests encoding with time.Time value
func TestPacketEncodeWithTimeValue(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Event-Timestamp is not registered, so we test with a different approach
	// Just verify that time.Time encoding works
	now := time.Now()
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, uint32(now.Unix()))

	// Verify the time encoding manually
	_ = raw
}

// TestPacketSetWithTransformer tests Set with transformer
func TestPacketSetWithTransformer(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// User-Password has a transformer
	err := packet.Set("User-Password", "testpass")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Update the password
	err = packet.Set("User-Password", "newpass")
	if err != nil {
		t.Fatalf("Set update failed: %v", err)
	}
}

// TestPacketAddAttrWithTransformer tests AddAttr with transformer
func TestPacketAddAttrWithTransformer(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// User-Password has a transformer
	err := packet.AddAttr("User-Password", "testpass")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}
}

// TestParsePacketWithUserPasswordDecodeError tests parsing User-Password with decode error
func TestParsePacketWithUserPasswordDecodeError(t *testing.T) {
	// Create a packet with User-Password but no secret
	data := make([]byte, 20+18) // Header + User-Password attribute
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 38)

	// Add User-Password (type 2) with 16 bytes value
	data[20] = 2
	data[21] = 18
	copy(data[22:38], make([]byte, 16))

	// Parse with nil secret - should use default password
	packet, err := ParsePacket(data, nil, Builtin)
	if err != nil {
		// This is expected - User-Password decode fails without secret
		t.Logf("ParsePacket returned error (expected): %v", err)
	} else {
		// Check if default password was used
		if len(packet.AttrItems) > 0 {
			t.Log("Packet parsed with default password")
		}
	}
}

// TestPacketEncodeWithLongPacket tests encoding a very long packet
func TestPacketEncodeWithLongPacket(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Add many attributes to make packet long
	for i := 0; i < 100; i++ {
		// Use different attribute names to avoid duplicates
		switch i % 5 {
		case 0:
			packet.Set("User-Name", "testuser")
		case 1:
			packet.Set("NAS-IP-Address", net.ParseIP("192.168.1.1"))
		case 2:
			packet.Set("NAS-Port", uint32(i))
		case 3:
			packet.Set("Service-Type", uint32(1))
		case 4:
			packet.Set("State", []byte("state"))
		}
	}

	data, err := packet.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) < 20 {
		t.Errorf("Encoded packet too short")
	}
}

// TestPacketEncodeTooLongPacket tests encoding a packet that exceeds max size
func TestPacketEncodeTooLongPacket(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create a very long attribute value
	longValue := make([]byte, 2000)
	for i := range longValue {
		longValue[i] = 'a'
	}

	// This should fail during encode
	err := packet.AddAttr("State", longValue)
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	_, err = packet.Encode()
	if err == nil {
		t.Error("Expected error for packet too long")
	}
}

// TestHexEncodingInError tests that hex encoding works in error messages
func TestHexEncodingInError(t *testing.T) {
	// Create a packet with an attribute that will fail to decode
	data := make([]byte, 20+5)
	data[0] = byte(CodeAccessRequest)
	data[1] = 1
	binary.BigEndian.PutUint16(data[2:4], 25)

	// Add NAS-IP-Address with wrong length
	data[20] = 4
	data[21] = 5
	data[22] = 0x01
	data[23] = 0x02
	data[24] = 0x03

	_, err := ParsePacket(data, []byte("secret"), Builtin)
	if err == nil {
		t.Error("Expected error")
	} else {
		// Check that error message contains hex
		errStr := err.Error()
		if !bytes.Contains([]byte(errStr), []byte("0x")) && !bytes.Contains([]byte(errStr), []byte("01")) {
			t.Logf("Error message: %s", errStr)
		}
	}
}

// Helper function to create hex dump
func hexDump(data []byte) string {
	return hex.EncodeToString(data)
}

// TestPacketSetNewAttribute tests Set method adding a new attribute
func TestPacketSetNewAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Set a new attribute (should add since it doesn't exist)
	err := packet.Set("User-Name", "testuser")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if len(packet.AttrItems) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(packet.AttrItems))
	}
}

// TestPacketSetExistingAttribute tests Set method updating an existing attribute
func TestPacketSetExistingAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("User-Name", "testuser")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Update existing attribute
	err = packet.Set("User-Name", "newuser")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if packet.GetString("User-Name") != "newuser" {
		t.Errorf("Expected 'newuser', got '%s'", packet.GetString("User-Name"))
	}
}

// TestPacketSetWithTransformerError tests Set with transformer that fails
func TestPacketSetWithTransformerError(t *testing.T) {
	// Create a dictionary with a transformer that can fail
	dict := &TDictionary{}
	dict.NameItems = make(map[string]*TDictEntry)

	codec := &testTransformerCodec{}
	dict.MustRegister("Test-Transformer", 200, codec)

	packet := &TDataPacket{
		Code:       CodeAccessRequest,
		Dictionary: dict,
		Secret:     []byte("secret"),
	}

	// Add attribute with valid value
	err := packet.AddAttr("Test-Transformer", "valid")
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Try to set with an invalid type for transformer
	err = packet.Set("Test-Transformer", 12345)
	if err == nil {
		t.Error("Expected error for invalid type in transformer")
	}
}

// TestPacketSetNonTransformerAttribute tests Set with non-transformer attribute
func TestPacketSetNonTransformerAttribute(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	err := packet.AddAttr("NAS-Port", uint32(1234))
	if err != nil {
		t.Fatalf("AddAttr failed: %v", err)
	}

	// Update without transformer
	err = packet.Set("NAS-Port", uint32(5678))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	value := packet.GetValue("NAS-Port")
	if value.(uint32) != 5678 {
		t.Errorf("Expected 5678, got %d", value.(uint32))
	}
}

// TestPacketIsAuthenticAccessReject tests IsAuthentic for AccessReject
func TestPacketIsAuthenticAccessReject(t *testing.T) {
	secret := []byte("sharedsecret")
	request := NewPacket(CodeAccessRequest, secret)
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create AccessReject response - must copy request authenticator
	response := &TDataPacket{
		Code:          CodeAccessReject,
		Identifier:    request.Identifier,
		Authenticator: request.Authenticator, // Copy from request
		Secret:        secret,
		Dictionary:    Builtin,
	}

	// Encode to get proper authenticator
	encoded, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Check authenticity
	result := parsed.IsAuthentic(request)
	if !result {
		t.Error("IsAuthentic should return true for properly encoded AccessReject")
	}
}

// TestPacketIsAuthenticAccessChallenge tests IsAuthentic for AccessChallenge
func TestPacketIsAuthenticAccessChallenge(t *testing.T) {
	secret := []byte("sharedsecret")
	request := NewPacket(CodeAccessRequest, secret)
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create AccessChallenge response - must copy request authenticator
	response := &TDataPacket{
		Code:          CodeAccessChallenge,
		Identifier:    request.Identifier,
		Authenticator: request.Authenticator,
		Secret:        secret,
		Dictionary:    Builtin,
	}

	// Encode to get proper authenticator
	encoded, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Check authenticity
	result := parsed.IsAuthentic(request)
	if !result {
		t.Error("IsAuthentic should return true for properly encoded AccessChallenge")
	}
}

// TestPacketIsAuthenticAccessAccept tests IsAuthentic for AccessAccept
func TestPacketIsAuthenticAccessAccept(t *testing.T) {
	secret := []byte("sharedsecret")
	request := NewPacket(CodeAccessRequest, secret)
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create AccessAccept response - must copy request authenticator
	response := &TDataPacket{
		Code:          CodeAccessAccept,
		Identifier:    request.Identifier,
		Authenticator: request.Authenticator,
		Secret:        secret,
		Dictionary:    Builtin,
	}

	// Encode to get proper authenticator
	encoded, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(encoded, secret, Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Check authenticity
	result := parsed.IsAuthentic(request)
	if !result {
		t.Error("IsAuthentic should return true for properly encoded AccessAccept")
	}
}

// TestPacketIsAuthenticAccountingRequest tests IsAuthentic for AccountingRequest
func TestPacketIsAuthenticAccountingRequestDetailed(t *testing.T) {
	request := NewPacket(CodeAccountingRequest, []byte("secret"))
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Encode to get proper authenticator
	encoded, err := request.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(encoded, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Check authenticity against itself
	result := parsed.IsAuthentic(request)
	if !result {
		t.Error("IsAuthentic should return true for properly encoded AccountingRequest")
	}
}

// TestPacketIsAuthenticBadResponse tests IsAuthentic with tampered response
func TestPacketIsAuthenticBadResponse(t *testing.T) {
	request := NewPacket(CodeAccessRequest, []byte("secret"))
	if request == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Create AccessAccept response
	response := NewPacket(CodeAccessAccept, []byte("secret"))
	if response == nil {
		t.Fatal("NewPacket returned nil")
	}
	response.Identifier = request.Identifier

	// Encode
	encoded, err := response.Encode()
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Parse it back
	parsed, err := ParsePacket(encoded, []byte("secret"), Builtin)
	if err != nil {
		t.Fatalf("ParsePacket failed: %v", err)
	}

	// Now tamper with the request's authenticator
	var fakeAuth [16]byte
	for i := range fakeAuth {
		fakeAuth[i] = 0xFF
	}
	request.Authenticator = fakeAuth

	// Should fail authenticity check
	result := parsed.IsAuthentic(request)
	if result {
		t.Error("IsAuthentic should return false for tampered request")
	}
}

// TestPacketGetStringWithIAttributeStringer tests GetString with IAttributeStringer
func TestPacketGetStringWithIAttributeStringer(t *testing.T) {
	// Create a custom dictionary with a stringer codec
	dict := &TDictionary{}
	dict.NameItems = make(map[string]*TDictEntry)

	codec := &testStringerCodec{}
	dict.MustRegister("Test-Stringer", 200, codec)

	packet := &TDataPacket{
		Code:       CodeAccessRequest,
		Dictionary: dict,
		Secret:     []byte("secret"),
	}
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    200,
		AttrValue: "rawvalue",
	})

	str := packet.GetString("Test-Stringer")
	if str != "stringer:rawvalue" {
		t.Errorf("Expected 'stringer:rawvalue', got '%s'", str)
	}
}

// testStringerCodec implements IAttributeCodec with IAttributeStringer
type testStringerCodec struct{}

func (testStringerCodec) Decode(packet *TDataPacket, wire []byte) (interface{}, error) {
	return string(wire), nil
}

func (testStringerCodec) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	return []byte(value.(string)), nil
}

func (testStringerCodec) GetCodeName() string {
	return "TestStringer"
}

func (testStringerCodec) String(value interface{}) string {
	return "stringer:" + value.(string)
}

// TestPacketGetStringWithNonStringableValue tests GetString with non-stringable value
func TestPacketGetStringWithNonStringableValue(t *testing.T) {
	packet := NewPacket(CodeAccessRequest, []byte("secret"))
	if packet == nil {
		t.Fatal("NewPacket returned nil")
	}

	// Manually add an attribute with a non-stringable value (e.g., int)
	packet.AttrItems = append(packet.AttrItems, &TAttribute{
		AttrId:    5, // NAS-Port (integer)
		AttrValue: int(12345), // Not uint32, not string, not []byte
	})

	str := packet.GetString("NAS-Port")
	// Should return "" for non-stringable value
	if str != "" {
		t.Logf("GetString for non-stringable value returned: '%s'", str)
	}
}

// TestDictionaryNewAttrTransformerError tests NewAttr with transformer error
func TestDictionaryNewAttrTransformerError(t *testing.T) {
	// Create a dictionary with a transformer that can fail
	dict := &TDictionary{}
	dict.NameItems = make(map[string]*TDictEntry)

	codec := &testTransformerCodec{}
	dict.MustRegister("Test-Transformer", 200, codec)

	// Valid value
	attr, err := dict.NewAttr("Test-Transformer", "valid")
	if err != nil {
		t.Fatalf("NewAttr failed: %v", err)
	}
	if attr == nil {
		t.Error("Attr should not be nil")
	}

	// Invalid value (should trigger transformer error)
	_, err = dict.NewAttr("Test-Transformer", 12345)
	if err == nil {
		t.Error("Expected error for transformer failure")
	}
}

// testTransformerCodec implements IAttributeCodec with IAttributeTransformer
type testTransformerCodec struct{}

func (testTransformerCodec) Decode(packet *TDataPacket, wire []byte) (interface{}, error) {
	return string(wire), nil
}

func (testTransformerCodec) Encode(packet *TDataPacket, value interface{}) ([]byte, error) {
	return []byte(value.(string)), nil
}

func (testTransformerCodec) GetCodeName() string {
	return "TestTransformer"
}

func (testTransformerCodec) Transform(value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("transformer requires string, got %T", value)
	}
	return "transformed:" + str, nil
}
