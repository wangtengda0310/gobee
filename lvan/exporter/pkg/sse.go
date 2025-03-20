package pkg

import (
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
	"sync"
	"time"
)

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
