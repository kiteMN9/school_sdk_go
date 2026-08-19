package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func parseKklxdmXkkzId(cfg *APIConfig, docNode *html.Node) {
	nodes := htmlquery.Find(docNode, `*//ul/li/a`)
	if len(nodes) != 0 {
		cfg.modeStore = nil
	}
	for _, item := range nodes {
		// 提取名称
		nameNode := htmlquery.FindOne(item, "./text()")
		if nameNode == nil {
			continue
		}

		// 提取 onclick 属性
		onclick := htmlquery.SelectAttr(item, "onclick")
		if onclick == "" {
			continue
		}

		// 处理 onclick 字符串
		tmp := strings.TrimPrefix(onclick, "queryCourse(this,'")
		parts := strings.Split(tmp, "','")
		//log.Println("parts:", parts)
		if len(parts) < 2 {
			continue
		}

		// xkkz_id = strings.TrimSuffix(xkkz_id, "')")
		var store ModeStore
		store.Kklxmc = nameNode.Data
		store.Kklxdm = parts[0]
		store.Xkkz_id = parts[1]
		cfg.modeStore = append(cfg.modeStore, store)
		//fmt.Println("store:", store)
	}
	log.Println("modeStore:", cfg.modeStore)
	if len(cfg.modeStore) != 0 {
		fmt.Printf("当前菜单项 %d\n", len(cfg.modeStore))
		for _, item := range cfg.modeStore {
			fmt.Printf("%s%-3s%s\n", BoldCyan, item.Kklxmc, Reset)
		}
		fmt.Println("=============")
	}
}

func (a *APIClient) chooseCourseWithXXXXX(cfg *APIConfig, co *CustomCourseDic, sigCh chan os.Signal) ChooseCourseResult {
	ctx, cancel := context.WithCancel(context.Background())
	var done = make(chan bool)
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-sigCh:
			cancel()
		}
	}()
	return a.chooseCourseRaw(cfg, co, ctx)
}

func zf9ChooseResultParse(result ChooseCourseResult) string {
	// 目前解决正方选课走后门导致自动选课异常的方法
	// {"msg":"0,394E58D5537473F8E065FCFCFE1D0407,130,","flag":"-1"}
	// fzjxb,jxb_id,yxzrs,blyxrs
	if result.Flag != "-1" {
		return ""
	}

	if strings.Contains(result.Msg, ",") {
		parts := strings.Split(result.Msg, ",")
		if len(parts) < 3 {
			log.Println("parts:", parts)
			return ""
		}
		// 获取第三部分
		valueStr := parts[2]
		// 验证是否只包含数字字符
		for _, r := range valueStr {
			if r < '0' || r > '9' {
				return ""
			}
		}
		// 验证长度和值
		if len(valueStr) == 0 {
			return ""
		}
		// 转换为数字进行大小验证
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return ""
		}
		// 确保大于10
		if value <= 10 {
			return ""
		}
		return valueStr
	}
	return ""
}

func (s *SafeCustomCourseSlice) fix(yl bool, list []CourseListDic) {
	if !yl {
		return
	}
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	for i := range s.items {
		found := false
		for j := range list {
			if list[j].Jxb_id == s.items[i].Jxb_id {
				found = true
				break
			}
		}
		if !found && s.items[i].Jxbrl != "" {
			s.items[i].Yxzrs = s.items[i].Jxbrl
		}
	}
}

func (s *SafeCustomCourseSlice) courseList2custom(list []CourseListDic) {
	// append_or_refersh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	found := false      // 发现是否已经在里面了，避免重复添加
	for i := range list {
		for j := range s.items {
			if list[i].Jxb_id == s.items[j].Jxb_id {
				// refersh
				var tmp = s.items[j]
				tmp.Jxbmc = list[i].Jxbmc
				tmp.Kch_id = list[i].Kch_id
				tmp.Kcmc = list[i].Kcmc
				tmp.Kklxdm = list[i].Kklxdm
				tmp.Kzmc = list[i].Kzmc
				tmp.XF = list[i].XF
				tmp.Xxkbj = list[i].Xxkbj
				tmp.Year = list[i].Year
				tmp.Yxzrs = list[i].Yxzrs
				tmp.Cxbj = list[i].Cxbj
				tmp.Date = list[i].Date
				tmp.DateDigit = list[i].DateDigit
				tmp.DateDigitSeparator = list[i].DateDigitSeparator
				// tmp. = list[i].
				// cust[j]. = list[i].DateDigitSeparator
				// nCustP.Update(j, tmp)
				if j >= len(s.items) {
					log.Println("list index out of range")
					fmt.Println("list index out of range")
					break
				}
				s.items[j] = tmp
				found = true
				break
			}
		}
		if found {
			found = false
			continue
		}
		// append
		var tmp CustomCourseDic
		tmp.Jxb_id = list[i].Jxb_id
		tmp.Jxbmc = list[i].Jxbmc
		tmp.Kch_id = list[i].Kch_id
		tmp.Kcmc = list[i].Kcmc
		tmp.Kklxdm = list[i].Kklxdm
		tmp.Kzmc = list[i].Kzmc
		tmp.XF = list[i].XF
		tmp.Xxkbj = list[i].Xxkbj
		tmp.Year = list[i].Year
		tmp.Yxzrs = list[i].Yxzrs // 没什么用
		tmp.Cxbj = list[i].Cxbj
		tmp.Date = list[i].Date
		tmp.DateDigit = list[i].DateDigit
		tmp.DateDigitSeparator = list[i].DateDigitSeparator
		// tmp. = list[i].
		// cust = append(cust, tmp)
		s.items = append(s.items, tmp)
		// s.Append(tmp) // 这里又锁了不知道会不会有问题
	}
}

func (s *SafeCustomCourseSlice) courseDetail2custom(list []CourseDetail) {
	// append_or_refresh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	for i := range list {
		for j := range s.items {
			if list[i].JxbId == s.items[j].Jxb_id {
				// refresh
				// var tmp CustomCourseDic
				// tmp := s.Get(j)
				tmp := s.items[j]
				tmp.Do_jxb_id = list[i].DoJxbId
				tmp.Jxbrl = list[i].Jxbrl
				tmp.Sksj = strings.ReplaceAll(list[i].Sksj, "<br/>", ",")
				tmp.Jxdd = list[i].Jxdd
				tmp.Jsxx = processLine(list[i].Jsxx)
				tmp.Xqumc = strings.ReplaceAll(list[i].Xqumc, "校", "")
				tmp.Xqh_id = list[i].Xqh_id
				tmp.Kcxzmc = list[i].Kcxzmc
				tmp.Kkxymc = list[i].Kkxymc
				tmp.Jxms = list[i].Jxms
				tmp.Kclbmc = list[i].Kclbmc
				// tmp. = list[i].
				// cust[j].DateDigit = list[i].DateDigit
				// s.Update(j, tmp)
				if j >= len(s.items) {
					log.Println("list index out of range")
					fmt.Println("list index out of range")
					break
				}
				s.items[j] = tmp
				break
			}
		}

	}
	return
}

// removePrefix 去掉单个字符串中的 "数字/" 前缀
func removePrefix(s string) string {
	idx := strings.Index(s, "/")
	if idx == -1 {
		// 没有斜杠，返回原字符串（或根据需求处理）
		return s
	}
	// 返回第一个斜杠之后的部分
	return s[idx+1:]
}

// processLine 处理可能包含多个条目的行（分号分隔）
func processLine(line string) string {
	parts := strings.Split(line, ";")
	for i, part := range parts {
		parts[i] = removePrefix(part)
	}
	return strings.Join(parts, ";")
}

func (s *SafeCustomCourseSlice) isKchIdAllSame() (bool, int) {
	s.mu.RLock()         // 加读锁（允许其他读，阻塞写）
	defer s.mu.RUnlock() // 确保解锁
	if len(s.items) == 0 {
		return false, 0
	}
	tmp := s.items[0].Kch_id
	for index := range s.items {
		if s.items[index].Kch_id == "" {
			fmt.Println("开发错误: kch_id 为空")
		}
		if tmp != s.items[index].Kch_id {
			return false, index
		}
	}
	// log.Println("kch_id is all same!")
	return true, -1
}

func FullPrintWithEnd(i int, d CustomCourseDic) {
	FullPrint(i, d)
	//fmt.Println("=====================<UNK>===============================")
	//topLine := fmt.Sprintf("--------------%d---------------%d---------------------\n", i, i)
	//endLine := "====================end==========================="
	//diff := len(topLine) - len(endLine)
	//diffStr := ""
	//for i := 0; i < diff; i++ {
	//	diffStr += "="
	//}
	//fmt.Println(endLine + diffStr)
	fmt.Println("====================end=============================")
}

func (s *SafeCustomCourseSlice) printCourse(cfg *APIConfig) {
	s.mu.Lock() // 加读锁（允许其他读，阻塞写）
	reference := guessGoodCourse(s.items)
	scanWant(cfg, s.items)
	s.mu.Unlock()

	s.mu.RLock()
	defer s.mu.RUnlock()
	fmt.Println("===================目录=============================")
	for i, d := range s.items {
		rs, err := strconv.Atoi(d.Yxzrs)
		if err != nil {
			log.Println(err)
		}
		if d.Do_jxb_id == "" {
			// 普通的 print list
			var showName string
			if strings.Contains(d.Jxbmc, d.Kcmc) {
				showName = d.Kcmc
			} else {
				showName = d.Jxbmc
			}
			if d.Want {
				fmt.Printf("\033[0;33;40m-----👇--------------%d-------------------------------\033[0m\n", i)
				fmt.Printf("\033[1;36m%2d\033[0m: %s\n", i, showName)
				fmt.Printf("\033[1;36m%2d\033[0m: %2s 人已选  %s  %2s学分\n", i, d.Yxzrs, d.XF, d.Kzmc)
			} else {
				fmt.Printf("--------------------%d-------------------------------\n", i)
				if rs > reference {
					fmt.Printf("\033[1;36m%2d\033[0m: %s\n", i, showName)
					fmt.Printf("\033[1;36m%2d\033[0m: %2s 人已选  %s  %2s学分\n", i, d.Yxzrs, d.Kzmc, d.XF)
				} else {
					fmt.Printf("%2d: %s\n", i, showName)
					fmt.Printf("%2d: %2s 人已选  %s  %2s学分\n", i, d.Yxzrs, d.Kzmc, d.XF)
				}
			}
		} else {
			// Full print
			FullPrint(i, d)
		}
	}
	fmt.Println("====================end==============================")
}

func FullPrint(i int, d CustomCourseDic) {
	var showName string
	if strings.Contains(d.Jxbmc, d.Kcmc) {
		showName = d.Kcmc
	} else {
		showName = d.Jxbmc
	}
	if d.Want {
		fmt.Printf("\033[0;33;40m-----👇--------%d----⬇-want-⬇---%d---------------------\033[0m\n", i, i)
		fmt.Printf("\033[1;36m%2d\033[0m: %-5s %3s %-2s\n", i, showName, d.Xqumc, d.Sksj)
		fmt.Printf("\033[1;36m%2d\033[0m: ", i)
	} else {
		fmt.Printf("\r--------------%d---------------%d---------------------\n", i, i)
		fmt.Printf("%2d: %-5s %3s %-2s\n", i, showName, d.Xqumc, d.Sksj)
		fmt.Printf("%2d: ", i)
	}
	jxbrl, err := strconv.Atoi(d.Jxbrl)
	if err != nil {
		log.Println(err)
	}
	rs, err1 := strconv.Atoi(d.Yxzrs)
	if err1 != nil {
		log.Println(err1)
	}
	if rs < jxbrl {
		// 绿色
		fmt.Printf("%-6s \033[0;32;40m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	} else if rs == jxbrl {
		// 红色
		fmt.Printf("%-6s \033[0;31;40m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	} else {
		// 大于 亮黄色
		fmt.Printf("%-6s \033[1;33m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	}
}

func guessGoodCourse(cust []CustomCourseDic) int {
	// 计算已选人数的平均值（忽略0）
	rsCount := 0
	zeroCount := 0
	for i := range cust {
		rs, err := strconv.Atoi(cust[i].Yxzrs)
		if err != nil {
			log.Println(err)
		}
		if rs == 0 {
			zeroCount += 1
		}
	}
	result := rsCount / (len(cust) - zeroCount + 1)
	return result
}

func scanWant(cfg *APIConfig, list []CustomCourseDic) {
	for index := range list {
		for i := range cfg.wantClassList {
			if strings.Contains(list[index].Jxbmc, cfg.wantClassList[i]) {
				list[index].Want = true
			}
		}
		for i := range cfg.wantTypeList {
			if strings.Contains(list[index].Kzmc, cfg.wantTypeList[i]) || strings.Contains(cfg.wantTypeList[i], list[index].Kzmc) {
				if len(cfg.wantTeacherList) == 0 {
					list[index].Want = true
				}
				for j := range cfg.wantTeacherList {
					if strings.Contains(list[index].Jsxx, cfg.wantTeacherList[j]) {
						list[index].Want = true
					}
				}
			}
		}
		if len(cfg.wantTypeList) == 0 {
			for i := range cfg.wantTeacherList {
				if strings.Contains(list[index].Jsxx, cfg.wantTeacherList[i]) {
					list[index].Want = true
				}
			}
		}
	}
}

// 遍历切片 read only
func (s *SafeCustomCourseSlice) scanWantAuto(cfg *APIConfig, index int) bool {
	var want = false
	s.mu.RLock()         // 加读锁（允许其他读，阻塞写）
	defer s.mu.RUnlock() // 确保解锁

	// 遍历副本（避免遍历过程中原切片被修改）
	//itemsCopy := make([]CustomCourseDic, len(s.items))
	//copy(itemsCopy, s.items)

	for i := range cfg.wantClassList {
		if strings.Contains(s.items[index].Jxbmc, cfg.wantClassList[i]) {
			want = true
		}
	}
	//slices.Contains(cfg.wantClassList, s.items[index].Jxbmc)
	for i := range cfg.wantTypeList {
		if strings.Contains(s.items[index].Kzmc, cfg.wantTypeList[i]) || strings.Contains(cfg.wantTypeList[i], s.items[index].Kzmc) {
			if len(cfg.wantTeacherList) == 0 {
				want = true
			}
			for j := range cfg.wantTeacherList {
				if strings.Contains(s.items[index].Jsxx, cfg.wantTeacherList[j]) {
					want = true
				}
			}
		}
	}
	if len(cfg.wantTypeList) == 0 {
		for i := range cfg.wantTeacherList {
			if strings.Contains(s.items[index].Jsxx, cfg.wantTeacherList[i]) {
				want = true
			}
		}
	}
	return want
}

func checkRank(cfg *APIConfig, currentClass string) int {
	var rank = len(cfg.wantClassList) // + 1
	if currentClass == "" {
		return rank
	}
	for i := range cfg.wantClassList {
		if strings.Contains(cfg.wantClassList[i], currentClass) || strings.Contains(currentClass, cfg.wantClassList[i]) {
			rank = i
			break
		}
	}
	return rank
}

func courseInList(JxbId string, list []CustomCourseDic) bool {
	if JxbId == "" {
		return false
	}
	for i := range list {
		if list[i].Jxb_id == JxbId {
			return true
		}
	}
	//slices.Contains(list, doJxbId)
	return false
}

func (s *SafeCustomCourseSlice) blockBanCourse(banList *[]CustomCourseDic) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.Do_jxb_id == "" {
			// 未获取到详细信息的课程跳过检索
			continue
		}
		jxbrl, err := strconv.Atoi(item.Jxbrl)
		if err != nil {
			log.Println(err)
		}
		yxrs, err1 := strconv.Atoi(item.Yxzrs)
		if err1 != nil {
			log.Println(err1)
			yxrs = jxbrl
		}
		if yxrs > jxbrl && !courseInList(item.Jxb_id, *banList) {
			*banList = append(*banList, item)
			//fmt.Println("banList = append(banList, item)")
		}
	}
}

func (s *SafeCustomCourseSlice) checkWantAndHave(cfg *APIConfig, tk string, banList []CustomCourseDic) (bool, int) {
	// 用于板块课捡漏
	// 课程列表的课和待退选的课谁的排名更靠前？
	// 读写锁？
	s.mu.RLock()         // 加读锁（允许其他读，阻塞写）
	defer s.mu.RUnlock() // 确保解锁
	rank := checkRank(cfg, tk)
	tmp := rank
	index := -1
	iErr := 0
	var teacherList []int
	for i := range s.items {
		if s.items[i].Do_jxb_id == "" {
			// 未获取到详细信息的课程跳过检索
			continue
		}
		for _, item2 := range cfg.wantTeacherList {
			if strings.Contains(s.items[i].Jsxx, item2) && !courseInList(s.items[i].Jxb_id, banList) {
				teacherList = append(teacherList, i)
			}
		}

		courseRank := checkRank(cfg, s.items[i].Jxbmc)
		jxbrl, err := strconv.Atoi(s.items[i].Jxbrl)
		if err != nil {
			log.Println(err)
		}
		yxrs, err1 := strconv.Atoi(s.items[i].Yxzrs)
		if err1 != nil {
			log.Println(err1)
			yxrs = jxbrl
		}
		// time_conflict
		if courseRank < tmp && yxrs < jxbrl && !courseInList(s.items[i].Jxb_id, banList) { // 这里应该考虑下改成 !=
			tmp = courseRank
			index = i
		} else if yxrs > jxbrl && !courseInList(s.items[i].Jxb_id, banList) {
			if yxrs-jxbrl >= 1 && yxrs-jxbrl <= 4 {
				// 极大概率是加容量了 TODO: 做个刷新记录功能刷新次数多了且人数超了列为禁选
				iErr = -3
			}
		}
	}
	if index < 0 {
		if index == -1 && len(cfg.wantClassList) == 0 {
			var input string
			fmt.Print("没设置愿望清单用个毛啊，快设置下愿望清单(Enter): ")
			_, err := fmt.Scanln(&input)
			if err != nil {
				return false, -2
			}
			return false, -2
		}
		if len(teacherList) != 0 {
			return true, teacherList[rand.Intn(len(teacherList))]
		}
		if iErr == -3 {
			fmt.Println("发现教学班容量异常，该刷新参数了")
			log.Println("发现教学班容量异常，教务系统有操作加容量嫌疑，该刷新参数了")
		}
		return false, iErr
	}
	if iErr == -3 {
		fmt.Println("由于已经找到要的课，忽略异常的容量")
	}
	courseName := s.items[index].Jxbmc
	if tmp < rank && !(strings.Contains(tk, courseName) || strings.Contains(courseName, tk)) {
		// if time_conflict
		return true, index
	}
	return false, -1
}
func checkCourseMsg(result ChooseCourseResult) bool {
	//选课频率过高，请稍后重试！
	if strings.Contains(result.Msg, "选课频率过高，请稍后重试！") {
		return true
	}
	//The frequency of course selection is too high, please try again in 20 seconds!
	if strings.Contains(result.Msg, "The frequency of course selection is too high") {
		return true
	}
	return false
}

// HandChooseCourse return 选课状态, flag
func (a *APIClient) HandChooseCourse(cfg *APIConfig, cust *SafeCustomCourseSlice, index int, sigCh chan os.Signal) (bool, ChooseCourseResult) {
	chooseResult := a.chooseCourseWithXXXXX(cfg, &cust.items[index], sigCh)
	if chooseResult.Flag == "1" {
		fmt.Println("*-选课成功✅-*-", cust.items[index].Jxbmc)
		log.Println("*-选课成功✅-*-", cust.items[index].Jxbmc)
		return true, chooseResult
	} else if chooseResult.Flag == "6" {
		fmt.Println("该教学班已选中，刷新页面可见！Msg:", chooseResult.Msg)
		log.Println("flag=6: ", chooseResult.Msg)
		return true, chooseResult
	} else if chooseResult.Flag == "0" { // 一门课程只能选一个教学班，不可再选、时间冲突
		if checkCourseMsg(chooseResult) {
			chooseResult.Flag = "-1"
			return false, chooseResult
		}
		fmt.Println("选课失败: ", chooseResult.Msg)
		log.Println("选课失败: ", chooseResult.Msg)
		// sleep
		return false, chooseResult
	} else if chooseResult.Flag == "-1" { // 人满了 //容量超出，重新修改页面上的选课人数信息
		yxrs := zf9ChooseResultParse(chooseResult)
		if yxrs != "" {
			cust.items[index].Yxzrs = yxrs
			cust.items[index].Jxbrl = yxrs
		}
		fmt.Println("选课失败: ", chooseResult.Msg)
		log.Println("选课失败: ", chooseResult.Msg)
		return false, chooseResult
	} else if chooseResult.Flag == "2" {
		fmt.Println("上课时间冲突且可查看冲突: ", chooseResult.Msg)
		log.Println("选课失败2: ", chooseResult.Msg, baseCfg.Conflict)
		return false, chooseResult
	} else if chooseResult.Flag == "-5" {
		//fmt.Println("傻卵不想等啊，那多等等吧")
		//log.Println("傻卵不想等啊，那多等等吧")
		return false, chooseResult
	} else {
		log.Printf("warning: msg:%s flag:%s\n", chooseResult.Msg, chooseResult.Flag)
		fmt.Printf("msg:%s\n", chooseResult.Msg)
		if strings.Contains(chooseResult.Msg, "警告:你正在非法操作！") {
			fmt.Println("一般发生这个错误是因为脚本编写错误导致")
		} else {
			fmt.Println("未知错误")
		}
		return false, chooseResult
	}
}

func (a *APIClient) getAlreadySelected(cfg *APIConfig) {
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigCh)
	go func() {
		printSelectedList(a.getHaveSelectedList(cfg.xkxnm, cfg.xkxqm))
		close(done)
	}()
	select {
	case <-sigCh:
		// fmt.Println("<-sigCh")
	case <-done:
		// fmt.Println("<-done")
	}
}

func truncateString(str string, maxLen int) string {
	if utf8.RuneCountInString(str) <= maxLen {
		return str
	}

	// 按rune截取，避免截断中文字符
	runes := []rune(str)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return str
}

func printSelectedList(selectedList []ChosenDic) {
	if selectedList == nil {
		fmt.Println("什么都没获取到")
		return
	}
	if len(selectedList) == 0 {
		fmt.Println("什么都没查到", selectedList)
		return
	}
	for i, item := range selectedList {
		fmt.Println("---------------------------------------------------")
		if item.Sfktk == "1" || item.IsInxksj == "1" {
			// 可以退课
			fmt.Printf("\033[1;36m%d\033[0m: \033[1;36m%s\033[0m %s", i, item.Jxbmc, item.Xf)
		} else {
			if strings.Contains(item.Jxbmc, item.Kcmc) {
				fmt.Printf("%d: %s \033[1;36m%s\033[0m", i, item.Kcmc, item.Xf)
			} else {
				fmt.Printf("%d: %s \033[1;36m%s\033[0m\n%d: %s ", i, item.Kcmc, item.Xf, i, item.Jxbmc)
			}
		}
		//if len(item.Jsxx) > 32 {
		fmt.Printf("\n%d: %s\n", i, truncateString(item.Jsxx, 35))
		//} else {
		//	fmt.Printf("\t%s\n", item.Jsxx)
		//}
	}
	fmt.Println("---------------------------------------------------")
}

func (a *APIClient) getAlreadySelectedTK(cfg *APIConfig) []ChosenDic {
	SelectedList := a.getHaveSelectedList(cfg.xkxnm, cfg.xkxqm)
	var quitList []ChosenDic
	if len(SelectedList) == 0 {
		fmt.Println("没有可退课程")
		return nil
	}
	var first = true
	i := 0
	fmt.Println("---------------------目录--------------------------")
	for _, item := range SelectedList {
		// isInxksj=="1" && sfxkbj=="1" && zcxkbj=="1")
		if item.Sfktk == "1" && (cfg.xztk || item.Sfxkbj == "1") {
			// 可以退课
			if first {
				first = false
			} else {
				fmt.Println("---------------------------------------------------")
			}
			var mc string
			if strings.Contains(item.Jxbmc, item.Kcmc) {
				mc = item.Kcmc
			} else {
				mc = item.Jxbmc
			}
			fmt.Printf("\033[1;36m%d\033[0m: \033[1;36m%s\033[0m %s\n", i, mc, item.Jsxx)
			var tmp ChosenDic
			tmp.Do_jxb_id = item.Do_jxb_id
			tmp.Jxbmc = item.Jxbmc
			quitList = append(quitList, tmp)
			i += 1
		} else {
			// fmt.Printf("%d: %s", i, item.Jxbmc)
		}
	}
	fmt.Println("---------------------end---------------------------")
	return quitList
}

func (a *APIClient) quitSelected(cfg *APIConfig) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigCh)
	defer wg.Done()
	go func() {
		defer close(done)
		fmt.Println("正在进行可退选课程查询...")
		quitList := a.getAlreadySelectedTK(cfg)
		if len(quitList) == 0 {
			fmt.Println("没有可以退的课")
			return
		}
		codeRow, err := utils.UserInputWithSigInt("输入要选择的课程前的序号(-1退出,其它刷新):")
		if err != nil {
			return
		}
		index, err1 := strconv.Atoi(strings.TrimSpace(codeRow))
		if err1 != nil {
			return
		}
		if 0 <= index && index < len(quitList) {
			fmt.Printf("退选课程: \033[1;36m%s\033[0m\n", quitList[index].Jxbmc)
			stat, msg := a.quitCourse(cfg, quitList[index].Do_jxb_id, quitList[index].Kch_id)
			if stat {
				fmt.Println("退课成功")
			} else {
				fmt.Println("退课失败:", msg)
				log.Println("quit msg:", msg)
			}
		}
		return
	}()
	select {
	case <-sigCh:
		// fmt.Println("<-sigCh")
		return
	case <-done:
		// fmt.Println("<-done")
		return
	}
}

func (a *APIClient) quitSelectedNormal(cfg *APIConfig) {
	fmt.Println("正在进行可退选课程查询...")
	quitList := a.getAlreadySelectedTK(cfg)
	if len(quitList) == 0 {
		fmt.Println("没有可以退的课")
		return
	}
	codeRow, err := utils.UserInputWithSigInt("请输入要退选的课程名字前的序号(-1退出):")
	if err != nil {
		return
	}
	index, err1 := strconv.Atoi(strings.TrimSpace(codeRow))
	if err1 != nil {
		return
	}
	if 0 <= index && index < len(quitList) {
		fmt.Printf("退选课程: %s\n", quitList[index].Jxbmc)
		//stat, msg := a.quitCourse(quitList[index].DoJxbId)
		stat, msg := a.quitCourse(cfg, quitList[index].Do_jxb_id, quitList[index].Kch_id)
		if stat {
			fmt.Println("退课成功")
		} else {
			fmt.Println("退课失败:", msg)
			log.Println("quit msg:", msg)
		}
	}

}

func (a *APIClient) cookie() {
	targetURL, _ := url.Parse(a.Http.BaseURL())
	cookies := a.Http.CookieJar().Cookies(targetURL)
	parts := make([]string, len(cookies))
	for i, c := range cookies {
		parts[i] = c.Name + "=" + c.Value
	}
	cookieStr := strings.Join(parts, "; ")
	fmt.Println(cookieStr)
}
