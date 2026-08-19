package client

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	baseCfg "school_sdk/config"
	"strconv"
	"strings"
	"time"
)

// 星期映射
var weekdayMap = map[string]int{
	"星期一": 1,
	"星期二": 2,
	"星期三": 3,
	"星期四": 4,
	"星期五": 5,
	"星期六": 6,
	"星期日": 7,
}

// TimeSlot 解析后的单个时间段
type TimeSlot struct {
	Weekday int
	Start   int
	End     int
	Weeks   string
}

// parseTeacher 解析教师字段，返回用顿号连接的姓名
func parseTeacher(jsxx string) string {
	if jsxx == "" || jsxx == "--" {
		return ""
	}
	parts := strings.Split(jsxx, ";")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		sub := strings.Split(p, "/")
		if len(sub) >= 2 {
			names = append(names, sub[1]) // 姓名在第二个位置
		} else {
			names = append(names, p)
		}
	}
	return strings.Join(names, ",")
}

// parseLocation 解析地点字段，返回地点切片
func parseLocation(jxdd string) []string {
	if jxdd == "" || jxdd == "--" {
		return []string{}
	}
	return strings.Split(jxdd, "<br/>")
}

// 预编译正则表达式
var (
	reWeekday = regexp.MustCompile(`(星期一|星期二|星期三|星期四|星期五|星期六|星期日)`)
	reSection = regexp.MustCompile(`第(\d+)-(\d+)节`)
	reWeeks   = regexp.MustCompile(`\{([^}]+)}`)
)

// parseTimeSlot 解析单个时间字符串，如 "星期一第5-6节{3-4周}"
func parseTimeSlot(timeStr string) *TimeSlot {
	weekdayMatch := reWeekday.FindStringSubmatch(timeStr)
	if len(weekdayMatch) < 2 {
		return nil
	}
	weekday := weekdayMap[weekdayMatch[1]]

	sectionMatch := reSection.FindStringSubmatch(timeStr)
	if len(sectionMatch) < 3 {
		return nil
	}
	start := atoi(sectionMatch[1])
	end := atoi(sectionMatch[2])

	weeksMatch := reWeeks.FindStringSubmatch(timeStr)
	weeks := ""
	if len(weeksMatch) >= 2 {
		weeksRaw := weeksMatch[1]
		weeks = strings.ReplaceAll(weeksRaw, "周", "")
		weeks = strings.TrimSpace(weeks)
	}

	return &TimeSlot{
		Weekday: weekday,
		Start:   start,
		End:     end,
		Weeks:   weeks,
	}
}

// parseSksj 解析整个 sksj 字段，返回多个时间段的解析结果
func parseSksj(sksj string) []*TimeSlot {
	if sksj == "" || sksj == "--" {
		return []*TimeSlot{}
	}
	timeStrings := strings.Split(sksj, "<br/>")
	slots := make([]*TimeSlot, 0, len(timeStrings))
	for _, ts := range timeStrings {
		if slot := parseTimeSlot(ts); slot != nil {
			slots = append(slots, slot)
		}
	}
	return slots
}

// 简易字符串转整数
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func outputWakeupCSV(courses []ChosenDic) {
	outputFile := "wakeup课表输出.csv"

	// 准备 CSV 行数据，第一行为表头
	rows := [][]string{
		{"课程名称", "星期", "开始节数", "结束节数", "老师", "地点", "周数"},
	}

	for _, course := range courses {
		// 跳过无时间或地点标记为 "--" 的课程（如形势与政策）
		if course.Sksj == "--" || course.Jxdd == "--" {
			continue
		}

		teacher := parseTeacher(course.Jsxx)
		locations := parseLocation(course.Jxdd)
		timeSlots := parseSksj(course.Sksj)

		// 遍历每个时间段，生成对应的行
		for i, slot := range timeSlots {
			location := ""
			if i < len(locations) {
				location = locations[i]
			} else if len(locations) > 0 {
				// 地点不足时重复最后一个
				location = locations[len(locations)-1]
			}
			row := []string{
				course.Kcmc,
				strconv.Itoa(slot.Weekday),
				strconv.Itoa(slot.Start),
				strconv.Itoa(slot.End),
				teacher,
				location,
				slot.Weeks,
			}
			rows = append(rows, row)
		}
	}

	// 创建输出文件
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 写入 UTF-8 BOM，便于 Excel 直接打开时不乱码
	_, err = f.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		fmt.Printf("写入 BOM 失败: %v\n", err)
		return
	}

	// 创建 CSV Writer 并写入所有行
	w := csv.NewWriter(f)
	err = w.WriteAll(rows)
	if err != nil {
		fmt.Printf("写入 CSV 失败: %v\n", err)
		return
	}

	fmt.Printf("成功生成课表文件：%s\n", outputFile)
}

func GetSuggestYearTerm2() (string, int) {
	now := time.Now()
	month := now.Month() // time.Month类型（1-12）
	var year = ""
	var term int
	switch {
	case month >= 2 && month <= 7:
		year = strconv.Itoa(now.Year() - 1)
		term = 2
	case month >= 8 && month <= 12 || month == 1:
		year = strconv.Itoa(now.Year())
		term = 1
	}
	return year, term
}

type KbList struct {
	Qsxqj string `json:"qsxqj"`
	Xsxx  struct {
		BJMC   string `json:"BJMC"`
		XNMC   string `json:"XNMC"`
		KXKXXQ string `json:"KXKXXQ"`
		XKKGXQ string `json:"XKKGXQ"`
		XKKG   string `json:"XKKG"`
		ZYHID  string `json:"ZYH_ID"`
		XHID   string `json:"XH_ID"`
		XH     string `json:"XH"`
		XQMMC  string `json:"XQMMC"`
		JFZT   int    `json:"JFZT"`
		XM     string `json:"XM"`
		XQM    string `json:"XQM"`
		XNM    string `json:"XNM"`
		YWXM   string `json:"YWXM"`
		NJDMID string `json:"NJDM_ID"`
		JSXM   string `json:"JSXM"`
		KCMS   int    `json:"KCMS"`
		ZYMC   string `json:"ZYMC"`
	} `json:"xsxx"`
	SjkList []struct { // 实践课列表
		Cxbj               string `json:"cxbj"`
		Date               string `json:"date"`
		DateDigit          string `json:"dateDigit"`
		DateDigitSeparator string `json:"dateDigitSeparator"`
		Day                string `json:"day"`
		Jgpxzd             string `json:"jgpxzd"`
		Jsxm               string `json:"jsxm"`             // 教师姓名
		Jxbzh              string `json:"jxbzh"`            // 一起上课的教学班级
		Kclb               string `json:"kclb"`             // 课程类别
		Kcmc               string `json:"kcmc"`             // 课程名称
		Khfsmc             string `json:"khfsmc,omitempty"` // 考核方式名称
		Kklxdm             string `json:"kklxdm"`           // 类型代码
		Listnav            string `json:"listnav"`
		LocaleKey          string `json:"localeKey"`
		Month              string `json:"month"`
		Njxh               int    `json:"njxh"`
		PageTotal          int    `json:"pageTotal"`
		Pageable           bool   `json:"pageable"`
		Qsjsz              string `json:"qsjsz"`  // "9-10周"
		Qtkcgs             string `json:"qtkcgs"` // "形势与政策61孙莉(共2周)\/9-10周\/无"
		QueryModel         struct {
			CurrentPage   int  `json:"currentPage"`
			CurrentResult int  `json:"currentResult"`
			EntityOrField bool `json:"entityOrField"`
			Limit         int  `json:"limit"`
			Offset        int  `json:"offset"`
			PageNo        int  `json:"pageNo"`
			PageSize      int  `json:"pageSize"`
			ShowCount     int  `json:"showCount"`
			//Sorts         []interface{} `json:"sorts"`
			TotalCount  int `json:"totalCount"`
			TotalPage   int `json:"totalPage"`
			TotalResult int `json:"totalResult"`
		} `json:"queryModel"`
		Rangeable   bool   `json:"rangeable"`
		Rsdzjs      int    `json:"rsdzjs"`
		Sfkckkb     bool   `json:"sfkckkb"`
		Sfsjk       string `json:"sfsjk,omitempty"`
		Sjkcgs      string `json:"sjkcgs"` // "形势与政策61孙莉(共2周)\/9-10周"
		TotalResult string `json:"totalResult"`
		UserModel   struct {
			Monitor    bool   `json:"monitor"`
			RoleCount  int    `json:"roleCount"`
			RoleKeys   string `json:"roleKeys"`
			RoleValues string `json:"roleValues"`
			Status     int    `json:"status"`
			Usable     bool   `json:"usable"`
		} `json:"userModel"`
		Xf    string `json:"xf"`     // 学分
		Xksj  string `json:"xksj"`   // 选课时间
		Xnmc  string `json:"xnmc"`   // "2025-2026"
		XqhId string `json:"xqh_id"` // 校区号id
		Xqmc  string `json:"xqmc"`   // 校区名称
		Xqmmc string `json:"xqmmc"`
		Year  string `json:"year"`
	} `json:"sjkList"`
	Sjfwkg   bool `json:"sjfwkg"`
	XqjmcMap struct {
		Field1 string `json:"1"`
		Field2 string `json:"2"`
		Field3 string `json:"3"`
		Field4 string `json:"4"`
		Field5 string `json:"5"`
		Field6 string `json:"6"`
		Field7 string `json:"7"`
	} `json:"xqjmcMap"`
	Xskbsfxstkzt string `json:"xskbsfxstkzt"`
	//RqazcList    []interface{} `json:"rqazcList"`
	KbList []struct { // 课表列表
		Bklxdjmc           string `json:"bklxdjmc"`
		CdId               string `json:"cd_id"`
		Cdbh               string `json:"cdbh"`
		Cdlbmc             string `json:"cdlbmc"`
		Cdmc               string `json:"cdmc"`
		Cxbj               string `json:"cxbj"`
		Cxbjmc             string `json:"cxbjmc"`
		Date               string `json:"date"`
		DateDigit          string `json:"dateDigit"`
		DateDigitSeparator string `json:"dateDigitSeparator"`
		Day                string `json:"day"`
		Jc                 string `json:"jc"`
		Jcor               string `json:"jcor"`
		Jcs                string `json:"jcs"`
		JghId              string `json:"jgh_id"`
		Jgpxzd             string `json:"jgpxzd"`
		JxbId              string `json:"jxb_id"`
		Jxbmc              string `json:"jxbmc"`
		Jxbsftkbj          string `json:"jxbsftkbj"`
		Jxbzc              string `json:"jxbzc"`
		Kcbj               string `json:"kcbj"`
		Kch                string `json:"kch"`
		KchId              string `json:"kch_id"`
		Kclb               string `json:"kclb"`
		Kcmc               string `json:"kcmc"`
		Kcxszc             string `json:"kcxszc"`
		Kcxz               string `json:"kcxz"`
		Kczxs              string `json:"kczxs"`
		Khfsmc             string `json:"khfsmc"`
		Kklxdm             string `json:"kklxdm"`
		Kkzt               string `json:"kkzt"`
		Ksfsmc             string `json:"ksfsmc"`
		Lh                 string `json:"lh"`
		Listnav            string `json:"listnav"`
		LocaleKey          string `json:"localeKey"`
		Month              string `json:"month"`
		Njxh               int    `json:"njxh"`
		Oldjc              string `json:"oldjc"`
		Oldzc              string `json:"oldzc"`
		PageTotal          int    `json:"pageTotal"`
		Pageable           bool   `json:"pageable"`
		Pkbj               string `json:"pkbj"`
		Px                 string `json:"px"`
		Qqqh               string `json:"qqqh"`
		QueryModel         struct {
			CurrentPage   int  `json:"currentPage"`
			CurrentResult int  `json:"currentResult"`
			EntityOrField bool `json:"entityOrField"`
			Limit         int  `json:"limit"`
			Offset        int  `json:"offset"`
			PageNo        int  `json:"pageNo"`
			PageSize      int  `json:"pageSize"`
			ShowCount     int  `json:"showCount"`
			//Sorts         []interface{} `json:"sorts"`
			TotalCount  int `json:"totalCount"`
			TotalPage   int `json:"totalPage"`
			TotalResult int `json:"totalResult"`
		} `json:"queryModel"`
		Rangeable   bool   `json:"rangeable"`
		Rk          string `json:"rk"`
		Rsdzjs      int    `json:"rsdzjs"`
		Sfjf        string `json:"sfjf"`
		Sfkckkb     bool   `json:"sfkckkb"`
		Skfsmc      string `json:"skfsmc"`
		Sxbj        string `json:"sxbj"`
		TotalResult string `json:"totalResult"`
		UserModel   struct {
			Monitor    bool   `json:"monitor"`
			RoleCount  int    `json:"roleCount"`
			RoleKeys   string `json:"roleKeys"`
			RoleValues string `json:"roleValues"`
			Status     int    `json:"status"`
			Usable     bool   `json:"usable"`
		} `json:"userModel"`
		Xf       string `json:"xf"`
		Xkbz     string `json:"xkbz"`
		Xkrs     string `json:"xkrs"`
		Xm       string `json:"xm"`
		Xnm      string `json:"xnm"`
		Xqdm     string `json:"xqdm"`
		Xqh1     string `json:"xqh1"`
		XqhId    string `json:"xqh_id"`
		Xqj      string `json:"xqj"`
		Xqjmc    string `json:"xqjmc"`
		Xqm      string `json:"xqm"`
		Xqmc     string `json:"xqmc"`
		Xsdm     string `json:"xsdm"`
		Xslxbj   string `json:"xslxbj"`
		Year     string `json:"year"`
		Zcd      string `json:"zcd"`
		Zcmc     string `json:"zcmc"`
		Zfjmc    string `json:"zfjmc"`
		Zhxs     string `json:"zhxs"`
		Zxs      string `json:"zxs"`
		Zxxx     string `json:"zxxx"`
		Zyfxmc   string `json:"zyfxmc"`
		Zyhxkcbj string `json:"zyhxkcbj"`
		Zzmm     string `json:"zzmm"`
		Zzrl     string `json:"zzrl"`
	} `json:"kbList"`
	XsbjList []struct {
		Xslxbj string `json:"xslxbj"`
		Xsmc   string `json:"xsmc"`
		Xsdm   string `json:"xsdm"`
		Ywxsmc string `json:"ywxsmc,omitempty"`
	} `json:"xsbjList"`
	Zckbsfxssj string `json:"zckbsfxssj"`
	//DjdzList     []interface{} `json:"djdzList"`
	Kblx    int    `json:"kblx"`
	Sfxsd   string `json:"sfxsd"`
	Jfckbkg bool   `json:"jfckbkg"`
	//XqbzxxszList []interface{} `json:"xqbzxxszList"`
	Xkkg     bool   `json:"xkkg"`
	Sxgykbbz string `json:"sxgykbbz"`
	//JxhjkcList   []interface{} `json:"jxhjkcList"`
	Xnxqsfkz string `json:"xnxqsfkz"`
}

func (a *APIClient) getSelectedList(xkxnm, xkxqm string) KbList {
	// 查询已选课程
	fmt.Println("个人课表查询")
	var result KbList
	for {
		resp, err := a.hedgeC.R().
			SetTimeout(time.Second*23).
			SetQueryParam("gnmkdm", "N2151").
			//SetQueryParam("su", a.Config.Account).
			SetFormData(map[string]string{
				"xnm":    xkxnm,
				"xqm":    xkxqm,
				"kzlx":   "ck",
				"xsdm":   "",
				"kclbdm": "",
				"kclxdm": "",
			}).
			SetResult(&result).
			Post(baseCfg.Kbcx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时", resp.Duration())
				return result
			} else {
				fmt.Println("请求发生错误")
				log.Println(err)
			}
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("getChoosedList 状态码:", resp.Status())
			continue
		}
		if resp.ResultError() != nil {
			log.Println(resp.ResultError(), resp.String())
			continue
		}
		if a.LoginCheck(resp) {
			log.Printf("已选课程查询: \n%s", resp.String())
			return result
		} else {
			a.ReLogin()
			continue
		}
	}
}

// outputKbListCSV 将 KbList 课表数据导出为 CSV 文件
func outputKbListCSV(kbList KbList) {
	outputFile := "课表输出.csv"

	// 准备 CSV 行数据，第一行为表头
	rows := [][]string{
		{"课程名称", "星期", "开始节数", "结束节数", "老师", "地点", "周数"},
	}

	// 遍历 kbList 中的每一条排课记录
	for _, item := range kbList.KbList {
		// 跳过无具体时间的课程（如无节次信息）
		if item.Jc == "" || item.Jc == "--" {
			continue
		}

		// 解析节次，提取开始和结束节数
		start, end := parseJc(item.Jc)

		// 星期直接使用 xqj（数字1-7）
		weekday := item.Xqj

		// 教师姓名（可能多人，逗号分隔）
		teacher := item.Xm

		// 地点
		location := item.Cdmc

		// 周次
		weeks := item.Zcd

		row := []string{
			item.Kcmc,
			weekday,
			strconv.Itoa(start),
			strconv.Itoa(end),
			teacher,
			location,
			weeks,
		}
		rows = append(rows, row)
	}

	// 创建输出文件
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("创建输出文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 写入 UTF-8 BOM
	_, err = f.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		fmt.Printf("写入 BOM 失败: %v\n", err)
		return
	}

	// 写入 CSV
	w := csv.NewWriter(f)
	err = w.WriteAll(rows)
	if err != nil {
		fmt.Printf("写入 CSV 失败: %v\n", err)
		return
	}

	fmt.Printf("成功生成课表文件：%s\n", outputFile)
}

// parseJc 解析节次字符串，如 "3-4节" 或 "1-4节"，返回开始和结束节数
func parseJc(jc string) (int, int) {
	// 去除 "节" 字，并按 "-" 分割
	trimmed := strings.TrimSuffix(jc, "节")
	parts := strings.Split(trimmed, "-")
	if len(parts) != 2 {
		// 如果格式不符，默认返回0，后续可根据需要处理
		return 0, 0
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return start, end
}
