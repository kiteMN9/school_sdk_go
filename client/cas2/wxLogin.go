package cas2

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"school_sdk/check_code"
	"school_sdk/client/cas2/utils"
	"school_sdk/config"
	"strings"
	"time"

	"resty.dev/v3"
)

func NewCasWX(account, password string) *Client {
	client := resty.New()
	client.SetBaseURL("https://cas2.ycit.edu.cn/").
		//SetUserAgent(config.ChromeUA).
		SetHeader("user-agent", config.ChromeUA).
		SetRedirectPolicy(resty.RedirectNoPolicy())
	//client.SetProxyURL("http://127.0.0.1:8866")

	hash := md5.Sum([]byte(account + "salt354waragthaswrg"))
	md5Str := hex.EncodeToString(hash[:])
	//fmt.Println("MD5:", md5Str)

	portalHttp := client.Clone(context.Background()).
		SetBaseURL("https://portal.ycit.edu.cn/").
		//SetUserAgent(config.ChromeUA).
		SetHeader("user-agent", config.ChromeUA).
		SetRedirectPolicy(resty.RedirectNoPolicy())

	portalHttp.SetRetryCount(0)

	//client.SetProxyURL("http://127.0.0.1:8866")
	if os.Getenv("trace") == "1" {
		portalHttp.SetTrace(true)
		//client.SetLogger()
	}
	if os.Getenv("proxy") == "1" {
		portalHttp.SetProxy("http://127.0.0.1:8866")
	}

	return &Client{
		Account:     account,
		password:    password,
		fpVisitorId: md5Str, // fingerprint
		http:        client,
		portalHttp:  portalHttp,
	}
}

func (c *Client) WXLogin() bool {
	c.http.SetRedirectPolicy(resty.RedirectNoPolicy())
	check_code.SaveImgStream(c.getQrCode(), "./", "qrcode")
	uuid, imgUrl, state := c.wxUUID()
	fmt.Println("imgUrl:", imgUrl)
	check_code.SaveImgStream(c.getWXQrCode(imgUrl), "./", "WXQrcode")
	fmt.Println("扫描二维码，完成微信登录")
	wxCode := c.GetScanResultCode(uuid)
	_, stat, location := c.WxLoginFinal(wxCode, state)
	if stat {
		//c.LoggedIn = true
		fmt.Println("已完成登录流程，点击下方链接可直达门户")
		fmt.Println(location)
		fmt.Println("已完成登录流程，点击上方链接可直达门户")
		return true
	}
	return false
}

func (c *Client) wxUUID() (string, string, string) {
	var Location, wxLocation string
	//https://cas2.ycit.edu.cn/cas/federatedRedirect?service=https://portal.ycit.edu.cn/?path%3Dhttps://portal.ycit.edu.cn/main.html%23/&federatedName=openweixin
outerLoop:
	for {
		resp, err := c.http.R().
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetQueryParam("federatedName", "openweixin").
			Get("https://cas2.ycit.edu.cn/cas/federatedRedirect")
		if err != nil {
			fmt.Println("wxUUID:", err)
			log.Println("wxUUID:", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if resp.StatusCode() != 302 {
			time.Sleep(1 * time.Second)
			continue
		}
		Location = resp.Header().Get("Location")
		if Location == "" {
			continue
		}
		for {
			//// https://cas2.ycit.edu.cn/cas/federation/federated/openweixin?redirectUri=https%3A%2F%2Fcas2.ycit.edu.cn%2Fcas%2FfederatedCallback%2Fopenweixin%3Fservice%3Dhttps%253A%252F%252Fportal.ycit.edu.cn%252F%253Fpath%253Dhttps%253A%252F%252Fportal.ycit.edu.cn%252Fmain.html%2523%252F&state=TST-762-ZDk2hJqqTP3RGEdnsfohEf-BR9dymPRu
			resp2, err2 := c.http.R().
				//	// https://cas2.ycit.edu.cn/cas/federatedCallback/openweixin?service=https%3A%2F%2Fportal.ycit.edu.cn%2F%3Fpath%3Dhttps%3A%2F%2Fportal.ycit.edu.cn%2Fmain.html%23%2F
				Get(Location)
			if err2 != nil {
				fmt.Println("wxUUIDl:", err2)
				log.Println("wxUUIDl:", err2)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if resp2.StatusCode() != 302 {
				time.Sleep(1 * time.Second)
				continue outerLoop
			}
			wxLocation = resp.Header().Get("Location")
			//wxLocation,err := resp2.Location()
			//https://open.weixin.qq.com/connect/qrconnect?appid=wxdf4bda39b1e560ab&redirect_uri=https%3A%2F%2Fcas2.ycit.edu.cn%2Fcas%2Ffederation%2FfederatedCallback%2Fopenweixin&response_type=code&scope=snsapi_login,snsapi_userinfo&state=XksYVW#wechat_redirect
			//fmt.Println("wxLocation", wxLocation)
			break
		}
		if wxLocation == "" {
			log.Println("wxUUID: location is null")
			time.Sleep(1 * time.Second)
			continue
		}

		// 解析 location
		parsedUrl, locationErr := url.Parse(wxLocation)
		if locationErr != nil {
			continue
		}
		state := parsedUrl.Query().Get("state")
		if state == "" {
			log.Println("wxUUID: state is null")
			time.Sleep(1 * time.Second)
			continue
		}

		for {
			resp3, err3 := c.http.R().
				Get(wxLocation)
			if err3 != nil {
				fmt.Println("wxUUIDw:", err3)
				time.Sleep(1 * time.Second)
				continue
			}

			html := resp3.String()
			var uuid, imgUrl string
			if strings.Contains(html, `"/connect/qrcode/`) && strings.Contains(html, `web_qrcode_img" src="`) {
				uuid = strings.Split(strings.Split(html, `"/connect/qrcode/`)[1], `"`)[0]
				imgUrl = "https://open.weixin.qq.com" + strings.Split(strings.Split(html, `web_qrcode_img" src="`)[1], `"`)[0]
			}
			log.Println("uuid:", uuid)
			log.Println("imgUrl:", imgUrl)
			log.Println("state:", state)
			return uuid, imgUrl, state
		}
	}
}

func (c *Client) getWXQrCode(url string) []byte {
	for range 5 {
		resp, err := c.http.R().
			Get(url)
		if err != nil {
			fmt.Println("getWXQrCode:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		qrcode := resp.Bytes()
		return qrcode
	}
	return []byte{}
}

func (c *Client) GetScanResultCode(uuid string) string {
	//SetUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 8_0 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Mobile/12A365 MicroMessenger/5.4.1 NetType/WIFI WebView/doc")
	defer func() {
		_ = os.Remove("./WXQrcode.png")
	}()
	time.Sleep(2 * time.Second)
	for range 15 {
		resp, err := c.http.R().
			SetQueryParam("uuid", uuid).
			Get("https://lp.open.weixin.qq.com/connect/l/qrconnect")
		if err != nil {
			fmt.Println("getScanResult:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		// window.wx_errcode = 408; window.wx_code = '';
		result := resp.String()
		if strings.Contains(result, `window.wx_code='';`) {
			time.Sleep(1 * time.Second)
			continue
		}
		//window.wx_errcode=405;window.wx_code='0117xIkl2JDozg4nZEkl2oVrZT37xIkl';
		if strings.Contains(result, `window.wx_code='`) {
			wxCode, _, _ := strings.Cut(strings.Split(result, `window.wx_code='`)[1], "';")
			log.Println("wxCode:", wxCode)
			return wxCode
		}
		fmt.Println("Code not found:", uuid)
	}
	log.Println("wxCode: null; uuid:", uuid)
	return ""
}

func (c *Client) WxLoginFinal(wxCode, state string) (string, bool, string) {
	//GET ?=0117xIkl2JDozg4nZEkl2oVrZT37xIkl&=Ui2AIW
	resp, err := c.http.R().
		SetQueryParam("code", wxCode).
		SetQueryParam("state", state).
		SetRetryCount(2).
		Get("https://cas2.ycit.edu.cn/cas/federation/federatedCallback/openweixin")
	if err != nil {
		return "", false, ""
	}
	if resp.StatusCode() != 302 {
		result := resp.String()
		fmt.Println(result)
	}
	location := resp.Header().Get("Location")
	log.Println("location:", location)
	//https://cas2.ycit.edu.cn/cas/federatedCallback/openweixin?service=https%3A%2F%2Fportal.ycit.edu.cn%2F%3Fpath%3Dhttps%3A%2F%2Fportal.ycit.edu.cn%2Fmain.html%23%2F&federatedCode=SI0aYr&state=TST-846-ClDtDL9bS4Lt0h6hm6DV2ZjBLPT8AtaT

	resp2, err2 := c.http.R().
		SetRetryCount(2).
		Get(location)
	if err2 != nil {
		return "", false, ""
	}
	if resp2.StatusCode() != 302 {
		result := resp2.String()
		fmt.Println(result)
	}
	location2 := resp2.Header().Get("Location")
	log.Println("location2:", location2)

	resp3, err3 := c.http.R().
		SetRetryCount(2).
		Get(location2)
	if err3 != nil {
		return "", false, ""
	}
	if resp3.StatusCode() != 302 {
		result := resp3.String()
		fmt.Println(result)
	}

	resp4, err4 := c.http.R().
		SetRetryCount(2).
		Get("https://cas2.ycit.edu.cn/cas/login?service=https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/")
	if err4 != nil {
		return "", false, ""
	}
	if resp4.StatusCode() != 302 {
		result := resp4.String()
		fmt.Println(result)
	}
	location4 := resp4.Header().Get("Location")
	if location4 == "" {
		log.Fatal("location is null")
	}
	log.Println("location4:", location4)
	//fmt.Println(location4)
	// 解析 location
	parsedUrl, _ := url.Parse(location4)
	query := parsedUrl.Query()
	ticketJWT := query.Get("ticket")
	if ticketJWT == "" {
		log.Fatal("ticket is null")
	}
	log.Println("ticketJWT:", ticketJWT)
	var idToken string
	var err1 error
	idToken, c.nextLoginTimeExp, c.Account, err1 = utils.ExtractIDToken(ticketJWT)
	if err1 != nil {
		fmt.Printf("错误: %v\n", err1)
		log.Println("ticketJWT:", ticketJWT)
		log.Println("ticket解析失败:", err1.Error())
		return "", false, ""
	}
	log.Println("idToken", idToken)
	c.portalHttp.SetHeader("x-id-token", idToken)
	c.portalHttp.SetHeader("x-device-info", "PC")
	c.portalHttp.SetHeader("x-terminal-info", "PC")
	c.portalHttp.SetHeader("cookie", "isLogin=true")
	log.Println("当前账号:", c.Account)
	c.fCfg.TicketJWT = ticketJWT
	c.fCfg.WriteConfig()
	return idToken, true, location4
}
