package utils

import (
	"fmt"
	"testing"
)

func Test_extractIDToken(*testing.T) {
	tokenString := "eyJhbGciOiJIUzUxMiJ9.哎呀这部分涉及隐私啦"

	idToken, err := ExtractIDToken(tokenString)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("成功提取idToken: %s\n", idToken)
}
