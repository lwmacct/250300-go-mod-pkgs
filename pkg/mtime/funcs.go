package mtime

import "time"

// TruncateFunc 向过去取整, 按指定的时间间隔步进,返回一个函数, 每次调用步进一次
//
// 如果传入 date 参数, 则以 date 为准, 否则以当前时间为准
func TruncateFunc(step time.Duration, date ...time.Time) func() time.Time {
	return defaultMust.TruncateFunc(step, date...)
}

// 获取 start 到 end 之间有多少个 step
func Points(start time.Time, end time.Time, step time.Duration) int {
	return defaultMust.Points(start, end, step)
}

// 过去今天 0 点到现在过去了多少分钟
func TodayPastMinutes() int {
	return defaultMust.TodayPastMinutes()
}

// 当前时间, 上海时区
func Now() time.Time {
	return defaultMust.Now()
}

// 上海时区
func LocCST() *time.Location {
	return time.FixedZone("CST", 8*3600)
}

// Since 返回自 start 以来经过的时间, 去除小数点, 格式化为字符串。
func Since(start time.Time) string {
	return defaultMust.Since(start)
}
