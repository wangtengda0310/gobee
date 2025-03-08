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
	}

	return logger, nil
}

// 检查并轮转日志文件
func (l *Logger) checkRotate(bytesWritten int64) error {
	l.curSize += bytesWritten

	// 如果当前文件大小超过最大大小，进行轮转
	if l.curSize >= l.maxSize {
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
	}

	return nil
}

// 格式化日志消息
func formatLogMessage(level int, format string, args ...interface{}) string {
	timeStr := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] %s", timeStr, levelNames[level], msg)
}

// 写入日志
func (l *Logger) writeLog(level int, format string, args ...interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	// 格式化日志消息
	msg := formatLogMessage(level, format, args...)

	// 加锁确保并发安全
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 写入日志
	bytesWritten, err := fmt.Fprintln(l.logWriter, msg)
	if err != nil {
		fmt.Printf("写入日志失败: %v\n", err)
		return
	}

	// 检查是否需要轮转日志文件
	if err := l.checkRotate(int64(bytesWritten)); err != nil {
		fmt.Printf("轮转日志文件失败: %v\n", err)
	}

	// 如果是致命错误，程序退出
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 输出调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.writeLog(DEBUG, format, args...)
}

// Info 输出信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.writeLog(INFO, format, args...)
}

// Warn 输出警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.writeLog(WARN, format, args...)
}

// Error 输出错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.writeLog(ERROR, format, args...)
}

// Fatal 输出致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.writeLog(FATAL, format, args...)
}

// 设置日志级别
func (l *Logger) SetLevel(level int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// 关闭日志
func (l *Logger) Close() {
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