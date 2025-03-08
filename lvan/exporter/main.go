package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/wangtengda/gobee/lvan/exporter/logger"
)

// 版本信息
const (
	Version = "0.1.0"
)

// 命令请求结构
type CommandRequest struct {
	Cmd     string   `json:"cmd" yaml:"cmd"`
	Version string   `json:"version" yaml:"version"`
	Args    []string `json:"args" yaml:"args"`
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

// 处理命令请求
func handleCommandRequest(w http.ResponseWriter, r *http.Request) {
	// 处理GET请求，格式为/cmd/command/param1/param2...
	if r.Method == http.MethodGet {
		// 解析URL路径
		pathParts := strings.Split(r.URL.Path, "/")

		// 检查路径格式是否正确
		if len(pathParts) < 3 {
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}

		// 获取命令和参数
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

		// 创建命令请求
		req := CommandRequest{
			Cmd:     cmd,
			Version: Version,
			Args:    args,
		}

		// 创建任务并处理
		task := taskManager.CreateTask(req)

		// 检查是否只返回ID
		onlyID := r.URL.Query().Get("onlyid") == "true"
		if onlyID {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(task.ID))

			// 异步执行命令
			go executeCommand(task)
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

			// 设置连接关闭时的清理（使用更现代的方式替代已弃用的http.CloseNotifier）
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
		} else {
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
		return
	}

	// 处理POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// 创建任务
	task := taskManager.CreateTask(req)

	// 检查是否只返回ID
	onlyID := r.URL.Query().Get("onlyid") == "true"
	if onlyID {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(task.ID))

		// 异步执行命令
		go executeCommand(task)
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

		// 设置连接关闭时的清理
		notifier, ok := w.(http.CloseNotifier)
		if ok {
			go func() {
				<-notifier.CloseNotify()
				cleanup()
			}()
		}

		// 异步执行命令
		go executeCommand(task)

		// 发送任务ID
		fmt.Fprintf(w, "data: {\"id\": \"%s\"}\n\n", task.ID)
		w.(http.Flusher).Flush()

		// 发送输出流
		for output := range outputChan {
			fmt.Fprintf(w, "data: %s\n\n", output)
			w.(http.Flusher).Flush()
		}
	} else {
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
}

// 执行命令
func executeCommand(task *Task) {
	// 记录开始执行
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", task.Request.Cmd))
	task.AddOutput(fmt.Sprintf("Version: %s\n", task.Request.Version))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(task.Request.Args, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 参数: %s", task.Request.Cmd, strings.Join(task.Request.Args, ", "))

	// 调用系统命令
	cmd := exec.Command(task.Request.Cmd, task.Request.Args...)

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
			output := scanner.Text()
			task.AddOutput(output + "\n")
			logger.Info("命令输出: %s", output)
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			error := scanner.Text()
			task.AddOutput("ERROR: " + error + "\n")
			logger.Warn("命令错误输出: %s", error)
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

		// 设置连接关闭时的清理
		notifier, ok := w.(http.CloseNotifier)
		if ok {
			go func() {
				<-notifier.CloseNotify()
				cleanup()
			}()
		}

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
