package client

import (
	"fmt"
	"school_sdk/utils"
	"sync"
	"time"
)

type APIConfig struct {
	syxs string // 距选课结束的小时数
	syts string // 距选课结束的天数
	zxfs string // 已修分数？

	Kklxmc  string
	xkkz_id string
	kklxdm  string
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
	njdm_id_list0 string

	sfkcfx     string // 没什么用参数
	bbhzxjxb   string
	xkxskcgskg string
	jxbzcxskg  string
	sfkknj     string

	sfktk string // 是否可退课
	sfkxk string // 是否可选课
	sfkxq string // 是否可选课?
	xxdm  string // 学校代码
	xklc  string // 轮次
	xz    string // 学制4年

	mzm     string
	ccdm    string
	xbm     string // 性别码 男1 女2
	kkbk    string
	kkbkdj  string
	xszxzt  string
	zyfx_id string // wfx
	xslbdm  string // wlb
	xsbj    string // 4294967296 or 1 ，学生标记,
	jg_id   string // 学院

	rlkz   string
	rlzlkz string
	cdrlkz string
	xkly   string

	tkzgcs_qt string
	currentsj string

	wantClassList   []string
	wantTeacherList []string
	wantTypeList    []string

	modeName       string      // 模式名称: 特殊课程、通识选修课
	startTimeStamp time.Time   // 开始选课时间戳
	modeStore      []ModeStore // 模式存储，用于模式切换

	listDump   bool
	detailDump bool
	needInit   bool
	yl         bool // 余量查询参数
	xztk       bool // 限制退课
	smtpConfig utils.SMTPConfig
}

type ModeStore struct {
	Kklxmc  string
	Kklxdm  string `json:"kklxdm"` // 关键参数，区分不同类型选课
	Xkkz_id string
}

type ChosenDic struct {
	Bdzcbj             string `json:"bdzcbj"`
	Cxbj               string `json:"cxbj"`
	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
	Ddkzbj             string `json:"ddkzbj"`
	Do_jxb_id          string `json:"do_jxb_id"`
	Xkkz_id            string `json:"xkkz_id"`
	Rwlx               string `json:"rwlx"`
	IsInxksj           string `json:"isInxksj"`
	Jdlx               string `json:"jdlx"`
	Jgpxzd             string `json:"jgpxzd"`
	Jsxx               string `json:"jsxx"`
	JxbId              string `json:"jxb_id"`
	Jxbmc              string `json:"jxbmc"`

	Jxbxf     string `json:"jxbxf"`
	Jxbzls    string `json:"jxbzls"`
	Jxdd      string `json:"jxdd"`
	Kcmc      string `json:"kcmc"`
	Kklxdm    string `json:"kklxdm"`
	Kklxmc    string `json:"kklxmc"`
	Xf        string `json:"xf"`
	Kch       string `json:"kch"`
	Kch_id    string `json:"kch_id"` // 课程号ID
	Kklxpx    string `json:"kklxpx"`
	Krrl      string `json:"krrl"`
	Listnav   string `json:"listnav"`
	LocaleKey string `json:"localeKey"`
	Month     string `json:"month"`
	PageTotal int    `json:"pageTotal"`
	Pageable  bool   `json:"pageable"`
	Sfktk     string `json:"sfktk"` // 是否可退课
	Sfxkbj    string `json:"sfxkbj"`
	JxbRS     string `json:"jxbrs"`
	YXzRS     string `json:"yxzrs"`

	Qz          string `json:"qz"`
	Rangeable   bool   `json:"rangeable"`
	Sksj        string `json:"sksj"`
	Sxbj        string `json:"sxbj"`
	TKchId      string `json:"t_kch_id"`
	TotalResult string `json:"totalResult"`
	Xxkbj       string `json:"xxkbj"`
	Year        string `json:"year"`
	Zixf        string `json:"zixf"`
	Zy          string `json:"zy"`

	QueryModel struct {
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
	TotalCount    int  `json:"totalCount"`
	TotalPage     int  `json:"totalPage"`
	TotalResult   int  `json:"totalResult"`
}

type CourseListDic struct {
	Jxb_id string `json:"jxb_id"`
	Jxbmc  string `json:"jxbmc"`  // 教学班名称
	Kklxdm string `json:"kklxdm"` // 关键参数，区分不同类型选课
	Kzmc   string `json:"kzmc"`   // 课程性质
	Kch_id string `json:"kch_id"` // 课程号id
	Kcmc   string `json:"kcmc"`   // 课程名称
	XF     string `json:"xf"`     // 学分
	Yxzrs  string `json:"yxzrs"`  // 已选人数
	Cxbj   string `json:"cxbj"`   // 重修标记 0
	Year   string `json:"year"`
	Xxkbj  string `json:"xxkbj"`

	Jxbzls string `json:"jxbzls"`
	Kch    string `json:"kch"`    // 课程号
	Blyxrs string `json:"blyxrs"` // 本轮已选人数
	Blzyl  string `json:"blzyl"`

	Day      string `json:"day"`
	Month    string `json:"month"`
	Fxbj     string `json:"fxbj"`
	Jgpxzd   string `json:"jgpxzd"`
	Pageable bool   `json:"pageable"`

	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
}

type GetCourseListResult struct {
	TmpList []CourseListDic `json:"tmpList"` // 搜索课程返回的清单
	Sfxsjc  string          `json:"sfxsjc"`
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
	Jxb_id string `json:"jxb_id"`
	Jxbmc  string `json:"jxbmc"` // 教学班名称
	Jxbzls string `json:"jxbzls"`
	Kch    string `json:"kch"`    // 课程号
	Kch_id string `json:"kch_id"` // 课程号id
	Kcmc   string `json:"kcmc"`   // 课程名称
	Kklxdm string `json:"kklxdm"` // 关键参数，区分不同类型选课
	Kzmc   string `json:"kzmc"`   // 课程性质
	XF     string `json:"xf"`     // 学分
	Xxkbj  string `json:"xxkbj"`
	Year   string `json:"year"`
	Yxzrs  string `json:"yxzrs"`
	Cxbj   string `json:"cxbj"` // 重修标记 '0'

	Do_jxb_id string `json:"do_jxb_id"`
	Jxbrl     string `json:"jxbrl"`
	Sksj      string `json:"sksj"`
	Jxdd      string `json:"jxdd"`
	Jsxx      string `json:"jsxx"`
	Xqumc     string `json:"xqumc"`
	Xqh_id    string `json:"xqh_id"`
	Kcxzmc    string `json:"kcxzmc"`
	Kkxymc    string `json:"kkxymc"` // 'kkxymc': '外国语学院'
	Jxms      string `json:"jxms"`   //'jxms': '理论'
	Kclbmc    string `json:"kclbmc"` //'kclbmc': '公共必修课'
	Day       string `json:"day"`

	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`

	Want bool
}

type CourseDetail struct {
	Jxb_id    string `json:"jxb_id"`
	Do_jxb_id string `json:"do_jxb_id"`
	Jxbrl     string `json:"jxbrl"`  // 教学班容量
	Sksj      string `json:"sksj"`   // 上课时间
	Jxdd      string `json:"jxdd"`   // 教学地点 知行楼
	Jsxx      string `json:"jsxx"`   // 教师信息
	Xqumc     string `json:"xqumc"`  // 校区名称
	Xqh_id    string `json:"xqh_id"` // 校区号
	Kcxzmc    string `json:"kcxzmc"` // kcxzmc: 必修
	Kkxymc    string `json:"kkxymc"` // kkxymc: 外国语学院
	Jxms      string `json:"jxms"`   // jxms: 理论
	Kclbmc    string `json:"kclbmc"` // kclbmc: 公共必修课
	Year      string `json:"year"`   // year: 2025

	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
}
