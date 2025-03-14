package mtime

import (
	"regexp"
	"strings"
	"time"
)

var (
	defaultMust = NewMust()
)

type must struct {
	loc *time.Location

	formatDecimalRegex *regexp.Regexp
}

type mustOpts func(*must)

func NewMust(opts ...mustOpts) *must {
	m := &must{
		loc: time.FixedZone("CST", 8*3600),

		// 正则表达式匹配数字和单位
		// 解释:
		// (-?\d+)    : 捕获整数部分（可选负号）
		// (\.\d+)?   : 可选的小数部分
		// ([a-zA-Z]+): 捕获单位
		formatDecimalRegex: regexp.MustCompile(`(-?\d+)(\.\d+)?([a-zA-Z]+)`),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithMustLoc(loc *time.Location) mustOpts {
	return func(t *must) {
		t.loc = loc
	}
}

// TruncateFunc 向过去取整, 按指定的时间间隔步进,返回一个函数, 每次调用步进一次
//
// 如果传入 date 参数, 则以 date 为准, 否则以当前时间为准
func (t *must) TruncateFunc(step time.Duration, date ...time.Time) func() time.Time {
	var now time.Time
	if len(date) > 0 {
		now = date[0].In(t.loc).Truncate(step)
	} else {
		now = time.Now().In(t.loc).Truncate(step)
	}

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

// 获取 start 到 end 之间有多少个 step
func (t *must) Points(start time.Time, end time.Time, step time.Duration) int {
	return int(end.Sub(start) / step)
}

// 过去今天 0 点到现在过去了多少分钟
func (t *must) TodayPastMinutes() int {
	points := t.Points(
		t.TruncateFunc(24*time.Hour)().Add(-8*time.Hour),
		t.TruncateFunc(1*time.Minute)(),
		time.Minute,
	)

	return points
}

// 当前时间, 上海时区
func (t *must) Now() time.Time {
	return time.Now().In(t.loc)
}

// Since 返回自 start 以来经过的时间, 去除小数点, 格式化为字符串。
func (t *must) Since(start time.Time) string {
	value := time.Since(start).String()
	digit := 0
	if digit < 0 {
		// 如果小数位数为负数，视为无效，直接返回原始字符串
		return value
	}

	matches := t.formatDecimalRegex.FindAllStringSubmatchIndex(value, -1)
	if matches == nil {
		// 如果没有匹配项，直接返回原始字符串
		return value
	}

	var result strings.Builder
	prevEnd := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// 将非匹配部分直接写入结果
		if start > prevEnd {
			result.WriteString(value[prevEnd:start])
		}

		// 提取各个捕获组的内容
		intStart, intEnd := match[2], match[3]
		decimalStart, decimalEnd := match[4], match[5]
		unitStart, unitEnd := match[6], match[7]

		intPart := value[intStart:intEnd]
		var decimalPart string
		if decimalStart != -1 && decimalEnd != -1 {
			decimalPart = value[decimalStart+1 : decimalEnd] // 去掉点号
		}
		unit := value[unitStart:unitEnd]

		// 根据 n 处理小数部分
		if digit == 0 {
			// 不保留小数位，直接使用整数部分
			result.WriteString(intPart + unit)
		} else {
			if decimalPart != "" {
				if len(decimalPart) > digit {
					// 保留前 n 位
					decimalPart = decimalPart[:digit]
				} else {
					// 如果小数位数不足 n 位，保留现有的小数部分
					// 可以选择是否补零，这里选择不补零
				}
				// 拼接整数部分、小数点和处理后的小数部分
				result.WriteString(intPart + "." + decimalPart + unit)
			} else {
				// 没有小数部分，直接拼接整数部分和单位
				result.WriteString(intPart + unit)
			}
		}

		prevEnd = end
	}

	// 将最后一部分写入结果
	if prevEnd < len(value) {
		result.WriteString(value[prevEnd:])
	}

	return result.String()
}
