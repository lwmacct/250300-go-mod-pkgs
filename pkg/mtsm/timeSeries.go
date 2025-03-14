package mtsm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

type timeSeries struct {
	conf    *timeSeriesConf
	labels  []prompb.Label
	samples []prompb.Sample
}

type timeSeriesConf struct {
	TimeLoc *time.Location
	Raw     *remote.ClientConfig
	Err     error
}

type timeSeriesOpts func(*timeSeriesConf)

func NewTimeSeries(labels map[string]string, UrlScheme string, opts ...timeSeriesOpts) *timeSeries {
	t := &timeSeries{
		labels:  make([]prompb.Label, 0),
		samples: make([]prompb.Sample, 0),
		conf: &timeSeriesConf{
			TimeLoc: time.FixedZone("CST", 8*3600),
			Raw: &remote.ClientConfig{
				Timeout: model.Duration(30 * time.Second),
			},
			Err: nil,
		},
	}

	opts = append(opts, WithTimeSeries(UrlScheme))
	for _, opt := range opts {
		opt(t.conf)
	}

	// 设置标签
	for k, v := range labels {
		t.labels = append(t.labels, prompb.Label{Name: k, Value: v})
	}

	return t
}

func WithTimeSeries(UrlScheme string, rewritePath ...string) timeSeriesOpts {
	return func(c *timeSeriesConf) {
		urlParse, err := url.Parse(UrlScheme)
		if err != nil {
			c.Err = fmt.Errorf("WithUrlParse: %w", err)
			return
		}
		if len(rewritePath) > 0 {
			urlParse.Path = rewritePath[0]
		}
		if urlParse.Path == "" {
			urlParse.Path = "/api/v1/write"
		}
		c.Raw.URL = &config_util.URL{
			URL: urlParse,
		}
	}
}

func WithTimeSeriesTimeout(timeout time.Duration) timeSeriesOpts {
	return func(c *timeSeriesConf) {
		c.Raw.Timeout = model.Duration(timeout)
	}
}

// 单个值, 单个时间戳
func (t *timeSeries) AddValue(value float64, timestamps ...int64) {
	timestamp := int64(0)
	// timestamp 取整为秒
	if len(timestamps) == 0 {
		timestamp = t.NowUnix(0)
	} else {
		timestamp = timestamps[0]
	}
	t.samples = append(t.samples, prompb.Sample{Value: value, Timestamp: timestamp * 1000})
}

// 多个值, 多个时间戳
func (t *timeSeries) AddMulti(values []float64, timestamps []int64) error {
	if len(values) != len(timestamps) {
		return errors.New("values and timestamps length mismatch")
	}
	for i, value := range values {
		t.AddValue(value, timestamps[i])
	}
	return nil
}

// 时间戳为key, 值为value
func (t *timeSeries) AddMap(values map[int64]float64) {
	for timestamp, value := range values {
		t.AddValue(value, timestamp)
	}
}

// 设置标签
func (t *timeSeries) SetLabel(key string, value string) {
	found := false
	for i, label := range t.labels {
		if label.Name == key {
			t.labels[i].Value = value
			found = true
			break
		}
	}
	if !found {
		t.labels = append(t.labels, prompb.Label{Name: key, Value: value})
	}
}

// tojson
func (t *timeSeries) ToJson() string {
	json, err := json.Marshal(t.ToObject())
	if err != nil {
		return ""
	}
	return string(json)
}

// func toObject
func (t *timeSeries) ToObject() map[string]interface{} {
	return map[string]interface{}{
		"labels":  t.labels,
		"samples": t.samples,
	}
}

func (t *timeSeries) NowUnix(offset ...int64) int64 {
	opts := []int64{0}
	if len(offset) > 0 {
		opts = offset
	}
	return time.Now().In(t.conf.TimeLoc).Unix() + opts[0]
}

// NowTruncate 向过去取整, 按指定的时间间隔步进,返回一个函数, 每次调用步进一次
func (t *timeSeries) NowTruncate(step time.Duration) func() time.Time {
	now := time.Now().In(t.conf.TimeLoc).Truncate(step)
	firstCall := true
	return func() time.Time {
		if firstCall {
			firstCall = false
		} else {
			now = now.Add(-step)
		}
		return now
	}
}

// 将压缩后的数据发送到远程服务器
func (t *timeSeries) Push(ctx context.Context, retryAttempt int) ([]byte, error) {

	req := &prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels:  t.labels,
				Samples: t.samples,
			},
		},
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 对序列化后的数据进行压缩（Prometheus使用snappy压缩）
	compressed := snappy.Encode(nil, data)

	pusher, err := remote.NewWriteClient("metrics_sender", t.conf.Raw)
	if err != nil {
		return nil, err
	}

	// 将压缩后的数据发送到远程服务器
	_, err = pusher.Store(ctx, compressed, retryAttempt)
	if err != nil {
		return nil, err
	}
	return compressed, nil
}
