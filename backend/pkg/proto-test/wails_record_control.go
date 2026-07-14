package prototest

import (
	"fmt"
	"strconv"
	"strings"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// RecordControlService 录制控制服务（对应前端 packet-tab.vue 录制按钮）
//
// 时序图：
// ┌─────────────────┐    StartRecord(filterMode) ┌────────────────────┐
// │ packet-tab.vue  │ ───────────────────> │ RecordControlService │
// │   (前端)        │                     │   (后端)             │
// └─────────────────┘                     └────────────────────┘
//
//		     │                                        │
//		     │                                        │ Start()
//		     │                                        ▼
//		     │                                  ┌──────────┐
//		     │                                  │RecordWorker│
//		     │                                  │  + TCP代理 │
//		     │                                  │  + HTTP代理│
//		     │                                  └──────────┘
//		     │                                        │
//	  ......................................................
//	  .      │          filterMode:客户端修改报文        │    .
//	  .      │ <──────────────────────────────────────┤    .
//	  .      │                                        │    .
//	  .      │                                        │    .
//	  .      │          filterMode:客户端发送修改后的报文 │    .
//	  .      │ -─────────────────────────────────────>│    .
//	  ......................................................
//		     │                                        │
//		     │          Event.Emit('record:progress')  │
//		     │ <──────────────────────────────────────┤
//		     │                                        │
//		     ▼                                        ▼
//		表格实时追加消息                          监听端口并录制
type RecordControlService struct {
	recordWorker *RecordWorker
	connPool     *protocol.AccountConnectionPool
}

// NewRecordControlService 创建录制控制服务实例
func NewRecordControlService(recordWorker *RecordWorker, connPool *protocol.AccountConnectionPool) *RecordControlService {
	recordWorker.SetConnPool(connPool)
	return &RecordControlService{
		recordWorker: recordWorker,
		connPool:     connPool,
	}
}

// Start 启动监听并立即开始录制（兼容旧接口）
// 新代码建议分别调用 StartListen + StartRecord
func (s *RecordControlService) Start(serverAddr string, httpAddr string, filterMode bool) error {
	colonIdx := strings.LastIndex(serverAddr, ":")
	if colonIdx < 0 {
		return fmt.Errorf("服务器地址格式错误: %s", serverAddr)
	}
	tcpPort, err := strconv.Atoi(serverAddr[colonIdx+1:])
	if err != nil {
		return fmt.Errorf("服务器地址端口格式错误: %s", serverAddr)
	}
	if err := s.StartListen(serverAddr, httpAddr, tcpPort, 20144); err != nil {
		return err
	}
	return s.StartRecord(filterMode)
}

// StartListen 启动本地 TCP/HTTP 监听（应用启动时调用）
// serverAddr: 目标 TCP 服务器地址（如 "10.254.114.204:18000"）
// httpAddr: 目标 HTTP 地址（如 "10.254.114.204:20144"）
// tcpListenPort: 本地 TCP 监听端口
// httpListenPort: 本地 HTTP 监听端口
func (s *RecordControlService) StartListen(serverAddr string, httpAddr string, tcpListenPort int, httpListenPort int) error {
	return s.recordWorker.StartListen(serverAddr, httpAddr, tcpListenPort, httpListenPort)
}

// StartRecord 开始录制（监听已启动后调用）
func (s *RecordControlService) StartRecord(filterMode bool) error {
	return s.recordWorker.StartRecord(filterMode)
}

// StopRecord 停止录制（保留监听）
func (s *RecordControlService) StopRecord() error {
	s.recordWorker.StopRecord()
	return nil
}

// StopListen 停止监听（关闭 TCP/HTTP 监听器）
func (s *RecordControlService) StopListen() error {
	s.recordWorker.StopListen()
	return nil
}

// Stop 停止录制和监听（完整停止，兼容旧逻辑）
func (s *RecordControlService) Stop() error {
	s.recordWorker.Stop()
	return nil
}

// GetRecording 获取当前内存中的录制数据（供"保存为用例"使用）
func (s *RecordControlService) GetRecording() *protocol.Recording {
	return s.recordWorker.GetRecording()
}

// GetRecordStatus 获取当前录制状态
func (s *RecordControlService) GetRecordStatus() *RecordProgress {
	p := s.recordWorker.GetProgress()
	return &p
}

// ReleasePendingMessages 放行指定连接的 pending 消息（代理内写入 serverConn）
func (s *RecordControlService) ReleasePendingMessages(connID uint64, editsJSON string) error {
	return s.recordWorker.ReleasePendingMessages(connID, editsJSON)
}

// ReleaseAllPending 放行所有连接的 pending 消息
func (s *RecordControlService) ReleaseAllPending(editsJSON string) error {
	return s.recordWorker.ReleaseAllPending(editsJSON)
}

// SetFilterMode 关闭/开启实时拦截（关闭后后续消息透传）
func (s *RecordControlService) SetFilterMode(enabled bool) error {
	s.recordWorker.SetFilterMode(enabled)
	return nil
}
