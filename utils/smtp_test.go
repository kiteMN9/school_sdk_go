package utils

import (
	"testing"
)

func Test_sendMail(t *testing.T) {
	cfg := SMTPReadConfig()

	content := "<b>HTML 内容</b>"
	SendMail(cfg, "Hello!", content)
}
