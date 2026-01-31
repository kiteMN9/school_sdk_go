package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"school_sdk/utils"
	"time"
)

type startTime struct {
	StartTime time.Time `json:"start_time"`
}

func readStartTimeConfig() time.Time {
	filename := "startTime.json"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return time.Unix(0, 0)
	}

	// 读取文件内容
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		// panic(err)
		log.Println("startTime.json 文件读取失败")
		return time.Unix(0, 0)
	}

	// 将 JSON 数据解析到结构体
	var timeConfig startTime
	err = json.Unmarshal(byteValue, &timeConfig)
	if err != nil {
		return time.Unix(0, 0)
	}
	log.Println("timeFromFile:", timeConfig.StartTime, timeConfig.StartTime.Format("2006-01-02_15:04:05"))
	return timeConfig.StartTime
}

func parseStartTime(timeStr string) time.Time {
	var timeObj time.Time
	var err error

	// 日期格式
	layout := "2006-01-02 15:04:05.000"
	// 解析日期字符串为time.Time对象
	loc, _ := time.LoadLocation("Local")
	//fmt.Printf("dateStr:'%s'\n", dateStr)
	timeObj, err = time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		log.Println("解析时间失败:", err)
		//fmt.Println()
		return time.Unix(0, 0)
	}
	// 获取时间戳
	//timestamp := timeObj.UnixMilli()
	// fmt.Println("时间戳:", timestamp)
	return timeObj
}

func getTargetTime() time.Time {
	now := time.Now()
	// 构建今天的12:30
	today1230 := time.Date(now.Year(), now.Month(), now.Day(), 12, 30, 0, 0, now.Location())

	// 判断当前时间是否超过今天的12:30
	if now.Before(today1230) {
		return today1230
	}
	// 显示明天的12:30
	return today1230.AddDate(0, 0, 1)
}

func setTimeKeepSession() time.Time {
	targetTime := getTargetTime()
	for {
		// 指定日期字符串
		dateStr := targetTime.Format("01-02") + " 12:30:01.500"
		//dateStr := "2025-09-04 12:30:00.000"
		input, err2 := utils.UserInputWithSigInt(fmt.Sprint("    参考:  ", dateStr, "\n输入时间: "))
		if err2 != nil {
			return time.Unix(0, 0)
		}
		timestamp := parseStartTime(targetTime.Format("2006-") + input)
		if timestamp == time.Unix(0, 0) {
			continue
		}
		log.Println("setStartTime:", input, timestamp)

		var configData startTime
		configData.StartTime = timestamp
		dataByte, err := json.Marshal(configData)
		if err != nil {
			panic(fmt.Sprintf("JSON序列化失败: %v", err))
			// continue
		}
		if err1 := os.WriteFile("startTime.json", dataByte, 0644); err1 != nil {
			panic(err1)
		}
		return timestamp
		//break
	}
}

func (a *APIClient) timeKeepSession(targetTime time.Time) {
	delay := 0 * time.Millisecond
	if time.Since(targetTime) > delay {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-sigCh:
			cancel()
			fmt.Println("请求已取消")
		}
		signal.Stop(sigCh)
		close(sigCh)
	}()
	var count int
	// var signMen = []string{"|", "/", "-", "\\"}
	var signMen = []string{"⠇", "⠏", "⠋", "⠙", "⠹", "⠼", "⠴", "⠦", "⠧"}
	var refreshCount int
	var sign = "|"
	var signNum int
	var sigTime time.Time
	signNumStat := 0
	first := true

	fmt.Println("开始时间:", targetTime.Format("2006-01-02 15:04:05.000"))
	for {
		if time.Since(targetTime) > delay {
			fmt.Print("\r=========开始========= \n")
			fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"))
			//close(done)
			return
		}
		if first {
			fmt.Println("未到指定时间，等待中...")
			first = false
		}
		time.Sleep(1 * time.Millisecond)
		count += 1
		// signNum = count / 79 % 4
		signNum = count / 77 % 9
		if signNumStat != signNum {
			sign = signMen[signNum]
			fmt.Printf("\r======%d=========  %s ", refreshCount, sign)
			signNumStat = signNum
		}

		if time.Since(sigTime) > 21*time.Second { // 每21秒检查会话
			if time.Until(targetTime) > 56*time.Second { // 最后56秒不检查会话
				fmt.Printf("\r======%d====c====  %s ", refreshCount, sign)
				// time.Sleep(1 * time.Second)
				a.CheckSession(ctx)
				// 定时刷新
				refreshCount += 1
				if errors.Is(ctx.Err(), context.Canceled) {
					//log.Println("保持登录已取消")
					return
				}
				fmt.Printf("\r======%d=========  %s ", refreshCount, sign)
			} else {
				//fmt.Println("test test")
			}
			sigTime = time.Now()
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			//log.Println("保持登录已取消")
			return
		}
	}
}
