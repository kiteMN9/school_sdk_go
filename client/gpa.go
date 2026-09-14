package client

import (
	"fmt"
	"log"
	"school_sdk/client/GPA"
	baseCfg "school_sdk/config"
	"time"
)

func (a *APIClient) getGPA() {
	resp, err := a.Http.R().
		SetRetryCount(0).
		//SetTimeout(time.Second*19).
		SetQueryParam("gnmkdm", "N105515").
		SetResponseDoNotParse(true).
		Get(baseCfg.AcademiaIndex)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsStatusFailure() {
		fmt.Println(resp.Status())
		fmt.Println(resp.Duration())
	}
	if resp.IsStatusSuccess() {
		GPA.GPA(resp.Body)
		//GPA.GPA(resp.String())
	}
	if resp.Body != nil {
		_ = resp.Body.Close()
	}
	//log.Println(utils.RemoveEmptyLines(resp.String()))
}

func (a *APIClient) getGPA_Params() {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetTimeout(time.Second * 19).
		SetQueryParams(map[string]string{
			"gnmkdm": "N309130",
			"_":      fmt.Sprint(time.Now().UnixMilli()),
			"su":     a.Config.Account,
		}).
		Get(baseCfg.GpaParamsUrl)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsStatusFailure() {
		fmt.Println(resp.Status())
		fmt.Println(resp.Duration())
	}
	if resp.IsStatusSuccess() {
		log.Println(resp.String())
	}
	//log.Println(utils.RemoveEmptyLines(resp.String()))
}

func (a *APIClient) getGPA_CACL() {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetTimeout(time.Second*19).
		SetQueryParam("gnmkdm", "N309131").
		SetFormData(map[string]string{}).
		Post(baseCfg.GpaCalcUrl)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsStatusFailure() {
		fmt.Println(resp.Status())
		fmt.Println(resp.Duration())
	}
	if resp.IsStatusSuccess() {
		log.Println(resp.String())
	}
	//log.Println(utils.RemoveEmptyLines(resp.String()))
}
func (a *APIClient) getGPA_Query() {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetTimeout(time.Second * 19).
		SetQueryParams(map[string]string{"gnmkdm": "N309131", "doType": "query", "su": a.Config.Account}).
		SetFormData(map[string]string{
			"_search":                "false",
			"nd":                     fmt.Sprint(time.Now().UnixMilli()),
			"queryModel.showCount":   "15",
			"queryModel.currentPage": "1",
			"queryModel.sortName":    "",
			"queryModel.sortOrder":   "asc",
			"time":                   "0",
		}).
		Post(baseCfg.GpaQueryUrl)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsStatusFailure() {
		fmt.Println(resp.Status())
		fmt.Println(resp.Duration())
	}
	if resp.IsStatusSuccess() {
		log.Println(resp.String())
	}
	//log.Println(utils.RemoveEmptyLines(resp.String()))
}
