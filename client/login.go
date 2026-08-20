package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"school_sdk/client/rsa"
	"strconv"
	"strings"
	"sync"
	"time"

	"school_sdk/check_code"
	baseCfg "school_sdk/config"
	"school_sdk/utils"

	"github.com/PuerkitoBio/goquery"
)

var ExistVerify = fmt.Errorf("请先滑动图片进行验证！")
var InputYzmErr = fmt.Errorf("验证码输入错误！")
var IncorrectPassword = fmt.Errorf("用户名或密码不正确，请重新输入！")
var CsrfNotExist = fmt.Errorf("CSRF not exist")
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
	if a.Login() {
		lastSuccessTime = time.Now()
		log.Println("重新登录用时:", time.Since(reStartTime))
		return true
	}
	return false
}

func (a *APIClient) Login() bool {
	if a.enableCas2 {
		if a.cas2LoginCtl() {
			if a.cas2Client.Account != a.Config.Account {
				a.Config.Account = a.cas2Client.Account
			}
			return true
		}
		return false
	}
	var LoginExtend = generateLoginExtend(a.Config.UserAgent)

	for range 15 {
		reqTime := strconv.FormatInt(time.Now().UnixMilli(), 10)
		csrfToken, yzm, stat := a.getRawCsrfToken()
		if stat {
			return true
		}
		if yzm {
			if a.kaptchaLogin(csrfToken, reqTime) {
				return true
			}
			continue
		}
		if a.Config.ExistVerify {
			// if verify_type
			if a.getCaptchaLogin(LoginExtend, csrfToken, reqTime) {
				return true
			}
			fmt.Println("重新开始登录流程")
			continue
		} else {
			var wg sync.WaitGroup
			var encryptedResult string
			//Eb := strconv.FormatInt(time.Now().UnixMilli(), 10)
			wg.Add(1)
			go a.getRsaPublicKey(context.TODO(), &wg, &reqTime, &encryptedResult)
			wg.Wait()

			//csrfToken = a.getRawCsrfToken()
			stat_, err := a.postLogin(csrfToken, reqTime, encryptedResult, "")
			if errors.Is(err, ExistVerify) {
				a.Config.UpdateConfigUserInfo(true)
				continue
			}
			if errors.Is(err, IncorrectPassword) {
				a.Config.SetConfigUserInfo(nil)
				continue
			}
			if errors.Is(err, CsrfNotExist) {
				fmt.Println("未获取到CSRF")
			}
			return stat_
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

	//Eb := strconv.FormatInt(time.Now().UnixMilli(), 10)
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

		stat, err := a.postLogin(csrfToken, reqTime, encryptedResult, "")
		if errors.Is(err, ExistVerify) {
			if a.Config.ExistVerify {
				log.Println("重试验证码")
				continue
			}
			a.Config.ExistVerify = true
			return false
		}
		if errors.Is(err, IncorrectPassword) {
			a.Config.SetConfigUserInfo(nil)
			//a.passwd = cfg.Passwd
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
		*t = strconv.FormatInt(time.Now().UnixMilli(), 10)
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
			u, _ := url.Parse(a.Http.BaseURL())
			a.Http.CookieJar().SetCookies(u, []*http.Cookie{})
			*csrfToken, _, _ = a.getRawCsrfToken()
			*rtk = a.getRTK()
			*t = strconv.FormatInt(time.Now().UnixMilli(), 10)
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
		check_code.SaveImgStream(imgStream, "fail/", "fail_"+strconv.Itoa(x)+"_"+strconv.FormatInt(time.Now().UnixMilli(), 10))
		return false // 一般来说出现验证失败是cookie问题，所以要重新登录流程而不是重试验证码
	}
	return false
}

func (a *APIClient) kaptchaLogin(csrfToken, reqTime string) bool {
	var encryptedResult string
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	// 将公钥获取放在这里以节省时间，并确保公钥是新鲜的
	go a.getRsaPublicKey(ctx, &wg, &reqTime, &encryptedResult)
	for range 3 {
		yzm := a.getKaptchaImage()
		if yzm == "" {
			continue
		}
		wg.Wait()
		stat, err := a.postLogin(csrfToken, reqTime, encryptedResult, yzm) // mm 似乎不用刷新
		if errors.Is(err, InputYzmErr) {
			fmt.Println("验证码输入错误！")
			continue
		}
		if errors.Is(err, IncorrectPassword) {
			a.Config.SetConfigUserInfo(nil)
			continue
		} else {
			return stat
		}
	}
	return false
}

func (a *APIClient) getKaptchaImage() string {
	resp, err := a.Http.R().
		SetQueryParam("time", strconv.FormatInt(time.Now().UnixMilli(), 10)).
		Get(baseCfg.KAPTCHA)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	//if err := check_code.DisplayCaptcha(resp.Bytes()); err != nil {
	//	fmt.Println(err)
	//	return ""
	//}
	check_code.SaveImgStream(resp.Bytes(), "./", "kaptcha")
	fmt.Println("请查看 kaptcha.png")
	input, err := utils.UserInputWithSigInt("输入验证码:")
	if err != nil {
		return ""
	}
	errF := os.Remove("./kaptcha.png")
	if errF != nil {
		fmt.Println(errF)
	}
	//fmt.Println("input:", input)
	return input
}

func (a *APIClient) getRawCsrfToken() (string, bool, bool) {
	// 获取CSRF令牌
	var failCount int
	var timeout int
	var csrfToken string
	var yzm bool
	var exists bool
	for {
		// log.Println("csrf debug")
		resp, err := a.hedgeC.R().
			//SetContext(ctx).
			//SetRetryCount(1).
			//SetQueryParam("time", strconv.FormatInt(time.Now().UnixMilli(), 10)).
			//SetQueryParams(map[string]string{ // ?language=zh_CN&_t=MiniSecond
			//	"language": "zh_CN",
			//	"_t":       strconv.FormatInt(time.Now().UnixMilli(), 10),
			//}).
			Get(baseCfg.LoginIndex)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				timeout++
				fmt.Println("CSRF 请求超时", timeout, resp.Duration())
				continue
			}
			if errors.Is(err, io.EOF) {
				fmt.Println(err)
				time.Sleep(3 * time.Second)
				continue
			} else {
				log.Println("CSRF HTTP 请求失败:", failCount, err)
				failCount++
			}

			fmt.Printf("\r%d %s\n", failCount, err.Error())
			time.Sleep(600 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			failCount++
			if resp.StatusCode() == 404 {
				fmt.Println("url:", a.Http.BaseURL())
				fmt.Println("404, url 填的有问题吧，是不是少了 /jwglxt 或者多了")
				log.Println("404, url 填的有问题吧")
				time.Sleep(4 * time.Second)
				continue
			}
			if resp.StatusCode() == 403 {
				fmt.Println("url:", a.Http.BaseURL())
			}
			log.Println("CSRF http:", resp.Status())
			fmt.Println("CSRF http:", resp.Status())
			time.Sleep(1 * time.Second)
			continue
		}

		if resp.IsStatusSuccess() {
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Bytes()))
			if err != nil {
				log.Println("CSRF 解析 HTML 失败:", err)
				time.Sleep(150 * time.Millisecond)
				continue
			}

			if doc.Find("#yzmDiv").Text() != "" {
				yzm = true
			}

			// 使用 CSS 选择器提取元素属性 "input#csrftoken"
			csrfToken, exists = doc.Find("input#csrftoken").Attr("value")
			if exists {
				return csrfToken, yzm, false
			}
			if utils.UserIsLogin(a.Config.Account, resp.String()) {
				return "nil", yzm, false
			}
			fmt.Println("未找到 #csrftoken 元素或其 value 属性")
			log.Println("未找到 #csrftoken 元素或其 value 属性")
			log.Println(resp.String())
		}

		if resp.StatusCode() == 302 {
			if strings.Contains(resp.Header().Get("Location"), baseCfg.MENU) {
				return "", yzm, true
			}
			fmt.Println(resp.Header().Get("Location"))
			return "", yzm, false
		}
		time.Sleep(1 * time.Second)
		continue
	}
}

func (a *APIClient) getRTK() string {
	// 获取 cookie rtk
	for {
		resp, err := a.hedgeC.R().
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
			fmt.Println(a.Http.BaseURL())
			fmt.Println("404, url 填的有问题，是不是少了 /jwglxt")
			fmt.Println("请填写baseURL，后面的部分如:" + baseCfg.LoginIndex + " 是不必要的")
			log.Println("404, url 填的有问题")
			time.Sleep(4 * time.Second)
			continue
		}

		if resp.IsStatusFailure() {
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
	T      int64  `json:"t"`
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
			//SetResult(&jsonResult).
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
			log.Println("getCaptchaParams:", err)
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println(resp.Status())
		}
		if resp.ResultError() != nil {
			log.Println(resp.ResultError())
		}
		if err := json.Unmarshal(resp.Bytes(), &jsonResult); err != nil {
			fmt.Println(err)
			log.Println(err, resp.String())
			time.Sleep(150 * time.Millisecond)
			continue
		}
		if jsonResult.Msg != "" {
			fmt.Println(jsonResult.Msg)
		}

		return jsonResult
	}
}

var noImage = fmt.Errorf("未获取到 image")

func (a *APIClient) getCaptchaImage(imtk, id string, T int64) ([]byte, error) {
	resp, err := a.Http.R().
		SetRetryCount(1).
		SetTimeout(31*time.Second). // 不能睡到79秒 (76-78)
		SetHeader("Accept", "image/*,*/*").
		SetQueryParams(map[string]string{
			"type":       "image",
			"id":         id,
			"imtk":       imtk,
			"t":          strconv.FormatInt(T, 10),
			"instanceId": "zfcaptchaLogin",
		}).Get(baseCfg.CAPTCHA)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		fmt.Println("get image http error")
		log.Println(err)
		time.Sleep(150 * time.Millisecond)
		return nil, noImage
	}
	if resp.StatusCode() == 404 { // 过期了，重试也没用
		return nil, noImage
	}
	if resp.IsStatusFailure() {
		return nil, noImage
	}

	if len(resp.Bytes()) == 0 {
		log.Println("未获取到 image")
		return nil, noImage
	}

	return resp.Bytes(), nil
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
	for range 3 {
		resp, err := a.Http.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json, */*").
			//SetResult(&jsonResult).
			SetQueryParams(map[string]string{
				"time": strconv.FormatInt(time.Now().UnixMilli(), 10),
				"_":    *t,
			}).SetRetryCount(1).
			Get(baseCfg.PublicKey)
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
			//continue
		}
		if resp.IsStatusFailure() {
			log.Println("pubkey HTTP 错误: 状态码 ", resp.Status())
			//continue
		}
		if err := json.Unmarshal(resp.Bytes(), &jsonResult); err != nil {
			fmt.Println(err)
			log.Println(err, resp.String())
		}
		if jsonResult.Modulus == "" || jsonResult.Exponent == "" {
			log.Println("pubkey 获取错误:", resp.Status(), resp.String(), *t)
			*t = strconv.FormatInt(time.Now().UnixMilli(), 10)
			time.Sleep(70 * time.Millisecond)
			continue
		}
		*enResult, err = rsa.EncryptRsa(jsonResult.Modulus, jsonResult.Exponent, a.Config.Passwd)
		if err != nil {
			*t = strconv.FormatInt(time.Now().UnixMilli(), 10)
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
			"time":       strconv.FormatInt(time.Now().UnixMilli(), 10),
			"mt":         base64.StdEncoding.EncodeToString(captchaVerifyResult),
			"instanceId": "zfcaptchaLogin",
			"extend":     base64.StdEncoding.EncodeToString(LoginExtend),
		}
		resp, err := a.Http.R().
			//SetResult(&result).
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
		if resp.IsStatusFailure() {
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
		if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
			fmt.Println(err)
			log.Println(err, resp.String())
			continue
		}

		if result.VS == "verified" && result.Status == "success" {
			return true
		} else if result.VS == "not_verify" {
			return false
		}
	}
	return false
}

func (a *APIClient) postLogin(csrf, t, mm, yzm string) (bool, error) {
	// fmt.Println("postLogin sleep 300")
	// time.Sleep(300 * time.Second)
	if csrf == "nil" || mm == "" {
		return false, CsrfNotExist
	}
	for range 6 {
		formData := map[string]string{
			"csrftoken": csrf, // 某些系统csrf就是空的
			"yhm":       a.Config.Account,
			"mm":        mm, // TODO:使用后是否需要刷新公钥？
		}
		if len(yzm) != 0 {
			formData["yzm"] = yzm
		}
		resp, err := a.Http.R().
			SetQueryParam("time", t).
			SetFormData(formData).Post(baseCfg.LoginIndex)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("postLogin 超时", resp.Duration())
			} else {
				fmt.Println("postLogin http error")
			}
			log.Println("postLogin 请求失败:", err)
			fmt.Println(err)
			time.Sleep(170 * time.Millisecond)
			continue
		}
		//log.Println()
		if resp.IsStatusFailure() {
			log.Println("postLogin HTTP 错误: 状态码 ", resp.Status())
			continue
		}
		//log.Println(resp.String())
		stat, err1 := isLogin(a.Config.Account, resp.String())
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
	accountPattern := `value="` + regexp.QuoteMeta(account) + `"`
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

	if strings.Contains(errMsg, "验证码输入错误！") {
		log.Println(errMsg)
		return false, InputYzmErr
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
	_, after, ok := strings.Cut(UserAgent, "/")
	modifiedUserAgent := UserAgent

	if ok {
		// 截取第一个 '/' 之后的内容
		modifiedUserAgent = after
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

func (a *APIClient) cas2LoginCtl() bool {
	log.Println("cas2Login=======")
	if !a.cas2Client.Login() {
		return false
	}
	//if !a.cas2Client.GetJwCookie() {
	//	return false
	//}
	location := a.ssoLogin()
	if location == "" {
		return false
	}
	location = a.cas2Client.GetJwCookie2(location) // https://jwglxt.ycit.edu.cn/ticketlogin?uid=&timestamp=&verify=
	if location == "" {
		return false
	}
	return a.ssoTicketLogin(location)
}

func (a *APIClient) ssoLogin() string {
	log.Println("ssoLogin=======")
	for range 8 {
		resp, err := a.Http.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			SetRetryCount(1).
			Get("https://jwglxt.ycit.edu.cn/sso/hnyyxyiotlogin")
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println(err)
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Println("ssoLogin:", err)
			log.Println("ssoLogin:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode() != 302 {
			fmt.Println("sso/hnyyxyiotlogin:", resp.Status())
			log.Println("sso/hnyyxyiotlogin:", resp.Status())
			time.Sleep(2 * time.Second)
			continue
		}
		location := resp.Header().Get("Location")
		log.Println(location) // https://cas2.ycit.edu.cn/cas/login?service=http://jwglxt.ycit.edu.cn/sso/hnyyxyiotlogin?targetUrl={base64}aHR0cDovL2p3Z2x4dC55Y2l0LmVkdS5jbi9zc28vc3NvL2luZGV4LmpzcA==
		return location
	}
	return ""
}

func (a *APIClient) ssoTicketLogin(location string) bool {
	if location == "" {
		log.Fatal("GetJwCookie2 location:", location)
	}
	var location2 string
	for range 8 {
		resp, err := a.hedgeC.R().
			SetRetryCount(1).
			Get(location) // https://jwglxt.ycit.edu.cn/sso/hnyyxyiotlogin?targetUrl={base64}aHR0cDovL2p3Z2x4dC55Y2l0LmVkdS5jbi9zc28vc3NvL2luZGV4LmpzcA==&ticket=ST-529025-5R1TqFz
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("ssoTicketLogin EOF")
			}
			fmt.Println(err)
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		if resp.StatusCode() == 301 {
			location = resp.Header().Get("Location")
			continue
		}
		if resp.StatusCode() != 302 {
			fmt.Println(resp.Status())
			log.Println("resp2:", resp.Status(), resp.Header(), resp.String())
			time.Sleep(2 * time.Second)
			continue
		}
		location2 = resp.Header().Get("Location") // https://jwglxt.ycit.edu.cn/ticketlogin?uid=&timestamp=&verify=
		if location2 == "" {
			log.Println("location:", location2)
			continue
		}
		log.Println("location2:", location2)
		break
	}

	if location2 == "" {
		a.cookie()
		return false
	}
	location2 = strings.Replace(location2, "http://", "https://", -1)

	for range 6 {
		resp, err := a.hedgeC.R().
			SetHeader("Referer", "https://portal.ycit.edu.cn/main.html").
			SetRetryCount(1).
			Get(location2)
		if err != nil {
			fmt.Println(err)
			log.Println("ssoTicketLogin:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode() == 301 {
			location2 = resp.Header().Get("Location")
			continue
		}
		if resp.StatusCode() != 302 {
			fmt.Println("ssoTicketLogin:", resp.Status())
			log.Println("ssoTicketLogin:", resp.Status())
			time.Sleep(2 * time.Second)
			continue
		}
		location2 = resp.Header().Get("Location") // /xtgl/login_slogin.html
		if location2 == "" {
			fmt.Println("location为空", resp.Status())
			time.Sleep(1 * time.Second)
			continue
		}
		if strings.Contains(location2, baseCfg.LoginIndex) {
			continue
		}
		if strings.Contains(location2, baseCfg.MENU) { // /xtgl/index_initMenu.html?jsdm=xs&_t=1785043153113&echarts=1
			// c80e782f5a3340e86274809ce311b6b4 1
			// 425b918000ed5b18d10afb85fbbf8ec7 1
			// 018f9ff65252ca4f51865070844ae0be ❌
			// 34ff17f478ebaa7e4063c9d5a95901d0 ❌
			fmt.Println("登录成功")
			return true
		}
		break
	}
	a.cookie()
	return false
}
