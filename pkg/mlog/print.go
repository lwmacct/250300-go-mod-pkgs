package mlog

import "os"

// 创建一个通用的日志级别处理函数
func (t *ts) logAt(level int, fields H, callDepth int) *ts {
	if t.config.Level < level {
		return t
	}
	fields["level"] = levelToString(level)
	return t.logWithLevel(fields, callDepth+1)
}

func (t *ts) Fatal(fields H) *ts {
	if t.config.Level < levelFatal {
		return t
	}
	fields["level"] = levelToString(levelFatal)
	t.logWithLevel(fields, 2)
	os.Exit(1)
	return t
}

func (t *ts) Error(fields H) *ts {
	return t.logAt(levelError, fields, 1)
}

func (t *ts) Warn(fields H) *ts {
	return t.logAt(levelWarn, fields, 1)
}

func (t *ts) Info(fields H) *ts {
	return t.logAt(levelInfo, fields, 1)
}

func (t *ts) Debug(fields H) *ts {
	return t.logAt(levelDebug, fields, 1)
}

func (t *ts) Trace(fields H) *ts {
	return t.logAt(levelTrace, fields, 1)
}
