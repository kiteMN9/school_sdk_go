package client

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"school_sdk/client/cas2"
	baseCfg "school_sdk/config"
	"strings"
	"time"

	"resty.dev/v3"
)

// APIClient 包装Resty客户端，动态应用配置
type APIClient struct {
	filename         string // 配置文件名称
	Config           *Config
	Account          string // 账号、学号
	passwd           string // 密码
	Http             *resty.Client
	onlyCookieMethod bool
	enableCas2       bool
	cas2Client       *cas2.Client
}

func NewBasicClient(baseURL string, timeout time.Duration) *resty.Client {
	client := resty.New().
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		SetBaseURL(baseURL)

	if os.Getenv("proxy") == "1" {
		client.SetProxy("http://127.0.0.1:8866")
		tls := client.TLSClientConfig()
		tls.InsecureSkipVerify = true
	}
	if os.Getenv("trace") == "1" {
		client.EnableTrace()
	}

	//client.EnableRetryDefaultConditions()
	client.SetRetryCount(0)    // resty 库有点bug一旦失败及其影响性能
	client.SetTimeout(timeout) // 整个请求的超时时间

	// 设置两次重试之间的基础等待时间
	client.SetRetryWaitTime(1129 * time.Millisecond).
		SetRetryMaxWaitTime(3 * time.Second) // 设置两次重试之间的最大等待时间

	refer, _ := JoinURL(baseURL, baseCfg.LoginIndex)
	client.SetHeader("Referer", refer)

	client.SetHeader("Accept", "*/*")

	// Add decompresser into Resty
	client.AddContentDecompresser("br", decompressBrotli)
	client.AddContentDecompresser("zstd", decompressZstd)
	return client
}

func NewAPIClient(config *Config, account, passwd string, cfgFileName string, isCas2, WX bool, casPasswd string) *APIClient {
	client := NewBasicClient(config.baseURL, config.timeout)
	client.SetHeader("user-agent", config.userAgent)
	//SetTLSFingerprintRandomized().
	//client.EnableDebugLog()

	//transport, _ := client.HTTPTransport()
	//fmt.Println(transport.MaxConnsPerHost, transport.MaxIdleConnsPerHost, transport.MaxIdleConns)
	//transport.MaxIdleConns = 100
	//transport.MaxIdleConnsPerHost = 16 // 每个host最大空闲连接数
	//transport.MaxConnsPerHost = 30     // 每个host最大连接数

	// client.SetDebug(false) // 启用调试日志
	//client.SetLogger(&CustomLogger{})

	apiClient := &APIClient{
		Config:     config,
		Http:       client,
		Account:    account,
		passwd:     passwd,
		filename:   cfgFileName,
		enableCas2: isCas2 || WX,
	}

	if isCas2 || WX {
		apiClient.cas2Client = cas2.NewCas(account, casPasswd, config.userAgent, WX)
		return apiClient
	}
	return apiClient
}

func NewClientWithCookieJar(config *Config, account string, jar *cookiejar.Jar) *APIClient {
	client := NewBasicClient(config.baseURL, config.timeout).
		SetCookieJar(jar)
	client.SetHeader("user-agent", config.userAgent)
	//SetTLSFingerprintRandomized().
	//client.SetProxyURL("http://127.0.0.1:8866")
	client.EnableTrace()

	//client.SetLogger(&CustomLogger{})

	return &APIClient{
		Config:           config,
		Http:             client,
		Account:          account,
		passwd:           "cookie!",
		filename:         "",
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

var TERM = map[int]string{1: "3", 2: "12", 3: "16"}

func (a *APIClient) CheckLogout302(resp *resty.Response) bool {
	if resp == nil {
		return false
	}
	if resp.StatusCode() == http.StatusFound {
		location := resp.Header().Get("Location")
		if strings.Contains(location, baseCfg.LoginIndex) || strings.Contains(location, a.Http.BaseURL()) {
			//println("Logout302")
			return true
		} else {
			log.Println("CheckLogout302:", resp.Header())
			fmt.Println("意料之外的错误！", resp.Header())
		}
	}
	return false
}
