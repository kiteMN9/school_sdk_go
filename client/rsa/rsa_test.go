package rsa

import (
	"fmt"
	"testing"
)

func Test_RsaEncrypt(*testing.T) {
	// Base64编码的公钥模数和指数
	modulusBase64 := "AKaDZR0GA7V7IokWdV.....pXOWKdbt7OGP53wOhryeOkd9GYB6"
	exponentBase64 := "AQAB"
	secret := "12345678"
	result, err := EncryptRsa(modulusBase64, exponentBase64, secret)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}

//func Test_RsaPEMEncrypt(*testing.T) {
//	pubKey := `-----BEGIN PUBLIC KEY-----
//MIIBIjANBgkqhkiG9w0BAQE.....JezURm7wDK4QVEoqpVrRdE....LtKblwIDAQAB
//-----END PUBLIC KEY-----`
//
//	// 5. 准备要加密的数据
//	secretMessage := []byte("这是需要加密的敏感信息")
//	encrypt, err := PEMEncrypt(pubKey, secretMessage)
//	if err != nil {
//		return
//	}
//	fmt.Printf("加密结果: %x\n", encrypt)
//	//fmt.Println(string(encrypt))
//}
