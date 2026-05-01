package client

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
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
	var n int
	fmt.Sscanf(s, "%d", &n)
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
				fmt.Sprintf("%d", slot.Weekday),
				fmt.Sprintf("%d", slot.Start),
				fmt.Sprintf("%d", slot.End),
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
		year = fmt.Sprintf("%d", now.Year()-1)
		term = 2
	case month >= 8 && month <= 12 || month == 1:
		year = fmt.Sprintf("%d", now.Year())
		term = 1
	}
	return year, term
}
