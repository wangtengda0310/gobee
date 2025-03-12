package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/spf13/pflag"
	"github.com/wangtengda/gobee/lvan/exporter/config"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/wangtengda/gobee/lvan/exporter/logger"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// 嵌入HTTP文档
//
//go:embed http-doc.txt
var httpDoc string

// 嵌入CLI文档
//
//go:embed cli-doc.txt
var cliDoc string

// 版本信息
const (
	Version = "0.0.0"
)

// 命令请求结构
type CommandRequest struct {
	Cmd     string            `json:"cmd" yaml:"cmd"`
	Version string            `json:"version" yaml:"version"`
	Args    []string          `json:"args" yaml:"args"`
	Env     map[string]string `json:"-" yaml:"env,omitempty"`
}

// 客户端信息
type Client struct {
	ID           string      // 客户端唯一标识
	OutputChan   chan string // 输出通道
	LastActivity time.Time   // 最后活动时间
	Active       bool        // 客户端是否活跃
}

// 客户端管理器，使用分片锁减少锁竞争
type ClientManager struct {
	Clients     map[string]*Client // 客户端映射
	Mutex       sync.RWMutex       // 读写锁
	BroadcastCh chan string        // 广播通道
	MaxClients  int                // 最大客户端数量
	IdleTimeout time.Duration      // 客户端空闲超时时间
	shutdown    chan struct{}      // 关闭信号
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
	CmdPath   string                 `json:"cmd_path"`
	WorkDir   string                 `json:"workdir"`
	Mutex     *sync.Mutex            `json:"-"`
	Logger    *logger.Logger         `json:"-"`
	Clients   map[string]chan string `json:"-"`
	ClientMgr *ClientManager         `json:"-"`
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

// 创建新的客户端管理器
func NewClientManager(maxClients int, idleTimeout time.Duration) *ClientManager {
	cm := &ClientManager{
		Clients:     make(map[string]*Client),
		BroadcastCh: make(chan string, 100),
		MaxClients:  maxClients,
		IdleTimeout: idleTimeout,
		shutdown:    make(chan struct{}),
	}

	// 启动广播处理协程
	go cm.broadcastWorker()
	// 启动空闲客户端清理协程
	go cm.cleanupWorker()

	return cm
}

// 广播工作协程
func (cm *ClientManager) broadcastWorker() {
	for {
		select {
		case msg := <-cm.BroadcastCh:
			// 读锁保护，允许并发读取
			cm.Mutex.RLock()
			for _, client := range cm.Clients {
				if client.Active {
					select {
					case client.OutputChan <- msg:
						// 消息发送成功
					default:
						// 通道已满，跳过
					}
				}
			}
			cm.Mutex.RUnlock()
		case <-cm.shutdown:
			return
		}
	}
}

// 清理空闲客户端
func (cm *ClientManager) cleanupWorker() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.Mutex.Lock()
			now := time.Now()
			for id, client := range cm.Clients {
				if now.Sub(client.LastActivity) > cm.IdleTimeout {
					// 关闭通道
					close(client.OutputChan)
					// 从映射中删除
					delete(cm.Clients, id)
					logger.Info("客户端 %s 因空闲超时被清理", id)
				}
			}
			cm.Mutex.Unlock()
		case <-cm.shutdown:
			return
		}
	}
}

// 添加客户端
func (cm *ClientManager) AddClient(clientID string) chan string {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	// 检查是否达到最大客户端数量限制
	if cm.MaxClients > 0 && len(cm.Clients) >= cm.MaxClients {
		logger.Warn("达到最大客户端数量限制 %d，拒绝新客户端 %s", cm.MaxClients, clientID)
		return nil
	}

	// 创建新客户端
	client := &Client{
		ID:           clientID,
		OutputChan:   make(chan string, 100),
		LastActivity: time.Now(),
		Active:       true,
	}

	cm.Clients[clientID] = client
	logger.Info("添加新客户端 %s，当前客户端数量: %d", clientID, len(cm.Clients))
	return client.OutputChan
}

// 移除客户端
func (cm *ClientManager) RemoveClient(clientID string) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	if client, exists := cm.Clients[clientID]; exists {
		close(client.OutputChan)
		delete(cm.Clients, clientID)
		logger.Info("移除客户端 %s，当前客户端数量: %d", clientID, len(cm.Clients))
	}
}

// 广播消息给所有客户端
func (cm *ClientManager) Broadcast(msg string) {
	select {
	case cm.BroadcastCh <- msg:
		// 消息已放入广播通道
	default:
		// 广播通道已满，记录警告
		logger.Warn("广播通道已满，消息丢弃")
	}
}

// 关闭客户端管理器
func (cm *ClientManager) Close() {
	// 发送关闭信号
	close(cm.shutdown)

	// 关闭所有客户端连接
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	for id, client := range cm.Clients {
		close(client.OutputChan)
		delete(cm.Clients, id)
	}

	logger.Info("客户端管理器已关闭")
}

// 创建新任务
func (tm *TaskManager) CreateTask(req CommandRequest) *Task {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	taskID := uuid.New().String()
	workdir := getTaskDirectory(taskID)
	logInstance, _ := logger.NewLogger(workdir, "output.log", logger.INFO, 10*1024*1024)
	task := &Task{
		ID:        taskID,
		StartTime: time.Now(),
		Request:   req,
		Status:    "running",
		WorkDir:   workdir,
		Mutex:     &sync.Mutex{},
		Logger:    logInstance,
		ClientMgr: NewClientManager(1000, 30*time.Minute), // 设置最大1000个客户端，30分钟超时
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

	if t.Logger != nil {
		t.Logger.Info("TASK OUTPUT: %s", output)
	}

	// 使用ClientManager广播消息给所有客户端
	if t.ClientMgr != nil {
		t.ClientMgr.Broadcast(output)
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

	// 关闭客户端管理器，会自动关闭所有客户端连接
	if t.ClientMgr != nil {
		t.ClientMgr.Close()
	}

	// 发送任务完成的最终消息
	completionMsg := fmt.Sprintf("\nTask completed with status: %s, exit code: %d\n", status, exitCode)
	t.Output += completionMsg
}

// 添加SSE客户端
func (t *Task) AddClient(clientID string) chan string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	// 如果任务已完成，返回nil
	if t.Status == "completed" || t.Status == "failed" {
		logger.Warn("尝试为已完成的任务 %s 添加客户端 %s", t.ID, clientID)
		return nil
	}

	// 使用ClientManager添加客户端
	if t.ClientMgr != nil {
		ch := t.ClientMgr.AddClient(clientID)

		// 如果是新客户端，发送已有的输出历史
		if ch != nil && t.Output != "" {
			select {
			case ch <- t.Output:
				// 发送成功
			default:
				// 通道已满，跳过历史输出
				logger.Warn("客户端 %s 通道已满，跳过历史输出", clientID)
			}
		}
		return ch
	}

	// 如果ClientManager不存在，创建一个
	t.ClientMgr = NewClientManager(1000, 30*time.Minute)
	return t.ClientMgr.AddClient(clientID)
}

// 移除SSE客户端
func (t *Task) RemoveClient(clientID string) {
	// 不需要锁定Task，ClientManager有自己的锁
	if t.ClientMgr != nil {
		t.ClientMgr.RemoveClient(clientID)
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
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建客户端ID
	clientID := uuid.New().String()
	outputChan := task.AddClient(clientID)

	// 如果无法添加客户端（例如任务已完成或达到最大连接数）
	if outputChan == nil {
		http.Error(w, "Cannot connect to task: either completed or connection limit reached", http.StatusServiceUnavailable)
		return
	}

	// 设置连接关闭时的清理
	go func() {
		<-r.Context().Done()
		task.RemoveClient(clientID)
		logger.Info("SSE客户端 %s 连接已关闭", clientID)
	}()

	// 异步执行命令（如果尚未执行）
	if task.Status == "running" && task.CmdPath == "" {
		go executeCommand(task)
	}

	// 发送任务ID
	fmt.Fprintf(w, "data: {\"id\": \"%s\", \"status\": \"%s\"}\n\n", task.ID, task.Status)
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

	// 如果输出通道关闭但任务仍在运行，发送最终状态
	task.Mutex.Lock()
	if task.Status != "running" {
		fmt.Fprintf(w, "data: {\"status\": \"%s\", \"exitCode\": %d}\n\n", task.Status, task.ExitCode)
		w.(http.Flusher).Flush()
	}
	task.Mutex.Unlock()
}

type CmdResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Id   string `json:"id"`
}
type ResultResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Id   string         `json:"id"`
	Job  CommandRequest `json:"job"`
}

// 处理同步执行命令请求
func handleSyncRequest(w http.ResponseWriter, task *Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	isRunning := task.Status
	task.Mutex.Unlock()

	var res *CmdResponse
	// 根据任务状态设置HTTP状态码
	if isRunning == "failed" {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &CmdResponse{
			Code: 1,
			Msg:  "任务执行失败",
			Id:   task.ID,
		}
	} else if isRunning == "running" {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &CmdResponse{
			Code: 1,
			Msg:  "任务处理中",
			Id:   task.ID,
		}
	} else {
		w.Header().Set("X-Exit-Code", "0")
		res = &CmdResponse{
			Code: 0,
			Msg:  "任务执行成功",
			Id:   task.ID,
		}
	}

	// 返回结果
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(task.Output))
	w.Header().Set("Content-Type", "application/json")
	marshal, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		w.Write([]byte(fmt.Sprintf("{\"code\":1,\"msg\":\"序列化错误\",\"id\":\"%s\"}", task.ID)))
		return
	}
	w.Write(marshal)
}

// 处理同步执行命令请求 // todo 跟新见任务的方法合并
func handleSyncResultRequest(w http.ResponseWriter, task *Task) {
	// 检查任务是否仍在运行
	task.Mutex.Lock()
	isRunning := task.Status
	task.Mutex.Unlock()

	var res *ResultResponse
	// 根据任务状态设置HTTP状态码
	if isRunning == "failed" {
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &ResultResponse{
			Code: 1,
			Msg:  "任务执行失败",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else if isRunning == "running" {
		w.WriteHeader(http.StatusAccepted) // 202
		w.Header().Set("X-Exit-Code", fmt.Sprintf("%d", task.ExitCode))
		res = &ResultResponse{
			Code: 1,
			Msg:  "任务处理中",
			Id:   task.ID,
			Job:  task.Request,
		}
	} else {
		w.Header().Set("X-Exit-Code", "0")
		res = &ResultResponse{
			Code: 0,
			Msg:  "任务执行成功",
			Id:   task.ID,
			Job:  task.Request,
		}
	}

	// 返回结果
	//w.Header().Set("Content-Type", "text/plain")
	//w.Write([]byte(task.Output))
	w.Header().Set("Content-Type", "application/json")
	marshal, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) // 500
		bytes, err := json.Marshal(task.Request)
		var job string
		if err == nil {
			job = string(bytes)
		}
		w.Write([]byte(fmt.Sprintf("{\"code\":1,\"msg\":\"序列化错误\",\"id\":\"%s\",\"job\":\"%s\"}", task.ID, job)))
		return
	}
	w.Write(marshal)
}

// 处理根路径请求
func handleRootRequest(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径请求
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 使用嵌入的HTTP文档内容
	w.Write([]byte(httpDoc))
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
			Version: "",
			Args:    args,
			Env:     make(map[string]string),
		}

		logger.Info("GET请求需要适配可用版本执行命令")

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
		return
	}

	handleSyncRequest(w, task)
	w.(http.Flusher).Flush()

	// 同步执行命令
	go executeCommand(task)
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

	// 使用版本管理获取可执行文件路径
	cmdPath, found, err := config.GetCommandPath(task.Request.Cmd, task.Request.Version)
	if err != nil || !found {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", task.Request.Cmd, task.Request.Version, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete("failed", 1)
		return
	}

	task.CmdPath = cmdPath

	// 记录使用的可执行文件路径
	logger.Info("使用可执行文件: %s", cmdPath)
	task.AddOutput(fmt.Sprintf("Using executable: %s\n", cmdPath))

	var timeout = 30 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	// 检查是否是 Windows 平台
	if runtime.GOOS == "windows" {
		// 检查文件扩展名是否为批处理文件
		ext := strings.ToLower(filepath.Ext(task.CmdPath))
		if ext == ".bat" || ext == ".cmd" {
			// 使用 cmd /c 执行批处理文件
			newArgs := append([]string{"/c", task.CmdPath}, task.Request.Args...)
			cmd = exec.CommandContext(ctx, "cmd", newArgs...)
		} else {
			// 非批处理文件直接执行
			cmd = exec.CommandContext(ctx, task.CmdPath, task.Request.Args...)
		}
	} else {
		// 非 Windows 平台直接执行命令
		cmd = exec.CommandContext(ctx, task.CmdPath, task.Request.Args...)
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
	cmd.Dir = task.WorkDir

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
		// 处理超时错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			task.AddOutput(fmt.Sprintf("命令执行超时 %v\n", timeout))
			exitCode = 124 // 通常用 124 表示超时退出码
		} else if exitErr, ok := err.(*exec.ExitError); ok { // 尝试获取退出码
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
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 创建客户端ID
		clientID := uuid.New().String()
		outputChan := task.AddClient(clientID)

		// 如果无法添加客户端（例如任务已完成或达到最大连接数）
		if outputChan == nil {
			http.Error(w, "Cannot connect to task: either completed or connection limit reached", http.StatusServiceUnavailable)
			return
		}

		// 设置连接关闭时的清理
		go func() {
			<-r.Context().Done()
			task.RemoveClient(clientID)
			logger.Info("结果SSE客户端 %s 连接已关闭", clientID)
		}()

		// 添加健康检查ping
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					// 发送ping消息
					fmt.Fprintf(w, "event: ping\ndata: %d\n\n", time.Now().Unix())
					w.(http.Flusher).Flush()
				case <-r.Context().Done():
					return
				}
			}
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
		handleSyncResultRequest(w, task)
	}
}

const tasksDir = "tasks"

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
	pflag.Parse()

	if *cleanFlag {
		cleanGeneratedFiles(tasksDir)
		return // 清理后直接退出
	}

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

	// 设置HTTP路由
	http.HandleFunc("/", handleRootRequest)
	http.HandleFunc("/cmd", handleCommandRequest)
	http.HandleFunc("/cmd/", handleCommandRequest)
	http.HandleFunc("/result/", handleResultRequest)

	// 启动HTTP服务器
	serverAddr := fmt.Sprintf(":%d", *port)
	logger.Info("启动exporter服务器，监听端口 %d...", *port)
	logger.Fatal("服务器停止: %v", http.ListenAndServe(serverAddr, nil))
}
