package execute

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

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
	timeout    = 124
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
func CreateTask(req internal.CommandRequest, w ...io.Writer) *Task {
	return taskManager.CreateTask(req, w...)
}

func (tm *TaskManager) CreateTask(req internal.CommandRequest, w ...io.Writer) *Task {
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
	internal.Create(filePath)

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
