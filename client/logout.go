package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"time"
)

func (a *APIClient) Logout() string {
	if a.onlyCookieMethod {
		os.Exit(0)
	}

	defer os.Exit(0)
	// 超时强制退出机制
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetTimeout(time.Second * 2).
		SetQueryParams(map[string]string{
			"t":          fmt.Sprint(time.Now().UnixMilli()),
			"login_type": "",
		}).
		Get(baseCfg.LOGOUT)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return ""
		} else {
			fmt.Println(err)
		}
	}

	// fmt.Println(resp.String()
	bodyString := resp.String()
	if utils.UserIsLogin(a.Account, bodyString) && !a.CheckLogout302(resp) {
		fmt.Println("退出失败")
	} else {
		fmt.Println("退出成功")
	}

	//err1 := a.Http.Close()
	//if err1 != nil {
	//	return ""
	//}

	return bodyString

}
