package prototest

// 发送单条自定义消息 CLI 子命令（send-msg）。

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	protocol "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/proto-test/msg"
)

func newSendMsgCmd() *cobra.Command {
	var (
		msgID       uint16
		msgName     string
		payload     string
		payloadFile string
		openID      string
		serverAddr  string
		httpAddr    string
		intervalMs  int
		ackWaitMs   int
		timeout     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "send-msg",
		Short: "发送单条自定义消息（需 --msg-id/--msg-name/--payload）",
		RunE: func(cmd *cobra.Command, args []string) error {
			if openID == "" {
				return fmt.Errorf("--openid 为必填项")
			}
			if msgName == "" {
				return fmt.Errorf("--msg-name 为必填项")
			}

			// 确定 payload
			payloadStr := payload
			if payloadFile != "" {
				data, err := os.ReadFile(payloadFile)
				if err != nil {
					return fmt.Errorf("读取 --payload-file 失败: %w", err)
				}
				payloadStr = string(data)
			}
			if payloadStr == "" {
				payloadStr = "{}"
			}

			// 确定 TCP 服务器地址
			tcpAddr := serverAddr
			if tcpAddr == "" {
				return fmt.Errorf("--server 为必填项（send-msg 无用例文件可读取默认地址）")
			}

			// 推导 HTTP 地址
			httpAddrFinal := httpAddr
			if httpAddrFinal == "" {
				httpAddrFinal = deriveHTTPAddr(tcpAddr)
			}

			// 构造单条消息（Direction 留空，SendMessages 内部按 Req 处理）
			messages := []protocol.RecordMessage{
				{
					MsgID:       msgID,
					MsgName:     msgName,
					PayloadJSON: payloadStr,
				},
			}

			// 设置全局参数
			protocol.SendIntervalMs = intervalMs
			protocol.AckWaitMs = ackWaitMs

			log.Printf("[send-msg] 账号=%s MsgID=%d MsgName=%s TCP=%s HTTP=%s",
				openID+"1", msgID, msgName, tcpAddr, httpAddrFinal)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			err := protocol.NewReplayer(protocol.ReplayOptions{
				ServerAddr:  tcpAddr,
				HTTPAddr:    httpAddrFinal,
				OpenID:      openID,
				Messages:    messages,
				RepeatCount: 1,
				RangeStart:  1,
				RangeEnd:    1,
				Context:     ctx,
			}).SendMessages()
			if err != nil {
				return fmt.Errorf("发送失败: %w", err)
			}

			cmd.Printf("[send-msg] 完成")
			return nil
		},
	}

	cmd.Flags().Uint16Var(&msgID, "msg-id", 0, "消息 ID（如 1001）")
	cmd.Flags().StringVar(&msgName, "msg-name", "", "消息名称（如 GmCommandReq，必填）")
	cmd.Flags().StringVar(&payload, "payload", "", "payload JSON 字符串（如 '{\"content\":\"//AddItem 1000001 999\"}'）")
	cmd.Flags().StringVar(&payloadFile, "payload-file", "", "从文件读取 payload（避免 shell 转义）")
	cmd.Flags().StringVar(&openID, "openid", "", "账号前缀（必填，实际账号为 <openid>1）")
	cmd.Flags().StringVar(&serverAddr, "server", "", "目标 TCP 服务器地址（必填，如 10.254.114.204:18000）")
	cmd.Flags().StringVar(&httpAddr, "http", "", "HTTP 认证地址（默认从 --server 推导，同 IP:20144）")
	cmd.Flags().IntVar(&intervalMs, "interval", 1000, "消息发送间隔（毫秒）")
	cmd.Flags().IntVar(&ackWaitMs, "ack-wait", 2000, "发送完成后等待 Ack 时间（毫秒）")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "整体超时")

	return cmd
}
