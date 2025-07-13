package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
)

func EncryptRsa(modulusBase64, exponentBase64, secret, enResult *string) {
	// Base64编码的公钥模数和指数
	// modulusBase64 := "AKaDZR0GA7V7IokWdV+r7J3QgV8ovozuHgFrShFVbQQJYf1+FjIlcxgHP1BFOv4efIgJ0yOm7e1rX0MqaA2974rbAojpXOWKdbt7OGP53wOhryeOkd9GYB6EmbUYT4bnWR94ALyYNTecE73rJrCTgd3hMI4FMxsyp2k0TuU6OYMN"
	// exponentBase64 := "AQAB"
	// 待加密的明文
	plaintext := []byte(*secret)

	// 解码Base64得到模数N和指数E的字节
	modulusBytes, err := base64.StdEncoding.DecodeString(*modulusBase64)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	exponentBytes, err1 := base64.StdEncoding.DecodeString(*exponentBase64)
	if err1 != nil {
		fmt.Println(err1)
		panic(err1)
	}

	// 将字节转换为大整数
	N := new(big.Int).SetBytes(modulusBytes)
	E := new(big.Int).SetBytes(exponentBytes).Int64()

	// 构造RSA公钥
	publicKey := &rsa.PublicKey{
		N: N,
		E: int(E),
	}

	// 使用PKCS#1 v1.5填充进行加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// fmt.Println(ciphertext)

	// 将密文转换为Base64编码
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	// fmt.Println(encoded)
	*enResult = encoded
}
