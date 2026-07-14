// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"os"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ResourceCheckRule 检查配表内的资源文件是否存在于客户端项目中
//
// 支持参数：
//   - clientPath: 客户端项目本地路径（如 D:\work\client），必填
//   - prefix: 父文件夹前缀（可选，用于字段值只有文件名时自动补全路径）
//   - allowEmpty: 是否允许空值，默认 false
//   - allowCommit: 是否允许注释，默认 false
//   - breakLine: 连续空行阈值，默认 3
type ResourceCheckRule struct{}

// Check 检查配表内的资源文件是否存在
//
// 执行流程：
//  1. 解析 BREAK_LINE 参数，设置空行连续阈值（默认3行）
//  2. 自动检测数据结束位置并提取列数据
//  3. 解析 ALLOW_EMPTY 和 ALLOW_COMMIT 参数
//  4. 解析 clientPath 参数（客户端项目路径），为空则直接返回
//  5. 解析 prefix 参数（可选的父文件夹前缀）
//  6. 遍历列数据，对每个单元格：
//     a. 处理空值和注释（根据配置决定是否跳过）
//     b. 拼接完整路径：clientPath + separator + prefix + separator + 单元格值
//     c. 检查文件是否存在，不存在则记录错误
//  7. 返回所有错误信息
func (c *ResourceCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	// 步骤1: 解析 BREAK_LINE 参数
	breakLine := helpers.ParseBreakLine(params)

	// 步骤2: 自动检测数据结束位置并提取列数据
	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]

	// 初始化检查结果切片
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 步骤3: 解析 ALLOW_EMPTY 和 ALLOW_COMMIT 参数
	allowEmpty := helpers.ParseAllowEmpty(params)
	allowCommit := helpers.ParseAllowCommit(params)

	// 步骤4: 解析 clientPath 参数
	clientPath, ok := params["clientPath"]
	if !ok || strings.TrimSpace(clientPath) == "" {
		// 没有配置客户端路径，无法执行检查，直接返回
		return res
	}
	clientPath = strings.TrimSpace(clientPath)

	// 步骤5: 解析 prefix 参数（可选）
	prefix := params["prefix"]

	// 步骤6: 遍历列数据
	for i, cellValue := range myColData {
		// 步骤6a: 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 步骤6b: 拼接完整路径
		fullPath := buildResourcePath(clientPath, prefix, strings.TrimSpace(cellValue))

		// 步骤6c: 检查文件是否存在
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   "资源文件不存在: " + cellValue,
			})
		}
	}

	// 步骤7: 返回所有错误信息
	return res
}

// buildResourcePath 拼接完整的资源文件路径
//
// 拼接规则：
//   - 有 prefix: clientPath + separator + prefix + separator + cellValue
//   - 无 prefix: clientPath + separator + cellValue
//
// cellValue 本身可能包含子目录（如 Prefabs/Prefab/Hero/xxx.prefab），
// separator 统一使用 os 路径分隔符
func buildResourcePath(clientPath, prefix, cellValue string) string {
	clientPath = filepath.Clean(clientPath)
	if prefix != "" {
		return filepath.Join(clientPath, prefix, cellValue)
	}
	return filepath.Join(clientPath, cellValue)
}
