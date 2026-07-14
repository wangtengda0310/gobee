// Package check_manager 单元测试
package engine

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestColRuleAdapter 测试列规则适配器
func TestColRuleAdapter(t *testing.T) {
	// 创建一个简单的 mock 规则
	mockRule := &mockColRuleForAdapter{
		checkFunc: func(data map[int]string) ColCheckResultAdaptor {
			// 验证数据格式
			assert.NotNil(t, data, "数据不应为 nil")
			assert.Contains(t, data, 0, "应该包含第0行")
			return &mockColCheckResultForAdapter{
				ok:     true,
				reason: "测试通过",
			}
		},
	}

	// 创建适配器
	adapter := NewColRuleAdapter(mockRule)

	// 准备测试数据
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	sheetName := "TestSheet"
	cols := [][]string{
		{"", "", "Id", "int", "1", "2", "3"},
		{"", "", "Name", "string", "数据1", "数据2", "数据3"},
	}
	colIdx := 0
	startRowIdx := 4
	params := map[string]string{
		string(json_rule.ALLOW_EMPTY): "false",
	}
	sheetMap := map[string]*excelize.File{
		sheetName: xlsxFile,
	}

	// 执行检查
	result := adapter.Check(sheetName, cols, colIdx, startRowIdx, params, sheetMap)

	// 验证：由于 mock 规则返回成功，结果应该是 nil
	assert.Nil(t, result, "mock 规则返回成功时应该无错误")
}

// TestColRuleAdapter_ErrorCase 测试适配器处理错误场景
func TestColRuleAdapter_ErrorCase(t *testing.T) {
	// 创建一个返回错误的 mock 规则
	mockRule := &mockColRuleForAdapter{
		checkFunc: func(data map[int]string) ColCheckResultAdaptor {
			return &mockColCheckResultForAdapter{
				ok:     false,
				reason: "发现 2 个单元格类型错误",
			}
		},
	}

	// 创建适配器
	adapter := NewColRuleAdapter(mockRule)

	// 准备测试数据
	xlsxFile := excelize.NewFile()
	defer xlsxFile.Close()

	sheetName := "TestSheet"
	cols := [][]string{
		{"", "", "Id", "int", "1", "2"},
	}
	colIdx := 0
	startRowIdx := 4
	params := map[string]string{}
	sheetMap := map[string]*excelize.File{
		sheetName: xlsxFile,
	}

	// 执行检查
	result := adapter.Check(sheetName, cols, colIdx, startRowIdx, params, sheetMap)

	// 验证：应该返回错误
	assert.NotNil(t, result, "应该返回错误")
	assert.Len(t, result, 1, "应该有1个错误")
	assert.Contains(t, result[0].Reason, "发现 2 个单元格类型错误", "错误原因应该匹配")
}

// mockColRuleForAdapter 模拟的列规则（用于适配器测试）
type mockColRuleForAdapter struct {
	checkFunc func(data map[int]string) ColCheckResultAdaptor
}

func (m *mockColRuleForAdapter) Check(data map[int]string) ColCheckResultAdaptor {
	return m.checkFunc(data)
}

// mockColCheckResultForAdapter 模拟的检查结果
type mockColCheckResultForAdapter struct {
	ok     bool
	reason string
}

func (m *mockColCheckResultForAdapter) IsOk() bool {
	return m.ok
}

func (m *mockColCheckResultForAdapter) GetReason() string {
	return m.reason
}
