package config

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	cfg "school_sdk/config"
	"school_sdk/utils"

	"github.com/AlecAivazis/survey/v2/terminal"
)

type Data struct {
	filename  string
	BaseURL   string `json:"url"`
	Account   string `json:"account"`
	Passwd    string `json:"password"`
	CasPasswd string `json:"casPasswd"`
	Timeout   string `json:"timeout"`
	Want      string `json:"want"`
	//Verify    string `json:"verify"`
	ExistVerify  bool     `json:"verify"`
	CasLogin     bool     `json:"casLogin"`
	UserAgent    string   `json:"ua"`
	PerInfo      bool     `json:"perInfo"`
	Hedging      bool     `json:"hedging"`
	HedgingDelay string   `json:"hedgingDelay"`
	TicketJWT    string   `json:"ticketJWT"`
	Routes       []string `json:"routes"`
	enableCas2   bool
}

func (c *Data) WriteConfig() {
	dataByte, err := json.Marshal(c, jsontext.WithIndent("  "))
	if err != nil {
		panic(fmt.Sprintf("JSON序列化失败: %v", err))
	}
	err1 := os.WriteFile(c.filename, dataByte, 0644)
	if err1 != nil {
		panic(err1)
	}
}

func initConfig(filename string) *Data {
	initialData := Data{
		filename:     filename,
		BaseURL:      "https://jwglxt.ycit.edu.cn/",
		Account:      "account",
		Passwd:       "password",
		CasPasswd:    "cas2password",
		Timeout:      "43s",
		Want:         "want.xlsx",
		UserAgent:    cfg.FireFoxUA,
		ExistVerify:  true,
		CasLogin:     false,
		PerInfo:      true,
		Hedging:      false,
		HedgingDelay: "21s",
		Routes:       []string{""},
	}
	initialData.WriteConfig()
	initialData.SetConfigUserInfo()
	return &initialData
}

func ReadConfig(filename string) *Data {
	if info, err := os.Stat(filename); os.IsNotExist(err) || info.Size() < 2 {
		return initConfig(filename)
	}
	// 读取文件内容
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	config := Data{
		filename:     filename,
		Timeout:      "43s",
		UserAgent:    cfg.ApppleUA,
		ExistVerify:  true,
		PerInfo:      true,
		HedgingDelay: "21s",
	}
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("json配置解析失败", err)
		log.Fatalln("json配置解析失败", err)
		return nil
	}
	return &config
}

func (c *Data) SetConfigUserInfo() {
	var Account, Passwd, newAccount, newPasswd string
	var err error
	if c.enableCas2 {
		fmt.Println("当前登录方式：门户登录")
	} else {
		fmt.Println("当前登录方式：教务系统登录，平静的配色方案、干净式美学...")
	}
	newAccount = c.Account
	if c.enableCas2 {
		newPasswd = c.CasPasswd
	} else {
		newPasswd = c.Passwd
	}

	fmt.Println("当前用户:", newAccount)
	for {
		Account, err = utils.UserInputWithSigInt("  账号:")
		if err == io.EOF || errors.Is(err, terminal.InterruptErr) {
			os.Exit(0)
		}
		if Account == "" && newAccount != "account" {
			Account = newAccount
			fmt.Printf("账号保持(%s)不变\n", Account)
		} else if Account == "account" {
			fmt.Println("你是认真的吗？")
			continue
		} else {
			fmt.Println("设置用户:", Account)
		}
		break
	}

	fmt.Printf("当前密码:(%s)\n", newPasswd)

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
	if c.enableCas2 {
		c.CasPasswd = Passwd
	} else {
		c.Passwd = Passwd
	}
	c.WriteConfig()
}

func (c *Data) UpdateConfigUserInfo(verify bool) {
	c.ExistVerify = verify
	c.SetConfigUserInfo()
}

func (c *Data) SetCas2(b bool) {
	c.enableCas2 = b
}
