package prototest

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
)

// ReplayControlService 重放控制服务（对应前端 packet-tab.vue 和 testcase-tab.vue 重放按钮）
//
// 时序图：
// ┌──────────────────┐  SendMessages    ┌─────────────────────┐
// │ packet-tab.vue   │ ────────────────> │ ReplayControlService │
// │ testcase-tab.vue │                  │   (后端)              │
// └──────────────────┘                  └─────────────────────┘
//
//	     │                                        │
//	     │                                        │ StartSend()
//	     │                                        ▼
//	     │                                  ┌──────────┐
//	     │                                  │ReplayWorker│
//	     │                                  │  HTTP登录 │
//	     │                                  │  TCP连接  │
//	     │                                  │  发送LoginReq
//	     │                                  │  逐条发送Req
//	     │                                  └──────────┘
//	     │                                        │
//	     │    Event.Emit('replay:progress')        │
//	     │ <────────────────────────────────────────┘
//	     │                                        │
//	     ▼                                        ▼
//	更新进度标签                             执行重放逻辑
//
//	     │    Event.Emit('record:progress', latest_msg)
//	     │ <────────────────────────────────────────┘
//	     │
//	     ▼
//	表格追加重放消息
type ReplayControlService struct {
	worker       *ReplayWorker
	connPool     *protocol.AccountConnectionPool
	recordWorker *RecordWorker // 录制工作器引用（检测代理连接冲突）
}

// NewReplayControlService 创建重放控制服务实例
func NewReplayControlService(worker *ReplayWorker, connPool *protocol.AccountConnectionPool, recordWorker *RecordWorker) *ReplayControlService {
	return &ReplayControlService{
		worker:       worker,
		connPool:     connPool,
		recordWorker: recordWorker,
	}
}

// SendMessages 启动异步消息发送（前端"开始重放"/"执行用例"/"重发"共用）
// 检测目标账号是否与录制代理中的活跃连接冲突：
//   - 不冲突 → 走连接池路径
//   - 冲突 → 走代理注入路径（使用 proxySeqId，不新建连接）
//
// 特殊限制：有变量依赖的消息不支持注入路径（因为注入路径没有 FrameMux 能力）
func (s *ReplayControlService) SendMessages(serverAddr, httpAddr, openID string, messagesJSON string, repeatCount int, rangeStart, rangeEnd int) error {
	// 冲突检测：如果录制代理中有活跃连接且目标账号匹配，走注入路径
	if s.recordWorker != nil && s.recordWorker.IsRecording() {
		for i := rangeStart; i <= rangeEnd; i++ {
			accountID := fmt.Sprintf("%s%d", openID, i)
			if s.recordWorker.HasAccountConnection(accountID) {
				// D5: 注入路径不支持变量（没有 FrameMux）
				if hasVariableInMessages(messagesJSON) {
					return fmt.Errorf("消息包含变量占位符，不支持通过录制代理注入路径发送，请先停止录制后再重放")
				}
				log.Printf("[ReplayControl] 账号 %s 在录制代理中活跃，切换到注入路径", accountID)
				return s.worker.StartInject(messagesJSON, repeatCount, s.recordWorker)
			}
		}
	}
	return s.worker.StartSend(serverAddr, httpAddr, openID, messagesJSON, repeatCount, rangeStart, rangeEnd)
}

// hasVariableInMessages 检测 messagesJSON 中是否包含变量占位符 {"__var__":"..."}
func hasVariableInMessages(messagesJSON string) bool {
	// 简单检查 JSON 字符串中是否包含 __var__ 标记
	return strings.Contains(messagesJSON, `"__var__"`)
}

// StartReplay 启动异步协议重放（从录制文件重放）
func (s *ReplayControlService) StartReplay(filePath string, serverAddr string, httpAddr string, openID string, repeatCount int) error {
	return s.worker.Start(filePath, serverAddr, httpAddr, openID, repeatCount)
}

// StopReplay 停止正在进行的协议重放
func (s *ReplayControlService) StopReplay() error {
	s.worker.Stop()
	return nil
}

// GetReplayStatus 获取当前重放状态
func (s *ReplayControlService) GetReplayStatus() *ReplayProgress {
	p := s.worker.GetProgress()
	return &p
}

// GetReplaySettings 获取当前重放参数设置（供前端设置面板初始化）
func (s *ReplayControlService) GetReplaySettings() map[string]int {
	return map[string]int{
		"send_interval_ms": protocol.SendIntervalMs,
		"ack_wait_ms":      protocol.AckWaitMs,
		"max_concurrency":  protocol.MaxConcurrency,
	}
}

// SetReplaySettings 设置重放参数（供前端设置面板修改）
func (s *ReplayControlService) SetReplaySettings(sendIntervalMs int, ackWaitMs int, maxConcurrency int) {
	protocol.SendIntervalMs = sendIntervalMs
	protocol.AckWaitMs = ackWaitMs
	protocol.MaxConcurrency = maxConcurrency
	log.Printf("[ReplayControl] 设置已更新: SendIntervalMs=%d, AckWaitMs=%d, MaxConcurrency=%d", sendIntervalMs, ackWaitMs, maxConcurrency)
}

// SendIterativeMessages 迭代发送消息
// 前端传入一条基础消息的 payload + 字段迭代配置（FieldValues），
// 后端根据配置生成多条消息后一次性批量发送。
//
// 流程：
//  1. 解析 payloadJSON 为 map[string]any
//  2. 将 FieldValues 转为内部 IterationConfig
//  3. 调用 params.GenerateIterativeMessages 生成分组独立迭代消息
//  4. 每条消息序列化回 JSON，组装为 []RecordMessage
//  5. 批量序列化 → worker.StartSend()（复用进度推送 + 取消机制）
//
// @frontend
func (s *ReplayControlService) SendIterativeMessages(
	serverAddr, httpAddr, openID string,
	msgID uint16, msgName string,
	payloadJSON string, direction string,
	fieldValues map[string]FieldValues,
) error {
	// 1. 解析基础 payload
	var basePayload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &basePayload); err != nil {
		return fmt.Errorf("解析 PayloadJSON 失败: %v", err)
	}

	// 2. FieldValues → IterationConfig
	// 无迭代配置时 params.GenerateIterativeMessages 会返回原始消息（单条发送）
	configs := params.FieldValuesToIterationConfig(fieldValues)

	// 3. 生成迭代消息
	payloads := params.GenerateIterativeMessages(basePayload, configs)

	// 安全检查：防止笛卡尔积爆炸
	const maxIterativeMessages = 10000
	if len(payloads) > maxIterativeMessages {
		return fmt.Errorf("迭代消息数量 %d 超过上限 %d，请减少迭代字段或范围", len(payloads), maxIterativeMessages)
	}

	// 4. 组装为 []RecordMessage
	messages := make([]protocol.RecordMessage, len(payloads))
	for i, p := range payloads {
		jsonBytes, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("序列化第 %d 条迭代消息失败: %v", i, err)
		}
		messages[i] = protocol.RecordMessage{
			MsgID:       msgID,
			MsgName:     msgName,
			PayloadJSON: string(jsonBytes),
			Direction:   direction,
		}
	}

	// 5. 冲突检测：迭代发送始终只针对第一个账号（rangeStart=1, rangeEnd=1）
	// 冲突时走注入路径（使用 proxySeqId）
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("序列化消息列表失败: %v", err)
	}
	if s.recordWorker != nil && s.recordWorker.IsRecording() {
		accountID := fmt.Sprintf("%s%d", openID, 1)
		if s.recordWorker.HasAccountConnection(accountID) {
			log.Printf("[ReplayControl] 账号 %s 在录制代理中活跃，迭代发送切换到注入路径", accountID)
			return s.worker.StartInject(string(messagesJSON), 1, s.recordWorker)
		}
	}
	return s.worker.StartSend(serverAddr, httpAddr, openID, string(messagesJSON), 1, 1, 1)
}

// GetConnectionPoolStatus 获取连接池状态
func (s *ReplayControlService) GetConnectionPoolStatus() []protocol.ConnPoolEntry {
	if s.connPool == nil {
		return []protocol.ConnPoolEntry{}
	}
	return s.connPool.List()
}

// CloseConnection 关闭指定账号的连接
func (s *ReplayControlService) CloseConnection(accountID string) error {
	if s.connPool == nil {
		return fmt.Errorf("连接池未初始化")
	}
	s.connPool.Close(accountID)
	return nil
}

// CloseAllConnections 关闭所有连接
func (s *ReplayControlService) CloseAllConnections() error {
	if s.connPool == nil {
		return fmt.Errorf("连接池未初始化")
	}
	s.connPool.CloseAll()
	return nil
}

// GetAvailableVariables 返回注册表中所有可用变量列表（供前端变量选择器使用）（E1/E2）
func (s *ReplayControlService) GetAvailableVariables() []params.VariableInfo {
	return params.GetAvailableVariables()
}

// SendMessagesWithFieldValues 统一发送消息（支持迭代/变量配置）
// 接受包含 field_values 的消息列表，内部展开迭代配置后统一一次 StartSend 发送
// 变量字段的 FieldValues 元数据会附带在 RecordMessage 中传递到 replay 层
//
// @frontend
func (s *ReplayControlService) SendMessagesWithFieldValues(
	serverAddr, httpAddr, openID string,
	entries []RecordEntryView,
	repeatCount, rangeStart, rangeEnd int,
) error {
	// 展开所有消息（包括带迭代/变量配置的）
	var allMessages []protocol.RecordMessage
	for _, entry := range entries {
		if len(entry.FieldValues) > 0 {
			// 有迭代配置：展开为多条消息
			configs := params.FieldValuesToIterationConfig(entry.FieldValues)
			payloads := params.GenerateIterativeMessages(entry.Payload, configs)

			// 收集 variable 类型字段的元数据
			fieldMeta := buildFieldMetaFromFieldValues(entry.FieldValues)

			for _, p := range payloads {
				jsonBytes, err := json.Marshal(p)
				if err != nil {
					return fmt.Errorf("序列化迭代消息失败: %v", err)
				}
				allMessages = append(allMessages, protocol.RecordMessage{
					MsgID:       entry.MsgID,
					MsgName:     entry.MsgName,
					PayloadJSON: string(jsonBytes),
					Direction:   entry.Direction,
					FieldValues: fieldMeta,
				})
			}
		} else {
			// 普通消息
			payloadJSON, err := json.Marshal(entry.Payload)
			if err != nil {
				return fmt.Errorf("序列化消息失败: %v", err)
			}
			allMessages = append(allMessages, protocol.RecordMessage{
				MsgID:       entry.MsgID,
				MsgName:     entry.MsgName,
				PayloadJSON: string(payloadJSON),
				Direction:   entry.Direction,
			})
		}
	}

	if len(allMessages) == 0 {
		return nil
	}

	messagesJSON, err := json.Marshal(allMessages)
	if err != nil {
		return fmt.Errorf("序列化消息列表失败: %v", err)
	}

	// 冲突检测：录制中且目标账号在代理中有活跃连接时，优先走注入路径复用连接。
	// 但变量解析需要 FrameMux，注入路径不支持，因此含变量配置的消息降级为直接发送。
	if s.recordWorker != nil && s.recordWorker.IsRecording() {
		hasVar := false
		for _, m := range allMessages {
			if len(m.FieldValues) > 0 {
				hasVar = true
				break
			}
		}
		if !hasVar {
			for i := rangeStart; i <= rangeEnd; i++ {
				accountID := fmt.Sprintf("%s%d", openID, i)
				if s.recordWorker.HasAccountConnection(accountID) {
					return s.worker.StartInject(string(messagesJSON), repeatCount, s.recordWorker)
				}
			}
		}
	}

	return s.worker.StartSend(serverAddr, httpAddr, openID, string(messagesJSON), repeatCount, rangeStart, rangeEnd)
}

// buildFieldMetaFromFieldValues 从前端 FieldValues 提取变量类型字段的元数据
// 用于在 RecordMessage 中传递到 replay 层
func buildFieldMetaFromFieldValues(fv map[string]FieldValues) map[string]protocol.FieldMetaValue {
	result := make(map[string]protocol.FieldMetaValue)
	for field, v := range fv {
		if v.InputType == "variable" && v.VariableName != "" {
			result[field] = protocol.FieldMetaValue{
				InputType:    v.InputType,
				VariableName: v.VariableName,
			}
		}
	}
	return result
}
