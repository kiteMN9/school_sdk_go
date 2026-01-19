package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strconv"
	"strings"
	"time"
)

func (a *APIClient) getScoreRaw(year string, term int) []Score {
	var score ScoreRaw
	for {
		resp, err := a.Http.R().
			SetQueryParams(map[string]string{
				"doType": "query",
				"gnmkdm": "N305005",
				"su":     a.Account,
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
				"time":                   "0", //0,1
			}).
			SetTimeout(12 * time.Second).
			SetResult(&score).
			Post(baseCfg.SCORE)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("查成绩请求取消")
				log.Println("成绩请求已取消")
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("查成绩请求超时")
				log.Println("查成绩请求超时")
			}
			//	return
			//} else {
			//	fmt.Println(err)
			//}
			fmt.Println(err.Error())
			log.Println(err.Error())

			log.Println(resp.String())
			if strings.Contains(resp.String(), "Sorry, the page you are looking for is currently unavailable.") {
				fmt.Println("http状态码:", resp.Status())
				fmt.Println(resp.String())
			}
			return nil
		}

		if a.LoginCheck(resp) {
			// Ctrl里有关重定向是302，不关是200
		} else {
			fmt.Println(resp.Status())
			a.ReLogin()
			continue
		}

		return score.Items
	}
}

func formatPrintScoreAll(items []Score) {
	if len(items) == 0 {
		fmt.Println("未查到成绩数据")
		return
	}
	gksl := 0
	jqxf := 0.0
	fmt.Printf("%s百分成绩%s学分%s绩点%s课程名称  %s成绩%s 考试性质 学位课程 课程类别 课程性质\n",
		BoldCyan, Reset,
		GreenBgText, Reset,
		BoldCyan, Reset,
	)
	for _, item := range items {
		if formatPrintScore(item) {
			gksl++
			xf, err := strconv.ParseFloat(item.Xf, 64)
			if err != nil {
				continue
			}
			jqxf += xf
		}
	}
	// 做一个总结功能，总结一下获得多少学分拖欠多少学分，获得多少绩点 平均多少
	if jqxf != 0 {
		fmt.Println("挂科数量:"+BoldYellow, gksl, Reset+"总课程数:", len(items), "积欠学分:"+BoldYellow, jqxf, Reset)
	} else {
		fmt.Println("挂科数量:"+BoldYellow, gksl, Reset+"总课程数:", len(items))
	}
	log.Printf("%+v\n", items)
}
func formatPrintScore(d Score) bool {
	bfcj, err := strconv.Atoi(d.Bfzcj)
	if err != nil {
		log.Println(err)
	}
	return printScore(d, bfcj)
}

// 定义颜色常量
const (
	Reset      = "\033[0m"
	BoldCyan   = "\033[1;36m"
	BoldYellow = "\033[1;33m"

	//RedText     = "\033[31m"
	//GreenText   = "\033[32m"
	//RedBgText   = "\033[31;40m" // 红色文字，黑色背景

	BrightRedBg = "\033[1;91m"  // 红色文字，黑色背景
	GreenBgText = "\033[32;40m" // 绿色文字，黑色背景
)

func printScore(d Score, bfcj int) bool {
	if bfcj == 0 {
		// 缺考
		s := fmt.Sprintf("%2s %s%3s%s学 %s%-6s %s%2s%s ",
			d.Bfzcj,                 // 百分制成绩
			BoldYellow, d.Xf, Reset, // 学分
			//d.Jd, // 绩点

			BrightRedBg,
			d.Kcmc, // 课程名称
			BoldYellow,
			d.Cj, // 成绩
			BrightRedBg)
		s2 := fmt.Sprintf("%s(%s)%s %s %s %s %s%s\n",
			BoldYellow,
			d.Cjbz,
			BrightRedBg,
			d.Ksxz, d.Sfxwkc, d.Kclbmc, d.Kcxzmc,
			Reset)
		fmt.Print(s + s2)
		return true
	} else if bfcj < 60 {
		// 不及格
		s := fmt.Sprintf("%2s %s%3s%s学 %s%-6s %s%2s%s ",
			d.Bfzcj,
			BoldYellow, d.Xf, Reset,
			//d.Jd,
			BrightRedBg,

			d.Kcmc,
			BoldYellow,
			d.Cj,
			BrightRedBg)
		s2 := fmt.Sprintf("%s %s %s %s%s\n",
			d.Ksxz, d.Sfxwkc, d.Kclbmc, d.Kcxzmc,
			Reset)
		fmt.Print(s + s2)
		return true
	}

	// 合格
	fmt.Printf("%s%2s%s %3s学 %s%3s%s %-6s %s%2s%s %s %s %s %s\n",
		BoldCyan, d.Bfzcj, Reset,
		d.Xf,
		GreenBgText, d.Jd, Reset,

		d.Kcmc,
		BoldCyan, d.Cj, Reset,
		d.Ksxz, d.Sfxwkc, d.Kclbmc, d.Kcxzmc)
	return false
}

func (a *APIClient) GetScore(year string, term int) {
	//fmt.Println("get score")
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
	//fmt.Println(items)
	formatPrintScoreAll(items)
}

func (a *APIClient) GetScoreWithInput() {
	year, termInt := GetUserInputYearTerm()
	a.GetScore(year, termInt)
}

func GetUserInputYearTerm() (string, int) {
	var year, term, line = "2025", "1", ""
	var termInt int

	fmt.Printf("\033[1;36m%2s\033[0m 年 \033[1;36m%2s\033[0m 学期\n", year, term)
	for {
		var err error
		line, err = utils.UserInputWithSigInt(fmt.Sprintf("年份(%s):", year))
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
	var term_ string
	for {
		var err error
		line, err = utils.UserInputWithSigInt("学期(\u001B[1;36m1\u001B[0m,2,3):")
		if err != nil {
			return "0", 0
		}
		if line != "" && len(line) == 1 {
			term_ = line[0:1]
		} else {
			term_ = term
		}
		var err4 error
		termInt, err4 = strconv.Atoi(term_)
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
	//# njdm_id -> 年级代码
	Bfzcj   string `json:"bfzcj"` // 百分制成绩
	Bh      string `json:"bh"`    //
	Bh_id   string `json:"bh_id"`
	Bj      string `json:"bj"`
	Bzxx    string `json:"bzxx"`    // 备注信息
	Cjbz    string `json:"cjbz"`    // 成绩备注 缺考
	Czr     string `json:"czr"`     // cz人
	Cj      string `json:"cj"`      // 成绩
	Cjbdczr string `json:"cjbdczr"` // 陈爱华
	Cjbdsj  string `json:"cjbdsj"`  // 成绩bd时间？
	Cjsfzf  string `json:"cjsfzf"`  // 成绩是否作废

	Date               string `json:"date"`               // 二○二五年六月一日
	DateDigit          string `json:"dateDigit"`          // 2025年6月01日
	DateDigitSeparator string `json:"dateDigitSeparator"` // 2025-6-01

	Day    string `json:"day"`    //
	Jd     string `json:"jd"`     // 绩点 1.00
	Jg_id  string `json:"jg_id"`  // 003
	Jgmc   string `json:"jgmc"`   // 化学化工学院、应急管理与安全工程学院（合署）
	Jgpxzd string `json:"jgpxzd"` // 1
	Jsxm   string `json:"jsxm"`   // 教师姓名
	Jxb_id string `json:"jxb_id"` //
	Jxbmc  string `json:"jxbmc"`  // 教学班名称
	Kcbj   string `json:"kcbj"`   // 课程标记 主修

	Kch    string `json:"kch"`    // 课程号
	Kch_id string `json:"kch_id"` // 课程号ID
	Kclbmc string `json:"kclbmc"` // 课程类别 专业必修课
	Kcmc   string `json:"kcmc"`   // 课程名称
	Kcxzdm string `json:"kcxzdm"` // 课程性质名称001
	Kcxzmc string `json:"kcxzmc"` // 课程性质名称 选修

	Key string `json:"key"` // =Jxb_id+"-"+Xh

	Ksxz   string `json:"ksxz"`   // 考试性质 正常考试、补考一、重修
	Kkbmmc string `json:"kkbmmc"` // 开课部门名称 开课学院	数理学院 马克思主义学院 人文社会科学学院
	Khfsmc string `json:"khfsmc"` // 考核方式 考查 考试
	Kklxdm string `json:"kklxdm"` // 板块课 主修课程 特殊课程 通识选修课

	Year    string `json:"year"`    //
	Rwzxs   string `json:"rwzxs"`   // 什么人数？ "40"
	Sfdkbcx string `json:"sfdkbcx"` // 否
	Sfxwkc  string `json:"sfxwkc"`  // 是否学位课程 否
	Sfzh    string `json:"sfzh"`    // 身份证号码
	Sfzx    string `json:"sfzx"`    // 是
	Tjrxm   string `json:"tjrxm"`   // 提交人姓名
	Tjsj    string `json:"tjsj"`    // 提交时间 2024-01-01 10:00:00
	Xb      string `json:"xb"`      // 性别 男
	Xbm     string `json:"xbm"`     // 性别码 男:1
	Xf      string `json:"xf"`      // 学分 2.5
	Xfjd    string `json:"xfjd"`    // 学分绩点 2.50
	Xh      string `json:"xh"`      // 学号
	Xh_id   string `json:"xh_id"`   // 001
	Xm      string `json:"xm"`      // 姓名
	Xnm     string `json:"xnm"`     // 2024
	Xnmmc   string `json:"xnmmc"`   // 2024-2025
	Xqm     string `json:"xqm"`     // 3
	Xqmmc   string `json:"xqmmc"`   // 1
	Zymc    string `json:"zymc"`    // 化学工程与工艺

	//Zymc string `json:"zymc"` // 化学工程与工艺
}
