package config

import "strings"

const EdgeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0"

const FireFoxUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0"

func CheckUALegal(UA string) bool {
	// 合法是true 非法是 false
	if UA == "" {
		return false
	}
	if strings.Contains(UA, "Mozilla/") || strings.Contains(UA, "/") {
		return true
	}
	return false
}
