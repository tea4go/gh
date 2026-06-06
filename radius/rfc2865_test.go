package radius

import (
	"bytes"
	"crypto/md5"
	"testing"
)

// TestRFC2865UserPasswordDecode tests rfc2865UserPassword Decode
func TestRFC2865UserPasswordDecode(t *testing.T) {
	codec := rfc2865UserPassword{}

	secret := []byte("testsecret")
	packet := &TDataPacket{
		Secret: secret,
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	// First encode a password
	password := "mypassword"
	encoded, err := codec.Encode(packet, password)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Now decode it
	decoded, err := codec.Decode(packet, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded != password {
		t.Errorf("Expected '%s', got '%s'", password, decoded)
	}
}

// TestRFC2865UserPasswordDecodeNilSecret tests Decode with nil secret
func TestRFC2865UserPasswordDecodeNilSecret(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: nil,
	}

	_, err := codec.Decode(packet, make([]byte, 16))
	if err == nil {
		t.Error("Expected error for nil secret")
	}
}

// TestRFC2865UserPasswordDecodeInvalidLength tests Decode with invalid length
func TestRFC2865UserPasswordDecodeInvalidLength(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: []byte("secret"),
	}

	// Wrong length (not 16)
	_, err := codec.Decode(packet, make([]byte, 10))
	if err == nil {
		t.Error("Expected error for invalid length")
	}

	// Also test with 0 length
	_, err = codec.Decode(packet, make([]byte, 0))
	if err == nil {
		t.Error("Expected error for zero length")
	}
}

// TestRFC2865UserPasswordEncodeNilSecret tests Encode with nil secret
func TestRFC2865UserPasswordEncodeNilSecret(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: nil,
	}

	_, err := codec.Encode(packet, "password")
	if err == nil {
		t.Error("Expected error for nil secret")
	}
}

// TestRFC2865UserPasswordEncodeString tests Encode with string password
func TestRFC2865UserPasswordEncodeString(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: []byte("secret"),
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	encoded, err := codec.Encode(packet, "password")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(encoded) != 16 {
		t.Errorf("Expected 16 bytes, got %d", len(encoded))
	}
}

// TestRFC2865UserPasswordEncodeBytes tests Encode with []byte password
func TestRFC2865UserPasswordEncodeBytes(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: []byte("secret"),
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	encoded, err := codec.Encode(packet, []byte("password"))
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(encoded) != 16 {
		t.Errorf("Expected 16 bytes, got %d", len(encoded))
	}
}

// TestRFC2865UserPasswordEncodeInvalidType tests Encode with invalid type
func TestRFC2865UserPasswordEncodeInvalidType(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: []byte("secret"),
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	_, err := codec.Encode(packet, 12345)
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

// TestRFC2865UserPasswordEncodeTooLong tests Encode with password too long
func TestRFC2865UserPasswordEncodeTooLong(t *testing.T) {
	codec := rfc2865UserPassword{}

	packet := &TDataPacket{
		Secret: []byte("secret"),
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	longPassword := make([]byte, 20)
	_, err := codec.Encode(packet, longPassword)
	if err == nil {
		t.Error("Expected error for password too long")
	}
}

// TestRFC2865UserPasswordGetCodeName tests GetCodeName
func TestRFC2865UserPasswordGetCodeName(t *testing.T) {
	codec := rfc2865UserPassword{}
	if codec.GetCodeName() != "RFC2865UserPassword" {
		t.Errorf("Expected 'RFC2865UserPassword', got '%s'", codec.GetCodeName())
	}
}

// TestRFC2865UserPasswordDecodeWithNullPadding tests Decode with null-padded password
func TestRFC2865UserPasswordDecodeWithNullPadding(t *testing.T) {
	codec := rfc2865UserPassword{}

	secret := []byte("testsecret")
	packet := &TDataPacket{
		Secret: secret,
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	// Manually create an encoded password with null padding
	password := "hi"
	passwordBytes := []byte(password)
	// Pad to 16 bytes
	padded := make([]byte, 16)
	copy(padded, passwordBytes)

	// Encrypt
	var mask [md5.Size]byte
	hash := md5.New()
	hash.Write(secret)
	hash.Write(packet.Authenticator[:])
	hash.Sum(mask[0:0])

	for i, b := range padded {
		padded[i] = b ^ mask[i]
	}

	// Decode
	decoded, err := codec.Decode(packet, padded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded != "hi" {
		t.Errorf("Expected 'hi', got '%s'", decoded)
	}
}

// TestRFC2865UserPasswordDecodeWithoutNullPadding tests Decode without null padding
func TestRFC2865UserPasswordDecodeWithoutNullPadding(t *testing.T) {
	codec := rfc2865UserPassword{}

	secret := []byte("testsecret")
	packet := &TDataPacket{
		Secret: secret,
	}
	copy(packet.Authenticator[:], make([]byte, 16))

	// Encode a password that fills all 16 bytes
	password := "1234567890123456" // Exactly 16 chars
	encoded, err := codec.Encode(packet, password)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode should return full string without truncation
	decoded, err := codec.Decode(packet, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if decoded != password {
		t.Errorf("Expected '%s', got '%s'", password, decoded)
	}
}

// TestRFC2865UserPasswordEncodeDecodeRoundTrip tests encode/decode round trip
func TestRFC2865UserPasswordEncodeDecodeRoundTrip(t *testing.T) {
	codec := rfc2865UserPassword{}

	passwords := []string{
		"a",
		"test",
		"mypassword",
		"1234567890123456", // Max length
		"",
	}

	for _, password := range passwords {
		secret := []byte("testsecret")
		packet := &TDataPacket{
			Secret: secret,
		}
		copy(packet.Authenticator[:], make([]byte, 16))

		if password == "" {
			continue // Empty password edge case
		}

		encoded, err := codec.Encode(packet, password)
		if err != nil {
			t.Errorf("Encode failed for '%s': %v", password, err)
			continue
		}

		decoded, err := codec.Decode(packet, encoded)
		if err != nil {
			t.Errorf("Decode failed for '%s': %v", password, err)
			continue
		}

		if decoded != password {
			t.Errorf("Round trip failed: expected '%s', got '%s'", password, decoded)
		}
	}
}

// TestRFC2865UserPasswordXOR tests the XOR operation manually
func TestRFC2865UserPasswordXOR(t *testing.T) {
	secret := []byte("sharedsecret")
	authenticator := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	packet := &TDataPacket{
		Secret:        secret,
		Authenticator: authenticator,
	}

	codec := rfc2865UserPassword{}

	password := "testpass"
	encoded, err := codec.Encode(packet, password)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify the encoding manually
	var mask [md5.Size]byte
	hash := md5.New()
	hash.Write(secret)
	hash.Write(authenticator[:])
	hash.Sum(mask[0:0])

	expected := make([]byte, 16)
	copy(expected, []byte(password))
	for i, b := range expected {
		expected[i] = b ^ mask[i]
	}

	if !bytes.Equal(encoded, expected) {
		t.Errorf("XOR encoding mismatch")
	}
}

// TestRFC2865Registration tests that User-Password is registered with rfc2865UserPassword codec
func TestRFC2865Registration(t *testing.T) {
	codec := Builtin.GetFunc(2)
	if codec == nil {
		t.Fatal("GetFunc returned nil for User-Password")
	}
	if codec.GetCodeName() != "RFC2865UserPassword" {
		t.Errorf("Expected RFC2865UserPassword, got %s", codec.GetCodeName())
	}
}
