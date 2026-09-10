package utils

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

const SMTPConfigFileName = "SMTP.json"

type SMTPConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Password string   `json:"password"`
	Enable   bool     `json:"enable"`
	d        *gomail.Dialer
}

func (c *SMTPConfig) WriteSMTPConfig() {
	dataByte, err := json.Marshal(c, jsontext.WithIndent("  ")) // 无前缀，两个空格缩进
	if err != nil {
		panic(fmt.Sprintf("JSON序列化失败: %v", err))
	}
	err1 := os.WriteFile(SMTPConfigFileName, dataByte, 0644)
	if err1 != nil {
		panic(err1)
	}
}

func initSMTPConfigDefault() *SMTPConfig {
	initialData := SMTPConfig{
		Host:     "smtp.qq.com",
		Port:     587,
		From:     "abcdefg@qq.com",
		To:       []string{"123456@qq.com", "456789@qq.com"},
		Password: "smtp password",
	}
	initialData.WriteSMTPConfig()
	return &initialData
}

func SMTPReadConfig() *SMTPConfig {
	filename := SMTPConfigFileName
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return initSMTPConfigDefault()
	}
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var config SMTPConfig
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("json配置解析失败")
		log.Println("json配置解析失败")
		config.WriteSMTPConfig()
		return &config
	}
	return &config
}

func (c *SMTPConfig) SendMail(subject, content string) {
	m := gomail.NewMessage()
	//m.SetHeader("From", "sender@example.com")
	m.SetHeader("From", c.From)
	//m.SetHeader("To", "recipient@example.com")
	m.SetHeader("Bcc", c.To...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	if c.d == nil {
		c.d = gomail.NewDialer(c.Host, c.Port, c.From, c.Password)
	}

	if err := c.d.DialAndSend(m); err != nil {
		log.Println(err)
	}
}
