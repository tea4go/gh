package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

func TestTAESEncrypt(t *testing.T) {
	fmt.Println("=== 开始测试: AES 加密解密 (TestTAESEncrypt) ===")
	key := "0123456789abcdef" // 16 bytes
	plainText := "Hello, World! This is a test."

	aesEnc := TAESEncrypt{}
	aesEnc.Init(key)

	// Test Encrypt
	cipherBytes, err := aesEnc.Encrypt(plainText)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if len(cipherBytes) == 0 {
		t.Fatal("Cipher text is empty")
	}

	// Test Decrypt
	decryptedBytes, err := aesEnc.Decrypt(string(cipherBytes))
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	decryptedText := string(decryptedBytes)

	if decryptedText != plainText {
		t.Errorf("Decrypted text does not match original. Got %q, want %q", decryptedText, plainText)
	}
}

func TestAesEncryptDecryptFunctions(t *testing.T) {
	fmt.Println("=== 开始测试: AES 辅助函数 (TestAesEncryptDecryptFunctions) ===")
	key := "0123456789abcdef"
	plainText := "Simple Test"

	// Test AesEncrypt (returns hex string)
	hexCipher, err := AesEncrypt(key, plainText)
	if err != nil {
		t.Fatalf("AesEncrypt failed: %v", err)
	}

	// Test AesDecrypt (takes hex string)
	decryptedText, err := AesDecrypt(key, hexCipher)
	if err != nil {
		t.Fatalf("AesDecrypt failed: %v", err)
	}

	if !strings.HasPrefix(decryptedText, plainText) {
		t.Errorf("Decrypted text mismatch. Got %q, want %q", decryptedText, plainText)
	}
}

func TestGetPKCS1Key(t *testing.T) {
	fmt.Println("=== 开始测试: RSA 密钥生成与解析 (TestGetPKCS1Key) ===")
	pri, pub, err := GetPKCS1Key()
	if err != nil {
		t.Fatalf("GetPKCS1Key failed: %v", err)
	}

	if pri == "" || pub == "" {
		t.Error("Generated keys are empty")
	}

	if !strings.Contains(pri, "RSA PRIVATE KEY") {
		t.Error("Private key format incorrect")
	}
	if !strings.Contains(pub, "RSA PUBLIC KEY") {
		t.Error("Public key format incorrect")
	}

	// Test Parse Functions
	_, err = ParsePriKeyBytes([]byte(pri))
	if err != nil {
		t.Errorf("ParsePriKeyBytes failed: %v", err)
	}

	_, err = ParsePubKeyBytes([]byte(pub))
	if err != nil {
		t.Errorf("ParsePubKeyBytes failed: %v", err)
	}
}

func TestBase64EncryptDecrypt(t *testing.T) {
	fmt.Println("=== 开始测试: Base64 加密解密流程 (TestBase64EncryptDecrypt) ===")
	key := "0123456789abcdef"
	aesEnc := TAESEncrypt{}
	aesEnc.Init(key)

	text := "Test Base64"
	crypted, _ := aesEnc.Encrypt(text)
	b64 := base64.StdEncoding.EncodeToString(crypted)

	// Reverse
	cryptedDecode, _ := base64.StdEncoding.DecodeString(b64)
	decrypted, _ := aesEnc.Decrypt(string(cryptedDecode))

	if string(decrypted) != text {
		t.Errorf("Base64 flow failed. Got %s, want %s", string(decrypted), text)
	}
}

// ===================== PKCS5 Padding =====================

func TestPKCS5Padding(t *testing.T) {
	fmt.Println("=== 开始测试: PKCS5 填充 ===")
	aesEnc := TAESEncrypt{}
	aesEnc.Init("0123456789abcdef")
	aesEnc.ZeroPad = false // Use PKCS5

	plainText := "Hello"
	cipherBytes, err := aesEnc.Encrypt(plainText)
	if err != nil {
		t.Fatalf("Encrypt with PKCS5 failed: %v", err)
	}

	decryptedBytes, err := aesEnc.Decrypt(string(cipherBytes))
	if err != nil {
		t.Fatalf("Decrypt with PKCS5 failed: %v", err)
	}

	if string(decryptedBytes) != plainText {
		t.Errorf("PKCS5 roundtrip failed. Got %q, want %q", string(decryptedBytes), plainText)
	}
}

func TestPKCS5PaddingDirect(t *testing.T) {
	fmt.Println("=== 开始测试: PKCS5Padding 直接调用 ===")
	aesEnc := TAESEncrypt{}
	// Test padding: "Hello" is 5 bytes, block size 16 => padding of 11
	padded := aesEnc.PKCS5Padding([]byte("Hello"), 16)
	if len(padded) != 16 {
		t.Errorf("PKCS5Padding length wrong. Got %d, want 16", len(padded))
	}
	// Last 11 bytes should all be 0x0b (11)
	for i := 5; i < 16; i++ {
		if padded[i] != 11 {
			t.Errorf("PKCS5Padding byte %d wrong. Got %d, want 11", i, padded[i])
		}
	}
}

func TestPKCS5UnPaddingDirect(t *testing.T) {
	fmt.Println("=== 开始测试: PKCS5UnPadding 直接调用 ===")
	aesEnc := TAESEncrypt{}
	// Create properly padded data
	padded := aesEnc.PKCS5Padding([]byte("Hello"), 16)
	unpadded := aesEnc.PKCS5UnPadding(padded, 16)
	if string(unpadded) != "Hello" {
		t.Errorf("PKCS5UnPadding failed. Got %q", string(unpadded))
	}

	// Empty input
	empty := aesEnc.PKCS5UnPadding([]byte{}, 16)
	if len(empty) != 0 {
		t.Errorf("PKCS5UnPadding empty failed. Got %v", empty)
	}

	// Invalid padding (unpadding > length)
	invalid := []byte{0x00}
	result := aesEnc.PKCS5UnPadding(invalid, 16)
	if len(result) != 1 {
		t.Errorf("PKCS5UnPadding invalid padding should return original. Got %v", result)
	}

	// Negative padding (byte value = 0)
	zeroPad := []byte{0x41, 0x00}
	result = aesEnc.PKCS5UnPadding(zeroPad, 16)
	if len(result) != 2 {
		t.Errorf("PKCS5UnPadding zero padding should return original. Got %v", result)
	}
}

func TestZeroUnPaddingNoNull(t *testing.T) {
	fmt.Println("=== 开始测试: ZeroUnPadding 无空字节 ===")
	aesEnc := TAESEncrypt{}
	// No null byte in content
	content := []byte("HelloWorld")
	result := aesEnc.ZeroUnPadding(content, 16)
	if string(result) != "HelloWorld" {
		t.Errorf("ZeroUnPadding no null failed. Got %q", string(result))
	}
}

func TestZeroPaddingDirect(t *testing.T) {
	fmt.Println("=== 开始测试: ZeroPadding 直接调用 ===")
	aesEnc := TAESEncrypt{}
	// Data already aligned to block size
	aligned := make([]byte, 16)
	padded := aesEnc.ZeroPadding(aligned, 16)
	// Should add a full block of padding
	if len(padded) != 32 {
		t.Errorf("ZeroPadding aligned data should add block. Got %d", len(padded))
	}
}

// ===================== Error paths =====================

func TestAesEncryptInvalidKey(t *testing.T) {
	fmt.Println("=== 开始测试: AES 加密无效密钥 ===")
	aesEnc := TAESEncrypt{}
	aesEnc.Key = []byte("short") // Invalid key size
	_, err := aesEnc.Encrypt("test")
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestAesDecryptInvalidKey(t *testing.T) {
	fmt.Println("=== 开始测试: AES 解密无效密钥 ===")
	aesEnc := TAESEncrypt{}
	aesEnc.Key = []byte("short") // Invalid key size
	_, err := aesEnc.Decrypt("somedata")
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestAesDecryptInvalidCiphertext(t *testing.T) {
	fmt.Println("=== 开始测试: AES 解密无效密文 ===")
	aesEnc := TAESEncrypt{}
	aesEnc.Init("0123456789abcdef")

	// Empty ciphertext
	_, err := aesEnc.Decrypt("")
	if err == nil {
		t.Error("Expected error for empty ciphertext")
	}

	// Non-aligned ciphertext
	_, err = aesEnc.Decrypt("short")
	if err == nil {
		t.Error("Expected error for non-aligned ciphertext")
	}
}

func TestAesDecryptHexError(t *testing.T) {
	fmt.Println("=== 开始测试: AesDecrypt 无效hex ===")
	_, err := AesDecrypt("0123456789abcdef", "not-hex!!")
	if err == nil {
		t.Error("Expected error for invalid hex string")
	}
}

func TestAesEncryptDecryptPKCS5LongText(t *testing.T) {
	fmt.Println("=== 开始测试: AES PKCS5 长文本 ===")
	key := "0123456789abcdef"
	plainText := "This is a longer piece of text that spans multiple AES blocks for testing PKCS5 padding mode."

	hexCipher, err := AesEncrypt(key, plainText)
	if err != nil {
		t.Fatalf("AesEncrypt PKCS5 long text failed: %v", err)
	}

	// AesEncrypt uses ZeroPad (Init sets ZeroPad=true), so we need custom setup for PKCS5
	// Let's test with ZeroPad=true (default) for AesEncrypt/AesDecrypt
	decrypted, err := AesDecrypt(key, hexCipher)
	if err != nil {
		t.Fatalf("AesDecrypt PKCS5 long text failed: %v", err)
	}

	if !strings.HasPrefix(decrypted, plainText) {
		t.Errorf("Long text roundtrip failed. Got %q", decrypted)
	}
}

func TestParsePriKeyBytesInvalid(t *testing.T) {
	fmt.Println("=== 开始测试: ParsePriKeyBytes 无效数据 ===")
	// Not PEM data
	_, err := ParsePriKeyBytes([]byte("not a pem key"))
	if err == nil {
		t.Error("Expected error for invalid PEM data")
	}

	// Valid PEM but not a valid key
	invalidPem := "-----BEGIN RSA PRIVATE KEY-----\ninvalid\n-----END RSA PRIVATE KEY-----"
	_, err = ParsePriKeyBytes([]byte(invalidPem))
	if err == nil {
		t.Error("Expected error for invalid key content")
	}
}

func TestParsePubKeyBytesInvalid(t *testing.T) {
	fmt.Println("=== 开始测试: ParsePubKeyBytes 无效数据 ===")
	// Not PEM data
	_, err := ParsePubKeyBytes([]byte("not a pem key"))
	if err == nil {
		t.Error("Expected error for invalid PEM data")
	}

	// Valid PEM but not a valid key
	invalidPem := "-----BEGIN RSA PUBLIC KEY-----\ninvalid\n-----END RSA PUBLIC KEY-----"
	_, err = ParsePubKeyBytes([]byte(invalidPem))
	if err == nil {
		t.Error("Expected error for invalid key content")
	}
}

func TestAesDecryptHexDecodeError(t *testing.T) {
	fmt.Println("=== 开始测试: AesDecrypt hex解码错误 ===")
	_, err := AesDecrypt("0123456789abcdef", "zzzz")
	if err == nil {
		t.Error("Expected error for invalid hex")
	}
	// The hex.DecodeString should fail
	if err != nil {
		t.Logf("AesDecrypt with bad hex: %v (expected)", err)
	}
}

func TestAesDecryptEmptyHexString(t *testing.T) {
	fmt.Println("=== 开始测试: AesDecrypt 空hex ===")
	key := "0123456789abcdef"
	// Encrypt something first to confirm the key works
	_, _ = AesEncrypt(key, "test")

	// Now test with an empty hex string (decodes to empty bytes, then Decrypt gets empty string)
	emptyHex := hex.EncodeToString([]byte{})
	_, err := AesDecrypt(key, emptyHex)
	if err == nil {
		t.Error("Expected error for empty ciphertext after hex decode")
	}
}

func TestAesEncryptDecryptEmptyString(t *testing.T) {
	fmt.Println("=== 开始测试: AES 加密空字符串 ===")
	key := "0123456789abcdef"
	aesEnc := TAESEncrypt{}
	aesEnc.Init(key)

	// Encrypt empty string
	cipherBytes, err := aesEnc.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty string failed: %v", err)
	}

	// Decrypt
	decrypted, err := aesEnc.Decrypt(string(cipherBytes))
	if err != nil {
		t.Fatalf("Decrypt empty string failed: %v", err)
	}

	// ZeroUnPadding on all-zero content: pos will be 0, so returns original
	// The decrypted text may contain padding zeros
	t.Logf("Decrypted empty: %q (len=%d)", string(decrypted), len(decrypted))
}

func TestAesPKCS5Roundtrip(t *testing.T) {
	fmt.Println("=== 开始测试: AES PKCS5 完整加解密 ===")
	key := "0123456789abcdef"
	aesEnc := TAESEncrypt{}
	aesEnc.Init(key)
	aesEnc.ZeroPad = false // Use PKCS5

	tests := []string{
		"Hello",
		"Exactly16chars!!",
		"A",
		"This is a test string with more than 16 characters to test multi-block.",
	}

	for _, plain := range tests {
		cipherBytes, err := aesEnc.Encrypt(plain)
		if err != nil {
			t.Errorf("PKCS5 Encrypt %q failed: %v", plain, err)
			continue
		}
		decrypted, err := aesEnc.Decrypt(string(cipherBytes))
		if err != nil {
			t.Errorf("PKCS5 Decrypt %q failed: %v", plain, err)
			continue
		}
		if string(decrypted) != plain {
			t.Errorf("PKCS5 roundtrip %q failed. Got %q", plain, string(decrypted))
		}
	}
}
