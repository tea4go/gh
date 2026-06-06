package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGUIDFromString(t *testing.T) {
	fmt.Println("=== 开始测试: GUID 字符串解析 (TestGUIDFromString) ===")
	// Valid GUID
	s := "11223344-5566-7788-9900-aabbccddeeff"
	g, err := GUIDFromString(s)
	if err != nil {
		t.Fatalf("GUIDFromString failed for valid GUID: %v", err)
	}

	if g.String() != s {
		t.Errorf("GUID string representation mismatch. Got %s, want %s", g.String(), s)
	}

	// Invalid Length
	_, err = GUIDFromString("123")
	if err == nil {
		t.Error("Expected error for short string, got nil")
	}

	// Invalid Format
	_, err = GUIDFromString("11223344-5566-7788-9900-aabbccddeefg") // 'g' is invalid hex
	if err == nil {
		t.Error("Expected error for invalid hex char, got nil")
	}
}

func TestGUIDFromBytes(t *testing.T) {
	fmt.Println("=== 开始测试: GUID 字节解析 (TestGUIDFromBytes) ===")
	// 16 bytes
	b, _ := hex.DecodeString("44332211665588779900aabbccddeeff") 
	// Note: GUID uses Little Endian for first 3 parts.
	// Data1: 11223344 (4 bytes) -> 44 33 22 11
	// Data2: 5566 (2 bytes) -> 66 55
	// Data3: 7788 (2 bytes) -> 88 77
	// Data4: 9900aabbccddeeff (8 bytes) -> 99 00 aa bb cc dd ee ff (Big Endian / Array order)
	
	// So input bytes for "11223344-5566-7788-9900-aabbccddeeff" should be:
	// 44 33 22 11 | 66 55 | 88 77 | 99 00 aa bb cc dd ee ff
	
	g := GUIDFromBytes(b)
	expected := "11223344-5566-7788-9900-aabbccddeeff"
	if g.String() != expected {
		t.Errorf("GUIDFromBytes failed. Got %s, want %s", g.String(), expected)
	}
}

func TestGUIDOctetString(t *testing.T) {
	fmt.Println("=== 开始测试: GUID OctetString (TestGUIDOctetString) ===")
	s := "11223344-5566-7788-9900-aabbccddeeff"
	g, _ := GUIDFromString(s)
	
	octet := g.OctetString()
	// Should be backslashed hex of the internal byte array (little endian structure)
	// Bytes: 44 33 22 11 66 55 88 77 99 00 aa bb cc dd ee ff
	expected := "\\44\\33\\22\\11\\66\\55\\88\\77\\99\\00\\aa\\bb\\cc\\dd\\ee\\ff"
	
	if octet != expected {
		t.Errorf("OctetString mismatch. Got %s, want %s", octet, expected)
	}
}

func TestGUIDToWindowsArray(t *testing.T) {
	fmt.Println("=== 开始测试: GUID Windows Array 转换 (TestGUIDToWindowsArray) ===")
	s := "11223344-5566-7788-9900-aabbccddeeff"
	g, _ := GUIDFromString(s)

	arr := g.ToWindowsArray()

	expectedBytes, _ := hex.DecodeString("44332211665588779900aabbccddeeff")

	for i := 0; i < 16; i++ {
		if arr[i] != expectedBytes[i] {
			t.Errorf("Byte at index %d mismatch. Got %x, want %x", i, arr[i], expectedBytes[i])
		}
	}
}

func TestGUIDFromWindowsArray(t *testing.T) {
	fmt.Println("=== 开始测试: GUIDFromWindowsArray ===")
	var b [16]byte
	b[0], b[1], b[2], b[3] = 0x44, 0x33, 0x22, 0x11
	b[4], b[5] = 0x66, 0x55
	b[6], b[7] = 0x88, 0x77
	copy(b[8:], []byte{0x99, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff})

	g := GUIDFromWindowsArray(b)
	expected := "11223344-5566-7788-9900-aabbccddeeff"
	if g.String() != expected {
		t.Errorf("GUIDFromWindowsArray failed. Got %s, want %s", g.String(), expected)
	}
}

func TestGUIDFromStringInvalidFormat(t *testing.T) {
	fmt.Println("=== 开始测试: GUIDFromString 各种无效格式 ===")
	tests := []string{
		"",                            // too short
		"12345678",                    // too short
		"11223344556677889900aabbccddeeff", // missing dashes
		"1122334-5566-7788-9900-aabbccddeeff", // wrong first dash position
		"11223344-5566-7788-9900-aabbccddeefz", // 'z' invalid hex in last part
		"11223344-5566-7788-9900-aabbccddeef!", // '!' invalid
	}
	for _, s := range tests {
		_, err := GUIDFromString(s)
		if err == nil {
			t.Errorf("Expected error for GUID %q, got nil", s)
		}
	}
}

func TestGUIDFromStringInvalidHex(t *testing.T) {
	fmt.Println("=== 开始测试: GUIDFromString 无效十六进制 ===")
	// Valid format but invalid hex in data1
	_, err := GUIDFromString("zzzzzzzz-5566-7788-9900-aabbccddeeff")
	if err == nil {
		t.Error("Expected error for invalid hex in data1")
	}

	// Invalid hex in data2
	_, err = GUIDFromString("11223344-zzzz-7788-9900-aabbccddeeff")
	if err == nil {
		t.Error("Expected error for invalid hex in data2")
	}

	// Invalid hex in data3
	_, err = GUIDFromString("11223344-5566-zzzz-9900-aabbccddeeff")
	if err == nil {
		t.Error("Expected error for invalid hex in data3")
	}

	// Invalid hex in Data4 part (byte positions)
	_, err = GUIDFromString("11223344-5566-7788-zz00-aabbccddeeff")
	if err == nil {
		t.Error("Expected error for invalid hex in data4")
	}
}
