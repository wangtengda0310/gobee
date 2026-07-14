package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/params"
)

// ReplayProgressCallback 重放进度回调
// total: 总消息数, sent: 已发送数, currentMsg: 当前消息名
// 返回 true 表示继续，返回 false 表示取消重放
type ReplayProgressCallback func(total, sent int, currentMsg string) bool

// ReplayMessageCallback 每条消息的推送回调（用于前端实时展示）
// direction: "→" 客户端→服务端, "←" 服务端→客户端
// accountID: 当前账号标识（如 test1、test2），录制时为空
type ReplayMessageCallback func(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string)

// SendIntervalMs 重放时每条消息的发送间隔（毫秒），默认 1 秒
var SendIntervalMs = 1000

// AckWaitMs 重放完成后等待服务器返回 Ack 的时间（毫秒）
var AckWaitMs = 2000

// RecordMessage 待发送的消息（前端传递的轻量结构）
type RecordMessage struct {
	MsgID       uint16                    `json:"msg_id"`
	MsgName     string                    `json:"msg_name"`
	PayloadJSON string                    `json:"payload_json"`
	OffsetMs    int                       `json:"offset_ms"`
	Direction   string                    `json:"direction"`
	SeqID       uint32                    `json:"seq_id"`
	FieldValues map[string]FieldMetaValue `json:"field_values,omitempty"` // 变量/迭代字段元数据
}

// FieldMetaValue 记录单个字段的元信息（跟随 RecordMessage 在重放管线中传递）
type FieldMetaValue struct {
	InputType    string `json:"input_type"`              // "original" | "range" | "enum" | "combo" | "variable"
	VariableName string `json:"variable_name,omitempty"` // 仅 variable 类型，如 "cityId"
}

// SendMessages 并发发送指定消息列表到服务器
// 每个账号独立完成 HTTP 登录 + TCP 连接 + 消息发送，发完后关闭连接
// 通过 MaxConcurrency 控制并发度，0=不限制
// 错误模式: 跳过失败账号继续执行其余账号，最终汇总失败列表
//
// 参数:
//   - serverAddr: TCP 游戏服务器地址 (如 10.254.114.204:18000)
//   - httpAddr: HTTP 登录服务地址 (如 10.254.114.204:20144)
//   - openID: 账号前缀，与 rangeStart/rangeEnd 拼接构成完整账号
//     例: openID="test", rangeStart=1, rangeEnd=3 → test1, test2, test3
//     与 AuthLogin 中 FakeSDK 的 open_id 相同，也与 LoginReq.Account 相同
//   - messages: 要发送的消息列表（←方向跳过只发→方向）
//   - repeatCount: 每个账号重复发送轮数
//   - rangeStart/rangeEnd: 账号序号范围，1-based
//   - ctx: 取消上下文，nil 表示不可取消
//   - onProgress: 进度回调（返回 false 取消重放）
//   - onMessage: 每条消息推送回调（前端实时展示）
//   - connPool: 连接池（可选）。非 nil 时优先从池中借出已登录连接，用完后归还。
//
// accountResult 单个账号的并发执行结果
type accountResult struct {
	accountID string
	sent      int
	err       error
}

// SendMessages 并发发送指定消息列表到服务器。
// 详细行为见 replayer.go 中的 Replayer。
func SendMessages(serverAddr, httpAddr, openID string, messages []RecordMessage, repeatCount int, rangeStart, rangeEnd int, ctx context.Context, onProgress ReplayProgressCallback, onMessage ReplayMessageCallback, connPool *AccountConnectionPool) error {
	return NewReplayer(ReplayOptions{
		ServerAddr:  serverAddr,
		HTTPAddr:    httpAddr,
		OpenID:      openID,
		Messages:    messages,
		RepeatCount: repeatCount,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
		Context:     ctx,
		OnProgress:  onProgress,
		OnMessage:   onMessage,
		ConnPool:    connPool,
		RetryConfig: GlobalRetryConfig,
		Concurrency: MaxConcurrency,
	}).SendMessages()
}

// replaySession 重放会话状态（连接 + 变量上下文 + 清理）
// 生命周期：acquireConnection → prepareVariableContext → 发送循环 → cleanup
//
// 并发安全：
//   - readerDone: readDrainer 退出信号，cleanup() 通过它同步等待 readDrainer 完全退出
//     防止 readDrainer 与下一个 borrower 并发读取同一 net.Conn
//   - borrowedFromPool: 标记连接是否来自连接池，决定 cleanup 时归还还是关闭
//   - lastSeqID: 本次发送使用的最大 seqId，归还连接池时用于续接
type replaySession struct {
	auth             Authenticator // 负责获取与归还/关闭连接
	conn             net.Conn
	borrowedFromPool bool
	accountID        string
	lastSeqID        uint32 // 服务端已知该连接的最大 seqId（连接池续接用）

	// 变量上下文（有变量依赖时使用 FrameMux，否则用 readDrainer）
	mux           *FrameMux
	stopReader    chan struct{}
	readerDone    chan struct{} // readDrainer 退出信号，cleanup 同步等待用
	variableStore map[string]any
}

// cleanup 统一清理会话资源
// 阶段：
//  1. 停止 FrameMux（如有）或同步等待 readDrainer 退出
//     - FrameMux: 调用 Stop() 内部已 wg.Wait() 同步等待 readLoop 退出
//     - readDrainer: close(stopReader) 通知退出 → SetReadDeadline(now) 强制中断 io.ReadFull
//     → 等待 readerDone 同步完成（500ms 超时）
//  2. 通过 Authenticator 归还连接池或关闭连接
//
// 关键：必须在 readDrainer 完全退出后再归还连接，否则下一个 borrower 会与之并发读取同一 net.Conn。
func (s *replaySession) cleanup() {
	if s.mux != nil {
		s.mux.Stop()
		// mux.Stop() 内部已 wg.Wait() 同步等待 readLoop 完全退出，无需额外 sleep
	} else if s.stopReader != nil {
		close(s.stopReader)
		// 强制中断 io.ReadFull：设置极短 deadline 让 ReadFull 立即返回错误
		if s.conn != nil {
			_ = s.conn.SetReadDeadline(time.Now())
		}
		// 同步等待 readDrainer 完全退出，避免与下一个 borrower 并发读取同一连接
		if s.readerDone != nil {
			select {
			case <-s.readerDone:
				log.Printf("[重放] [account=%s] readDrainer 已同步退出", s.accountID)
			case <-time.After(500 * time.Millisecond):
				log.Printf("[重放] [account=%s] readDrainer 同步退出超时(500ms)，继续归还连接", s.accountID)
			}
		}
	}

	if s.auth != nil {
		s.auth.Return(s.accountID, s.conn, s.lastSeqID)
	} else if s.conn != nil {
		_ = s.conn.Close()
	}
}

// prepareVariableContext 根据消息的 FieldValues 元数据准备变量解析上下文
//
// 惰性提取设计 (2026-06-15 修复时序 bug):
// 此函数只负责启动 FrameMux 和创建空的 variableStore，不再立即调用 ExtractVariableValues。
// 变量提取推迟到发送循环中，发送依赖变量的消息前才按需触发 (ExtractVariablesForMessage)。
// 这样确保 Ntf 帧在前面消息发出后、变量消息发出前的窗口内被 readLoop 缓存，WaitMsg 能命中 cache。
//
// openid-only 场景 (2026-06-17 修复):
// 当用例只有账号级变量（如 openid）而无 Ntf 变量（watchedIDs 为空）时：
//   - 池连接需先 DrainConn 排空积压帧（防止旧帧干扰），再启动 readDrainer 持续消费
//   - 新连接无需 DrainConn，直接启动 readDrainer
//   - 这是 F1 修复：原实现 skipDrain=true 但 watchedIDs 为空，导致池连接积压帧未清理
//
// 有变量依赖时创建 variableStore（并预置账号级变量 openid）；
// 若同时存在 Ntf 变量（watchedIDs 非空），才创建 FrameMux 缓存帧；
// 否则仅启动 readDrainer 消费帧。
func prepareVariableContext(conn net.Conn, messages []RecordMessage, borrowedFromPool bool, onMessage ReplayMessageCallback, accountID string) (mux *FrameMux, stopReader chan struct{}, readerDone chan struct{}, variableStore map[string]any) {
	hasVariable, watchedIDs := ScanFieldValuesForVariables(messages)

	if !hasVariable {
		stopReader = make(chan struct{})
		readerDone = make(chan struct{})
		go readDrainer(conn, stopReader, readerDone, onMessage, accountID)
		return
	}

	// 创建变量存储，预置账号级变量 openid（每个账号独立）
	variableStore = make(map[string]any)
	variableStore[params.OpenIDShortName] = accountID

	if len(watchedIDs) > 0 {
		mux = NewFrameMux(conn, watchedIDs)
		if borrowedFromPool {
			mux.DrainAndStart(100*time.Millisecond, onMessage, accountID)
		} else {
			mux.wg.Add(1)
			go mux.readLoop(onMessage, accountID)
		}
		log.Printf("[重放] [account=%s] FrameMux 已启动(惰性提取模式), watchedIDs=%v", accountID, watchedIDs)
	} else {
		// 只有账号级变量（如 openid），无需缓存 Ntf 帧
		// 但池连接可能积压旧帧，需先按帧边界排空，再启动 readDrainer 持续消费
		if borrowedFromPool {
			if drainErr := DrainConn(conn, 100*time.Millisecond); drainErr != nil {
				log.Printf("[重放] [account=%s] openid-only 场景 DrainConn 失败: %v，readDrainer 后续遇到帧错误会自行退出", accountID, drainErr)
			}
		}
		stopReader = make(chan struct{})
		readerDone = make(chan struct{})
		go readDrainer(conn, stopReader, readerDone, onMessage, accountID)
		log.Printf("[重放] [account=%s] 仅账号级变量，启动 readDrainer", accountID)
	}
	return
}

// resolveVariablePayload 根据字段元数据解析消息中的变量字段，返回可发送的 payloadJSON
func resolveVariablePayload(msg RecordMessage, variableStore map[string]any) string {
	if len(msg.FieldValues) == 0 || len(variableStore) == 0 {
		return msg.PayloadJSON
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(msg.PayloadJSON), &payload); err != nil {
		return msg.PayloadJSON
	}
	resolved, resolveErr := ResolveMessageVariables(payload, msg.FieldValues, variableStore)
	if resolveErr != nil {
		log.Printf("[变量解析] %v", resolveErr)
	}
	resolvedJSON, err := json.Marshal(resolved)
	if err != nil {
		return msg.PayloadJSON
	}
	return string(resolvedJSON)
}

// accountRunOptions 描述对单个账号执行一次完整发送流程所需参数。
// 嵌入 ReplayOptions 复用批量配置字段，仅保留单账号差异字段。
type accountRunOptions struct {
	ReplayOptions
	accountID   string        // 当前账号
	auth        Authenticator // 负责获取与归还/关闭连接
	grandTotal  int           // 全局总消息数（用于进度回调）
	alreadySent int           // 已发送消息数（用于进度回调）
}

// sendMessagesOnce 对单个账号执行一次完整的发送流程
//
// 阶段:
//  0. scanVariables         → 扫描变量依赖，决定 Authenticate 的 skipDrain
//  1. auth.Authenticate     → 通过 Authenticator 获取已准备好的 TCP 连接
//  2. prepareVariableContext → FrameMux/变量提取上下文
//  3. 发送循环               → 逐条发送 + 惰性变量提取/替换
//  4. cleanup                → 通过 Authenticator 归还或关闭连接
func sendMessagesOnce(opts accountRunOptions) (sent int, err error) {
	log.Printf("[重放] [account=%s] 消息数量: %d, 重复次数: %d", opts.accountID, len(opts.Messages), opts.RepeatCount)
	log.Printf("[重放] [account=%s] 目标服务器: %s", opts.accountID, opts.ServerAddr)

	// 阶段 0: 扫描变量依赖（用于提示 Authenticator 是否跳过 DrainConn）
	hasVariableCtx, _ := ScanFieldValuesForVariables(opts.Messages)

	// 阶段 1: 获取连接
	// Authenticator 内部可自主选择新建连接、复用连接池连接或其他方式；
	// sendMessagesOnce 不再直接操作连接池。
	conn, borrowedFromPool, startSeqID, connErr := opts.auth.Authenticate(opts.Context, opts.accountID, hasVariableCtx)
	if connErr != nil {
		return 0, connErr
	}

	// 阶段 2: 准备变量上下文
	sess := &replaySession{
		auth:             opts.auth,
		conn:             conn,
		borrowedFromPool: borrowedFromPool,
		accountID:        opts.accountID,
		lastSeqID:        startSeqID,
	}
	defer sess.cleanup()

	sess.mux, sess.stopReader, sess.readerDone, sess.variableStore = prepareVariableContext(
		conn, opts.Messages, borrowedFromPool, opts.OnMessage, opts.accountID,
	)

	// 有变量上下文即可能走变量解析（包括仅账号级变量 openid 的场景）
	hasVariableContext := sess.variableStore != nil

	// 阶段 3: 发送循环
	// seqID 从池连接续接（首次使用的新连接 startSeqID=0，从 1 开始）
	seqID := sess.lastSeqID
	roundTotal := 0
	skipped := 0
	failed := 0

	log.Printf("[重放] [account=%s] 开始发送循环: %d 条消息 × %d 轮 (SendIntervalMs=%d)",
		opts.accountID, len(opts.Messages), opts.RepeatCount, SendIntervalMs)

	for r := range opts.RepeatCount {
		roundStart := time.Now()
		log.Printf("[重放] [account=%s] ==== 第 %d/%d 轮 ====", opts.accountID, r+1, opts.RepeatCount)
		roundSent := 0
		roundSkipped := 0
		roundFailed := 0

		for i, msg := range opts.Messages {
			// 跳过 Ack/Ntf（只发 "→" 方向的）
			if msg.Direction != "" && msg.Direction != "→" && msg.Direction != DirClientToServer {
				log.Printf("[重放] [account=%s] [跳过] 第%d轮-第%d条 %s direction=%q", opts.accountID, r+1, i+1, msg.MsgName, msg.Direction)
				skipped++
				roundSkipped++
				continue
			}

			// 惰性变量提取（仅在该消息依赖变量且有 FrameMux 上下文时触发）
			// 提取失败(如 Ntf 超时)则跳过该消息发送并计为失败，而非静默用写死值兜底——
			// QA 工具最怕"测试了错误数据却以为成功"
			payloadToSend := msg.PayloadJSON
			if hasVariableContext && msgNeedsVariable(msg) {
				if extractErr := ExtractVariablesForMessage(msg, sess.mux, sess.variableStore); extractErr != nil {
					log.Printf("[重放] [account=%s] [变量提取失败-跳过] 第%d轮-第%d条 %s: %v", opts.accountID, r+1, i+1, msg.MsgName, extractErr)
					// 通过 onMessage 回传跳过原因，让前端可见（而非静默消失）
					if opts.OnMessage != nil {
						opts.OnMessage(msg.MsgName+"(跳过:变量提取失败)", msg.MsgID, msg.SeqID, "", msg.OffsetMs, DirClientToServer, opts.accountID)
					}
					failed++
					roundFailed++
					continue
				}
				payloadToSend = resolveVariablePayload(msg, sess.variableStore)
			}

			// 消息回调（传递 accountID 供前端展示账号列）
			if opts.OnMessage != nil {
				opts.OnMessage(msg.MsgName, msg.MsgID, msg.SeqID, payloadToSend, msg.OffsetMs, DirClientToServer, opts.accountID)
			}

			if sendErr := sendRawMessage(conn, msg.MsgID, payloadToSend, seqID); sendErr != nil {
				log.Printf("[重放] [account=%s] [发送失败] 第%d轮-第%d条 %s MsgID=%d SeqID=%d: %v", opts.accountID, r+1, i+1, msg.MsgName, msg.MsgID, seqID, sendErr)
				failed++
				roundFailed++
				continue
			}
			seqID++
			roundSent++
			sess.lastSeqID = seqID

			currentTotal := opts.alreadySent + roundTotal + roundSent
			if opts.OnProgress != nil {
				if !opts.OnProgress(opts.grandTotal, currentTotal, msg.MsgName) {
					log.Printf("[重放] [account=%s] 用户取消", opts.accountID)
					return roundTotal + roundSent, fmt.Errorf("重放已取消")
				}
			}
			log.Printf("[重放] [account=%s] [发送成功] 第%d轮-第%d条 %s (MsgID=%d, SeqID=%d)", opts.accountID, r+1, i+1, msg.MsgName, msg.MsgID, seqID-1)

			if SendIntervalMs > 0 {
				time.Sleep(time.Duration(SendIntervalMs) * time.Millisecond)
			}
		}

		roundTotal += roundSent
		log.Printf("[重放] [account=%s] 第%d轮完成: 成功=%d 跳过=%d 失败=%d 耗时=%v", opts.accountID, r+1, roundSent, roundSkipped, roundFailed, time.Since(roundStart))
	}

	if opts.OnProgress != nil {
		opts.OnProgress(opts.grandTotal, opts.alreadySent+roundTotal, "")
	}
	log.Printf("[重放] [account=%s] 发送循环结束: 成功=%d, 跳过=%d, 失败=%d", opts.accountID, roundTotal, skipped, failed)
	if failed > 0 || skipped > 0 {
		log.Printf("[重放] [account=%s] ⚠️ 存在未成功发送的消息！失败=%d 跳过=%d", opts.accountID, failed, skipped)
	}

	// 等待服务器返回最后一批 Ack
	if AckWaitMs > 0 {
		time.Sleep(time.Duration(AckWaitMs) * time.Millisecond)
	}
	return roundTotal, nil
}

// Replay 从录制文件重放消息到服务端（保留给命令行工具 cmd/tests/streamproxy 使用）
func Replay(filename string, serverAddr string, httpAddr string, openID string, onProgress ReplayProgressCallback, onMessage ReplayMessageCallback) error {
	return SendMessages(serverAddr, httpAddr, openID, nil, 1, 1, 1, nil, onProgress, onMessage, nil)
}
