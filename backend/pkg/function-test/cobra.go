package functiontest

// fight-test CLI 子命令定义。
// 子命令实现拆分到本文件（list / run）。
// 执行路径复用 RunRobotTest（与 MCP run_fight_test 同一条代码路径），
// 配置默认值复用 FuncCaseConfigService.GetConfig()。

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

//go:embed cobra-help.md
var fightTestHelpText string

// NewFightTestCmd 创建 fight-test 子命令。
// 该命令用于执行 cases/fight_cases 目录下的战斗测试用例。
func NewFightTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fight-test",
		Short: "战斗测试用例工具",
		Long:  fightTestHelpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	cmd.AddCommand(
		newFightTestListCmd(),
		newFightTestRunCmd(),
	)

	return cmd
}

// newFightTestCaseServiceForCLI 为 CLI 场景构造 JsonCaseService。
// emitter 传 newCLILogEmitter()：CLI 把每条 robotLog 事件实时格式化打印到 stdout，
// 让 fight-test run 子命令能看到逐条 step 执行过程（UseHeroSkill/PlayCard/断言结果…），
// 体验对齐前端 robot-test-log 渲染。详见 cli_emitter.go。
func newFightTestCaseServiceForCLI() *JsonCaseService {
	return NewJsonCaseService(newCLILogEmitter())
}

// newFightTestListCmd 创建 list 子命令：列出 fight_cases 目录下的用例。
func newFightTestListCmd() *cobra.Command {
	var (
		dir    string
		heroID int
		format string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出战斗测试用例",
		RunE: func(cmd *cobra.Command, args []string) error {
			// --dir 未显式指定时取配置默认值（cases/fight_cases）
			if !cmd.Flags().Changed("dir") {
				if cfg, err := NewFuncCaseConfigService().GetConfig(); err == nil && cfg.JsonsDir != "" {
					dir = cfg.JsonsDir
				}
			}

			svc := newFightTestCaseServiceForCLI()
			cases, err := svc.GetCaseList(dir)
			if err != nil {
				return fmt.Errorf("读取用例目录 %q 失败: %w", dir, err)
			}

			// 按英雄 ID 过滤文件（支持下划线/连字符两种命名格式）
			if heroID > 0 {
				cases = filterCasesByHeroID(cases, dir, heroID)
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
					cmd.Println("（无战斗测试用例）")
					return nil
				}
				cmd.Printf("%-28s %6s %6s  %s\n", "CASE", "STEPS", "HEROS", "FILE")
				for _, c := range cases {
					cmd.Printf("%-28s %6d %6d  %s\n", c.Case, c.StepCount, c.HeroCount, c.FileName)
				}
				cmd.Printf("\n共 %d 条用例（目录: %s）\n", len(cases), dir)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "cases/fight_cases", "用例目录（默认取配置 jsons_dir）")
	cmd.Flags().IntVar(&heroID, "hero", 0, "按英雄 ID 过滤（匹配 {id}_*.json 和 {id}-*.json）")
	cmd.Flags().StringVar(&format, "format", "table", "输出格式：table（默认）或 json")
	return cmd
}

// filterCasesByHeroID 按英雄 ID 过滤用例。
// 复用 MCP buildFightTestRunner（tools.go）的 glob 约定：
// 文件名格式为 {heroID}_*.json 或 {heroID}-*.json。
func filterCasesByHeroID(cases []CaseInfo, dir string, heroID int) []CaseInfo {
	want := make(map[string]bool)
	for _, p := range []string{fmt.Sprintf("%d_*.json", heroID), fmt.Sprintf("%d-*.json", heroID)} {
		matches, _ := filepath.Glob(filepath.Join(dir, p))
		for _, m := range matches {
			want[filepath.Base(m)] = true
		}
	}
	if len(want) == 0 {
		return nil
	}
	var out []CaseInfo
	for _, c := range cases {
		if want[c.FileName] {
			out = append(out, c)
		}
	}
	return out
}

// newFightTestRunCmd 创建 run 子命令：运行战斗测试。
// 执行路径与 MCP run_fight_test（buildFightTestRunner）一致：组装过滤 → RunRobotTest → GetTestLogs。
func newFightTestRunCmd() *cobra.Command {
	var (
		serverAddr  string
		serverPort  int
		dir         string
		heroID      int
		caseName    string
		files       []string
		robotPrefix string
		opTimeMs    int
		concurrency int
		timeout     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "运行战斗测试用例",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 读取配置，为未显式指定的 flag 提供默认值
			configSvc := NewFuncCaseConfigService()
			cfg, err := configSvc.GetConfig()
			if err != nil {
				return fmt.Errorf("读取配置失败: %w", err)
			}

			// 未显式指定的 flag 回退到配置值
			if !cmd.Flags().Changed("server") {
				serverAddr = cfg.ServerAddr
			}
			if !cmd.Flags().Changed("port") {
				serverPort = cfg.ServerPort
			}
			if !cmd.Flags().Changed("dir") {
				dir = cfg.JsonsDir
			}
			if !cmd.Flags().Changed("prefix") {
				robotPrefix = cfg.RobotPrefix
			}
			if !cmd.Flags().Changed("concurrency") {
				concurrency = cfg.Concurrency
			}

			if serverAddr == "" {
				return fmt.Errorf("缺少目标服务器地址，请通过 --server 指定或在配置中设置 server_addr")
			}

			// 组装过滤用例文件名（优先级：--file > --hero > 全部）
			var caseFiles []string
			var filterCaseNames []string
			if caseName != "" {
				filterCaseNames = []string{caseName}
			}
			if len(files) > 0 {
				caseFiles = files
			} else if heroID > 0 {
				for _, p := range []string{fmt.Sprintf("%d_*.json", heroID), fmt.Sprintf("%d-*.json", heroID)} {
					matches, _ := filepath.Glob(filepath.Join(dir, p))
					for _, m := range matches {
						caseFiles = append(caseFiles, filepath.Base(m))
					}
				}
				if len(caseFiles) == 0 {
					return fmt.Errorf("未找到英雄 ID %d 对应的用例文件（在 %s 中）", heroID, dir)
				}
			}

			log.Printf("[fight-test] server=%s:%d dir=%s hero=%d case=%q files=%v prefix=%s concurrency=%d",
				serverAddr, serverPort, dir, heroID, caseName, caseFiles, robotPrefix, concurrency)

			svc := newFightTestCaseServiceForCLI()
			err = svc.RunRobotTest(
				serverAddr,                    // ip
				fmt.Sprintf("%d", serverPort), // port
				robotPrefix,                   // prefix
				"cli-fight-test",              // desc
				"",                            // feishuGuid
				float64(cfg.LoginTime),        // loginTime
				caseFiles,                     // filterCasesOption（文件名列表）
				filterCaseNames,               // filterCaseNameOption（用例名列表）
				&dir,                          // dir
				nil,                           // filesData
				opTimeMs,                      // opTimeMs
				false,                         // feishuNtf（CLI 不发飞书通知）
				false,                         // debugLevel
				false,                         // debugLog
				uint(concurrency),             // concurrency
				timeout,                       // maxTimeoutPerCase（--timeout 透传）
			)
			if err != nil {
				return fmt.Errorf("运行战斗测试失败: %w", err)
			}

			// 输出测试日志摘要
			logs := svc.GetTestLogs(caseName)
			printFightTestLogSummary(cmd, logs)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server", "", "目标服务器 IP（默认取配置 server_addr）")
	cmd.Flags().IntVar(&serverPort, "port", 0, "目标服务器端口（默认取配置 server_port）")
	cmd.Flags().StringVar(&dir, "dir", "", "用例目录（默认取配置 jsons_dir）")
	cmd.Flags().IntVar(&heroID, "hero", 0, "英雄 ID（>0 时匹配 {id}_*.json / {id}-*.json）")
	cmd.Flags().StringVar(&caseName, "case", "", "仅运行指定用例名（Case 字段）")
	cmd.Flags().StringArrayVar(&files, "file", nil, "直接指定用例文件名（可多次指定，优先级高于 --hero）")
	cmd.Flags().StringVar(&robotPrefix, "prefix", "", "机器人账号前缀（默认取配置 robot_prefix）")
	cmd.Flags().IntVar(&opTimeMs, "op-time", 30000, "操作超时（毫秒）")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "并发数（默认取配置 concurrency）")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "整体超时（单批用例最长执行时间）")
	return cmd
}

// printFightTestLogSummary 输出每个用例的日志摘要（条目数 + 错误级日志）。
func printFightTestLogSummary(cmd *cobra.Command, logs map[string][]LogEntry) {
	if len(logs) == 0 {
		cmd.Println("[fight-test] 完成（无日志输出）")
		return
	}

	totalEntries := 0
	totalErrors := 0
	cmd.Println()
	for caseName, entries := range logs {
		errCount := 0
		for _, e := range entries {
			if e.Level == "error" || e.Level == "Error" {
				errCount++
			}
		}
		totalEntries += len(entries)
		totalErrors += errCount
		mark := "✓"
		if errCount > 0 {
			mark = "✗"
		}
		cmd.Printf("  %s %-28s  日志 %d 条", mark, caseName, len(entries))
		if errCount > 0 {
			cmd.Printf("（含 %d 条错误）", errCount)
		}
		cmd.Println()
	}
	cmd.Printf("\n[fight-test] 完成：共 %d 个用例，%d 条日志", len(logs), totalEntries)
	if totalErrors > 0 {
		cmd.Printf("，%d 条错误", totalErrors)
	}
	cmd.Println()
}
