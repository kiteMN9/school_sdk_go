package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type startTime struct {
	StartTime int64 `json:"start_time"`
}

func readStartTimeConfig() int64 {
	filename := "startTime.json"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return 0
	}

	// 读取文件内容
	byteValue, err := os.ReadFile(filename)
	if err != nil {
		// panic(err)
		log.Println("startTime.json 文件读取失败")
		return 0
	}

	// 将 JSON 数据解析到结构体
	var timeConfig startTime
	err = json.Unmarshal(byteValue, &timeConfig)
	if err != nil {
		return 0
	}
	log.Println("timeFromFile:", timeConfig.StartTime, time.UnixMilli(timeConfig.StartTime).Format("2006-01-02_15:04:05"))
	return timeConfig.StartTime
}

func parseStartTime(timeStr string) int64 {
	var timeObj time.Time
	var err error

	// 日期格式
	layout := "2006-01-02 15:04:05"
	// 解析日期字符串为time.Time对象
	loc, _ := time.LoadLocation("Local")

	timeObj, err = time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		log.Println("解析时间失败:", err)
		//fmt.Println()
		return 0
	}
	// 获取时间戳
	timestamp := timeObj.UnixMilli()

	return timestamp
}

func setTimeKeepSession() int64 {
	for {
		// 指定日期字符串
		dateStr := "2025-09-04 12:30:00"
		fmt.Print("    参考: ", dateStr, "\n输入时间: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		if err := scanner.Err(); err != nil {
			continue

		}

		timestamp := parseStartTime(input)
		if timestamp == 0 {
			continue
		}
		log.Println("setStartTime:", input, timestamp)

		var configData startTime
		configData.StartTime = timestamp
		dataByte, err := json.Marshal(configData)
		if err != nil {
			panic(fmt.Sprintf("JSON序列化失败: %v", err))
			// continue
		}
		if err1 := os.WriteFile("startTime.json", dataByte, 0644); err1 != nil {
			panic(err1)
		}
		return timestamp

	}

}

func (a *APIClient) timeKeepSession(timestamp int64) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer close(done)
	go func() {
		select {
		case <-done:
		case <-sigCh:
			cancel()
			fmt.Println("请求已取消")
		}
		signal.Stop(sigCh)
		close(sigCh)
	}()
	var count int
	// var signMen = []string{"|", "/", "-", "\\"}
	var signMen = []string{"⠇", "⠏", "⠋", "⠙", "⠹", "⠼", "⠴", "⠦", "⠧"}
	var refreshCount int
	var sign = "|"
	var signNum int
	var start int64
	signNumStat := 0
	first := true

	fmt.Println("开始时间:", time.UnixMilli(timestamp).Format("2006-01-02_15:04:05"))
	for {
		if time.Now().UnixMilli() > timestamp+6 {
			fmt.Print("\r=========开始========= ")
			//close(done)
			return
		}
		if first {
			fmt.Println("未到指定时间，等待中...")
			first = false
		}
		time.Sleep(3 * time.Millisecond)
		count += 1
		// signNum = count / 79 % 4
		signNum = count / 35 % 9
		if signNumStat != signNum {
			sign = signMen[signNum]
			fmt.Printf("\r======%d=========  %s ", refreshCount, sign)
			signNumStat = signNum
		}
		if time.Now().UnixMilli()-start > 21000 {
			// 定时刷新
			refreshCount += 1
			if timestamp-16500 > time.Now().UnixMilli() {
				fmt.Printf("\r======%d====c====  %s ", refreshCount, sign)
				// time.Sleep(1 * time.Second)
				a.CheckSession(ctx)
				fmt.Printf("\r======%d=========  %s ", refreshCount, sign)
			}
			start = time.Now().UnixMilli()
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			//log.Println("保持登录已取消")
			return
		}
	}
}

func getXpathValue(docNode *html.Node, name string) string {
	nodes := htmlquery.FindOne(docNode, `//*[@id="`+name+`"]`)
	return htmlquery.SelectAttr(nodes, "value")
}

func (a *APIClient) getPubParams(ctx context.Context, cfg *APIConfig) {
	log.Println("=======================get_pub_params()=======================")
	needEnter := false
	i := 0
	for {
		i++

		resp, err := a.http.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			Get(baseCfg.CHOOSE_COURSE_INDEX)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Println("请求已取消")
				return
			}
			fmt.Println("请求出错: ", err)
			log.Println("请求发生错误")
			continue
		}

		htmlContent := utils.RemoveEmptyLines(resp.String())
		if utils.UserIsLogin(a.Account, htmlContent) && !a.CheckLogout302(resp) {
			// fmt.Println(htmlContent)
			// return
		} else {
			a.ReLogin()
			continue
		}
		// fmt.Println(htmlContent)
		docNode, err1 := htmlquery.Parse(strings.NewReader(htmlContent)) // 相当于etree.HTML()
		if err1 != nil {
			log.Println(err1)
			fmt.Println("\r他妈个逼这什么情况，完成请求然后解析出错？")
			log.Println("他妈个逼这什么情况，完成请求然后解析出错？")
			continue
		}
		statNode := htmlquery.Find(docNode, `//div[@class="nodata"]/span/text()`)
		if len(statNode) != 0 {
			jdStr := strings.TrimSpace(htmlquery.InnerText(statNode[0]))
			fmt.Printf("\r%d %s", i, jdStr) // 对不起，当前不属于选课阶段
			log.Printf("%d %s", i, jdStr)
			needEnter = true
			time.Sleep(650 * time.Millisecond)

			continue
		}
		if needEnter {
			fmt.Println()
			needEnter = false
		}
		cfg.Kklxmc = getXpathValue(docNode, "firstKklxmc")
		cfg.xkkz_id = getXpathValue(docNode, "firstXkkzId")
		cfg.kklxdm = getXpathValue(docNode, "firstKklxdm")
		cfg.njdm_id = getXpathValue(docNode, "firstNjdmId")
		cfg.njdm_id_list0 = cfg.njdm_id
		cfg.bh_id = getXpathValue(docNode, "bh_id")
		cfg.xkxnm = getXpathValue(docNode, "xkxnm")
		cfg.xkxqm = getXpathValue(docNode, "xkxqm")
		cfg.zyh_id = getXpathValue(docNode, "firstZyhId")
		cfg.xqh_id = getXpathValue(docNode, "xqh_id")
		cfg.jg_id = getXpathValue(docNode, "jg_id_1")
		cfg.xz = getXpathValue(docNode, "xz")
		cfg.zyfx_id = getXpathValue(docNode, "zyfx_id")
		cfg.ccdm = getXpathValue(docNode, "ccdm")
		cfg.xbm = getXpathValue(docNode, "xbm")
		cfg.mzm = getXpathValue(docNode, "mzm")
		cfg.xsbj = getXpathValue(docNode, "xsbj")
		cfg.xslbdm = getXpathValue(docNode, "xslbdm")
		cfg.xszxzt = getXpathValue(docNode, "xszxzt")
		cfg.xxdm = getXpathValue(docNode, "xxdm")
		cfg.zxfs = getXpathValue(docNode, "zxfs")

		if cfg.kklxdm != "06" {
			fmt.Println("\r将要选 \033[1;36m" + cfg.Kklxmc + "\033[0m 这可能不太对(体育和英语进阶属于板块课)")
		} else {
			fmt.Println("\r将要选 \033[1;36m", cfg.Kklxmc, "\033[0m !!")
		}
		log.Println("\r将要选", cfg.Kklxmc, "!!")
		cfg.modeName = cfg.Kklxmc
		log.Println(htmlContent)
		parseKklxdmXkkzId(cfg, docNode)
		return
	}
}

func parseKklxdmXkkzId(cfg *APIConfig, docNode *html.Node) {
	// 未经测试
	log.Println("parse_kklxdm_xkkz_id debug")
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

		log.Println("name:", name)
		log.Println("kklxdm:", kklxdm)
		log.Println("xkkz_id:", xkkz_id)
		var store ModeStore
		store.Kklxmc = name
		store.Kklxdm = kklxdm
		store.Xkkz_id = xkkz_id
		cfg.modeStore = append(cfg.modeStore, store)
		log.Println("modeStore:", cfg.modeStore)
	}

	log.Println("parse_kklxdm_xkkz_id debug end")
}

func (a *APIClient) getCourseListPre(ctx context.Context, cfg *APIConfig, xkkz_id, xszxzt string) {
	// 补 齐搜索课程需要的发包参数
	log.Println("===============getCourseList_pre()=================")
	for {

		resp, err := a.http.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			SetFormData(map[string]string{
				"xkkz_id": xkkz_id,
				"xszxzt":  xszxzt, // 1
				"kspage":  "0",
				"jspage":  "0",
			}).
			Post(baseCfg.CHOOSE_COURSE_List_pre)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(ctx.Err(), context.Canceled) {
				fmt.Println("请求已取消")
				return
			}
			fmt.Println("请求出错: ", err)
			log.Println("请求发生错误")
			continue
		}

		htmlContent := utils.RemoveEmptyLines(resp.String())
		if a.LoginCheck(resp) {
			// fmt.Println(htmlContent)
			// return
		} else {
			a.ReLogin()
			continue
		}
		// fmt.Println(htmlContent)
		docNode, err := htmlquery.Parse(strings.NewReader(htmlContent))
		if err != nil {
			log.Println(err)
			fmt.Println("他妈个逼这什么情况，完成请求然后解析出错？")
			continue
		}
		cfg.bklx_id = getXpathValue(docNode, "bklx_id")
		cfg.rwlx = getXpathValue(docNode, "rwlx")
		cfg.xkly = getXpathValue(docNode, "xkly")
		cfg.sfkknj = getXpathValue(docNode, "sfkknj")
		cfg.rlkz = getXpathValue(docNode, "rlkz")
		cfg.kkbk = getXpathValue(docNode, "kkbk")
		cfg.kkbkdj = getXpathValue(docNode, "kkbkdj")
		cfg.jxbzcxskg = getXpathValue(docNode, "jxbzcxskg")
		cfg.xkxskcgskg = getXpathValue(docNode, "xkxskcgskg")
		cfg.sfkcfx = getXpathValue(docNode, "sfkcfx")
		cfg.sfkkjyxdxnxq = getXpathValue(docNode, "sfkkjyxdxnxq")
		cfg.gnjkxdnj = getXpathValue(docNode, "gnjkxdnj")
		cfg.sfkkzy = getXpathValue(docNode, "sfkkzy")
		cfg.kzybkxy = getXpathValue(docNode, "kzybkxy")
		cfg.sfznkx = getXpathValue(docNode, "sfznkx")
		cfg.zdkxms = getXpathValue(docNode, "zdkxms")
		cfg.sfkxq = getXpathValue(docNode, "sfkxq")
		cfg.bbhzxjxb = getXpathValue(docNode, "bbhzxjxb")
		cfg.xklc = getXpathValue(docNode, "xklc")
		cfg.rlzlkz = getXpathValue(docNode, "rlzlkz")
		cfg.sfkxk = getXpathValue(docNode, "sfkxk") // 是否可选课
		cfg.sfktk = getXpathValue(docNode, "sfktk") // 是否可退课
		cfg.jdlx = getXpathValue(docNode, "jdlx")   // 体育课多志愿开关
		cfg.syts = getXpathValue(docNode, "syts")   // 距选课结束还剩{0}天
		cfg.syxs = getXpathValue(docNode, "syxs")   // 距选课结束还剩{0}小时
		log.Println(htmlContent)
		return
	}
	// zdzys //"一门课程最多可选"+zdzys+"个志愿！"
	// sfqzxk //"一门课程只能选一个教学班！"
	// self.lnzgxkxf # 历学期选课最高学分要求为
	// bxqzgxkxf //基本选课规则设置中设置的最高选课学分
	// 本学期本类型课程选课最高学分要求为"+bxqzgxkxf+"，当前本学期本类型课程选课总学分为("+kklxzxfs+"+"+$("#xf_"+kch_id).text()+"="+c_kklxzxfs+")，超出选课最高学分要求，不可选！"
	// 本学期本类型课程选课最高门次要求为 bxqzgxkmc ，不可选
}

func (a *APIClient) getCourseList(ctx context.Context, cfg *APIConfig) *[]CourseListDic {
	// 搜索课程，主页面的查询
	if !cfg.listDump {
		log.Println("========搜索课程 getCourseList()========")
		fmt.Println("搜索课程")
	}
	for {
		formData := map[string]string{
			"rwlx":         cfg.rwlx,
			"xkly":         cfg.xkly,
			"bklx_id":      cfg.bklx_id,
			"sfkkjyxdxnxq": cfg.sfkkjyxdxnxq,
			"xqh_id":       cfg.xqh_id,
			"jg_id":        cfg.jg_id,
			"njdm_id_1":    cfg.njdm_id,
			"zyh_id_1":     cfg.zyh_id,
			"zyh_id":       cfg.zyh_id,
			"zyfx_id":      cfg.zyfx_id,
			"njdm_id":      cfg.njdm_id,
			"bh_id":        cfg.bh_id,
			"xbm":          cfg.xbm,
			"xslbdm":       cfg.xslbdm,
			"mzm":          cfg.mzm,
			"xz":           cfg.xz,
			"ccdm":         cfg.ccdm,
			"xsbj":         cfg.xsbj,
			"sfkknj":       cfg.sfkknj,
			"gnjkxdnj":     cfg.gnjkxdnj,
			"sfkkzy":       cfg.sfkkzy,
			"kzybkxy":      cfg.kzybkxy,
			"sfznkx":       cfg.sfznkx,
			"zdkxms":       cfg.zdkxms,
			"sfkxq":        cfg.sfkxq,
			"sfkcfx":       cfg.sfkcfx,
			"kkbk":         cfg.kkbk,
			"kkbkdj":       cfg.kkbkdj,

			"xkxnm":    cfg.xkxnm,
			"xkxqm":    cfg.xkxqm,
			"kklxdm":   cfg.kklxdm,
			"bbhzxjxb": cfg.bbhzxjxb,
			"rlkz":     cfg.rlkz,

			"kspage": "1",
			"jspage": "200",
			"jxbzb":  "",
		}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		req := a.http.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).
			SetFormData(formData).
			SetResult(&GetCourseListResult{})

		resp, err := req.Post(baseCfg.CHOOSE_COURSE_courseList)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Println("请求已取消")
				return &[]CourseListDic{}
			}
			// fmt.Printf("请求出错: %v\n", err)
			log.Println("请求发生错误")
			continue
		}

		if a.LoginCheck(resp) {
			if !cfg.listDump {
				log.Println(resp.String())
				cfg.listDump = true
			}
			result := resp.Result().(*GetCourseListResult)
			// fmt.Println(result.TmpList)
			// log.Println(result)
			return &result.TmpList
		} else {
			a.ReLogin()
			continue
		}
	}
}

func (a *APIClient) getCourseDetail(ctx context.Context, cfg *APIConfig, kch_id string) *[]CourseDetail {
	// 	查询课程具体信息
	// 	获得do_jxb_id（选课的必要参数）以及很重要的容量信息等等
	if !cfg.detailDump {
		fmt.Println("\r正在获取详细信息")
		log.Println("========查询课程具体信息 getCourseDetail()========")
	}
	for {
		formData := map[string]string{
			"rwlx":         cfg.rwlx,
			"xkly":         cfg.xkly,
			"bklx_id":      cfg.bklx_id,
			"sfkkjyxdxnxq": cfg.sfkkjyxdxnxq,
			"xqh_id":       cfg.xqh_id,
			"jg_id":        cfg.jg_id,
			"njdm_id_1":    cfg.njdm_id,
			"zyh_id":       cfg.zyh_id,
			"zyfx_id":      cfg.zyfx_id,
			"njdm_id":      cfg.njdm_id,
			"bh_id":        cfg.bh_id,

			"xbm":      cfg.xbm,
			"xslbdm":   cfg.xslbdm,
			"mzm":      cfg.mzm,
			"xz":       cfg.xz,
			"ccdm":     cfg.ccdm,
			"xsbj":     cfg.xsbj,
			"sfkknj":   cfg.sfkknj,
			"gnjkxdnj": cfg.gnjkxdnj,
			"sfkkzy":   cfg.sfkkzy,
			"kzybkxy":  cfg.kzybkxy,
			"sfznkx":   cfg.sfznkx,
			"zdkxms":   cfg.zdkxms,
			"sfkxq":    cfg.sfkxq,
			"sfkcfx":   cfg.sfkcfx,
			"kkbk":     cfg.kkbk,
			"kkbkdj":   cfg.kkbkdj,

			"xkxnm":    cfg.xkxnm,
			"xkxqm":    cfg.xkxqm,
			"kklxdm":   cfg.kklxdm,
			"bbhzxjxb": cfg.bbhzxjxb,
			"rlkz":     cfg.rlkz,

			"kch_id":  kch_id,
			"xkkz_id": cfg.xkkz_id,
		}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		req := a.http.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).
			SetFormData(formData).
			SetResult(&[]CourseDetail{})

		resp, err := req.Post(baseCfg.CHOOSE_COURSE_courseDetail)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(ctx.Err(), context.Canceled) {
				fmt.Println("请求已取消")
				return &[]CourseDetail{}
			} else if "0" == resp.String() {
				// json: cannot unmarshal string into Go value of type []client.CourseDetial
				//"0"
				fmt.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
				log.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
				// 选课数据重新获取，而不是原地重试请求
				cfg.needInit = true
				return nil
			}
			// fmt.Printf("请求出错: %v\n", err)
			log.Println("请求发生错误")
			continue
		}

		if a.LoginCheck(resp) {
			if !cfg.detailDump {
				log.Println(resp.String())
				cfg.detailDump = true
			}
			result := resp.Result().(*[]CourseDetail)
			return result
		} else {
			a.ReLogin()
			continue
		}

	}
}

func (a *APIClient) chooseCourseWithXXXXX(cfg *APIConfig, co *CustomCourseDic, sigCh chan os.Signal) *ChooseCourseResult {
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

func (a *APIClient) chooseCourseRaw(cfg *APIConfig, co *CustomCourseDic, ctx context.Context) *ChooseCourseResult {
	// 	选课
	// 	若flag==1则表示选课成功
	// 	已部分测试
	log.Println("=========chooseCourse()=========")
	var sxbj = "0"
	if cfg.rlkz == "1" || cfg.rlzlkz == "1" {
		sxbj = "1"
	}
	for {
		resp, err := a.http.R().
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).SetContext(ctx).
			SetFormData(map[string]string{
				"jxb_ids": co.Do_jxb_id,
				"kch_id":  co.Kch_id,

				"rwlx":    cfg.rwlx,
				"rlkz":    cfg.rlkz,
				"rlzlkz":  cfg.rlzlkz,
				"sxbj":    sxbj,
				"xxkbj":   co.Xxkbj,
				"qz":      "0",
				"cxbj":    co.Cxbj,
				"xkkz_id": cfg.xkkz_id,
				"njdm_id": cfg.njdm_id,
				"zyh_id":  cfg.zyh_id,
				"kklxdm":  cfg.kklxdm,
				"xklc":    cfg.xklc,
				"xkxnm":   cfg.xkxnm,
				"xkxqm":   cfg.xkxqm,
			}).
			SetResult(&ChooseCourseResult{}).
			Post(baseCfg.CHOOSE_COURSE_chooseCourse)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				fmt.Println("请求已取消")
				return &ChooseCourseResult{Flag: "-5"}
			}
			fmt.Println("选课请求发生错误", err)
			continue
		}

		if CheckStatusCode(resp) {
			log.Println("chooseCourse HTTP 错误: 状态码 ", resp.StatusCode())
			continue
		}
		if a.LoginCheck(resp) {
			fmt.Println(resp.String())
			log.Println("chooseCourse:", resp.String())
			result := resp.Result().(*ChooseCourseResult)
			//fmt.Println(result)
			return result
		} else {
			a.ReLogin()
			continue
		}
	}
}

func (a *APIClient) getHaveSelectedList(xkxnm, xkxqm string) *[]ChosenDic {
	// 查询已选课程
	fmt.Println("查询已选课程")
	for {
		resp, err := a.http.R().
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Account).
			SetFormData(map[string]string{
				"xkxnm": xkxnm,
				"xkxqm": xkxqm,
			}).
			SetResult(&[]ChosenDic{}).
			Post(baseCfg.CHOOSE_COURSE_SelectedList)
		if err != nil {
			fmt.Println("请求发生错误")
			continue
		}

		if CheckStatusCode(resp) {
			log.Println("getHaveChoosedList HTTP 错误: 状态码 ", resp.StatusCode())
			continue
		}
		if a.LoginCheck(resp) {
			// fmt.Println(resp.String())
			log.Printf("已选课程查询: \n%s", resp.String())
			result := resp.Result().(*[]ChosenDic)
			return result
		} else {
			a.ReLogin()
			continue
		}
	}
}

func (a *APIClient) quitCourse(jxb_ids string) (bool, string) {
	// 退课
	log.Println("========quitCourse()========")
	for range 3 {
		resp, err := a.http.R().
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Account).
			SetFormData(map[string]string{
				"jxb_ids": jxb_ids,
			}).
			Post(baseCfg.CHOOSE_COURSE_quitCourse)
		if err != nil {
			fmt.Println("退课请求发生错误")
			continue
		}

		if a.LoginCheck(resp) {
			log.Println(resp.String()) // "1"
			if resp.String() == "1" {
				fmt.Println("退课成功")
				return true, resp.String()
			} else {
				return false, resp.String()
			}
		} else {
			a.ReLogin()
			// continue
			return false, resp.String()
		}
	}
	return false, "??"
}

func (s *SafeCustomCourseSlice) courseList2custom(listP *[]CourseListDic) {
	// append_or_refersh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	list := *listP
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

func (s *SafeCustomCourseSlice) courseDetail2custom(detailP *[]CourseDetail) {
	// append_or_refresh
	s.mu.Lock()         // 加锁
	defer s.mu.Unlock() // 确保解锁
	list := *detailP
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

func scanWant(cfg *APIConfig, listP *[]CustomCourseDic) {
	list := *listP
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
	reference := guessGoodCourse(itemsCopy)
	scanWant(cfg, &itemsCopy)

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
				fmt.Printf("\033[1;36m%d\033[0m: %s\n", i, d.Jxbmc)
				fmt.Printf("\033[1;36m%d\033[0m: %s 人已选  %s  %s学分\n", i, d.Yxzrs, d.XF, d.Kzmc)
			} else {
				fmt.Printf("--------------------%d-------------------------------\n", i)
				if rs > reference {
					fmt.Printf("\033[1;36m%d\033[0m: %s\n", i, d.Jxbmc)
					fmt.Printf("\033[1;36m%d\033[0m: %s 人已选  %s  %s学分\n", i, d.Yxzrs, d.Kzmc, d.XF)
				} else {
					fmt.Printf("%d: %s\n", i, d.Jxbmc)
					fmt.Printf("%d: %s 人已选  %s  %s学分\n", i, d.Yxzrs, d.Kzmc, d.XF)
				}
			}
		} else {
			FullPrint(i, d)
		}
	}
	fmt.Println("====================end==============================")
}

func FullPrint(i int, d CustomCourseDic) {
	if d.Want {
		fmt.Printf("\033[0;33;40m-----👇--------%d----⬇-want-⬇---%d---------------------\033[0m\n", i, i)
		fmt.Printf("\033[1;36m%d\033[0m: %s %s %s  %s\n", i, d.Jxbmc, d.Xqumc, d.Sksj, d.Year)
		fmt.Printf("\033[1;36m%d\033[0m: ", i)
	} else {
		fmt.Printf("--------------%d---------------%d---------------------\n", i, i)
		fmt.Printf("%d: %s %s %s %s\n", i, d.Jxbmc, d.Xqumc, d.Sksj, d.Year)
		fmt.Printf("%d: ", i)
	}
	jxbrl, err := strconv.Atoi(d.Jxbrl)
	if err != nil {
		log.Println(err)
	}
	rs, err := strconv.Atoi(d.Yxzrs)
	if err != nil {
		log.Println(err)
	}
	if rs < jxbrl {
		fmt.Printf("%s \033[0;32;40m%s/%s\033[0m %s分 %s %s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
	} else {
		fmt.Printf("%s \033[0;31;40m%s/%s\033[0m %s分 %s %s\n", d.Jsxx, d.Yxzrs, d.Jxbrl, d.XF, d.Kzmc, d.Jxdd)
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
		return false, 0
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

func printSelectedList(selectedListP *[]ChosenDic) {
	selectedList := *selectedListP
	if len(selectedList) == 0 {
		fmt.Println("什么都没查到")
		return
	}
	for i, item := range selectedList {
		fmt.Println("---------------------------------------------------")
		if item.Sfktk == "1" {
			// 可以退课
			fmt.Printf("\033[1;36m%d\033[0m: \033[1;36m%s\033[0m\n", i, item.Jxbmc)
		} else {
			fmt.Printf("%d: %s\n", i, item.Jxbmc)
		}
	}
	fmt.Println("---------------------------------------------------")
}

func (a *APIClient) getAlreadySelectedTK(cfg *APIConfig) *[]ChosenDic {
	SelectedList := *a.getHaveSelectedList(cfg.xkxnm, cfg.xkxqm)
	var quitList []ChosenDic
	if len(SelectedList) == 0 {
		fmt.Println("没有可退课程")
		return nil
	}
	var first = true
	i := 0
	fmt.Println("---------------------目录--------------------------")
	for _, item := range SelectedList {
		if item.Sfktk == "1" {
			// 可以退课
			if first {
				first = false
			} else {
				fmt.Println("---------------------------------------------------")
			}
			fmt.Printf("\033[1;36m%d\033[0m: \033[1;36m%s\033[0m\n", i, item.Jxbmc)
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
	return &quitList
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
		var codeRow string
		// var code int
		fmt.Println("正在进行可退选课程查询...")
		quitList := *a.getAlreadySelectedTK(cfg)
		fmt.Print("请输入要退选的课程名字前的序号(-1退出 退课): ")
		_, err := fmt.Scanln(&codeRow)
		if err == io.EOF {
			wg.Wait()
			return
		}
		if err != nil {
			return
		}
		index, err1 := strconv.Atoi(strings.TrimSpace(codeRow))
		if err1 != nil {
			return
		}
		if 0 <= index && index < len(quitList) {
			fmt.Printf("退选课程: \033[1;36m%s\033[0m\n", quitList[index].Jxbmc)
			stat, msg := a.quitCourse(quitList[index].Do_jxb_id)
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
