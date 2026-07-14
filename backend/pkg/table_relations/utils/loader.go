package utils

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"github.com/xuri/excelize/v2"
)

// LoadSheetData 加载指定表的列数据
//
// 参数：
//   - sheetMap: 表映射表，key为表名，value为Excel文件对象
//   - sheetName: 要加载的表名（如 "Activity"、"DrawSkin"）
//
// 返回：
//   - [][]string: 表的列数据（按列组织的二维数组）
//   - string: 实际的表名（可能包含中文部分）
//   - error: 错误信息
//
// 示例：
//
//	cols, actualName, err := LoadSheetData(sheetMap, "Activity")
//	if err != nil {
//	    return fmt.Errorf("加载活动表失败: %w", err)
//	}
func LoadSheetData(sheetMap map[string]*excelize.File, sheetName string) ([][]string, string, error) {
	// 查找 Sheet（支持精确匹配和后缀匹配）
	file, actualName, found := helpers.FindSheetBySuffix(sheetMap, sheetName)
	if !found {
		return nil, "", fmt.Errorf("表 '%s' 不存在", sheetName)
	}

	// 获取列数据
	cols, err := file.GetCols(actualName)
	if err != nil {
		return nil, "", fmt.Errorf("获取表 '%s' 列数据失败: %w", actualName, err)
	}

	return cols, actualName, nil
}

// GetStartRowIndex 获取数据起始行索引
//
// 名将杀 MJS 格式的表有固定行数的表头：
//   - 第0行：中文名称行
//   - 第1行：类型定义行
//   - 第2行：属性名称行
//   - 第3行：导出标识行
//   - 第4行开始：实际数据
//
// 返回：
//   - int: 数据起始行索引（固定为 4）
//
// 示例：
//
//	startRow := GetStartRowIndex()
//	for i, id := range idCol[startRow:] {
//	    // 处理数据行
//	}
func GetStartRowIndex() int {
	return excelio.MJS_FIXED_ROWS_NUM
}

// GetDataEndIndex 获取数据结束行索引
//
// 使用 helpers.AutoDetectEndIndex 检测数据结束位置
// 连续 3 行空单元格视为数据结束（避免读取到注释区）
//
// 参数：
//   - cols: 表的列数据
//   - idColIdx: ID 列的索引（用于判断数据是否结束）
//   - startRow: 数据起始行索引
//
// 返回：
//   - int: 数据结束行索引（相对于 cols 的索引）
//
// 示例：
//
//	startRow := GetStartRowIndex()
//	endRow := GetDataEndIndex(cols, idColIdx, startRow)
//	for i := startRow; i < endRow; i++ {
//	    // 处理数据行
//	}
func GetDataEndIndex(cols [][]string, idColIdx, startRow int) int {
	const breakLine = 3 // 连续3行空单元格视为数据结束
	return helpers.AutoDetectEndIndex(cols, idColIdx, startRow, breakLine)
}
