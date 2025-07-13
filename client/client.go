package client

import (
	"net/http"
	"net/url"
	baseCfg "school_sdk/config"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	// "github.com/google/brotli/go/cbrotli"
	// "github.com/klauspost/compress/zstd"
)

type APIClient struct {
	config  *Config
	Account string // 账号、学号
	passwd  string // 密码
	http    *resty.Client
}

// var zstdDecoder, _ = zstd.NewReader(nil)

//var servers = []string{
//	"http://10.0.4.8",
//	"http://10.0.4.9",
//	"http://10.0.4.22",
//	"http://10.0.4.23",
//}

func NewAPIClient(config *Config, account, passwd string) *APIClient {
	transport := &http.Transport{
		DisableKeepAlives: false, // 禁用 Keep-Alive（默认false，即启用）
		// 连接池设置
		MaxIdleConns:        50,               // 全局最大空闲连接数 default 100
		MaxIdleConnsPerHost: 20,               // 对于每个主机，保持最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时，空闲连接保留时间
		ProxyConnectHeader: http.Header{
			"User-Agent": {config.userAgent},
			"Connection": {"keep-alive"},
		},
		TLSHandshakeTimeout: 11 * time.Second, // TLS 握手超时时间，default 10
		//ResponseHeaderTimeout: 25 * time.Second, // 等待响应头的超时时间
		ForceAttemptHTTP2:     true,
		ExpectContinueTimeout: 1 * time.Second,
		//Proxy: ProxyFromEnvironment,
		//DialContext: defaultTransportDialContext(&net.Dialer{
		//	Timeout:   30 * time.Second,
		//	KeepAlive: 30 * time.Second,
		//}),
	}
	client := resty.New().
		SetTransport(transport).
		SetHeader("user-agent", config.userAgent).
		SetHeader("accept", "*/*")
	// 设置初始配置
	//client.SetProxy("http://127.0.0.1:8866")
	client.SetRetryCount(5).
		SetRetryWaitTime(1743 * time.Millisecond). // 设置两次重试直接的基础等待时间
		SetRetryMaxWaitTime(config.timeout)
	//client.SetTimeout(config.GetTimeout())           // 整个请求的超时时间

	refer, _ := JoinURL(config.baseURL, baseCfg.LoginIndex)
	client.SetHeader("Referer", refer)
	client.SetBaseURL(config.baseURL)

	//client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(2))
	//client.SetRedirectPolicy(resty.NoRedirectPolicy())
	// 去掉这个就是正常的验证登录的流程
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// 返回此错误会强制停止自动重定向，并返回当前响应（如 302）
		return http.ErrUseLastResponse
	}))
	client.SetLogger(&CustomLogger{})

	return &APIClient{
		config:  config,
		http:    client,
		Account: account,
		passwd:  passwd,
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

var TERM = map[int]string{1: "3", 2: "12", 3: "16"}

func (a *APIClient) CheckLogout302(resp *resty.Response) bool {
	if resp.StatusCode() == http.StatusFound {
		location := resp.Header().Get("Location")
		if strings.Contains(location, baseCfg.LoginIndex) {
			//println("Logout302")
			return true
		}
	}
	return false
}

// 自定义日志记录器
type CustomLogger struct{}

// Implement the Log方法
// func (c *CustomLogger) Log(v ...interface{}) {
// 	log.Println(v...)
// }

// Debugf 实现了 resty.Logger 接口的 Debugf 方法
func (l *CustomLogger) Debugf(format string, v ...interface{}) {
	// log.Printf("DEBUG: "+format, v...)
}

// Errorf 实现了 resty.Logger 接口的 Errorf 方法
func (l *CustomLogger) Errorf(format string, v ...interface{}) {
	// log.Printf("ERROR: "+format, v...)
}
func (l *CustomLogger) Warnf(format string, v ...interface{}) {
	// log.Printf("ERROR: "+format, v...)
}
