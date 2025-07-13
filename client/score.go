package client

import (
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strconv"
	"strings"
	"time"
)

func (a *APIClient) getScoreRaw(year string, term int) []Score {
	for {
		resp, err := a.http.R().
			SetQueryParams(map[string]string{
				"doType": "query",
				"gnmkdm": "N305005",
				//"layout": "default",
				"su": a.Account,
			}).
			SetFormData(map[string]string{
				"xnm":                    year,
				"xqm":                    TERM[term],
				"_search":                "false",
				"nd":                     fmt.Sprint(time.Now().UnixMilli()),
				"queryModel.showCount":   "500", // 展示数量？
				"queryModel.currentPage": "1",
				"queryModel.sortName":    "",
				"queryModel.sortOrder":   "asc",
				"time":                   "0",
			}).
			//SetContext(ctx).
			SetResult(&ScoreRaw{}).
			Post(baseCfg.SCORE)

		if err != nil {

			fmt.Println(err.Error())
			return []Score{}
		}

		if a.LoginCheck(resp) {
			// Ctrl里有关重定向是302，不关是200
		} else {
			fmt.Println(resp.StatusCode())
			a.ReLogin()
			continue
		}
		result := resp.Result().(*ScoreRaw)

		log.Println(resp.String())
		if strings.Contains(resp.String(), "Sorry, the page you are looking for is currently unavailable.") {
			fmt.Println("http状态码:", resp.StatusCode())
			fmt.Println(resp.String())
		}
		return result.Items
	}
}

func formatPrintScoreAll(items []Score) {
	for _, item := range items {
		formatPrintScore(item)
	}
}
func formatPrintScore(d Score) {
	bfcj, err := strconv.Atoi(d.Bfzcj)
	if err != nil {
		log.Println(err)
	}
	if bfcj == 0 {
		fmt.Printf("\033[0;31;40m%s %s\t%s学分 %s (%s) %s绩点 %s %s %s %s\033[0m\n", d.Kcxzmc, d.Kcmc, d.Xf, d.Cj, d.Cjbz, d.Jd, d.Ksxz, d.Sfxwkc, d.Kkbmmc, d.Kclbmc)
	} else if bfcj < 60 {
		fmt.Printf("\033[0;31;40m%s %s\t%s学分 %s分 %s绩点 %s %s %s %s\033[0m\n", d.Kcxzmc, d.Kcmc, d.Xf, d.Cj, d.Jd, d.Ksxz, d.Sfxwkc, d.Kkbmmc, d.Kclbmc)
	} else {
		fmt.Printf("%s %s\t%s学分 \u001B[1;36m%s\u001B[0m分 %s绩点 %s %s %s %s %s\n", d.Kcxzmc, d.Kcmc, d.Xf, d.Cj, d.Jd, d.Bfzcj, d.Ksxz, d.Sfxwkc, d.Kkbmmc, d.Kclbmc)
	}
}

func (a *APIClient) GetScore(year string, term int) {

	if term == 0 {
		return
	}
	if len(year) != 4 || (term != 1 && term != 2 && term != 3) {
		fmt.Println("score 查询参数不合法")
		log.Println("score 查询参数不合法")
		time.Sleep(1 * time.Second)
		return
	}
	items := a.getScoreRaw(year, term)

	formatPrintScoreAll(items)
}

func (a *APIClient) GetScoreWithInput() {
	year, termInt := GetUserInputYearTerm()
	a.GetScore(year, termInt)
}

func GetUserInputYearTerm() (string, int) {
	var year, term, line = "2024", "2", ""
	var termInt int

	fmt.Printf("\033[1;36m%s\033[0m 年 \033[1;36m%s\033[0m 学期\n", year, term)
	for {
		var err error
		line, err = utils.UserInputWithSigInt("年份(2024):")
		if err != nil {
			return "0", 0
		}
		if line != "" && len(line) == 4 {
			year = line[0:4]
			break
		}
		if len(line) == 0 {
			break
		}
	}

	for {
		var err error
		line, err = utils.UserInputWithSigInt("学期(1,2,3):")
		if err != nil {
			return "0", 0
		}
		if line != "" && len(line) == 1 {
			term = line[0:1]
		}
		var err4 error
		termInt, err4 = strconv.Atoi(term)
		if err4 != nil {
			log.Println(err4)
		}
		if termInt >= 1 && termInt <= 3 {
			break
		}
	}
	return year, termInt
}

type ScoreRaw struct {
	CurrentPage   int     `json:"currentPage"`
	CurrentResult int     `json:"currentResult"`
	EntityOrField bool    `json:"entityOrField"`
	Items         []Score `json:"items"`
	Limit         int     `json:"limit"`
	Offset        int     `json:"offset"`
	PageNo        int     `json:"pageNo"`
	PageSize      int     `json:"pageSize"`
	ShowCount     int     `json:"showCount"`
	TotalCount    int     `json:"totalCount"`
	TotalPage     int     `json:"totalPage"`
	TotalResult   int     `json:"totalResult"`
}
type Score struct {
	Bfzcj   string `json:"bfzcj"`
	Bh      string `json:"bh"`
	Bh_id   string `json:"bh_id"`
	Bj      string `json:"bj"`
	Bzxx    string `json:"bzxx"`
	Cjbz    string `json:"cjbz"`
	Czr     string `json:"czr"`
	Cj      string `json:"cj"`
	Cjbdczr string `json:"cjbdczr"`
	Cjbdsj  string `json:"cjbdsj"`
	Cjsfzf  string `json:"cjsfzf"`

	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`

	Day    string `json:"day"`
	Jd     string `json:"jd"`
	Jg_id  string `json:"jg_id"`
	Jgmc   string `json:"jgmc"`
	Jgpxzd string `json:"jgpxzd"`
	Jsxm   string `json:"jsxm"`
	Jxb_id string `json:"jxb_id"`
	Jxbmc  string `json:"jxbmc"`
	Kcbj   string `json:"kcbj"`

	Kch    string `json:"kch"`
	Kch_id string `json:"kch_id"`
	Kclbmc string `json:"kclbmc"`
	Kcmc   string `json:"kcmc"`
	Kcxzdm string `json:"kcxzdm"`
	Kcxzmc string `json:"kcxzmc"`

	Key string `json:"key"`

	Ksxz   string `json:"ksxz"`
	Kkbmmc string `json:"kkbmmc"`
	Khfsmc string `json:"khfsmc"`
	Kklxdm string `json:"kklxdm"`

	Year    string `json:"year"`
	Rwzxs   string `json:"rwzxs"`
	Sfdkbcx string `json:"sfdkbcx"`
	Sfxwkc  string `json:"sfxwkc"`
	Sfzh    string `json:"sfzh"`
	Sfzx    string `json:"sfzx"`
	Tjrxm   string `json:"tjrxm"`
	Tjsj    string `json:"tjsj"`
	Xb      string `json:"xb"`
	Xbm     string `json:"xbm"`
	Xf      string `json:"xf"`
	Xfjd    string `json:"xfjd"`
	Xh      string `json:"xh"`
	Xh_id   string `json:"xh_id"`
	Xm      string `json:"xm"`
	Xnm     string `json:"xnm"`
	Xnmmc   string `json:"xnmmc"`
	Xqm     string `json:"xqm"`
	Xqmmc   string `json:"xqmmc"`
	Zymc    string `json:"zymc"`
}
