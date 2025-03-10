package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 日志级别
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

// 日志级别名称
var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

// 日志条目结构
type logEntry struct {
	level   int
	message string
}

// Logger 结构体定义
type Logger struct {
	level      int
	logFile    *os.File
	logWriter  io.Writer
	consoleLog *log.Logger
	fileLog    *log.Logger
	mutex      sync.Mutex
	filePath   string
	maxSize    int64 // 日志文件最大大小（字节）
	curSize    int64 // 当前日志文件大小
	
	// 异步日志相关
	logChan    chan logEntry // 日志消息通道
	closeChan  chan struct{} // 关闭信号通道
	wg         sync.WaitGroup // 等待组，确保安全关闭
	batchSize  int           // 批量写入大小
	flushInterval time.Duration // 定时刷新间隔
}

// 全局日志实例
var defaultLogger *Logger
var once sync.Once

// 初始化默认日志实例
func init() {
	once.Do(func() {
		var err error
		defaultLogger, err = NewLogger("logs", "exporter.log", INFO, 10*1024*1024) // 默认10MB
		if err != nil {
			fmt.Printf("初始化日志失败: %v\n", err)
			os.Exit(1)
		}
	})
}

// NewLogger 创建新的日志记录器
func NewLogger(logDir, logFileName string, level int, maxSize int64) (*Logger, error) {
    // 确保日志目录存在
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return nil, fmt.Errorf("创建日志目录失败: %v", err)
    }

    // 构建日志文件路径
    filePath := filepath.Join(logDir, logFileName)

    // 打开日志文件
    file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, fmt.Errorf("打开日志文件失败: %v", err)
    }

    // 获取当前文件大小
    fileInfo, err := file.Stat()
    if err != nil {
        file.Close()
        return nil, fmt.Errorf("获取文件信息失败: %v", err)
    }

	// 创建多输出writer
	multiWriter := io.MultiWriter(os.Stdout, file)

    // 创建日志记录器
    logger := &Logger{
        level:      level,
        logFile:    file,
        logWriter:  multiWriter,
        consoleLog: log.New(os.Stdout, "", 0),
        fileLog:    log.New(file, "", 0),
        filePath:   filePath,
        maxSize:    maxSize,
        curSize:    fileInfo.Size(),
        logChan:    make(chan logEntry, 10000), // 缓冲大小可调整
        closeChan:  make(chan struct{}),
        batchSize:  100,                        // 每批处理100条日志
        flushInterval: 100 * time.Millisecond,  // 100ms刷新一次
    }

    // 启动异步日志处理
    logger.wg.Add(1)
    go logger.processLogs()

    return logger, nil
}

// 异步处理日志
func (l *Logger) processLogs() {
	defer l.wg.Done()

	// 创建定时器，定期刷新日志
	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	// 批量日志缓冲
	buffer := make([]logEntry, 0, l.batchSize)
	
	for {
		select {
		case entry := <-l.logChan:
			// 添加到缓冲
			buffer = append(buffer, entry)
			
			// 如果达到批处理大小，立即写入
			if len(buffer) >= l.batchSize {
				l.writeBatch(buffer)
				buffer = buffer[:0] // 清空缓冲区但保留容量
			}
			
		case <-ticker.C:
			// 定时刷新，即使未达到批处理大小
			if len(buffer) > 0 {
				l.writeBatch(buffer)
				buffer = buffer[:0]
			}
			
			// 检查是否需要轮转日志文件
			l.checkRotate(0)
			
		case <-l.closeChan:
			// 关闭前写入剩余日志
			if len(buffer) > 0 {
				l.writeBatch(buffer)
			}
			return
		}
	}
}

// 批量写入日志
func (l *Logger) writeBatch(entries []logEntry) {
    if len(entries) == 0 {
        return
    }

    l.mutex.Lock()
    defer l.mutex.Unlock()

    var totalBytes int64
    
    // 一次性构建所有日志消息
    for _, entry := range entries {
        msg := formatLogMessage(entry.level, entry.message)
        n, err := fmt.Fprintln(l.logWriter, msg)
        if err != nil {
            // 处理编码错误，直接尝试写入原始消息
            n, err = fmt.Fprintln(l.logWriter, entry.message)
            if err != nil {
                fmt.Printf("写入日志失败: %v\n", err)
                continue
            }
        }
        totalBytes += int64(n)
    }
    
    // 更新文件大小
    l.curSize += totalBytes
}

// 检查并轮转日志文件
func (l *Logger) checkRotate(additionalBytes int64) error {
	// 预估文件大小
	estimatedSize := l.curSize + additionalBytes

	// 如果预估大小未超过最大大小，直接返回
	if estimatedSize < l.maxSize {
		return nil
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 再次检查（可能在获取锁的过程中已经被其他goroutine轮转）
	if l.curSize < l.maxSize {
		return nil
	}

	// 关闭当前日志文件
	l.logFile.Close()

	// 生成新的文件名（添加时间戳）
	timeStr := time.Now().Format("20060102-150405")
	dir := filepath.Dir(l.filePath)
	base := filepath.Base(l.filePath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	newPath := filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, timeStr, ext))

	// 重命名当前日志文件
	if err := os.Rename(l.filePath, newPath); err != nil {
		return fmt.Errorf("重命名日志文件失败: %v", err)
	}

	// 创建新的日志文件
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建新日志文件失败: %v", err)
	}

	// 更新日志记录器
	l.logFile = file
	l.logWriter = io.MultiWriter(os.Stdout, file)
	l.fileLog = log.New(file, "", 0)
	l.curSize = 0

	return nil
}

// 格式化日志消息
func formatLogMessage(level int, message string) string {
	timeStr := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("[%s] [%s] %s", timeStr, levelNames[level], message)
}

// 添加日志到通道
func (l *Logger) addLog(level int, format string, args ...interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	// 格式化消息
	msg := fmt.Sprintf(format, args...)
	
	// 如果是致命错误，直接写入并退出
	if level == FATAL {
		l.mutex.Lock()
		formattedMsg := formatLogMessage(level, msg)
		fmt.Fprintln(l.logWriter, formattedMsg)
		l.mutex.Unlock()
		os.Exit(1)
	}

	// 发送到日志通道
	select {
	case l.logChan <- logEntry{level: level, message: msg}:
		// 成功发送
	default:
		// 通道已满，直接写入控制台
		fmt.Printf("日志通道已满，丢弃日志: %s\n", msg)
	}
}

// Debug 输出调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.addLog(DEBUG, format, args...)
}

// Info 输出信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.addLog(INFO, format, args...)
}

// Warn 输出警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.addLog(WARN, format, args...)
}

// Error 输出错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.addLog(ERROR, format, args...)
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.addLog(FATAL, format, args...)
}

// 设置日志级别
func (l *Logger) SetLevel(level int) {
	l.level = level
}

// 关闭日志
func (l *Logger) Close() {
	// 发送关闭信号
	close(l.closeChan)
	
	// 等待日志处理完成
	l.wg.Wait()
	
	// 关闭文件
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
}

// 以下是全局函数，使用默认日志记录器

// Debug 输出调试级别日志
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info 输出信息级别日志
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn 输出警告级别日志
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error 输出错误级别日志
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// SetLevel 设置默认日志记录器的日志级别
func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

// Close 关闭默认日志记录器
func Close() {
	defaultLogger.Close()
}