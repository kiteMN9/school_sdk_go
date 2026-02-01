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
	"github.com/antchfx/htmlquery"

	"school_sdk/check_code"
	baseCfg "school_sdk/config"
	"school_sdk/utils"
)

var ExistVerify = fmt.Errorf("请先滑动图片进行验证！")
var IncorrectPassword = fmt.Errorf("用户名或密码不正确，请重新输入！")
var CsrfEmpty = fmt.Errorf("CSRF is empty")
var loginMU sync.Mutex
var lastSuccessTime = time.Unix(0, 0)

func (a *APIClient) ReLogin() bool {
	loginMU.Lock()
	defer loginMU.Unlock()
	// 多线程情况下还得加个1~2秒的成功登录冷静期，防止一解锁就重复登录
	if time.Since(lastSuccessTime) < 1100*time.Millisecond {
		return true
	}
	if a.onlyCookieMethod {
		fmt.Println("登录可能过期，需要更新cookie")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}
	fmt.Println("\r重新登录")
	reStartTime := time.Now()
	//log.Println("清空 cookie")
	//jar, _ := cookiejar.New(nil)
	//a.Http.SetCookieJar(jar)
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
		}
		return false
	}
	var LoginExtend = generateLoginExtend(a.Config.userAgent)
	count := 0
	for count < 15 {
		reqTime := fmt.Sprint(time.Now().UnixMilli())
		csrfToken := a.getRawCsrfToken()
		if a.Config.ExistVerify {
			// if verify_type
			if a.getCaptchaLogin(LoginExtend, csrfToken, reqTime) {
				return true
			}
			fmt.Println("重新开始登录流程")
			count++
			continue
		} else {
			var wg sync.WaitGroup
			var encryptedResult string
			//Eb := fmt.Sprint(time.Now().UnixMilli())
			wg.Add(1)
			go a.getRsaPublicKey(context.TODO(), &wg, &reqTime, &encryptedResult)
			wg.Wait()

			//csrfToken = a.getRawCsrfToken()
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

func (a *APIClient) getCaptchaLogin(LoginExtend []byte, csrfToken, reqTime string) bool {
	// 控制整个滑块验证码登录
	var rtk string
	var encryptedResult string
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Eb := fmt.Sprint(time.Now().UnixMilli())
	rtk = a.getRTK()

	// fmt.Println(encryptedResult)
	// 发起登录请求
	for range 3 {
		wg.Add(1)
		// 验证码获取与识别，顺便获取公钥
		if !a.captchaControl(ctx, &wg, LoginExtend, &csrfToken, &rtk, &encryptedResult, &reqTime) {
			return false
		}
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
			//wg.Add(1)
			//a.getRsaPublicKey(ctx, &wg, &reqTime, &encryptedResult)
			//wg.Wait()
			continue
		} else {
			return stat
		}
	}
	return false
}

func (a *APIClient) captchaControl(ctx context.Context, wg *sync.WaitGroup, LoginExtend []byte, csrfToken, rtk, encryptedResult, t *string) bool {
	// 控制除了RTK的整个验证码识别过程
	defer wg.Done()
	captchaStartTime := time.Now()
	for range 4 {
		*t = fmt.Sprint(time.Now().UnixMilli())
		captchaParams := a.getCaptchaParams(*rtk, *t)
		if captchaParams.VS == "verified" {
			log.Println("验证码已通过验证")
			return true
		}
		for captchaParams.Msg != "" {
			//log.Println("验证码已通过验证")
			log.Println(a.Http.Cookies())
			fmt.Println("清空cookie")
			log.Println("清空cookie")
			jar, _ := cookiejar.New(nil)
			a.Http.SetCookieJar(jar)
			*csrfToken = a.getRawCsrfToken()
			*rtk = a.getRTK()
			*t = fmt.Sprint(time.Now().UnixMilli())
			captchaParams = a.getCaptchaParams(*rtk, *t)
		}
		imgStream, err := a.getCaptchaImage(captchaParams.Imtk, captchaParams.Mi, captchaParams.T)
		if err != nil {
			continue
		}
		wg.Add(1)
		// 将公钥获取放在这里以节省时间，并确保公钥是新鲜的
		go a.getRsaPublicKey(ctx, wg, t, encryptedResult)
		capStartTime := time.Now()
		x := check_code.FindBestMatch(imgStream)
		log.Println("识别用时:", time.Since(capStartTime))
		verResult := a.captchaVerify(*rtk, LoginExtend, x)
		// log.Println("captcha_verify:", ver_result)
		if verResult {
			// wg.Wait()
			log.Println("验证用时:", time.Since(captchaStartTime))
			return true
		}

		fmt.Println(":( 滑块验证失败")
		log.Println(":( 滑块验证失败")
		check_code.SaveImgStream(imgStream, "fail/", "fail_"+fmt.Sprint(x)+"_"+fmt.Sprint(time.Now().UnixMilli()))
		return false // 一般来说出现验证失败是cookie问题，所以要重新登录流程而不是重试验证码
	}
	return false
}

func (a *APIClient) getKaptchaImage() {
	resp, err := a.Http.R().
		SetQueryParam("time", fmt.Sprint(time.Now().UnixMilli())).
		Get(baseCfg.KAPTCHA)
	if err != nil {
		fmt.Println(err)
		return
	}
	check_code.SaveImgStream(resp.Bytes(), "./", "kaptcha")
}

func (a *APIClient) getRawCsrfToken() string {
	// 获取CSRF令牌
	var failCount int
	var timeout int
	var csrfToken string
	for {
		// log.Println("csrf debug")
		resp, err := a.Http.R().
			//SetContext(ctx).
			//SetRetryCount(1).
			//SetQueryParam("time", fmt.Sprint(time.Now().UnixMilli())).
			//SetQueryParams(map[string]string{ // ?language=zh_CN&_t=MiniSecond
			//	"language": "zh_CN",
			//	"_t":       fmt.Sprint(time.Now().UnixMilli()),
			//}).
			Get(baseCfg.LoginIndex)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				timeout++
				fmt.Println("CSRF 请求超时", timeout, resp.Duration())
				continue
			}
			log.Println("CSRF HTTP 请求失败:", failCount, err)
			failCount++
			if failCount > 1 {
				fmt.Printf("\r%d %s", failCount, err.Error())
			}
			if failCount > 4 {
				return "" // 重新开始流程
			}
			continue
		}
		if resp.IsError() {
			failCount++
			log.Println("CSRF http:", resp.Status())
		}
		if failCount > 2 {
			fmt.Println()
		}

		// 解析 HTML 文档
		docNode, err := htmlquery.Parse(bytes.NewReader(resp.Bytes()))
		if err != nil {
			log.Println("CSRF 解析 HTML 失败:", err)
			time.Sleep(150 * time.Millisecond)
			continue
		}

		if node := htmlquery.FindOne(docNode, `//*[@id="yzm" or @name="yzm"]]`); node != nil {
			fmt.Println(a.Http.BaseURL())
			if htmlquery.SelectAttr(node, "placeholder") != "" {
				fmt.Println("不支持图形验证码")
				fmt.Println("支持滑块验证码，无验证码")
				a.getKaptchaImage()
				time.Sleep(1 * time.Second)
				os.Exit(0)
			}
		}
		//if node := htmlquery.FindOne(docNode, `//*[@id="ydType" or @name="ydType"]`); node != nil {
		//	ydType := htmlquery.SelectAttr(node, "value")
		//	fmt.Println(ydType)
		//}
		if node := htmlquery.FindOne(docNode, `//*[@id="csrftoken" or @name="csrftoken"]`); node != nil {
			csrfToken = htmlquery.SelectAttr(node, "value")
			//fmt.Println(csrfToken)
			return csrfToken
		}

		bodyStr := resp.String()
		if utils.UserIsLogin(a.Account, bodyStr) {
			return ""
		}

		if node := htmlquery.FindOne(docNode, `//*[@role="form" or @class="form-horizontal"]`); node != nil {
			action := htmlquery.SelectAttr(node, "action")
			fmt.Println(action)
			fmt.Println(a.Http.BaseURL())
			fmt.Println("请检查url填写是否有误，是否带有 /jwglxt 注意后面留空")
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println(bodyStr)
		log.Println("未找到 #csrftoken 元素或其 value 属性")
		log.Println("请检查url填写是否有误，是否带有 /jwglxt 注意后面留空")
		fmt.Println(a.Http.BaseURL())
		fmt.Println("请检查url填写是否有误，是否带有 /jwglxt 注意后面留空")
		fmt.Println("baseUrl: http?://????.????.edu.cn/jwglxt/")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (a *APIClient) getRTK() string {
	// 获取 cookie rtk
	for {
		resp, err := a.Http.R().
			SetQueryParams(map[string]string{
				"type":       "resource",
				"instanceId": "zfcaptchaLogin",
				"name":       "zfdun_captcha.js",
			}).
			Get(baseCfg.CAPTCHA)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				time.Sleep(275 * time.Millisecond)
				fmt.Println("rtk 请求超时", resp.Duration())
				continue
			} else {
				fmt.Println("rtk http:", err)
				log.Println("rtk http:", err)
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

		if resp.IsError() {
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
			//log.Println("rtk:", matches[1])
			return matches[1]
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
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("capParams 请求超时:", resp.Duration())
			} else {
				fmt.Println("\ncapParams http:", err)
			}
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
	for i := 0; i < 2; i++ {
		resp2, err := a.Http.R().
			SetRetryCount(0).
			SetTimeout(76*time.Second). // 不能睡到79秒 (76-78)
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
			return nil, fmt.Errorf("未获取到image")
		}

		return resp2.Bytes(), nil
	}
	return nil, fmt.Errorf("未获取到 image")
}

type rsaResponseData struct {
	Modulus  string `json:"modulus"`
	Exponent string `json:"exponent"`
}

func (a *APIClient) getRsaPublicKey(ctx context.Context, wg *sync.WaitGroup, t *string, enResult *string) {
	// 获取RSA公钥信息
	// 注意：公钥会经常刷新
	var jsonResult rsaResponseData
	defer wg.Done()
	for range 4 {
		resp, err := a.Http.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json, */*").
			SetResult(&jsonResult).
			SetQueryParams(map[string]string{
				"time": fmt.Sprint(time.Now().UnixMilli()),
				"_":    *t,
			}).Get(baseCfg.PublicKey)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("pubkey 超时", resp.Duration())
			} else {
				fmt.Println("pubkey 获取错误:", err)
				log.Println("pubkey HTTP 请求失败:", err)
			}
			time.Sleep(150 * time.Millisecond)
			//continue
		}
		if resp.IsError() {
			log.Println("pubkey HTTP 错误: 状态码 ", resp.Status())
			//continue
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

func (a *APIClient) captchaVerify(rtk string, LoginExtend []byte, x int) bool {
	captchaVerifyResult := check_code.GetTrackByte(x, 480)
	if captchaVerifyResult == nil {
		return false
	}
	var result captchaVerifyData
	for range 2 {
		formData := map[string]string{
			"type":       "verify",
			"rtk":        rtk,
			"time":       fmt.Sprint(time.Now().UnixMilli()),
			"mt":         base64.StdEncoding.EncodeToString(captchaVerifyResult),
			"instanceId": "zfcaptchaLogin",
			"extend":     base64.StdEncoding.EncodeToString(LoginExtend),
		}
		resp, err := a.Http.R().
			SetResult(&result).
			SetFormData(formData). // 这里不支持json
			Post(baseCfg.CAPTCHA)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("captcha_verify 超时", resp.Duration())
			} else {
				fmt.Println("captcha_verify http error:", err)
			}
			log.Println("captcha_verify HTTP 请求失败:", err)
			// fmt.Println(err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp.IsError() {
			log.Println("captcha_verify HTTP 错误: 状态码 ", resp.Status())
			fmt.Println("captcha_verify ", resp.Status())
			if resp.StatusCode() == 404 {
				return false
			}
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
	if csrf == "" || mm == "" {
		return false, CsrfEmpty
	}
	for range 6 {
		resp, err := a.Http.R().
			SetQueryParam("time", t).
			SetFormData(map[string]string{
				"csrftoken": csrf,
				"yhm":       a.Account,
				"mm":        mm,
				//"yzm":       "",
			}).Post(baseCfg.LoginIndex)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("postLogin 超时", resp.Duration())
			} else {
				fmt.Println("postLogin http error")
			}
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
			// CSRF 没必要重复获取，同cookie下是一样的
			return false, err1
		}
		if resp.StatusCode() == 302 || stat {
			//fmt.Println("postLogin", resp.Status())
			fmt.Println("登录成功")
			// 这个location 并不是很有参考意义
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

func generateLoginExtend(UserAgent string) []byte {
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
	//LoginExtend := string(jsonBytes)
	return jsonBytes
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
	for range 4 {
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
	for range 8 {
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
	for range 8 {
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
