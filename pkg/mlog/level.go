package mlog

// Log levels

const (
	levelFatal = iota // 0
	levelError        // 1
	levelWarn         // 2
	levelInfo         // 3
	levelDebug        // 4
	levelTrace        // 5
)

// 预定义日志级别字符串常量
var levelNames = [...]string{
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
	"UNKNOWN", // 用于默认情况
}
