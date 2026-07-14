package activitywikicheck

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/mcp"
	mcpgo "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterActivityWikiCheckTools 注册活动 Wiki 检查相关的 MCP Tools
func RegisterActivityWikiCheckTools(s *mcpgo.Server, svc *ActivityWikiCheckService) {
	// 注册 check_activity_wiki 工具
	s.AddTool(&mcpgo.Tool{
		Name: "check_activity_wiki",
		Description: "执行活动Wiki检查，返回活动配置数据。" +
			"返回包含活动基础信息、抽奖配置、掉落规则、次数奖励等完整数据。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"excelDir": map[string]any{
					"type":        "string",
					"description": "Excel 配置目录路径（如 D:/work/config/excel）",
				},
			},
			"required": []string{"excelDir"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			ExcelDir string `json:"excelDir"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		result, err := svc.Check(args.ExcelDir)
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult(string(data)), nil
	})

	// 注册 open_excel 工具
	s.AddTool(&mcpgo.Tool{
		Name: "open_excel",
		Description: "打开指定 Sheet 对应的 Excel 文件" +
			"根据操作系统自动选择打开方式（Windows/macOS/Linux）。",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"sheetName": map[string]any{
					"type":        "string",
					"description": "Sheet 名称（如\"活动表|Activity\"）",
				},
				"excelDir": map[string]any{
					"type":        "string",
					"description": "Excel 配置目录路径（如 D:/work/config/excel）",
				},
			},
			"required": []string{"sheetName", "excelDir"},
		},
	}, func(ctx context.Context, req *mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		var args struct {
			SheetName string `json:"sheetName"`
			ExcelDir  string `json:"excelDir"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return mcp.ErrorResult("解析参数失败: " + err.Error()), nil
		}

		if err := svc.OpenExcel(args.SheetName, args.ExcelDir); err != nil {
			return mcp.ErrorResultFromError(err), nil
		}

		return mcp.TextResult("Excel 文件已打开: " + args.SheetName), nil
	})
}
