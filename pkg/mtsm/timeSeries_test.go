package mtsm

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestTimeSeries_NowTruncate(t *testing.T) {

	// 定义测试用例
	testCases := []struct {
		name   string        // 测试名称
		step   time.Duration // 步进时间
		count  int           // 测试计数
		format string        // 时间格式化字符串
		adjust int64         // 时间戳调整（秒）
	}{
		{name: "步进 10 秒", step: 10 * time.Second, count: 4, format: "2006-01-02 15:04:05.000", adjust: 0},
		{name: "步进 30 秒", step: 30 * time.Second, count: 4, format: "2006-01-02 15:04:05.000", adjust: 0},
		{name: "步进 1 分钟", step: time.Minute, count: 4, format: "2006-01-02 15:04:05", adjust: 0},
		{name: "步进 5 分钟", step: 5 * time.Minute, count: 4, format: "2006-01-02 15:04:05", adjust: 0},
		{name: "步进 1 小时", step: time.Hour, count: 4, format: "2006-01-02 15:04:05", adjust: 0},
		{name: "步进 1 天", step: 24 * time.Hour, count: 4, format: "2006-01-02 15:04:05", adjust: -8 * 60 * 60}, // 步进一天对齐 0 点需要减去 8 小时
	}

	ts := NewTimeSeries(
		map[string]string{"__name__": "test"},
		os.Getenv("ACF_VMETRICS_URL"),
	)
	fmt.Println(" ------ 当前时间: ", time.Now().In(ts.conf.TimeLoc).Format("2006-01-02 15:04:05"))

	// 执行测试用例
	for _, tc := range testCases {
		fmt.Printf(" ------ %s:\n", tc.name)
		now := ts.NowTruncate(tc.step)

		for i := 0; i < tc.count; i++ {
			timeValue := now()
			// 应用调整（如果有的话）
			if tc.adjust != 0 {
				timeValue = timeValue.Add(time.Duration(tc.adjust) * time.Second)
			}
			timestamp := timeValue.Unix()
			fmt.Printf("步进 %d: 时间戳 %d, 时间 %s\n", i, timestamp, timeValue.Format(tc.format))
		}
	}
}
