package prototest

// 用例重放 CLI 子命令（replay）。

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

func newReplayCmd() *cobra.Command {
	var (
		serverAddr      string
		httpAddr        string
		openID          string
		rangeSpec       string
		repeat          int
		intervalMs      int
		ackWaitMs       int
		concurrency     int
		maxRetries      int
		retryIntervalMs int
		selectSpec      string
		selectNames     []string
		printAck        bool
		timeout         time.Duration
	)

	cmd := &cobra.Command{
		Use:   "replay <case-name>",
		Short: "重放测试用例（支持账号范围/重复/消息子集选择）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseName := args[0]

			if openID == "" {
				return fmt.Errorf("--openid 为必填项")
			}

			// 解析账号范围
			rangeStart, rangeEnd, err := parseRange(rangeSpec)
			if err != nil {
				return fmt.Errorf("解析 --range 失败: %w", err)
			}

			// 加载用例文件
			casePath := caseFilePath(caseName)
			rec, err := LoadTestCaseFromFile(casePath)
			if err != nil {
				return fmt.Errorf("加载用例 %q 失败: %w", caseName, err)
			}

			// 转换为 RecordMessage 列表
			messages := recordingToRecordMessages(rec)
			if len(messages) == 0 {
				return fmt.Errorf("用例 %q 无可发送的消息", caseName)
			}

			// 应用消息子集过滤
			messages, err = filterMessages(messages, selectSpec, selectNames)
			if err != nil {
				return err
			}
			if len(messages) == 0 {
				return fmt.Errorf("过滤后无可发送的消息")
			}

			// 确定服务器地址
			tcpAddr := serverAddr
			if tcpAddr == "" {
				tcpAddr = rec.ServerAddr
			}
			if tcpAddr == "" {
				return fmt.Errorf("未指定 TCP 服务器地址（用例文件无 server_addr 且未传 --server）")
			}

			// 推导 HTTP 地址（同 IP，端口 20144）
			httpAddrFinal := httpAddr
			if httpAddrFinal == "" {
				httpAddrFinal = deriveHTTPAddr(tcpAddr)
			}

			// 设置全局重放参数
			protocol.SendIntervalMs = intervalMs
			protocol.AckWaitMs = ackWaitMs
			protocol.MaxConcurrency = concurrency
			protocol.GlobalRetryConfig = protocol.RetryConfig{
				MaxRetries:     maxRetries,
				RetryInterval:  time.Duration(retryIntervalMs) * time.Millisecond,
				MinConcurrency: 1,
			}

			log.Printf("[replay] 用例=%s 账号前缀=%s 范围=%d-%d TCP=%s HTTP=%s 消息数=%d 重复=%d 最大重试=%d 重试间隔=%dms",
				caseName, openID, rangeStart, rangeEnd, tcpAddr, httpAddrFinal, len(messages), repeat, maxRetries, retryIntervalMs)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if printAck {
				log.Printf("[replay] --print-ack 已启用，服务端返回的消息将以 NDJSON 输出到 stdout")
			}

			err = protocol.NewReplayer(protocol.ReplayOptions{
				ServerAddr:  tcpAddr,
				HTTPAddr:    httpAddrFinal,
				OpenID:      openID,
				Messages:    messages,
				RepeatCount: repeat,
				RangeStart:  rangeStart,
				RangeEnd:    rangeEnd,
				Context:     ctx,
				OnMessage:   makePrintAckCallback(printAck, cmd),
				RetryConfig: protocol.GlobalRetryConfig,
				Concurrency: protocol.MaxConcurrency,
			}).SendMessagesWithRetry()
			if err != nil {
				return fmt.Errorf("重放失败: %w", err)
			}

			cmd.Printf("[replay] 完成")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server", "", "目标 TCP 服务器地址（默认读用例文件 server_addr）")
	cmd.Flags().StringVar(&httpAddr, "http", "", "HTTP 认证地址（默认从 --server 推导，同 IP:20144）")
	cmd.Flags().StringVar(&openID, "openid", "", "账号前缀，如 test（必填，实际账号为 test1、test2...）")
	cmd.Flags().StringVar(&rangeSpec, "range", "1-1", "账号范围，如 1-10 表示 10 个账号")
	cmd.Flags().IntVarP(&repeat, "repeat", "n", 1, "每个账号重复轮数")
	cmd.Flags().IntVar(&intervalMs, "interval", 1000, "消息发送间隔（毫秒），0 表示不等待")
	cmd.Flags().IntVar(&ackWaitMs, "ack-wait", 2000, "重放完成后等待 Ack 时间（毫秒）")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "最大并发账号数（0=使用默认上限 10）")
	cmd.Flags().IntVar(&maxRetries, "max-retries", 3, "因登录限流失败时的最大重试次数")
	cmd.Flags().IntVar(&retryIntervalMs, "retry-interval", 500, "重试间隔（毫秒）")
	cmd.Flags().StringVar(&selectSpec, "select", "", "按 1-based 序号选择消息子集，如 3-7、3,5,9、1-3,7")
	cmd.Flags().StringSliceVar(&selectNames, "select-name", nil, "按消息名筛选（可多次指定），如 GmCommandReq")
	cmd.Flags().BoolVar(&printAck, "print-ack", false, "输出服务端返回的每条消息（Ack/Ntf）为 NDJSON")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "整体超时")

	return cmd
}

// recordingToRecordMessages 将 Recording 转换为 SendMessages 所需的 RecordMessage 列表。
// 仅保留 Req 方向（客户端→服务端）的消息。
func recordingToRecordMessages(rec *protocol.Recording) []protocol.RecordMessage {
	if rec == nil {
		return nil
	}
	out := make([]protocol.RecordMessage, 0, len(rec.Messages))
	for _, e := range rec.Messages {
		if e == nil || !IsReqDirection(e.Direction) {
			continue
		}
		out = append(out, protocol.RecordMessage{
			MsgID:       e.MsgID,
			MsgName:     e.MsgName,
			PayloadJSON: string(e.PayloadJSON),
			Direction:   e.Direction,
		})
	}
	return out
}

// parseRange 解析账号范围字符串，如 "1-10"、"3-3"、"5"。
func parseRange(spec string) (int, int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 1, 1, nil
	}
	parts := strings.SplitN(spec, "-", 2)
	if len(parts) == 1 {
		n, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("无效的数字 %q: %w", parts[0], err)
		}
		return n, n, nil
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("无效的起始值 %q: %w", parts[0], err)
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("无效的结束值 %q: %w", parts[1], err)
	}
	if start > end {
		return 0, 0, fmt.Errorf("起始值 %d 大于结束值 %d", start, end)
	}
	return start, end, nil
}

// filterMessages 按 --select（1-based 序号）和 --select-name（消息名）过滤消息子集。
// 两个条件取交集（同时满足）。序号是 1-based，与 case show summary 输出一致。
func filterMessages(messages []protocol.RecordMessage, selectSpec string, selectNames []string) ([]protocol.RecordMessage, error) {
	if selectSpec == "" && len(selectNames) == 0 {
		return messages, nil
	}

	// 解析 --select 序号集合
	var selectedIdx map[int]bool
	if selectSpec != "" {
		selectedIdx = make(map[int]bool)
		segments := strings.Split(selectSpec, ",")
		for _, seg := range segments {
			seg = strings.TrimSpace(seg)
			if seg == "" {
				continue
			}
			if strings.Contains(seg, "-") {
				parts := strings.SplitN(seg, "-", 2)
				start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				if err != nil {
					return nil, fmt.Errorf("无效的序号段 %q: %w", seg, err)
				}
				end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					return nil, fmt.Errorf("无效的序号段 %q: %w", seg, err)
				}
				for i := start; i <= end; i++ {
					selectedIdx[i] = true
				}
			} else {
				n, err := strconv.Atoi(seg)
				if err != nil {
					return nil, fmt.Errorf("无效的序号 %q: %w", seg, err)
				}
				selectedIdx[n] = true
			}
		}
	}

	// 构建 select-name 集合
	var nameSet map[string]bool
	if len(selectNames) > 0 {
		nameSet = make(map[string]bool, len(selectNames))
		for _, n := range selectNames {
			nameSet[n] = true
		}
	}

	var out []protocol.RecordMessage
	for i, m := range messages {
		idx := i + 1 // 1-based
		// --select 过滤
		if selectedIdx != nil && !selectedIdx[idx] {
			continue
		}
		// --select-name 过滤
		if nameSet != nil && !nameSet[m.MsgName] {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// printAckEntry 是 --print-ack 输出的 NDJSON 行结构。
type printAckEntry struct {
	Account string `json:"account"`
	MsgName string `json:"msg_name"`
	MsgID   uint16 `json:"msg_id"`
	SeqID   uint32 `json:"seq_id"`
	Payload any    `json:"payload"`
}

// makePrintAckCallback 构造一个 onMessage 回调。
// 当 enabled=true 且消息方向为服务端→客户端（Ack/Ntf）时，以 NDJSON 行输出到 stdout。
// 多账号并发时通过互斥锁保证输出不交错。
func makePrintAckCallback(enabled bool, cmd *cobra.Command) protocol.ReplayMessageCallback {
	if !enabled {
		return nil
	}
	var mu sync.Mutex
	return func(msgName string, msgID uint16, seqID uint32, payloadJSON string, offsetMs int, direction string, accountID string) {
		if direction != protocol.DirServerToClient {
			return
		}
		entry := printAckEntry{
			Account: accountID,
			MsgName: msgName,
			MsgID:   msgID,
			SeqID:   seqID,
		}
		if payloadJSON != "" {
			_ = json.Unmarshal([]byte(payloadJSON), &entry.Payload)
		}
		line, err := json.Marshal(entry)
		if err != nil {
			return
		}
		mu.Lock()
		cmd.Println(string(line))
		mu.Unlock()
	}
}
