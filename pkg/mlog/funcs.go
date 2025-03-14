package mlog

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

var g_levelColors map[string]string
var g_keyColors map[string]string

func init() {
	g_levelColors = map[string]string{
		"FATAL": "\033[95m",
		"ERROR": "\033[91m",
		"WARN":  "\033[93m",
		"INFO":  "\033[92m",
		"DEBUG": "\033[94m",
		"TRACE": "\033[90m",
	}

	g_keyColors = map[string]string{
		"time":  "\033[30m",
		"msg":   "\033[34m",
		"error": "\033[31m",
		"warn":  "\033[33m",
		"info":  "\033[34m",
		"data":  "\033[32m",
		"other": "\033[36m",
		"call":  "\033[35m",
	}
}

func colorizeJSONValues(fields H, orderedKeys []string) string {
	resetColor := "\033[0m"

	var otherKeys []string
	for key := range fields {
		if key != "call" && !slices.Contains(orderedKeys, key) {
			otherKeys = append(otherKeys, key)
		}
	}
	sort.Strings(otherKeys)

	allKeys := append(orderedKeys, otherKeys...)

	if _, exists := fields["call"]; exists {
		allKeys = append(allKeys, "call")
	}

	// 预估 Builder 需要的容量 - 每个字段约需要 50-100 字节
	// 这将减少 Builder 内部缓冲区的重新分配次数
	estimatedSize := 20 + len(allKeys)*100
	var builder strings.Builder
	builder.Grow(estimatedSize)

	builder.WriteByte('{')

	first := true
	for _, key := range allKeys {
		value, exists := fields[key]
		if !exists {
			continue
		}

		if !first {
			builder.WriteByte(',')
		}
		first = false

		builder.WriteByte('"')
		builder.WriteString(key)
		builder.WriteString(`":`)

		// 声明valueBytes变量
		var valueBytes []byte

		// 处理error类型值
		if err, ok := value.(error); ok {
			// 直接使用一个JSON安全的字符串表示
			escapedStr := strings.Replace(err.Error(), `"`, `\"`, -1)
			valueBytes = []byte(`"` + escapedStr + `"`)
		} else {
			// 非error类型才尝试JSON序列化
			var err error
			valueBytes, err = json.Marshal(value)
			if err != nil {
				// 序列化失败时提供更安全的回退方案
				escapedStr := strings.Replace(fmt.Sprintf("%v", value), `"`, `\"`, -1)
				valueBytes = []byte(`"` + escapedStr + `"`)
			}
		}

		if key == "level" {
			level := fmt.Sprintf("%v", value)
			if color, ok := g_levelColors[level]; ok {
				builder.WriteString(color)
				builder.Write(valueBytes)
				builder.WriteString(resetColor)
				continue
			}
		}

		if color, ok := g_keyColors[key]; ok {
			builder.WriteString(color)
			builder.Write(valueBytes)
			builder.WriteString(resetColor)
		} else {
			builder.WriteString(g_keyColors["other"])
			builder.Write(valueBytes)
			builder.WriteString(resetColor)
		}
	}
	builder.WriteByte('}')

	return builder.String()
}

// 为全局日志实例提供同样的工具函数
func IsLevelEnabled(level int) bool {
	return this.IsLevelEnabled(level)
}

func levelToString(level int) string {
	if level >= 0 && level <= levelTrace {
		return levelNames[level]
	}
	return levelNames[len(levelNames)-1] // 返回 "UNKNOWN"
}
