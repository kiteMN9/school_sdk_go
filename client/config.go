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

	"github.com/LiZhiqiang0/go_deep_copy"

	"github.com/AlecAivazis/survey/v2/terminal"
)

type ConfigData struct {
	filename  string
	BaseURL   string `json:"url"`
	Account   string `json:"account"`
	Passwd    string `json:"password"`
	CasPasswd string `json:"casPasswd"`
	Timeout   string `json:"timeout"`
	Want      string `json:"want"`
	//Verify    string `json:"verify"`
	ExistVerify bool   `json:"verify" default:"true"`
	CasLogin    bool   `json:"casLogin" default:"false"`
	UserAgent   string `json:"ua"`
}

func (c *ConfigData) WriteConfig() {
	dataByte, err := json.MarshalIndent(c, "", "  ") // 无前缀，两个空格缩进
	if err != nil {
		panic(fmt.Sprintf("JSON序列化失败: %v", err))
	}
	err1 := os.WriteFile(c.filename, dataByte, 0644)
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
			Timeout:     "16s",
			Want:        "want.xlsx",
			ExistVerify: true,
			CasLogin:    false,
			filename:    filename,
		}
		initialData.WriteConfig()
		initialData.SetConfigUserInfo(nil)
		return &initialData
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
	config.filename = filename
	return &config
}

func (c *ConfigData) SetConfigUserInfo(config *ConfigData) {
	var Account, Passwd string
	var err error
	if config == nil {
		config = &ConfigData{}
		if err := go_deep_copy.DeepCopy(c, config); err != nil {
			panic(err)
		}
	}
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
	}

	if Passwd == "password" {
		fmt.Println("认真的？改改密码吧")
		//break
	} else {
		fmt.Printf("设置密码:(%s)\n", Passwd)
	}

	c.Account = Account
	c.Passwd = Passwd

	c.WriteConfig()
}

func (c *ConfigData) UpdateConfigUserInfo(verify bool) {
	c.ExistVerify = verify
	c.SetConfigUserInfo(nil)
}
