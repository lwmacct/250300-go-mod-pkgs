package mlog

import (
	"os"
)

var this *ts = New(nil)

// 全局日志公共逻辑函数
func logAt(level int, fields H, callDepth int) *ts {
	if this.config.Level < level {
		return this
	}
	fields["level"] = levelToString(level)
	return this.logWithLevel(fields, callDepth+1)
}

func Fatal(fields H) *ts {
	if this.config.Level < levelFatal {
		return this
	}

	fields["level"] = levelToString(levelFatal)
	this.logWithLevel(fields, 2)
	os.Exit(1)
	return this
}

func Error(fields H) *ts {
	return logAt(levelError, fields, 1)
}

func Warn(fields H) *ts {
	return logAt(levelWarn, fields, 1)
}

func Info(fields H) *ts {
	return logAt(levelInfo, fields, 1)
}

func Debug(fields H) *ts {
	return logAt(levelDebug, fields, 1)
}

func Trace(fields H) *ts {
	return logAt(levelTrace, fields, 1)
}

func ShowLevel() {
	println(this.config.Level)
}

// Close 关闭全局日志实例，确保所有日志被处理
func Close() {
	this.Close()
}

func GetLevel() int {
	return this.GetLevel()
}
