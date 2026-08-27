package client

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"school_sdk/client/cas2"
	"school_sdk/client/config"
	"school_sdk/client/hedge"
	"school_sdk/client/internal"
	baseCfg "school_sdk/config"
	"strings"
	"time"

	"resty.dev/v3"
)

// APIClient 包装Resty客户端，动态应用配置
type APIClient struct {
	Config           *config.Data
	Name             string
	Http             *resty.Client
	hedgeC           *resty.Client
	onlyCookieMethod bool
	enableCas2       bool
	cas2Client       *cas2.Client
	lastRequestTime  time.Time
}

func baseURLLegalCheck(baseURL string) string {
	var illegal bool
	if strings.Contains(baseURL, baseCfg.LoginIndex) {
		fmt.Println("请去掉", baseCfg.LoginIndex)
		baseURL = strings.Replace(baseURL, baseCfg.LoginIndex, "", 1)
		illegal = true
	}
	if strings.Contains(baseURL, baseCfg.MENU) {
		fmt.Println("请去掉", baseCfg.MENU)
		baseURL = strings.Replace(baseURL, baseCfg.MENU, "", 1)
		illegal = true
	}
	if illegal {
		fmt.Println("需要 baseURL 不要带具体路径")
		fmt.Println(baseURL)
	}
	return baseURL
}

func NewBasicClient(baseURL string, timeout time.Duration, fCfg *config.Data) (*resty.Client, *resty.Client) {
	baseURLLegalCheck(baseURL)
	client := resty.New().
		SetRedirectPolicy(resty.RedirectNoPolicy()).
		SetBaseURL(baseURL)
	client.Client().Timeout = 50 * time.Second

	if os.Getenv("proxy") == "1" {
		client.SetProxy("http://127.0.0.1:8866")
		if tls_ := client.TLSClientConfig(); tls_ != nil {
			tls_.InsecureSkipVerify = true
			//client.SetCloseConnection(true)
		}
	}

	if os.Getenv("trace") == "1" {
		client.SetTrace(true)
	}
	if timeout < 4*time.Second {
		timeout = 4 * time.Second
	}
	client.SetTimeout(timeout) // 整个请求的超时时间
	client.SetRetryCount(3).
		AddRetryConditions(resty.RetryConditionStatus5XX)

	//refer, err := JoinURL(baseURL, baseCfg.LoginIndex)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//client.SetHeader("Referer", refer)
	client.SetHeader("user-agent", fCfg.UserAgent)

	client.AddContentDecompresser("br", internal.DecompressBrotli)

	if strings.HasPrefix(fCfg.BaseURL, "https://") {
		if transport, _ := client.HTTPTransport(); transport != nil {
			transport.IdleConnTimeout = 68 * time.Second
		}
	}

	if fCfg.Hedging {
		delay, err := time.ParseDuration(fCfg.HedgingDelay)
		if err != nil {
			delay = 16 * time.Second
			fCfg.HedgingDelay = "16s"
			fCfg.WriteConfig()
		}
		ht := hedge.NewHedging().
			SetDelay(delay).
			SetMaxRequest(3).
			SetTransport(client.Transport())
		hedgedClient := &http.Client{
			Transport: ht,
			Jar:       client.Client().Jar,
			Timeout:   50 * time.Second,
		}
		htc := resty.NewWithClient(hedgedClient).SetBaseURL(baseURL).
			SetTimeout(timeout).SetRedirectPolicy(resty.RedirectNoPolicy())
		//htc.SetHeader("Referer", refer)
		htc.SetHeader("user-agent", fCfg.UserAgent)
		htc.AddContentDecompresser("br", internal.DecompressBrotli)
		return client, htc
	}
	return client, client
}

func NewAPIClient(timeout time.Duration, cfg *config.Data, isCas2, WX bool, route string) *APIClient {
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	client, htc := NewBasicClient(cfg.BaseURL, timeout, cfg)
	if route != "" {
		cookie := &http.Cookie{ // 过 nginx有这个
			Name:  "route",
			Value: route, // 手动设置成一样的会有非常明显的挤号问题
		}
		client.CookieJar().SetCookies(u, []*http.Cookie{cookie})
	} else if strings.HasPrefix(cfg.BaseURL, "https://jwglxt.ycit") || strings.HasPrefix(cfg.BaseURL, "http://jwglxt.ycit") || strings.HasPrefix(cfg.BaseURL, "http://202.119.141") {
		routes := []string{
			//"018f9ff65252ca4f51865070844ae0be", // 慢且cas登录失败
			//"34ff17f478ebaa7e4063c9d5a95901d0", // 慢且cas登录失败
			//"425b918000ed5b18d10afb85fbbf8ec7", // 快
			//"8ed16c15842922decba77aa1ed63b61f", // 快❌
			//"c80e782f5a3340e86274809ce311b6b4", // 快

			"c8aa7be12690eaa40200741aded427f8", // 55.3428ms cas✅
			"8ed16c15842922decba77aa1ed63b61f", // 52.1706ms cas✅
		}
		if len(cfg.Routes) != 0 {
			routes = cfg.Routes
		}
		if len(cfg.Routes) != 0 {
			selected := routes[rand.Intn(len(routes))]
			cookie := &http.Cookie{ // 过 nginx有这个
				Name:  "route",
				Value: selected, // 手动设置成一样的会有非常明显的挤号问题
			}
			//log.Println(selected)
			client.CookieJar().SetCookies(u, []*http.Cookie{cookie})
		}
	}

	//transport, _ := client.HTTPTransport()
	//fmt.Println(transport.MaxConnsPerHost, transport.MaxIdleConnsPerHost, transport.MaxIdleConns)
	//transport.MaxIdleConns = 100
	//transport.MaxIdleConnsPerHost = 16 // 每个host最大空闲连接数
	//transport.MaxConnsPerHost = 30     // 每个host最大连接数

	// client.SetDebug(false) // 启用调试日志
	//client.SetLogger(&CustomLogger{})

	apiClient := &APIClient{
		Config:     cfg,
		Http:       client,
		hedgeC:     htc,
		enableCas2: isCas2 || WX,
	}

	if isCas2 || WX {
		apiClient.cas2Client = cas2.NewCas(cfg.Account, cfg.CasPasswd, cfg.UserAgent, WX, cfg)
		return apiClient
	}
	return apiClient
}

func NewClientWithCookieJar(cfg *config.Data, timeout time.Duration, jar *cookiejar.Jar) *APIClient {
	client, htc := NewBasicClient(cfg.BaseURL, timeout, cfg)
	client.SetCookieJar(jar)
	htc.SetCookieJar(jar)
	client.SetHeader("user-agent", cfg.UserAgent)
	//SetTLSFingerprintRandomized().
	//client.SetProxyURL("http://127.0.0.1:8866")
	client.SetTrace(true)

	//client.SetLogger(&CustomLogger{})

	return &APIClient{
		Config:           cfg,
		Http:             client,
		hedgeC:           htc,
		onlyCookieMethod: true,
	}
}

func JoinURL(base, endpoint string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	endpointURL, err1 := url.Parse(endpoint)
	if err1 != nil {
		return "", err1
	}
	fullURL := baseURL.ResolveReference(endpointURL)
	return fullURL.String(), nil
}

var TERM = map[int]string{0: "", 1: "3", 2: "12", 3: "16"}
