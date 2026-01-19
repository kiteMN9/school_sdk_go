package cas2

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"resty.dev/v3"
)

type mainPageResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Id                     string      `json:"id"`
		OwnerApplication       interface{} `json:"ownerApplication"`
		ServiceName            string      `json:"serviceName"`
		ServicePicUrl          string      `json:"servicePicUrl"`
		ServiceProfile         interface{} `json:"serviceProfile"`
		Status                 interface{} `json:"status"`
		CreateTime             time.Time   `json:"createTime"` // RFC 3339; 2023-05-20T02:27:12.501+00:00
		UpdateTime             time.Time   `json:"updateTime"`
		ServiceDesc            *string     `json:"serviceDesc"`
		ServicePinYin          interface{} `json:"servicePinYin"`
		CollectNum             int         `json:"collectNum"`
		ServiceNo              interface{} `json:"serviceNo"`
		ContactInformation     *string     `json:"contactInformation"`
		ServiceSource          interface{} `json:"serviceSource"`
		ServiceType            interface{} `json:"serviceType"`
		ClickNum               int         `json:"clickNum"`
		CreateUserCode         interface{} `json:"createUserCode"`
		UpdateUserCode         interface{} `json:"updateUserCode"`
		ServiceDepartmentCode  interface{} `json:"serviceDepartmentCode"`
		HaveGuide              string      `json:"haveGuide"`
		FlowId                 interface{} `json:"flowId"`
		SortNum                float64     `json:"sortNum"`
		PublicAccess           string      `json:"publicAccess"`
		Recommend              interface{} `json:"recommend"`
		RecommendMonths        interface{} `json:"recommendMonths"`
		ServiceDepartmentName  string      `json:"serviceDepartmentName"`
		CheckUrl               interface{} `json:"checkUrl"`
		FormId                 interface{} `json:"formId"`
		PrintId                interface{} `json:"printId"`
		YyId                   interface{} `json:"yyId"`
		ForceRecommend         *bool       `json:"forceRecommend"`
		SourceType             interface{} `json:"sourceType"`
		UnlineHand             interface{} `json:"unlineHand"`
		RunTime                interface{} `json:"runTime"`
		MaintainDepartmentCode interface{} `json:"maintainDepartmentCode"`
		MaintainDepartmentName interface{} `json:"maintainDepartmentName"`
		EnableEvaluation       interface{} `json:"enableEvaluation"`
		EvaluationId           interface{} `json:"evaluationId"`
		EvaluationObjectId     interface{} `json:"evaluationObjectId"`
		ReleaseArea            interface{} `json:"releaseArea"`
		ServiceUrl             string      `json:"serviceUrl"` // 重要的
		IconUrl                interface{} `json:"iconUrl"`
		TerminalName           interface{} `json:"terminalName"`
		TokenAccept            interface{} `json:"tokenAccept"`
		TechType               interface{} `json:"techType"`
		LabelNames             []string    `json:"labelNames"`
		Collect                string      `json:"collect"`
		Terminals              interface{} `json:"terminals"`
		LabelId                interface{} `json:"labelId"`
		UseVpn                 string      `json:"useVpn"`
		LabelIds               interface{} `json:"labelIds"`
		ReleaseTime            string      `json:"releaseTime"`
		NeedLocalNetWork       bool        `json:"needLocalNetWork"`
		Rights                 interface{} `json:"rights"`
		ServiceTerminals       interface{} `json:"serviceTerminals"`
		CollectTime            interface{} `json:"collectTime"`
		VisitNum               interface{} `json:"visitNum"`
		VisitTime              interface{} `json:"visitTime"`
		ServiceCas             interface{} `json:"serviceCas"`
	} `json:"data"`
}

func (c *Client) mainPage() {
	var mainPage mainPageResp
	resp, err := c.portalHttp.R().
		SetResult(&mainPage).
		SetQueryParam("random_number", fmt.Sprint(rand.IntN(900)+100)).
		SetHeader("referer", "https://portal.ycit.edu.cn/main.html").
		Get("https://portal.ycit.edu.cn/portal-api/v2/service/showAll?type=4&recommend=false&showPublic=false")
	if err != nil {
		log.Println(err)
		log.Println(resp.String())
		return
	}
	if resp.IsError() {
		log.Println(resp.Error())
		log.Println("mainPage HTTP 状态码错误:", resp.Status())
	}
	if resp.Error() != nil {
		log.Println(resp.Error(), resp.String())
	}
	if mainPage.Code == -1001 {
		fmt.Println(mainPage.Message)
		fmt.Println("需要重新登录")
		//c.LoggedIn = false
		c.Login()
		return
	}
	log.Printf("mainPage:%+v\n", mainPage)
	return
}

func (c *Client) GetJwCookie() bool {
	log.Println("GetJwCookie=======")
	if c.needLogin() {
		c.Login()
	}
	c.mainPage()
	// https://portal.ycit.edu.cn/portal-api/v1/service/useTime/save?id=8aaa844d8804cd12018836fb81f6166d
	fmt.Println("从已登录的门户中得到教务系统临时cookie")
	for range 2 {
		_, err := c.portalHttp.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			SetQueryParam("id", "8aaa844d8804cd12018836fb81f6166d"). // 讲道理这边不该写死在代码里
			Get("https://portal.ycit.edu.cn/portal-api/v1/service/useTime/save")
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			return false
		}
		return true
		//fmt.Println(resp.Status, resp.String())
		//time.Sleep(2 * time.Second)
	}
	return false
}

func (c *Client) GetJwCookie2(location string) string {
	log.Println("GetJwCookie2=======")
	if c.needLogin() {
		c.Login()
	}
	if location == "" {
		return ""
	}
	var location1 string
	for range 3 {
		resp1, err := c.portalHttp.R().
			Get(location)
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		if resp1.StatusCode() != 302 {
			fmt.Println(resp1.Status())
			log.Println("GetJwCookie2 resp1:", resp1.Status())
			continue
		}
		location1 = resp1.Header().Get("Location")
		if location1 == "" {
			log.Println("GetJwCookie2 req location:", location)
			continue
		}
		log.Println("GetJwCookie2 location1:", location1)
		break
	}
	if location1 == "" {
		log.Fatal("GetJwCookie2 location:", location)
	}
	for range 2 {
		resp2, err := c.portalHttp.R().
			Get(location1)
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		if resp2.StatusCode() != 302 {
			fmt.Println(resp2.Status())
			log.Println("GetJwCookie2 resp2:", resp2.Status())
			continue
		}
		location2 := resp2.Header().Get("Location")
		if location2 == "" {
			log.Println("GetJwCookie2 location:", location2)
			continue
		}
		log.Println("location2:", location2)
		return location2
	}
	return ""
}

func (c *Client) needLogin() bool {
	// 检查是否过期
	if time.Now().After(c.nextLoginTimeExp) {
		fmt.Println("Token has expired")
		log.Println("Token has expired")
		return true
	} else {
		//fmt.Println("Token is still valid")
		return false
	}
}

type netCheckResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"`
}

func (c *Client) netCheckIdToken() bool {
	if !checkHeader(c.portalHttp, "x-id-token") {
		return false
	}
	var result netCheckResp
	resp, err := c.portalHttp.R().
		SetResult(&result).
		Get("https://portal.ycit.edu.cn/portal-api/v2/service/networkCheck")
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		log.Println("netCheck:", resp.String())
		return false
	}
	if resp.IsError() {
		log.Println(resp.Error())
		log.Println("netCheckIdToken HTTP 状态码错误:", resp.Status())
	}
	if resp.Error() != nil {
		log.Println(resp.Error(), resp.String())
	}
	if result.Code == 0 {
		return true
	}
	log.Printf("netCheck:%+v\n", result)
	return false
}

// https://authx-service.ycit.edu.cn/personal/api/v1/personal/me/user

// 检查指定头部是否在公共头部中设置
func checkHeader(client *resty.Client, headerKey string) bool {
	// 从 client.Headers 中获取头部值

	values := client.Header().Get(http.CanonicalHeaderKey(headerKey))

	if len(values) > 0 {
		//fmt.Printf("✅ Header '%s' is SET (value: %s)\n", headerKey, values[0])
		return true
	}

	//fmt.Printf("❌ Header '%s' is NOT SET\n", headerKey)
	return false
}
