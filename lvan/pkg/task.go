package pkg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	intern "github.com/wangtengda0310/gobee/lvan/internal"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
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

// 执行命令
func ExecuteTask(task *Task) {
	// 记录开始执行
	cmdName := task.Request.Cmd
	cmdVersion := task.Request.Version
	cmdArgs := task.Request.Args
	task.AddOutput(fmt.Sprintf("Starting command: %s\n", cmdName))
	task.AddOutput(fmt.Sprintf("Version: %s\n", cmdVersion))
	task.AddOutput(fmt.Sprintf("Arguments: %s\n", strings.Join(cmdArgs, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 版本: %s, 参数: %s", cmdName, cmdVersion, strings.Join(cmdArgs, ", "))

	// 使用版本管理获取可执行文件路径
	versionDir, err := GetCommandVersionPath(cmdName, cmdVersion)
	if err != nil {
		logger.Error("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		task.AddOutput(fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err))
		task.Complete(Failed, cmdNotExit)
		return
	}
	task.CmdMeta = intern.TryMeta(filepath.Join(versionDir, "meta.yaml"))

	// 查找可执行文件
	executable, err := FindExecutable(versionDir, cmdName)
	if err != nil {
		logger.Error("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		task.AddOutput(fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err))
		task.Complete(Failed, cmdNotExit)
		return
	}

	task.CmdPath, err = filepath.Abs(executable)
	if err != nil {
		logger.Error("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		task.AddOutput(fmt.Sprintf("找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err))
		task.Complete(Failed, cmdNotExit)
		return
	}

	// 记录使用的可执行文件路径
	task.AddOutput(fmt.Sprintf("使用可执行文件: %s\n", task.CmdPath))

	var timeout = 10 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	task.Cancel = cancel

	var cmd *exec.Cmd
	if task.CmdMeta != nil && task.CmdMeta.Shell != nil {
		newArgs := append(task.CmdMeta.Shell[1:], task.CmdPath)
		newArgs = append(newArgs, cmdArgs...)
		cmd = exec.CommandContext(ctx, task.CmdMeta.Shell[0], newArgs...)

	} else if runtime.GOOS == "windows" {
		// 检查是否是 Windows 平台 尝试使用cmd执行.bat和.cmd
		// 检查文件扩展名是否为批处理文件
		ext := strings.ToLower(filepath.Ext(task.CmdPath))
		if ext == ".bat" || ext == ".cmd" {
			// 使用 cmd /c 执行批处理文件
			newArgs := append([]string{"/c", task.CmdPath}, cmdArgs...)
			cmd = exec.CommandContext(ctx, "cmd", newArgs...)
		} else {
			// 非批处理文件直接执行
			cmd = exec.CommandContext(ctx, task.CmdPath, cmdArgs...)
		}
	} else {
		// 非 Windows 平台直接执行命令
		cmd = exec.CommandContext(ctx, task.CmdPath, cmdArgs...)
	}

	// 获取当前环境变量
	var env = os.Environ()
	if task.CmdMeta != nil && len(task.CmdMeta.Resources) > 0 {
		retries := 40
		logger.Info("默认重试次数为 %d 可以通过环境变量 exporter_retry_times 设置", retries)
		if os.Getenv("exporter_retry_times") != "" {
			retry, err := strconv.Atoi(os.Getenv("exporter_retry_times"))
			if err == nil {
				logger.Info("使用环境变量 exporter_retry_times 设置重试次数为 %d", retry)
				retries = retry
			}
		}
		a.Add(1)
		defer a.Add(-1)
		resource, err, lock := intern.ExclusiveOneResource(task.CmdMeta.Resources, TasksDir, retries)
		if err != nil {
			// 无法获取资源，记录错误并继续执行
			logger.Warn("无法获取资源锁: %v，任务将继续执行但可能影响性能", err)
			message := fmt.Sprintf("排队超时 当前排队人数 %d\n", a.Load())
			task.Result.Stderr = append(task.Result.Stderr, message)
			task.AddOutput(message)
			task.Complete(Failed, exclusive)
			return
		}
		defer func(resource string) {
			if lock != nil {
				err := lock.Unlock()
				if err != nil {
					logger.Error("解锁失败: %v", err)
				}
			}

			// 释放资源锁
			if resource != "" {
				if err := intern.ReleaseResource(resource, lock); err != nil {
					logger.Error("释放资源锁失败: %s, %v", resource, err)
				} else {
					logger.Info("成功释放资源锁: %s", resource)
				}
			}
		}(resource)

		// 添加自定义环境变量
		env = append(env, fmt.Sprintf("exporter_cmd_%s_resource=%s", cmdName, resource))
	}

	env = append(env, fmt.Sprintf("exporter_cmd_%s_id=%s", cmdName, task.ID))

	// 设置环境变量
	if len(task.Request.Env) > 0 {

		// 添加自定义环境变量
		for key, value := range task.Request.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
			logger.Debug("设置环境变量: %s=%s", key, value)
		}
	}

	var encodingf func([]byte) string
	var encoding intern.Charset
	if task.CmdMeta != nil && task.CmdMeta.Encoding != "" {
		encoding = task.CmdMeta.Encoding
		encodingf = func(s []byte) string {
			return UtfFrom(s, encoding)
		}
	}
	status, err, stdout, stderr := Cmd(cmd, task.WorkDir, env)
	task.Status = status
	if err != nil {
		return
	}

	CatchStdout(stdout, encodingf, task.AddOutput)

	log := func(s string) {
		task.AddOutput(s)
		task.Result.Stderr = append(task.Result.Stderr, s)
	}
	CatchStderr(stderr, encodingf, log)

	logger.Info("等待命令完成")
	err = cmd.Wait()
	if err != nil {
		var exitCode int
		// 处理超时错误
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			task.AddOutput("命令执行超时")
			exitCode = 124 // 通常用 124 表示超时退出码
			task.AddOutput(fmt.Sprintf("命令执行超时，退出码 %d: %s", exitCode, err.Error()))
		} else if exitErr, ok := err.(*exec.ExitError); ok { // 尝试获取退出码
			exitCode = exitErr.ExitCode()
			task.AddOutput(fmt.Sprintf("命令执行失败，退出码 %d: %s", exitCode, err.Error()))
		} else {
			exitCode = 1 // 默认错误码
			task.AddOutput(fmt.Sprintf("命令执行未知错误，退出码 %d: %s", exitCode, err.Error()))
		}
		task.Complete(Failed, exitCode)
	} else {
		task.Complete(Completed, success)
	}

}
