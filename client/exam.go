package client

import (
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"strconv"
	"time"
)

func (a *APIClient) GetExam(year string, term int) {
	//var result Exam
	resp, err := a.Http.R().
		SetTimeout(12 * time.Second).
		//SetResult(&result).
		SetQueryParams(map[string]string{
			"gnmkdm": "N358105",
			"doType": "query",
		}).
		SetFormData(map[string]string{
			"xnm":                    year,
			"xqm":                    TERM[term],
			"_search":                "false",
			"ksmcdmb_id":             "",
			"kch":                    "",
			"kc":                     "",
			"ksrq":                   "",
			"kkbm_id":                "",
			"nd":                     strconv.FormatInt(time.Now().UnixMilli(), 10),
			"queryModel.showCount":   "15",
			"queryModel.currentPage": "1",
			"queryModel.sortName":    "",
			"queryModel.sortOrder":   "asc",
			"time":                   "1", // 查询次数
		}).
		Post(baseCfg.Exam)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.IsStatusFailure() {
		log.Println("GetExam HTTP 状态码错误:", resp.Status())
	}
	if resp.ResultError() != nil {
		log.Println(resp.ResultError(), resp.String())
	}
	if a.LoginCheck(resp) {
		// Ctrl里有关掉重定向是302，不关是200
		//return true
		fmt.Println(resp.String())
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}

	return
}

type Exam struct {
	CurrentPage   int  `json:"currentPage"`
	CurrentResult int  `json:"currentResult"`
	EntityOrField bool `json:"entityOrField"`
	Items         []struct {
		XhId        string `json:"xh_id"`
		Ksfs        string `json:"ksfs"`
		Bj          string `json:"bj"`
		Cdbh        string `json:"cdbh"`
		Ksmc        string `json:"ksmc"`
		Kssj        string `json:"kssj"`
		Kch         string `json:"kch"`
		Kkxy        string `json:"kkxy"`
		Cxbj        string `json:"cxbj"`
		Xqm         string `json:"xqm"`
		Khfs        string `json:"khfs"`
		Cdmc        string `json:"cdmc"`
		Cdxqmc      string `json:"cdxqmc"`
		Sksj        string `json:"sksj"`
		Kcmc        string `json:"kcmc"`
		Pycc        string `json:"pycc"`
		Totalresult int    `json:"totalresult"`
		Njmc        string `json:"njmc"`
		Jgmc        string `json:"jgmc"`
		Jxbmc       string `json:"jxbmc"`
		Sjbh        string `json:"sjbh"`
		Xb          string `json:"xb"`
		Zymc        string `json:"zymc"`
		Xf          string `json:"xf"`
		Xh          string `json:"xh"`
		Jxbzc       string `json:"jxbzc"`
		Xnmc        string `json:"xnmc"`
		Cdjc        string `json:"cdjc"`
		Xm          string `json:"xm"`
		Xnm         string `json:"xnm"`
		Xqmc        string `json:"xqmc"`
		Jsxx        string `json:"jsxx"`
		Xqmmc       string `json:"xqmmc"`
		RowId       int    `json:"row_id"`
		Zxbj        string `json:"zxbj"`
		Jxdd        string `json:"jxdd"`
	} `json:"items"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	PageNo    int    `json:"pageNo"`
	PageSize  int    `json:"pageSize"`
	ShowCount int    `json:"showCount"`
	SortName  string `json:"sortName"`
	SortOrder string `json:"sortOrder"`
	//Sorts       []interface{} `json:"sorts"`
	TotalCount  int `json:"totalCount"`
	TotalPage   int `json:"totalPage"`
	TotalResult int `json:"totalResult"`
}
