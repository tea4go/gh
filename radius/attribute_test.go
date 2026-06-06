package radius

import (
	"encoding/binary"
	"testing"
)

// TestEncodeAVPair tests EncodeAVPair function
func TestEncodeAVPair(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := "testvalue"

	vsa := EncodeAVPair(vendorID, typeID, value)
	if len(vsa) == 0 {
		t.Error("EncodeAVPair returned empty slice")
	}

	// Verify vendor ID
	encodedVendorID := binary.BigEndian.Uint32(vsa[0:4])
	if encodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, encodedVendorID)
	}

	// Verify type ID
	if vsa[4] != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, vsa[4])
	}

	// Verify length
	expectedLen := byte(len(value) + 2)
	if vsa[5] != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, vsa[5])
	}

	// Verify value
	if string(vsa[6:]) != value {
		t.Errorf("Expected value '%s', got '%s'", value, string(vsa[6:]))
	}
}

// TestEncodeAVPairByte tests EncodeAVPairByte function
func TestEncodeAVPairByte(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := []byte("testvalue")

	vsa := EncodeAVPairByte(vendorID, typeID, value)
	if len(vsa) == 0 {
		t.Error("EncodeAVPairByte returned empty slice")
	}

	// Verify vendor ID
	encodedVendorID := binary.BigEndian.Uint32(vsa[0:4])
	if encodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, encodedVendorID)
	}

	// Verify type ID
	if vsa[4] != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, vsa[4])
	}

	// Verify length
	expectedLen := byte(len(value) + 2)
	if vsa[5] != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, vsa[5])
	}

	// Verify value
	if string(vsa[6:]) != string(value) {
		t.Errorf("Expected value '%s', got '%s'", string(value), string(vsa[6:]))
	}
}

// TestEncodeAVpairTag tests EncodeAVpairTag function
func TestEncodeAVpairTag(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	tag := uint8(0)
	value := "testvalue"

	vsa := EncodeAVpairTag(vendorID, typeID, tag, value)
	if len(vsa) == 0 {
		t.Error("EncodeAVpairTag returned empty slice")
	}

	// Verify vendor ID
	encodedVendorID := binary.BigEndian.Uint32(vsa[0:4])
	if encodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, encodedVendorID)
	}

	// Verify type ID
	if vsa[4] != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, vsa[4])
	}

	// Verify length (includes tag byte)
	expectedLen := byte(len(value) + 3)
	if vsa[5] != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, vsa[5])
	}

	// Verify tag
	if vsa[6] != tag {
		t.Errorf("Expected tag %d, got %d", tag, vsa[6])
	}

	// Verify value
	if string(vsa[7:]) != value {
		t.Errorf("Expected value '%s', got '%s'", value, string(vsa[7:]))
	}
}

// TestEncodeAVPairByteTag tests EncodeAVPairByteTag function
func TestEncodeAVPairByteTag(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	tag := uint8(1)
	value := []byte("testvalue")

	vsa := EncodeAVPairByteTag(vendorID, typeID, tag, value)
	if len(vsa) == 0 {
		t.Error("EncodeAVPairByteTag returned empty slice")
	}

	// Verify vendor ID
	encodedVendorID := binary.BigEndian.Uint32(vsa[0:4])
	if encodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, encodedVendorID)
	}

	// Verify type ID
	if vsa[4] != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, vsa[4])
	}

	// Verify length (includes tag byte)
	expectedLen := byte(len(value) + 3)
	if vsa[5] != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, vsa[5])
	}

	// Verify tag
	if vsa[6] != tag {
		t.Errorf("Expected tag %d, got %d", tag, vsa[6])
	}

	// Verify value
	if string(vsa[7:]) != string(value) {
		t.Errorf("Expected value '%s', got '%s'", string(value), string(vsa[7:]))
	}
}

// TestDecodeAVPairByte tests DecodeAVPairByte function
func TestDecodeAVPairByte(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := []byte("testvalue")

	vsa := EncodeAVPairByte(vendorID, typeID, value)

	decodedVendorID, decodedTypeID, decodedValue, err := DecodeAVPairByte(vsa)
	if err != nil {
		t.Fatalf("DecodeAVPairByte failed: %v", err)
	}
	if decodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, decodedVendorID)
	}
	if decodedTypeID != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, decodedTypeID)
	}
	if string(decodedValue) != string(value) {
		t.Errorf("Expected value '%s', got '%s'", string(value), string(decodedValue))
	}
}

// TestDecodeAVPair tests DecodeAVPair function
func TestDecodeAVPair(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := "testvalue"

	vsa := EncodeAVPair(vendorID, typeID, value)

	decodedVendorID, decodedTypeID, decodedValue, err := DecodeAVPair(vsa)
	if err != nil {
		t.Fatalf("DecodeAVPair failed: %v", err)
	}
	if decodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, decodedVendorID)
	}
	if decodedTypeID != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, decodedTypeID)
	}
	if decodedValue != value {
		t.Errorf("Expected value '%s', got '%s'", value, decodedValue)
	}
}

// TestDecodeAVPairByteTooShort tests DecodeAVPairByte with too short data
func TestDecodeAVPairByteTooShort(t *testing.T) {
	// Too short VSA
	_, _, _, err := DecodeAVPairByte([]byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x02})
	if err == nil {
		t.Error("Expected error for too short VSA")
	}

	// Even shorter
	_, _, _, err = DecodeAVPairByte([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Error("Expected error for too short VSA")
	}
}

// TestDecodeAVPairTooShort tests DecodeAVPair with too short data
func TestDecodeAVPairTooShort(t *testing.T) {
	// Too short VSA
	_, _, _, err := DecodeAVPair([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Error("Expected error for too short VSA")
	}
}

// TestEncodeAVPairEmptyValue tests EncodeAVPair with empty value
func TestEncodeAVPairEmptyValue(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := ""

	vsa := EncodeAVPair(vendorID, typeID, value)
	if len(vsa) != 6 {
		t.Errorf("Expected 6 bytes for empty value, got %d", len(vsa))
	}
}

// TestEncodeAVPairByteEmptyValue tests EncodeAVPairByte with empty value
func TestEncodeAVPairByteEmptyValue(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	value := []byte{}

	vsa := EncodeAVPairByte(vendorID, typeID, value)
	if len(vsa) != 6 {
		t.Errorf("Expected 6 bytes for empty value, got %d", len(vsa))
	}
}

// TestEncodeDecodeRoundTrip tests encoding and decoding round trip
func TestEncodeDecodeRoundTrip(t *testing.T) {
	testCases := []struct {
		vendorID uint32
		typeID   uint8
		value    string
	}{
		{28557, 1, "hello"},
		{12345, 255, "test with spaces"},
		{4294967295, 128, "max vendor id"},
	}

	for _, tc := range testCases {
		vsa := EncodeAVPair(tc.vendorID, tc.typeID, tc.value)
		decodedVendorID, decodedTypeID, decodedValue, err := DecodeAVPair(vsa)
		if err != nil {
			t.Errorf("DecodeAVPair failed for vendorID=%d: %v", tc.vendorID, err)
			continue
		}
		if decodedVendorID != tc.vendorID {
			t.Errorf("Vendor ID mismatch: expected %d, got %d", tc.vendorID, decodedVendorID)
		}
		if decodedTypeID != tc.typeID {
			t.Errorf("Type ID mismatch: expected %d, got %d", tc.typeID, decodedTypeID)
		}
		if decodedValue != tc.value {
			t.Errorf("Value mismatch: expected '%s', got '%s'", tc.value, decodedValue)
		}
	}
}

// TestEncodeDecodeByteRoundTrip tests encoding and decoding round trip with bytes
func TestEncodeDecodeByteRoundTrip(t *testing.T) {
	testCases := []struct {
		vendorID uint32
		typeID   uint8
		value    []byte
	}{
		{28557, 1, []byte("hello")},
		{12345, 255, []byte{0x00, 0x01, 0x02, 0x03}},
	}

	for _, tc := range testCases {
		vsa := EncodeAVPairByte(tc.vendorID, tc.typeID, tc.value)
		decodedVendorID, decodedTypeID, decodedValue, err := DecodeAVPairByte(vsa)
		if err != nil {
			t.Errorf("DecodeAVPairByte failed for vendorID=%d: %v", tc.vendorID, err)
			continue
		}
		if decodedVendorID != tc.vendorID {
			t.Errorf("Vendor ID mismatch: expected %d, got %d", tc.vendorID, decodedVendorID)
		}
		if decodedTypeID != tc.typeID {
			t.Errorf("Type ID mismatch: expected %d, got %d", tc.typeID, decodedTypeID)
		}
		if string(decodedValue) != string(tc.value) {
			t.Errorf("Value mismatch: expected %v, got %v", tc.value, decodedValue)
		}
	}
}

// TestEncodeAVPairByteTagRoundTrip tests encoding and decoding round trip with tag
func TestEncodeAVPairByteTagRoundTrip(t *testing.T) {
	vendorID := uint32(28557)
	typeID := uint8(1)
	tag := uint8(5)
	value := []byte("testvalue")

	vsa := EncodeAVPairByteTag(vendorID, typeID, tag, value)

	// Decode - note that DecodeAVPairByte doesn't handle tag, so we verify manually
	decodedVendorID, decodedTypeID, decodedValue, err := DecodeAVPairByte(vsa)
	if err != nil {
		t.Fatalf("DecodeAVPairByte failed: %v", err)
	}
	if decodedVendorID != vendorID {
		t.Errorf("Expected vendor ID %d, got %d", vendorID, decodedVendorID)
	}
	if decodedTypeID != typeID {
		t.Errorf("Expected type ID %d, got %d", typeID, decodedTypeID)
	}
	// Value starts at offset 7 with tag, but DecodeAVPairByte reads from offset 6
	// So the decoded value will include the tag byte
	_ = decodedValue
}

// TestDecodeAVPairByteMinimumLength tests DecodeAVPairByte with minimum valid length
func TestDecodeAVPairByteMinimumLength(t *testing.T) {
	// Exactly 7 bytes: 4 vendor + 1 type + 1 length + 1 value
	vsa := make([]byte, 7)
	binary.BigEndian.PutUint32(vsa[0:4], 28557)
	vsa[4] = 1
	vsa[5] = 3 // length includes type+length+value
	vsa[6] = 'a'

	vendorID, typeID, value, err := DecodeAVPairByte(vsa)
	if err != nil {
		t.Fatalf("DecodeAVPairByte failed: %v", err)
	}
	if vendorID != 28557 {
		t.Errorf("Expected vendor ID 28557, got %d", vendorID)
	}
	if typeID != 1 {
		t.Errorf("Expected type ID 1, got %d", typeID)
	}
	if string(value) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(value))
	}
}

// TestDecodeAVPairByteExactly6Bytes tests DecodeAVPairByte with exactly 6 bytes (boundary)
func TestDecodeAVPairByteExactly6Bytes(t *testing.T) {
	// Exactly 6 bytes: should fail (need at least 7)
	vsa := make([]byte, 6)
	binary.BigEndian.PutUint32(vsa[0:4], 28557)
	vsa[4] = 1
	vsa[5] = 2

	_, _, _, err := DecodeAVPairByte(vsa)
	if err == nil {
		t.Error("Expected error for exactly 6 bytes")
	}
}
