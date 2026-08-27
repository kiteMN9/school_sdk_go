package cas2

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http/cookiejar"
	"net/url"
	"os"
	"school_sdk/client/cas2/utils"
	"school_sdk/client/config"
	"school_sdk/client/internal"
	bcfg "school_sdk/config"
	"strings"
	"time"

	"golang.org/x/net/html"
	"resty.dev/v3"
)

type Client struct {
	Account     string
	password    string
	fpVisitorId string
	http        *resty.Client
	portalHttp  *resty.Client
	//LoggedIn      bool
	enableWxLogin    bool
	nextLoginTimeExp time.Time
	fCfg             *config.Data
}

func NewCas(account, password, UA string, wx bool, fCfg *config.Data) *Client {
	if UA == "" {
		UA = bcfg.EdgeUA
	}
	client := resty.New()
	client.SetBaseURL("https://cas2.ycit.edu.cn/").
		SetHeader("user-agent", UA).
		SetRedirectPolicy(resty.RedirectNoPolicy())
	client.AddContentDecompresser("br", internal.DecompressBrotli)

	client.SetRetryCount(1).AddRetryConditions(resty.RetryConditionStatus5XX)

	if os.Getenv("trace") == "1" {
		client.SetTrace(true)
		//client.SetLogger()
	}
	if os.Getenv("proxy") == "1" {
		//.EnableInsecureSkipVerify()
		client.SetProxy("http://127.0.0.1:8866")
		tls_ := client.TLSClientConfig()
		tls_.InsecureSkipVerify = true
	}

	// 共享transport
	portalHttp := client.Clone(context.Background()).
		SetBaseURL("https://portal.ycit.edu.cn/")

	if fCfg.TicketJWT != "" {
		idToken, nextLoginTimeExp, Account, err1 := utils.ExtractIDToken(fCfg.TicketJWT)
		if err1 == nil && time.Now().Before(nextLoginTimeExp) {
			portalHttp.SetHeader("x-id-token", idToken)
			portalHttp.SetHeader("x-device-info", "PC")
			portalHttp.SetHeader("x-terminal-info", "PC")
			portalHttp.SetHeader("cookie", "isLogin=true")
			hash := md5.Sum([]byte(Account + "salt354waragthaswrg"))
			md5Str := hex.EncodeToString(hash[:])
			return &Client{
				Account:          account,
				password:         password,
				fpVisitorId:      md5Str, // fingerprint
				http:             client,
				portalHttp:       portalHttp,
				enableWxLogin:    wx,
				nextLoginTimeExp: nextLoginTimeExp,
				fCfg:             fCfg,
			}
		}
	}

	hash := md5.Sum([]byte(account + "salt354waragthaswrg"))
	md5Str := hex.EncodeToString(hash[:])
	//fmt.Println("MD5:", md5Str)

	return &Client{
		Account:       account,
		password:      password,
		fpVisitorId:   md5Str, // fingerprint
		http:          client,
		portalHttp:    portalHttp,
		enableWxLogin: wx,
		fCfg:          fCfg,
	}
}

func (c *Client) Login() bool {
	if c.netCheckIdToken() {
		return true
	}
	if c.enableWxLogin {
		return c.WXLogin()
	}
	execution := c.getHtml()
	//fmt.Println(execution)
	encryptResult := c.getRsaPublicKey()

	//check_code.SaveImgStream(c.getQrCode(), "./", "qrcode")
	if c.postLogin(encryptResult, execution) {
		//c.LoggedIn = true
		return true
	}

	//c.LoggedIn = false
	fmt.Println("清空cookie")
	log.Println("清空cookie")
	jar, _ := cookiejar.New(nil)
	c.http.SetCookieJar(jar)
	return false
}

func extractLoginParams(body io.Reader) (execution, failN string, err error) {
	tokenizer := html.NewTokenizer(body)

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// 遇到错误（如 EOF）则结束
			if tokenizer.Err() == io.EOF {
				return execution, failN, nil
			}
			return execution, failN, tokenizer.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			// 只处理 <input> 标签（自闭合或开始标签）
			tagName, _ := tokenizer.TagName()
			if !bytes.Equal(tagName, []byte("input")) {
				continue
			}
			// 遍历属性
			var name, value string
			for {
				key, val, moreAttr := tokenizer.TagAttr()
				if bytes.Equal(key, []byte("name")) {
					name = string(val)
				} else if bytes.Equal(key, []byte("value")) {
					value = string(val)
				}
				if !moreAttr {
					break
				}
			}
			if name == "execution" {
				execution = value
			} else if name == "failN" {
				failN = value
			}
			// 如果两个值都找到了，可以提前结束（但需注意 tokenizer 可能还有后续，但我们可以返回）
			if execution != "" && failN != "" {
				return execution, failN, nil
			}
		}
	}
}

func (c *Client) getHtml() string {
	for {
		resp, err := c.http.R().
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetRetryCount(1).
			SetResponseDoNotParse(true).
			Get("/cas/login")

		if err != nil {
			// 错误处理（注意 resp 可能为 nil，此时不能调用 resp.Body.Close）
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("cas getHtml 请求超时", resp.Duration())
				continue
			}
			if errors.Is(err, io.EOF) {
				fmt.Println(err)
				time.Sleep(3 * time.Second)
				continue
			}
			log.Println("cas getHtml 请求失败:", err)
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		execution, failN, parseErr := extractLoginParams(resp.Body)
		resp.Body.Close() // 注意：提前 return 前会关闭
		if parseErr != nil {
			fmt.Println(parseErr)
			continue
		}

		if failN != "-1" && failN != "0" {
			fmt.Println("failN:", failN, "，有一定失败次数")
			time.Sleep(2 * time.Second)
		}
		if execution != "" {
			//fmt.Println(time.Since(start))
			return execution
		}
		// 如果没有拿到 execution，继续循环
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) getRsaPublicKey() string {
	publicKeyPEM := ""
	for {
		resp, err := c.http.R().
			SetRetryCount(1).
			Get("/cas/jwt/publicKey")
		if err != nil {
			log.Println("getRsaPublicKey", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.IsStatusFailure() {
			log.Println("getRsaPublicKey", resp.Status())
			time.Sleep(1 * time.Second)
			continue
		}
		publicKeyPEM = resp.String()
		//fmt.Println(publicKeyPEM)
		// 从 PEM 创建加密器
		encryptor, rsaDecErr := NewRSAEncryptorFromPEM(publicKeyPEM)
		if rsaDecErr != nil {
			fmt.Printf("创建加密器失败: %v\n", rsaDecErr)
			log.Printf("创建加密器失败: %v\n", rsaDecErr)
			time.Sleep(1 * time.Second)
			continue
		}

		// 加密得到 Base64 结果
		base64Result, encErr := encryptor.EncryptWithBase64(c.password)
		if encErr != nil {
			fmt.Printf("加密失败: %v\n", encErr)
			continue
		}
		encryptResult := "__RSA__" + base64Result
		//fmt.Printf("结果: %s\n", encryptResult)
		//log.Println("encryptResult:", encryptResult)
		return encryptResult
	}
}

func (c *Client) getCaptchaImage() []byte {
	// set cookie
	resp, err := c.http.R().
		SetQueryParam("r", fmt.Sprint(time.Now().UnixMicro()/100)).
		SetRetryCount(2).
		Get("/cas/captcha.jpg")
	if err != nil {
		log.Println("getCaptchaImage:", err)
		return []byte{}
	}
	captchaImage := resp.Bytes()
	return captchaImage
}

func (c *Client) getQrCode() []byte {
	// if not set cookie SESSION
	resp, err := c.http.R().
		SetQueryParam("r", fmt.Sprint(time.Now().UnixMicro()/100)).
		SetRetryCount(2).
		Get("/cas/qr/qrcode")
	if err != nil {
		log.Println("getCaptchaImage:", err)
		return []byte{}
	}
	captchaImage := resp.Bytes()
	return captchaImage
}

func (c *Client) postLogin(encryptResult, execution string) bool {
	for range 5 {
		resp, err := c.http.R().
			SetRetryCount(1).
			SetRetryAllowNonIdempotent(true).
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetFormData(map[string]string{
				"username":    c.Account,
				"password":    encryptResult,
				"captcha":     "",
				"currentMenu": "1",
				"failN":       "0",
				"mfaState":    "",
				"execution":   execution,
				"_eventId":    "submit", // submitPasswordlessToken
				"geolocation": "",
				"fpVisitorId": c.fpVisitorId,
				"submit1":     "Login1",
			}).Post("/cas/login")
		if err != nil {
			fmt.Println("postLogin error:", err)
			log.Println("postLogin err:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		//log.Println("cas2 postLogin:", resp.Status()) // 401失败 302成功
		switch resp.StatusCode() {
		case 302:
			location := resp.Header().Get("Location")
			if location == "" {
				log.Fatal("location is null")
			}

			// 解析 location
			parsedUrl, _ := url.Parse(location)
			query := parsedUrl.Query()
			// 提取 ticket
			ticketJWT := query.Get("ticket")
			if ticketJWT == "" {
				log.Fatal("ticketJWT is null")
			}

			fmt.Println("cas2登录成功")
			fmt.Println("====点击下方连接可访问门户=============")
			fmt.Println(location)
			fmt.Println("====点击上方连接可访问门户=============")

			log.Println("====点击下方连接可访问门户==============")
			log.Println("\n", location)
			log.Println("====点击上方连接可访问门户==============")

			//fmt.Println("ticketJWT:", ticketJWT)

			// 从 ticketJWT 提取 idToken 作为x-id-token
			// ticket分成三段，中间的base64解码后得到json里的idToken是结果
			var idToken string
			var err1 error
			idToken, c.nextLoginTimeExp, c.Account, err1 = utils.ExtractIDToken(ticketJWT)
			if err1 != nil {
				fmt.Printf("错误: %v\n", err1)
				log.Println("ticketJWT:", ticketJWT)
				log.Println("ticket解析失败:", err1)
				return false
			}
			// portal header
			c.portalHttp.SetHeader("x-id-token", idToken)
			c.portalHttp.SetHeader("x-device-info", "PC")
			c.portalHttp.SetHeader("x-terminal-info", "PC")
			c.portalHttp.SetHeader("cookie", "isLogin=true")
			c.fCfg.TicketJWT = ticketJWT
			c.fCfg.WriteConfig()
			return true
		case 200:
			if strings.Contains(resp.String(), "这个账户已经被锁住了。") {
				fmt.Println("这个账户已经被锁住了。")
			} else {
				fmt.Println("不成功，登录实现有问题", resp.Status())
				log.Println("不成功，登录实现有问题", resp.Status(), resp.String())
				time.Sleep(time.Second * 12)
			}
		case 401:
			fmt.Println("账户或密码错误？")
			time.Sleep(3 * time.Second)
			panic("账户或密码错误")
		case 500:
			log.Println("postLogin status:", resp.Status())
			log.Println(resp.String())
			time.Sleep(2 * time.Second)
			return false
		default:
			fmt.Println("cas2 postLogin:", resp.Status(), resp.String())
			log.Println("cas2 postLogin:", resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		break
		//fmt.Println(resp.String())
	}
	return false
}
