package settings

import (
	"context"
	"encoding/json"
	"fmt"

	internal "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	exceltest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/excel-test"
	functiontest "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/function-test"

	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ConfigTools 配置管理工具集合
type ConfigTools struct {
	funcConfigSvc  *functiontest.FuncCaseConfigService
	excelConfigSvc *exceltest.ExcelConfigService
	mcpConfigSvc   *MCPConfigService
}

// NewConfigTools 创建配置管理工具实例
func NewConfigTools(funcConfigSvc *functiontest.FuncCaseConfigService, excelConfigSvc *exceltest.ExcelConfigService, mcpConfigSvc *MCPConfigService) *ConfigTools {
	return &ConfigTools{
		funcConfigSvc:  funcConfigSvc,
		excelConfigSvc: excelConfigSvc,
		mcpConfigSvc:   mcpConfigSvc,
	}
}

// RegisterConfigTools 注册配置管理相关的 MCP Tools
func RegisterConfigTools(s *mcpgo.Server, funcConfigSvc *functiontest.FuncCaseConfigService, excelConfigSvc *exceltest.ExcelConfigService, mcpConfigSvc *MCPConfigService) {
	tools := NewConfigTools(funcConfigSvc, excelConfigSvc, mcpConfigSvc)

	// 注册 get_func_config 工具
	s.AddTool(tools.getFuncConfigTool(), tools.handleGetFuncConfig)

	// 注册 save_func_config 工具
	s.AddTool(tools.saveFuncConfigTool(), tools.handleSaveFuncConfig)

	// 注册 get_excel_config 工具
	s.AddTool(tools.getExcelConfigTool(), tools.handleGetExcelConfig)

	// 注册 save_excel_config 工具
	s.AddTool(tools.saveExcelConfigTool(), tools.handleSaveExcelConfig)

	// ========== 全局设置相关工具 ==========

	// 注册 get_feishu_config 工具
	s.AddTool(tools.getFeishuConfigTool(), tools.handleGetFeishuConfig)

	// 注册 update_feishu_config 工具
	s.AddTool(tools.updateFeishuConfigTool(), tools.handleUpdateFeishuConfig)

	// 注册 get_mcp_config 工具
	s.AddTool(tools.getMCPConfigTool(), tools.handleGetMCPConfig)

	// 注册 save_mcp_config 工具
	s.AddTool(tools.saveMCPConfigTool(), tools.handleSaveMCPConfig)

	// 注册 get_mcp_status 工具
	s.AddTool(tools.getMCPStatusTool(), tools.handleGetMCPStatus)

	// 注册 send_feishu_message 工具
	s.AddTool(tools.sendFeishuMessageTool(), tools.handleSendFeishuMessage)
}

// ========== get_func_config ==========

func (t *ConfigTools) getFuncConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_func_config",
		Description: "获取功能测试配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

func (t *ConfigTools) handleGetFuncConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	config, err := t.funcConfigSvc.GetConfig()
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// ========== save_func_config ==========

func (t *ConfigTools) saveFuncConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "save_func_config",
		Description: "保存功能测试配置",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"type":        "object",
					"description": "配置对象",
				},
			},
			"required": []string{"config"},
		},
	}
}

func (t *ConfigTools) handleSaveFuncConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	// req.Params.Arguments 是 json.RawMessage 类型
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	configArg, ok := args["config"]
	if !ok {
		return mcp.ErrorResult("缺少 config 参数"), nil
	}

	// 将参数转换为 JSON 再转换为 FuncCaseConfig 结构体
	configData, err := json.Marshal(configArg)
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	var config internal.FuncCaseConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	if err := t.funcConfigSvc.SaveConfig(&config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult("功能测试配置保存成功"), nil
}

// ========== get_excel_config ==========

func (t *ConfigTools) getExcelConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_excel_config",
		Description: "获取 Excel 配置",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

func (t *ConfigTools) handleGetExcelConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	config, err := t.excelConfigSvc.GetConfig()
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// ========== save_excel_config ==========

func (t *ConfigTools) saveExcelConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "save_excel_config",
		Description: "保存 Excel 配置",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"config": map[string]any{
					"type":        "object",
					"description": "配置对象",
				},
			},
			"required": []string{"config"},
		},
	}
}

func (t *ConfigTools) handleSaveExcelConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	// req.Params.Arguments 是 json.RawMessage 类型
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	configArg, ok := args["config"]
	if !ok {
		return mcp.ErrorResult("缺少 config 参数"), nil
	}

	// 将参数转换为 JSON 再转换为 ExcelConfig 结构体
	configData, err := json.Marshal(configArg)
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	var config exceltest.ExcelConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	if err := t.excelConfigSvc.SaveConfig(&config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult("Excel 配置保存成功"), nil
}

// ========== 全局设置工具：飞书配置 ==========

// getFeishuConfigTool 获取飞书配置工具定义
func (t *ConfigTools) getFeishuConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_feishu_config",
		Description: "获取飞书通知配置（包括飞书通知开关和机器人GUID）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

// handleGetFeishuConfig 处理获取飞书配置请求
func (t *ConfigTools) handleGetFeishuConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	config, err := t.funcConfigSvc.GetConfig()
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	// 只返回飞书相关配置
	feishuConfig := map[string]any{
		"fei_shu_ntf":  config.FeiShuNtf,
		"fei_shu_guid": config.FeiShuGUID,
	}

	data, err := json.MarshalIndent(feishuConfig, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// updateFeishuConfigTool 更新飞书配置工具定义
func (t *ConfigTools) updateFeishuConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "update_feishu_config",
		Description: "更新飞书通知配置（部分更新，不会覆盖其他配置项）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"fei_shu_ntf": map[string]any{
					"type":        "boolean",
					"description": "是否开启飞书通知",
				},
				"fei_shu_guid": map[string]any{
					"type":        "string",
					"description": "飞书机器人GUID",
				},
			},
		},
	}
}

// handleUpdateFeishuConfig 处理更新飞书配置请求
func (t *ConfigTools) handleUpdateFeishuConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	// 构建部分更新参数
	updates := make(map[string]interface{})
	if val, ok := args["fei_shu_ntf"]; ok {
		updates["fei_shu_ntf"] = val
	}
	if val, ok := args["fei_shu_guid"]; ok {
		updates["fei_shu_guid"] = val
	}

	if len(updates) == 0 {
		return mcp.ErrorResult("没有提供要更新的配置项"), nil
	}

	// 使用部分更新方法
	updatedConfig, err := t.funcConfigSvc.UpdateConfig(updates)
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	// 返回更新后的飞书配置
	result := map[string]any{
		"fei_shu_ntf":  updatedConfig.FeiShuNtf,
		"fei_shu_guid": updatedConfig.FeiShuGUID,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult("飞书配置更新成功:\n" + string(data)), nil
}

// ========== 全局设置工具：MCP 配置 ==========

// getMCPConfigTool 获取 MCP 配置工具定义
func (t *ConfigTools) getMCPConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_mcp_config",
		Description: "获取 MCP 服务配置（包括启用状态、端口、绑定地址）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

// handleGetMCPConfig 处理获取 MCP 配置请求
func (t *ConfigTools) handleGetMCPConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	config, err := t.mcpConfigSvc.GetConfig()
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// saveMCPConfigTool 保存 MCP 配置工具定义
func (t *ConfigTools) saveMCPConfigTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "save_mcp_config",
		Description: "保存 MCP 服务配置（修改后服务会自动重启）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"enabled": map[string]any{
					"type":        "boolean",
					"description": "是否启用 MCP 服务",
				},
				"port": map[string]any{
					"type":        "integer",
					"description": "MCP 服务端口 (1-65535)",
					"minimum":     1,
					"maximum":     65535,
				},
				"host": map[string]any{
					"type":        "string",
					"description": "MCP 服务绑定地址",
				},
			},
			"required": []string{"enabled", "port", "host"},
		},
	}
}

// handleSaveMCPConfig 处理保存 MCP 配置请求
func (t *ConfigTools) handleSaveMCPConfig(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	// 解析参数
	enabled, ok := args["enabled"].(bool)
	if !ok {
		return mcp.ErrorResult("缺少或无效的 enabled 参数"), nil
	}

	port, ok := args["port"].(float64)
	if !ok {
		return mcp.ErrorResult("缺少或无效的 port 参数"), nil
	}

	host, ok := args["host"].(string)
	if !ok {
		return mcp.ErrorResult("缺少或无效的 host 参数"), nil
	}

	// 保存配置（会自动重启服务）
	config := &MCPConfig{
		Enabled: enabled,
		Port:    int(port),
		Host:    host,
	}

	if err := t.mcpConfigSvc.SaveConfig(config); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	result := map[string]any{
		"enabled": enabled,
		"port":    int(port),
		"host":    host,
		"address": config.Address(),
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult("MCP 配置保存成功（服务已自动重启）:\n" + string(data)), nil
}

// getMCPStatusTool 获取 MCP 状态工具定义
func (t *ConfigTools) getMCPStatusTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_mcp_status",
		Description: "获取 MCP 服务运行状态（包括运行状态、地址、连接数等）",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}

// handleGetMCPStatus 处理获取 MCP 状态请求
func (t *ConfigTools) handleGetMCPStatus(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	status := t.mcpConfigSvc.GetMCPStatus()

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	return mcp.TextResult(string(data)), nil
}

// ========== 全局设置工具：发送飞书消息 ==========

// sendFeishuMessageTool 发送飞书消息工具定义
func (t *ConfigTools) sendFeishuMessageTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "send_feishu_message",
		Description: "发送飞书消息（文本）到指定机器人，用于测试飞书配置或手动发送通知",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "要发送的消息内容（纯文本）",
				},
				"robot_guid": map[string]any{
					"type":        "string",
					"description": "飞书机器人GUID（可选，不填则使用全局配置中的GUID）",
				},
			},
			"required": []string{"message"},
		},
	}
}

// handleSendFeishuMessage 处理发送飞书消息请求
func (t *ConfigTools) handleSendFeishuMessage(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return mcp.ErrorResultFromError(err), nil
	}

	// 获取消息内容
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return mcp.ErrorResult("缺少 message 参数或消息为空"), nil
	}

	// 获取机器人GUID（优先使用参数，否则使用全局配置）
	robotGUID, _ := args["robot_guid"].(string)
	if robotGUID == "" {
		config, err := t.funcConfigSvc.GetConfig()
		if err != nil {
			return mcp.ErrorResult(fmt.Sprintf("获取全局配置失败: %v", err)), nil
		}
		robotGUID = config.FeiShuGUID
		if robotGUID == "" {
			return mcp.ErrorResult("未配置飞书机器人GUID，请先在全局设置中配置或通过 robot_guid 参数传入"), nil
		}
	}

	// 发送消息（使用 %s 格式化避免 go vet 警告)
	feishu.SendFeiShuRobotText(robotGUID, "%s", message)

	return mcp.TextResult(fmt.Sprintf("消息已发送到飞书机器人: %s", robotGUID)), nil
}
