package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"net/http"
	//_ "net/http/pprof"
	"school_sdk/client"
	"school_sdk/utils"
)

func main() {
	var cfgFileName, modeCode string
	var version = "school_sdk_go 1.2.7"
	// 版本信息
	verFlag := flag.Bool("V", false, "Print version information")
	// 绑定命令行参数到变量
	flag.StringVar(&cfgFileName, "c", "config.json", "Specify config file path")
	//// 也可以添加长格式别名 (可选)
	//flag.StringVar(&cfgFileName, "config", "config.json", "Specify config file path (long format)")
	perInfo := flag.Bool("d", false, "不查个人信息")

	flag.StringVar(&modeCode, "code", "", "模式代码")

	campus := flag.Bool("cam", false, "校园网模式")

	cas2 := flag.Bool("cas", false, "启用cas2登录方式")
	cas2wx := flag.Bool("wx", false, "启用cas2微信登录")
	// 解析命令行参数
	flag.Parse()
	if *verFlag {
		fmt.Printf("%s\n", version)
		os.Exit(0)
	}
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()
	utils.Exit()
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-done:
			signal.Stop(sigCh)
		case <-sigCh:
			signal.Stop(sigCh)
			fmt.Println()
			os.Exit(0)
		}
	}()

	log.Println("程序启动") // 写入文件和控制台
	startTime := time.Now()

	fileConfig := utils.ReadConfig(cfgFileName)

	if *campus && !*cas2wx {
		fmt.Println("当前用户:", utils.MaskString(fileConfig.Account, 2, 7))
	} else if !*cas2wx {
		fmt.Println("当前用户:", fileConfig.Account)
	}

	url := ""
	if *campus {
		list := []string{
			//"http://10.0.4.8",
			"http://10.0.4.9",
			"http://10.0.4.22",
			//"http://10.0.4.23",
			//"http://202.119.141.5:81",
		}
		randomIndex := rand.Intn(len(list))
		selected := list[randomIndex]
		url = selected
	} else {
		url = fileConfig.URL
	}

	cliConfig := client.NewConfig(url, fileConfig.Verify, 15*time.Second, fileConfig.UserAgent)
	apiClient := client.NewAPIClient(cliConfig, fileConfig.Account, fileConfig.Passwd, cfgFileName, *cas2, *cas2wx, fileConfig.CasPasswd)

	if apiClient.Login() {
		diffTime := time.Since(startTime)
		log.Println("登录总用时:", diffTime)
	} else {
		log.Printf("登录失败\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	}
	if !(*campus || *perInfo) {
		client.PrintStudentInfo2(apiClient.GetJsonInfo())
	}

	//apiClient.GetScore("2024", 2)
	close(done)

	apiClient.GetCourseCtl(modeCode)
}
