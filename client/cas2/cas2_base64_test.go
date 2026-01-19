package cas2

import (
	"fmt"
	"testing"
)

func Test_hex2Base64(*testing.T) {
	// 测试示例
	hexStr := "9e1d4fec82450374460159deea3088f29624cbe....."
	fmt.Printf("结果:\n%s\n__RSA__%s\n", hexStr, hexToBase64(hexStr))
}

func Test_rsa(*testing.T) {
	// 示例 PEM 公钥（这里需要替换为实际的 PEM 公钥）
	publicKeyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBg........FAAOCAQlwIDAQAB
-----END PUBLIC KEY-----`

	// 从 PEM 创建加密器
	encryptor, err := NewRSAEncryptorFromPEM(publicKeyPEM)
	if err != nil {
		fmt.Printf("创建加密器失败: %v\n", err)
		return
	}

	password := "myPassword123"

	fmt.Println("=== RSA 加密测试 ===")
	fmt.Printf("原始密码: %s\n", password)

	// 加密得到 16 进制结果
	hexResult, err := encryptor.Encrypt(password)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("16进制结果: %s\n", hexResult)

	// 加密得到 Base64 结果
	base64Result, err := encryptor.EncryptWithBase64(password)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("Base64结果: %s\n", base64Result)

	// 测试自定义 Base64 编码
	//customBase64 := hexToBase64Custom(hexResult)
	//fmt.Printf("自定义Base64: %s\n", customBase64)
}
