package client

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log"
	"regexp"
	"school_sdk/check_code"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var loginWg sync.WaitGroup

func (a *APIClient) getPubParams(ctx context.Context, cfg *APIConfig, save bool) {
	log.Println("=======================get_pub_params()=======================")
	fmt.Println("Ctrl+C 退出")
	needEnter := false
	i := 0
	for {
		i++
		reqAt := time.Now()
		resp, err := a.hedgeC.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			SetFormData(map[string]string{
				"gnmkdm":  "N253512", // 这个接口能post
				"czdmKey": "00",
			}).
			Get(baseCfg.ChooseCourseIndex)

		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
				log.Println("index 请求已取消")
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("index 请求超时:", resp.Duration(), err)
				log.Println("index 请求超时:", resp.Duration(), err)
			} else {
				fmt.Println("index请求出错:", err)
				log.Println("index请求出错:", err)
				time.Sleep(time.Millisecond * 1500)
			}
			continue
		}

		if !a.CheckLogout302(resp) && utils.UserIsLogin(a.Config.Account, resp.String()) {
		} else {
			a.ReLogin()
			continue
		}
		if resp.IsStatusFailure() {
			fmt.Println("getPubParams:", resp.Status())
			continue
		}
		if len(resp.Bytes()) == 0 {
			log.Println("getPubParams len(resp.Bytes()) == 0", resp.Status())
			fmt.Println("getPubParams len(resp.Bytes()) == 0", resp.Status())
			time.Sleep(time.Millisecond * 1500)
			continue
		}
		docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes()))
		//docNode, err1 := htmlquery.Parse(resp.Body)
		if err1 != nil {
			log.Println(err1)
			fmt.Println("\r完成请求然后解析出错？")
			log.Println("完成请求然后解析出错？", err1)
			continue
		}
		if getXpathValue(docNode, "iskxk") == "0" {
			// 当前不属于选课阶段
			statNode := htmlquery.Find(docNode, `//div[@class="nodata"]/span/text()`)
			if len(statNode) != 0 {
				jdStr := strings.TrimSpace(htmlquery.InnerText(statNode[0]))
				// Sorry, it is not in the elective stage at present. If necessary, please contact the administrator.
				// 对不起，当前不属于选课阶段，如有需要，请与管理员联系！
				re := regexp.MustCompile(`ver=(\d+)`)
				match := re.FindStringSubmatch(resp.String())
				ver := ""
				if len(match) > 1 {
					ver = match[1]
				} else {
					log.Println("未找到 ver=数字")
				}
				fmt.Printf("\r%d %s ver=%s", i, jdStr, ver)
				log.Printf("%d %s ver=%s", i, jdStr, ver)
				needEnter = true
				time.Sleep(650 * time.Millisecond)
			}
			continue
		}

		if needEnter {
			fmt.Println()
			needEnter = false
		}
		if parseYzbIndexHtml(cfg, docNode) {
			log.Println("getPubParams:", resp.Status(), resp.String())
			continue
		}
		CheckTime(cfg.currentsj, reqAt)
		htmlContent := utils.RemoveEmptyLines(resp.String())
		if !save {
			log.Println(htmlContent)
		}
		a.Name = getXpathValue(docNode, "xm")
		if save {
			check_code.SaveFile("./", "zzxkyzb_cxZzxkYzbIndex.html", resp.Bytes())
		}
		return
	}
}

func parseYzbIndexHtml(cfg *APIConfig, docNode *html.Node) bool {
	cfg.xkkz_id = getXpathValue(docNode, "firstXkkzId")
	cfg.xkkz_xh = getXpathValue(docNode, "firstXkkzXh")
	cfg.kklxdm = getXpathValue(docNode, "firstKklxdm")
	cfg.Kklxmc = getXpathValue(docNode, "firstKklxmc")
	cfg.njdm_id = getXpathValue(docNode, "firstNjdmId")
	cfg.zyh_id = getXpathValue(docNode, "firstZyhId")
	//cfg.njdm_id_list0 = cfg.njdm_id
	cfg.bh_id = getXpathValue(docNode, "bh_id")   // 班号
	cfg.xkxnm = getXpathValue(docNode, "xkxnm")   // 学年
	cfg.xkxqm = getXpathValue(docNode, "xkxqm")   // 学期
	cfg.xqh_id = getXpathValue(docNode, "xqh_id") // 校区号
	cfg.jg_id = getXpathValue(docNode, "jg_id_1")
	cfg.xz = getXpathValue(docNode, "xz") // 学制
	cfg.zyfx_id = getXpathValue(docNode, "zyfx_id")
	cfg.ccdm = getXpathValue(docNode, "ccdm")
	cfg.xbm = getXpathValue(docNode, "xbm")   // 性别码 男1 女2
	cfg.mzm = getXpathValue(docNode, "mzm")   // 民族码
	cfg.xsbj = getXpathValue(docNode, "xsbj") // 学生标记
	cfg.xslbdm = getXpathValue(docNode, "xslbdm")
	cfg.xszxzt = getXpathValue(docNode, "xszxzt")
	cfg.xxdm = getXpathValue(docNode, "xxdm") // 学校代码
	cfg.zxfs = getXpathValue(docNode, "zxfs")
	cfg.tkzgcs_qt = getXpathValue(docNode, "tkzgcs_qt")
	cfg.currentsj = getXpathValue(docNode, "currentsj") // 当前时间

	// 解析总学分最低
	minCreditNode := htmlquery.FindOne(docNode, `//h5[contains(., "总学分最低")]/font[@color='red'][1]`)
	if minCreditNode != nil {
		cfg.minCredit = htmlquery.InnerText(minCreditNode)
	}
	// 解析总学分最高
	maxCreditNode := htmlquery.FindOne(docNode, `//h5[contains(., "总学分最低")]/font[@color='red'][2]`)
	if maxCreditNode != nil {
		cfg.maxCredit = htmlquery.InnerText(maxCreditNode)
	}
	// 解析本学期已选学分
	selectedCreditNode := htmlquery.FindOne(docNode, `//font[@id='yxxfs']`)
	if selectedCreditNode != nil {
		cfg.selectedCredit = htmlquery.InnerText(selectedCreditNode)
	}
	if cfg.xkkz_id == "" && cfg.kklxdm == "" && cfg.xkkz_xh == "" && cfg.xkxnm == "" {
		fmt.Println("❌ Step 1 index failed")
		return true
	}
	if cfg.xkkz_xh != "" {
		fmt.Println("正方 V9")
	}
	fmt.Println("Step 1 index finished")
	fmt.Println("\n\r将要选 \033[1;36m", cfg.Kklxmc, "\033[0m !!")
	log.Println("将要选", cfg.Kklxmc, "!!")
	parseKklxdmXkkz__(cfg, docNode)
	cfg.modeName = cfg.Kklxmc
	return false
}

func (a *APIClient) getCourseListPre(ctx context.Context, cfg *APIConfig, save bool) {
	// 补齐搜索课程需要的发包参数 页签
	log.Println("===============getCourseList_pre()=================")
	formData := map[string]string{
		"xszxzt": cfg.xszxzt, // 1
		"kspage": "0",
		"jspage": "0",
	}
	if cfg.xkkz_id != "" {
		formData["xkkz_id"] = cfg.xkkz_id
	}
	if cfg.xkkz_xh != "" {
		formData["xkkz_xh"] = cfg.xkkz_xh
	}
	for {
		resp, err := a.hedgeC.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			SetFormData(formData).
			Post(baseCfg.ChooseCourseListPre)

		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				fmt.Println("YzbDisplay 请求已取消")
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("YzbDisplay 请求超时:", resp.Duration())
				log.Println("YzbDisplay 请求超时:", resp.Duration(), err)
			} else {
				fmt.Println("YzbDisplay 请求出错:", err)
				log.Println("YzbDisplay 请求出错:", err)
				time.Sleep(time.Millisecond * 500)
			}
			continue
		}
		if resp.IsStatusFailure() {
			// 这里容易有 404 问题和
			// 200 错误提示 系统运行异常，请稍后再试 问题
			fmt.Println("ListPre:", resp.Status(), resp.String())
			continue
		}
		if a.LoginCheck(resp) {
		} else {
			a.ReLogin()
			continue
		}
		if resp.IsStatusSuccess() {
			docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes()))
			//docNode, err1 := htmlquery.Parse(strings.NewReader(htmlContent))
			//docNode, err1 := htmlquery.Parse(resp.Body)
			if err1 != nil {
				log.Println("htmlquery:", err1)
				fmt.Println("完成请求然后解析出错？")
				continue
			}
			parseListPreHtml(cfg, docNode)
			htmlContent := utils.RemoveEmptyLines(resp.String())
			if !save {
				log.Println(htmlContent)
			}
			if save {
				check_code.SaveFile("./", "zzxkyzb_cxZzxkYzbDisplay.html", resp.Bytes())
			}
			return
		}
		fmt.Println("ListPre:", resp.Status(), resp.String())
	}
	// zdzys //"一门课程最多可选"+zdzys+"个志愿！"
	// sfqzxk //"一门课程只能选一个教学班！"
	// self.lnzgxkxf # 历学期选课最高学分要求为
	// bxqzgxkxf //基本选课规则设置中设置的最高选课学分
	// 本学期本类型课程选课最高学分要求为"+bxqzgxkxf+"，当前本学期本类型课程选课总学分为("+kklxzxfs+"+"+$("#xf_"+kch_id).text()+"="+c_kklxzxfs+")，超出选课最高学分要求，不可选！"
	// 本学期本类型课程选课最高门次要求为 bxqzgxkmc ，不可选
}

func getXpathValue(docNode *html.Node, name string) string {
	if docNode == nil || name == "" {
		return ""
	}
	escapedName := html.EscapeString(name)
	// 使用单个XPath同时匹配name或id
	xpath := `//*[@id="` + escapedName + `" or @name="` + escapedName + `"]`
	if node := htmlquery.FindOne(docNode, xpath); node != nil {
		return htmlquery.SelectAttr(node, "value")
	}
	return ""
}

func parseListPreHtml(cfg *APIConfig, docNode *html.Node) {
	cfg.bklx_id = getXpathValue(docNode, "bklx_id")
	cfg.rwlx = getXpathValue(docNode, "rwlx")
	cfg.xkly = getXpathValue(docNode, "xkly")
	cfg.xklc = getXpathValue(docNode, "xklc")     // 选课轮次
	cfg.xklcmc = getXpathValue(docNode, "xklcmc") // 选课轮次
	cfg.sfkknj = getXpathValue(docNode, "sfkknj")
	cfg.rlkz = getXpathValue(docNode, "rlkz")
	cfg.cdrlkz = getXpathValue(docNode, "cdrlkz")
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
	cfg.rlzlkz = getXpathValue(docNode, "rlzlkz")
	cfg.sfkxk = getXpathValue(docNode, "sfkxk")   // 是否可选课
	cfg.sfktk = getXpathValue(docNode, "sfktk")   // 是否可退课
	cfg.jdlx = getXpathValue(docNode, "jdlx")     // 体育课多志愿开关
	cfg.syts = getXpathValue(docNode, "syts")     // 距选课结束还剩{0}天
	cfg.syxs = getXpathValue(docNode, "syxs")     // 距选课结束还剩{0}小时
	cfg.xkkssj = getXpathValue(docNode, "xkkssj") // 选课开始时间
	cfg.xkjssj = getXpathValue(docNode, "xkjssj") // 选课结束时间
	if cfg.bklx_id == "" && cfg.rwlx == "" && cfg.xkly == "" && cfg.syxs == "" {
		fmt.Println("❌ Step 2 params failed")
		return
	}
	fmt.Println("Step 2 params finished")
}

func (a *APIClient) getCourseList(ctx context.Context, cfg *APIConfig) []CourseListDic {
	// 搜索课程，主页面的查询
	if !cfg.listDump {
		log.Println("========搜索课程 getCourseList()========")
		fmt.Println("搜索课程")
	}
	var result GetCourseListResult
	for {
		formData := map[string]string{ // 25
			"bbhzxjxb": cfg.bbhzxjxb,
			"bh_id":    cfg.bh_id,
			"bklx_id":  cfg.bklx_id,
			//"xkkz_id":      cfg.xkkz_id, //
			"rwlx":         cfg.rwlx, // 校选是2 专选是1 没有这两个会蹦出来选不了的课，主修课：✓ 选修课：✗
			"xkly":         cfg.xkly, // 1 选择无限制是0，主修课：✓ 选修课：✗
			"sfkkjyxdxnxq": cfg.sfkkjyxdxnxq,
			"xqh_id":       cfg.xqh_id,
			"jg_id":        cfg.jg_id,
			//"zyh_id_1":     cfg.zyh_id,
			"zyh_id":    cfg.zyh_id,
			"zyh_id_xs": cfg.zyh_id,
			//"zyfx_id":      cfg.zyfx_id,
			"njdm_id":    cfg.njdm_id,
			"njdm_id_xs": cfg.njdm_id,
			//"njdm_id_1":    cfg.njdm_id,
			// bjgkczxbbjwcx: 0
			"xbm": cfg.xbm,
			//"xslbdm":   cfg.xslbdm,
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
			// 	sfkgbcx: 0
			// 	sfrxtgkcxd: 0
			// 	tykczgxdcs: 0
			"xkxnm":  cfg.xkxnm,  // 当前学期年份, 如2021-2022 即2021，必须
			"xkxqm":  cfg.xkxqm,  // 3 12 16
			"kklxdm": cfg.kklxdm, // 01为主修课 10为选修课，校选10 专选01，英语进阶06，必须
			"rlkz":   cfg.rlkz,
			// 	xkzgbj: 0
			"kspage": "1",
			"jspage": "222", // 页号，一页显示的数量，必须
			//"jxbzb":  "",
		}
		if cfg.xkkz_id != "" {
			formData["xkkz_id"] = cfg.xkkz_id
		}
		if cfg.xkkz_xh != "" {
			formData["xkkz_xh"] = cfg.xkkz_xh
		}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		if cfg.yl {
			formData["yl_list[0]"] = "1"
		}
		requ := a.hedgeC.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				//"su":     a.Config.Account,
			}).
			SetFormData(formData)
		//SetResult(&result).SetError(&respStr)

		resp, err := requ.Post(baseCfg.ChooseCourseCourseList)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				log.Println("PartD 请求已取消")
				return nil
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("PartD 请求超时", resp.Duration())
				log.Println("PartD 请求超时", resp.Duration())
				continue
			}

			fmt.Println("PartD 请求发生错误:", err)
			log.Println("PartD 请求发生错误:", err, resp.String())
			time.Sleep(370 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if a.LoginCheck(resp) {
		} else {
			// fmt.Println("重新登录")
			a.ReLogin()
			continue
		}
		if resp.IsStatusSuccess() {
			if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
				if resp.String() == `"0"` || resp.String() == `"1"` {
					fmt.Println("搜索课程失败 msg:", resp.String())
					log.Println("搜索课程失败 msg:", resp.String())
					cfg.needInit = true
					return nil
				}
				fmt.Println(err)
				log.Println(err, resp.String())
			}
			if !cfg.listDump {
				log.Println(resp.String())
				cfg.listDump = true
			}
			if result.TmpList != nil {
				return result.TmpList
			}
			// {"msg":"加密串错误，可以清除浏览器缓存后刷新网页重试！","flag":"0"}
			if result.Msg != "" || result.Flag != "" {
				log.Println(resp.String())
				fmt.Println(resp.String())
				cfg.needInit = true
				return result.TmpList
			}
			fmt.Println("课程列表为空", len(result.TmpList))
			time.Sleep(1 * time.Second)
			continue
		}

	}
}

func (a *APIClient) getCourseDetail(ctx context.Context, cfg *APIConfig, kch_id string) []CourseDetail {
	// 	查询课程具体信息
	// 	获得do_jxb_id（选课的必要参数）以及很重要的容量信息等等
	if !cfg.detailDump {
		fmt.Println("\r正在获取详细信息")
		log.Println("========查询课程具体信息 getCourseDetail()========")
	}
	type MsgFlag struct {
		Msg  string `json:"msg"`
		Flag string `json:"flag"`
	}
	var result []CourseDetail
	for {
		formData := map[string]string{
			"rwlx":         cfg.rwlx, // 校选是2 专选是1 没有这两个会蹦出来选不了的课，主修课：✓ 选修课：✗
			"xkly":         cfg.xkly, // 1 选择无限制是0，主修课：✓ 选修课：✗
			"bklx_id":      cfg.bklx_id,
			"sfkkjyxdxnxq": cfg.sfkkjyxdxnxq,

			"xqh_id":    cfg.xqh_id,
			"jg_id":     cfg.jg_id,
			"zyh_id":    cfg.zyh_id,
			"zyh_id_xs": cfg.zyh_id,
			//"zyfx_id":   cfg.zyfx_id,

			//"njdm_id_1":  cfg.njdm_id,
			"njdm_id":    cfg.njdm_id,
			"njdm_id_xs": cfg.njdm_id,
			"bh_id":      cfg.bh_id,
			// bjgkczxbbjwcx: 0
			"xbm": cfg.xbm,
			//"xslbdm":   cfg.xslbdm,
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
			"bbhzxjxb": cfg.bbhzxjxb,
			"kkbk":     cfg.kkbk,
			"kkbkdj":   cfg.kkbkdj,

			// 	sfkgbcx: 0
			// 	sfrxtgkcxd: 0
			// 	tykczgxdcs: 0
			"xkxnm": cfg.xkxnm, // 当前学期年份, 如2021-2022 即2021，必须
			"xkxqm": cfg.xkxqm, // 3 12 16

			"rlkz":   cfg.rlkz,
			"kklxdm": cfg.kklxdm, // 01为主修课 10为选修课，校选10 专选01，英语进阶06，必须

			// 	xkzgbj: 0
			"kch_id": kch_id, // 课程号，必须
			"xklc":   cfg.xklc,
			//"xkkz_id": cfg.xkkz_id,
		}
		if cfg.xkkz_id != "" {
			formData["xkkz_id"] = cfg.xkkz_id
		} else {
			formData["xkkz_xh"] = cfg.xkkz_xh
		}
		//if cfg.xkkz_xh != "" {
		//	formData["xkkz_xh"] = cfg.xkkz_xh
		//}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		if cfg.yl {
			formData["yl_list[0]"] = "1"
		}
		requ := a.hedgeC.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				//"su":     a.Config.Account,
			}).
			SetFormData(formData)

		loginWg.Wait()
		resp, err := requ.Post(baseCfg.ChooseCourseCourseDetail)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				fmt.Println("WithKch 请求已取消")
				return result
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("WithKch 请求超时", resp.Duration())
				continue
			}
			fmt.Println("WithKch 请求发生错误:", err)
			log.Println("WithKch 请求发生错误:", err, resp.String())
			time.Sleep(370 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if a.LoginCheck(resp) {
			if !cfg.detailDump {
				log.Println(resp.String())
				cfg.detailDump = true
			}
			if result != nil {
				return result
			}
		} else {
			loginWg.Add(1)
			a.ReLogin()
			loginWg.Done()
			continue
		}
		if resp.IsStatusSuccess() {
			if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
				if resp.String() == `"0"` {
					fmt.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
					log.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
					cfg.needInit = true
					return nil
				}
				fmt.Println(err, resp.String())
				log.Println(err, resp.String())
				var mf MsgFlag
				if err := json.Unmarshal(resp.Bytes(), &mf); err != nil {
					cfg.needInit = true
					log.Println("getCourseDetail:", resp.String())
					fmt.Println(resp.String())
					return nil
				}
				log.Println("getCourseDetail?:", resp.String())
				fmt.Println(resp.String())
				return nil
			}
		}

	}
}

func (a *APIClient) getCourseDoJxb(ctx context.Context, cfg *APIConfig, jxb_ids []string) []CourseDetail {
	// 貌似只有实验课能用这个
	var result []CourseDetail
	for {
		formData := map[string]string{
			"jxb_ids": strings.Join(jxb_ids, ","),
			"bklx_id": cfg.bklx_id,
			"kklxdm":  cfg.kklxdm,
			"rlkz":    cfg.rlkz,
			"xklc":    cfg.xklc,
			"zyh_id":  cfg.zyh_id,
			"njdm_id": cfg.njdm_id,
		}
		if cfg.xkkz_id != "" {
			formData["xkkz_id"] = cfg.xkkz_id
		} else {
			formData["xkkz_xh"] = cfg.xkkz_xh
		}
		//if cfg.xkkz_xh != "" {
		//	formData["xkkz_xh"] = cfg.xkkz_xh
		//}
		requ := a.hedgeC.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Config.Account,
			}).
			SetFormData(formData)

		loginWg.Wait()
		resp, err := requ.Post(baseCfg.ChooseSimpleDoJxb)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				fmt.Println("请求已取消")
				return result
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时", resp.Duration())
				continue
			}
			fmt.Println("请求发生错误:", err)
			log.Println("请求发生错误:", err, resp.String())
			time.Sleep(370 * time.Millisecond)
			continue
		}

		if resp.IsStatusFailure() {
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if a.LoginCheck(resp) {
			if !cfg.detailDump {
				log.Println(resp.String())
				cfg.detailDump = true
			}
			if result != nil {
				return result
			}
			//log.Println(resp.String())
			//continue
		} else {
			loginWg.Add(1)
			a.ReLogin()
			loginWg.Done()
			continue
		}
		if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
			if resp.String() == `"0"` {
				fmt.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
				log.Println(`"0"，未查询到信息，可能没到选课时间，可能程序编写错误，也可能教务系统临时调整了选课`)
				cfg.needInit = true
				return nil
			}
			fmt.Println(err)
			log.Println(err, resp.String())
		}
	}
}

func (a *APIClient) chooseCourseRaw(cfg *APIConfig, co *CustomCourseDic, ctx context.Context) ChooseCourseResult {
	// 	选课
	// 	若flag==1则表示选课成功
	var sxbj = "0"
	if cfg.rlkz == "1" || cfg.rlzlkz == "1" || cfg.cdrlkz == "1" {
		sxbj = "1"
	}
	var result ChooseCourseResult
	for {
		formData := map[string]string{
			// "bklx_id": cfg.bklx_id,  // 英语进阶，这一个能顶掉很多个
			// 选课第一阶段 不允许跨年级跨专业选课 选课第二阶段 允许跨年级跨专业选课 不带下面的参数也可以
			"jxb_ids": co.Do_jxb_id,
			"kch_id":  co.Kch_id,
			// "kcmc":    co.kcmc,
			"rwlx":   cfg.rwlx,
			"rlkz":   cfg.rlkz,
			"cdrlkz": cfg.cdrlkz,
			"rlzlkz": cfg.rlzlkz,
			"sxbj":   sxbj,
			"xxkbj":  co.Xxkbj,
			"qz":     "0",
			"cxbj":   co.Cxbj,
			//"xkkz_id":    cfg.xkkz_id,
			"njdm_id":    cfg.njdm_id,
			"njdm_id_xs": cfg.njdm_id,
			"zyh_id":     cfg.zyh_id,
			"zyh_id_xs":  cfg.zyh_id,
			"kklxdm":     cfg.kklxdm, // 校选10 专选01 英语06
			"xklc":       cfg.xklc,   // 选课轮次
			"xkxnm":      cfg.xkxnm,
			"xkxqm":      cfg.xkxqm,
			//"jcxx_id":    "[]jcxx_arr",
		}
		if cfg.xkkz_id != "" {
			formData["xkkz_id"] = cfg.xkkz_id
		} else {
			formData["xkkz_xh"] = cfg.xkkz_xh
		}
		//if cfg.xkkz_xh != "" {
		//	formData["xkkz_xh"] = cfg.xkkz_xh
		//}
		loginWg.Wait()
		resp, err := a.hedgeC.R().
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Config.Account,
			}).SetContext(ctx).
			SetFormData(formData).
			//SetResult(&result).
			Post(baseCfg.ChooseCourse)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("选课请求已取消")
				return ChooseCourseResult{Flag: "-5"}
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("选课请求超时", resp.Duration())
				log.Println("选课请求超时", resp.Duration())
			} else {
				fmt.Println("选课请求发生错误", err)
				time.Sleep(100 * time.Millisecond)
			}
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("chooseCourse", resp.Status())
			log.Println("chooseCourse 状态码:", resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if a.LoginCheck(resp) {
		} else {
			loginWg.Add(1)
			a.ReLogin()
			loginWg.Done()
			continue
		}
		if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
			fmt.Println(err, resp.String())
			log.Println(err, resp.String())
			continue
		}
		return result
	}
}

func (a *APIClient) isCourseRegistered(cfg *APIConfig, co *CustomCourseDic) bool {
	log.Println("====isCourseRegistered====")
	for {
		formData := map[string]string{
			"jxb_id": co.Do_jxb_id,
			"xnm":    cfg.xkxnm,
			"xqm":    cfg.xkxqm,
		}
		if cfg.xkkz_id != "" {
			formData["xkkz_id"] = cfg.xkkz_id
		}
		if cfg.xkkz_xh != "" {
			formData["xkkz_xh"] = cfg.xkkz_xh
		}
		resp, err := a.hedgeC.R().
			SetTimeout(time.Second*19).
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Config.Account).
			SetFormData(formData).
			Post(baseCfg.CourseRegistered)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("Reg请求超时", resp.Duration())
				continue
			} else {
				fmt.Println("Reg请求发生错误")
				log.Println(err)
				time.Sleep(150 * time.Millisecond)
			}
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("isCourseRegistered:", resp.Status())
			log.Println("isCourseRegistered 状态码:", resp.Status())
			continue
		}
		if resp.ResultError() != nil {
			log.Println(resp.ResultError(), resp.String())
			continue
		}
		if a.LoginCheck(resp) {
			return resp.String() == "1"
		}

		a.ReLogin()
		continue
	}
}

func (a *APIClient) getHaveSelectedList(xkxnm, xkxqm string) []ChosenDic {
	// 查询已选课程
	fmt.Println("查询已选课程")
	var result []ChosenDic
	for {
		resp, err := a.hedgeC.R().
			SetTimeout(time.Second*23).
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Config.Account).
			SetFormData(map[string]string{
				"xkxnm": xkxnm,
				"xkxqm": xkxqm,
			}).
			//SetResult(&result).
			Post(baseCfg.CourseSelectedList)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("已选课程请求超时", resp.Duration())
				return nil
			}

			fmt.Println("已选课程请求错误:", err)
			log.Println("已选课程请求错误:", err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("Choosed:", resp.Status())
			log.Println("getHaveChoosedList 状态码:", resp.Status())
			continue
		}
		if a.LoginCheck(resp) {
		} else {
			a.ReLogin()
			continue
		}

		if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
			fmt.Println(err, resp.String())
			log.Println(err, resp.String())
			continue
		}
		log.Printf("已选课程查询: \n%s", resp.String())
		return result
	}
}

func (a *APIClient) quitCourse(cfg *APIConfig, jxb_ids, kch_id string) (bool, string) {
	// 退课
	log.Println("========quitCourse()========")
	for range 3 {
		resp, err := a.hedgeC.R().
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Config.Account).
			SetFormData(map[string]string{
				"kch_id":  kch_id,
				"jxb_ids": jxb_ids,
				"xkxnm":   cfg.xkxnm,
				"xkxqm":   cfg.xkxqm,
				"txbsfrl": "0",
			}).
			Post(baseCfg.QuitCourse)
		if err != nil {
			fmt.Println("退课请求发生错误")
			log.Println(err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("quitCourse", resp.Status())
			fmt.Println("quitCourse", resp.Status())
		}
		if a.LoginCheck(resp) {
			log.Println(resp.String()) // "1"
			if resp.String() == `"1"` {
				//fmt.Println("退课成功")
				return true, resp.String()
			} else if resp.String() == `"2"` {
				// 服务器繁忙
				return false, resp.String()
			} else if resp.String() == `"3"` {
				// 未知错误
				return false, resp.String()
			} else if resp.String() == `"4"` {
				// 非法访问
				return false, resp.String()
			} else if resp.String() == `"5"` {
				// 验证失败
				return false, resp.String()
			}
			return false, resp.String()
		}

		a.ReLogin()
		// continue
		return false, resp.String()
	}
	return false, "??"
}
