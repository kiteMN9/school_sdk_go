package utils

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"time"
)

func ExtractExpManual(tokenString string) (time.Time, string) {
	// 分割 JWT 的三个部分
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		log.Fatal("Invalid JWT format")
	}

	// 解码 payload 部分
	payload := parts[1]

	// 添加 padding 如果必要（base64 解码需要正确的 padding）
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		// 如果 RawURLEncoding 失败，尝试标准 URLEncoding
		decoded, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			log.Fatalf("Failed to decode payload: %v", err)
		}
	}

	// 解析 JSON claims
	var claims JWTContent
	if unMarshaErr := json.Unmarshal(decoded, &claims); unMarshaErr != nil {
		log.Fatalf("Failed to unmarshal claims: %v", unMarshaErr)
	}

	var expTime time.Time
	// 提取 exp 字段
	if claims.Exp != 0 {
		expTime = time.Unix(int64(claims.Exp), 0)
		// 创建东八区时区
		east8Zone := time.FixedZone("CST", 8*60*60)
		log.Printf("Token expires at: %s\n", expTime.In(east8Zone).Format(time.RFC3339))
		return expTime, claims.Sub
	}
	return time.Time{}, claims.Sub
}

type JWTContent struct {
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
