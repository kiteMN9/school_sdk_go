package utils

import "strings"

func MaskString(s string, start, length int) string {
	// 将字符串转为rune切片以支持Unicode（此处数字可省略）
	runes := []rune(s)
	total := len(runes)

	// 边界检查
	if start < 0 || start >= total || length <= 0 {
		return s
	}
	end := start + length
	if end > total {
		end = total
	}

	// 构建结果
	var result []rune
	result = append(result, runes[:start]...)
	result = append(result, []rune(strings.Repeat("*", end-start))...)
	result = append(result, runes[end:]...)
	return string(result)
}
