package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"school_sdk/check_code"
	baseCfg "school_sdk/config"
	"school_sdk/rsa"
	"school_sdk/utils"
)

var ExistVerify = fmt.Errorf("请先滑动图片进行验证！")

func (a *APIClient) ReLogin() bool {
	fmt.Println("\r重新登录")
	return a.Login()
}

func (a *APIClient) Login() bool {
	var LoginExtend = generateLoginExtend(a.config.userAgent)
	for {
		if a.config.ExistVerify {

			return a.getCaptchaLogin(LoginExtend)
		} else {
			var csrfToken string
			var wg sync.WaitGroup
			var encryptedResult string
			var reqTime string

			wg.Add(1)
			go a.getRsaPublicKey(&wg, &reqTime, &a.passwd, &encryptedResult)
			wg.Wait()
			wg.Add(1)
			go a.getRawCsrfToken(&wg, &csrfToken)

			wg.Wait()
			stat, err := a.postLogin(&csrfToken, &a.Account, &reqTime, &encryptedResult)
			if errors.Is(err, ExistVerify) {
				a.config.ExistVerify = true
				//utils.UpdateConfigUserInfo(a.config.ExistVerify)
				continue
			}
			return stat
		}
	}
}

func (a *APIClient) getCaptchaLogin(LoginExtend *string) bool {

	var csrfToken string
	var rtk string
	var encryptedResult string
	var wg sync.WaitGroup
	var reqTime string

	wg.Add(1)

	a.getRTK(&wg, &rtk)
	wg.Wait()

	wg.Add(1)

	go a.getRawCsrfToken(&wg, &csrfToken)

	wg.Add(1)

	a.captchaControl(&wg, &rtk, &a.passwd, LoginExtend, &encryptedResult, &reqTime)

	wg.Wait()

	for {
		stat, err := a.postLogin(&csrfToken, &a.Account, &reqTime, &encryptedResult)
		if err != nil {
			//cfg := utils.UpdateConfigUserInfo(, a.config.ExistVerify)
			//a.Account = cfg.Account
			//a.passwd = cfg.Passwd
			wg.Add(1)
			a.getRsaPublicKey(&wg, &reqTime, &a.passwd, &encryptedResult)
			wg.Wait()
		} else {
			return stat
		}
	}

}

func (a *APIClient) captchaControl(wg *sync.WaitGroup, rtk, passwd, LoginExtend, encryptedResult, t *string) {
	// 控制除了RTK的整个验证码识别过程
	defer wg.Done()
	captchaStartTime := time.Now().UnixMilli()
	for {
		*t = fmt.Sprint(time.Now().UnixMilli())

		captchaParams := a.getCaptchaParams(rtk, t)
		//imgStream, err := a.getCaptchaImage(&captchaParams.Imtk, &captchaParams.Mi, captchaParams.T)
		imgStream, err := a.getCaptchaImage(&captchaParams.Imtk, &captchaParams.Si, captchaParams.T)
		if err != nil {
			continue
		}
		wg.Add(1)
		*t = fmt.Sprint(time.Now().UnixMilli())

		go a.getRsaPublicKey(wg, t, passwd, encryptedResult)
		capStartTime := time.Now().UnixMilli()
		x := check_code.Identify(imgStream)
		log.Printf("识别用时: %dms\n", time.Now().UnixMilli()-capStartTime)
		verResult := a.captchaVerify(rtk, LoginExtend, x)

		if verResult {

			log.Printf("验证用时: %dms\n", time.Now().UnixMilli()-captchaStartTime)
			return
		} else {
			fmt.Println(":( 滑块验证失败")
			log.Println("滑块验证失败")
			check_code.SaveImg(imgStream, "fail_"+fmt.Sprint(x)+"_")
		}
	}
}

func (a *APIClient) getRawCsrfToken(wg *sync.WaitGroup, csrf *string) {
	// 获取CSRF令牌
	defer wg.Done()

	var exists = false
	for {

		resp, err := a.http.R().Get(baseCfg.LoginIndex)
		if err != nil {
			log.Println("csrf HTTP 请求失败:", err)
			continue
		}

		htmlReader := bytes.NewReader(resp.Body())
		bodyStr := resp.String()
		doc, err1 := goquery.NewDocumentFromReader(htmlReader)
		if err1 != nil {
			log.Println("csrf 解析 HTML 失败:", err1)
			continue
		}

		*csrf, exists = doc.Find("#csrftoken").Attr("value")
		if !exists {
			if utils.UserIsLogin(a.Account, bodyStr) {
				return
			}
			log.Println("未找到 #csrftoken 元素或其 value 属性")
			log.Println("url 填的有问题吧")
			log.Println(bodyStr)
			time.Sleep(1 * time.Second)
			continue
		} else {

			return
		}
	}

}

func (a *APIClient) getRTK(wg *sync.WaitGroup, rtk *string) {
	// 获取 cookie rtk
	defer wg.Done()
	for {
		resp, err := a.http.R().
			SetQueryParams(map[string]string{
				"type":       "resource",
				"instanceId": "zfcaptchaLogin",
				"name":       "zfdun_captcha.js",
			}).
			//SetQueryString("type=resource&instanceId=zfcaptchaLogin&name=zfdun_captcha.js").
			Get(baseCfg.CAPTCHA)
		if err != nil {
			fmt.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}

		var re = regexp.MustCompile(`tk:'(.*)',`)
		matches := re.FindStringSubmatch(resp.String())
		if len(matches) < 2 {
			fmt.Println("未找到rtk, url 填的有问题吧")
			log.Println("未找到rtk, url 填的有问题吧")
			time.Sleep(2 * time.Second)
		} else {
			*rtk = matches[1]

			return
		}
	}

}

type captchaData struct {
	Msg    string `json:"msg"`
	T      int    `json:"t"`
	Si     string `json:"si"`
	Imtk   string `json:"imtk"`
	Mi     string `json:"mi"`
	VS     string `json:"vs"`
	Status string `json:"status"`
}

func (a *APIClient) getCaptchaParams(rtk, t *string) *captchaData {

	for {
		resp1, err := a.http.R().
			SetResult(&captchaData{}).
			SetQueryParams(map[string]string{
				"type":       "refresh",
				"rtk":        *rtk,
				"time":       *t,
				"instanceId": "zfcaptchaLogin",
			}).Get(baseCfg.CAPTCHA)

		if err != nil {
			fmt.Println(err)
			continue
		}

		jsonResult := resp1.Result().(*captchaData)
		return jsonResult
	}
}

func (a *APIClient) getCaptchaImage(imtk, id *string, T int) (*[]byte, error) {

	for i := 0; i < 3; i++ {
		resp2, err := a.http.R().
			SetQueryParams(map[string]string{
				"type":       "image",
				"id":         *id,
				"imtk":       *imtk,
				"t":          fmt.Sprint(T),
				"instanceId": "zfcaptchaLogin",
			}).Get(baseCfg.CAPTCHA)
		if err != nil {
			log.Println(err)
			continue
		}
		if resp2.StatusCode() == 404 {
			break
		}

		if len(resp2.Body()) == 0 {
			log.Println("未获取到 image")
			return &[]byte{}, fmt.Errorf("未获取到image")
		}

		body := resp2.Body()
		if resp2.Body() != nil {
			err2 := resp2.RawResponse.Body.Close()
			if err2 != nil {
				return nil, err2
			}
		}
		return &body, nil
	}
	return &[]byte{}, fmt.Errorf("未获取到 image")
}

type rsaResponseData struct {
	Modulus  string `json:"modulus"`
	Exponent string `json:"exponent"`
}

func (a *APIClient) getRsaPublicKey(wg *sync.WaitGroup, t *string, secret, enResult *string) {
	// 获取RSA公钥信息
	// 注意：公钥会经常刷新
	defer wg.Done()
	for {
		resp, err := a.http.R().
			SetResult(&rsaResponseData{}).
			SetQueryParams(map[string]string{
				"time": *t,
				"_":    *t,
			}).Get(baseCfg.PublicKey)
		if err != nil {
			log.Println("pubkey HTTP 请求失败:", err)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Println("pubkey HTTP 错误: 状态码 ", resp.StatusCode())
			continue
		}
		result := resp.Result().(*rsaResponseData)
		rsa.EncryptRsa(&result.Modulus, &result.Exponent, secret, enResult)
		return
	}
}

type captchaVerifyData struct {
	Msg    string `json:"msg"`
	VS     string `json:"vs"`
	Status string `json:"status"`
}

func (a *APIClient) captchaVerify(rtk, LoginExtend *string, x int) bool {
	captchaVerifyResult := check_code.GetTrackString(x, 480)
	if captchaVerifyResult == "" {
		return false
	}
	for range 2 {
		resp, err := a.http.R().
			SetResult(&captchaVerifyData{}).
			SetFormData(map[string]string{
				"type":       "verify",
				"rtk":        *rtk,
				"time":       fmt.Sprint(time.Now().UnixMilli()),
				"mt":         base64.StdEncoding.EncodeToString([]byte(captchaVerifyResult)),
				"instanceId": "zfcaptchaLogin",
				"extend":     base64.StdEncoding.EncodeToString([]byte(*LoginExtend)),
			}).Post(baseCfg.CAPTCHA)

		if err != nil {
			log.Println("captcha_verify HTTP 请求失败:", err)

			continue
		}

		if resp.StatusCode() != 200 {
			log.Println("captcha_verify HTTP 错误: 状态码 ", resp.StatusCode())
		}

		result := resp.Result().(*captchaVerifyData)
		if result.VS == "verified" && result.Status == "success" {
			return true
		} else if result.VS == "not_verify" {
			return false
		}
	}
	return false
}

func (a *APIClient) postLogin(csrf, account, t, mm *string) (bool, error) {

	for range 6 {
		resp, err := a.http.R().
			SetQueryParam("time", *t).
			SetFormData(map[string]string{
				"csrftoken": *csrf,
				"yhm":       *account,
				"mm":        *mm,
			}).Post(baseCfg.LoginIndex)
		if err != nil {
			log.Println("postLogin HTTP 请求失败:", err)

			continue
		}

		if resp.StatusCode() != 200 && resp.StatusCode() != 302 {
			log.Println("postLogin HTTP 错误: 状态码 ", resp.StatusCode())
		}
		stat, err1 := isLogin(*account, resp.String())
		if err1 != nil {
			return false, err1
		}
		if resp.StatusCode() == 302 || stat {
			fmt.Println("登录成功")
			log.Println("登录成功")
			return true, nil
		}
	}
	return false, nil
}

func isLogin(account, html string) (bool, error) {
	accountPattern := fmt.Sprintf(`value="%s"`, regexp.QuoteMeta(account))
	re1 := regexp.MustCompile(accountPattern)
	if re1.MatchString(html) {
		return true, nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return false, nil
	}

	errMsg := strings.TrimSpace(doc.Find("#tips").Text())

	if errMsg == "" {
		return false, nil
	}

	if strings.Contains(errMsg, "验证码") {
		log.Println(errMsg)
		return false, nil
	}

	fmt.Println(errMsg)
	log.Println(errMsg)
	if strings.Contains(errMsg, "用户名或密码不正确") {
		return false, fmt.Errorf("用户名或密码不正确，请重新输入！")
	}
	if strings.Contains(errMsg, "请先滑动图片进行验证！") {
		return false, ExistVerify
	}
	return false, nil
}

func generateLoginExtend(UserAgent string) *string {

	slashIndex := strings.Index(UserAgent, "/")
	modifiedUserAgent := UserAgent

	if slashIndex != -1 {

		modifiedUserAgent = UserAgent[slashIndex+1:]
	}

	loginExtend := struct {
		AppName    string `json:"appName"`
		UserAgent  string `json:"userAgent"`
		AppVersion string `json:"appVersion"`
	}{
		AppName:    "Netscape",
		UserAgent:  UserAgent,
		AppVersion: modifiedUserAgent,
	}

	jsonBytes, _ := json.Marshal(loginExtend)
	LoginExtend := string(jsonBytes)
	return &LoginExtend
}
