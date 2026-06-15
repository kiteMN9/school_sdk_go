package client

import (
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"time"
)

func (a *APIClient) SwitchLanguage(lang string) {
	resp, err := a.Http.R().
		SetTimeout(12 * time.Second).
		SetFormData(map[string]string{
			"language": lang,
		}).
		Post(baseCfg.Language)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsStatusFailure() {
		log.Println("language HTTP 状态码错误:", resp.Status())
	}
	if resp.ResultError() != nil {
		log.Println(resp.ResultError(), resp.String())
	}
	if a.LoginCheck(resp) {
		// Ctrl里有关掉重定向是302，不关是200
		//return true
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}
	fmt.Println(resp.String())
}
