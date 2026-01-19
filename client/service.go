package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strings"
	"time"

	"resty.dev/v3"
)

func (a *APIClient) CheckSession(ctx context.Context) bool {
	resp, err := a.Http.R().
		SetRetryCount(0).
		//SetQueryParams(map[string]string{
		//"gnmkdm": "N100801",
		//"layout": "default",
		//"su": a.Account,
		//}).
		SetContext(ctx).
		SetTimeout(11 * time.Second).
		//Get(baseCfg.InfoJson)
		Get("/xtgl/index_cxGxDlztxx.html?dlztxxtj_id=")

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			log.Println("保持登录已取消")
			return true
		} else {
			fmt.Println(err)
		}
	}

	if utils.UserIsLogin(a.Account, resp.String()) && !a.CheckLogout302(resp) {
		// Ctrl里有关重定向是302，不关是200
		return true
	} else {
		fmt.Println(resp.Status())
		return a.ReLogin()
	}
}

func (a *APIClient) CheckSession2(ctx context.Context) bool {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetQueryParams(map[string]string{
			"xt":        "jw",
			"localeKey": "zh_CN",
			"_":         fmt.Sprint(time.Now().UnixMilli()),
			"gnmkdm":    "index",
		}).
		SetContext(ctx).
		SetTimeout(10 * time.Second).
		Get(baseCfg.StudentName) // /xtgl/index_cxGxDlztxx.html?dlztxxtj_id=

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			log.Println("保持登录已取消")
			return true
		} else {
			fmt.Println(err)
		}
	}

	if utils.UserIsLogin(a.Account, resp.String()) && !a.CheckLogout302(resp) {
		// Ctrl里有关重定向是302，不关是200
		return true
	} else {
		fmt.Println(resp.Status())
		return a.ReLogin()
	}
}

func (a *APIClient) LoginCheck(resp *resty.Response) bool {
	if resp == nil {
		return true
	}
	if strings.Contains(resp.String(), "Sorry, the page you are looking for is currently unavailable.") || resp.StatusCode() >= 400 {
		//if strings.Contains(resp.String(), "Sorry, the page you are looking for is currently unavailable.") {
		fmt.Println("http状态码:", resp.Status())
		fmt.Println(resp.String())
		fmt.Print("程序已暂停，Enter以继续 ")
		_, err := fmt.Scanln()
		if err != nil {
			return false
		}
		return true
	}
	return utils.UserIsLogin(a.Account, resp.String()) && !a.CheckLogout302(resp)
}
