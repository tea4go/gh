package utils

import (
	"encoding/base64"
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
	// Note: The Decrypt method expects a string which is essentially the raw bytes cast to string
	// based on how it is used in the testMyAES function in aes.go.
	// However, usually cipher text is binary. Let's look at Decrypt implementation:
	// func (this *TAESEncrypt) Decrypt(ciphertext string) ([]byte, error)
	// It converts string to []byte.

	decryptedBytes, err := aesEnc.Decrypt(string(cipherBytes))
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Remove padding is handled inside Decrypt (ZeroUnPadding or PKCS5UnPadding)
	// TAESEncrypt defaults to ZeroPad=true in Init.
	
	decryptedText := string(decryptedBytes)
	
	// ZeroPadding might leave null bytes if the original string contained them? 
	// Or rather, ZeroUnPadding cuts off at the first null byte.
	// "Hello..." doesn't contain null bytes.
	
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

	// AesDecrypt implementation:
	// pass, _ := aes.Decrypt(string(temp_text))
	// return string(pass), nil
	// Note: AesEncrypt uses ZeroPadding by default (via Init).
	// ZeroUnPadding stops at first null byte.
	
	// If plainText is short, it works.
	if !strings.HasPrefix(decryptedText, plainText) {
         // ZeroUnPadding implementation:
         // pos := strings.IndexByte(string(ciphertext), 0)
         // if pos > 0 { return ciphertext[:pos] }
         // So it should be exact match if there are no internal nulls.
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
	// Checking the manual usage in testMyAES
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
