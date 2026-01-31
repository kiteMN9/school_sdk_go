package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	cfg "school_sdk/config"

	"github.com/AlecAivazis/survey/v2/terminal"
)

type ConfigData struct {
	URL       string `json:"url"`
	Account   string `json:"account"`
	Passwd    string `json:"password"`
	CasPasswd string `json:"casPasswd"`
	UserAgent string `json:"ua"`
	//Verify    string `json:"verify"`
	Verify   bool `json:"verify" default:"true"`
	CasLogin bool `json:"casLogin" default:"false"`
}

func SetConfigDefault(filename string) {
	initialData := ConfigData{
		URL:       "https://jwglxt.ycit.edu.cn/",
		Account:   "account",
		Passwd:    "password",
		CasPasswd: "cas2password",
		UserAgent: cfg.FireFoxUA,
		Verify:    true,
		CasLogin:  false,
	}
	//initialData := ConfigData{
	//	URL:       "http://www.gdjw.zjut.edu.cn/jwglxt",
	//	Account:   "Zjuter",
	//	Passwd:    "Zjut12@@",
	//	UserAgent: cfg.FireFoxUA,
	//}
	SetConfig(filename, initialData)
}

func SetConfig(filename string, configData ConfigData) {
	// filename := "config.json"

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
		SetConfigDefault(filename)
	}
	byteValue, err := os.ReadFile(filename)
	// file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	// defer file.Close()

	// 读取文件内容
	// byteValue, _ := ioutil.ReadAll(file)

	// 将 JSON 数据解析到结构体
	var config ConfigData
	config.Verify = true
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("json配置解析失败", err)
		log.Fatalln("json配置解析失败", err)
		return nil
	}
	//fmt.Println("<UNK>:", config)
	if config.Account == "account" {
		config_ := SetConfigUserInfo(filename, &config)
		config.Account = config_.Account
		config.Passwd = config_.Passwd
	}
	//fmt.Println("<UNK>:", config)
	return &config
}
func SetConfigUserInfo(filename string, config *ConfigData) *ConfigData {
	var configNew ConfigData
	configNew.Verify = config.Verify
	fmt.Println("当前用户:", config.Account)
	for {
		//fmt.Print("    账号: ")
		var err error
		//_, err := fmt.Scanln(&configNew.Account)
		configNew.Account, err = UserInputWithSigInt("    账号:")
		if err == io.EOF || errors.Is(err, terminal.InterruptErr) {
			os.Exit(0)
		}
		if configNew.Account == "" && config.Account != "account" {
			configNew.Account = config.Account
			fmt.Printf("账号保持(%s)不变\n", configNew.Account)
		} else if configNew.Account == "account" {
			fmt.Println("你是认真的吗？")
			continue
		} else {
			fmt.Println("设置用户:", configNew.Account)
		}
		break
	}

	fmt.Printf("当前密码:(%s)\n", config.Passwd)
	//for {
	var err error
	configNew.Passwd, err = UserInputWithSigInt("    密码:")
	if err != nil {
		if err == io.EOF || errors.Is(err, terminal.InterruptErr) {
			os.Exit(0)
		}
		return nil
	}

	if configNew.Passwd == "password" {
		fmt.Println("认真的？改改密码吧")
		//break
	} else {
		fmt.Printf("设置密码:(%s)\n", configNew.Passwd)
	}
	//break
	//}
	configNew.URL = config.URL
	configNew.UserAgent = config.UserAgent

	SetConfig(filename, configNew)
	return &configNew
}
func UpdateConfigUserInfo(filename string, verify bool) *ConfigData {
	config := ReadConfig(filename)
	config.Verify = verify
	info := SetConfigUserInfo(filename, config)
	if info == nil {
		return config
	}
	return info
}
