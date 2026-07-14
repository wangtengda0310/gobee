package exceltest

import (
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterExcelCheckTools 注册 Excel 检查相关的 MCP Tools
// @mcp
func RegisterExcelCheckTools(s *mcpgo.Server, svc *ExcelCheckService) {
	// TODO: 重新注册 MCP 工具
}

// RegisterExcelConfigTools 注册 Excel 配置相关的 MCP Tools
// @mcp
func RegisterExcelConfigTools(s *mcpgo.Server, svc *ExcelConfigService) {
	// TODO: 重新注册 MCP 工具
}

// RegisterGameExcelTools 注册游戏数据相关的 MCP Tools（仅 Excel 测试页面使用）
// 注意：这个函数只注册 get_all_hero_cfg 工具，其他游戏数据工具由 mcp/common.RegisterGameExcelTools 注册
// @mcp
func RegisterGameExcelTools(s *mcpgo.Server, svc *ExcelTestGameService) {
	// TODO: 重新注册 MCP 工具
}
