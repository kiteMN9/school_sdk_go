package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var loginWg sync.WaitGroup

func (a *APIClient) getPubParams(ctx context.Context, cfg *APIConfig) {
	log.Println("=======================get_pub_params()=======================")
	needEnter := false
	i := 0
	for {
		i++
		resp, err := a.Http.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			Get(baseCfg.ChooseCourseIndex)

		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				log.Println("请求已取消")
				return
			}
			fmt.Println("请求出错: ", err)
			log.Println("请求发生错误:", err)
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
		//docNode, err1 := htmlquery.Parse(strings.NewReader(htmlContent)) // 相当于etree.HTML()
		docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes())) // 相当于etree.HTML()
		//docNode, err1 := htmlquery.Parse(resp.Body) // 相当于etree.HTML()
		//resp.Body.Close()
		if err1 != nil {
			log.Println(err1)
			fmt.Println("\r他妈个逼这什么情况，完成请求然后解析出错？")
			log.Println("他妈个逼这什么情况，完成请求然后解析出错？", err1)
			continue
		}
		if getXpathValue(docNode, "iskxk") == "0" {
			//fmt.Printf("\r%d 当前还不能选课", i)
			// 当前不属于选课阶段
			statNode := htmlquery.Find(docNode, `//div[@class="nodata"]/span/text()`)
			if len(statNode) != 0 {
				jdStr := strings.TrimSpace(htmlquery.InnerText(statNode[0]))
				// Sorry, it is not in the elective stage at present. If necessary, please contact the administrator.
				// 对不起，当前不属于选课阶段，如有需要，请与管理员联系！
				fmt.Printf("\r%d %s", i, jdStr)
				log.Printf("%d %s", i, jdStr)
				needEnter = true
				time.Sleep(650 * time.Millisecond)
			}
			continue
		}

		if needEnter {
			fmt.Println()
			needEnter = false
		}
		parseYzbIndexHtml(cfg, docNode)
		fmt.Println("\r将要选 \033[1;36m", cfg.Kklxmc, "\033[0m !!")
		log.Println("\r将要选", cfg.Kklxmc, "!!")
		cfg.modeName = cfg.Kklxmc
		log.Println(htmlContent)
		parseKklxdmXkkzId(cfg, docNode)
		return
	}
}

func parseYzbIndexHtml(cfg *APIConfig, docNode *html.Node) {
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
	cfg.xbm = getXpathValue(docNode, "xbm") // 性别码 男1 女2
	cfg.mzm = getXpathValue(docNode, "mzm")
	cfg.xsbj = getXpathValue(docNode, "xsbj")
	cfg.xslbdm = getXpathValue(docNode, "xslbdm")
	cfg.xszxzt = getXpathValue(docNode, "xszxzt")
	cfg.xxdm = getXpathValue(docNode, "xxdm")
	cfg.zxfs = getXpathValue(docNode, "zxfs")
	cfg.tkzgcs_qt = getXpathValue(docNode, "tkzgcs_qt")
	cfg.currentsj = getXpathValue(docNode, "currentsj")
	fmt.Println("index finished")
}

func (a *APIClient) getCourseListPre(ctx context.Context, cfg *APIConfig, xkkz_id, xszxzt string) {
	// 补 齐搜索课程需要的发包参数
	log.Println("===============getCourseList_pre()=================")
	for {
		resp, err := a.Http.R().
			SetContext(ctx).
			SetQueryParam("gnmkdm", "N253512").
			SetFormData(map[string]string{
				"xkkz_id": xkkz_id,
				"xszxzt":  xszxzt, // 1
				"kspage":  "0",
				"jspage":  "0",
			}).
			Post(baseCfg.ChooseCourseListPre)

		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				fmt.Println("请求已取消")
				return
			}
			fmt.Println("请求出错: ", err)
			log.Println("请求发生错误", err)
			continue
		}
		if resp.IsError() {
			// 这里容易有 404 问题和
			// 200 错误提示 系统运行异常，请稍后再试 问题
			fmt.Println("ListPre:", resp.Status(), resp.String())
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
		docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes())) // 相当于etree.HTML()
		//docNode, err1 := htmlquery.Parse(strings.NewReader(htmlContent)) //etree.HTML()
		//docNode, err1 := htmlquery.Parse(resp.Body) //etree.HTML()
		if err1 != nil {
			log.Println("htmlquery:", err1)
			fmt.Println("他妈个逼这什么情况，完成请求然后解析出错？")
			continue
		}
		parseListPreHtml(cfg, docNode)
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
	cfg.xklc = getXpathValue(docNode, "xklc")
	cfg.rlzlkz = getXpathValue(docNode, "rlzlkz")
	cfg.sfkxk = getXpathValue(docNode, "sfkxk") // 是否可选课
	cfg.sfktk = getXpathValue(docNode, "sfktk") // 是否可退课
	cfg.jdlx = getXpathValue(docNode, "jdlx")   // 体育课多志愿开关
	cfg.syts = getXpathValue(docNode, "syts")   // 距选课结束还剩{0}天
	cfg.syxs = getXpathValue(docNode, "syxs")   // 距选课结束还剩{0}小时
	fmt.Println("params finished")
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
			"bbhzxjxb":     cfg.bbhzxjxb,
			"bh_id":        cfg.bh_id,
			"bklx_id":      cfg.bklx_id,
			"xkkz_id":      cfg.xkkz_id,
			"rwlx":         cfg.rwlx, // 校选是2 专选是1 没有这两个会蹦出来选不了的课，主修课：✓ 选修课：✗
			"xkly":         cfg.xkly, // 1 选择无限制是0，主修课：✓ 选修课：✗
			"sfkkjyxdxnxq": cfg.sfkkjyxdxnxq,
			"xqh_id":       cfg.xqh_id,
			"jg_id":        cfg.jg_id,
			"zyh_id_1":     cfg.zyh_id,
			"zyh_id":       cfg.zyh_id,
			"zyh_id_xs":    cfg.zyh_id,
			"zyfx_id":      cfg.zyfx_id,
			"njdm_id":      cfg.njdm_id,
			"njdm_id_xs":   cfg.njdm_id,
			"njdm_id_1":    cfg.njdm_id,
			// bjgkczxbbjwcx: 0
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
			// 	sfkgbcx: 0
			// 	sfrxtgkcxd: 0
			// 	tykczgxdcs: 0
			"xkxnm":  cfg.xkxnm,  // 当前学期年份, 如2021-2022 即2021，必须
			"xkxqm":  cfg.xkxqm,  // 3 12 16
			"kklxdm": cfg.kklxdm, // 01为主修课 10为选修课，校选10 专选01，英语进阶06，必须
			"rlkz":   cfg.rlkz,
			// 	xkzgbj: 0
			"kspage": "1",
			"jspage": "200", // 页号，一页显示的数量，必须
			"jxbzb":  "",
		}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		if cfg.yl {
			formData["yl_list[0]"] = "1"
		}
		requ := a.Http.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).
			SetFormData(formData)
		//SetResult(&result).SetError(&respStr)

		resp, err := requ.Post(baseCfg.ChooseCourseCourseList)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				log.Println("请求已取消")
				return nil
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时")
				log.Println("请求超时")
				continue
			}

			fmt.Println("请求发生错误:", err)
			log.Println("请求发生错误:", err, resp.String())
			time.Sleep(370 * time.Millisecond)
			continue
		}
		if resp.IsError() {
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}

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
		//resp.IsSuccess()
		if a.LoginCheck(resp) {
			if !cfg.listDump {
				log.Println(resp.String())
				cfg.listDump = true
			}
			if result.TmpList != nil {
				return result.TmpList
			}

			//fmt.Println(resp.String())
			fmt.Println("课程列表为空")
			time.Sleep(1 * time.Second)
			continue
		} else {
			// fmt.Println("重新登录")
			a.ReLogin()
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
			"zyfx_id":   cfg.zyfx_id,

			"njdm_id_1":  cfg.njdm_id,
			"njdm_id":    cfg.njdm_id,
			"njdm_id_xs": cfg.njdm_id,
			"bh_id":      cfg.bh_id,
			// bjgkczxbbjwcx: 0
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
			"kch_id":  kch_id, // 课程号，必须
			"xklc":    cfg.xklc,
			"xkkz_id": cfg.xkkz_id,
		}
		if cfg.njdm_id_list0 != "" {
			formData["njdm_id_list[0]"] = cfg.njdm_id_list0 // 这个就是选课的时候筛选的条件，建议只填个年级就好了
		}
		if cfg.yl {
			formData["yl_list[0]"] = "1"
		}
		requ := a.Http.R().
			SetContext(ctx).
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).
			SetFormData(formData)

		loginWg.Wait()
		resp, err := requ.Post(baseCfg.ChooseCourseCourseDetail)
		if err != nil {
			// 判断是否因Context取消导致的错误
			if errors.Is(err, context.Canceled) {
				fmt.Println("请求已取消")
				return result
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时")
				continue
			}
			fmt.Println("请求发生错误:", err)
			log.Println("请求发生错误:", err, resp.String())
			time.Sleep(370 * time.Millisecond)
			continue
		}

		if resp.IsError() {
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
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
	}
}

func (a *APIClient) chooseCourseRaw(cfg *APIConfig, co *CustomCourseDic, ctx context.Context) ChooseCourseResult {
	// 	选课
	// 	若flag==1则表示选课成功
	// 	已部分测试
	//log.Println("=========chooseCourse()=========")
	var sxbj = "0"
	if cfg.rlkz == "1" || cfg.rlzlkz == "1" || cfg.cdrlkz == "1" {
		sxbj = "1"
	}
	var result ChooseCourseResult
	for {
		data := map[string]string{
			// "bklx_id": cfg.bklx_id,  // 英语进阶，这一个能顶掉很多个
			// 选课第一阶段 不允许跨年级跨专业选课 选课第二阶段 允许跨年级跨专业选课 不带下面的参数也可以
			"jxb_ids": co.Do_jxb_id, // 有些开源代码这边要拼接不知道是什么蜜汁操作 jxb_arr
			"kch_id":  co.Kch_id,
			// "kcmc":    co.kcmc,
			"rwlx":       cfg.rwlx,
			"rlkz":       cfg.rlkz,
			"cdrlkz":     cfg.cdrlkz,
			"rlzlkz":     cfg.rlzlkz,
			"sxbj":       sxbj,
			"xxkbj":      co.Xxkbj,
			"qz":         "0",
			"cxbj":       co.Cxbj,
			"xkkz_id":    cfg.xkkz_id,
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
		loginWg.Wait()
		resp, err := a.Http.R().
			SetQueryParams(map[string]string{
				"gnmkdm": "N253512",
				"su":     a.Account,
			}).SetContext(ctx).
			SetFormData(data).
			SetResult(&result).
			Post(baseCfg.ChooseCourse)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Println("请求已取消")
				return ChooseCourseResult{Flag: "-5"}
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时")
				log.Println("请求超时")
			} else {
				fmt.Println("选课请求发生错误", err)
			}
			continue
		}
		if resp.IsError() {
			log.Println("chooseCourse", resp.Status())
			log.Println("chooseCourse 状态码:", resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.Error() != nil {
			log.Println(resp.Error(), resp.String())
			continue
		}
		if a.LoginCheck(resp) {
			//log.Println("chooseCourse:", result.Flag, result.Msg)
			return result
		} else {
			loginWg.Add(1)
			a.ReLogin()
			loginWg.Done()
			continue
		}
	}
}

func (a *APIClient) getHaveSelectedList(xkxnm, xkxqm string) []ChosenDic {
	// 查询已选课程
	fmt.Println("查询已选课程")
	var result []ChosenDic
	for {
		resp, err := a.Http.R().
			SetTimeout(time.Second*23).
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Account).
			SetFormData(map[string]string{
				"xkxnm": xkxnm,
				"xkxqm": xkxqm,
			}).
			SetResult(&result).
			Post(baseCfg.CourseSelectedList)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("请求超时")
				return result
			} else {
				fmt.Println("请求发生错误")
				log.Println(err)
			}
			continue
		}
		if resp.IsError() {
			log.Println("Choosed:", resp.Status())
			log.Println("getHaveChoosedList 状态码:", resp.Status())
			continue
		}
		if resp.Error() != nil {
			log.Println(resp.Error(), resp.String())
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

func (a *APIClient) quitCourse(cfg *APIConfig, jxb_ids, kch_id string) (bool, string) {
	// 退课
	// fmt.Println("========quitCourse()========")
	log.Println("========quitCourse()========")
	for range 3 {
		resp, err := a.Http.R().
			SetQueryParam("gnmkdm", "N253512").
			SetQueryParam("su", a.Account).
			SetFormData(map[string]string{
				"kch_id":  kch_id,
				"jxb_ids": jxb_ids,
				//"xkxnm":   "2025",
				"xkxnm": cfg.xkxnm,
				//"xkxqm":   "3",
				"xkxqm":   cfg.xkxqm,
				"txbsfrl": "0",
			}).
			Post(baseCfg.QuitCourse)
		if err != nil {
			fmt.Println("退课请求发生错误")
			log.Println(err)
			continue
		}
		if resp.IsError() {
			log.Println("quitCourse", resp.Status())
			fmt.Println("quitCourse", resp.Status())
		}
		if a.LoginCheck(resp) {
			log.Println(resp.String()) // "1"
			if resp.String() == `"1"` {
				//fmt.Println("退课成功")
				return true, resp.String()
			}
			return false, resp.String()
		}

		a.ReLogin()
		// continue
		return false, resp.String()
	}
	return false, "??"
}

func (a *APIClient) getGPA() {
	resp, err := a.Http.R().
		SetRetryCount(0).
		SetTimeout(time.Second*19).
		SetQueryParam("gnmkdm", "N105515").
		Get(baseCfg.AcademiaIndex) // 这玩意实在太低效了
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsError() {
		fmt.Println(resp.Status())
		fmt.Println(resp.Duration())
	}
	if resp.IsSuccess() {

	}
	log.Println(utils.RemoveEmptyLines(resp.String()))
}
