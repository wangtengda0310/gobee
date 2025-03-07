package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
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
	ID        string          `json:"id"`
	StartTime time.Time       `json:"start_time"`
	EndTime   *time.Time      `json:"end_time,omitempty"`
	Request   CommandRequest  `json:"request"`
	Output    string          `json:"output"`
	Status    string          `json:"status"` // running, completed, failed
	Mutex     *sync.Mutex     `json:"-"`
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
func (t *Task) Complete(status string) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Status = status
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

		// 返回结果
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(task.Output))
	}
}

// 执行命令
func executeCommand(task *Task) {
	// 这里应该实现实际的命令执行逻辑
	// 为了演示，我们只是模拟一些输出
	
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", task.Request.Cmd))
	task.AddOutput(fmt.Sprintf("Version: %s\n", task.Request.Version))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(task.Request.Args, ", ")))
	
	// 模拟命令执行
	for i := 0; i < 5; i++ {
		time.Sleep(500 * time.Millisecond)
		task.AddOutput(fmt.Sprintf("Processing step %d...\n", i+1))
	}
	
	task.AddOutput("Command completed successfully!\n")
	task.Complete("completed")
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
			fmt.Fprintf(w, "data: %s\n\n", output)
			w.(http.Flusher).Flush()
		}
	} else {
		// 检查任务是否仍在运行
		task.Mutex.Lock()
		isRunning := task.Status == "running"
		output := task.Output
		task.Mutex.Unlock()

		if isRunning {
			w.WriteHeader(http.StatusAccepted) // 202
			return
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

	flag.Parse()

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
	http.HandleFunc("/result/", handleResultRequest)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf(":%d", listenPort)
	fmt.Printf("Starting exporter server on port %d...\n", listenPort)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}