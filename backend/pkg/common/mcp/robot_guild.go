package mcp

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/robotext"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RobotGuildTools robot 工会操作工具集合
type RobotGuildTools struct {
	svc *robotext.RobotExtService
}

// NewRobotGuildTools 创建 robot 工会操作工具实例
func NewRobotGuildTools(svc *robotext.RobotExtService) *RobotGuildTools {
	return &RobotGuildTools{
		svc: svc,
	}
}

// RegisterRobotGuildTools 注册 robot 工会操作相关的 MCP Tools
func RegisterRobotGuildTools(s *mcpgo.Server, svc *robotext.RobotExtService) {
	tools := NewRobotGuildTools(svc)

	// 注册 get_guild_leaders 工具
	s.AddTool(tools.getGuildLeadersTool(), tools.handleGetGuildLeaders)

	// 注册 create_guild_with_city 工具
	s.AddTool(tools.createGuildWithCityTool(), tools.handleCreateGuildWithCity)

	// 注册 upgrade_guild_city 工具
	s.AddTool(tools.upgradeGuildCityTool(), tools.handleUpgradeGuildCity)
}

// ========== get_guild_leaders ==========

func (t *RobotGuildTools) getGuildLeadersTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "get_guild_leaders",
		Description: "查询服务器工会列表及会长账号信息，返回可用于登录的会长账号列表",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"serverIp", "serverPort"},
		},
	}
}

func (t *RobotGuildTools) handleGetGuildLeaders(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args struct {
		ServerIp   string `json:"serverIp"`
		ServerPort string `json:"serverPort"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return robotErrorResult("解析参数失败: " + err.Error()), nil
	}

	result, err := t.svc.GetGuildLeaders(args.ServerIp, args.ServerPort)
	if err != nil {
		return ErrorResultFromError(err), nil
	}

	return TextResult(result), nil
}

// ========== create_guild_with_city ==========

func (t *RobotGuildTools) createGuildWithCityTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "create_guild_with_city",
		Description: "创建工会并设置野战城池。使用机器人账号创建新工会，并创建指定数量的城池。机器人账号格式为 {prefix}x{process}x{idx}",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"prefix": map[string]any{
					"type":        "string",
					"description": "机器人账号前缀",
				},
				"process": map[string]any{
					"type":        "integer",
					"description": "机器人进程编号（默认为 1）",
				},
				"idx": map[string]any{
					"type":        "integer",
					"description": "机器人序号（默认为 1）",
				},
				"country": map[string]any{
					"type":        "integer",
					"description": "势力ID（1-15）：1=秦 2=西楚 3=西汉 4=东汉 5=黄 6=曹魏 7=蜀 8=孙吴 9=张楚 10=魏 11=楚 12=汉 13=燕 14=齐 15=赵",
				},
				"cityCount": map[string]any{
					"type":        "integer",
					"description": "要创建的城池数量（最多 8 个）",
				},
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"prefix", "country", "cityCount", "serverIp", "serverPort"},
		},
	}
}

func (t *RobotGuildTools) handleCreateGuildWithCity(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args struct {
		Prefix     string `json:"prefix"`
		Process    int    `json:"process"`
		Idx        int    `json:"idx"`
		Country    int32  `json:"country"`
		CityCount  int    `json:"cityCount"`
		ServerIp   string `json:"serverIp"`
		ServerPort string `json:"serverPort"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return robotErrorResult("解析参数失败: " + err.Error()), nil
	}

	// 设置默认值
	if args.Process == 0 {
		args.Process = 1
	}
	if args.Idx == 0 {
		args.Idx = 1
	}

	result, err := t.svc.CreateGuildWithCity(args.Prefix, args.Process, args.Idx, args.Country, args.CityCount, args.ServerIp, args.ServerPort)
	if err != nil {
		return ErrorResultFromError(err), nil
	}

	return TextResult(result), nil
}

// ========== upgrade_guild_city ==========

func (t *RobotGuildTools) upgradeGuildCityTool() *mcpgo.Tool {
	return &mcpgo.Tool{
		Name:        "upgrade_guild_city",
		Description: "升级工会城池等级。使用会长账号登录，循环升级城池直到目标等级。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"accountId": map[string]any{
					"type":        "string",
					"description": "会长账号（从 get_guild_leaders 获取）",
				},
				"cityId": map[string]any{
					"type":        "integer",
					"description": "城池ID",
				},
				"targetLevel": map[string]any{
					"type":        "integer",
					"description": "目标等级",
				},
				"serverIp": map[string]any{
					"type":        "string",
					"description": "服务器 IP 地址",
				},
				"serverPort": map[string]any{
					"type":        "string",
					"description": "服务器端口",
				},
			},
			"required": []string{"accountId", "cityId", "targetLevel", "serverIp", "serverPort"},
		},
	}
}

func (t *RobotGuildTools) handleUpgradeGuildCity(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	var args struct {
		AccountId   string `json:"accountId"`
		CityId      uint64 `json:"cityId"`
		TargetLevel int32  `json:"targetLevel"`
		ServerIp    string `json:"serverIp"`
		ServerPort  string `json:"serverPort"`
	}
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return robotErrorResult("解析参数失败: " + err.Error()), nil
	}

	result, err := t.svc.UpgradeGuildCity(args.AccountId, args.CityId, args.TargetLevel, args.ServerIp, args.ServerPort)
	if err != nil {
		return ErrorResultFromError(err), nil
	}

	return TextResult(result), nil
}

// ========== 辅助函数 ==========

// robotErrorResult 创建字符串错误结果
func robotErrorResult(msg string) *mcpgo.CallToolResult {
	return &mcpgo.CallToolResult{
		IsError: true,
		Content: []mcpgo.Content{
			&mcpgo.TextContent{Text: "错误: " + msg},
		},
	}
}
