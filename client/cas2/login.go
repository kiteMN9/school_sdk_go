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
	"school_sdk/config"
	"time"

	"github.com/antchfx/htmlquery"
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
}

func NewCas(account, password, UA string, wx bool) *Client {
	if UA == "" {
		UA = config.EdgeUA
	}
	client := resty.New()
	client.SetBaseURL("https://cas2.ycit.edu.cn/").
		SetHeader("user-agent", UA).
		SetRedirectPolicy(resty.RedirectNoPolicy())

	client.SetRetryCount(1).AddRetryConditions(resty.RetryConditionStatus5XX)

	if os.Getenv("trace") == "1" {
		client.SetTrace(true)
		//client.SetLogger()
	}
	if os.Getenv("proxy") == "1" {
		//.EnableInsecureSkipVerify()
		client.SetProxy("http://127.0.0.1:8866")
	}

	portalHttp := client.Clone(context.Background()).
		SetBaseURL("https://portal.ycit.edu.cn/").
		SetHeader("user-agent", UA).
		SetRedirectPolicy(resty.RedirectNoPolicy())

	portalHttp.SetRetryCount(1).AddRetryConditions(resty.RetryConditionStatus5XX)

	if os.Getenv("trace") == "1" {
		portalHttp.SetTrace(true)
		//client.SetLogger()
	}
	if os.Getenv("proxy") == "1" {
		portalHttp.SetProxy("http://127.0.0.1:8866")
		tls := client.TLSClientConfig()
		tls.InsecureSkipVerify = true
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

func getXpathValue(docNode *html.Node, name string) string {
	nodes := htmlquery.FindOne(docNode, `//*[@name="`+name+`"]`)
	return htmlquery.SelectAttr(nodes, "value")
}

func (c *Client) getHtml() string {
	for {
		resp, err := c.http.R().
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetRetryCount(1).
			Get("/cas/login")
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("cas getHtml 请求超时", resp.Duration())
				continue
			}
			if errors.Is(err, io.EOF) {
				fmt.Println(err)
				time.Sleep(3 * time.Second)
				continue
			} else {
				log.Println("cas getHtml 请求失败:", err)
				fmt.Println(err)
			}
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.IsStatusFailure() {
			fmt.Println(resp.Status())
			time.Sleep(2 * time.Second)
			continue
		}
		docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes()))
		//docNode, err1 := htmlquery.Parse(resp.Body)
		if err1 != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		execution := getXpathValue(docNode, "execution")
		failN := getXpathValue(docNode, "failN")
		log.Println("failN:", failN)
		if failN != "-1" && failN != "0" {
			fmt.Println("failN:", failN)
			fmt.Println("有一定的失败次数，这可能导致验证码变成必须项")
			log.Println("有一定的失败次数，这可能导致验证码变成必须项", failN)
			time.Sleep(2 * time.Second)
		}
		return execution
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
			//SetRetryCount(1).
			//SetRetryAllowNonIdempotent(true).
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetFormData(map[string]string{
				"username":    c.Account,
				"password":    encryptResult,
				"captcha":     "",
				"currentMenu": "",
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
			log.Println(location)
			log.Println("====点击上方连接可访问门户==============")

			//fmt.Println("ticketJWT:", ticketJWT)

			// 从 ticketJWT 提取 idToken 作为x-id-token
			// ticket分成三段，中间的base64解码后得到json里的idToken是结果
			idToken, err1 := utils.ExtractIDToken(ticketJWT)
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
			c.nextLoginTimeExp, c.Account = utils.ExtractExpManual(ticketJWT)
			return true
		case 200:
			fmt.Println("不成功，登录实现有问题", resp.Status())
			time.Sleep(time.Second * 12)
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
