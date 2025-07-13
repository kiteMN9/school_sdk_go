package client

import (
	"fmt"
	"sync"
)

type APIConfig struct {
	syxs string
	syts string
	zxfs string

	Kklxmc  string
	xkkz_id string
	kklxdm  string
	rwlx    string
	bklx_id string
	bh_id   string
	xqh_id  string
	zyh_id  string
	njdm_id string
	xkxqm   string
	xkxnm   string

	jdlx          string
	sfznkx        string
	kzybkxy       string
	sfkkzy        string
	zdkxms        string
	gnjkxdnj      string
	sfkkjyxdxnxq  string
	njdm_id_list0 string

	sfkcfx     string
	bbhzxjxb   string
	xkxskcgskg string
	jxbzcxskg  string
	sfkknj     string

	sfktk  string
	sfkxk  string
	sfkxq  string
	xxdm   string
	rlzlkz string
	xklc   string
	xz     string

	mzm     string
	ccdm    string
	xbm     string // 性别？
	kkbk    string
	kkbkdj  string
	xszxzt  string
	zyfx_id string
	xslbdm  string
	xsbj    string
	jg_id   string

	rlkz string
	xkly string

	wantClassList   []string
	wantTeacherList []string
	wantTypeList    []string

	modeName       string      // 模式名称: 特殊课程、通识选修课
	startTimeStamp int64       // 开始选课时间戳
	modeStore      []ModeStore // 模式存储，用于模式切换

	listDump   bool
	detailDump bool
	needInit   bool
}

type ModeStore struct {
	Kklxmc  string
	Kklxdm  string `json:"kklxdm"`
	Xkkz_id string
}

type ChosenDic struct {
	Do_jxb_id string `json:"do_jxb_id"`
	Xkkz_id   string `json:"xkkz_id"`
	Rwlx      string `json:"rwlx"`
	Jxbmc     string `json:"jxbmc"`
	Kcmc      string `json:"kcmc"`
	Xf        string `json:"xf"`
	Kch_id    string `json:"kch_id"`
	Kch       string `json:"kch"`
	Sfktk     string `json:"sfktk"`
	JxbRS     string `json:"jxbrs"`
	YXzRS     string `json:"yxzrs"`
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
	Jxbmc  string `json:"jxbmc"`
	Kklxdm string `json:"kklxdm"`
	Kzmc   string `json:"kzmc"`
	Kch_id string `json:"kch_id"`
	Kcmc   string `json:"kcmc"`
	XF     string `json:"xf"`
	Yxzrs  string `json:"yxzrs"`
	Cxbj   string `json:"cxbj"`
	Year   string `json:"year"`
	Xxkbj  string `json:"xxkbj"`

	Jxbzls string `json:"jxbzls"`
	Kch    string `json:"kch"`
	Blyxrs string `json:"blyxrs"`
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
	Jxbmc  string `json:"jxbmc"`
	Jxbzls string `json:"jxbzls"`
	Kch    string `json:"kch"`
	Kch_id string `json:"kch_id"`
	Kcmc   string `json:"kcmc"`
	Kklxdm string `json:"kklxdm"`
	Kzmc   string `json:"kzmc"`
	XF     string `json:"xf"`
	Xxkbj  string `json:"xxkbj"`
	Year   string `json:"year"`
	Yxzrs  string `json:"yxzrs"`
	Cxbj   string `json:"cxbj"`

	Do_jxb_id string `json:"do_jxb_id"`
	Jxbrl     string `json:"jxbrl"`
	Sksj      string `json:"sksj"`
	Jxdd      string `json:"jxdd"`
	Jsxx      string `json:"jsxx"`
	Xqumc     string `json:"xqumc"`
	Xqh_id    string `json:"xqh_id"`
	Kcxzmc    string `json:"kcxzmc"`
	Kkxymc    string `json:"kkxymc"`
	Jxms      string `json:"jxms"`
	Kclbmc    string `json:"kclbmc"`
	Day       string `json:"day"`

	Date               string `json:"date"`
	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`

	Want bool
}

type CourseDetail struct {
	Jxb_id    string `json:"jxb_id"`
	Do_jxb_id string `json:"do_jxb_id"`
	Jxbrl     string `json:"jxbrl"`
	Sksj      string `json:"sksj"`
	Jxdd      string `json:"jxdd"`
	Jsxx      string `json:"jsxx"`
	Xqumc     string `json:"xqumc"`
	Xqh_id    string `json:"xqh_id"`
	Kcxzmc    string `json:"kcxzmc"`
	Kkxymc    string `json:"kkxymc"`
	Jxms      string `json:"jxms"`
	Kclbmc    string `json:"kclbmc"`
	Year      string `json:"year"`

	DateDigit          string `json:"dateDigit"`
	DateDigitSeparator string `json:"dateDigitSeparator"`
	Day                string `json:"day"`
}
