package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	//_ "net/http/pprof"
	"school_sdk/client"
	"school_sdk/utils"
)

var version = "school_sdk_go 1.2.8"

func main() {
	var cfgFileName, modeCode string

	// 版本信息
	verFlag := flag.Bool("V", false, "Print version information")
	// 绑定命令行参数到变量
	flag.StringVar(&cfgFileName, "c", "config.json", "Specify config file path")
	//// 也可以添加长格式别名 (可选)
	//flag.StringVar(&cfgFileName, "config", "config.json", "Specify config file path (long format)")
	perInfo := flag.Bool("d", false, "不查个人信息")

	flag.StringVar(&modeCode, "code", "", "模式代码")

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
	fCfg := client.ReadConfig(cfgFileName)
	apiClient := client.NewAPIClient(16*time.Second, fCfg, cfgFileName, *cas2 || fCfg.CasLogin, *cas2wx)
	//apiClient := client.NewClientWithCookieJar(cliConfig, fCfg.Account, jar)
	fmt.Println("当前用户:", fCfg.Account)

	if apiClient.Login() {
		diffTime := time.Since(startTime)
		log.Println("登录总用时:", diffTime)
	} else {
		fmt.Println("登录失败")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
	if !(*perInfo) {
		client.PrintStudentInfo2(apiClient.GetJsonInfo())
	}

	//apiClient.GetScore("2024", 2)
	close(done)
	//apiClient.GetScore(client.GetSuggestYearTerm())
	apiClient.GetCourseCtl(modeCode)
}
