package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"school_sdk/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2/terminal"
)

func (a *APIClient) userSetMode(cfg *APIConfig) string {
	fmt.Printf(`
********************************
功能代码如下:  %s
---------------------
[1;36m1[0m.【xk】选课
2.【yxkc】已选课程查询
3.【tk】退课
6.【sx】刷新愿望清单
7.【rf】重新获取参数
[1;36m9[0m.设定开始时间
0.【0】其他(建设中)
---------------------
ps:【】内的值为功能代码
********************************`+"\n", cfg.modeName)
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

		case "clear":
			sCustL = NewCustomCourseSlice()
			continue
		case "1", "xk", "2", "yxkc", "3", "tk":
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
					a.getPubParams(ctx, &cfg)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					fmt.Printf("本学期已选学分 \033[1;36m%s\033[0m\n", cfg.zxfs)
					a.getCourseListPre(ctx, &cfg, cfg.xkkz_id, cfg.xszxzt)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					fmt.Printf("距离选课结束还有 \033[1;36m%s\033[0m 天 共 \033[1;36m%s\033[0m 小时\n", cfg.syts, cfg.syxs)
					log.Printf("距离选课结束还有 %s 天 共 %s 小时", cfg.syts, cfg.syxs)
					switch code {
					case "1", "xk":
						list := a.getCourseList(ctx, &cfg)
						if errors.Is(ctx.Err(), context.Canceled) {
							return
						}
						sCustL.courseList2custom(list)
						sCustL.printCourse(&cfg)
						if len(sCustL.items) == 0 {
							fmt.Println("你可能没有课可选")
							return
						}
						same, _ := sCustL.isKchIdAllSame()
						if same {
							detail := a.getCourseDetail(ctx, &cfg, sCustL.Get(0).Kch_id)
							if errors.Is(ctx.Err(), context.Canceled) {
								return
							}
							if detail == nil {

								fmt.Println("将重新开始")
								continue
							}
							sCustL.courseDetail2custom(detail)
						}
					}

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
		case "7":
		default:
			fmt.Println("开发阶段错误")
			log.Println("开发阶段错误")
		}
	}
}

func (a *APIClient) Other(cfg *APIConfig) {
	log.Println("进入 Other 功能")
	for {
		var code string
		var err error
		fmt.Printf(`
********************************
1.课程模式切换（待测试）
2.启用邮件功能(SMTP)
3.设置教务系统课程查询参数
4.查询成绩
5.自定义已选课程查询
mail.测试邮件功能
gpa.查看GPA
color.色彩测试
********************************` + "\n")
		code, err = utils.UserInputWithSigInt("请输入功能代码(-1 退出其他):")
		if err != nil {
			return
		}
		log.Println("userInputCode:", code)
		switch code {
		case "-1", ".", "@":
			return
		case "1":
			setMode(cfg)
		case "2":
			cfg.smtpConfig = utils.SMTPReadConfig()
			cfg.smtpConfig.Enable = true
			fmt.Println(cfg.smtpConfig.Host, cfg.smtpConfig.Port)
			fmt.Println(cfg.smtpConfig.From)
			fmt.Println(cfg.smtpConfig.To)
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
			if cfg.xztk {
				cfg.xztk = false
				fmt.Println("限制退选课程")
			} else {
				cfg.xztk = true
				fmt.Println("不限制退课")
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
		default:
			fmt.Printf("没有 %s 哦\n", code)
		}
	}
}

func setMode(cfg *APIConfig) {
	log.Println("特殊课程、通识选修课模式切换:", cfg.modeStore)
	if len(cfg.modeStore) == 0 {
		fmt.Println("没有模式切换选项哦")
		return
	}
	for _, item := range cfg.modeStore {
		fmt.Println(item.Kklxmc)
		fmt.Println(item.Kklxdm)
		fmt.Println(item.Xkkz_id)
		fmt.Println()
	}
	toChooseIdRow, err := utils.UserInputWithSigInt("输入模式前的序号: ")
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
		fmt.Println("模式设置为:", cfg.modeName)
	} else {
		fmt.Println("无效的选择")
	}
	//fmt.Println("然后就没了")
	//time.Sleep(1 * time.Second)
}

func (a *APIClient) devMode(cfg *APIConfig) {

}

func (a *APIClient) customGetSelected() {
	year, termInt := GetUserInputYearTerm()
	if termInt == 0 {
		return
	}
	printSelectedList(a.getHaveSelectedList(year, TERM[termInt]))
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
		} else {
			list := a.getCourseList(ctx, cfg)
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			cust.courseList2custom(list)
			same, _ := cust.isKchIdAllSame()
			if same {
				detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
				if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
					return
				}
				cust.courseDetail2custom(detail)
			}
		}
		var toChooseId string
		for toChooseId != "-1" {
			select {
			case <-ctx.Done():
				return
			default:
				// 模拟主逻辑
				cust.printCourse(cfg)
				toChooseIdRow, err := utils.UserInputWithSigInt("输入要选择的课程前的序号(-1退出,其它刷新): ")
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
					log.Println("user select:", index, cust.items[index].Jxbmc)
					jxbrl, err := strconv.Atoi(cust.Get(index).Jxbrl)
					if err != nil {
						log.Println(err)
					}
					rs, err := strconv.Atoi(cust.Get(index).Yxzrs)
					if err != nil {
						log.Println(err)
					}
					if cust.items[index].Do_jxb_id == "" || rs > jxbrl {
						detail := a.getCourseDetail(ctx, cfg, cust.Get(index).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
							return
						}
						cust.courseDetail2custom(detail)
					}
					// 单独 printDetail
					FullPrintWithEnd(index, cust.items[index])
					// 让用户确认选择
					userInput, err2 := utils.UserInputWithSigInt(fmt.Sprintf("确认选择课程 \033[1;36m%s\033[0m ? (Y/n,默认Y): ", cust.items[index].Jxbmc))
					if err2 != nil {
						return
					}
					userInput = strings.ToLower(userInput)
					log.Printf("确认选择课程, userInput: (%s)", userInput)
					if strings.Contains(userInput, "n") || strings.Contains(userInput, ".") {
						continue
					}
					// 选课
					a.HandChooseCourse(cfg, cust, index, sigCh)
					time.Sleep(600 * time.Millisecond)
				} else if index == -2 {
					// refresh print
					list := a.getCourseList(ctx, cfg)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					cust.courseList2custom(list)
					cust.fix(cfg.yl, list)
					same, _ := cust.isKchIdAllSame()
					if same {
						detail := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detail == nil {
							return
						}
						cust.courseDetail2custom(detail)
					}

				}
			}
		}
	}()
	select {
	case <-sigCh:
		cancel()
		fmt.Println("退出 XK")
	case <-done:
		cancel()
	}
	return
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

func refreshWant(cfg *APIConfig) {
	log.Println("刷新愿望清单")
	cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList = utils.ReadExcel()
	fmt.Println("课程:", cfg.wantClassList)
	fmt.Println("教师:", cfg.wantTeacherList)
	fmt.Println("类型:", cfg.wantTypeList)
	log.Println(cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList)
}
