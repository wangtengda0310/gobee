package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	intern "github.com/wangtengda/gobee/lvan/exporter/internal"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type TaskStatus struct {
	Status   string `json:"status"`    // running, completed, failed
	ExitCode int    `json:"exit_code"` // 命令退出码
}

var (
	completed = TaskStatus{Status: "completed", ExitCode: 0}
	failed    = TaskStatus{Status: "failed", ExitCode: 1}
	running   = TaskStatus{Status: "running", ExitCode: 2}
	blocking  = TaskStatus{Status: "blocking", ExitCode: 3}
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
