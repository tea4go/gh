package radius

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// TestAttributeTextDecode tests attributeText Decode
func TestAttributeTextDecode(t *testing.T) {
	codec := attributeText{}

	// Valid UTF-8
	value, err := codec.Decode(nil, []byte("hello"))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if value != "hello" {
		t.Errorf("Expected 'hello', got %v", value)
	}

	// Invalid UTF-8
	_, err = codec.Decode(nil, []byte{0xff, 0xfe, 0xfd})
	if err == nil {
		t.Error("Expected error for invalid UTF-8")
	}
}

// TestAttributeTextEncode tests attributeText Encode
func TestAttributeTextEncode(t *testing.T) {
	codec := attributeText{}

	// String value
	data, err := codec.Encode(nil, "hello")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("Expected 'hello', got %s", data)
	}

	// []byte value
	data, err = codec.Encode(nil, []byte("world"))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("Expected 'world', got %s", data)
	}

	// Invalid type
	_, err = codec.Encode(nil, 12345)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestAttributeTextGetCodeName tests attributeText GetCodeName
func TestAttributeTextGetCodeName(t *testing.T) {
	codec := attributeText{}
	if codec.GetCodeName() != "AttributeText" {
		t.Errorf("Expected 'AttributeText', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeStringDecode tests attributeString Decode
func TestAttributeStringDecode(t *testing.T) {
	codec := attributeString{}

	value, err := codec.Decode(nil, []byte("hello"))
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	result, ok := value.([]byte)
	if !ok {
		t.Fatalf("Expected []byte, got %T", value)
	}
	if string(result) != "hello" {
		t.Errorf("Expected 'hello', got %s", result)
	}
}

// TestAttributeStringEncode tests attributeString Encode
func TestAttributeStringEncode(t *testing.T) {
	codec := attributeString{}

	// []byte value
	data, err := codec.Encode(nil, []byte("hello"))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("Expected 'hello', got %s", data)
	}

	// String value
	data, err = codec.Encode(nil, "world")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if string(data) != "world" {
		t.Errorf("Expected 'world', got %s", data)
	}

	// Invalid type
	_, err = codec.Encode(nil, 12345)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestAttributeStringGetCodeName tests attributeString GetCodeName
func TestAttributeStringGetCodeName(t *testing.T) {
	codec := attributeString{}
	if codec.GetCodeName() != "AttributeString" {
		t.Errorf("Expected 'AttributeString', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeAddressDecode tests attributeAddress Decode
func TestAttributeAddressDecode(t *testing.T) {
	codec := attributeAddress{}

	// Valid IPv4
	ip := net.ParseIP("192.168.1.1").To4()
	value, err := codec.Decode(nil, ip)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	result, ok := value.(net.IP)
	if !ok {
		t.Fatalf("Expected net.IP, got %T", value)
	}
	if !result.Equal(net.ParseIP("192.168.1.1")) {
		t.Errorf("Expected 192.168.1.1, got %s", result)
	}

	// Invalid length
	_, err = codec.Decode(nil, []byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for invalid address length")
	}
}

// TestAttributeAddressEncode tests attributeAddress Encode
func TestAttributeAddressEncode(t *testing.T) {
	codec := attributeAddress{}

	// Valid IPv4
	ip := net.ParseIP("192.168.1.1")
	data, err := codec.Encode(nil, ip)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if !bytes.Equal(data, ip.To4()) {
		t.Errorf("Expected %v, got %v", ip.To4(), data)
	}

	// Invalid type
	_, err = codec.Encode(nil, "not an IP")
	if err == nil {
		t.Error("Expected error for invalid type")
	}

	// IPv6 address (should fail)
	ipv6 := net.ParseIP("::1")
	_, err = codec.Encode(nil, ipv6)
	if err == nil {
		t.Error("Expected error for IPv6 address")
	}
}

// TestAttributeAddressGetCodeName tests attributeAddress GetCodeName
func TestAttributeAddressGetCodeName(t *testing.T) {
	codec := attributeAddress{}
	if codec.GetCodeName() != "AttributeAddress" {
		t.Errorf("Expected 'AttributeAddress', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeIntegerDecode tests attributeInteger Decode
func TestAttributeIntegerDecode(t *testing.T) {
	codec := attributeInteger{}

	// Valid integer
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, 12345)
	value, err := codec.Decode(nil, raw)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if value.(uint32) != 12345 {
		t.Errorf("Expected 12345, got %d", value.(uint32))
	}

	// Invalid length
	_, err = codec.Decode(nil, []byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for invalid integer length")
	}
}

// TestAttributeIntegerEncode tests attributeInteger Encode
func TestAttributeIntegerEncode(t *testing.T) {
	codec := attributeInteger{}

	data, err := codec.Encode(nil, uint32(12345))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("Expected 4 bytes, got %d", len(data))
	}
	if binary.BigEndian.Uint32(data) != 12345 {
		t.Errorf("Expected 12345, got %d", binary.BigEndian.Uint32(data))
	}

	// Invalid type
	_, err = codec.Encode(nil, "not an integer")
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestAttributeIntegerGetCodeName tests attributeInteger GetCodeName
func TestAttributeIntegerGetCodeName(t *testing.T) {
	codec := attributeInteger{}
	if codec.GetCodeName() != "AttributeInteger" {
		t.Errorf("Expected 'AttributeInteger', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeTimeDecode tests attributeTime Decode
func TestAttributeTimeDecode(t *testing.T) {
	codec := attributeTime{}

	// Valid time
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, uint32(1609459200)) // 2021-01-01 00:00:00 UTC
	value, err := codec.Decode(nil, raw)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	result, ok := value.(time.Time)
	if !ok {
		t.Fatalf("Expected time.Time, got %T", value)
	}
	if result.Unix() != 1609459200 {
		t.Errorf("Expected Unix timestamp 1609459200, got %d", result.Unix())
	}

	// Invalid length
	_, err = codec.Decode(nil, []byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for invalid time length")
	}
}

// TestAttributeTimeEncode tests attributeTime Encode
func TestAttributeTimeEncode(t *testing.T) {
	codec := attributeTime{}

	ts := time.Unix(1609459200, 0)
	data, err := codec.Encode(nil, ts)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) != 4 {
		t.Errorf("Expected 4 bytes, got %d", len(data))
	}
	if binary.BigEndian.Uint32(data) != 1609459200 {
		t.Errorf("Expected 1609459200, got %d", binary.BigEndian.Uint32(data))
	}

	// Invalid type
	_, err = codec.Encode(nil, "not a time")
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestAttributeTimeGetCodeName tests attributeTime GetCodeName
func TestAttributeTimeGetCodeName(t *testing.T) {
	codec := attributeTime{}
	if codec.GetCodeName() != "AttributeTime" {
		t.Errorf("Expected 'AttributeTime', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeUnknownDecode tests attributeUnknown Decode
func TestAttributeUnknownDecode(t *testing.T) {
	codec := attributeUnknown{}

	_, err := codec.Decode(nil, []byte("test"))
	if err == nil {
		t.Error("Expected error for unknown attribute decode")
	}
}

// TestAttributeUnknownEncode tests attributeUnknown Encode
func TestAttributeUnknownEncode(t *testing.T) {
	codec := attributeUnknown{}

	_, err := codec.Encode(nil, "test")
	if err == nil {
		t.Error("Expected error for unknown attribute encode")
	}
}

// TestAttributeUnknownGetCodeName tests attributeUnknown GetCodeName
func TestAttributeUnknownGetCodeName(t *testing.T) {
	codec := attributeUnknown{}
	if codec.GetCodeName() != "AttributeUnknown" {
		t.Errorf("Expected 'AttributeUnknown', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeVendorDecode tests attributeVendor Decode
func TestAttributeVendorDecode(t *testing.T) {
	codec := attributeVendor{}

	// Valid VSA
	vsaData := EncodeAVPair(28557, 1, "testvalue")
	value, err := codec.Decode(nil, vsaData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if value != "testvalue" {
		t.Errorf("Expected 'testvalue', got %v", value)
	}

	// Invalid VSA (too short)
	_, err = codec.Decode(nil, []byte{0x00, 0x00, 0x00})
	if err != nil {
		// Should fall back to text decode
		t.Logf("Decode with short VSA returned error: %v", err)
	}
}

// TestAttributeVendorEncode tests attributeVendor Encode
func TestAttributeVendorEncode(t *testing.T) {
	codec := attributeVendor{}

	// With vendor ID set
	VerdorID = 28557
	VerdorTypeID = 1
	data, err := codec.Encode(nil, "testvalue")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	// Without vendor ID set
	VerdorID = 0
	data, err = codec.Encode(nil, "testvalue")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if string(data) != "testvalue" {
		t.Errorf("Expected 'testvalue', got '%s'", data)
	}
}

// TestAttributeVendorGetCodeName tests attributeVendor GetCodeName
func TestAttributeVendorGetCodeName(t *testing.T) {
	codec := attributeVendor{}
	if codec.GetCodeName() != "AttributeVendor" {
		t.Errorf("Expected 'AttributeVendor', got '%s'", codec.GetCodeName())
	}
}

// TestAttributeVendorDecodeLargeVendorID tests attributeVendor Decode with large vendor ID
func TestAttributeVendorDecodeLargeVendorID(t *testing.T) {
	codec := attributeVendor{}

	// VSA with vendor ID > 99999 should fall back to text decode
	vsaData := EncodeAVPair(999999, 1, "testvalue")
	value, err := codec.Decode(nil, vsaData)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	// Should fall back to text decode
	strValue, ok := value.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", value)
	}
	// The value should be the raw bytes as string since it falls back to text
	_ = strValue
}

// TestAttributeCodecInterfaces tests that all codecs implement IAttributeCodec
func TestAttributeCodecInterfaces(t *testing.T) {
	var _ IAttributeCodec = attributeText{}
	var _ IAttributeCodec = attributeString{}
	var _ IAttributeCodec = attributeAddress{}
	var _ IAttributeCodec = attributeInteger{}
	var _ IAttributeCodec = attributeTime{}
	var _ IAttributeCodec = attributeUnknown{}
	var _ IAttributeCodec = attributeVendor{}
}

// TestAttributeGlobalVariables tests that global attribute variables are initialized
func TestAttributeGlobalVariables(t *testing.T) {
	if AttributeText == nil {
		t.Error("AttributeText should not be nil")
	}
	if AttributeString == nil {
		t.Error("AttributeString should not be nil")
	}
	if AttributeAddress == nil {
		t.Error("AttributeAddress should not be nil")
	}
	if AttributeInteger == nil {
		t.Error("AttributeInteger should not be nil")
	}
	if AttributeTime == nil {
		t.Error("AttributeTime should not be nil")
	}
	if AttributeUnknown == nil {
		t.Error("AttributeUnknown should not be nil")
	}
	if AttributeVendor == nil {
		t.Error("AttributeVendor should not be nil")
	}
}

// TestAttributeStringDecodeCopy tests that attributeString Decode makes a copy
func TestAttributeStringDecodeCopy(t *testing.T) {
	codec := attributeString{}

	original := []byte("hello")
	value, err := codec.Decode(nil, original)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	result := value.([]byte)

	// Modify original - should not affect decoded value
	original[0] = 'x'
	if result[0] == 'x' {
		t.Error("Decode should return a copy, not a reference")
	}
}

// TestAttributeAddressDecodeCopy tests that attributeAddress Decode makes a copy
func TestAttributeAddressDecodeCopy(t *testing.T) {
	codec := attributeAddress{}

	original := net.ParseIP("192.168.1.1").To4()
	value, err := codec.Decode(nil, original)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	result := value.(net.IP)

	// Modify original - should not affect decoded value
	original[0] = 10
	if result[0] == 10 {
		t.Error("Decode should return a copy, not a reference")
	}
}
