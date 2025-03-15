package mlog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var this *ts = New()

const (
	g_levelFatal = iota // 0
	g_levelError        // 1
	g_levelWarn         // 2
	g_levelInfo         // 3
	g_levelDebug        // 4
	g_levelTrace        // 5
)

var g_levelColors = map[string]string{
	"FATAL": "\033[95m",
	"ERROR": "\033[91m",
	"WARN":  "\033[93m",
	"INFO":  "\033[92m",
	"DEBUG": "\033[94m",
	"TRACE": "\033[90m",
}
var g_keyColors = map[string]string{
	"time":  "\033[30m",
	"msg":   "\033[34m",
	"error": "\033[31m",
	"warn":  "\033[33m",
	"info":  "\033[34m",
	"data":  "\033[32m",
	"other": "\033[36m",
	"call":  "\033[35m",
}

// 预定义日志级别字符串常量
var g_levelNames = [...]string{
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
	"UNKNOWN", // 用于默认情况
}

type H map[string]interface{}

type Opts struct {
	File   string `group:"mlog" note:"日志文件名" default:""`
	Level  int    `group:"mlog" note:"日志级别" default:"3"`
	Stdout bool   `group:"mlog" note:"是否输出到标准输出" default:"true"`
	Color  bool   `group:"mlog" note:"是否启用终端颜色显示" default:"true"`

	// lumberjack.Logger fields
	RotateMaxSize    int  `group:"mlog" note:"日志文件最大尺寸" default:"100"`
	RotateMaxAge     int  `group:"mlog" note:"日志文件最大保存天数" default:"30"`
	RotateMaxBackups int  `group:"mlog" note:"日志文件最大备份数" default:"3"`
	RotateLocalTime  bool `group:"mlog" note:"是否使用本地时间" default:"false"`
	RotateCompress   bool `group:"mlog" note:"是否压缩" default:"false"`

	// 异步配置
	AsyncEnabled   bool `group:"mlog" note:"是否启用异步日志" default:"true"`
	AsyncQueueSize int  `group:"mlog" note:"异步日志队列大小" default:"1000"`
	AsyncBatchSize int  `group:"mlog" note:"异步日志批处理大小" default:"32"`
	AsyncWorkers   int  `group:"mlog" note:"异步日志处理工作线程数" default:"1"`

	CallerClip string `group:"mlog" note:"裁剪调用路径" default:""`
}

type ts struct {
	opts *Opts

	H      map[string]any
	logger *log.Logger

	orderedKeys []string

	// 时间缓存相关
	cachedTimeStr  string
	lastTimeUpdate int64
	timeMutex      sync.RWMutex

	// 异步处理相关
	logChan  chan logEntry
	wg       sync.WaitGroup
	shutdown chan struct{}
}

// 日志条目结构，用于异步传递
type logEntry struct {
	fields      H
	orderedKeys []string
}

type tsOpts func(*ts)

func New(opts ...tsOpts) *ts {
	t := &ts{
		opts: &Opts{
			Stdout: true,
			Level:  3,
			File:   "",
			Color:  true,

			AsyncEnabled:   true,
			AsyncQueueSize: 1000,

			RotateMaxSize:    100,
			RotateMaxAge:     90,
			RotateMaxBackups: 3,
			RotateLocalTime:  true,
			RotateCompress:   true,
		},
		shutdown:    make(chan struct{}),
		orderedKeys: []string{"time", "level", "msg", "info", "error", "warn", "data", "flags"},
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(t)
	}

	// 初始化异步处理
	if t.opts.AsyncEnabled {
		t.logChan = make(chan logEntry, t.opts.AsyncQueueSize)

		// 确定工作线程数
		workers := t.opts.AsyncWorkers
		if workers <= 0 {
			workers = 1 // 至少启动一个工作线程
		}

		// 启动多个工作线程处理日志
		for i := 0; i < workers; i++ {
			t.wg.Add(1)
			go t.processLogs()
		}
	}

	if t.opts.File != "" {
		logger := &lumberjack.Logger{
			Filename:   t.opts.File,
			MaxSize:    t.opts.RotateMaxSize,
			MaxBackups: t.opts.RotateMaxBackups,
			MaxAge:     t.opts.RotateMaxAge,
			Compress:   t.opts.RotateCompress,
			LocalTime:  t.opts.RotateLocalTime,
		}
		t.logger = log.New(logger, "", 0)
	}
	return t
}

// 日志文件名, 指定文件名
func WithFile(file string) tsOpts {
	return func(t *ts) {
		t.opts.File = file
	}
}

// 日志文件名, 默认使用程序名
func WithFileDefault() tsOpts {
	return func(t *ts) {
		t.opts.File = fmt.Sprintf("/var/log/%s.log", filepath.Base(os.Args[0]))
	}
}

// WithLevel 设置日志级别
func WithLevel(level int) tsOpts {
	return func(t *ts) {
		t.opts.Level = level
	}
}

// 控制是否启用终端颜色显示
func WithColor(enabled bool) tsOpts {
	return func(t *ts) {
		t.opts.Color = enabled
	}
}

// WithCallerClip 设置调用路径裁剪
func WithCallerClip(clip string) tsOpts {
	return func(t *ts) {
		t.opts.CallerClip = clip
	}
}

func GetOpts() *Opts {
	return this.opts
}

func SetNew(opts ...tsOpts) *ts {
	this.Close()
	this = New(opts...)
	return this
}

// Close 关闭全局日志实例，确保所有日志被处理
func Close() {
	this.Close()
}

func Fatal(fields H) *ts {
	if this.opts.Level < g_levelFatal {
		return this
	}

	fields["level"] = levelToString(g_levelFatal)
	this.print(fields, 2)
	os.Exit(1)
	return this
}

func Error(fields H) *ts {
	return logAt(g_levelError, fields, 1)
}

func Warn(fields H) *ts {
	return logAt(g_levelWarn, fields, 1)
}

func Info(fields H) *ts {
	return logAt(g_levelInfo, fields, 1)
}

func Debug(fields H) *ts {
	return logAt(g_levelDebug, fields, 1)
}

func Trace(fields H) *ts {
	return logAt(g_levelTrace, fields, 1)
}

// 创建一个通用的日志级别处理函数
func (t *ts) logAt(level int, fields H, callDepth int) *ts {
	if t.opts.Level < level {
		return t
	}
	fields["level"] = levelToString(level)
	return t.print(fields, callDepth+1)
}

// 记录 fatal 级别日志
func (t *ts) Fatal(fields H) *ts {
	if t.opts.Level < g_levelFatal {
		return t
	}
	fields["level"] = levelToString(g_levelFatal)
	t.print(fields, 2)
	os.Exit(1)
	return t
}

// 记录 error 级别日志
func (t *ts) Error(fields H) *ts {
	return t.logAt(g_levelError, fields, 1)
}

// 记录 warn 级别日志
func (t *ts) Warn(fields H) *ts {
	return t.logAt(g_levelWarn, fields, 1)
}

// 记录 info 级别日志
func (t *ts) Info(fields H) *ts {
	return t.logAt(g_levelInfo, fields, 1)
}

// 记录 debug 级别日志
func (t *ts) Debug(fields H) *ts {
	return t.logAt(g_levelDebug, fields, 1)
}

// 记录 trace 级别日志
func (t *ts) Trace(fields H) *ts {
	return t.logAt(g_levelTrace, fields, 1)
}

// 关闭日志系统，确保所有日志都被处理
func (t *ts) Close() {
	if t.opts.AsyncEnabled {
		close(t.shutdown)
		t.wg.Wait()
	}
}

func (t *ts) print(fields H, callDepth int) *ts {
	t.setCaller(fields, callDepth+2)

	// 对于同步日志处理
	if !t.opts.AsyncEnabled {
		// 生成有颜色的输出（用于终端）
		coloredOutput := colorizeJSONValues(fields, t.orderedKeys, t.opts)
		// 生成无颜色的输出（用于文件）
		plainOutput := colorizeJSONValues(fields, t.orderedKeys, &Opts{Color: false})

		// 同步处理日志
		if t.opts.File != "" {
			// 文件输出使用无颜色版本
			t.logger.Println(plainOutput)
		}
		if t.opts.Stdout {
			// 终端输出使用有颜色版本（如果启用了颜色）
			fmt.Println(coloredOutput)
		}
	} else {
		// 对于异步日志处理，只传递字段和键序列
		entry := logEntry{
			fields:      copyFields(fields),
			orderedKeys: t.orderedKeys,
		}

		// 异步发送日志
		select {
		case t.logChan <- entry:
			// 成功发送到通道
		default:
			// 通道已满，输出警告并丢弃
			if t.opts.Stdout {
				fmt.Println("警告：日志缓冲区已满，日志被丢弃")
			}
		}
	}

	return t
}

// 处理异步日志的goroutine
func (t *ts) processLogs() {
	defer t.wg.Done()

	// 批量处理缓冲区
	batchSize := t.opts.AsyncBatchSize
	if batchSize <= 0 {
		batchSize = 32 // 默认批处理大小
	}

	// 批量日志缓冲区
	batch := make([]logEntry, 0, batchSize)

	// 定期刷新计时器，确保低流量时日志也能及时写入
	// 提高刷新间隔以减少CPU消耗
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// 预分配一个足够大的临时缓冲区用于处理logChan中的消息
	tmpBatch := make([]logEntry, 0, batchSize*2)

	for {
		select {
		case logEntry := <-t.logChan:
			// 添加到批处理缓冲区
			batch = append(batch, logEntry)

			// 如果达到批处理阈值，立即处理
			if len(batch) >= batchSize {
				t.processBatch(batch)
				batch = batch[:0] // 重置批处理缓冲区，保留容量
			}

			// 尝试一次性获取更多消息（非阻塞方式）
			drainCount := 0
			for len(batch) < batchSize && drainCount < batchSize*2 {
				select {
				case entry := <-t.logChan:
					batch = append(batch, entry)
				default:
					// 没有更多消息
					drainCount = batchSize * 2 // 退出循环
				}
				drainCount++
			}

		case <-ticker.C:
			// 定期刷新，即使没有填满批处理缓冲区
			if len(batch) > 0 {
				t.processBatch(batch)
				batch = batch[:0]
			}

		case <-t.shutdown:
			// 关闭前处理剩余日志
			// 先处理批处理缓冲区中的日志
			if len(batch) > 0 {
				t.processBatch(batch)
			}

			// 尝试处理通道中剩余的日志 - 更高效的方式
			draining := true
			for draining {
				// 一次性获取尽可能多的消息
				tmpBatch = tmpBatch[:0] // 重置但保留容量

				// 持续非阻塞地消费通道
				drainLoop := true
				for drainLoop && len(tmpBatch) < cap(tmpBatch) {
					select {
					case entry, ok := <-t.logChan:
						if !ok {
							drainLoop = false
							draining = false
							break
						}
						tmpBatch = append(tmpBatch, entry)
					default:
						// 通道暂时为空
						drainLoop = false
					}
				}

				// 处理收集到的消息
				if len(tmpBatch) > 0 {
					t.processBatch(tmpBatch)
				} else {
					draining = false
				}
			}
			return
		}
	}
}

// 批量处理日志消息
func (t *ts) processBatch(entries []logEntry) {
	for _, entry := range entries {
		// 生成有颜色的输出（用于终端）
		coloredOutput := colorizeJSONValues(entry.fields, entry.orderedKeys, t.opts)
		// 生成无颜色的输出（用于文件）
		plainOutput := colorizeJSONValues(entry.fields, entry.orderedKeys, &Opts{Color: false})

		if t.opts.File != "" {
			// 文件输出使用无颜色版本
			t.logger.Println(plainOutput)
		}
		if t.opts.Stdout {
			// 终端输出使用有颜色版本
			fmt.Println(coloredOutput)
		}
	}
}

// IsLevelEnabled 检查指定日志级别是否启用
// 这允许调用者在创建日志消息前检查级别，避免不必要的对象创建
func (t *ts) IsLevelEnabled(level int) bool {
	return t.opts.Level >= level
}

// GetLevel 返回当前设置的日志级别
func (t *ts) GetLevel() int {
	return t.opts.Level
}

// GetColor 返回当前颜色设置状态
func (t *ts) GetColor() bool {
	return t.opts.Color
}

// 获取格式化的时间，使用缓存减少格式化开销
func (t *ts) getFormattedTime() string {
	now := time.Now().Unix()

	// 尝试使用缓存的时间
	t.timeMutex.RLock()
	if now-t.lastTimeUpdate < 1 { // 缓存1秒内的时间
		timeStr := t.cachedTimeStr
		t.timeMutex.RUnlock()
		return timeStr
	}
	t.timeMutex.RUnlock()

	// 需要更新时间
	t.timeMutex.Lock()
	defer t.timeMutex.Unlock()

	// 双重检查，避免多线程重复更新
	if now-t.lastTimeUpdate < 1 {
		return t.cachedTimeStr
	}

	currentTime := time.Now()
	t.cachedTimeStr = currentTime.Format("2006-01-02 15:04:05")
	t.lastTimeUpdate = now
	return t.cachedTimeStr
}

func (t *ts) setCaller(fields H, callDepth int) {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "unknown"
		line = 0
	}

	fields["call"] = t.pathClipping(fmt.Sprintf("%s:%d", file, line))

	// 使用优化的时间格式化方法
	fields["time"] = t.getFormattedTime()
}

func (t *ts) pathClipping(path string) string {
	if t.opts.CallerClip == "" {
		if len(path) > 0 && path[0] == '/' {
			// 优化: 避免使用 strings.Split 产生的临时切片
			// 从后向前找到第三个 '/'
			count := 0
			pos := len(path) - 1

			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' {
					count++
					if count == 3 {
						pos = i
						break
					}
				}
			}

			if count >= 3 {
				// 找到了至少3个斜杠
				return path[pos:]
			}
			return path
		}
		return path
	}
	return strings.Replace(path, t.opts.CallerClip, "", -1)
}

func colorizeJSONValues(fields H, orderedKeys []string, config *Opts) string {
	resetColor := "\033[0m"
	isColorEnabled := config.Color

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

		if isColorEnabled {
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
		} else {
			// 无颜色模式，直接写入值
			builder.Write(valueBytes)
		}
	}
	builder.WriteByte('}')

	return builder.String()
}

func levelToString(level int) string {
	if level >= 0 && level <= g_levelTrace {
		return g_levelNames[level]
	}
	return g_levelNames[len(g_levelNames)-1] // 返回 "UNKNOWN"
}

// 全局日志公共逻辑函数
func logAt(level int, fields H, callDepth int) *ts {
	if this.opts.Level < level {
		return this
	}

	fields["level"] = levelToString(level)
	return this.print(fields, callDepth+1)
}

// 复制字段映射，避免并发问题
func copyFields(fields H) H {
	result := make(H, len(fields))
	for k, v := range fields {
		result[k] = v
	}
	return result
}
