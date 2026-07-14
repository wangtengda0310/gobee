package mcp

import (
	activitywikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/activity-wiki-check"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/robotext"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"
	herowikicheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/hero-wiki-check"
	resourcechecker "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/settings"
)

// McpServices 统一的 MCP 服务集合。
// stdio 与 HTTP 两种传输模式共享同一套服务实例，传输层的差异保留在 server 类型上
// （MCPServer 用于 HTTP，StdioMCPServer 用于 stdio）。
type McpServices struct {
	McpJsonCaseService           *functiontest.JsonCaseService
	McpFuncCaseConfigService     *functiontest.FuncCaseConfigService
	McpExcelCheckService         *exceltest.ExcelCheckService
	McpExcelConfigService        *exceltest.ExcelConfigService
	McpGameExcelService          *game.GameExcelService
	McpExcelTestGameService      *exceltest.ExcelTestGameService
	McpHeroResCheckService       *resourcechecker.HeroResCheckService
	McpHeroWikiResCheckService   *herowikicheck.HeroWikiResCheckService
	McpActivityWikiCheckService  *activitywikicheck.ActivityWikiCheckService
	McpConfigService             *settings.MCPConfigService
	McpFeishuNotifyConfigService *settings.FeishuNotifyConfigService
	McpRobotExtService           *robotext.RobotExtService
}

// buildMcpServiceDeps 创建所有 MCP 服务实际依赖的私有 helper。
// 两个传输模式使用相同的底层 Service 实例创建逻辑，避免重复代码。
// 当前所有服务初始化均不会失败，保留 err 以便未来扩展。
func buildMcpServiceDeps() (*McpServices, error) {
	funcConfigService := functiontest.NewFuncCaseConfigService()
	jsonCaseService := functiontest.NewJsonCaseService(nil)
	excelCheckService := exceltest.NewExcelCheckServiceWithConfig(funcConfigService)
	excelConfigService := exceltest.NewExcelConfigService()
	gameExcelService := game.NewGameExcelService()
	if funcConfig, cfgErr := funcConfigService.GetConfig(); cfgErr == nil && funcConfig != nil && funcConfig.ExcelResourcesDir != "" {
		_ = gameExcelService.InitExcel(funcConfig.ExcelResourcesDir)
	}
	excelTestGameService := exceltest.NewExcelTestGameService()
	heroResCheckService := resourcechecker.NewHeroResCheckService()
	heroWikiResCheckService := herowikicheck.NewHeroWikiResCheckService()
	activityWikiCheckService := activitywikicheck.NewActivityWikiCheckService()
	mcpConfigService := settings.NewMCPConfigService()
	feishuNotifyConfigService := settings.NewFeishuNotifyConfigService()
	robotExtService := robotext.NewRobotExtService()
	return &McpServices{
		McpJsonCaseService:           jsonCaseService,
		McpFuncCaseConfigService:     funcConfigService,
		McpExcelCheckService:         excelCheckService,
		McpExcelConfigService:        excelConfigService,
		McpGameExcelService:          gameExcelService,
		McpExcelTestGameService:      excelTestGameService,
		McpHeroResCheckService:       heroResCheckService,
		McpHeroWikiResCheckService:   heroWikiResCheckService,
		McpActivityWikiCheckService:  activityWikiCheckService,
		McpConfigService:             mcpConfigService,
		McpFeishuNotifyConfigService: feishuNotifyConfigService,
		McpRobotExtService:           robotExtService,
	}, nil
}

// BuildDefaultMcpServices 创建默认的 MCP 服务集合，stdio 与 HTTP 模式共用。
func BuildDefaultMcpServices() (*McpServices, error) {
	return buildMcpServiceDeps()
}
