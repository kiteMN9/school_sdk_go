package client

import (
	"encoding/json"
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"strings"
	"time"
)

func (a *APIClient) GetJsonInfo() UserInfo {
	var result UserInfo
	resp, err := a.hedgeC.R().
		SetTimeout(12 * time.Second).
		//SetResult(&result).
		SetQueryParams(map[string]string{
			"gnmkdm": "N100801",
			"su":     a.Config.Account,
		}).Get(baseCfg.InfoJson)
	if err != nil {
		fmt.Println(err)
		return result
	}
	if resp.IsStatusFailure() {
		fmt.Println("GetJsonInfo:", resp.Status())
		log.Println("GetJsonInfo HTTP 状态码错误:", resp.Status())
		return result
	}
	if resp.ResultError() != nil {
		log.Println(resp.ResultError(), resp.String())
		return result
	}
	if a.LoginCheck(resp) {
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}

	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		log.Println("获取个人信息失败 msg:", err, resp.String())
		if strings.Contains(resp.String(), "无功能权限") {
			fmt.Println("获取个人信息无功能权限")
			return result
		}
		fmt.Println(err)
	}

	a.Name = result.Xm
	return result
}

func PrintStudentInfo2(info UserInfo) {
	if info.Xh == "" {
		return
	}
	//fmt.Printf("姓名:%-3s 班级:%-6s 学号:%-6s 毕业学校:%s\n", info.Xm, info.BhId, info.XhId, info.Byzx)
	fmt.Printf("%-3s %-6s %-6s %-1s 毕业学校:%s\n", info.Xm, info.BhId, info.XhId, info.Xbm, info.Byzx)
	fmt.Printf("学院:%-6s 年级:%-4s\n", info.JgId, info.NjdmId)
}

type UserInfo struct {
	Bdzcbj             string `json:"bdzcbj"` // 已注册
	BhId               string `json:"bh_id"`  // 班级名称
	Byzx               string `json:"byzx"`   // 毕业中学
	Bz                 string `json:"bz"`     // 普本
	Csrq               string `json:"csrq"`   // 出生日期
	CyNum              int    `json:"cyNum"`
	Date               string `json:"date"`               // 二○二六年七月十三日
	DateDigit          string `json:"dateDigit"`          // 查询时间 2026年7月13日
	DateDigitSeparator string `json:"dateDigitSeparator"` // 2026-7-13
	Day                string `json:"day"`
	Gddh               string `json:"gddh"`   // 固定电话
	Fdyjgh             string `json:"fdyjgh"` // 辅导员姓名
	HasXszp            string `json:"has_xszp"`
	JdNum              int    `json:"jdNum"`
	Jg                 string `json:"jg"`    // 生员所在地区？
	JgId               string `json:"jg_id"` // 学院
	Jgpxzd             string `json:"jgpxzd"`
	JlNum              int    `json:"jlNum"`
	Jtdh               string `json:"jtdh"` // 家庭电话
	Jtdz               string `json:"jtdz"` // 家庭地址
	Ksh                string `json:"ksh"`  // 考生号
	Listnav            string `json:"listnav"`
	LocaleKey          string `json:"localeKey"` // zh_CN
	Month              string `json:"month"`
	Mzm                string `json:"mzm"`     // 民族 汉族
	NjdmId             string `json:"njdm_id"` // 年级代码入学 2023
	PageTotal          int    `json:"pageTotal"`
	Pageable           bool   `json:"pageable"`
	Pyccdm             string `json:"pyccdm"`    // 本科
	PyfaxxId           string `json:"pyfaxx_id"` // AAAAAAAAAFFFFFFEEEEEECCCCCBBBB11
	Qqhm               string `json:"qqhm"`      // QQ号码
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
	Rxrq        string `json:"rxrq"` // 入学日期 202?-09-0? YYYY-MM-DD
	Rxzf        string `json:"rxzf"` // 入学分数
	Sfzx        string `json:"sfzx"` // 是否？
	Syd         string `json:"syd"`  // 生源地
	Sjhm        string `json:"sjhm"` // 手机号码
	TotalResult string `json:"totalResult"`
	Txdz        string `json:"txdz"`
	UserModel   struct {
		Monitor    bool   `json:"monitor"`
		RoleCount  int    `json:"roleCount"`
		RoleKeys   string `json:"roleKeys"`
		RoleValues string `json:"roleValues"`
		Status     int    `json:"status"`
		Usable     bool   `json:"usable"`
	} `json:"userModel"`
	Xbm    string `json:"xbm"`    // 性别
	Xh     string `json:"xh"`     // 学号
	XhId   string `json:"xh_id"`  // 学号
	Xjztdm string `json:"xjztdm"` // 学籍状态代码:在读
	Xm     string `json:"xm"`     // 姓名
	Xlccdm string `json:"xlccdm"` // 本科
	Xmpy   string `json:"xmpy"`   // 姓名拼音
	Ywxm   string `json:"ywxm"`   // 拼音姓名
	Xnm    string `json:"xnm"`
	Xnmc   string `json:"xnmc"`
	Xqm    string `json:"xqm"`
	Xqmc   string `json:"xqmc"`  // 校区名称？
	Xz     string `json:"xz"`    // 学制 4 年
	Year   string `json:"year"`  // 查询日期年
	Ylzd1  string `json:"ylzd1"` // 分数1 语文
	Ylzd2  string `json:"ylzd2"` // 分数2 数学
	Ylzd3  string `json:"ylzd3"` // 分数3 英语

	Zjhm  string `json:"zjhm"`   // 证件号码
	Zjlxm string `json:"zjlxm"`  // 证件类型
	ZyhId string `json:"zyh_id"` // 专业名称
	Zzmmm string `json:"zzmmm"`  // 政治面貌
}

func (a *APIClient) GetRawInfo() []byte {
	resp, err := a.Http.R().
		SetQueryParams(map[string]string{
			"gnmkdm": "N100801",
			//"layout": "default",
			"su": a.Config.Account,
		}).
		Get(baseCfg.InfoHtm)
	//Get(baseCfg.PersonalInfo)

	if err != nil {
		fmt.Println(err)
	}
	if a.LoginCheck(resp) {
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}
	return resp.Bytes()
}
