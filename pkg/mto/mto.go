package mto

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// 通用的 float64 转换函数
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return 0
	case bool:
		if val {
			return 1
		}
		return 0
	case int, int8, int16, int32, int64:
		// 利用反射或直接分支都可以，这里用反射做示例
		return float64(anyToInt64(val))
	case uint, uint8, uint16, uint32, uint64:
		return float64(anyToUint64(val))
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// 将各种整型统一转为 int64
func anyToInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint, uint8, uint16, uint32, uint64:
		return int64(anyToUint64(val))
	default:
		return 0
	}
}

// 将各种无符号整型统一转为 uint64
func anyToUint64(v interface{}) uint64 {
	switch val := v.(type) {
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case uint64:
		return val
	default:
		return 0
	}
}

func Bool(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return val == "1" || val == "true" || val == "True"
	case bool:
		return val
	default:
		// 其他数值类型先转 float64 再判断是否等于 1.0
		return toFloat64(v) == 1.0
	}
}

func String(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		// 优先尝试做数值格式化
		f := toFloat64(v)
		if f != 0 {
			// 判断是否是整数形式
			if math.Floor(f) == f {
				return fmt.Sprintf("%d", int64(f))
			}
			return fmt.Sprintf("%.2f", f)
		}
		// 如果无法作为数值转换，则尝试做 JSON
		data, err := json.Marshal(val)
		if err != nil {
			return ""
		}
		return string(data)
	}
}

func Int(v interface{}) int {
	return int(Int64(v))
}

func Int64(v interface{}) int64 {
	// 这里直接先转 float64，再转 int64
	return int64(toFloat64(v))
}

func Float32(v interface{}) float32 {
	return float32(toFloat64(v))
}

func Float64(v interface{}) float64 {
	return toFloat64(v)
}

func Uint(v interface{}) uint {
	return uint(Uint64(v))
}

func Uint64(v interface{}) uint64 {
	// 这里也先转 float64，再转 uint64
	f := toFloat64(v)
	if f < 0 {
		return 0
	}
	return uint64(f)
}

func Uint32(v interface{}) uint32 {
	return uint32(Uint64(v))
}

func Json(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}
