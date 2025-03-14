package mlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type H map[string]interface{}

type ts struct {
	config *Config

	H      map[string]any
	logger *log.Logger

	orderedKeys []string

	// 时间缓存相关
	cachedTimeStr  string
	lastTimeUpdate int64
	timeMutex      sync.RWMutex

	// 异步处理相关
	logChan  chan string
	wg       sync.WaitGroup
	shutdown chan struct{}
}

type Config struct {
	File   string `group:"mlog" note:"日志文件名" default:""`
	Level  int    `group:"mlog" note:"日志级别" default:"3"`
	Stdout bool   `group:"mlog" note:"是否输出到标准输出" default:"true"`

	// lumberjack.Logger fields
	RotateMaxSize    int  `group:"mlog" note:"日志文件最大尺寸" default:"100"`
	RotateMaxAge     int  `group:"mlog" note:"日志文件最大保存天数" default:"30"`
	RotateMaxBackups int  `group:"mlog" note:"日志文件最大备份数" default:"3"`
	RotateLocalTime  bool `group:"mlog" note:"是否使用本地时间" default:"false"`
	RotateCompress   bool `group:"mlog" note:"是否压缩" default:"false"`

	// 异步配置
	AsyncEnabled   bool `group:"mlog" note:"是否启用异步日志" default:"true"`
	AsyncQueueSize int  `group:"mlog" note:"异步日志队列大小" default:"1000"`
	// 新增配置: 批处理大小
	AsyncBatchSize int `group:"mlog" note:"异步日志批处理大小" default:"32"`
	// 新增配置: 工作线程数
	AsyncWorkers int `group:"mlog" note:"异步日志处理工作线程数" default:"1"`

	CallerClip string `group:"mlog" note:"裁剪调用路径" default:""`
}

func New(config *Config) *ts {
	if config == nil {
		config = NewConfig()
		config.File = ""
	}
	t := &ts{
		config:      config,
		shutdown:    make(chan struct{}),
		orderedKeys: []string{"time", "level", "msg", "info", "error", "warn", "data", "flags"},
	}

	// 初始化异步处理
	if config.AsyncEnabled {
		t.logChan = make(chan string, config.AsyncQueueSize)

		// 确定工作线程数
		workers := config.AsyncWorkers
		if workers <= 0 {
			workers = 1 // 至少启动一个工作线程
		}

		// 启动多个工作线程处理日志
		for i := 0; i < workers; i++ {
			t.wg.Add(1)
			go t.processLogs()
		}
	}

	if t.config.File != "" {
		logger := &lumberjack.Logger{
			Filename:   t.config.File,
			MaxSize:    t.config.RotateMaxSize,
			MaxBackups: t.config.RotateMaxBackups,
			MaxAge:     t.config.RotateMaxAge,
			Compress:   t.config.RotateCompress,
			LocalTime:  t.config.RotateLocalTime,
		}
		t.logger = log.New(logger, "", 0)
	}
	return t
}

func NewConfig() *Config {
	execName := filepath.Base(os.Args[0])
	fileName := fmt.Sprintf("/var/log/%s.log", execName)

	t := &Config{
		Stdout: true,
		Level:  3,
		File:   fileName,

		RotateMaxSize:    100,
		RotateMaxAge:     90,
		RotateMaxBackups: 3,
		RotateLocalTime:  true,
		RotateCompress:   true,
		AsyncEnabled:     true,
		AsyncQueueSize:   1000,
	}

	return t
}

func (e *ts) logWithLevel(fields H, callDepth int) *ts {
	e.setCaller(fields, callDepth+2)

	coloredOutput := colorizeJSONValues(fields, e.orderedKeys)

	if e.config.AsyncEnabled {
		// 异步发送日志
		select {
		case e.logChan <- coloredOutput:
			// 成功发送到通道
		default:
			// 通道已满，输出警告并丢弃
			if e.config.Stdout {
				fmt.Println("警告：日志缓冲区已满，日志被丢弃")
			}
		}
	} else {
		// 同步处理日志
		if e.config.File != "" {
			e.logger.Println(coloredOutput)
		}
		if e.config.Stdout {
			fmt.Println(coloredOutput)
		}
	}

	return e
}

// 处理异步日志的goroutine
func (e *ts) processLogs() {
	defer e.wg.Done()

	// 批量处理缓冲区
	batchSize := e.config.AsyncBatchSize
	if batchSize <= 0 {
		batchSize = 32 // 默认批处理大小
	}

	// 批量日志缓冲区
	batch := make([]string, 0, batchSize)

	// 定期刷新计时器，确保低流量时日志也能及时写入
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case logMsg := <-e.logChan:
			// 添加到批处理缓冲区
			batch = append(batch, logMsg)

			// 如果达到批处理阈值，立即处理
			if len(batch) >= batchSize {
				e.processBatch(batch)
				batch = batch[:0] // 重置批处理缓冲区，保留容量
			}

		case <-ticker.C:
			// 定期刷新，即使没有填满批处理缓冲区
			if len(batch) > 0 {
				e.processBatch(batch)
				batch = batch[:0]
			}

		case <-e.shutdown:
			// 关闭前处理剩余日志

			// 先处理批处理缓冲区中的日志
			if len(batch) > 0 {
				e.processBatch(batch)
				batch = batch[:0]
			}

			// 尝试处理通道中剩余的日志
			for {
				select {
				case logMsg := <-e.logChan:
					batch = append(batch, logMsg)
					if len(batch) >= batchSize {
						e.processBatch(batch)
						batch = batch[:0]
					}
				default:
					// 处理最后的批次
					if len(batch) > 0 {
						e.processBatch(batch)
					}
					return // 通道已空，退出
				}
			}
		}
	}
}

// 批量处理日志消息
func (e *ts) processBatch(messages []string) {
	if e.config.File != "" {
		for _, msg := range messages {
			e.logger.Println(msg)
		}
	}
	if e.config.Stdout {
		for _, msg := range messages {
			fmt.Println(msg)
		}
	}
}

// 关闭日志系统，确保所有日志都被处理
func (e *ts) Close() {
	if e.config.AsyncEnabled {
		close(e.shutdown)
		e.wg.Wait()
	}
}

// IsLevelEnabled 检查指定日志级别是否启用
// 这允许调用者在创建日志消息前检查级别，避免不必要的对象创建
func (e *ts) IsLevelEnabled(level int) bool {
	return e.config.Level >= level
}

// GetLevel 返回当前设置的日志级别
func (e *ts) GetLevel() int {
	return e.config.Level
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
	if t.config.CallerClip == "" {
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
	return strings.Replace(path, t.config.CallerClip, "", -1)
}
