package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http/cookiejar"
	"os"
	"regexp"
	"school_sdk/client/rsa"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"school_sdk/check_code"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
)

var ExistVerify = fmt.Errorf("请先滑动图片进行验证！")
var IncorrectPassword = fmt.Errorf("用户名或密码不正确，请重新输入！")
var loginMU sync.Mutex
var lastSuccessTime = time.Unix(0, 0)

func (a *APIClient) ReLogin() bool {
	loginMU.Lock()
	defer loginMU.Unlock()
	// 多线程情况下还得加个1~2秒的成功登录冷静期，防止一解锁就重复登录
	if time.Since(lastSuccessTime) < 1000*time.Millisecond {
		return true
	}
	if a.onlyCookieMethod {
		fmt.Println("登录可能过期，需要更新cookie")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}
	fmt.Println("\r重新登录")
	reStartTime := time.Now()
	if a.Login() {
		lastSuccessTime = time.Now()
		log.Println("重新登录用时:", time.Since(reStartTime))
		return true
	}
	return false
}

func (a *APIClient) Login() bool {
	if a.enableCas2 {
		if a.cas2Login() {
			if a.cas2Client.Account != a.Account {
				a.Account = a.cas2Client.Account
			}
			return true
		} else {
			return false
		}
	}
	var LoginExtend = generateLoginExtend(a.Config.userAgent)
	count := 0
	for count < 5 {
		if a.Config.ExistVerify {
			// if verify_type
			if a.getCaptchaLogin(LoginExtend) {
				return true
			}
			fmt.Println("重新开始登录流程")
			count++
			continue
		} else {
			var csrfToken string
			var wg sync.WaitGroup
			var encryptedResult string
			var reqTime string
			//Eb := fmt.Sprint(time.Now().UnixMilli())
			wg.Add(1)
			go a.getRsaPublicKey(&wg, &reqTime, &encryptedResult)
			wg.Wait()
			wg.Add(1)
			go a.getRawCsrfToken(&wg, &csrfToken)

			wg.Wait()
			stat, err := a.postLogin(csrfToken, reqTime, encryptedResult)
			if errors.Is(err, ExistVerify) {
				a.Config.ExistVerify = true
				utils.UpdateConfigUserInfo(a.filename, a.Config.ExistVerify)
				continue
			}
			return stat
		}
	}
	return false
}

func (a *APIClient) cas2Login() bool {
	log.Println("cas2Login=======")
	if !a.cas2Client.Login() {
		return false
	}
	if !a.cas2Client.GetJwCookie() {
		return false
	}
	location := a.ssoLogin()
	if location == "" {
		return false
	}
	location = a.cas2Client.GetJwCookie2(location)
	if location == "" {
		return false
	}
	return a.ssoLogin2(location)
}

func (a *APIClient) ssoLogin() string {
	log.Println("ssoLogin=======")
	for range 3 {
		resp, err := a.Http.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			Get("https://jwglxt.ycit.edu.cn/sso/hnyyxyiotlogin")
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			continue
		}
		if resp.StatusCode() != 302 {
			fmt.Println(resp.Status())
			log.Println("sso/hnyyxyiotlogin not 302")
			continue
		}
		location := resp.Header().Get("Location")
		log.Println(location)
		return location
	}
	return ""
}

// set-cookie
func (a *APIClient) ssoLogin2(location string) bool {
	log.Println("ssoLogin2======")
	if location == "" {
		return false
	}
	location = strings.Replace(location, "http://", "https://", -1)
	log.Println("ssoLogin2 replaced url:", location)
	var location2 string
	for range 5 {
		resp, err := a.Http.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			Get(location)
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode() != 302 {
			fmt.Println(resp.Status())
			log.Println(resp.Status())
			continue
		}
		location2 = resp.Header().Get("Location")
		if location2 == "" {
			continue
		} else {
			break
		}
	}

	log.Println(location2)
	location2 = strings.Replace(location2, "http://", "https://", -1)
	for range 5 {
		resp2, err2 := a.Http.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			Get(location2)
		if err2 != nil {
			fmt.Println(err2)
			log.Println(err2)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp2.StatusCode() == 302 {
			return true
		} else {
			continue
		}
	}
	return false
}

func (a *APIClient) getCaptchaLogin(LoginExtend string) bool {
	// 控制整个滑块验证码登录
	var csrfToken string
	var rtk string
	var encryptedResult string
	var wg sync.WaitGroup
	var reqTime string

	wg.Add(1)
	// 选课时csrf获取时间过长，所以先获取rtk并设置cookie
	a.getRTK(&wg, &rtk)
	wg.Wait()
	//Eb := fmt.Sprint(time.Now().UnixMilli())
	wg.Add(1)
	// 希望在整个验证码识别过程中能获取到csrf以节省时间
	go a.getRawCsrfToken(&wg, &csrfToken)

	// fmt.Println(encryptedResult)
	// 发起登录请求
	for range 3 {
		wg.Add(1)
		// 验证码获取与识别，顺便获取公钥
		a.captchaControl(&wg, LoginExtend, &csrfToken, &rtk, &encryptedResult, &reqTime)
		wg.Wait()

		stat, err := a.postLogin(csrfToken, reqTime, encryptedResult)
		if errors.Is(err, ExistVerify) {
			if a.Config.ExistVerify {
				log.Println("重试验证码")
				continue
			}
			return false
		}
		if errors.Is(err, IncorrectPassword) {
			cfg := utils.UpdateConfigUserInfo(a.filename, a.Config.ExistVerify)
			a.Account = cfg.Account
			a.passwd = cfg.Passwd
			wg.Add(1)
			a.getRsaPublicKey(&wg, &reqTime, &encryptedResult)
			wg.Wait()
		} else {
			return stat
		}
	}
	return false
}

func (a *APIClient) captchaControl(wg *sync.WaitGroup, LoginExtend string, csrfToken, rtk, encryptedResult, t *string) {
	// 控制除了RTK的整个验证码识别过程
	defer wg.Done()
	captchaStartTime := time.Now()
	for {
		*t = fmt.Sprint(time.Now().UnixMilli())
		captchaParams := a.getCaptchaParams(*rtk, *t)
		if captchaParams.VS == "verified" {
			log.Println("验证码已通过验证")
			return
		}
		for captchaParams.Msg != "" {
			//log.Println("验证码已通过验证")
			fmt.Println("清空cookie")
			log.Println("清空cookie")
			jar, _ := cookiejar.New(nil)
			a.Http.SetCookieJar(jar)
			wg.Add(2)
			a.getRawCsrfToken(wg, csrfToken)
			a.getRTK(wg, rtk)
			*t = fmt.Sprint(time.Now().UnixMilli())
			captchaParams = a.getCaptchaParams(*rtk, *t)
		}
		imgStream, err := a.getCaptchaImage(captchaParams.Imtk, captchaParams.Mi, captchaParams.T)
		if err != nil {
			continue
		}
		wg.Add(1)
		*t = fmt.Sprint(time.Now().UnixMilli())
		// 将公钥获取放在这里以节省时间，并确保公钥是新鲜的
		go a.getRsaPublicKey(wg, t, encryptedResult)
		capStartTime := time.Now()
		x := check_code.FindBestMatch(imgStream)
		log.Println("识别用时:", time.Since(capStartTime))
		verResult := a.captchaVerify(*rtk, LoginExtend, x)
		// log.Println("captcha_verify:", ver_result)
		if verResult {
			// wg.Wait()
			log.Println("验证用时:", time.Since(captchaStartTime))
			return
		}

		fmt.Println(":( 滑块验证失败")
		log.Println(":( 滑块验证失败")
		check_code.SaveImgStream(imgStream, "fail/", "fail_"+fmt.Sprint(x)+"_"+fmt.Sprint(time.Now().UnixMilli()))
	}
}

func (a *APIClient) getRawCsrfToken(wg *sync.WaitGroup, csrf *string) {
	// 获取CSRF令牌
	defer wg.Done()
	var exists = false
	var failCount int
	for {
		// log.Println("csrf debug")
		resp, err := a.Http.R().
			SetRetryCount(1).
			Get(baseCfg.LoginIndex) // ?language=zh_CN&_t=MiniSecond

		if err != nil {
			log.Println("csrf HTTP 请求失败:", failCount, err)
			failCount++
			if failCount > 2 {
				fmt.Printf("\r%d %s", failCount, err.Error())
			}
			continue
		}
		if failCount > 2 {
			fmt.Println()
		}

		// 解析 HTML 文档
		htmlReader := bytes.NewReader(resp.Bytes())
		doc, err1 := goquery.NewDocumentFromReader(htmlReader)
		if err1 != nil {
			log.Println("csrf 解析 HTML 失败:", err1)
			time.Sleep(150 * time.Millisecond)
			continue
		}

		// 使用 CSS 选择器提取元素属性
		*csrf, exists = doc.Find("#csrftoken").Attr("value")
		if !exists {
			bodyStr := resp.String()
			if utils.UserIsLogin(a.Account, bodyStr) {
				return
			}
			log.Println("未找到 #csrftoken 元素或其 value 属性")
			log.Println("url 填的有问题吧")
			log.Println(bodyStr)
			fmt.Println("请检查url填写是否有误，特别是 /jwglxt")
			time.Sleep(1 * time.Second)
			continue
		} else {
			// fmt.Println("CSRF Token:\n" + *csrf)
			return
		}
	}
	// log.Fatal("csrf 出错")
}

func (a *APIClient) getRTK(wg *sync.WaitGroup, rtk *string) {
	// 获取 cookie rtk
	defer wg.Done()
	for {
		resp, err := a.Http.R().
			EnableTrace().
			SetQueryParams(map[string]string{
				"type":       "resource",
				"instanceId": "zfcaptchaLogin",
				"name":       "zfdun_captcha.js",
			}).
			Get(baseCfg.CAPTCHA)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				time.Sleep(275 * time.Millisecond)
				fmt.Println("rtk 请求超时")
				continue
			} else {
				fmt.Println("rtk http:", err)
				log.Println(err)
			}
			time.Sleep(1475 * time.Millisecond)
			continue
		}

		if resp.StatusCode() == 404 {
			fmt.Println("404, url 填的有问题吧")
			log.Println("404, url 填的有问题吧")
			time.Sleep(4 * time.Second)
			continue
		}

		if resp.StatusCode() != 200 {
			fmt.Println("rtk HTTP 错误: 状态码 ", resp.Status())
			log.Println("rtk HTTP 错误: 状态码 ", resp.Status())
			time.Sleep(275 * time.Millisecond)
			continue
		}

		var re = regexp.MustCompile(`tk:'(.*)',`)
		matches := re.FindStringSubmatch(resp.String())
		if len(matches) < 2 {
			fmt.Println("未找到rtk, url 填的有问题吧")
			log.Println("未找到rtk, url 填的有问题吧")
			time.Sleep(4 * time.Second)
		} else {
			*rtk = matches[1]
			// fmt.Println("rtk:\n" + rtk) // csrfToken}
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
	VS     string `json:"vs"`     // not_verify
	Status string `json:"status"` // success
}

func (a *APIClient) getCaptchaParams(rtk, t string) captchaData {
	var jsonResult captchaData
	for {
		resp, err := a.Http.R().
			SetResult(&jsonResult).
			SetQueryParams(map[string]string{
				"type":       "refresh",
				"rtk":        rtk,
				"time":       t,
				"instanceId": "zfcaptchaLogin",
			}).Get(baseCfg.CAPTCHA)

		if err != nil {
			fmt.Println("cap http:", err)
			log.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp.IsError() {
			log.Println(resp.Error())
		}
		if resp.Error() != nil {
			log.Println(resp.Error())
		}
		if jsonResult.Msg != "" {
			fmt.Println(jsonResult.Msg)
		}

		return jsonResult
	}
}

func (a *APIClient) getCaptchaImage(imtk, id string, T int) ([]byte, error) {
	// time.Sleep(76 * time.Second)
	for i := 0; i < 2; i++ {
		resp2, err := a.Http.R().
			SetRetryCount(0).
			SetTimeout(79*time.Second). // 不能睡到79秒 (76-78)
			SetHeader("Accept", "image/apng,image/*,*/*").
			SetQueryParams(map[string]string{
				"type":       "image",
				"id":         id,
				"imtk":       imtk,
				"t":          fmt.Sprint(T),
				"instanceId": "zfcaptchaLogin",
			}).Get(baseCfg.CAPTCHA)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			fmt.Println("get image http error")
			log.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp2.StatusCode() == 404 { // 过期了，重试也没用
			break
		}
		if resp2.IsError() {
			continue
		}

		if len(resp2.Bytes()) == 0 {
			log.Println("未获取到 image")
			return []byte{}, fmt.Errorf("未获取到image")
		}

		return resp2.Bytes(), nil
	}
	return []byte{}, fmt.Errorf("未获取到 image")
}

type rsaResponseData struct {
	Modulus  string `json:"modulus"`
	Exponent string `json:"exponent"`
}

func (a *APIClient) getRsaPublicKey(wg *sync.WaitGroup, t *string, enResult *string) {
	// 获取RSA公钥信息
	// 注意：公钥会经常刷新
	var jsonResult rsaResponseData
	defer wg.Done()
	for {
		resp, err := a.Http.R().
			SetHeader("Accept", "application/json, */*").
			SetResult(&jsonResult).
			SetQueryParams(map[string]string{
				"time": *t,
				//"_":    *Eb,
			}).Get(baseCfg.PublicKey)
		if err != nil {
			fmt.Println("pubkey 获取错误:", err)
			log.Println("pubkey HTTP 请求失败:", err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp.IsError() {
			log.Println("pubkey HTTP 错误: 状态码 ", resp.Status())
			continue
		}
		if resp.Error() != nil {
			log.Println(resp.Error(), resp.String())
		}
		if jsonResult.Modulus == "" || jsonResult.Exponent == "" {
			log.Println("pubkey 获取错误:", resp.Status(), resp.String(), *t)
			*t = fmt.Sprint(time.Now().UnixMilli())
			continue
		}
		*enResult, err = rsa.EncryptRsa(jsonResult.Modulus, jsonResult.Exponent, a.passwd)
		if err != nil {
			*t = fmt.Sprint(time.Now().UnixMilli())
			continue
		}
		return
	}
}

type captchaVerifyData struct {
	Msg    string `json:"msg"`    // 验证失败,请稍后重试
	VS     string `json:"vs"`     // verified, not_verify
	Status string `json:"status"` // success, fail
}

func (a *APIClient) captchaVerify(rtk, LoginExtend string, x int) bool {
	captchaVerifyResult := check_code.GetTrackString(x, 480)
	if captchaVerifyResult == "" {
		return false
	}
	var result captchaVerifyData
	for range 2 {
		resp, err := a.Http.R().
			SetResult(&result).
			SetFormData(map[string]string{ // 这里不支持json
				"type":       "verify",
				"rtk":        rtk,
				"time":       fmt.Sprint(time.Now().UnixMilli()),
				"mt":         base64.StdEncoding.EncodeToString([]byte(captchaVerifyResult)),
				"instanceId": "zfcaptchaLogin",
				"extend":     base64.StdEncoding.EncodeToString([]byte(LoginExtend)),
			}).Post(baseCfg.CAPTCHA)

		if err != nil {
			fmt.Println("captcha_verify http error:", err)
			log.Println("captcha_verify HTTP 请求失败:", err)
			// fmt.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Println("captcha_verify HTTP 错误: 状态码 ", resp.Status())
		}
		// fmt.Println(resp)
		// {"msg":"","vs":"verified","status":"success"}

		if result.VS == "verified" && result.Status == "success" {
			return true
		} else if result.VS == "not_verify" {
			return false
		}
	}
	return false
}

func (a *APIClient) postLogin(csrf, t, mm string) (bool, error) {
	// fmt.Println("postLogin sleep 300")
	// time.Sleep(300 * time.Second)
	for range 6 {
		resp, err := a.Http.R().
			SetQueryParam("time", t).
			SetFormData(map[string]string{
				"csrftoken": csrf,
				"yhm":       a.Account,
				"mm":        mm,
			}).Post(baseCfg.LoginIndex)
		if err != nil {
			fmt.Println("postLogin http error")
			log.Println("postLogin HTTP 请求失败:", err)
			// fmt.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		//log.Println()
		if resp.IsError() {
			log.Println("postLogin HTTP 错误: 状态码 ", resp.Status())
			continue
		}
		// fmt.Println(resp)
		stat, err1 := isLogin(a.Account, resp.String())
		if err1 != nil {
			return false, err1
		}
		if resp.StatusCode() == 302 || stat {
			//fmt.Println("postLogin", resp.Status())
			fmt.Println("登录成功")
			log.Println("登录成功 Location:", resp.Header().Get("Location"))
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

	// 检查是否存在 id="tips"
	// re2 := regexp.MustCompile(`id="tips"`)
	// if !re2.MatchString(html) {
	// 	return true, nil
	// }

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return false, nil
	}

	// 获取错误提示信息
	errMsg := strings.TrimSpace(doc.Find("#tips").Text())

	if errMsg == "" {
		return false, nil
	}

	if strings.Contains(errMsg, "验证码") {
		log.Println(errMsg)
		return false, nil
	}

	// fmt.Printf("UserIsLogin(): %s\n", errMsg)
	fmt.Println(errMsg)
	log.Println(errMsg)
	if strings.Contains(errMsg, "用户名或密码不正确") {
		return false, IncorrectPassword
	}
	if strings.Contains(errMsg, "请先滑动图片进行验证！") {
		return false, ExistVerify
	}
	return false, nil
}

func generateLoginExtend(UserAgent string) string {
	// 查找第一个 '/' 的位置
	slashIndex := strings.Index(UserAgent, "/")
	modifiedUserAgent := UserAgent

	if slashIndex != -1 {
		// 截取第一个 '/' 之后的内容
		modifiedUserAgent = UserAgent[slashIndex+1:]
	}

	// 创建 JSON 结构体
	loginExtend := struct {
		AppName    string `json:"appName"`
		UserAgent  string `json:"userAgent"`
		AppVersion string `json:"appVersion"`
	}{
		AppName:    "Netscape",
		UserAgent:  UserAgent,
		AppVersion: modifiedUserAgent,
	}

	// 序列化为 JSON 字符串
	jsonBytes, _ := json.Marshal(loginExtend)
	LoginExtend := string(jsonBytes)
	return LoginExtend
}
