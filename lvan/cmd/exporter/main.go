package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/wangtengda/gobee/lvan/api"
	"github.com/wangtengda/gobee/lvan/internal"
	"github.com/wangtengda/gobee/lvan/pkg"
	"github.com/wangtengda/gobee/lvan/pkg/logger"

	"github.com/spf13/pflag"
)

// 嵌入CLI文档
//
//go:embed cli-doc.txt
var cliDoc string

// 版本信息
const (
	Version = "0.0.0"
)

// 全局工作目录变量
var (
	logsDir string // 日志目录
)

// 新增清理方法
func cleanGeneratedFiles(tasksDir string) {
	// 删除任务目录
	if err := os.RemoveAll(tasksDir); err != nil {
		logger.Error("清理失败: %v", err)
	} else {
		logger.Info("已清理所有任务数据")
	}

}

// 新增环境变量辅助函数
func getEnvString(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.EqualFold(value, "true") || value == "1"
	}
	return defaultVal
}
func main() {
	// 设置程序说明
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Exporter 服务程序 v%s\n\n", Version)
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  exporter [选项]\n\n")
		fmt.Fprintf(os.Stderr, "选项:\n")
		pflag.PrintDefaults()
	}

	// 解析命令行参数
	port := pflag.IntP("port", "p", getEnvInt("EXPORTER_PORT", 80), "指定服务监听的TCP端口默认 80 支持环境变量 EXPORTER_PORT")
	showVersion := pflag.BoolP("version", "v", false, "显示版本号")
	showHelp := pflag.BoolP("help", "h", false, "本说明文档")
	showMoreHelp := pflag.Bool("morehelp", false, "展示更详细的文档")
	logLevel := pflag.String("log-level", "info", "Log level (debug, info, warn, error, fatal)")
	cleanFlag := pflag.Bool("clean", false, "清除任务的工作目录")
	workDirFlag := pflag.StringP("workdir", "w", getEnvString("EXPORTER_WORKDIR", ""), "指定工作目录，默认为程序所在目录，支持环境变量 EXPORTER_WORKDIR")

	// 为参数添加长格式说明
	pflag.Lookup("port").Usage = "指定服务监听的TCP端口默认 80 支持环境变量 EXPORTER_PORT\n" +
		"  示例:\n" +
		"    EXPORTER_PORT=8080	通过环境变量指定端口\n" +
		"    --port 8080     		监听8080端口\n" +
		"    -p 8080        		使用短格式指定端口"

	pflag.Lookup("log-level").Usage = "设置日志输出级别\n" +
		"  可选值:\n" +
		"    debug   调试信息\n" +
		"    info    一般信息\n" +
		"    warn    警告信息\n" +
		"    error   错误信息\n" +
		"    fatal   致命错误"

	pflag.Lookup("workdir").Usage = "指定工作目录，默认为程序所在目录\n" +
		"  示例:\n" +
		"    EXPORTER_WORKDIR=/path/to/dir  通过环境变量指定工作目录\n" +
		"    --workdir /path/to/dir         使用长格式指定工作目录\n" +
		"    -w /path/to/dir                使用短格式指定工作目录"
	pflag.Parse()

	// 初始化工作目录
	if *workDirFlag != "" {
		// 使用命令行参数指定的工作目录
		internal.WorkDir = *workDirFlag
	} else {
		// 默认使用可执行文件所在目录
		execPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("无法获取程序路径: %v\n", err)
			os.Exit(1)
		}
		internal.WorkDir = execPath
	}

	// 确保工作目录存在
	if err := os.MkdirAll(internal.WorkDir, 0755); err != nil {
		fmt.Printf("无法创建工作目录: %v\n", err)
		os.Exit(1)
	}

	// 设置相关目录
	pkg.TasksDir = filepath.Join(internal.WorkDir, "tasks")
	logsDir = filepath.Join(internal.WorkDir, "logs")

	// 确保相关目录存在
	if err := os.MkdirAll(pkg.TasksDir, 0755); err != nil {
		fmt.Printf("无法创建任务目录: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("无法创建日志目录: %v\n", err)
		os.Exit(1)
	}

	// 先初始化日志系统，使用新的日志目录
	loggerInstance, err := logger.NewLogger(logsDir, "exporter.log", logger.INFO, 10*1024*1024, os.Stdout)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	logger.SetDefaultLogger(loggerInstance)

	pkg.CommandDir = filepath.Join(internal.WorkDir, "cmd")

	if *cleanFlag {
		cleanGeneratedFiles(pkg.TasksDir)
		return // 清理后直接退出
	}
	go internal.ScheduleCleaner(pkg.TasksDir, time.Hour*24)

	// 设置日志级别
	switch strings.ToLower(*logLevel) {
	case "debug":
		logger.SetLevel(logger.DEBUG)
	case "info":
		logger.SetLevel(logger.INFO)
	case "warn":
		logger.SetLevel(logger.WARN)
	case "error":
		logger.SetLevel(logger.ERROR)
	case "fatal":
		logger.SetLevel(logger.FATAL)
	default:
		logger.SetLevel(logger.INFO)
	}

	// 处理版本信息
	if *showVersion {
		fmt.Printf("Exporter version %s\n", Version)
		return
	}

	// 处理帮助信息
	if *showHelp {
		pflag.Usage()
		return
	}

	// 处理更多帮助信息
	if *showMoreHelp {
		fmt.Println(cliDoc)
		return
	}

	// 获取非标志参数（即不带--或-的参数）
	args := pflag.Args()
	if len(args) != 0 {
		// 根据第一个参数判断子命令
		switch args[0] {
		case "cmd", "command":
			var cmd = args[1]
			args := args[2:]
			logger.Info("执行命令: %s, %s", cmd, args)

			flags := pflag.NewFlagSet("cmd", pflag.ExitOnError)
			version := flags.StringP("version", "v", "", "被调用的版本号")
			flags.Parse(args)
			//
			//// 必要参数检查
			//if *message == "" {
			//	fmt.Println("必须提供提交信息（-m）")
			//	cmd.PrintDefaults()
			//	os.Exit(1)
			//}

			var req = internal.CommandRequest{
				Cmd:     cmd,
				Version: *version,
				Args:    args,
			}
			var task = pkg.CreateTask(req, os.Stdout)
			pkg.ExecuteCommand(task)
		case "exec", "run":
			var cmd = args[1]
			args := args[2:]
			logger.Info("执行命令: %s, %s", cmd, args)

			flags := pflag.NewFlagSet("cmd", pflag.ExitOnError)
			encoding := flags.String("encoding", "", "被调用的版本号")
			flags.Parse(args)

			var encodingFunc func([]byte) string
			if *encoding != "" {
				encodingFunc = func(s []byte) string {
					return pkg.UtfFrom(s, internal.Charset(*encoding))
				}
			}

			c := exec.Command(cmd, args...)
			dir, err := os.Getwd()
			if err != nil {
				logger.Warn("获取当前工作目录失败: %v", err)
			}

			log := func(s string) {
				logger.Info(s)
			}
			status, err, stdout, stderr := pkg.Cmd(c, dir, []string{})
			if err != nil {
				logger.Warn("命令执行失败: %v", err)
			}

			pkg.CacthStdout(stdout, encodingFunc, log)

			pkg.CacthStderr(stderr, encodingFunc, log)

			logger.Info("命令执行完成: %v", status)
		default:
			fmt.Printf("未知子命令: %s\n", args[0])
			os.Exit(1)
		}
		return
	}

	// 创建路由器并应用中间件
	router := http.NewServeMux()
	router.HandleFunc("/", api.HandleRootRequest)
	router.HandleFunc("/cmd", api.HandleCommandRequest)
	router.HandleFunc("/cancel/", api.HandleCancelRequest)
	router.HandleFunc("/cmd/", api.HandleCommandRequest)
	router.HandleFunc("/result/", api.HandleResultRequest)
	router.HandleFunc("/backup/", api.HandleBackupRequest)

	// 使用中间件包装路由
	protectedRouter := recoveryMiddleware(router)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf(":%d", *port)
	logger.Info("启动exporter服务器，监听端口 %d...", *port)
	logger.Fatal("服务器停止: %v", http.ListenAndServe(serverAddr, protectedRouter))
}

// 新增恢复中间件
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("HTTP处理崩溃恢复: %v\n%s", err, debug.Stack())
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
