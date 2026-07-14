package prototest

// Unity NetRecorder 日志分析命令组（unity-log）。
//
// 解析客户端 NetRecorder 输出的 .log 文件，提供：
//   - list:  列出日志目录下的 .log 文件
//   - show:  显示单条日志的摘要（方向、MsgID、MsgName、SeqID、登录边界）
//   - replay: 提取 C→S 游戏协议消息并用 proto-test 登录/重放管线发送
//
// 客户端日志格式（NetRecorder.RecordItem.ToString()）：
//
//	[YYYY-MM-DD HH:MM:SS:mmm] RecordType: StartSend; Time:HH:mm:ss:ms; MsgIdName:...; MsgId:N; SeqId: N; Msg:
//	<JSON body>
//
// RecordType 映射：
//   - StartSend / SendSuccess / WaitingToSend → C→S (direction="→")
//   - ReceiveSuccess → S→C (direction="←")
//
// 会话边界：取日志中最后一次 LoginResp（MsgId:2）之后的所有 C→S 消息。
// 登录由 proto-test AuthLogin + sendLoginReq 自行完成，不依赖客户端日志中的 LoginReq。

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/cases"
	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

// unityLogEntry 客户端 NetRecorder 日志中的单条记录
type unityLogEntry struct {
	RecordType  string // StartSend / SendSuccess / ReceiveSuccess / WaitingToSend / SendFail
	MsgID       uint32 // 消息 ID
	SeqID       uint32 // 序列号
	MsgName     string // 消息枚举名（来自日志 MsgIdName，可能不完整）
	PayloadJSON string // 原始 JSON body
}

// headerRe 匹配 NetRecorder 日志头部行
// 格式: [timestamp] RecordType: X; Time:...; MsgIdName:X; MsgId:N; SeqId: N; Msg:
var headerRe = regexp.MustCompile(
	`^\[[\d\s\-:]+\] RecordType: (\w+); Time:[\d:]+; MsgIdName:(\S+); MsgId:(\d+); SeqId: (\d+); Msg:$`,
)

// parseNetRecorderLog 解析客户端 NetRecorder .log 文件，返回所有记录条目
func parseNetRecorderLog(filePath string) ([]unityLogEntry, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer func() { _ = f.Close() }()

	var entries []unityLogEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 10<<20) // 10 MB 缓冲区，支持大型日志

	var (
		inBody      bool
		braceDepth  int
		bodyBuilder strings.Builder
		recordType  string
		msgName     string
		msgID       uint32
		seqID       uint32
	)

	for scanner.Scan() {
		line := scanner.Text()

		if !inBody {
			matches := headerRe.FindStringSubmatch(strings.TrimSpace(line))
			if matches == nil {
				continue
			}
			id, _ := strconv.Atoi(matches[3])
			sid, _ := strconv.Atoi(matches[4])
			recordType = matches[1]
			msgName = matches[2]
			msgID = uint32(id)
			seqID = uint32(sid)
			inBody = true
			braceDepth = 0
			bodyBuilder.Reset()
			continue
		}

		// 在 JSON body 内，统计花括号深度
		for _, ch := range line {
			switch ch {
			case '{':
				braceDepth++
			case '}':
				braceDepth--
			}
		}

		if bodyBuilder.Len() > 0 {
			bodyBuilder.WriteByte('\n')
		}
		bodyBuilder.WriteString(line)

		// braceDepth 回到 0 表示 JSON body 完整接收
		if braceDepth == 0 {
			entries = append(entries, unityLogEntry{
				RecordType:  recordType,
				MsgName:     msgName,
				MsgID:       msgID,
				SeqID:       seqID,
				PayloadJSON: bodyBuilder.String(),
			})
			inBody = false
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %w", err)
	}

	return entries, nil
}

// recordTypeToDirection 将客户端 ERecordType 映射为 proto-test 方向标记
func recordTypeToDirection(rt string) string {
	switch rt {
	case "StartSend", "SendSuccess", "WaitingToSend":
		return cases.DirClientToServer // "→"
	case "ReceiveSuccess":
		return cases.DirServerToClient // "←"
	default:
		return "?"
	}
}

// findLastLoginRespIndex 找最后一次 LoginResp（MsgId=2, ReceiveSuccess）的索引
// 返回索引和总记录数。未找到返回 -1。
func findLastLoginRespIndex(entries []unityLogEntry) int {
	lastIdx := -1
	for i, e := range entries {
		if e.MsgID == 2 && e.RecordType == "ReceiveSuccess" {
			lastIdx = i
		}
	}
	return lastIdx
}

// newUnityLogCmd 创建 unity-log 命令组
func newUnityLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unity-log",
		Short: "解析 Unity 客户端 NetRecorder 日志",
		Long: `解析 Unity 客户端 NetRecorder 输出的 .log 文件。

提供三个子命令：
  list   列出日志目录下的 .log 文件
  show   显示单条日志的消息摘要
  replay 提取 C→S 游戏协议消息并重放

日志默认目录：D:\work\client\Master\Card\Log\net（可通过 --log-dir 指定）。`,
	}

	cmd.AddCommand(
		newUnityLogListCmd(),
		newUnityLogShowCmd(),
		newUnityLogReplayCmd(),
	)

	return cmd
}

// newUnityLogListCmd 创建 unity-log list 子命令
func newUnityLogListCmd() *cobra.Command {
	var logDir string

	cmd := &cobra.Command{
		Use:   "list [log-dir]",
		Short: "列出日志目录下的 .log 文件",
		Long: `列出指定日志目录（及其子目录按日期组织）下的 NetRecorder .log 文件。

默认目录可通过 --log-dir 指定；未指定时尝试常见路径。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := logDir
			if len(args) > 0 {
				dir = args[0]
			}
			if dir == "" {
				dir = defaultUnityLogDir()
			}

			files, err := findUnityLogFiles(dir)
			if err != nil {
				return fmt.Errorf("查找日志文件失败: %w", err)
			}

			if len(files) == 0 {
				cmd.Printf("目录 %q 下未找到 .log 文件\n", dir)
				return nil
			}

			cmd.Printf("%s 下的日志文件（共 %d 个）：\n", dir, len(files))
			for _, f := range files {
				rel, _ := filepath.Rel(dir, f)
				info, err := os.Stat(f)
				sizeStr := "?"
				if err == nil {
					sizeStr = formatBytes(info.Size())
				}
				cmd.Printf("  %-60s %8s\n", rel, sizeStr)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&logDir, "log-dir", "", "日志根目录（默认 D:\\work\\client\\Master\\Card\\Log\\net）")

	return cmd
}

// newUnityLogShowCmd 创建 unity-log show 子命令
func newUnityLogShowCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show <log-file>",
		Short: "显示 NetRecorder 日志摘要",
		Long: `解析单条 NetRecorder 日志并输出摘要信息。

summary 格式：消息序号、方向、MsgID、MsgName、SeqID。
json 格式：完整记录数组（含 payload）。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logFilePath := args[0]

			entries, err := parseNetRecorderLog(logFilePath)
			if err != nil {
				return fmt.Errorf("解析日志失败: %w", err)
			}

			lastLoginIdx := findLastLoginRespIndex(entries)

			switch format {
			case "json":
				out := struct {
					File         string          `json:"file"`
					Total        int             `json:"total"`
					LastLoginIdx int             `json:"last_login_idx"`
					Entries      []unityLogEntry `json:"entries"`
				}{
					File:         logFilePath,
					Total:        len(entries),
					LastLoginIdx: lastLoginIdx,
					Entries:      entries,
				}
				b, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return fmt.Errorf("JSON 编码失败: %w", err)
				}
				cmd.Println(string(b))

			case "summary", "":
				cmd.Printf("日志文件: %s\n", logFilePath)
				cmd.Printf("总记录数: %d\n", len(entries))
				if lastLoginIdx >= 0 {
					cmd.Printf("最后一次登录完成: 第 %d 条记录\n", lastLoginIdx+1)
					cmd.Printf("登录后可重放 C→S 消息: %d 条\n", countReplayableEntries(entries, lastLoginIdx))
				} else {
					cmd.Println("未找到 LoginResp，无法确定会话边界")
				}
				cmd.Println()
				cmd.Println("消息摘要：")
				cmd.Printf("%-6s %-4s %-8s %-35s %-10s\n", "序号", "方向", "MsgID", "MsgName", "SeqID")
				for i, e := range entries {
					dir := recordTypeToDirection(e.RecordType)
					marker := ""
					if i == lastLoginIdx {
						marker = "  <-- 登录边界"
					}
					cmd.Printf("%-6d %-4s %-8d %-35s %-10d%s\n", i+1, dir, e.MsgID, e.MsgName, e.SeqID, marker)
				}

			default:
				return fmt.Errorf("未知格式 %q，支持 summary|json", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "summary", "输出格式：summary|json")

	return cmd
}

// newUnityLogReplayCmd 创建 unity-log replay 子命令
func newUnityLogReplayCmd() *cobra.Command {
	var (
		serverAddr  string
		httpAddr    string
		openID      string
		rangeSpec   string
		repeat      int
		intervalMs  int
		ackWaitMs   int
		concurrency int
		printAck    bool
		timeout     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "replay <log-file>",
		Short: "解析客户端 NetRecorder 日志并重放其中的 C→S 消息",
		Long: `解析客户端 NetRecorder 输出的 .log 文件，提取其中客户端→服务端方向的消息，
在内存中构建为测试用例并通过 proto-test 登录/重放管线发送。

会话边界：取日志中最后一次 LoginResp（MsgId=2）之后的所有 C→S 消息。
登录由 proto-test AuthLogin 完成，不依赖客户端日志中的 LoginReq。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logFilePath := args[0]

			// 验证必要参数
			if serverAddr == "" {
				return fmt.Errorf("--server 为必填（客户端日志不包含服务器地址）")
			}
			if openID == "" {
				return fmt.Errorf("--openid 为必填")
			}

			// 解析账号范围
			rangeStart, rangeEnd, err := parseRange(rangeSpec)
			if err != nil {
				return fmt.Errorf("解析 --range 失败: %w", err)
			}

			// 解析客户端日志
			entries, err := parseNetRecorderLog(logFilePath)
			if err != nil {
				return fmt.Errorf("解析客户端日志失败: %w", err)
			}
			log.Printf("[unity-log replay] 解析日志 %q: 共 %d 条记录", logFilePath, len(entries))

			lastLoginIdx := findLastLoginRespIndex(entries)
			if lastLoginIdx < 0 {
				return fmt.Errorf("未在日志中找到 LoginResp（MsgId=2, ReceiveSuccess），无法确定会话边界")
			}
			log.Printf("[unity-log replay] 最后一次登录完成于第 %d 条记录（共 %d 条），之后的消息作为重放用例", lastLoginIdx+1, len(entries))

			// 构建 Recording：只取登录后的 C→S 游戏协议消息
			rec := &cases.Recording{
				Version:    cases.RecordingVersion,
				RecordedAt: time.Now().Format(time.RFC3339),
				ServerAddr: serverAddr,
				Messages:   make([]*cases.RecordEntry, 0),
			}

			for i := lastLoginIdx + 1; i < len(entries); i++ {
				e := entries[i]
				dir := recordTypeToDirection(e.RecordType)
				if dir != cases.DirClientToServer {
					continue
				}
				if e.MsgID < 1000 {
					continue
				}
				if !json.Valid([]byte(e.PayloadJSON)) {
					log.Printf("[unity-log replay] 第 %d 条消息 Payload 不是合法 JSON，跳过 (MsgID=%d)", i+1, e.MsgID)
					continue
				}

				msgName := protocol.GetMsgName(uint16(e.MsgID))
				rec.Messages = append(rec.Messages, &cases.RecordEntry{
					MsgID:       uint16(e.MsgID),
					MsgName:     msgName,
					SeqID:       e.SeqID,
					Direction:   dir,
					PayloadJSON: json.RawMessage(e.PayloadJSON),
				})
			}

			if len(rec.Messages) == 0 {
				return fmt.Errorf("登录后未找到可重放的 C→S 游戏协议消息（MsgID>=1000）")
			}
			log.Printf("[unity-log replay] 提取 %d 条 C→S 消息作为用例（跳过 %d 条 S→C/框架消息）",
				len(rec.Messages), len(entries)-lastLoginIdx-1-len(rec.Messages))

			messages := recordingToRecordMessages(rec)
			if len(messages) == 0 {
				return fmt.Errorf("无可发送的消息")
			}

			httpAddrFinal := httpAddr
			if httpAddrFinal == "" {
				httpAddrFinal = deriveHTTPAddr(serverAddr)
			}

			protocol.SendIntervalMs = intervalMs
			protocol.AckWaitMs = ackWaitMs
			protocol.MaxConcurrency = concurrency

			log.Printf("[unity-log replay] TCP=%s HTTP=%s 账号前缀=%s 范围=%d-%d 消息数=%d 重复=%d",
				serverAddr, httpAddrFinal, openID, rangeStart, rangeEnd, len(messages), repeat)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if printAck {
				log.Printf("[unity-log replay] --print-ack 已启用")
			}

			err = protocol.NewReplayer(protocol.ReplayOptions{
				ServerAddr:  serverAddr,
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

			cmd.Printf("[unity-log replay] 完成")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server", "", "目标 TCP 服务器地址（必填，如 10.254.114.204:18000）")
	cmd.Flags().StringVar(&httpAddr, "http", "", "HTTP 认证地址（默认从 --server 推导，同 IP:20144）")
	cmd.Flags().StringVar(&openID, "openid", "", "账号前缀，如 test（必填，实际账号为 test1、test2...）")
	cmd.Flags().StringVar(&rangeSpec, "range", "1-1", "账号范围，如 1-10 表示 10 个账号")
	cmd.Flags().IntVarP(&repeat, "repeat", "n", 1, "每个账号重复轮数")
	cmd.Flags().IntVar(&intervalMs, "interval", 1000, "消息发送间隔（毫秒），0 表示不等待")
	cmd.Flags().IntVar(&ackWaitMs, "ack-wait", 2000, "重放完成后等待 Ack 时间（毫秒）")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "最大并发账号数（0=不限）")
	cmd.Flags().BoolVar(&printAck, "print-ack", false, "输出服务端返回的每条消息（Ack/Ntf）为 NDJSON")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "整体超时")

	return cmd
}

// defaultUnityLogDir 返回默认 Unity 客户端日志目录
func defaultUnityLogDir() string {
	candidate := `D:\work\client\Master\Card\Log\net`
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return "Log/net"
}

// findUnityLogFiles 递归查找目录下的所有 .log 文件，按完整路径排序
func findUnityLogFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无权限目录
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".log") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// countReplayableEntries 统计登录后可重放的 C→S 游戏协议消息数量
func countReplayableEntries(entries []unityLogEntry, lastLoginIdx int) int {
	count := 0
	for i := lastLoginIdx + 1; i < len(entries); i++ {
		e := entries[i]
		if recordTypeToDirection(e.RecordType) != cases.DirClientToServer {
			continue
		}
		if e.MsgID < 1000 {
			continue
		}
		count++
	}
	return count
}

// formatBytes 将字节数格式化为人类可读字符串
func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n >= div*unit && exp < 3 {
		div *= unit
		exp++
	}
	switch exp {
	case 0:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(div))
	case 1:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(div))
	case 2:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(div))
	}
	return fmt.Sprintf("%d B", n)
}
