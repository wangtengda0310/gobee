package herowikicheck

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterWikiCheckTools 注册 Wiki 检查相关的 MCP Tools
func RegisterWikiCheckTools(s *mcpgo.Server, svc *HeroWikiResCheckService) {
	// 注册 check_hero_wiki 工具
	s.AddTool(&mcpgo.Tool{
		Name: "check_hero_wiki",
		Description: "执行武将Wiki检查，对比新旧数据返回差异结果。" +
			"返回包含新增武将、删除武将、修改武将的详细信息，以及每个武将的变化字段数量。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"excelDir": map[string]any{
					"type":        "string",
					"description": "Excel 配置目录路径（如 D:/work/config/excel）",
				},
				"oldJsonPath": map[string]any{
					"type":        "string",
					"description": "历史数据 JSON 文件路径，用于对比差异（可选）",
				},
			},
			"required": []string{"excelDir"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			ExcelDir    string `json:"excelDir"`
			OldJsonPath string `json:"oldJsonPath"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		result, err := svc.Check(args.ExcelDir, args.OldJsonPath)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(data)), nil
	})

	// 注册 save_hero_wiki 工具
	s.AddTool(&mcpgo.Tool{
		Name:        "save_hero_wiki",
		Description: "保存武将Wiki检查结果到指定路径。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"savePath": map[string]any{
					"type":        "string",
					"description": "保存路径（如 tmp/test.json）",
				},
				"data": map[string]any{
					"type":        "object",
					"description": "要保存的数据（DataContainer 结构）",
				},
			},
			"required": []string{"savePath", "data"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			SavePath string          `json:"savePath"`
			Data     json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		var container diff.DataContainer
		if err := json.Unmarshal(args.Data, &container); err != nil {
			return mcp.ErrorResult("解析数据失败: " + err.Error()), nil
		}

		if err := svc.Save(args.SavePath, &container); err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult("保存成功: " + args.SavePath), nil
	})
}
