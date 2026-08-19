package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"school_sdk/client/config"
	"syscall"
	"time"

	//_ "net/http/pprof"
	"school_sdk/client"
	"school_sdk/utils"
)

var version = "school_sdk_go 1.2.17-dev"

func main() {
	var cfgFileName, wantFile, modeCode, route string

	// 版本信息
	verFlag := flag.Bool("V", false, version)
	// 绑定命令行参数到变量
	flag.StringVar(&cfgFileName, "c", "config.json", "配置文的件路径")
	flag.StringVar(&wantFile, "want", "want.xlsx", "选课愿望单文件路径")
	//// 也可以添加长格式别名 (可选)
	//flag.StringVar(&cfgFileName, "config", "config.json", "Specify config file path (long format)")

	flag.StringVar(&modeCode, "code", "", "模式代码建议5")

	cas2 := flag.Bool("cas", false, "启用cas2登录方式")
	cas2wx := flag.Bool("wx", false, "启用cas2微信登录")
	flag.StringVar(&route, "route", "", "教务系统route")

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
	fCfg := config.ReadConfig(cfgFileName)
	if wantFile != "want.xlsx" || fCfg.Want == "" {
		fCfg.Want = wantFile
		fCfg.WriteConfig()
	}
	duration, err := time.ParseDuration(fCfg.Timeout)
	if err != nil {
		duration = 47 * time.Second
		fCfg.Timeout = "47s"
		fCfg.WriteConfig()
	}
	apiClient := client.NewAPIClient(duration, fCfg, *cas2 || fCfg.CasLogin, *cas2wx, route)
	//apiClient := client.NewClientWithCookieJar(cliConfig, fCfg.Account, jar)
	fmt.Println("当前用户:", fCfg.Account)

	if apiClient.Login() {
		log.Println("登录总用时:", time.Since(startTime))
	} else {
		fmt.Println("登录失败")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
	if fCfg.PerInfo {
		client.PrintStudentInfo2(apiClient.GetJsonInfo())
	}

	//apiClient.GetScore("2024", 2)
	close(done)
	//apiClient.GetScore(client.GetSuggestYearTerm())
	apiClient.GetCourseCtl(modeCode)
}
