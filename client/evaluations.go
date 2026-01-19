package client

import (
	"fmt"
	baseCfg "school_sdk/config"
	"time"
)

// 瞎做的没什么用的模块

func (a *APIClient) Evaluation() {
	resp, err := a.Http.R().
		SetQueryParams(map[string]string{
			"gnmkdm": "N408125",
			//"layout": "default",
			//"su": a.account,
		}).Get(baseCfg.Evaluations)
	fmt.Println(resp, err)
}

func (a *APIClient) EvaluationQuery() T {
	var result T
	resp, err := a.Http.R().
		SetQueryParams(map[string]string{
			"gnmkdm": "N408125",
			"doType": "query",
			//"layout": "default",
			//"su": a.account,
		}).SetResult(&result).
		Post(baseCfg.Evaluations)
	if err != nil {
		fmt.Println(err.Error())
		return result
	}
	if a.LoginCheck(resp) {
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}
	fmt.Println(resp, err)
	fmt.Println(result)
	return result
}

func (a *APIClient) GetTeacherPhoto(id string) T {
	var result T
	resp, err := a.Http.R().
		SetQueryParams(map[string]string{
			"t":      fmt.Sprint(time.Now().UnixMilli()),
			"ignore": "1",
			"jgh_id": id,
			//"layout": "default",
			//"su": a.account,
		}).SetResult(&result).
		Get(baseCfg.TeacherPhoto)
	if err != nil {
		fmt.Println(err.Error())
		return result
	}
	if a.LoginCheck(resp) {
		// Ctrl里有关重定向是302，不关是200
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
		//continue
	}

	fmt.Println(resp, err)
	fmt.Println(result)
	return result
}

type T struct {
	Pxzt    string `json:"pxzt"`
	Message string `json:"message"`
	List    []struct {
		JghId string `json:"jgh_id"`
		Jszc  string `json:"jszc,omitempty"`
		Jsbm  string `json:"jsbm"`
		Pxzt  string `json:"pxzt"`
		Jgh   string `json:"jgh"`
		Jsxm  string `json:"jsxm"`
	} `json:"list"`
	Status string `json:"status"`
}

type T2 struct {
	Model struct {
		Bprwjsxnxqm        string `json:"bprwjsxnxqm"`
		Bprwqsxnxqm        string `json:"bprwqsxnxqm"`
		Date               string `json:"date"`
		DateDigit          string `json:"dateDigit"`
		DateDigitSeparator string `json:"dateDigitSeparator"`
		Day                string `json:"day"`
		Jgpxzd             string `json:"jgpxzd"`
		Listnav            string `json:"listnav"`
		LocaleKey          string `json:"localeKey"`
		Month              string `json:"month"`
		Mrpzjjss           string `json:"mrpzjjss"`
		PageTotal          int    `json:"pageTotal"`
		Pageable           bool   `json:"pageable"`
		Pjjssj             string `json:"pjjssj"`
		Pjkssj             string `json:"pjkssj"`
		Pjmc               string `json:"pjmc"`
		Pkvalue            string `json:"pkvalue"`
		Pxms               string `json:"pxms"`
		QueryModel         struct {
			CurrentPage   int           `json:"currentPage"`
			CurrentResult int           `json:"currentResult"`
			EntityOrField bool          `json:"entityOrField"`
			Limit         int           `json:"limit"`
			Offset        int           `json:"offset"`
			PageNo        int           `json:"pageNo"`
			PageSize      int           `json:"pageSize"`
			ShowCount     int           `json:"showCount"`
			Sorts         []interface{} `json:"sorts"`
			TotalCount    int           `json:"totalCount"`
			TotalPage     int           `json:"totalPage"`
			TotalResult   int           `json:"totalResult"`
		} `json:"queryModel"`
		Rangeable   bool   `json:"rangeable"`
		Sysdate     string `json:"sysdate"`
		TotalResult string `json:"totalResult"`
		UserModel   struct {
			Monitor    bool   `json:"monitor"`
			RoleCount  int    `json:"roleCount"`
			RoleKeys   string `json:"roleKeys"`
			RoleValues string `json:"roleValues"`
			Status     int    `json:"status"`
			Usable     bool   `json:"usable"`
		} `json:"userModel"`
		Xnm   string `json:"xnm"`
		Xnmmc string `json:"xnmmc"`
		Xqm   string `json:"xqm"`
		Xqmmc string `json:"xqmmc"`
		Year  string `json:"year"`
	} `json:"model"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type T3 struct {
	Pxzt    string `json:"pxzt"`
	Message string `json:"message"`
	List    []struct {
		JghId string `json:"jgh_id"`
		Jszc  string `json:"jszc,omitempty"`
		Jsbm  string `json:"jsbm"`
		Pxzt  string `json:"pxzt"`
		Jgh   string `json:"jgh"`
		Jsxm  string `json:"jsxm"`
	} `json:"list"`
	Status string `json:"status"`
}
