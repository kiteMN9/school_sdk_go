package cas2

// 这部分代码写的比较烂，但是能跑起来
import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http/cookiejar"
	"net/url"
	"os"
	"school_sdk/check_code"
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
		SetRedirectPolicy(resty.NoRedirectPolicy())

	client.SetRetryCount(5)

	if os.Getenv("trace") == "1" {
		client.EnableTrace()
		//client.SetLogger()
	}
	if os.Getenv("proxy") == "1" {
		//.EnableInsecureSkipVerify()
		client.SetProxy("http://127.0.0.1:8866")
		tls := client.TLSClientConfig()
		tls.InsecureSkipVerify = true
	}

	portalHttp := resty.New().
		SetBaseURL("https://portal.ycit.edu.cn/").
		SetHeader("user-agent", UA).
		SetRedirectPolicy(resty.NoRedirectPolicy())

	portalHttp.SetRetryCount(5)

	if os.Getenv("trace") == "1" {
		portalHttp.EnableTrace()
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

	check_code.SaveImgStream(c.getQrCode(), "./", "qrcode")
	if c.postLogin(encryptResult, execution) {
		//c.LoggedIn = true
		return true
	} else {
		//c.LoggedIn = false
		fmt.Println("清空cookie")
		log.Println("清空cookie")
		jar, _ := cookiejar.New(nil)
		c.http.SetCookieJar(jar)
		return false
	}
}

func getXpathValue(docNode *html.Node, name string) string {
	nodes := htmlquery.FindOne(docNode, `//*[@name="`+name+`"]`)
	return htmlquery.SelectAttr(nodes, "value")
}

func (c *Client) getHtml() string {
	for {
		resp, err := c.http.R().
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			Get("/cas/login")
		if err != nil {
			log.Println("getHtml", err)
			time.Sleep(1 * time.Second)
			continue
		}
		//htmlContent := resp.String()
		//docNode, err1 := htmlquery.Parse(strings.NewReader(htmlContent))
		docNode, err1 := htmlquery.Parse(bytes.NewReader(resp.Bytes()))
		//docNode, err1 := htmlquery.Parse(resp.Body)
		if err1 != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		execution := getXpathValue(docNode, "execution")
		failN := getXpathValue(docNode, "failN")
		log.Println("execution:", execution)
		log.Println("failN:", failN)
		if failN != "-1" && failN != "0" {
			fmt.Println("failN:", failN)
			fmt.Println("有一定的失败次数，这可能导致验证码变成必须项")
			time.Sleep(2 * time.Second)
		}
		return execution
	}
}

func (c *Client) getRsaPublicKey() string {
	publicKeyPEM := ""
	for {
		resp, err := c.http.R().
			Get("/cas/jwt/publicKey")
		if err != nil {
			log.Println("getRsaPublicKey", err)
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

		//fmt.Printf("原始密码: %s\n", c.password)

		// 加密得到 Base64 结果
		base64Result, encErr := encryptor.EncryptWithBase64(c.password)
		if encErr != nil {
			fmt.Printf("加密失败: %v\n", encErr)
			continue
		}
		encryptResult := "__RSA__" + base64Result
		//fmt.Printf("结果: %s\n", encryptResult)
		log.Println("encryptResult:", encryptResult)
		return encryptResult
	}
	//return publicKeyPEM
}

func (c *Client) getCaptchaImage() []byte {
	for range 3 { // set cookie
		resp, err := c.http.R().
			SetQueryParam("r", fmt.Sprint(time.Now().UnixMicro()/100)).
			Get("/cas/captcha.jpg")
		if err != nil {
			log.Println("getCaptchaImage:", err)
			continue
		}
		captchaImage := resp.Bytes()
		return captchaImage
	}
	return []byte{}
}

func (c *Client) getQrCode() []byte {
	for range 3 {
		// if not set cookie SESSION
		resp, err := c.http.R().
			SetQueryParam("r", fmt.Sprint(time.Now().UnixMicro()/100)).
			Get("/cas/qr/qrcode")
		if err != nil {
			log.Println("getCaptchaImage:", err)
			continue
		}
		captchaImage := resp.Bytes()
		return captchaImage
	}
	return []byte{}
}

func (c *Client) postLogin(encryptResult, execution string) bool {
	for range 5 {
		resp, err := c.http.R().
			SetQueryParam("service", "https://portal.ycit.edu.cn/?path=https://portal.ycit.edu.cn/main.html#/").
			SetFormData(map[string]string{
				"username":    c.Account,
				"password":    encryptResult,
				"captcha":     "",
				"currentMenu": "",
				"failN":       "0",
				"mfaState":    "",
				"execution":   execution,
				"_eventId":    "submit",
				"geolocation": "",
				"fpVisitorId": c.fpVisitorId,
				"submit1":     "Login1",
			}).Post("/cas/login")
		if err != nil {
			log.Println("postLogin err:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Println("cas2 postLogin:", resp.Status()) // 401失败 302成功
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
			fmt.Println("点击下方连接可访问门户==========")
			fmt.Println(location)
			fmt.Println("点击上方连接可访问门户==========")

			log.Println("点击下方连接可访问门户==========")
			log.Println(location)
			log.Println("点击上方连接可访问门户==========")

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
			fmt.Println("不成功，登录实现有问题")
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
			fmt.Println(resp.Status(), resp.String())
			log.Println(resp.Status(), resp.String())
			time.Sleep(1 * time.Second)
			continue
		}
		break
		//fmt.Println(resp.String())
	}
	return false
}
