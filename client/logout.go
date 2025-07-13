package client

import (
	"context"
	"errors"
	"fmt"
	baseCfg "school_sdk/config"
	"school_sdk/utils"

	"time"
)

func (a *APIClient) Logout() string {
	RetryCount := a.http.RetryCount
	// 超时强制退出机制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	a.http.SetRetryCount(0)
	resp, err := a.http.R().SetContext(ctx).
		SetQueryParams(map[string]string{
			"t":          fmt.Sprint(time.Now().UnixMilli()),
			"login_type": "",
		}).
		Get(baseCfg.LOGOUT)
	a.http.SetRetryCount(RetryCount)

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ""
		} else {
			fmt.Println(err)
		}
	}

	bodyString := resp.String()
	if utils.UserIsLogin(a.Account, bodyString) && !a.CheckLogout302(resp) {
		fmt.Println("退出失败")
	} else {
		fmt.Println("退出成功")
	}

	return bodyString

}
