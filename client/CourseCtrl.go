package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"school_sdk/utils"
	"school_sdk/utils/color"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/antchfx/htmlquery"
)

func (a *APIClient) userSetMode(cfg *APIConfig) string {
	fmt.Printf(`
********************************
功能代码如下:  %s%s%s
---------------------
[1;36m1[0m.选课
2,yxkc.已选课程查询
3,tk.退课
[1;36m4[0m.自动贪婪选课
[1;36m5[0m.自动单次选课
6,sx.刷新愿望清单
7,rf.重新获取参数
[1;36m9[0m.设定开始时间
0.其他
---------------------
ps:.前的值为功能代码
********************************`+"\n", color.Bold, cfg.modeName, color.Reset)
	code, err := utils.UserInputWithSigInt("请输入功能代码(-2 退出系统):")
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, terminal.InterruptErr) {
			a.Logout()
		}
		log.Println("input code err:", err)
		fmt.Println("", err)
		time.Sleep(1 * time.Second)
		//return code
	}
	log.Println("userInputCode:", code)
	if code == "-2" {
		a.Logout()
	}
	return code
}

func (a *APIClient) GetCourseCtl(modeCode string) {
	if modeCode != "" {
		log.Println("modeCode:", modeCode)
	}
	utils.PrintNotise()
	var cfg APIConfig
	var code string
	cfg.needInit = true
	// var tkList CustomCourseDic
	// tkList.Jxbmc = "/**-**/"
	sCustL := NewCustomCourseSlice()
	cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList = utils.ReadExcel()
	cfg.startTimeStamp = readStartTimeConfig()
	cfg.smtpConfig = utils.SMTPReadConfig()
	cfg.modeName = "(未初始化)"

	for {
		if modeCode != "" {
			code = modeCode
			modeCode = ""
		} else {
			code = a.userSetMode(&cfg)
		}
		switch code {
		case "6", "sx":
			refreshWant(&cfg)
			continue
		case "7", "rf":
			cfg.needInit = true
			cfg.listDump = false
			cfg.detailDump = false
		case "9":
			cfg.startTimeStamp = setTimeKeepSession()
			continue
		case "0":
			a.Other(&cfg)
			continue
		case "4", "jl", "5", "xk2":
			if !cfg.needInit && len(sCustL.items) != 0 {
				// 已经初始化之后，列表中有课，设置退课
				a.setQuitCourse(&cfg)
			}
		case "clear":
			sCustL = NewCustomCourseSlice()
			fmt.Println("已清空课程缓存")
			continue
		case "1", "xk", "2", "yxkc", "3", "tk", "boom":
		default:
			fmt.Println("无效的输入: ", code)
			continue
		}
		if cfg.startTimeStamp != time.Unix(0, 0) {
			a.timeKeepSession(cfg.startTimeStamp)
		}
		if cfg.needInit {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)
			go func() {
				defer close(done)
				for {
					a.getPubParams(ctx, &cfg, false)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					fmt.Printf("本学期已选学分 \033[1;36m%s\033[0m\n", cfg.zxfs)
					fmt.Printf("总学分最低: %s\n", cfg.minCredit)
					fmt.Printf("总学分最高: %s\n", cfg.maxCredit)
					fmt.Printf("本学期已选学分: %s\n", cfg.selectedCredit)

					a.getCourseListPre(ctx, &cfg, cfg.xkkz_id, cfg.xszxzt, false)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					fmt.Printf("距离选课结束还有 \033[1;36m%s\033[0m 天 共 \033[1;36m%s\033[0m 小时\n", cfg.syts, cfg.syxs)
					log.Printf("距离选课结束还有 %s 天 共 %s 小时", cfg.syts, cfg.syxs)

					switch code {
					case "1", "xk", "4", "jl", "5", "xk2":
						list := a.getCourseList(ctx, &cfg)
						if errors.Is(ctx.Err(), context.Canceled) || len(list) == 0 {
							return
						}
						sCustL.courseList2custom(list)
						sCustL.printCourse(&cfg)
						// fmt.Println("选课初始化完成")
						if len(sCustL.items) == 0 {
							fmt.Println("你可能没有课可选")
							return
							//panic("开发错误: 课程列表长度为0")
						}
						same, _ := sCustL.isKchIdAllSame()
						if same {
							// kch_id 都相同的逻辑
							// fmt.Println("kch_id 都相同")
							detail := a.getCourseDetail(ctx, &cfg, sCustL.Get(0).Kch_id)
							if errors.Is(ctx.Err(), context.Canceled) {
								return
							}
							if detail == nil {
								//needInit = true
								//return
								fmt.Println("将重新开始")
								continue
							}
							sCustL.courseDetail2custom(detail)
							//sCustL.printCourse()
						}
					}
					// signal.Stop(sigCh)
					cfg.needInit = false
					if code == "7" {
						return
					}
					break
				}
				return
			}()

			select {
			case <-done:
				signal.Stop(sigCh)
				close(sigCh)
			case <-sigCh:
				signal.Stop(sigCh)
				cancel()
				close(sigCh)
				fmt.Println("请求已取消")
				continue
			}

		}
		if cfg.needInit {
			continue
		}

		switch code {
		case "1", "xk":
			a.XK(&cfg, sCustL)
		case "2", "yxkc":
			a.getAlreadySelected(&cfg)
		case "3", "tk":
			a.quitSelected(&cfg)
		case "4", "jl":
			a.JL(&cfg, sCustL, false)
		case "5", "xk2":
			a.JL(&cfg, sCustL, true)
		case "7":
		case "boom":
			a.Boom(&cfg, sCustL)
		default:
			fmt.Println("开发阶段错误")
			log.Println("开发阶段错误")
		}
	}
}

func (a *APIClient) setQuitCourse(cfg *APIConfig) {
	// tkList.Jxbmc = "/**-**/"
	input, err := utils.UserInputWithSigInt("是否设置退课(y/N):")
	//var input string
	//fmt.Printf("是否设置退课(y/n默认):")
	//_, err := fmt.Scanln(&input)
	if err != nil {
		return
	}
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	if strings.Contains(input, "y") {
		fmt.Println(cfg.xkxnm, cfg.xkxqm)
		a.quitSelectedNormal(cfg)
		// fmt.Println("功能没做好")
	}
}

func (a *APIClient) Other(cfg *APIConfig) {
	log.Println("进入 Other 功能")
	for {
		var code string
		var err error
		fmt.Printf(`
********************************
1.课程模式切换 %d
2.邮件功能(SMTP) %t
3.设置余量 %t
4.查询成绩
5.自定义已选课程查询
6.设置timeout (%.1fs)
mail.测试邮件功能
gpa.查看GPA
color.色彩测试
en,zh.中英文切换
who,info.我是谁 %s
wakeup.输出csv课表
********************************`+"\n", len(cfg.modeStore), cfg.smtpConfig.Enable, cfg.yl, a.Http.Timeout().Seconds(), a.Config.Account)
		code, err = utils.UserInputWithSigInt("请输入功能代码(-1 退出其他):")
		if err != nil {
			return
		}
		log.Println("userInputCode:", code)
		switch code {
		case "-1", ".", "@":
			return
		case "0":
			a.detectTime(context.Background())
		case "1":
			a.setMode(cfg)
		case "2":
			if cfg.smtpConfig.Enable {
				cfg.smtpConfig.Enable = false
			} else {
				cfg.smtpConfig = utils.SMTPReadConfig()
				cfg.smtpConfig.Enable = true
				fmt.Println(cfg.smtpConfig.Host, cfg.smtpConfig.Port)
				fmt.Println(cfg.smtpConfig.From)
				fmt.Println(cfg.smtpConfig.To)
			}
		case "3":
			if cfg.yl {
				cfg.yl = false
				fmt.Println("设置余量 无")
			} else {
				cfg.yl = true
				fmt.Println("设置余量 有")
			}
		case "4":
			a.GetScoreWithInput()
		case "5":
			a.customGetSelected()
		case "6":
			input, err := utils.UserInputWithSigInt("请输入持续时间:")
			if err != nil {
				continue
			}
			duration, err := time.ParseDuration(input)
			if err != nil {
				fmt.Printf("无效的时间格式\n")
				fmt.Println("支持的格式：")
				fmt.Println(" 300ms - 300毫秒")
				fmt.Println(" 30s   - 30秒")
				fmt.Println(" 1m30s - 1分30秒")
				continue
			}
			fmt.Printf("  %v 秒\n", duration.Seconds())
			fmt.Printf("  %.3f 分钟\n", duration.Minutes())
			a.Http.SetTimeout(duration)
		case "7":
			if cfg.xztk {
				cfg.xztk = false
				fmt.Println("限制退选课程")
			} else {
				cfg.xztk = true
				fmt.Println("🚨不限制退课，你解锁了禁忌功能🚨")
			}
		case "mail":
			if cfg.smtpConfig.Enable {
				smtpContent := "<b>%s\n%s</b>"
				fmt.Println("Send mail")
				utils.SendMail(cfg.smtpConfig, "选课提醒测试", fmt.Sprintf(smtpContent, "*-选课成功✅?-*-", "游戏电竞课"))
			}
		case "gpa":
			a.getGPA()
		case "color":
			utils.TestTerminalColors()
		case "en":
			a.SwitchLanguage("en_US")
		case "zh":
			a.SwitchLanguage("zh_CN")
		case "who":
			if a.Name == "" {
				a.GetJsonInfo()
			}
			fmt.Println(a.Config.Account, a.Name)
		case "info":
			PrintStudentInfo2(a.GetJsonInfo())
		case "wakeup":
			a.wakeup()
		case "detect":
			a.detectKeepAliveTime()
		case "save":
			a.save(cfg)
		case "dev":
			a.devMode(cfg)
		default:
			fmt.Printf("没有 %s 哦\n", code)
		}
	}
}

func (a *APIClient) devMode(cfg *APIConfig) {
	for {
		fmt.Printf(`
********************************
1.设置代理 (%t)
2.选课参数来源于YzbIndex.html
3.选课参数来源于YzbDisplay.html
4.init = %t
5.打印cfg
url.Set BaseURL  (%s)
trace.
********************************`+"\n", a.Http.IsProxySet(), !cfg.needInit, a.Http.BaseURL())
		code, err := utils.UserInputWithSigInt("请输入功能代码:")
		if err != nil {
			return
		}
		switch code {
		case "-1", ".", "@":
			return
		case "1":
			if a.Http.IsProxySet() {
				a.Http.RemoveProxy() // It will not take effect immediately
				tls := a.Http.TLSClientConfig()
				tls.InsecureSkipVerify = false
				transport, err := a.Http.HTTPTransport()
				if err != nil {
					continue
				}
				transport.CloseIdleConnections()
				fmt.Println("取消代理")
			} else {
				proxy := "http://127.0.0.1:8866"
				fmt.Println("设置代理", proxy)
				a.Http.SetProxy(proxy)
				tls := a.Http.TLSClientConfig()
				tls.InsecureSkipVerify = true
			}
		case "2":
			fmt.Println("从 zzxkyzb_cxZzxkYzbIndex.html 文件设置选课参数")
			docNode, err := htmlquery.LoadDoc("zzxkyzb_cxZzxkYzbIndex.html")
			if err != nil {
				fmt.Println("大沙币:", err)
				continue
			}
			parseYzbIndexHtml(cfg, docNode)
		case "3":
			fmt.Println("从 zzxkyzb_cxZzxkYzbDisplay.html 文件设置选课参数")
			docNode, err := htmlquery.LoadDoc("zzxkyzb_cxZzxkYzbDisplay.html")
			if err != nil {
				//panic(err2)
				fmt.Println("大沙币:", err)
				continue
			}
			parseListPreHtml(cfg, docNode)
		case "3-":
			fmt.Println("等待模式修改:", cfg.modeName)
			a.getCourseListPre(context.Background(), cfg, cfg.xkkz_id, cfg.xszxzt, false)
			fmt.Println("模式设置为:", cfg.modeName)
		case "4":
			if cfg.needInit {
				cfg.needInit = false
				if cfg.modeName == "" {
					cfg.modeName = "develop"
				}
			} else {
				cfg.needInit = true
				if cfg.modeName == "develop" {
					cfg.modeName = ""
				}
			}
		case "5":
			fmt.Println("打印cfg")
			fmt.Printf("%+v", cfg)
		case "url":
			baseUrl, err := utils.UserInputWithSigInt("baseURL:")
			if err != nil {
				continue
			}
			if !strings.HasPrefix(baseUrl, "https://") && !strings.HasPrefix(baseUrl, "http://") {
				fmt.Println("无效的输入")
				continue
			}
			a.Http.SetBaseURL(baseUrl)
			fmt.Println("baseUrl设置为:", baseUrl)
		case "trace":
			a.Http.SetTrace(true)
		case "debugLog":
			a.Http.SetDebug(true)
			//a.Http.EnableDebugLog()
			//a.Http.EnableDumpAll()
			//case "devMode":
			//a.Http.DevMode()
		}
	}

}

func (a *APIClient) setMode(cfg *APIConfig) {
	context.Background()
	log.Println("特殊课程、通识选修课模式切换:", cfg.modeStore)
	if len(cfg.modeStore) == 0 {
		fmt.Println("没有模式切换选项哦")
		return
	}
	fmt.Println()
	for i, item := range cfg.modeStore {
		fmt.Println("====================")
		//fmt.Println(i, item.Kklxmc)
		//fmt.Println(i, item.Kklxdm)
		//fmt.Println(i, item.Xkkz_id)
		fmt.Printf("%d: %s  %s\n", i, item.Kklxmc, item.Kklxdm)
	}
	fmt.Println("========end=========")
	toChooseIdRow, err := utils.UserInputWithSigInt("输入模式前的序号:")
	if err != nil {
		return
	}
	if toChooseIdRow == "-1" {
		return
	}
	index, err1 := strconv.Atoi(strings.TrimSpace(toChooseIdRow))
	if err1 != nil {
		return
	}
	if 0 <= index && index < len(cfg.modeStore) {
		cfg.modeName = cfg.modeStore[index].Kklxmc
		cfg.kklxdm = cfg.modeStore[index].Kklxdm
		cfg.xkkz_id = cfg.modeStore[index].Xkkz_id
		fmt.Println("等待模式修改:", cfg.modeName)
		a.getCourseListPre(context.Background(), cfg, cfg.xkkz_id, cfg.xszxzt, false)
		fmt.Println("模式设置为:", cfg.modeName)
	} else {
		fmt.Println("无效的选择")
	}
	//fmt.Println("然后就没了")
	//time.Sleep(1 * time.Second)
}

func (a *APIClient) customGetSelected() {
	year, termInt := GetUserInputYearTerm(GetSuggestYearTerm2())
	if termInt == 0 {
		return
	}
	printSelectedList(a.getHaveSelectedList(year, TERM[termInt]))
}

func (a *APIClient) wakeup() {
	year, termInt := GetUserInputYearTerm(GetSuggestYearTerm2())
	if termInt == 0 {
		return
	}
	outputWakeupCSV(a.getHaveSelectedList(year, TERM[termInt]))
}

func show(cust *SafeCustomCourseSlice, workList []int) {
	fmt.Println("===================start============================")
	for _, I := range workList {
		FullPrint(I, cust.Get(I))
	}
	fmt.Println("====================end=============================")
}

func (a *APIClient) Boom(cfg *APIConfig, cust *SafeCustomCourseSlice) {
	if len(cust.items) == 0 {
		fmt.Println("当前没有课程缓存，需要初始化")
		return
	}
	log.Println("BOOM!")
	// 虽然可以进行课程爆破但是设计的太差了
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	done := make(chan struct{})
	defer close(done)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	go func() {
		select {
		case <-sigCh:
			fmt.Println("准备退出")
			cancel()
			// fmt.Println("<-sigCh")
		case <-done:
			// fmt.Println("<-done")
			cancel()
		}
	}()
	//numWorker := 2
	var workList []int
	var toChooseIdRow string
	var err error
	var wild bool
	cust.printCourse(cfg)
	for {
		fmt.Println("当前线程数:", len(workList))
		fmt.Println(workList)
		toChooseIdRow, err = utils.UserInputWithSigInt("(-1退出,start启动,show,wild):")
		if err != nil {
			close(sigCh)
			wg.Wait()
			return
		}
		if toChooseIdRow == "start" {
			break
		}
		switch toChooseIdRow {
		case "-1", ".", "@":
			return
		case "show":
			show(cust, workList)
			continue
		case "wild":
			if wild {
				wild = false
				fmt.Println("当前无法选的课将放弃")
				fmt.Println("时间，多选...")
			} else {
				wild = true
				fmt.Println("启用野蛮模式，不选到不退出")
			}
			continue
		}
		cust.printCourse(cfg)
		index, err1 := strconv.Atoi(strings.TrimSpace(toChooseIdRow))
		if err1 != nil {
			//fmt.Println(err1)
			index = -2
			// refresh print
			// continue
		}
		if 0 <= index && index < len(cust.items) {
			log.Println("b user select:", index, cust.items[index].Jxbmc)
			if cust.items[index].Do_jxb_id == "" {
				detail := a.getCourseDetail(ctx, cfg, cust.Get(index).Kch_id)
				if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
					return
				}
				cust.courseDetail2custom(detail)
			}
			FullPrintWithEnd(len(workList), cust.Get(index))
			workList = append(workList, index)
		}
	}
	show(cust, workList)
	fmt.Println(workList)
	fmt.Println("马上开始!")
	time.Sleep(3 * time.Second)
	for i := 0; i < len(workList); i++ {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(time.Duration(rand.Intn(130)+135) * time.Millisecond)
					chResult := a.chooseCourseRaw(cfg, &cust.items[workList[i]], ctx)
					if chResult.Flag == "1" {
						fmt.Printf("线程 %d 退出: %s\n", i, chResult.Msg)
						return
					}
					if chResult.Flag == "0" {
						if !wild && !checkCourseMsg(chResult) {
							fmt.Printf("线程 %d 退出: %s\n", i, chResult.Msg)
							return
						}
						fmt.Printf("线程 %d: %s\n", i, chResult.Msg)
					}
				}
			}
		})
	}
	wg.Wait()
}

func (a *APIClient) JL(cfg *APIConfig, cust *SafeCustomCourseSlice, single bool) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	defer close(sigCh)
	if len(cust.items) == 0 {
		list := a.getCourseList(ctx, cfg)
		if len(list) == 0 {
			cancel()
			return
		}
		cust.courseList2custom(list)
	}
	if len(cust.items) != 0 {
		cust.printCourse(cfg)
	}
	go func() {
		defer close(done)
		log.Println("进入 JL 功能")
		if len(cust.items) == 0 {
			return
		}
		same, _ := cust.isKchIdAllSame()
		signMent := []string{"|", "/", "-", "\\"}
		needEnter := false
		signNum := 0
		tryCount := 0
		var banList SafeCustomCourseSlice
		if same {
			log.Println("进入板块课选课")
			// 板块课
			//var runStat = true
			var tkDic CustomCourseDic

			// tkDic := tkList
			tkDic.Jxbmc = "/**-**/"
			var successQuit bool
			if cust.items[0].Do_jxb_id == "" {
				// get detail
				detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
				if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
					return
				}
				cust.courseDetail2custom(detail)
				cust.printCourse(cfg)
			}
			// user setup
			for {
				select {
				case <-ctx.Done():
					// 这里可以做清理工作，退出 Goroutine
					// fmt.Println("程序主逻辑已停止。")
					return
				default:
					// 自动决定选什么课
					stat, index := cust.checkWantAndHave(cfg, tkDic.Jxbmc, banList.items)
					if index == -3 {
						detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
							return
						}
						cust.courseDetail2custom(detail)
						if needEnter {
							fmt.Println()
						}
						cust.blockBanCourse(&banList.items)
						//fmt.Println(len(banList.items))
						cust.printCourse(cfg)
					} else if index == -2 {
						if needEnter {
							fmt.Println()
						}
						refreshWant(cfg)
						continue
					}
					// 满足选课条件
					if stat {
						FullPrintWithEnd(index, cust.items[index])
						// 操作退课，如果有
						if tkDic.Do_jxb_id != "" {
							successQuit, _ = a.quitCourse(cfg, tkDic.Do_jxb_id, tkDic.Kch_id)
						}
						// 选课
						succCh, chooseResult := a.HandChooseCourse(cfg, cust, index, sigCh)
						if succCh {
							// 选课成功，根据条件判断是否继续或退出
							if checkRank(cfg, cust.items[index].Jxbmc) == 0 || single {
								// 当选到第一志愿的课时可以返回
								if cfg.smtpConfig.Enable {
									smtpContent := "<b>%s\n%s</b>"
									go utils.SendMail(cfg.smtpConfig, "选课提醒", fmt.Sprintf(smtpContent, "*-选课成功✅-*-", cust.items[index].Jxbmc))
								}
								return
							}
							tkDic.Jxbmc = cust.items[index].Jxbmc
							tkDic.Do_jxb_id = cust.items[index].Do_jxb_id
							fmt.Println("自动设置退选课程为:", tkDic.Jxbmc)
							log.Println("自动设置退选课程为:", tkDic.Jxbmc)
						} else {
							if chooseResult.Flag == "0" {
								// 临时选课黑名单，将无法选的课纳入黑名单，避免盯着选
								fmt.Println("由于", chooseResult.Msg, "暂时将", cust.items[index].Jxbmc, "列为禁选")
								log.Println("由于", chooseResult.Msg, "暂时将", cust.items[index].Jxbmc, "列为禁选")
								banList.Append(cust.items[index])
							}
							if chooseResult.Flag == "-1" {
								yxrs_ := zf9ChooseResultParse(chooseResult)
								if yxrs_ == "" {
									cust.items[index].Jxbrl = cust.items[index].Yxzrs // zf8.0.0 不妥当
								}
							}
							// 选课失败，判断是否选回退的课
							if successQuit {
								chCResult := a.chooseCourseWithXXXXX(cfg, &tkDic, sigCh)
								fmt.Println("选回退课:", chCResult)
								log.Println("选回退课:", chCResult)
								// if chCResult.Flag == "1" {
								// 	successQuit = false
								// }
								successQuit = false
							}
						}
					}
				}
				// 刷新选课人数信息
				sign := signMent[signNum%4]
				fmt.Printf("\r%s 正在进行课程查询...", sign)
				needEnter = true
				signNum += 1
				time.Sleep(700 * time.Millisecond)
				list := a.getCourseList(ctx, cfg)
				if errors.Is(ctx.Err(), context.Canceled) || len(list) == 0 {
					return
				}
				cust.courseList2custom(list)
				cust.fix(cfg.yl, list)
				for _, item := range cust.items {
					if item.Do_jxb_id == "" {
						fmt.Println("发现教务系统有课程变化，自动刷新")
						log.Println("发现教务系统有课程变化，自动刷新")
						detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
							return
						}
						cust.courseDetail2custom(detail)
						cust.printCourse(cfg)
						continue
					}
				}
			}

		} else {
			log.Println("进入通识选修课选课")
			//cfg.yl = true
			// 校级选修课等其他课
			for {
				for i := range cust.items {
					select {
					case <-ctx.Done():
						// 这里可以做清理工作，退出 Goroutine
						// fmt.Println("程序主逻辑已停止。")
						return
					default:
						yxrs, err := strconv.Atoi(cust.items[i].Yxzrs)
						if err != nil {
							log.Println(err)
							fmt.Println(err)
							continue
						}
						if cust.items[i].Do_jxb_id == "" {
							// 判断获取价值
							if yxrs > guessGoodCourse(cust.items)*4/7 || rand.Float32() < 0.23 {
								time.Sleep(40 * time.Millisecond)
								detail := a.getCourseDetail(ctx, cfg, cust.Get(i).Kch_id)
								if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
									return
								}
								cust.courseDetail2custom(detail)
								fmt.Println("\r初始化选课信息", i)
								// cust.printCourse()
							} else {
								continue
							}
						}
						if cust.Get(i).Do_jxb_id != "" {
							// 判断余量，想要？
							jxbrl, err1 := strconv.Atoi(cust.Get(i).Jxbrl)
							if err1 != nil {
								log.Println(err1)
							}
							if cust.scanWantAuto(cfg, i) && yxrs != jxbrl && !courseInList(cust.Get(i).Jxb_id, banList.items) {
								// 发送选课
								if needEnter {
									fmt.Println()
								}
								FullPrintWithEnd(i, cust.Get(i)) // 发送前打印
								succCh, chooseResult := a.HandChooseCourse(cfg, cust, i, sigCh)
								if succCh && !single && tryCount < 22 {
									// 选成功了则不再选该课
									banList.Append(cust.items[i])
									tryCount += 1
									continue
								} else if succCh && single {
									if cfg.smtpConfig.Enable {
										smtpContent := "<b>%s\n%s</b>"
										go utils.SendMail(cfg.smtpConfig, "选课提醒", fmt.Sprintf(smtpContent, "*-选课成功✅-*-", cust.items[i].Jxbmc))
									}
									return
								} else if chooseResult.Flag == "0" {
									tryCount += 1
									banList.Append(cust.items[i])
									if tryCount > 11 {
										return
									}
								}
								//if chooseResult.Flag == "-1" {
								//	yxrs_ := zf9ChooseResultParse(chooseResult)
								//	if yxrs_ != "" {
								//		cust.items[i].Yxzrs = yxrs_
								//		cust.items[i].Jxbrl = yxrs_
								//	}
								//}
								if tryCount >= 20 && succCh {
									if cfg.smtpConfig.Enable {
										smtpContent := "<b>%s\n%s</b>"
										go utils.SendMail(cfg.smtpConfig, "选课提醒", fmt.Sprintf(smtpContent, "*-选课成功✅-*-", cust.items[i].Jxbmc))
									}
									return
								}
							}
							if yxrs > jxbrl && yxrs < jxbrl+3 && !courseInList(cust.Get(i).Jxb_id, banList.items) {
								// 刷新容量
								fmt.Println("\n发现教学班容量异常", i, cust.Get(i).Jxbmc)
								log.Println("发现教学班容量异常", i, cust.Get(i).Jxbmc)
								detail := a.getCourseDetail(ctx, cfg, cust.Get(i).Kch_id)
								if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
									return
								}
								cust.courseDetail2custom(detail)
								cust.printCourse(cfg)
								newJxbrl, err := strconv.Atoi(cust.Get(i).Jxbrl)
								if err != nil {
									log.Println(err)
								}
								if newJxbrl == jxbrl && err != nil {
									fmt.Println("将教学班容量拉平到已选人数")
									log.Println("将教学班容量拉平到已选人数")
									cust.items[i].Jxbrl = cust.items[i].Yxzrs
								}
							}
						}
					}

				}
				// 刷新人数
				sign := signMent[signNum%4]
				fmt.Printf("\r%s 正在进行课程查询...", sign)
				signNum += 1
				time.Sleep(850 * time.Millisecond)
				list := a.getCourseList(ctx, cfg)
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				cust.courseList2custom(list)
				cust.fix(cfg.yl, list)
			}
		}
	}()
	select {
	case <-sigCh:
		cancel()
		// fmt.Println("<-sigCh")
	case <-done:
		// fmt.Println("<-done")
		cancel()
	}
	return
}

func (a *APIClient) XK(cfg *APIConfig, cust *SafeCustomCourseSlice) {
	var wg sync.WaitGroup
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(sigCh)
	defer wg.Done()
	defer close(sigCh)
	go func() {
		defer close(done)
		log.Println("进入 XK 功能")
		if len(cust.items) != 0 {
			//fmt.Println("!=0")
		} else {
			//fmt.Println("==0")
			list := a.getCourseList(ctx, cfg)
			if errors.Is(ctx.Err(), context.Canceled) || len(list) == 0 {
				return
			}
			cust.courseList2custom(list)
			same, _ := cust.isKchIdAllSame()
			if same {
				detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
				if errors.Is(ctx.Err(), context.Canceled) || len(detail) == 0 {
					return
				}
				cust.courseDetail2custom(detail)
			}
			//cust.printCourse(cfg)
		}
		var toChooseId string
		for toChooseId != "-1" && toChooseId != "exit" {
			select {
			case <-ctx.Done():
				return
			default:
				// 模拟主逻辑
				cust.printCourse(cfg)
				toChooseIdRow, err := utils.UserInputWithSigInt("输入课程前的序号(-1退出,其它刷新):")
				if err != nil {
					return
				}
				toChooseId = strings.TrimSpace(toChooseIdRow)
				toChooseIdRow = "."
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				index, err1 := strconv.Atoi(toChooseId)
				if err1 != nil {
					//fmt.Println(err1)
					index = -2
					// refresh print
					// continue
				}
				if 0 <= index && index < len(cust.items) {
					log.Println("xk user select:", index, cust.items[index].Jxbmc)
					rs, err := strconv.Atoi(cust.Get(index).Yxzrs)
					if err != nil {
						log.Println(err)
					}
					if cust.items[index].Do_jxb_id == "" || cust.items[index].Jxbrl == "" {
						detail := a.getCourseDetail(ctx, cfg, cust.Get(index).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || len(detail) == 0 {
							return
						}
						cust.courseDetail2custom(detail)
					} else {
						jxbrl, err := strconv.Atoi(cust.Get(index).Jxbrl)
						if err != nil {
							log.Println(err)
						}
						if rs > jxbrl {
							detail := a.getCourseDetail(ctx, cfg, cust.Get(index).Kch_id)
							if errors.Is(ctx.Err(), context.Canceled) || len(detail) == 0 {
								return
							}
							cust.courseDetail2custom(detail)
						}
					}
					// 单独 printDetail
					FullPrintWithEnd(index, cust.items[index])
					// 让用户确认选择
					userInput, err2 := utils.UserInputWithSigInt(fmt.Sprintf("确认选择课程 \033[1;36m%s\033[0m ? (Y/n):", cust.items[index].Jxbmc))
					if err2 != nil {
						return
					}
					//var userInput string
					//fmt.Printf("确认选择课程 {\033[1;36m%s\033[0m} ? (Y/n,默认Y): ", cust.items[index].Jxbmc)
					//fmt.Printf("确认选择课程 \033[1;36m%s\033[0m ? (Y/n,默认Y): ", cust.items[index].Jxbmc)
					userInput = strings.ToLower(userInput)
					log.Printf("确认选择课程, userInput: (%s)", userInput)
					if strings.Contains(userInput, "n") || strings.Contains(userInput, ".") {
						continue
					}
					// 选课
					a.HandChooseCourse(cfg, cust, index, sigCh)
					//if chooseResult.Flag == "-1" {
					//yxrs := zf9ChooseResultParse(chooseResult)
					//if yxrs != "" {
					//	cust.items[index].Yxzrs = yxrs
					//	cust.items[index].Jxbrl = yxrs
					//}
					//}
					time.Sleep(600 * time.Millisecond)
				} else if index == -2 {
					// refresh print
					list := a.getCourseList(ctx, cfg)
					if errors.Is(ctx.Err(), context.Canceled) || len(list) == 0 {
						return
					}
					cust.courseList2custom(list)
					cust.fix(cfg.yl, list)
					same, _ := cust.isKchIdAllSame()
					if same {
						//a.getCourseDoJxb(ctx, cfg, jxbSlice(cust)) // 不顶用
						detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || len(detail) == 0 {
							return
						}
						cust.courseDetail2custom(detail)
					}
					//cust.printCourse(cfg)
				}
			}
		}
	}()
	select {
	case <-sigCh:
		cancel()
		fmt.Println("退出 XK")
		// fmt.Println("<-sigCh")
	case <-done:
		// fmt.Println("<-done")
		cancel()
	}
	return
}

//func jxbSlice(cust *SafeCustomCourseSlice) []string {
//	cust.mu.RLock()
//	defer cust.mu.RUnlock()
//	var jxb_ids []string
//	for _, item := range cust.items {
//		jxb_ids = append(jxb_ids, item.Jxb_id)
//	}
//	return jxb_ids
//}

func refreshWant(cfg *APIConfig) {
	log.Println("刷新愿望清单")
	cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList = utils.ReadExcel()
	fmt.Println("课程:", cfg.wantClassList)
	fmt.Println("教师:", cfg.wantTeacherList)
	fmt.Println("类型:", cfg.wantTypeList)
	log.Println(cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList)
}
