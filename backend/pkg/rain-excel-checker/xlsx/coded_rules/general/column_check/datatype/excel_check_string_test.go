// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestStringCheckRule 测试字符串类型检查规则
// 验证数值型单元格被错误地配在字符串列时能够正确检测
func TestStringCheckRule(t *testing.T) {
	// 创建测试用的 Excel 文件
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 创建测试数据：包含数值型和文本型单元格
	// A1-A4 为表头
	// A5: 文本型 "abc" (CellTypeSharedString) — 正确
	// A6: 数值型 123 (CellTypeUnset 或 CellTypeNumber) — 错误
	// A7: 数值型 456 (CellTypeUnset 或 CellTypeNumber) — 错误

	// 设置表头
	f.SetCellValue(sheetName, "A1", "")
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "A3", "Id")
	f.SetCellValue(sheetName, "A4", "")

	// 设置文本型单元格（正确）
	f.SetCellValue(sheetName, "A5", "abc")

	// 设置数值型单元格（错误，会被检测出来）
	f.SetCellValue(sheetName, "A6", 123)
	f.SetCellValue(sheetName, "A7", 456)

	// 获取列数据
	cols, err := f.GetCols(sheetName)
	if err != nil {
		t.Fatalf("获取列数据失败: %v", err)
	}

	// 构造 sheetMap
	sheetMap := map[string]*excelize.File{
		sheetName: f,
	}

	// 创建检查器并执行检查
	rule := new(StringCheckRule)
	params := map[string]string{
		"allowEmpty":  "true",
		"allowCommit": "true",
		"breakLine":   "3",
	}

	// 列索引 0，数据起始行索引 4
	res := rule.Check(sheetName, cols, 0, 4, params, sheetMap)

	// 验证结果：应该检测到 A6 和 A7 的类型错误
	assert.Len(t, res, 2, "应检测到 2 个数值格式错误")

	// 验证错误信息包含预期内容
	for _, errCell := range res {
		assert.Contains(t, errCell.Reason, "数值格式", "错误信息应包含'数值格式'")
		assert.Contains(t, errCell.Reason, "应为字符串", "错误信息应包含'应为字符串'")
	}
}

// TestStringCheckRule_AllString 测试全部为文本型的情况
// 验证当所有单元格都是文本型时，不应该有错误
func TestStringCheckRule_AllString(t *testing.T) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 设置表头
	f.SetCellValue(sheetName, "A1", "")
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "A3", "Name")
	f.SetCellValue(sheetName, "A4", "")

	// 设置文本型单元格（全部正确）
	f.SetCellValue(sheetName, "A5", "value1")
	f.SetCellValue(sheetName, "A6", "value2")
	f.SetCellValue(sheetName, "A7", "value3")

	cols, err := f.GetCols(sheetName)
	if err != nil {
		t.Fatalf("获取列数据失败: %v", err)
	}

	sheetMap := map[string]*excelize.File{
		sheetName: f,
	}

	rule := new(StringCheckRule)
	res := rule.Check(sheetName, cols, 0, 4, nil, sheetMap)

	assert.Empty(t, res, "全部为文本型时不应报错")
}

// TestStringCheckRule_AllNumeric 测试全部为数值型的情况
// 验证当所有单元格都是数值型时，每个非空单元格都应该报错
func TestStringCheckRule_AllNumeric(t *testing.T) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 设置表头
	f.SetCellValue(sheetName, "A1", "")
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "A3", "Count")
	f.SetCellValue(sheetName, "A4", "")

	// 设置数值型单元格（全部错误）
	f.SetCellValue(sheetName, "A5", 100)
	f.SetCellValue(sheetName, "A6", 200)
	f.SetCellValue(sheetName, "A7", 300)

	cols, err := f.GetCols(sheetName)
	if err != nil {
		t.Fatalf("获取列数据失败: %v", err)
	}

	sheetMap := map[string]*excelize.File{
		sheetName: f,
	}

	rule := new(StringCheckRule)
	res := rule.Check(sheetName, cols, 0, 4, nil, sheetMap)

	assert.Len(t, res, 3, "全部为数值型时应检测到 3 个错误")
}

// TestStringCheckRule_EmptyAndCommit 测试空单元格和注释
// 验证 allowEmpty 和 allowCommit 参数是否生效
func TestStringCheckRule_EmptyAndCommit(t *testing.T) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 设置表头
	f.SetCellValue(sheetName, "A1", "")
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "A3", "Value")
	f.SetCellValue(sheetName, "A4", "")

	// 空单元格
	f.SetCellValue(sheetName, "A5", "")

	// 注释行
	f.SetCellValue(sheetName, "A6", "# 这是注释")

	// 数值型（应该检测到）
	f.SetCellValue(sheetName, "A7", 456)

	cols, err := f.GetCols(sheetName)
	if err != nil {
		t.Fatalf("获取列数据失败: %v", err)
	}

	sheetMap := map[string]*excelize.File{
		sheetName: f,
	}

	rule := new(StringCheckRule)
	params := map[string]string{
		"allowEmpty":  "true",
		"allowCommit": "true",
	}

	res := rule.Check(sheetName, cols, 0, 4, params, sheetMap)

	// 应该只检测到 A7 的类型错误
	assert.Len(t, res, 1, "应只检测到 1 个错误")

	// 验证错误是 A7
	assert.Equal(t, 2, res[0].Index, "错误应在第 3 个数据行（索引 2）")
}

// TestStringCheckRule_DisallowEmpty 测试不允许空值的情况
// 验证 allowEmpty=false 时空单元格会报错
func TestStringCheckRule_DisallowEmpty(t *testing.T) {
	f := excelize.NewFile()
	sheetName := "Sheet1"

	// 设置表头
	f.SetCellValue(sheetName, "A1", "")
	f.SetCellValue(sheetName, "A2", "")
	f.SetCellValue(sheetName, "A3", "Value")
	f.SetCellValue(sheetName, "A4", "")

	f.SetCellValue(sheetName, "B1", "")
	f.SetCellValue(sheetName, "B2", "")
	f.SetCellValue(sheetName, "B3", "Other")
	f.SetCellValue(sheetName, "B4", "")

	// A5 为空，但 B5 非空（确保整行不空）
	f.SetCellValue(sheetName, "A5", "")
	f.SetCellValue(sheetName, "B5", "x")

	// A6 文本型（正确）
	f.SetCellValue(sheetName, "A6", "abc")
	f.SetCellValue(sheetName, "B6", "y")

	cols, err := f.GetCols(sheetName)
	if err != nil {
		t.Fatalf("获取列数据失败: %v", err)
	}

	sheetMap := map[string]*excelize.File{
		sheetName: f,
	}

	rule := new(StringCheckRule)
	params := map[string]string{
		"allowEmpty":  "false",
		"allowCommit": "true",
	}

	res := rule.Check(sheetName, cols, 0, 4, params, sheetMap)

	// 应该只检测到 A5 的空值错误
	assert.Len(t, res, 1, "应只检测到 1 个错误")
	assert.True(t, strings.Contains(res[0].Reason, "空") || strings.Contains(res[0].Reason, "为空"),
		"错误信息应说明空值问题: %s", res[0].Reason)
}

// TestStringCheckRule_NoXlsxFile 测试 sheetMap 为空的情况
// 验证降级处理：当无法获取 excelize.File 时，返回空结果
func TestStringCheckRule_NoXlsxFile(t *testing.T) {
	rule := new(StringCheckRule)
	cols := [][]string{
		{"", "", "Id", "", "123", "456"},
	}

	// 传入空的 sheetMap
	res := rule.Check("TestSheet", cols, 0, 4, nil, nil)

	// 应该返回空结果（降级处理）
	assert.Empty(t, res, "无法获取 excelize.File 时应返回空结果")
}
