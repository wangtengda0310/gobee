package execute

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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
	WorkDir    string                `json:"workdir"`
	Mutex      *sync.Mutex           `json:"-"`
	Logger     *logger.Logger        `json:"-"`
	sseClients *ClientManager
	CmdMeta    *CommandMeta       `json:"-"` // 可能为空
	Cancel     context.CancelFunc `json:"-"`
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
func (t *Task) Completef(status TaskStatus, exitCode int, format string, args ...any) {
	t.AddOutput(fmt.Sprintf(format, args...))
	t.Complete(status, exitCode)

}
func (t *Task) Complete(status TaskStatus, exitCode int) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	t.Status = status
	t.Result.ExitCode = exitCode
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

func findExec(ctx context.Context, versionDir, cmdName string, cmdMeta *CommandMeta, cmdArgs ...string) (*exec.Cmd, error) {
	// 查找可执行文件
	executable, err := FindExecutable(versionDir, cmdName)
	if err != nil {
		return nil, err
	}

	cmdpath, err := filepath.Abs(executable)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if cmdMeta != nil && cmdMeta.Shell != nil {
		newArgs := append(cmdMeta.Shell[1:], cmdpath)
		newArgs = append(newArgs, cmdArgs...)
		cmd = exec.CommandContext(ctx, cmdMeta.Shell[0], newArgs...)

	} else if runtime.GOOS == "windows" {
		// 检查是否是 Windows 平台 尝试使用cmd执行.bat和.cmd
		// 检查文件扩展名是否为批处理文件
		ext := strings.ToLower(filepath.Ext(cmdpath))
		if ext == ".bat" || ext == ".cmd" {
			// 使用 cmd /c 执行批处理文件
			newArgs := append([]string{"/c", cmdpath}, cmdArgs...)
			cmd = exec.CommandContext(ctx, "cmd", newArgs...)
		} else {
			// 非批处理文件直接执行
			cmd = exec.CommandContext(ctx, cmdpath, cmdArgs...)
		}
	} else {
		// 非 Windows 平台直接执行命令
		cmd = exec.CommandContext(ctx, cmdpath, cmdArgs...)
	}

	return cmd, nil
}

// ExecuteTask 执行命令
func ExecuteTask(task *Task) {
	// 记录开始执行
	cmdName := task.Request.Cmd
	cmdVersion := task.Request.Version
	cmdArgs := task.Request.Args
	taskoutput := task.AddOutput
	taskoutput(fmt.Sprintf("Starting command: %s\n", cmdName))
	taskoutput(fmt.Sprintf("Version: %s\n", cmdVersion))
	taskoutput(fmt.Sprintf("Arguments: %s\n", strings.Join(cmdArgs, ", ")))

	// 记录日志
	logger.Info("执行命令: %s, 版本: %s, 参数: %s", cmdName, cmdVersion, strings.Join(cmdArgs, ", "))

	var execute = executing{env: os.Environ(), task: task}

	// 使用版本管理获取可执行文件路径
	versionDir, err := GetCommandVersionPath(cmdName, cmdVersion)
	if err != nil {
		commands, e := ListCommands()
		if e == nil {
			err = fmt.Errorf("%w: %s, 可用命令有: %v", err, cmdName, commands)
		}
		task.Completef(Failed, cmdNotExit, "找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		return
	}
	execute.versionDir = versionDir

	task.CmdMeta = TryMeta(filepath.Join(versionDir, "meta.yaml"))

	ctx, cancel := execute.taskTimeout(task.CmdMeta)
	task.Cancel = cancel

	cmd, err := findExec(ctx, execute.versionDir, cmdName, task.CmdMeta, task.Request.Args...)
	if err != nil {
		task.Completef(Failed, cmdNotExit, "找不到命令 %s 版本 %s: %v\n", cmdName, cmdVersion, err)
		return
	}
	// 记录使用的可执行文件路径
	execute.outputTask(fmt.Sprintf("使用可执行文件: %s\n", cmd))

	var resources []string
	if task.CmdMeta != nil {
		resources = task.CmdMeta.Resources
		execute.encoding = task.CmdMeta.Encoding
	}
	defer execute.excludeResource(task, resources)()

	execute.appendEnv(cmdName, task)

	err, stdout, stderr := Cmd(cmd, task.WorkDir, execute.env)
	if err != nil {
		task.Status = Failed
		return
	}
	task.Status = Running

	execute.CatchStdout(stdout, stderr)

	logger.Info("等待命令完成")
	err = cmd.Wait()
	switch {
	case err != nil:
		var exitCode int
		if errors.Is(context.Cause(ctx), context.DeadlineExceeded) || errors.Is(context.Cause(ctx), context.Canceled) {
			task.Completef(Failed, timeout, "exporter 命令执行超时，退出码 %d: %s", exitCode, err.Error())
		} else if exitErr := new(exec.ExitError); errors.As(err, &exitErr) { // 尝试获取退出码
			exitCode = exitErr.ExitCode()
			task.Completef(Failed, exitCode, "exporter 命令执行失败，退出码 %d: %s", exitCode, err.Error())
		} else {
			exitCode = 1 // 默认错误码
			task.Completef(Failed, exitCode, "exporter 命令执行未知错误，退出码 %d: %s", exitCode, err.Error())
		}
		(&errHandler{executed: &execute}).errorCallback(task)
	default:
		task.Complete(Completed, success)
	}

}

type errHandler struct {
	executed *executing
}

func (e *errHandler) errorCallback(task *Task) {
	var execute = e.executed
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd, err := findExec(ctx, execute.versionDir, fmt.Sprintf("error_%d", task.Result.ExitCode), nil)
	if err == nil {
		logger.Info("执行错误回调脚本%s", cmd)

		err, stdout, stderr := Cmd(cmd, task.WorkDir, execute.env)

		if err != nil {
			return
		}

		execute.CatchStdout(stdout, stderr)

		err = cmd.Wait()
		if err != nil {
			logger.Warn("%v", err)
		}

	}
}
