package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"time"
)

//func (a *APIClient) detectT(delay time.Duration, ctx context.Context) {
//	a.detectTime(ctx)
//	fmt.Println(delay)
//	time.Sleep(delay)
//	a.detectTime(ctx)
//}

func (a *APIClient) detectKeepAliveTime() {
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
	// VPN下 Tengine 服务器大概70秒关闭连接
	timeTestList := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		60 * time.Second,
		67 * time.Second,
		69100 * time.Millisecond, //✅
		//69930 * time.Millisecond, //✅
		//69940 * time.Millisecond, //?
		//69950 * time.Millisecond, //?
		//69960 * time.Millisecond, //❌
		70 * time.Second, //❌
	}
	//timeTestList =
	a.detectTime(ctx)
	for _, t := range timeTestList {
		//a.detectT(t, ctx)
		a.detectTime(ctx)
		if ctx.Err() != nil {
			return
		}
		fmt.Println(t)
		//time.Sleep(t)
		err := SleepWithContext(ctx, t)
		if err != nil {
			return
		}
		a.detectTime(ctx)
		fmt.Println()
		if ctx.Err() != nil {
			return
		}
	}
	//fmt.Printf("%#v\n", a.Http.CookieJar()) // route 会影响时间
	fmt.Printf("%+v\n", a.Http.CookieJar()) // route 会影响时间
	// 34ff17f478ebaa7e4063c9d5a95901d0 慢 147.7819ms
	// 018f9ff65252ca4f51865070844ae0be 慢 145.1344ms
	// 8ed16c15842922decba77aa1ed63b61f 快 84.7719ms
	// 425b918000ed5b18d10afb85fbbf8ec7 快 87ms
	// c80e782f5a3340e86274809ce311b6b4 快 80.198ms
}

func (a *APIClient) detectTime(ctx context.Context) bool {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetContext(ctx).
		SetTimeout(11 * time.Second).
		//SetQueryParam("dlztxxtj_id", "").
		Get(baseCfg.LoginStatus)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("请求已取消")
			return true
		}
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("请求超时:", resp.Duration())
			return true
		}
		fmt.Println(err)
	}
	fmt.Println(resp.Duration())
	log.Println(resp.Duration())

	if utils.UserIsLogin(a.Config.Account, resp.String()) && !a.CheckLogout302(resp) {
		return true
	}
	fmt.Println("Login check:", resp.Status())
	return a.ReLogin()
}

func SleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop() // 确保函数退出时停止定时器，避免泄露

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
