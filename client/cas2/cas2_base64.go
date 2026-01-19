package cas2

import (
	"encoding/base64"
	"encoding/hex"
)

// hexToBase64 16进制转Base64
func hexToBase64(hexStr string) string {
	bytes, _ := hex.DecodeString(hexStr)
	return base64.StdEncoding.EncodeToString(bytes)
}
