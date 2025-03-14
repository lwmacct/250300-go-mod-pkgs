package mloki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/lwmacct/250109-m-bdflow/module/v00/mlog"
)

type Client struct {
	Conf  *clientConf
	resty *resty.Client
}

// clientConf 客户端配置
type clientConf struct {
	Err error
}

// clientOpts 客户端选项
type clientOpts func(*Client)

// QueryRange 查询范围
func (t *Client) QueryRange(queryParams map[string]string) (*TsQueryRange, error) {
	resp := &TsQueryRange{}
	_, err := t.resty.R().
		SetQueryParams(queryParams).
		SetResult(resp).
		Get("/loki/api/v1/query_range")
	return resp, err
}

// Tail 流式查询日志，使用 WebSocket 连接
// 返回一个接收通道，客户端可以从该通道持续接收日志数据
// 同时返回一个用于取消订阅的函数和可能的错误
func (t *Client) Tail(queryParams map[string]string) (chan *TsTail, func(), error) {
	// 创建数据通道用于传递结果
	dataCh := make(chan *TsTail, 10000)

	// 创建上下文，用于取消请求
	ctx, cancel := context.WithCancel(context.Background())

	// 格式化查询参数
	// 确保start参数格式正确
	if _, ok := queryParams["start"]; !ok {
		// 如果没有提供start参数，默认设置为1小时前
		queryParams["start"] = fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).Unix())
	}

	// 确保limit参数存在
	if _, ok := queryParams["limit"]; !ok {
		queryParams["limit"] = "100"
	}

	// 构建WebSocket URL
	baseURL := t.resty.BaseURL
	// 将http或https转换为ws或wss
	wsURL := baseURL
	if len(baseURL) > 5 && baseURL[:5] == "https" {
		wsURL = "wss" + baseURL[5:]
	} else if len(baseURL) > 4 && baseURL[:4] == "http" {
		wsURL = "ws" + baseURL[4:]
	}

	u, err := url.Parse(wsURL + "/loki/api/v1/tail")
	if err != nil {
		return nil, cancel, fmt.Errorf("构建WebSocket URL失败: %v", err)
	}

	// 添加查询参数
	q := u.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	// 设置WebSocket连接的HTTP头
	header := http.Header{}
	header.Add("Sec-WebSocket-Protocol", "json")

	// 创建WebSocket拨号器
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	// 启动goroutine处理WebSocket连接
	go func() {
		defer close(dataCh)
		// 连接WebSocket
		conn, resp, err := dialer.DialContext(ctx, u.String(), header)
		if err != nil {
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				mlog.Error(mlog.H{
					"msg":        "WebSocket连接失败",
					"err":        err.Error(),
					"statusCode": resp.StatusCode,
					"body":       string(body),
				})
			} else {
				mlog.Error(mlog.H{
					"msg": "WebSocket连接失败",
					"err": err.Error(),
				})
			}
			return
		}
		defer conn.Close()

		mlog.Info(mlog.H{
			"msg": "WebSocket连接成功",
		})

		// 设置处理函数，当ctx被取消时关闭连接
		go func() {
			<-ctx.Done()
			// 发送关闭消息
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			)
			conn.Close()
			mlog.Info(mlog.H{"msg": "WebSocket连接已关闭 (由用户取消)"})
		}()

		// 持续读取WebSocket消息
		for {
			// 读取消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway) {
					mlog.Error(mlog.H{
						"msg": "读取WebSocket消息出错",
						"err": err.Error(),
					})
				} else {
					mlog.Info(mlog.H{
						"msg": "WebSocket连接已关闭",
						"err": err.Error(),
					})
				}
				return
			}

			if len(message) == 0 {
				continue
			}

			// 解析JSON响应
			var tailResponse TsTail
			if err := json.Unmarshal(message, &tailResponse); err != nil {
				mlog.Warn(mlog.H{
					"msg":     "解析WebSocket JSON响应失败",
					"err":     err.Error(),
					"message": string(message),
				})
				continue
			}

			// 发送到通道
			select {
			case dataCh <- &tailResponse:
				// 成功发送
			default:
				// 通道已满，记录警告
				mlog.Warn(mlog.H{
					"msg": "数据通道已满，丢弃消息",
				})
			}
		}
	}()

	return dataCh, cancel, nil
}

// LabelValues 查询标签值
func (t *Client) LabelValues(label string, queryParams ...map[string]string) ([]string, error) {
	queryParam := map[string]string{}
	if len(queryParams) > 0 {
		queryParam = queryParams[0]
	}

	resp := &TsLabelValues{}
	_, err := t.resty.R().
		SetQueryParams(queryParam).
		SetResult(resp).
		Get(fmt.Sprintf("/loki/api/v1/label/%s/values", label))
	return resp.Data, err
}

// NewClient 创建客户端
func NewClient(baseUrl string, opts ...clientOpts) *Client {
	c := &Client{
		Conf:  &clientConf{},
		resty: resty.New(),
	}

	{
		// 设置 resty 配置
		c.resty.SetTimeout(30 * time.Second)
		c.resty.SetDisableWarn(true)
		c.resty.SetBaseURL(baseUrl)
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithClientBasicAuth 设置基本认证
func WithClientBasicAuth(username, password string) clientOpts {
	return func(c *Client) {
		c.resty.SetBasicAuth(username, password)
	}
}

// WithClientBearerAuth 设置 Bearer 认证
func WithClientBearerAuth(token string) clientOpts {
	return func(c *Client) {
		c.resty.SetHeader("Authorization", "Bearer "+token)
	}
}

// WithClientHeader 设置请求头
func WithClientHeader(key, value string) clientOpts {
	return func(c *Client) {
		c.resty.SetHeader(key, value)
	}
}

// WithClientDebug 设置 debug
func WithClientDebug(debug bool) clientOpts {
	return func(c *Client) {
		c.resty.SetDebug(debug)
	}
}

// WithClientDisableWarn 设置禁用警告
func WithClientDisableWarn(disableWarn bool) clientOpts {
	return func(c *Client) {
		c.resty.SetDisableWarn(disableWarn)
	}
}

// WithClientTimeout 设置超时时间
func WithClientTimeout(timeout time.Duration) clientOpts {
	return func(c *Client) {
		c.resty.SetTimeout(timeout)
	}
}

// WithClientRetry 设置重试次数
func WithClientRetry(retry int) clientOpts {
	return func(c *Client) {
		c.resty.SetRetryCount(retry)
	}
}
