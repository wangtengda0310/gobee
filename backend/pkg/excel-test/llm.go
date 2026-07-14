package exceltest

import (
	"context"
	"encoding/json"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

var (
	llmRegistry *tool.Registry
	llmSvc      *ExcelCheckService
)

// InitLLMTools 初始化 LLM 工具注册（由 ChatService 调用）
func InitLLMTools(registry *tool.Registry, svc *ExcelCheckService) {
	llmRegistry = registry
	llmSvc = svc

	if registry == nil || svc == nil {
		return
	}

	// ========== 规则管理 ==========
	registry.MustRegister(
		tool.NewFunction("get_all_excel_rules", "获取指定目录下的 Excel 检查规则",
			func(ctx context.Context, args map[string]any) (any, error) {
				dir, _ := args["dir"].(string)
				return svc.GetAllExcelRules(dir)
			},
			tool.WithStringParam("dir", "规则目录路径", true),
		),
		tool.NewFunction("save_all_excel_rules", "保存 Excel 检查规则到指定目录",
			func(ctx context.Context, args map[string]any) (any, error) {
				dir, _ := args["dir"].(string)
				rules, _ := args["rules"].(string)
				var sheetRules []*json_rule.SheetRule
				if err := json.Unmarshal([]byte(rules), &sheetRules); err != nil {
					return nil, err
				}
				return "保存成功", svc.SaveAllExcelRules(dir, sheetRules)
			},
			tool.WithStringParam("dir", "规则目录路径", true),
			tool.WithStringParam("rules", "Excel 检查规则 JSON 数组", true),
		),
		tool.NewFunction("check_all_excel_rules", "根据规则执行 Excel 全量检查",
			func(ctx context.Context, args map[string]any) (any, error) {
				dir, _ := args["dir"].(string)
				rules, _ := args["rules"].(string)
				var sheetRules []*json_rule.SheetRule
				if err := json.Unmarshal([]byte(rules), &sheetRules); err != nil {
					return nil, err
				}
				return svc.CheckAllExcelRules(dir, sheetRules)
			},
			tool.WithStringParam("dir", "Excel 文件目录路径", true),
			tool.WithStringParam("rules", "Excel 检查规则 JSON 数组", true),
		),
		tool.NewFunction("check_incremental", "增量检查 Excel 配表（基于 git diff 过滤变更文件）",
			func(ctx context.Context, args map[string]any) (any, error) {
				dir, _ := args["dir"].(string)
				rules, _ := args["rules"].(string)
				var sheetRules []*json_rule.SheetRule
				if err := json.Unmarshal([]byte(rules), &sheetRules); err != nil {
					return nil, err
				}
				return svc.CheckIncremental(dir, sheetRules)
			},
			tool.WithStringParam("dir", "Excel 文件目录路径", true),
			tool.WithStringParam("rules", "Excel 检查规则 JSON 数组", true),
		),
	)

	// ========== Excel 文件操作 ==========
	registry.MustRegister(
		tool.NewFunction("get_all_excels", "获取指定目录下所有 Excel 文件信息",
			func(ctx context.Context, args map[string]any) (any, error) {
				dirPath, _ := args["dirPath"].(string)
				return svc.GetAllExcels(dirPath)
			},
			tool.WithStringParam("dirPath", "Excel 文件目录路径", true),
		),
		tool.NewFunction("get_excel_sheets", "获取单个 Excel 文件中的所有 Sheet 名称",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				return svc.GetExcelSheets(filePath)
			},
			tool.WithStringParam("filePath", "Excel 文件路径", true),
		),
		tool.NewFunction("preview_excel_sheet", "预览 Excel Sheet 数据，返回表头和前 N 行数据",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				sheetName, _ := args["sheetName"].(string)
				rows := 10
				if v, ok := args["rows"].(float64); ok {
					rows = int(v)
				}
				return svc.PreviewExcelSheet(filePath, sheetName, rows)
			},
			tool.WithStringParam("filePath", "Excel 文件路径", true),
			tool.WithStringParam("sheetName", "Sheet 名称", true),
			tool.WithNumberParam("rows", "预览行数（默认 10）", false),
		),
		tool.NewFunction("get_excel_column_info", "获取 Excel Sheet 的列详细信息（中文名、字段名、类型、列状态）",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				sheetName, _ := args["sheetName"].(string)
				return svc.GetExcelColumnInfo(filePath, sheetName)
			},
			tool.WithStringParam("filePath", "Excel 文件路径", true),
			tool.WithStringParam("sheetName", "Sheet 名称", true),
		),
		tool.NewFunction("create_excel_file", "创建符合项目规范的 Excel 文件（4行表头：中文名/类型/字段名/导出标识）",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				sheets, _ := args["sheets"].(string)
				var sheetDefs []*ExcelSheetDef
				if err := json.Unmarshal([]byte(sheets), &sheetDefs); err != nil {
					return nil, err
				}
				return svc.CreateExcelFile(filePath, sheetDefs)
			},
			tool.WithStringParam("filePath", "输出文件路径", true),
			tool.WithStringParam("sheets", "Sheet 定义 JSON 数组", true),
		),
		tool.NewFunction("filter_excel_data", "根据指定列的条件过滤 Excel 数据",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				sheetName, _ := args["sheetName"].(string)
				includeHeader, _ := args["includeHeader"].(bool)
				conditions, _ := args["conditions"].(string)
				var conds []*FilterCondition
				if err := json.Unmarshal([]byte(conditions), &conds); err != nil {
					return nil, err
				}
				return svc.FilterExcelData(filePath, sheetName, conds, includeHeader)
			},
			tool.WithStringParam("filePath", "Excel 文件路径", true),
			tool.WithStringParam("sheetName", "Sheet 名称", true),
			tool.WithStringParam("conditions", "过滤条件 JSON 数组（columnName/value/operator）", true),
			tool.WithBooleanParam("includeHeader", "是否包含表头行", false),
		),
		tool.NewFunction("query_excel_range", "查询 Excel 指定行范围的数据",
			func(ctx context.Context, args map[string]any) (any, error) {
				filePath, _ := args["filePath"].(string)
				sheetName, _ := args["sheetName"].(string)
				includeHeader, _ := args["includeHeader"].(bool)
				startRow := 1
				endRow := -1
				if v, ok := args["startRow"].(float64); ok {
					startRow = int(v)
				}
				if v, ok := args["endRow"].(float64); ok {
					endRow = int(v)
				}
				return svc.QueryExcelRange(filePath, sheetName, startRow, endRow, includeHeader)
			},
			tool.WithStringParam("filePath", "Excel 文件路径", true),
			tool.WithStringParam("sheetName", "Sheet 名称", true),
			tool.WithNumberParam("startRow", "起始行号（数据行从1开始）", true),
			tool.WithNumberParam("endRow", "结束行号（-1 表示到末尾）", false),
			tool.WithBooleanParam("includeHeader", "是否包含表头行", false),
		),
	)

	// ========== 表变更检测 ==========
	registry.MustRegister(
		tool.NewFunction("get_table_changes", "获取 Excel 表的变更信息（基于 git diff）",
			func(ctx context.Context, args map[string]any) (any, error) {
				excelPath, _ := args["excelPath"].(string)
				sheetName, _ := args["sheetName"].(string)
				baseCommit, _ := args["baseCommit"].(string)
				idColName, _ := args["idColName"].(string)
				nameColName, _ := args["nameColName"].(string)
				return svc.GetTableChanges(excelPath, sheetName, baseCommit, idColName, nameColName)
			},
			tool.WithStringParam("excelPath", "Excel 文件绝对路径", true),
			tool.WithStringParam("sheetName", "Sheet 名称", true),
			tool.WithStringParam("baseCommit", "基准 commit（默认 HEAD~1）", false),
			tool.WithStringParam("idColName", "ID 列名（默认 Id）", false),
			tool.WithStringParam("nameColName", "名称列名（默认 Name）", false),
		),
	)
}
