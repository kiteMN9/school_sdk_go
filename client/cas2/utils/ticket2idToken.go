package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type JWTContent struct {
	IdentityTypeCode string `json:"identityTypeCode"`
	Aud              string `json:"aud"`
	Sub              string `json:"sub"` // Account
	OrganizationCode string `json:"organizationCode"`
	Iss              string `json:"iss"`
	IdToken          string `json:"idToken"`
	Exp              int64  `json:"exp"`
	Iat              int64  `json:"iat"`
	Jti              string `json:"jti"`
}

// ExtractIDToken idToken, expTime, Account
func ExtractIDToken(tokenString string) (string, time.Time, string, error) {
	// 分割 JWT 的三个部分
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", time.Time{}, "", fmt.Errorf("无效的token格式，期望3部分，实际得到%d部分", len(parts))
	}
	// 解码 payload 部分
	payload := parts[1]
	// 添加 padding 如果必要（base64 解码需要正确的 padding）
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	// Base64解码
	decodedBytes, err := base64.RawStdEncoding.DecodeString(payload)
	if err != nil {
		// 如果 RawURLEncoding 失败，尝试标准 URLEncoding
		decodedBytes, err = base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return "", time.Time{}, "", fmt.Errorf("Base64解码失败: %w", err)
		}
	}

	// 解析JSON
	var claims JWTContent
	if unMarshaErr := json.Unmarshal(decodedBytes, &claims); unMarshaErr != nil {
		return "", time.Time{}, "", fmt.Errorf("JSON解析失败: %w", unMarshaErr)
	}

	var expTime time.Time
	// 提取 exp 字段
	if claims.Exp != 0 {
		expTime = time.Unix(claims.Exp, 0) // nextLoginTimeExp
		// 创建东八区时区
		east8Zone := time.FixedZone("CST", 8*60*60)
		log.Printf("Token expires at: %s\n", expTime.In(east8Zone).Format(time.RFC3339))
		return claims.IdToken, expTime, claims.Sub, nil
	}

	return claims.IdToken, expTime, claims.Sub, nil
}
