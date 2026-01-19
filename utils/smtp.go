package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

const SMTPConfigFileName = "SMTP.json"

func SendMail(cfg SMTPConfig, subject, content string) {
	m := gomail.NewMessage()
	//m.SetHeader("From", "sender@example.com")
	m.SetHeader("From", cfg.From)
	//m.SetHeader("To", "recipient@example.com")
	m.SetHeader("Bcc", cfg.To...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.From, cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
	}
}

func SetSMTPConfigDefault() {
	initialData := SMTPConfig{
		Host:     "smtp.qq.com",
		Port:     587,
		From:     "abcdefg@qq.com",
		To:       []string{"123456@qq.com", "456789@qq.com"},
		Password: "qq smtp password",
	}
	SetSMTPConfig(initialData)
}

func SetSMTPConfig(configData SMTPConfig) {
	dataByte, err := json.MarshalIndent(configData, "", "  ") // 无前缀，两个空格缩进
	if err != nil {
		panic(fmt.Sprintf("JSON序列化失败: %v", err))
	}
	err1 := os.WriteFile(SMTPConfigFileName, dataByte, 0644)
	if err1 != nil {
		panic(err1)
	}
}

func SMTPReadConfig() SMTPConfig {
	filename := SMTPConfigFileName
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		SetSMTPConfigDefault()
	}
	byteValue, err := os.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	var config SMTPConfig
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("json配置解析失败")
		log.Fatalln("json配置解析失败")
		return SMTPConfig{}
	}
	return config
}

type SMTPConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Password string   `json:"password"`
	Enable   bool     `json:"enable"`
}
