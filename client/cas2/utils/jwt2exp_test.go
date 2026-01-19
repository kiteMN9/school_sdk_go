package utils

import (
	"fmt"
	"testing"
)

func Test_ExtractExpManual(*testing.T) {
	tokenString := `eyJhbGciOiJIUzUxMiJ9.涉及隐私`
	expTime, account := ExtractExpManual(tokenString)
	fmt.Println("\n", expTime, account)
}
