package mtsm

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/go-stack/stack"
)

// client 表示指标客户端
// clientOption 是函数选项模式的选项函数类型
type clientOption func(*clientConfig)

// clientConfig 包含客户端配置选项
type clientConfig struct {
	err error
	// 基本配置
	BaseURL  string
	Username string
	Password string

	// resty 配置
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	Headers       map[string]string
	Debug         bool // 调试模式
	DisableWarn   bool // 是否禁用警告

	resty *resty.Client
}

// WithClientBaseURL 设置基本 URL
func WithClientBaseURL(baseURL string) clientOption {
	return func(c *clientConfig) {
		c.BaseURL = baseURL
	}
}

// WithClientBasicAuth 设置基本认证
func WithClientBasicAuth(username, password string) clientOption {
	return func(c *clientConfig) {
		c.Username = username
		c.Password = password
	}
}

// WithClientDebug 设置调试模式
func WithClientDebug(debug bool) clientOption {
	return func(c *clientConfig) {
		c.Debug = debug
	}
}

// WithClientDisableWarn 设置是否禁用警告
func WithClientDisableWarn(disableWarn bool) clientOption {
	return func(c *clientConfig) {
		c.DisableWarn = disableWarn
	}
}

// WithClientResty 设置 resty 客户端
func WithClientResty(r *resty.Client) clientOption {
	return func(c *clientConfig) {
		c.resty = r
	}
}

// WithClientTimeout 设置请求超时时间
func WithClientTimeout(timeout time.Duration) clientOption {
	return func(c *clientConfig) {
		c.Timeout = timeout
	}
}

// WithClientRetry 设置重试策略
func WithClientRetry(count int, waitTime time.Duration) clientOption {
	return func(c *clientConfig) {
		c.RetryCount = count
		c.RetryWaitTime = waitTime
	}
}

// WithClientHeader 添加请求头
func WithClientHeader(key, value string) clientOption {
	return func(c *clientConfig) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		c.Headers[key] = value
	}
}

// WithClient 从 url 中自动解析用户名和密码,并设置 BaseURL
func WithClient(URLScheme string) clientOption {
	return func(c *clientConfig) {
		if URLScheme == "" {
			c.err = fmt.Errorf("%v [%#v]", "URLScheme is empty", stack.Caller(0))
			return
		}
		urlParse, err := url.Parse(URLScheme)
		if err != nil {
			c.err = fmt.Errorf("%v [%#v]", err, stack.Caller(0))
			return
		}
		c.BaseURL = fmt.Sprintf("%s://%s%s%s", urlParse.Scheme, urlParse.Host, urlParse.Path, urlParse.RawQuery)
		c.Username = urlParse.User.Username()
		c.Password, _ = urlParse.User.Password()
	}
}

// NewClient 创建一个新的客户端实例
func NewClient(UrlScheme string, opts ...clientOption) (*Client, error) {
	// 默认配置
	config := &clientConfig{
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryWaitTime: time.Second,
		Debug:         false,
		DisableWarn:   true,
	}

	// 应用选项
	opts = append(opts, WithClient(UrlScheme))
	for _, opt := range opts {
		opt(config)
	}

	if config.err != nil {
		return nil, config.err
	}

	if config.resty == nil {
		config.resty = resty.New()
		// 创建 HTTP 客户端
		r := config.resty
		r.SetTimeout(config.Timeout)
		r.SetRetryCount(config.RetryCount)
		r.SetRetryWaitTime(config.RetryWaitTime)
		r.SetBaseURL(config.BaseURL)
		r.SetDebug(config.Debug)
		r.SetDisableWarn(config.DisableWarn)

		// 添加基本认证（如果提供）
		if config.Username != "" && config.Password != "" {
			r.SetBasicAuth(config.Username, config.Password)
		}
		// 添加请求头
		for k, v := range config.Headers {
			r.SetHeader(k, v)
		}
	}

	t := &Client{
		config: config,
	}
	return t, nil
}

type Client struct {
	config *clientConfig
}

// ApiQuery 执行 Prometheus 查询
func (c *Client) ApiQuery(params map[string]string) (*api_query, error) {
	t := &api_query{
		client: c,
	}
	err := t.request(params)
	return t, err
}

// ApiQueryRange 执行 Prometheus 范围查询
func (c *Client) ApiQueryRange(params map[string]string) (*api_query_range, error) {
	t := &api_query_range{
		client: c,
	}
	err := t.request(params)
	return t, err
}

// ApiLabelValues 执行 Prometheus 标签值查询
func (c *Client) ApiLabelValues(labelName string, params map[string]string) (*api_label_values, error) {
	t := &api_label_values{
		client: c,
	}
	err := t.request(labelName, params)
	return t, err
}

// ApiDeleteSeries 执行 Prometheus 删除系列
//
// 参数示例:
//   - match[]={__name__="m_250312_test_2148"}
//   - match[]={__name__=~"m_250312_test_.*"}
func (c *Client) ApiDeleteSeries(body string) (*api_delete_series, error) {
	t := &api_delete_series{
		client: c,
	}
	err := t.request(body)
	return t, err
}

// Close 关闭客户端并释放资源
func (c *Client) Close() {
	// 关闭底层连接
	c.config.resty.SetCloseConnection(true)
}

func (t *Client) GetConfig() *clientConfig {
	return t.config
}

func (t *Client) GetClient() *Client {
	return t
}
