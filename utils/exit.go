package utils

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

func Exit() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var (
		mu        sync.Mutex
		count     int
		lastPress time.Time
	)

	go func() {
		for range c {
			mu.Lock()

			now := time.Now()
			if now.Sub(lastPress) > 900*time.Millisecond {
				count = 1
			} else {
				count++
			}
			lastPress = now

			currentCount := count
			mu.Unlock()

			if currentCount >= 6 {
				fmt.Println("Force exiting...")
				fmt.Println("连续6次 Ctrl+C 且每两次之间小于0.9秒将触发强制退出机制")
				time.Sleep(1 * time.Second)
				os.Exit(1)
			}

		}
	}()

}
