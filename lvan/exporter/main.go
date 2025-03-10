package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/wangtengda/gobee/lvan/exporter/config"
	"github.com/wangtengda/gobee/lvan/exporter/logger"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 版本信息
const (
	Version = "0.1.0"
)

// 命令请求结构
type CommandRequest struct {
	Cmd     string            `json:"cmd" yaml:"cmd"`
	Version string            `json:"version" yaml:"version"`
	Args    []string          `json:"args" yaml:"args"`
	Env     map[string]string `json:"env" yaml:"env"`
}

// 任务信息
type Task struct {
	ID        string                 `json:"id"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Request   CommandRequest         `json:"request"`
	Output    string                 `json:"output"`
	Status    string                 `json:"status"`    // running, completed, failed
	ExitCode  int                    `json:"exit_code"` // 命令退出码
	Mutex     *sync.Mutex            `json:"-"`
	Clients   map[string]chan string `json:"-"`
}

// 任务管理器
type TaskManager struct {
	Tasks map[string]*Task
	Mutex sync.Mutex
}

// 全局任务管理器
var taskManager = TaskManager{
	Tasks: make(map[string]*Task),
}

// 创建新任务
func (tm *TaskManager) CreateTask(req CommandRequest) *Task {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	taskID := uuid.New().String()
	task := &Task{
		ID:        taskID,
		StartTime: time.Now(),
		Request:   req,
		Status:    "running",
		Mutex:     &sync.Mutex{},
		Clients:   make(map[string]chan string),
	}

	tm.Tasks[taskID] = task
	return task
}

// 获取任务
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	task, exists := tm.Tasks[id]
	return task, exists
}

// 添加输出到任务
func (t *Task) AddOutput(output string) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Output += output

	// 向所有监听的客户端发送输出
	for _, ch := range t.Clients {
		select {
		case ch <- output:
		default:
			// 如果客户端没有准备好接收，跳过
		}
	}
}

// 完成任务
func (t *Task) Complete(status string, exitCode int) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Status = status
	t.ExitCode = exitCode
	now := time.Now()
	t.EndTime = &now

	// 关闭所有客户端通道
	for clientID, ch := range t.Clients {
		close(ch)
		delete(t.Clients, clientID)
	}
}

// 添加SSE客户端
func (t *Task) AddClient(clientID string) chan string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	ch := make(chan string, 100)
	t.Clients[clientID] = ch
	return ch
}

// 移除SSE客户端
func (t *Task) RemoveClient(clientID string) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	if ch, exists := t.Clients[clientID]; exists {
		close(ch)
		delete(t.Clients, clientID)
	}
}

// 处理只返回ID的请求
func handleOnlyIDRequest(w http.ResponseWriter, task *Task) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(task.ID))

	// 异步执行命令
	go executeCommand(task)
}

// 处理SSE请求
func handleSSERequest(w http.ResponseWriter, r *http.Request, task *Task) {
	// 设置SSE头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建客户端ID
	clientID := uuid.New().String()
	outputChan := task.AddClient(clientID)

	// 清理函数
	cleanup := func() {
		task.RemoveClient(clientID)
	}

	// 设置连接关闭时的清理
	go func() {
		<-r.Context().Done()
		cleanup()
	}()

	// 异步执行命令
	go executeCommand(task)

	// 发送任务ID
	fmt.Fprintf(w, "data: {\"id\": \"%s\"}\n\n", task.ID)
	w.(http.Flusher).Flush()

	// 发送输出流
	for output := range outputChan {
		// 确保每行输出都有正确的SSE格式
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Fprintf(w, "data: %s\n\n", line)
				w.(http.Flusher).Flush()
			}
		}
	}
}

// 处理同步执行命令请求
func handleSyncRequest(w http.ResponseWriter, task *Task) {
	// 同步执行命令
	executeCommand(task)

	// 根据任务状态设置HTTP状态码
	if task.Status == "failed" {
		w.WriteHeader(http.StatusInternalServerError) // 500
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
	} else {
		w.WriteHeader(http.StatusOK) // 200
		w.Header().Set("X-Exit-Code", "0")
	}

	// 返回结果
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(task.Output))
}

// 处理命令请求
func handleCommandRequest(w http.ResponseWriter, r *http.Request) {
	var task *Task
	switch r.Method {
	// 处理GET请求，格式为/cmd/command/param1/param2...
	case http.MethodGet:
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}
		cmd := pathParts[2]
		args := []string{}
		if len(pathParts) > 3 {
			// 处理URL路径中的参数，支持/和-引导的参数
			for _, arg := range pathParts[3:] {
				if arg != "" {
					// 处理特殊前缀，将__slash__转换为/开头的参数
					if strings.HasPrefix(arg, "__slash__") {
						arg = "/" + strings.TrimPrefix(arg, "__slash__")
					}
					args = append(args, arg)
				}
			}
		}

		req := CommandRequest{
			Cmd:     cmd,
			Version: "latest",
			Args:    args,
			Env:     make(map[string]string),
		}

		logger.Info("GET请求使用latest版本执行命令: %s", cmd)

		task = taskManager.CreateTask(req)

	// 处理POST请求
	case http.MethodPost:
		// 解析请求体格式
		bodyType := r.URL.Query().Get("body")
		if bodyType == "" {
			bodyType = "yaml" // 默认YAML
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// 解析请求体
		var req CommandRequest
		if bodyType == "json" {
			err = json.Unmarshal(body, &req)
		} else {
			err = yaml.Unmarshal(body, &req)
		}

		if err != nil {
			http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 验证请求
		if req.Cmd == "" {
			http.Error(w, "Command is required", http.StatusBadRequest)
			return
		}

		// 如果未指定版本，使用latest
		if req.Version == "" {
			req.Version = "latest"
			logger.Info("POST请求未指定版本，使用latest版本: %s", req.Cmd)
		}

		// 确保环境变量字段已初始化
		if req.Env == nil {
			req.Env = make(map[string]string)
		}

		// 创建任务
		task = taskManager.CreateTask(req)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否只返回ID
	onlyID := r.URL.Query().Get("onlyid") == "true"
	if onlyID {
		handleOnlyIDRequest(w, task)
		return
	}

	// 检查是否使用SSE
	useSSE := r.URL.Query().Get("sse") == "true"
	if useSSE {
		handleSSERequest(w, r, task)
	} else {
		handleSyncRequest(w, task)
	}
}

type charset string

const (
	UTF8    charset = "UTF-8"
	GB18030 charset = "GB18030"
)

// ByteToString 将字节切片转换为指定编码的字符串
func ByteToString(byte []byte, charset charset) string {
	var str string
	switch charset {
	case GB18030:
		decoder := simplifiedchinese.GB18030.NewDecoder()
		var err error
		str, err = decoder.String(string(byte))
		if err != nil {
			fmt.Println("解码错误:", err)
			return ""
		}
	case UTF8:
		str = string(byte)
	default:
		str = string(byte)
	}
	return str
}

// 执行命令
func executeCommand(task *Task) {
	// 记录开始执行
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", task.Request.Cmd))
	task.AddOutput(fmt.Sprintf("Version: %s\n", task.Request.Version))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(task.Request.Args, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 版本: %s, 参数: %s", task.Request.Cmd, task.Request.Version, strings.Join(task.Request.Args, ", "))

	// 获取命令路径
	version := task.Request.Version
	if version == "" {
		version = "latest"
		logger.Info("未指定版本，使用latest版本: %s", task.Request.Cmd)
	}

	// 使用版本管理获取可执行文件路径
	cmdPath, found, err := config.GetCommandPath(task.Request.Cmd, version)
	if err != nil || !found {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", task.Request.Cmd, version, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete("failed", 1)
		return
	}

	// 记录使用的可执行文件路径
	logger.Info("使用可执行文件: %s", cmdPath)
	task.AddOutput(fmt.Sprintf("Using executable: %s\n", cmdPath))

	var cmd *exec.Cmd
	// 检查是否是 Windows 平台
	if runtime.GOOS == "windows" {
		// 检查文件扩展名是否为批处理文件
		ext := strings.ToLower(filepath.Ext(cmdPath))
		if ext == ".bat" || ext == ".cmd" {
			// 使用 cmd /c 执行批处理文件
			newArgs := append([]string{"/c", cmdPath}, task.Request.Args...)
			cmd = exec.Command("cmd", newArgs...)
		} else {
			// 非批处理文件直接执行
			cmd = exec.Command(cmdPath, task.Request.Args...)
		}
	} else {
		// 非 Windows 平台直接执行命令
		cmd = exec.Command(cmdPath, task.Request.Args...)
	}

	// 获取当前环境变量
	cmd.Env = os.Environ()

	// 设置环境变量
	if len(task.Request.Env) > 0 {

		// 添加自定义环境变量
		for key, value := range task.Request.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
			logger.Debug("设置环境变量: %s=%s", key, value)
		}
	}

	// 设置工作目录（任务沙箱）
	taskDir := getTaskDirectory(task.ID)
	cmd.Dir = taskDir

	// 创建管道获取命令输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("创建stdout管道失败: %s", err.Error())
		task.AddOutput(fmt.Sprintf("Error creating stdout pipe: %s\n", err.Error()))
		task.Complete("failed", 1)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("创建stderr管道失败: %s", err.Error())
		task.AddOutput(fmt.Sprintf("Error creating stderr pipe: %s\n", err.Error()))
		task.Complete("failed", 1)
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Error("启动命令失败: %s", err.Error())
		task.AddOutput(fmt.Sprintf("Error starting command: %s\n", err.Error()))
		task.Complete("failed", 1)
		return
	}

	// 读取标准输出
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			output := scanner.Bytes()
			convertGbToUtf8 := ByteToString(output, GB18030)
			task.AddOutput(convertGbToUtf8 + "\n")
			logger.Info("命令输出: %s", convertGbToUtf8)
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			error := scanner.Bytes()
			convertGbToUtf8 := ByteToString(error, GB18030)
			task.AddOutput("ERROR: " + convertGbToUtf8 + "\n")
			logger.Warn("命令错误输出: %s", convertGbToUtf8)
		}
	}()

	// 等待命令完成
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		// 尝试获取退出码
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1 // 默认错误码
		}
		logger.Warn("命令执行失败，退出码 %d: %s", exitCode, err.Error())
		task.AddOutput(fmt.Sprintf("Command failed with exit code %d: %s\n", exitCode, err.Error()))
		task.Complete("failed", exitCode)
	} else {
		logger.Info("命令执行成功，退出码 0")
		task.AddOutput("Command completed successfully with exit code 0!\n")
		task.Complete("completed", 0)
	}
}

// getTaskDirectory 获取任务的工作目录
func getTaskDirectory(taskID string) string {
	// 示例实现，根据实际项目调整
	tasksDir := "tasks"
	taskDir := filepath.Join(tasksDir, taskID)

	// 确保目录存在
	os.MkdirAll(taskDir, 0755)

	return taskDir
}

// 处理结果请求
func handleResultRequest(w http.ResponseWriter, r *http.Request) {
	// 解析路径中的任务ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	taskID := pathParts[2]

	// 处理帮助请求
	if taskID == "help" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>API Documentation</h1><p>This is the API documentation for the exporter service.</p></body></html>"))
		return
	}

	// 获取任务
	task, exists := taskManager.GetTask(taskID)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// 检查是否使用SSE
	useSSE := r.URL.Query().Get("sse") == "true"
	if useSSE {
		// 设置SSE头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 创建客户端ID
		clientID := uuid.New().String()
		outputChan := task.AddClient(clientID)

		// 清理函数
		cleanup := func() {
			task.RemoveClient(clientID)
		}

		// 设置连接关闭时的清理（使用context替代已弃用的http.CloseNotifier）
		go func() {
			<-r.Context().Done()
			cleanup()
		}()

		// 发送现有输出
		task.Mutex.Lock()
		existingOutput := task.Output
		task.Mutex.Unlock()

		if existingOutput != "" {
			fmt.Fprintf(w, "data: %s\n\n", existingOutput)
			w.(http.Flusher).Flush()
		}

		// 如果任务已完成，关闭连接
		task.Mutex.Lock()
		isCompleted := task.Status != "running"
		task.Mutex.Unlock()

		if isCompleted {
			cleanup()
			return
		}

		// 发送新输出
		for output := range outputChan {
			// 确保每行输出都有正确的SSE格式
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Fprintf(w, "data: %s\n\n", line)
					w.(http.Flusher).Flush()
				}
			}
		}
	} else {
		// 检查任务是否仍在运行
		task.Mutex.Lock()
		isRunning := task.Status == "running"
		output := task.Output
		status := task.Status
		exitCode := task.ExitCode
		task.Mutex.Unlock()

		if isRunning {
			w.WriteHeader(http.StatusAccepted) // 202
			return
		}

		// 根据任务状态设置HTTP状态码
		if status == "failed" {
			w.WriteHeader(http.StatusInternalServerError) // 500
			w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", exitCode))
		} else {
			w.WriteHeader(http.StatusOK) // 200
			w.Header().Set("X-Exit-Code", "0")
		}

		// 返回结果
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output))
	}
}

func main() {
	// 解析命令行参数
	port := flag.Int("p", 80, "Port to listen on")
	portLong := flag.Int("port", 80, "Port to listen on")
	showVersion := flag.Bool("v", false, "Show version")
	showVersionLong := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("h", false, "Show help")
	showHelpLong := flag.Bool("help", false, "Show help")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error, fatal)")

	flag.Parse()

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
	if *showVersion || *showVersionLong {
		fmt.Printf("Exporter version %s\n", Version)
		return
	}

	// 处理帮助信息
	if *showHelp || *showHelpLong {
		flag.Usage()
		return
	}

	// 确定使用哪个端口
	listenPort := *port
	if *portLong != 80 {
		listenPort = *portLong
	}

	// 设置HTTP路由
	http.HandleFunc("/cmd", handleCommandRequest)
	http.HandleFunc("/cmd/", handleCommandRequest)
	http.HandleFunc("/result/", handleResultRequest)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf(":%d", listenPort)
	logger.Info("启动exporter服务器，监听端口 %d...", listenPort)
	logger.Fatal("服务器停止: %v", http.ListenAndServe(serverAddr, nil))
}
