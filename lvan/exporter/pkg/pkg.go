package pkg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	intern "github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// 执行命令
func ExecuteCommand(task *Task) {
	// 记录开始执行
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", task.Request.Cmd))
	task.AddOutput(fmt.Sprintf("Version: %s\n", task.Request.Version))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(task.Request.Args, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 版本: %s, 参数: %s", task.Request.Cmd, task.Request.Version, strings.Join(task.Request.Args, ", "))

	// 使用版本管理获取可执行文件路径
	versionDir, found, err := GetCommandVersionPath(task.Request.Cmd, task.Request.Version)
	if err != nil || !found {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", task.Request.Cmd, task.Request.Version, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete("failed", 2)
		return
	}
	task.CmdMeta = intern.TryMeta(filepath.Join(versionDir, "meta.yaml"))

	// 查找可执行文件
	executable, found, err := findExecutable(versionDir, task.Request.Cmd)
	if err != nil || !found {
		errMsg := fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", task.Request.Cmd, task.Request.Version, err)
		logger.Error(errMsg)
		task.AddOutput(errMsg)
		task.Complete("failed", 2)
		return
	}

	task.CmdPath = executable

	// 记录使用的可执行文件路径
	logger.Info("使用可执行文件: %s", executable)
	task.AddOutput(fmt.Sprintf("Using executable: %s\n", executable))

	var timeout = 30 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if task.CmdMeta != nil && task.CmdMeta.Shell != nil {
		newArgs := append(task.CmdMeta.Shell[1:], task.CmdPath)
		newArgs = append(newArgs, task.Request.Args...)
		cmd = exec.CommandContext(ctx, task.CmdMeta.Shell[0], newArgs...)

	} else if runtime.GOOS == "windows" {
		// 检查是否是 Windows 平台 尝试使用cmd执行.bat和.cmd
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
	var resource string
	if task.CmdMeta != nil && len(task.CmdMeta.Resources) > 0 {
		retries := 3
		logger.Info("默认重试次数为 %d 可以通过环境变量 exporter_retry_times 设置", retries)
		if os.Getenv("exporter_retry_times") != "" {
			retry, err := strconv.Atoi(os.Getenv("exporter_retry_times"))
			if err == nil {
				logger.Info("使用环境变量 exporter_retry_times 设置重试次数为 %d", retry)
				retries = retry
			}
		}
		resource, err = intern.ExclusiveOneResource(task.CmdMeta.Resources, TasksDir, retries)
		if err != nil {
			// 无法获取资源，记录错误并继续执行
			logger.Warn("无法获取资源锁: %v，任务将继续执行但可能影响性能", err)
			task.AddOutput("超时获取资源锁\n")
			task.Complete("blocking", 3)
			return
		}
		// 添加自定义环境变量
		cmd.Env = append(cmd.Env, fmt.Sprintf("exporter_cmd_%s_resource=%s", task.Request.Cmd, resource))
	}

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
		task.Complete("failed", 2)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("创建stderr管道失败: %s", err.Error())
		task.AddOutput(fmt.Sprintf("Error creating stderr pipe: %s\n", err.Error()))
		task.Complete("failed", 2)
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		logger.Error("启动命令失败: %s", err.Error())
		task.AddOutput(fmt.Sprintf("Error starting command: %s\n", err.Error()))
		task.Complete("failed", 2)
		return
	}

	// 读取标准输出
	// 读取标准输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var encoding intern.Charset
			if task.CmdMeta != nil && task.CmdMeta.Encoding != "" {
				encoding = task.CmdMeta.Encoding
			}
			convertGbToUtf8 := ByteToString(scanner.Bytes(), encoding)
			task.AddOutput(convertGbToUtf8 + "\n")
			logger.Info("命令输出: %s", convertGbToUtf8)
		}
	}()

	// 读取标准错误
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			var encoding intern.Charset
			if task.CmdMeta != nil && task.CmdMeta.Encoding != "" {
				encoding = task.CmdMeta.Encoding
			}
			convertGbToUtf8 := ByteToString(scanner.Bytes(), encoding)
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

	// 释放资源锁
	if task.CmdMeta != nil && task.Request.Cmd != "" {
		if resource != "" {
			if err := intern.ReleaseResource(resource); err != nil {
				logger.Error("释放资源锁失败: %s, %v", resource, err)
			} else {
				logger.Info("成功释放资源锁: %s", resource)
			}
		}
	}

}

// 任务信息
type Task struct {
	ID        string                 `json:"id"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Request   intern.CommandRequest  `json:"request"`
	Output    string                 `json:"output"`
	Status    string                 `json:"status"`    // running, completed, failed
	ExitCode  int                    `json:"exit_code"` // 命令退出码
	CmdPath   string                 `json:"cmd_path"`
	WorkDir   string                 `json:"workdir"`
	Mutex     *sync.Mutex            `json:"-"`
	Logger    *logger.Logger         `json:"-"`
	Clients   map[string]chan string `json:"-"`
	ClientMgr *ClientManager         `json:"-"`
	CmdMeta   *intern.CommandMeta    `json:"-"`
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

const (
	UTF8    intern.Charset = "UTF-8"
	utf8    intern.Charset = "utf-8"
	GB18030 intern.Charset = "GB18030"
	gb18030 intern.Charset = "gb18030"
	GBK     intern.Charset = "GBK"
	gbk     intern.Charset = "gbk"
)

// ByteToString 将字节切片转换为指定编码的字符串
func ByteToString(byte []byte, charset intern.Charset) string {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("解码错误:%v", err)
		}
	}()
	var str string
	switch charset {
	case GB18030, gb18030:
		decoder := simplifiedchinese.GB18030.NewDecoder()
		var err error
		str, err = decoder.String(string(byte))
		if err != nil {
			return ""
		}
	case GBK, gbk:
		decoder := simplifiedchinese.GBK.NewDecoder()
		var err error
		str, err = decoder.String(string(byte))
		if err != nil {
			panic(err)
		}
	case UTF8, utf8:
		str = string(byte)
	default:
		str = string(byte)
	}
	return str
}

// GetTaskDirectory 获取任务的工作目录
func getTaskDirectory(taskID string) string {
	// 示例实现，根据实际项目调整
	TaskDir := filepath.Join(TasksDir, taskID)

	// 确保目录存在
	os.MkdirAll(TaskDir, 0755)

	return TaskDir
}

var TasksDir string // 任务目录

// 创建新任务
func CreateTask(req intern.CommandRequest) *Task {
	return taskManager.CreateTask(req)
}

func (tm *TaskManager) CreateTask(req intern.CommandRequest) *Task {
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
		ExitCode:  2,
		WorkDir:   workdir,
		Mutex:     &sync.Mutex{},
		Logger:    logInstance,
		ClientMgr: NewClientManager(1000, 30*time.Minute), // 设置最大1000个客户端，30分钟超时
	}

	// 构建日志文件路径
	filePath := filepath.Join(workdir, "createtimestamp")
	intern.Create(filePath)

	tm.Tasks[taskID] = task
	return task
}

// 获取任务
func GetTask(id string) (*Task, bool) {
	return taskManager.GetTask(id)
}
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	task, exists := tm.Tasks[id]
	return task, exists
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
