// Package engine 提供检查功能的入口函数和接口定义
// 本包提供执行所有检查的顶层函数和检查器接口
package engine

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// Checker 列级检查器接口
// 所有列级检查器都必须实现此接口
type Checker interface {
	// Check 执行列级检查
	// 参数:
	//   - sheetName: 表名
	//   - cols: 所有列数据
	//   - colIdx: 当前检查的列索引
	//   - startRowIdx: 数据起始行索引
	//   - params: 规则参数
	//   - sheetMap: 其他表的数据（用于跨表检查）
	// 返回:
	//   - 错误单元格列表
	Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError
}

// TableChecker 表级检查器接口
// 所有表级检查器都必须实现此接口
// 表级检查器用于检查整表的业务规则，可能涉及跨表数据验证
type TableChecker interface {
	// Check 执行表级检查
	// 参数:
	//   - param: 包含所有检查参数的结构体
	// 返回:
	//   - 表检查结果，包含错误信息和详细数据
	Check(param json_rule.CheckParam) *json_rule.TableCheckResult

	// Meta 返回规则元数据（用于前端和 MCP）
	// 返回:
	//   - 规则元数据，包括规则类型、名称、描述、参数定义等
	Meta() *json_rule.TableRuleMeta
}
