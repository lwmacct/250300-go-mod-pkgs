package mtsm

import "strconv"

// convertToFloat64 是一个辅助函数，用于将 interface{} 转换为 float64。
// 如果转换失败，则返回 0。
func convertToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		parsedVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return parsedVal
	default:
		return 0
	}
}
