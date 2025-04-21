package pkg

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	intern "github.com/wangtengda/gobee/lvan/internal"
	"github.com/wangtengda/gobee/lvan/pkg/logger"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 任务信息
type Task struct {
	ID         string                `json:"id"`
	StartTime  time.Time             `json:"start_time"`
	EndTime    *time.Time            `json:"end_time,omitempty"`
	Request    intern.CommandRequest `json:"request"`
	Status     TaskStatus            `json:"status"` // running, completed, failed
	Result     *TaskResult           `json:"result"` // completed, failed, running, blocking
	CmdPath    string                `json:"cmd_path"`
	WorkDir    string                `json:"workdir"`
	Mutex      *sync.Mutex           `json:"-"`
	Logger     *logger.Logger        `json:"-"`
	sseClients *ClientManager
	CmdMeta    *intern.CommandMeta `json:"-"`
	Cancel     context.CancelFunc  `json:"-"`
}

// 添加输出到任务
func (t *Task) AddOutput(output string) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Result.Output += output

	if t.Logger != nil {
		t.Logger.Info("[%s]: %s", t.ID, output)
	}

	// 使用ClientManager广播消息给所有客户端
	if t.sseClients != nil {
		t.sseClients.Broadcast(output)
	}
}

// 完成任务
func (t *Task) Complete(status TaskStatus, exitCode int) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Status = status
	t.Result.ExitCode = int(exitCode)
	now := time.Now()
	t.EndTime = &now

	// 关闭客户端管理器，会自动关闭所有客户端连接
	if t.sseClients != nil {
		t.sseClients.Close()
	}

	// 发送任务完成的最终消息
	t.Result.Output += fmt.Sprintf("\nTask completed with status: %v, exit code: %d\n", status, t.Result.ExitCode)
	t.Logger.Close()
}

// 添加SSE客户端
func (t *Task) AddClient(clientID string) chan string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	// 如果任务已完成，返回nil
	if t.Status == Completed || t.Status == Failed {
		logger.Warn("尝试为已完成的任务 %s 添加客户端 %s", t.ID, clientID)
		return nil
	}

	// 使用ClientManager添加客户端
	if t.sseClients != nil {
		ch := t.sseClients.AddClient(clientID)

		// 如果是新客户端，发送已有的输出历史
		if ch != nil && t.Result.Output != "" {
			select {
			case ch <- t.Result.Output:
				// 发送成功
			default:
				// 通道已满，跳过历史输出
				logger.Warn("客户端 %s 通道已满，跳过历史输出", clientID)
			}
		}
		return ch
	}

	// 如果ClientManager不存在，创建一个
	t.sseClients = NewClientManager(1000, 30*time.Minute)
	return t.sseClients.AddClient(clientID)
}

// 移除SSE客户端
func (t *Task) RemoveClient(clientID string) {
	// 不需要锁定Task，ClientManager有自己的锁
	if t.sseClients != nil {
		t.sseClients.RemoveClient(clientID)
	}
}

// GetTaskDirectory 获取任务的工作目录
func GetTaskDirectory(taskID string) string {
	// 示例实现，根据实际项目调整
	TaskDir := filepath.Join(TasksDir, taskID)

	// 确保目录存在
	_ = os.MkdirAll(TaskDir, 0755)

	return TaskDir
}

var TasksDir string // 任务目录
type TaskStatus int

const (
	Completed TaskStatus = 0
	Failed    TaskStatus = 1
	Running   TaskStatus = 2
	Blocking  TaskStatus = 3
)

const (
	cmdNotExit = 400
	exclusive  = 401
	success    = 0
)

type TaskResult struct {
	ExitCode int      `json:"exit_code"` // 命令退出码
	Stderr   []string `json:"stderr"`    // new stderr
	Output   string   `json:"output"`    // stdout + stderr
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
func CreateTask(req intern.CommandRequest, w ...io.Writer) *Task {
	return taskManager.CreateTask(req, w...)
}

func (tm *TaskManager) CreateTask(req intern.CommandRequest, w ...io.Writer) *Task {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	taskID := uuid.New().String()
	workdir := GetTaskDirectory(taskID)
	logInstance, _ := logger.NewLogger(workdir, "output.log", logger.INFO, 10*1024*1024, w...)
	task := &Task{
		ID:        taskID,
		StartTime: time.Now(),
		Request:   req,
		Status:    Blocking,
		Result: &TaskResult{
			ExitCode: -1,
			Stderr:   []string{},
			Output:   "",
		},
		WorkDir:    workdir,
		Mutex:      &sync.Mutex{},
		Logger:     logInstance,
		sseClients: NewClientManager(1000, 30*time.Minute), // 设置最大1000个客户端，30分钟超时
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
