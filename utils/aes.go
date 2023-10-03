package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strings"
)

type TAESEncrypt struct {
	Key     []byte
	IV      []byte
	ZeroPad bool
}

func (this *TAESEncrypt) Init(key_str string) {
	this.Key = []byte(key_str) //H84XGAppu17nHdEu
	this.ZeroPad = true
	this.IV = []byte("zs_radius_server")
}
func (this *TAESEncrypt) Encrypt(strMesg string) ([]byte, error) {

	aesBlockEncrypter, err := aes.NewCipher(this.Key)
	if err != nil {
		return nil, err
	}

	aesEncrypter := cipher.NewCBCEncrypter(aesBlockEncrypter, this.IV)

	content := []byte(strMesg)
	if this.ZeroPad {
		content = this.ZeroPadding(content, aesBlockEncrypter.BlockSize())
	} else {
		content = this.PKCS5Padding(content, aesBlockEncrypter.BlockSize())
	}

	crypted := make([]byte, len(content))
	aesEncrypter.CryptBlocks(crypted, content)

	//fmt.Printf("16进制加密结果： %x\n", crypted)
	//fmt.Println("Base64加密结果:", base64.StdEncoding.EncodeToString(crypted))

	return crypted, nil
}

func (this *TAESEncrypt) Decrypt(ciphertext string) ([]byte, error) {

	aesBlockEncrypter, err := aes.NewCipher(this.Key) //选择加密算法
	if err != nil {
		return nil, err
	}

	aesEncrypter := cipher.NewCBCDecrypter(aesBlockEncrypter, this.IV)
	content := make([]byte, len(ciphertext))
	aesEncrypter.CryptBlocks(content, []byte(ciphertext))

	if this.ZeroPad {
		content = this.ZeroUnPadding(content, aesBlockEncrypter.BlockSize())
	} else {
		content = this.PKCS5UnPadding(content, aesBlockEncrypter.BlockSize())
	}

	return content, nil
}

func (this *TAESEncrypt) ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func (this *TAESEncrypt) ZeroUnPadding(ciphertext []byte, blockSize int) []byte {
	pos := strings.IndexByte(string(ciphertext), 0)
	if pos > 0 {
		return ciphertext[:pos]
	}
	return ciphertext
}

func (this *TAESEncrypt) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (this *TAESEncrypt) PKCS5UnPadding(ciphertext []byte, blockSize int) []byte {
	length := len(ciphertext)
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

func testMyAES() {
	aesEnc := TAESEncrypt{}
	aesEnc.Init("0123456789abcdef")
	out_src, err := aesEnc.Encrypt("0123456789012345678901234567890123456789")
	out_src, err = aesEnc.Encrypt("1")
	if err == nil {
		aesEnc.Decrypt(string(out_src))
	}
}

func AesEncrypt(key, test_str string) (string, error) {
	aes := TAESEncrypt{}
	aes.Init(key)
	pass, err := aes.Encrypt(test_str)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", pass), nil
}

func AesDecrypt(key, test_str string) (string, error) {
	temp_text, err := hex.DecodeString(test_str)
	if err != nil {
		return "", err
	}
	aes := TAESEncrypt{}
	aes.Init(key)
	pass, _ := aes.Decrypt(string(temp_text))
	return string(pass), nil
}
