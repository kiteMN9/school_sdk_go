package rsa

import (
	"fmt"
	"testing"
)

func Test_RsaEncrypt(*testing.T) {
	// Base64编码的公钥模数和指数
	modulusBase64 := "AKaDZR0GA7V7IokWdV+r7J3QgV8ovozuHgFrShFVbQQJYf1+FjIlcxgHP1BFOv4efIgJ0yOm7e1rX0MqaA2974rbAojpXOWKdbt7OGP53wOhryeOkd9GYB6EmbUYT4bnWR94ALyYNTecE73rJrCTgd3hMI4FMxsyp2k0TuU6OYMN"
	exponentBase64 := "AQAB"
	secret := "12345678"
	result := ""
	EncryptRsa(&modulusBase64, &exponentBase64, &secret, &result)
	fmt.Println(result)
}
