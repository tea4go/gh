package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
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
	//fmt.Printf("Base64加密结果： %s\n", base64.StdEncoding.EncodeToString(crypted))

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

// 生成密钥对文件，返回私钥和公钥字符串
func GetPKCS1Key() (string, string, error) {
	var prikey, pubkey string

	// 生成 RSA 私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("密钥生成失败，%v", err)
	}

	// 编码私钥为 PKCS#1 DER 格式
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateDER,
	}
	prikey = string(pem.EncodeToMemory(privateBlock))

	publicKey := &privateKey.PublicKey
	// 编码公钥为 PKCS#1 DER 格式
	publicDER := x509.MarshalPKCS1PublicKey(publicKey)
	publicBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicDER,
	}
	pubkey = string(pem.EncodeToMemory(publicBlock))

	return prikey, pubkey, nil
}

// 将PEM格式的私钥转换为可用于验证JWT签名的RSA公钥对象。
func ParsePriKeyBytes(pri_key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pri_key)

	if block == nil {
		return nil, errors.New("解析私钥失败")
	}

	// 解析PKCS1格式的私钥
	pri_ret, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析PKCS1私钥出错，%v", err)
	}

	return pri_ret, nil
}

// 将PEM格式的公钥转换为可用于验证JWT签名的RSA公钥对象。
func ParsePubKeyBytes(pub_key []byte) (*rsa.PublicKey, error) {
	// 解码PEM格式的公钥数据
	block, _ := pem.Decode(pub_key)
	if block == nil {
		return nil, errors.New("解析公钥失败")
	}

	// 解析PKCS1格式的公钥
	pub_ret, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析PKCS1公钥出错，%v", err)
	}

	return pub_ret, nil
}
