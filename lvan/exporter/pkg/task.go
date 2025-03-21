package pkg

import (
	"fmt"
	"github.com/google/uuid"
	intern "github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 任务信息
type Task struct {
	ID        string                 `json:"id"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Request   intern.CommandRequest  `json:"request"`
	Output    string                 `json:"output"`
	Status    TaskStatus             `json:"status"` // completed, failed, running, blocking
	CmdPath   string                 `json:"cmd_path"`
	WorkDir   string                 `json:"workdir"`
	Mutex     *sync.Mutex            `json:"-"`
	Logger    *logger.Logger         `json:"-"`
	Clients   map[string]chan string `json:"-"`
	ClientMgr *ClientManager         `json:"-"`
	CmdMeta   *intern.CommandMeta    `json:"-"`
}

// 添加输出到任务
func (t *Task) AddOutput(output string) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Output += output

	if t.Logger != nil {
		t.Logger.Info("[%s]: %s", t.ID, output)
	}

	// 使用ClientManager广播消息给所有客户端
	if t.ClientMgr != nil {
		t.ClientMgr.Broadcast(output)
	}
}

// 完成任务
func (t *Task) Complete(status TaskStatus) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Status = status
	now := time.Now()
	t.EndTime = &now

	// 关闭客户端管理器，会自动关闭所有客户端连接
	if t.ClientMgr != nil {
		t.ClientMgr.Close()
	}

	// 发送任务完成的最终消息
	completionMsg := fmt.Sprintf("\nTask completed with status: %s, exit code: %d\n", status, status.ExitCode)
	t.Output += completionMsg
	t.Logger.Close()
}

// 添加SSE客户端
func (t *Task) AddClient(clientID string) chan string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	// 如果任务已完成，返回nil
	if t.Status.Status == "completed" || t.Status.Status == "failed" {
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

// GetTaskDirectory 获取任务的工作目录
func GetTaskDirectory(taskID string) string {
	// 示例实现，根据实际项目调整
	TaskDir := filepath.Join(TasksDir, taskID)

	// 确保目录存在
	os.MkdirAll(TaskDir, 0755)

	return TaskDir
}

var TasksDir string // 任务目录

type TaskStatus struct {
	Status   string `json:"status"`    // running, completed, failed
	ExitCode int    `json:"exit_code"` // 命令退出码
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
func CreateTask(req intern.CommandRequest) *Task {
	return taskManager.CreateTask(req)
}

func (tm *TaskManager) CreateTask(req intern.CommandRequest) *Task {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()

	taskID := uuid.New().String()
	workdir := GetTaskDirectory(taskID)
	logInstance, _ := logger.NewLogger(workdir, "output.log", logger.INFO, 10*1024*1024)
	task := &Task{
		ID:        taskID,
		StartTime: time.Now(),
		Request:   req,
		Status:    blocking,
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
