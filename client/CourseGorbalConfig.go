package client

import (
	"fmt"
	"sync"
	"time"
)

type APIConfig struct {
	// mu sync.RWMutex
	minCredit      string // 本学期选课要求总学分最低
	maxCredit      string // 总学分最高
	selectedCredit string // 本学期已选学分
	syxs           string // 距选课结束的小时数
	syts           string // 距选课结束的天数
	zxfs           string // 已修分数？
	xkkssj         string // 2026-09-09 12:30:00
	xkjssj         string // 选课结束时间 2026-09-10 11:00:0

	//account string // 账号、学号
	// firstKklxmc string
	Kklxmc  string // 课程类型名称
	xkkz_id string
	xkkz_xh string
	kklxdm  string // 课程类型代码
	rwlx    string
	bklx_id string
	bh_id   string // 班号
	xqh_id  string // 校区号
	zyh_id  string // 专业号
	njdm_id string // 年级代码
	xkxqm   string // 学期 3 12 16
	xkxnm   string // 学年

	jdlx          string
	sfznkx        string
	kzybkxy       string
	sfkkzy        string
	zdkxms        string
	gnjkxdnj      string
	sfkkjyxdxnxq  string
	njdm_id_list0 string //

	sfkcfx     string // 没什么用参数
	bbhzxjxb   string
	xkxskcgskg string
	jxbzcxskg  string
	sfkknj     string

	sfktk  string // 是否可退课
	sfkxk  string // 是否可选课
	sfkxq  string // 是否开学前
	xxdm   string // 学校代码
	xklc   string // 轮次
	xklcmc string // 轮次名称
	xz     string // 学制4年

	mzm     string
	ccdm    string // 层次代码
	xbm     string // 性别码 男1 女2
	kkbk    string
	kkbkdj  string
	xszxzt  string
	zyfx_id string // wfx
	xslbdm  string // wlb
	xsbj    string // 4294967296 or 1 ，学生标记,
	jg_id   string // 学院, 机构ID

	rlkz   string // 容量控制
	rlzlkz string // 容量总量控制
	cdrlkz string
	xkly   string // 选课来源

	tkzgcs_qt string
	currentsj string

	wantClassList   []string
	wantTeacherList []string
	wantTypeList    []string

	modeName       string      // 当前模式名称: 特殊课程、通识选修课
	startTimeStamp time.Time   // 开始选课时间戳
	modeStore      []ModeStore // 模式存储，用于模式切换

	listDump   bool
	detailDump bool
	needInit   bool
	yl         bool // 余量查询参数
	xztk       bool // 限制退课
}

// ModeStore zzxkYzb.js
// function queryCourse(a_element,kklxdm,xkkz_id,njdm_id,zyh_id,xkkz_xh){}
type ModeStore struct {
	Kklxmc  string
	Kklxdm  string
	Xkkz_id string
	Njdm_id string
	Zyh_id  string
	Xkkz_xh string
}

type ChosenDic struct {
	Bdzcbj             string `json:"bdzcbj"`
	Cxbj               string `json:"cxbj"`
	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
	Ddkzbj             string `json:"ddkzbj"`
	Do_jxb_id          string `json:"do_jxb_id"` //
	Xkkz_id            string `json:"xkkz_id"`   // 'xkkz_id': '0',
	Rwlx               string `json:"rwlx"`      // 'rwlx': '1',
	IsInxksj           string `json:"isInxksj"`
	Jdlx               string `json:"jdlx"`
	Jgpxzd             string `json:"jgpxzd"`
	Jsxx               string `json:"jsxx"`   // "320057/陈爱华/副教授"
	JxbId              string `json:"jxb_id"` //
	Jxbmc              string `json:"jxbmc"`  // 教学班名称 '化工原理（上）-0002',

	Jxbxf     string `json:"jxbxf"`
	Jxbzls    string `json:"jxbzls"`
	Jxdd      string `json:"jxdd"`
	Kcmc      string `json:"kcmc"` // 课程名称 "化工原理（上）"
	Kklxdm    string `json:"kklxdm"`
	Kklxmc    string `json:"kklxmc"`
	Xf        string `json:"xf"`     // 学分 'xf': '3'
	Kch       string `json:"kch"`    // 'kch': '0003021031',
	Kch_id    string `json:"kch_id"` // 课程号ID 'kch_id': '0003021031',
	Kklxpx    string `json:"kklxpx"`
	Krrl      string `json:"krrl"`
	Listnav   string `json:"listnav"`
	LocaleKey string `json:"localeKey"`
	Month     string `json:"month"`
	PageTotal int    `json:"pageTotal"`
	Pageable  bool   `json:"pageable"`
	Sfktk     string `json:"sfktk"`  // 是否可退课
	Sfxkbj    string `json:"sfxkbj"` // 是否处于选课
	JxbRS     string `json:"jxbrs"`  // 'jxbrs': '68'
	YXzRS     string `json:"yxzrs"`  // 'yxzrs': '68'

	Qz          string `json:"qz"`
	Rangeable   bool   `json:"rangeable"`
	Sksj        string `json:"sksj"`
	Sxbj        string `json:"sxbj"`
	TKchId      string `json:"t_kch_id"`
	TotalResult string `json:"totalResult"`
	Xxkbj       string `json:"xxkbj"` // 选修课标记?
	Year        string `json:"year"`
	Zixf        string `json:"zixf"`
	Zy          string `json:"zy"`

	QueryModel struct {
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

	UserModel struct {
		Monitor    bool   `json:"monitor"`
		RoleCount  int    `json:"roleCount"`
		RoleKeys   string `json:"roleKeys"`
		RoleValues string `json:"roleValues"`
		Status     int    `json:"status"`
		Usable     bool   `json:"usable"`
	} `json:"userModel"`
}

type ChooseCourseResult struct {
	Flag string `json:"flag"`
	Msg  string `json:"msg"`
}

type CourseListDicQueryModel struct {
	CurrentPage   int  `json:"currentPage"`
	CurrentResult int  `json:"currentResult"`
	EntityOrField bool `json:"entityOrField"`
	Limit         int  `json:"limit"`
	Offset        int  `json:"offset"`
	PageNo        int  `json:"pageNo"`
	PageSize      int  `json:"pageSize"`
	ShowCount     int  `json:"showCount"`
	// sorts // 'sorts': []
	TotalCount  int `json:"totalCount"`
	TotalPage   int `json:"totalPage"`
	TotalResult int `json:"totalResult"`
}

// type userModel struct {
// 	Monitor    bool   `json:"monitor"`
// 	RoleCount  int    `json:"roleCount"`
// 	RoleKeys   string `json:"roleKeys"`
// 	RoleValues string `json:"roleValues"`
// 	Status     int    `json:"status"`
// 	Usable     bool   `json:"usable"`
// }

type CourseListDic struct {
	Blyxrs             string `json:"blyxrs"` // 本轮已选人数
	Blzyl              string `json:"blzyl"`
	Cxbj               string `json:"cxbj"`               // 重修标记 0
	Date               string `json:"date"`               // '二○二五年二月二十六日'
	DateDigit          string `json:"dateDigit"`          // '2025年2月26日'
	DateDigitSeparator string `json:"dateDigitSeparator"` // '2025-2-26'
	Day                string `json:"day"`
	Fxbj               string `json:"fxbj"`
	Jgpxzd             string `json:"jgpxzd"`
	JxbId              string `json:"jxb_id"` // 教学班id，用于连接List和Detail
	Jxbmc              string `json:"jxbmc"`  // 教学班名称  "艺术哲学：美是如何诞生的(艺术类)-0001"
	Jxbxf              string `json:"jxbxf"`
	Jxbzls             string `json:"jxbzls"` // 'jxbzls': '1'
	Kch                string `json:"kch"`    // 课程号  '9000000398'
	KchId              string `json:"kch_id"` // 课程号 id
	Kclxmc             string `json:"kclxmc"`
	Kcmc               string `json:"kcmc"` // 课程名称  "艺术哲学：美是如何诞生的(艺术类)"
	Kcrow              string `json:"kcrow"`
	Kklxdm             string `json:"kklxdm"` // 关键参数，区分不同类型选课  '10'
	Kzmc               string `json:"kzmc"`   // 课程性质  "艺术类"
	Listnav            string `json:"listnav"`
	LocaleKey          string `json:"localeKey"`
	Month              string `json:"month"`
	PageTotal          int    `json:"pageTotal"`
	Pageable           bool   `json:"pageable"`
	QueryModel         struct {
		CurrentPage   int   `json:"currentPage"`
		CurrentResult int   `json:"currentResult"`
		EntityOrField bool  `json:"entityOrField"`
		Limit         int   `json:"limit"`
		Offset        int   `json:"offset"`
		PageNo        int   `json:"pageNo"`
		PageSize      int   `json:"pageSize"`
		ShowCount     int   `json:"showCount"`
		Sorts         []any `json:"sorts"`
		TotalCount    int   `json:"totalCount"`
		TotalPage     int   `json:"totalPage"`
		TotalResult   int   `json:"totalResult"`
	} `json:"queryModel"`
	Rangeable   bool   `json:"rangeable"`
	Rwzxs       string `json:"rwzxs"`
	Sftj        string `json:"sftj"`
	TotalResult string `json:"totalResult"`
	UserModel   struct {
		Monitor    bool   `json:"monitor"`
		RoleCount  int    `json:"roleCount"`
		RoleKeys   string `json:"roleKeys"`
		RoleValues string `json:"roleValues"`
		Status     int    `json:"status"`
		Usable     bool   `json:"usable"`
	} `json:"userModel"`
	Xf      string `json:"xf"`    // 学分  "1.5"
	Xxkbj   string `json:"xxkbj"` // 选修课标记?
	Year    string `json:"year"`  // '2025'
	Yxzrs   string `json:"yxzrs"` // 已选人数  "70"
	Zcongbj string `json:"zcongbj"`
}

type GetCourseListResult struct {
	TmpList []CourseListDic `json:"tmpList"` // 搜索课程返回的清单
	Sfxsjc  string          `json:"sfxsjc"`
	Msg     string          `json:"msg"`
	Flag    string          `json:"flag"` // 成功 flag=1
}

type CourseDetail struct {
	Date               string `json:"date"`               //'date': '二○二五年二月二十日'
	DateDigit          string `json:"dateDigit"`          //'dateDigit': '2025年2月20日'
	DateDigitSeparator string `json:"dateDigitSeparator"` //'dateDigitSeparator': '2025-2-20'
	Day                string `json:"day"`                //'day': '20'

	DoJxbId   string `json:"do_jxb_id"`
	FjxbId    string `json:"fjxb_id"`
	Jgpxzd    string `json:"jgpxzd"`
	Jsxx      string `json:"jsxx"`   // 教师信息 440015/裴如意/副教授
	JxbId     string `json:"jxb_id"` // 教学班id，用于连接List和Detial 以 jxb_id 为课程唯一标识符
	Jxbmc     string `json:"jxbmc"`
	Jxbrs     string `json:"jxbrs"` // 教学班人数
	Jxdd      string `json:"jxdd"`  // 教学地点 知行楼108
	Listnav   string `json:"listnav"`
	LocaleKey string `json:"localeKey"`
	Month     string `json:"month"`
	PageTotal int    `json:"pageTotal"`
	Pageable  bool   `json:"pageable"`

	Rangeable   bool   `json:"rangeable"`
	Sksj        string `json:"sksj"` // 上课时间 星期二第1-2节{2-13周}
	TotalResult string `json:"totalResult"`

	Xsdm string `json:"xsdm"`
	Xsmc string `json:"xsmc"`
	Year string `json:"year"` // year: 2025

	Jxbrl  string `json:"jxbrl"`  // 教学班容量
	Xqumc  string `json:"xqumc"`  // 校区名称 北校区
	Xqh_id string `json:"xqh_id"` // 校区号 3
	Kcxzmc string `json:"kcxzmc"` // kcxzmc: 必修
	Kkxymc string `json:"kkxymc"` // kkxymc: 外国语学院
	Jxms   string `json:"jxms"`   // jxms: 理论
	Kclbmc string `json:"kclbmc"` // kclbmc: 公共必修课
	// Yqmc string `json:"yqmc"` //'yqmc': '--'

	// kcxzmc string `json:"kcxzmc"` //
}

type SafeCustomCourseSlice struct {
	mu    sync.RWMutex
	items []CustomCourseDic
}

func NewCustomCourseSlice() *SafeCustomCourseSlice {
	return &SafeCustomCourseSlice{
		items: make([]CustomCourseDic, 0),
	}
}

func (s *SafeCustomCourseSlice) Append(item CustomCourseDic) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.items {
		if v.Jxb_id == item.Jxb_id {
			return
		}
	}
	s.items = append(s.items, item)
}

func (s *SafeCustomCourseSlice) Get(index int) CustomCourseDic {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index >= len(s.items) {
		return CustomCourseDic{}
	}
	return s.items[index]
}

func (s *SafeCustomCourseSlice) Update(index int, newItem CustomCourseDic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index >= len(s.items) {
		return fmt.Errorf("list index out of range")
	}
	s.items[index] = newItem
	return nil
}

type CustomCourseDic struct {
	Jxb_id string `json:"jxb_id"` // 教学班id，用于连接 List 和 Detail
	Jxbmc  string `json:"jxbmc"`  // 教学班名称  艺术哲学：美是如何诞生的(艺术类)-0001
	Jxbzls string `json:"jxbzls"` // 'jxbzls': '1'
	Kch    string `json:"kch"`    // 课程号
	Kch_id string `json:"kch_id"` // 课程号id
	Kcmc   string `json:"kcmc"`   // 课程名称  艺术哲学：美是如何诞生的(艺术类)
	Kklxdm string `json:"kklxdm"` // 关键参数，区分不同类型选课  '10'
	Kzmc   string `json:"kzmc"`   // 课程性质  'kzmc': '艺术类'
	XF     string `json:"xf"`     // 学分  1.5
	Xxkbj  string `json:"xxkbj"`  // 选修课标记?
	Year   string `json:"year"`   // 'year': '2025'
	Yxzrs  string `json:"yxzrs"`  // 已选人数  'yxzrs': '70'
	Cxbj   string `json:"cxbj"`   // 重修标记 '0'

	Do_jxb_id string `json:"do_jxb_id"` //
	Jxbrl     string `json:"jxbrl"`     // 'jxbrl': '66'
	Sksj      string `json:"sksj"`      //'sksj': '星期二第1-2节{2-13周}
	Jxdd      string `json:"jxdd"`      // 'jxdd': '知行楼108',
	Jsxx      string `json:"jsxx"`      //jsxx': '440015/裴如意/副教授'
	Xqumc     string `json:"xqumc"`     //'xqumc': '北校区'
	Xqh_id    string `json:"xqh_id"`    //'xqh_id': '3'
	Kcxzmc    string `json:"kcxzmc"`    // 'kcxzmc': '必修'
	Kkxymc    string `json:"kkxymc"`    // 'kkxymc': '外国语学院'
	Jxms      string `json:"jxms"`      //'jxms': '理论'
	Kclbmc    string `json:"kclbmc"`    //'kclbmc': '公共必修课'
	Day       string `json:"day"`       // 'day': '26'
	// 'date': '二○二五年二月二十六日', 'dateDigit': '2025年2月26日', 'dateDigitSeparator': '2025-2-26'
	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"` //'dateDigit': '2025年2月20日'
	DateDigitSeparator string `json:"dateDigitSeparator"`

	Want bool
}
