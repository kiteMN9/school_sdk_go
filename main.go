package main

import (
	"fmt"
	"log"
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

	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	log.Println("程序启动") // 写入文件和控制台
	startTime := time.Now().UnixMilli()

	Account := "123456"
	Passwd := "123456"
	URL := "https://jwglxt.xxxx.edu.cn"
	UserAgent := ""
	fmt.Println("当前用户:", Account)

	cliConfig := client.NewConfig(URL, true, 30*time.Second, UserAgent)
	apiClient := client.NewAPIClient(cliConfig, Account, Passwd)

	if apiClient.Login() {
		diffTime := time.Now().UnixMilli() - startTime
		if diffTime > 2000 {
			log.Printf("登录总用时: %d.%ds\n", diffTime/1000, diffTime/10%100)
		} else {
			log.Printf("登录总用时: %dms\n", diffTime)
		}
	} else {
		log.Printf("登录失败\n")
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	client.PrintStudentInfo(apiClient.GetInfo())
	//apiClient.GetScore("2024", 2)
	close(done)

	apiClient.GetCourseCtl()
}
