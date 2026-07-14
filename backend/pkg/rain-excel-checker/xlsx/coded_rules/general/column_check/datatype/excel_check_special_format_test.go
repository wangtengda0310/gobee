// Package datatype 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package datatype

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestSpecialFormatCheckRule(t *testing.T) {
	cols, cIdx, c1idx, params, sheetMap := fakeDataSpecialFormatCheckRule()

	bc := new(SpecialFormatCheckRule)
	res := bc.Check("", cols, cIdx, c1idx, params, sheetMap)

	jsonData, _ := json.MarshalIndent(res, "", " ")
	t.Log(string(jsonData))
}

func fakeDataSpecialFormatCheckRule() (cols [][]string, cIdx, c1idx int, params map[string]string, sheetMap map[string]*excelize.File) {
	return [][]string{
		{"{2123;356}", "{12:1}", "{213,34}{345,123}", "{1}-{1}", "1{1}-{2}"},
	}, 0, 0, nil, nil
}

func TestSpecialFormatCheckRule_ArrayMode(t *testing.T) {
	// 数组模式测试：逗号分隔的多组 {id;count}
	cols := [][]string{
		{"{2000001;1},{2000001;1},{2000001;1}", "{7000001;3},{4000001;5}", "{1;2}", "invalid", "{1;2},{3;4"},
	}

	bc := new(SpecialFormatCheckRule)

	// 数组模式
	params := map[string]string{
		"isArray": "true",
	}
	res := bc.Check("", cols, 0, 0, params, nil)

	// 第0行: {2000001;1},{2000001;1},{2000001;1} - 应该通过
	// 第1行: {7000001;3},{4000001;5} - 应该通过
	// 第2行: {1;2} - 应该通过（单元素数组）
	// 第3行: invalid - 应该失败
	// 第4行: {1;2},{3;4 - 应该失败（缺少右花括号）
	assert.Len(t, res, 2, "应该只有2个错误")
	assert.Equal(t, 3, res[0].Index, "第3行应该报错")
	assert.Equal(t, 4, res[1].Index, "第4行应该报错")
}

func TestSpecialFormatCheckRule_SingleMode(t *testing.T) {
	// 单值模式测试（默认模式）
	cols := [][]string{
		{"{2123;356}", "{1;2}{3;4}", "{1;2},{3;4}"},
	}

	bc := new(SpecialFormatCheckRule)

	// 单值模式（默认）
	params := map[string]string{}
	res := bc.Check("", cols, 0, 0, params, nil)

	// 第0行: {2123;356} - 应该通过
	// 第1行: {1;2}{3;4} - 应该通过
	// 第2行: {1;2},{3;4} - 应该失败（逗号分隔在单值模式下不允许）
	assert.Len(t, res, 1, "应该只有1个错误")
	assert.Equal(t, 2, res[0].Index, "第2行应该报错")
}
