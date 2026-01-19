package client

import (
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"time"
)

func (a *APIClient) GetJsonInfo() UserInfo {
	var result UserInfo
	resp, err := a.Http.R().
		SetTimeout(12 * time.Second).
		SetResult(&result).
		SetQueryParams(map[string]string{
			"gnmkdm": "N100801",
			"su":     a.Account,
		}).Get(baseCfg.InfoJson)
	if err != nil {
		fmt.Println(err)
	}
	if resp.IsError() {
		log.Println(resp.Error())
		log.Println("GetJsonInfo HTTP 状态码错误:", resp.Status())
	}
	if resp.Error() != nil {
		log.Println(resp.Error(), resp.String())
	}
	if a.LoginCheck(resp) {
	} else {
		fmt.Println(resp.Status())
		a.ReLogin()
	}

	return result
}

func PrintStudentInfo2(info UserInfo) {
	if info.Xm == "" {
		return
	}
	fmt.Printf("姓名:%-3s 班级:%-6s 学号:%-6s 毕业学校:%s\n", info.Xm, info.BhId, info.XhId, info.Byzx)
	fmt.Printf("学院:%-6s 性别:%-1s 年级:%-4s\n", info.JgId, info.Xbm, info.NjdmId)
}

type UserInfo struct {
	Bdzcbj             string `json:"bdzcbj"` // 已注册
	BhId               string `json:"bh_id"`  // 班级
	Byzx               string `json:"byzx"`   // 毕业中学
	Bz                 string `json:"bz"`     // 普本
	Csrq               string `json:"csrq"`   //
	CyNum              int    `json:"cyNum"`
	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
	Gddh               string `json:"gddh"`   // 固定电话
	Fdyjgh             string `json:"fdyjgh"` // 辅导员姓名
	HasXszp            string `json:"has_xszp"`
	JdNum              int    `json:"jdNum"`
	JgId               string `json:"jg_id"` // 学院
	Jgpxzd             string `json:"jgpxzd"`
	JlNum              int    `json:"jlNum"`
	Jtdh               string `json:"jtdh"` // 家庭电话
	Jtdz               string `json:"jtdz"` // 家庭地址
	Ksh                string `json:"ksh"`  // 考生号
	Listnav            string `json:"listnav"`
	LocaleKey          string `json:"localeKey"`
	Month              string `json:"month"`
	Mzm                string `json:"mzm"` // 民族 汉族
	NjdmId             string `json:"njdm_id"`
	PageTotal          int    `json:"pageTotal"`
	Pageable           bool   `json:"pageable"`
	Pyccdm             string `json:"pyccdm"` // 本科
	PyfaxxId           string `json:"pyfaxx_id"`
	Qqhm               string `json:"qqhm"` // QQ号码
	QueryModel         struct {
		CurrentPage   int           `json:"currentPage"`
		CurrentResult int           `json:"currentResult"`
		EntityOrField bool          `json:"entityOrField"`
		Limit         int           `json:"limit"`
		Offset        int           `json:"offset"`
		PageNo        int           `json:"pageNo"`
		PageSize      int           `json:"pageSize"`
		ShowCount     int           `json:"showCount"`
		Sorts         []interface{} `json:"sorts"`
		TotalCount    int           `json:"totalCount"`
		TotalPage     int           `json:"totalPage"`
		TotalResult   int           `json:"totalResult"`
	} `json:"queryModel"`
	Rangeable   bool   `json:"rangeable"`
	Rxrq        string `json:"rxrq"` // 入学日期 202?-09-0? YYYY-MM-DD
	Rxzf        string `json:"rxzf"` // 入学分数
	Sfzx        string `json:"sfzx"`
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
	Xjztdm string `json:"xjztdm"` // 在读
	Xm     string `json:"xm"`     // 姓名
	Xnm    string `json:"xnm"`
	Xnmc   string `json:"xnmc"`
	Xqm    string `json:"xqm"`
	Xqmc   string `json:"xqmc"`
	Xz     string `json:"xz"` // 学制 4 年
	Year   string `json:"year"`
	Ylzd1  string `json:"ylzd1"` // 分数1 语文
	Ylzd2  string `json:"ylzd2"` // 分数2 数学
	Ylzd3  string `json:"ylzd3"` // 分数3 英语

	Zjhm  string `json:"zjhm"`   // 证件号码
	Zjlxm string `json:"zjlxm"`  // 证件类型
	ZyhId string `json:"zyh_id"` // 专业名称
	Zzmmm string `json:"zzmmm"`  // 政治面貌
}

//func (a *APIClient) GetRawInfo() *[]byte {
//	resp, err := a.Http.R().
//		SetQueryParams(map[string]string{
//			"gnmkdm": "N100801",
//			//"layout": "default",
//			"su": a.Account,
//		}).Get(baseCfg.INFO)
//	//resp, err := a.Http.R().
//	//	SetQueryParams(map[string]string{
//	//		"gnmkdm": "N100801",
//	//		//"layout": "default",
//	//		"su": a.Account,
//	//	}).Get(baseCfg.PersonalInfo)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//	if a.LoginCheck(resp) {
//		// Ctrl里有关掉重定向是302，不关是200
//		//return true
//	} else {
//		fmt.Println(resp.Status)
//		a.ReLogin()
//	}
//	info := resp.Bytes()
//	return &info
//}

//type StudentInfo struct {
//	StudentNumber    string // 学号
//	Name             string // 姓名
//	DepartmentName   string // 学院
//	ClassName        string // 班级
//	Grade            string // 年级
//	GraduationSchool string // 毕业学校
//	Major            string // 专业方向，基本没用
//	Gender           string // 性别
//	ID               string // 证件号
//	PhoneNum         string // 手机号
//	HomeAddress      string // 家庭住址
//	PostalCode       string // 邮政编码
//	PoliticalStatus  string // 政治面貌
//	Nationality      string // 民族
//}

//func parseHTML(html *[]byte) (*StudentInfo, error) {
//	//doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
//	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(*html))
//	if err != nil {
//		return nil, err
//	}
//
//	info := StudentInfo{
//		StudentNumber:    strings.TrimSpace(doc.Find("#ajaxForm > div > div.panel-heading > div > div:nth-child(1) > div > div > p").Text()),
//		Name:             strings.TrimSpace(doc.Find("#ajaxForm > div > div.panel-heading > div > div:nth-child(2) > div > div > p").Text()),
//		DepartmentName:   strings.TrimSpace(doc.Find("#col_jg_id > p").Text()),
//		ClassName:        strings.TrimSpace(doc.Find("#col_bh_id > p").Text()),
//		Grade:            strings.TrimSpace(doc.Find("#col_njdm_id > p").Text()),
//		GraduationSchool: strings.TrimSpace(doc.Find("#col_byzx > p").Text()),
//		Major:            strings.TrimSpace(doc.Find("#col_zyfx_id > p").Text()),
//		Gender:           strings.TrimSpace(doc.Find("#col_xbm > p").Text()),
//		ID:               strings.TrimSpace(doc.Find("#col_zjhm > p").Text()),
//		PhoneNum:         strings.TrimSpace(doc.Find("#col_gddh > p").Text()),
//		HomeAddress:      strings.TrimSpace(doc.Find("#col_jtdz > p").Text()),
//		PostalCode:       strings.TrimSpace(doc.Find("#col_yzbm > p").Text()),
//	}
//
//	return &info, nil
//}

//func (a *APIClient) GetInfo() *StudentInfo {
//	raw := a.GetRawInfo()
//	info, err := parseHTML(raw)
//	if err != nil {
//		fmt.Println(err)
//		log.Println("GetInfo err:", err)
//		return &StudentInfo{}
//		// panic(err)
//	}
//	return info
//}

//func PrintStudentInfo(info *StudentInfo) {
//	// info.ClassName
//	fmt.Printf("姓名:%s 班级:%s 学号:%s 毕业学校:%s\n", info.Name, info.ClassName, info.StudentNumber, info.GraduationSchool)
//	fmt.Printf("学院:%s 性别:%s 年级:%s\n", info.DepartmentName, info.Gender, info.Grade)
//}
