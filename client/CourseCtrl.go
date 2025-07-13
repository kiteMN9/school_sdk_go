package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/go-resty/resty/v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"school_sdk/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func (a *APIClient) userSetMode(cfg *APIConfig) string {
	var code string
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
			a.http.GetClient().CloseIdleConnections()
			os.Exit(0)
		}
		return ""
	}
	log.Println("userInputCode:", code)
	if code == "-2" {
		a.Logout()
		os.Exit(0)
	}
	return code
}

func (a *APIClient) GetCourseCtl() {
	// 这里设计的有点屎了
	utils.PrintNotise()
	a.http.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))
	var cfg APIConfig
	var code string
	cfg.needInit = true
	// var tkList CustomCourseDic
	// tkList.Jxbmc = "/**-**/"
	sCustL := NewCustomCourseSlice()
	cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList = utils.ReadExcel()
	cfg.startTimeStamp = readStartTimeConfig()
	cfg.modeName = "(未初始化)"

	for {
		code = a.userSetMode(&cfg)
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

		case "1", "xk", "2", "yxkc", "3", "tk":
		default:
			fmt.Println("无效的输入")
			continue
		}
		if cfg.startTimeStamp != 0 {
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
					fmt.Printf("\033[1;36m%s\033[0m\n", cfg.zxfs)
					a.getCourseListPre(ctx, &cfg, cfg.xkkz_id, cfg.xszxzt)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					fmt.Printf("距离选课结束还有 \033[1;36m%s\033[0m 天 共 \033[1;36m%s\033[0m 小时\n", cfg.syts, cfg.syxs)
					log.Printf("距离选课结束还有 %s 天 共 %s 小时", cfg.syts, cfg.syxs)
					switch code {
					case "1", "xk":
						listP := a.getCourseList(ctx, &cfg)
						if errors.Is(ctx.Err(), context.Canceled) {
							return
						}
						sCustL.courseList2custom(listP)
						sCustL.printCourse(&cfg)
						if len(sCustL.items) == 0 {
							panic("开发错误: 课程列表长度为0")
						}
						same, _ := sCustL.isKchIdAllSame()
						if same {
							detailP := a.getCourseDetail(ctx, &cfg, sCustL.Get(0).Kch_id)
							if errors.Is(ctx.Err(), context.Canceled) {
								return
							}
							if detailP == nil {

								fmt.Println("将重新开始")
								continue
							}
							sCustL.courseDetail2custom(detailP)

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

		switch code {
		case "1", "xk":
			a.XK(&cfg, sCustL)
		case "2", "yxkc":
			a.getAlreadySelected(&cfg)
		case "3", "tk":
			a.quitSelected(&cfg)

		default:
			fmt.Println("开发阶段错误")
			log.Println("开发阶段错误")
		}
	}
}

func (a *APIClient) Other(cfg *APIConfig) {
	log.Println("进入 Other 功能")
	fmt.Println("特殊课程、通识选修课模式切换:", cfg.modeStore)
	for {
		var code string
		var err error
		fmt.Printf(`
********************************
1.特殊课程、通识选修课模式切换（建设中）
2.配置邮件功能（没做）
3.设置教务系统课程查询参数（没做）
4.查询成绩
5.自定义已选课程查询
********************************` + "\n")
		code, err = utils.UserInputWithSigInt("请输入功能代码(-1 退出其他):")
		if err != nil {
			return
		}
		log.Println("userInputCode:", code)
		switch code {
		case "-1":
			return
		case "1":
			fmt.Println("特殊课程、通识选修课模式切换:", cfg.modeStore)
			fmt.Println("然后就没了")
		case "2", "3":
			fmt.Println("没做")
		case "4":
			a.GetScoreWithInput()
		case "5":
			a.customGetSelected()
		case "dev":
			fmt.Println("debug")

		default:
			fmt.Printf("没有 %s 哦\n", code)
		}
	}
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
			listP := a.getCourseList(ctx, cfg)
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			cust.courseList2custom(listP)
			same, _ := cust.isKchIdAllSame()
			if same {
				detailP := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
				if errors.Is(ctx.Err(), context.Canceled) || detailP == nil {
					return
				}
				cust.courseDetail2custom(detailP)
			}

		}
		var toChooseId string
		var toChooseIdRow string
		for toChooseId != "-1" {
			select {
			case <-ctx.Done():
				return
			default:
				// 模拟主逻辑
				cust.printCourse(cfg)
				fmt.Print("输入要选择的课程前的序号(-1退出,其它刷新): ")
				_, err := fmt.Scanln(&toChooseIdRow)
				if err == io.EOF {
					wg.Wait()
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
					log.Println("index in range")
					if cust.items[index].Do_jxb_id == "" {
						detailP := a.getCourseDetail(ctx, cfg, cust.Get(index).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detailP == nil {
							return
						}
						cust.courseDetail2custom(detailP)
					}
					// 单独 printDetail
					FullPrintWithEnd(index, cust.items[index])
					// 让用户确认选择
					var userInput string
					fmt.Printf("确认选择课程 \033[1;36m%s\033[0m ? (Y/n,默认Y): ", cust.items[index].Jxbmc)
					_, err2 := fmt.Scanln(&userInput)
					if err2 == io.EOF {
						wg.Wait()
						break
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
					listP := a.getCourseList(ctx, cfg)
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}
					cust.courseList2custom(listP)
					same, _ := cust.isKchIdAllSame()
					if same {
						detailP := a.getCourseDetail(ctx, cfg, cust.Get(0).Kch_id)
						if errors.Is(ctx.Err(), context.Canceled) || detailP == nil {
							return
						}
						cust.courseDetail2custom(detailP)
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

func refreshWant(cfg *APIConfig) {
	log.Println("刷新愿望清单")
	cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList = utils.ReadExcel()
	fmt.Println("课程:", cfg.wantClassList)
	fmt.Println("教师:", cfg.wantTeacherList)
	fmt.Println("类型:", cfg.wantTypeList)
	log.Println(cfg.wantClassList, cfg.wantTeacherList, cfg.wantTypeList)
}
