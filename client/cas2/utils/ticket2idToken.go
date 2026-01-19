package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type ticketR struct {
	IdentityTypeCode string `json:"identityTypeCode"`
	Aud              string `json:"aud"`
	Sub              string `json:"sub"`
	OrganizationCode string `json:"organizationCode"`
	Iss              string `json:"iss"`
	IdToken          string `json:"idToken"`
	Exp              int    `json:"exp"`
	Iat              int    `json:"iat"`
	Jti              string `json:"jti"`
}

func ExtractIDToken(tokenString string) (string, error) {
	// 分割字符串
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("无效的token格式，期望3部分，实际得到%d部分", len(parts))
	}

	// Base64解码
	decodedBytes, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		decodedBytes, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("Base64解码失败: %w", err)
		}
	}

	// 解析JSON
	var payload ticketR
	if err2 := json.Unmarshal(decodedBytes, &payload); err2 != nil {
		return "", fmt.Errorf("JSON解析失败: %w", err2)
	}

	return payload.IdToken, nil

	//// 解析JSON
	//var payload map[string]interface{}
	//if err := json.Unmarshal(decodedBytes, &payload); err != nil {
	//	return "", fmt.Errorf("JSON解析失败: %w", err)
	//}
	//
	//// 提取idToken
	//idToken, exists := payload["idToken"]
	//if !exists {
	//	return "", fmt.Errorf("未找到idToken字段")
	//}
	//
	//idTokenStr, ok := idToken.(string)
	//if !ok {
	//	return "", fmt.Errorf("idToken不是字符串类型")
	//}
	//
	//return idTokenStr, nil
}
