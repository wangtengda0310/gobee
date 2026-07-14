package mcp

import (
	"fmt"
	"path/filepath"

	activitywikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check"
	mcputil "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	herowikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerAllMcpTools 向给定的 MCP server 实例注册所有业务 tools。
// 该函数被 HTTP 和 stdio 两种传输模式共享，避免注册逻辑重复。
func registerAllMcpTools(server *mcpgo.Server, services *McpServices) {
	// 注册功能测试相关 tools
	functiontest.RegisterJsonCaseTools(server, services.McpJsonCaseService)

	// 注册配置管理 tools
	settings.RegisterConfigTools(server, services.McpFuncCaseConfigService, services.McpExcelConfigService, services.McpConfigService)

	// 注册 Excel 检查 tools
	exceltest.RegisterExcelCheckTools(server, services.McpExcelCheckService)

	// 注册 Excel 配置 tools
	exceltest.RegisterExcelConfigTools(server, services.McpExcelConfigService)

	// 注册 Wiki 检查 tools
	herowikicheck.RegisterWikiCheckTools(server, services.McpHeroWikiResCheckService)

	// 注册活动 Wiki 检查 tools
	activitywikicheck.RegisterActivityWikiCheckTools(server, services.McpActivityWikiCheckService)

	// 注册游戏数据 tools（Excel 测试页面专用）
	exceltest.RegisterGameExcelTools(server, services.McpExcelTestGameService)

	// 注册通用游戏数据 tools（完整版，包含 hero/card/skill/msg/error/property）
	mcputil.RegisterGameExcelTools(server, services.McpGameExcelService)

	// 注册 Robot 扩展工具（工会操作）
	mcputil.RegisterRobotGuildTools(server, services.McpRobotExtService)

	// 注册战斗测试 tools
	// 传递飞书自动通知配置服务
	functiontest.RegisterFightCaseTools(server, "", buildFightTestRunner(services), services.McpFeishuNotifyConfigService)
}

// buildFightTestRunner 构建战斗测试 runner 闭包。
//
// 参数说明：
//   - heroID: 英雄ID。当 heroID > 0 时，根据英雄ID查找对应的用例文件（支持下划线和连字符格式）；
//     当 heroID <= 0 时，不限制英雄，依赖 filterCases 或 caseName 过滤。
//   - caseName: 用例名称。当不为空字符串时，仅运行匹配该名称的用例，优先级低于 filterCases。
//   - filterCases: 指定运行的用例文件名列表。当非空时，直接运行这些文件，忽略 heroID 和 caseName。
//
// 返回值：
//   - map[string][]functiontest.LogEntry: 键为用例名称（caseName），值为该用例的日志条目列表。
//
// 内部实现：
//   - 通过 services.McpFuncCaseConfigService.GetConfig() 动态读取战斗用例目录路径、服务器地址等配置。
func buildFightTestRunner(services *McpServices) func(heroID int, caseName string, filterCases []string) (map[string][]functiontest.LogEntry, error) {
	return func(heroID int, caseName string, filterCases []string) (map[string][]functiontest.LogEntry, error) {
		// 从配置动态读取路径
		config, err := services.McpFuncCaseConfigService.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("获取配置失败: %w", err)
		}
		fightCasesDir := config.JsonsDir

		// 构建过滤用例文件名
		var caseFiles []string
		var filterCaseNames []string

		// 如果指定了用例名称，设置过滤
		if caseName != "" {
			filterCaseNames = []string{caseName}
		}

		if len(filterCases) > 0 {
			caseFiles = filterCases
		} else if heroID > 0 {
			// 根据英雄ID查找对应文件，同时支持下划线和连字符格式
			// 格式1: 10103_关羽.json
			pattern1 := filepath.Join(fightCasesDir, fmt.Sprintf("%d_*.json", heroID))
			matches1, _ := filepath.Glob(pattern1)
			// 格式2: 10112-魏延.json
			pattern2 := filepath.Join(fightCasesDir, fmt.Sprintf("%d-*.json", heroID))
			matches2, _ := filepath.Glob(pattern2)

			caseFiles = append(caseFiles, matches1...)
			caseFiles = append(caseFiles, matches2...)
			for i, match := range caseFiles {
				caseFiles[i] = filepath.Base(match)
			}
		}

		// 执行测试
		err = services.McpJsonCaseService.RunRobotTest(
			config.ServerAddr,                    // serverAddr
			fmt.Sprintf("%d", config.ServerPort), // serverPort
			"mcp-fight",                          // clientName
			fmt.Sprintf("战斗测试-英雄%d", heroID),     // robotName
			"",                       // feishuGuid
			5,                        // loginTime
			caseFiles,                // caseFiles
			filterCaseNames,          // filterCaseNames 传递用例名称过滤
			&fightCasesDir,           // fightCasesDir
			nil,                      // filesData
			30000,                    // opTimeMs
			false,                    // feishuNtf
			false,                    // debugLevel
			false,                    // debugLog
			uint(config.Concurrency), // concurrency — 复用前端设置页配置，不再写死
			0,                        // maxTimeoutPerCase: 0=默认 10 分钟
		)
		if err != nil {
			return nil, err
		}

		// 获取测试日志
		return services.McpJsonCaseService.GetTestLogs(caseName), nil
	}
}
