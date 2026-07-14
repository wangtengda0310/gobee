package prototest

// 用例管理 CLI 子命令（case list/show/edit/create/delete）。
// list/show 已实现，edit/create/delete 仍为占位。

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// newTestCaseServiceForCLI 为 CLI 场景构造 TestCaseService。
// RecordFileService 是无状态空结构体，可直接构造。
func newTestCaseServiceForCLI() *TestCaseService {
	return NewTestCaseService(NewRecordFileService())
}

// newCaseCmd 创建 case 父命令及其子命令。
func newCaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "case",
		Short: "用例管理（列表/查看/编辑/创建/删除）",
	}

	cmd.AddCommand(
		newCaseListCmd(),
		newCaseShowCmd(),
		newCaseEditCmd(),
		newCaseCreateCmd(),
		newCaseDeleteCmd(),
	)

	return cmd
}

func newCaseListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有测试用例",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := newTestCaseServiceForCLI()
			cases, err := svc.LoadTestCaseList()
			if err != nil {
				return fmt.Errorf("加载用例列表失败: %w", err)
			}

			switch format {
			case "json":
				if len(cases) == 0 {
					cmd.Println("[]")
					return nil
				}
				b, err := json.MarshalIndent(cases, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化失败: %w", err)
				}
				cmd.Println(string(b))
			default: // table
				if len(cases) == 0 {
					cmd.Println("（无测试用例）")
					return nil
				}
				cmd.Printf("%-20s %8s  %-30s %s\n", "NAME", "MSGS", "SERVER", "CREATED_AT")
				for _, c := range cases {
					cmd.Printf("%-20s %8d  %-30s %s\n", c.Name, c.MessageCount, c.ServerAddr, c.CreatedAt)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "输出格式：table（默认）或 json")
	return cmd
}

func newCaseShowCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "查看指定用例的消息详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			svc := newTestCaseServiceForCLI()
			data, err := svc.LoadTestCase(name)
			if err != nil {
				return fmt.Errorf("加载用例 %q 失败: %w", name, err)
			}

			switch format {
			case "json":
				out := buildCaseShowJSON(data)
				b, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化失败: %w", err)
				}
				cmd.Println(string(b))
			default: // summary
				cmd.Printf("用例: %s\n", name)
				cmd.Printf("服务器: %s\n", data.ServerAddr)
				cmd.Printf("录制时间: %s\n", data.RecordedAt)
				cmd.Printf("消息数: %d\n", len(data.Messages))
				cmd.Println()
				for i, m := range data.Messages {
					desc := m.Descript
					if desc == "" {
						desc = "-"
					}
					cmd.Printf("%d. %s (%s)\n", i+1, m.MsgName, desc)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "summary", "输出格式：summary（默认，仅序号+名称+描述）或 json（完整含 payload）")
	return cmd
}

func newCaseEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: "修改已有用例的消息内容（--merge/--payload/--desc/--append/--remove）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: 实现 — 按 --msg N 定位消息，--merge(JSON merge patch)/--payload(整体替换)/--desc 修改；
			//  --append 追加新消息；--remove N 删除指定序号消息
			name := args[0]
			cmd.Printf("[debug] proto-test case edit %s — 待实现\n", name)
			return nil
		},
	}
}

func newCaseCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "从文件（--file）或标准输入（--stdin）创建用例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: 实现 — 读取 --file 或 stdin 的 JSON，校验后写入 cases/proto_cases/<name>.json
			name := args[0]
			cmd.Printf("[debug] proto-test case create %s — 待实现\n", name)
			return nil
		},
	}
}

func newCaseDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "删除指定用例（--force 跳过确认）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: 实现 — 删除 cases/proto_cases/<name>.json，未带 --force 时提示确认
			name := args[0]
			cmd.Printf("[debug] proto-test case delete %s — 待实现\n", name)
			return nil
		},
	}
}

// caseShowMessage 是 case show --format json 的单条消息输出结构。
// 使用 1-based 的 seq 字段（与 summary 序号、--select、--msg 一致），
// 刻意不输出 RecordEntryView.Index（0-based，会与 1-based 约定混淆）。
type caseShowMessage struct {
	Seq      int            `json:"seq"` // 1-based 序号，与 --select/--msg 一致
	MsgID    uint16         `json:"msg_id"`
	MsgName  string         `json:"msg_name"`
	Descript string         `json:"descript"`
	Payload  map[string]any `json:"payload"`
}

// caseShowJSON 是 case show --format json 的顶层输出结构。
type caseShowJSON struct {
	ServerAddr   string            `json:"server_addr"`
	RecordedAt   string            `json:"recorded_at"`
	MessageCount int               `json:"message_count"`
	Messages     []caseShowMessage `json:"messages"`
}

// buildCaseShowJSON 从 RecordFileData 构造 CLI json 输出，统一使用 1-based seq。
func buildCaseShowJSON(data *RecordFileData) caseShowJSON {
	msgs := make([]caseShowMessage, len(data.Messages))
	for i, m := range data.Messages {
		msgs[i] = caseShowMessage{
			Seq:      i + 1, // 1-based
			MsgID:    m.MsgID,
			MsgName:  m.MsgName,
			Descript: m.Descript,
			Payload:  m.Payload,
		}
	}
	return caseShowJSON{
		ServerAddr:   data.ServerAddr,
		RecordedAt:   data.RecordedAt,
		MessageCount: len(data.Messages),
		Messages:     msgs,
	}
}
