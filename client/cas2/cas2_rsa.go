package cas2

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
)

// RSAEncryptor 封装 RSA 加密功能
type RSAEncryptor struct {
	publicKey *rsa.PublicKey
}

// NewRSAEncryptorFromPEM 从 PEM 格式公钥创建加密器
func NewRSAEncryptorFromPEM(publicKeyPEM string) (*RSAEncryptor, error) {
	// 解析 PEM 块
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	// 根据 PEM 类型处理
	var pub interface{}
	var err error

	switch block.Type {
	case "PUBLIC KEY":
		// PKIX 格式公钥
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		// PKCS#1 格式公钥
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// 类型断言为 RSA 公钥
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return &RSAEncryptor{publicKey: rsaPub}, nil
}

// NewRSAEncryptorFromComponents 从模数和指数创建加密器
func NewRSAEncryptorFromComponents(n, e string) (*RSAEncryptor, error) {
	// 解析模数 n
	nBytes, err := hex.DecodeString(n)
	if err != nil {
		return nil, fmt.Errorf("解析模数 n 失败: %v", err)
	}

	// 解析公钥指数 e
	eBytes, err := hex.DecodeString(e)
	if err != nil {
		return nil, fmt.Errorf("解析指数 e 失败: %v", err)
	}

	// 创建公钥
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}

	return &RSAEncryptor{publicKey: publicKey}, nil
}

// Encrypt 加密得到 16 进制结果
// Encrypt 使用标准库进行 RSA 加密（PKCS#1 v1.5）
func (r *RSAEncryptor) Encrypt(password string) (string, error) {
	// 使用标准库的 RSA 加密
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, []byte(password))
	if err != nil {
		return "", fmt.Errorf("RSA 加密失败: %v", err)
	}

	// 转换为 16 进制
	hexResult := hex.EncodeToString(encrypted)

	// 确保长度为偶数（与原 JavaScript 逻辑一致）
	if len(hexResult)%2 != 0 {
		hexResult = "0" + hexResult
	}

	return hexResult, nil
}

// EncryptWithBase64 加密并转换为 Base64
func (r *RSAEncryptor) EncryptWithBase64(password string) (string, error) {
	hexResult, err := r.Encrypt(password)
	if err != nil {
		return "", err
	}

	// 转换为 Base64
	base64Result := hexToBase64(hexResult)
	return base64Result, nil
}
