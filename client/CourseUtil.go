package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"school_sdk/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func parseKklxdmXkkzId(cfg *APIConfig, docNode *html.Node) {
	nodes := htmlquery.Find(docNode, `*//ul/li/a`)
	for _, item := range nodes {
		// 提取名称
		nameNode := htmlquery.FindOne(item, "./text()")
		if nameNode == nil {
			continue
		}
		name := nameNode.Data

		// 提取 onclick 属性
		onclick := htmlquery.SelectAttr(item, "onclick")
		if onclick == "" {
			continue
		}

		// 处理 onclick 字符串
		tmp := strings.TrimPrefix(onclick, "queryCourse(this,'")
		parts := strings.Split(tmp, "','")
		log.Println("parts:", parts)
		if len(parts) < 2 {
			continue
		}

		kklxdm := parts[0]
		xkkz_id := parts[1]
		// xkkz_id = strings.TrimSuffix(xkkz_id, "')")
		var store ModeStore
		store.Kklxmc = name
		store.Kklxdm = kklxdm
		store.Xkkz_id = xkkz_id
		cfg.modeStore = append(cfg.modeStore, store)
		fmt.Println("store:", store)
	}
	log.Println("modeStore:", cfg.modeStore)
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

func (s *SafeCustomCourseSlice) courseList2custom(list []CourseListDic) {
	// append_or_refersh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	found := false
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
		tmp.Yxzrs = list[i].Yxzrs
		tmp.Cxbj = list[i].Cxbj
		tmp.Date = list[i].Date
		tmp.DateDigit = list[i].DateDigit
		tmp.DateDigitSeparator = list[i].DateDigitSeparator

		s.items = append(s.items, tmp)

	}
}

func (s *SafeCustomCourseSlice) courseDetail2custom(list []CourseDetail) {
	// append_or_refresh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	for i := range list {
		for j := range s.items {
			if list[i].Jxb_id == s.items[j].Jxb_id {

				tmp := s.items[j]
				tmp.Do_jxb_id = list[i].Do_jxb_id
				tmp.Jxbrl = list[i].Jxbrl
				tmp.Sksj = list[i].Sksj
				tmp.Jxdd = list[i].Jxdd
				tmp.Jsxx = list[i].Jsxx
				tmp.Xqumc = list[i].Xqumc
				tmp.Xqh_id = list[i].Xqh_id
				tmp.Kcxzmc = list[i].Kcxzmc
				tmp.Kkxymc = list[i].Kkxymc
				tmp.Jxms = list[i].Jxms
				tmp.Kclbmc = list[i].Kclbmc

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

}

func guessGoodCourse(cust []CustomCourseDic) int {
	// 计算已选人数的平均值（忽略0）
	rsCount := 0
	zeroCount := 0
	for _, d := range cust {
		rs, err := strconv.Atoi(d.Yxzrs)
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
		for _, item := range cfg.wantClassList {
			if strings.Contains(list[index].Jxbmc, item) {
				list[index].Want = true
			}
		}
		for _, item := range cfg.wantTypeList {
			if strings.Contains(list[index].Kzmc, item) || strings.Contains(item, list[index].Kzmc) {
				if len(cfg.wantTeacherList) == 0 {
					list[index].Want = true
				}
				for _, item2 := range cfg.wantTeacherList {
					if strings.Contains(list[index].Jsxx, item2) {
						list[index].Want = true
					}
				}
			}
		}
		if len(cfg.wantTypeList) == 0 {
			for _, item2 := range cfg.wantTeacherList {
				if strings.Contains(list[index].Jsxx, item2) {
					list[index].Want = true
				}
			}
		}
	}
}

func (s *SafeCustomCourseSlice) printCourse(cfg *APIConfig) {
	s.mu.RLock()         // 加读锁（允许其他读，阻塞写）
	defer s.mu.RUnlock() // 确保解锁

	// 遍历副本（避免遍历过程中原切片被修改）
	itemsCopy := make([]CustomCourseDic, len(s.items))
	copy(itemsCopy, s.items)
	//itemsCopy := s.items // 浅拷贝
	reference := guessGoodCourse(itemsCopy)
	scanWant(cfg, itemsCopy)
	fmt.Println("===================目录=============================")
	for i, d := range itemsCopy {
		rs, err := strconv.Atoi(d.Yxzrs)
		if err != nil {
			log.Println(err)
		}
		if d.Do_jxb_id == "" {
			// 普通的 print list
			if d.Want {
				fmt.Printf("\033[0;33;40m-----👇--------------%d-------------------------------\033[0m\n", i)
				fmt.Printf("\033[1;36m%2d\033[0m: %s\n", i, d.Jxbmc)
				fmt.Printf("\033[1;36m%2d\033[0m: %2s 人已选  %s  %2s学分\n", i, d.Yxzrs, d.XF, d.Kzmc)
			} else {
				fmt.Printf("--------------------%d-------------------------------\n", i)
				if rs > reference {
					fmt.Printf("\033[1;36m%2d\033[0m: %s\n", i, d.Jxbmc)
					fmt.Printf("\033[1;36m%2d\033[0m: %2s 人已选  %s  %2s学分\n", i, d.Yxzrs, d.Kzmc, d.XF)
				} else {
					fmt.Printf("%2d: %s\n", i, d.Jxbmc)
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
	//var showName string
	//if strings.Contains(d.Jxbmc, d.Kcmc) {
	//	showName = d.Kcmc
	//} else {
	//	showName = d.Jxbmc
	//}
	if d.Want {
		fmt.Printf("\033[0;33;40m-----👇--------%d----⬇-want-⬇---%d---------------------\033[0m\n", i, i)
		fmt.Printf("\033[1;36m%2d\033[0m: %-5s %3s %-2s\n", i, d.Jxbmc, d.Xqumc, d.Sksj)
		fmt.Printf("\033[1;36m%2d\033[0m: ", i)
	} else {
		fmt.Printf("\r--------------%d---------------%d---------------------\n", i, i)
		fmt.Printf("%2d: %-5s %3s %-2s\n", i, d.Jxbmc, d.Xqumc, d.Sksj)
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
		fmt.Printf("%-11s \033[0;32;40m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	} else if rs == jxbrl {
		// 红色
		fmt.Printf("%-11s \033[0;31;40m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	} else {
		// 大于 亮黄色
		fmt.Printf("%-11s \033[1;33m%1s/%-2s\033[0m %1s分 %2s %2s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	}
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

func (s *SafeCustomCourseSlice) isKchIdAllSame() (bool, int) {
	s.mu.RLock()         // 加读锁（允许其他读，阻塞写）
	defer s.mu.RUnlock() // 确保解锁
	if len(s.items) == 0 {
		return false, -1
	}
	tmp := s.items[0].Kch_id
	for index, p := range s.items {
		if p.Kch_id == "" {
			fmt.Println("开发错误: kch_id 为空")
		}
		if tmp != p.Kch_id {
			return false, index
		}
	}
	return true, -1
}

func checkRank(cfg *APIConfig, currentClass string) int {
	var rank = len(cfg.wantClassList) // + 1
	if currentClass == "" {
		return rank
	}
	for i, item := range cfg.wantClassList {
		if strings.Contains(item, currentClass) || strings.Contains(currentClass, item) {
			rank = i
			break
		}
	}
	return rank
}

func courseInList(doJxbId string, list []CustomCourseDic) bool {
	if doJxbId == "" {
		return false
	}
	for _, d := range list {
		if d.Do_jxb_id == doJxbId {
			return true
		}
	}
	//slices.Contains(list, doJxbId)
	return false
}

// HandChooseCourse return 选课状态, flag
func (a *APIClient) HandChooseCourse(cfg *APIConfig, cust *SafeCustomCourseSlice, index int, sigCh chan os.Signal) (bool, string) {
	chooseResult := a.chooseCourseWithXXXXX(cfg, &cust.items[index], sigCh)
	if chooseResult.Flag == "1" {
		fmt.Println("*-选课成功✅-*-", cust.items[index].Jxbmc)
		log.Println("*-选课成功✅-*-", cust.items[index].Jxbmc)
		return true, chooseResult.Flag
	} else if chooseResult.Flag == "6" {
		fmt.Println("该教学班已选中，刷新页面可见！Msg:", chooseResult.Msg)
		log.Println("flag=6: ", chooseResult.Msg)
		return true, chooseResult.Flag
	} else if chooseResult.Flag == "0" {
		fmt.Println("选课失败: ", chooseResult.Msg)
		log.Println("选课失败: ", chooseResult.Msg)
		// sleep
		return false, chooseResult.Flag
	} else if chooseResult.Flag == "-1" {
		fmt.Println("选课失败: ", chooseResult.Msg)
		log.Println("选课失败: ", chooseResult.Msg)
		return false, chooseResult.Flag
	} else if chooseResult.Flag == "2" {
		fmt.Println("上课时间冲突且可查看冲突: ", chooseResult.Msg)
		log.Println("选课失败2: ", chooseResult.Msg)
		return false, chooseResult.Flag
	} else if chooseResult.Flag == "-5" {
		return false, chooseResult.Flag
	} else {
		log.Printf("warning: msg:%s flag:%s\n", chooseResult.Msg, chooseResult.Flag)
		fmt.Printf("msg:%s\n", chooseResult.Msg)
		if strings.Contains(chooseResult.Msg, "警告:你正在非法操作！") {
			fmt.Println("一般发生这个错误是因为脚本编写错误导致")
		} else {
			fmt.Println("未知错误")
		}
		return false, chooseResult.Flag
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
			fmt.Printf("\033[1;36m%d\033[0m: \033[1;36m%s\033[0m ", i, item.Jxbmc)
		} else {
			if strings.Contains(item.Jxbmc, item.Kcmc) {
				fmt.Printf("%d: %s ", i, item.Kcmc)
			} else {
				fmt.Printf("%d: %s ", i, item.Jxbmc)
			}
		}
		if len(item.Jsxx) > 32 {
			fmt.Printf("\n%s\n", item.Jsxx)
		} else {
			fmt.Printf("\t%s\n", item.Jsxx)
		}
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
		codeRow, err := utils.UserInputWithSigInt("输入要选择的课程前的序号(-1退出,其它刷新): ")
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
