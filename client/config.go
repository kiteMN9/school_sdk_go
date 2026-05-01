package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	cfg "school_sdk/config"
	"school_sdk/utils"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
)

type Config struct {
	baseURL     string
	timeout     time.Duration
	ExistVerify bool
	//Verify      string
	userAgent string
}

func NewConfig(baseURL string, existVerify bool, timeout time.Duration, userAgent string) *Config {
	if !cfg.CheckUALegal(userAgent) {
		userAgent = cfg.EdgeUA
	}
	return &Config{
		baseURL:     baseURL,
		timeout:     timeout,
		ExistVerify: existVerify,
		userAgent:   userAgent,
	}
}

type ConfigData struct {
	BaseURL   string `json:"url"`
	Account   string `json:"account"`
	Passwd    string `json:"password"`
	CasPasswd string `json:"casPasswd"`
	UserAgent string `json:"ua"`
	//Verify    string `json:"verify"`
	ExistVerify bool `json:"verify" default:"true"`
	CasLogin    bool `json:"casLogin" default:"false"`
}

func SetConfig(filename string, configData ConfigData) {
	dataByte, err := json.MarshalIndent(configData, "", "  ") // 无前缀，两个空格缩进
	if err != nil {
		panic(fmt.Sprintf("JSON序列化失败: %v", err))
	}
	err1 := os.WriteFile(filename, dataByte, 0644)
	if err1 != nil {
		panic(err1)
	}
}

func ReadConfig(filename string) *ConfigData {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initialData := ConfigData{
			BaseURL:     "https://jwglxt.ycit.edu.cn/",
			Account:     "account",
			Passwd:      "password",
			CasPasswd:   "cas2password",
			UserAgent:   cfg.FireFoxUA,
			ExistVerify: true,
			CasLogin:    false,
		}
		SetConfig(filename, initialData)
		return SetConfigUserInfo(filename, &initialData)
	}
	// 读取文件内容
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// 将 JSON 数据解析到结构体
	var config ConfigData
	config.ExistVerify = true
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("json配置解析失败", err)
		log.Fatalln("json配置解析失败", err)
		return nil
	}
	return &config
}
func SetConfigUserInfo(filename string, config *ConfigData) *ConfigData {
	var Account, Passwd string
	var err error
	fmt.Println("当前用户:", config.Account)
	for {

		Account, err = utils.UserInputWithSigInt("  账号:")
		if err == io.EOF || errors.Is(err, terminal.InterruptErr) {
			os.Exit(0)
		}
		if Account == "" && config.Account != "account" {
			Account = config.Account
			fmt.Printf("账号保持(%s)不变\n", Account)
		} else if Account == "account" {
			fmt.Println("你是认真的吗？")
			continue
		} else {
			fmt.Println("设置用户:", Account)
		}
		break
	}

	fmt.Printf("当前密码:(%s)\n", config.Passwd)

	Passwd, err = utils.UserInputWithSigInt("  密码:")
	if err != nil {
		if err == io.EOF || errors.Is(err, terminal.InterruptErr) {
			os.Exit(0)
		}
		return nil
	}

	if Passwd == "password" {
		fmt.Println("认真的？改改密码吧")
		//break
	} else {
		fmt.Printf("设置密码:(%s)\n", Passwd)
	}

	config.Account = Account
	config.Passwd = Passwd

	SetConfig(filename, *config)
	return config
}

func UpdateConfigUserInfo(filename string, verify bool) *ConfigData {
	config := ReadConfig(filename)
	config.ExistVerify = verify
	info := SetConfigUserInfo(filename, config)
	if info == nil {
		return config
	}
	return info
}
