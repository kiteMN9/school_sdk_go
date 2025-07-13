package client

import (
	"bytes"
	"fmt"
	"log"
	baseCfg "school_sdk/config"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (a *APIClient) GetRawInfo() *[]byte {
	resp, err := a.http.R().
		SetQueryParams(map[string]string{
			"gnmkdm": "N100801",
			//"layout": "default",
			"su": a.Account,
		}).Get(baseCfg.INFO)
	//resp, err := a.http.R().
	//	SetQueryParams(map[string]string{
	//		"gnmkdm": "N100801",
	//		//"layout": "default",
	//		"su": a.Account,
	//	}).Get(baseCfg.PersonalInfo)

	if err != nil {
		fmt.Println(err)
	}
	if a.LoginCheck(resp) {
		// Ctrl里有关掉重定向是302，不关是200
		//return true
	} else {
		fmt.Println(resp.StatusCode())
		a.ReLogin()
	}
	info := resp.Body()
	return &info
}

type StudentInfo struct {
	StudentNumber    string // 学号
	Name             string // 姓名
	DepartmentName   string // 学院 jg_id
	ClassName        string // 班级 bh_id
	Grade            string // 年级 njdm_id
	GraduationSchool string // 毕业学校 byzx
	Major            string // 专业方向，基本没用
	Gender           string // 性别 xbm
}

func parseHTML(html *[]byte) (*StudentInfo, error) {

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(*html))
	if err != nil {
		return nil, err
	}

	info := StudentInfo{
		StudentNumber:    strings.TrimSpace(doc.Find("#ajaxForm > div > div.panel-heading > div > div:nth-child(1) > div > div > p").Text()),
		Name:             strings.TrimSpace(doc.Find("#ajaxForm > div > div.panel-heading > div > div:nth-child(2) > div > div > p").Text()),
		DepartmentName:   strings.TrimSpace(doc.Find("#col_jg_id > p").Text()),
		ClassName:        strings.TrimSpace(doc.Find("#col_bh_id > p").Text()),
		Grade:            strings.TrimSpace(doc.Find("#col_njdm_id > p").Text()),
		GraduationSchool: strings.TrimSpace(doc.Find("#col_byzx > p").Text()),
		Major:            strings.TrimSpace(doc.Find("#col_zyfx_id > p").Text()),
		Gender:           strings.TrimSpace(doc.Find("#col_xbm > p").Text()),
	}

	return &info, nil
}

func (a *APIClient) GetInfo() *StudentInfo {
	raw := a.GetRawInfo()
	info, err := parseHTML(raw)
	if err != nil {
		fmt.Println(err)
		log.Println("GetInfo err:", err)
		return &StudentInfo{}

	}
	return info
}

func PrintStudentInfo(info *StudentInfo) {
	// info.ClassName
	fmt.Printf("姓名:%s 班级:%s 学号:%s 毕业学校:%s\n", info.Name, info.ClassName, info.StudentNumber, info.GraduationSchool)
	fmt.Printf("学院:%s 性别:%s 年级:%s\n", info.DepartmentName, info.Gender, info.Grade)
}
