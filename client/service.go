package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strings"
	"time"
)

func (a *APIClient) CheckSession(ctx context.Context) bool {
	resp, err := a.http.R().
		SetQueryParams(map[string]string{
			"gnmkdm": "N100801",
			//"layout": "default",
			"su": a.Account,
		}).
		SetContext(ctx).
		Get(baseCfg.INFO)

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			log.Println("保持登录已取消")
			return true
		} else {
			fmt.Println(err)
		}
	}

	if utils.UserIsLogin(a.Account, resp.String()) && !a.CheckLogout302(resp) {

		return true
	} else {
		fmt.Println(resp.StatusCode())
		return a.ReLogin()
	}
}

func (a *APIClient) LoginCheck(resp *resty.Response) bool {
	if strings.Contains(resp.String(), "Sorry, the page you are looking for is currently unavailable.") || resp.StatusCode() >= 400 {
		fmt.Println("http状态码:", resp.StatusCode())
		fmt.Println(resp.String())
		fmt.Print("程序已暂停，Enter以继续")
		_, err := fmt.Scanln()
		if err != nil {
			return false
		}
		return true
	}
	return utils.UserIsLogin(a.Account, resp.String()) && !a.CheckLogout302(resp)
}

func CheckStatusCode(resp *resty.Response) bool {
	if resp.StatusCode() != 200 && resp.StatusCode() != 302 {
		if resp.StatusCode() == 502 || resp.StatusCode() == 429 {
			fmt.Println(resp.StatusCode())
			fmt.Println(resp.String())
			time.Sleep(1 * time.Second)
			return true
		}
	}
	return false
}
