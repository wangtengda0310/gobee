// Package engine 提供检查器管理功能
//
// 本文件实现适配器，将 src.ColRule 接口适配到现有的 Checker 接口
// 这样可以在现有工厂中注册使用新接口的规则
package engine

import (
	"context"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// ColRuleAdapter 列规则适配器
// 将 src.ColRule 接口适配到现有的 Checker 接口
//
// 职责：
//  1. 在调用 colRule.Check() 前设置上下文到 adaptor
//  2. 调用完成后清理上下文
//  3. 转换数据格式（现有列数据 → map[int]string）
//
// 使用示例：
//
//	// 创建使用新接口的规则
//	rule := &MyStringCheckRule{}
//
//	// 适配到旧接口
//	adapter := engine.NewColRuleAdapter(rule)
//
//	// 注册到工厂（使用旧接口）
//	Manager.Reg(json_rule.STRING, adapter)
type ColRuleAdapter struct {
	rule ColRuleInterface
}

// ColRuleInterface 列规则接口（来自 src 包的设计）
// 保持接口简洁，只接收列数据
//
// 参数：
//   - data: map[int]string，key 为相对行号（从0开始），value 为单元格值
//
// 返回：
//   - ColCheckResultAdaptor: 检查结果
type ColRuleInterface interface {
	Check(data map[int]string) ColCheckResultAdaptor
}

// ColCheckResultAdaptor 列检查结果接口
type ColCheckResultAdaptor interface {
	IsOk() bool
	GetReason() string
}

// NewColRuleAdapter 创建列规则适配器
func NewColRuleAdapter(rule ColRuleInterface) *ColRuleAdapter {
	return &ColRuleAdapter{rule: rule}
}

// Check 实现 Checker 接口
// 这是现有工厂调用的方法签名
func (a *ColRuleAdapter) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	// 1. 构造检查上下文
	checkCtx := &diff.CheckContext{
		CheckParam: &json_rule.CheckParam{
			SheetName:   sheetName,
			Cols:        cols,
			StartRowIdx: startRowIdx,
			Params:      params,
			SheetMap:    sheetMap,
		},
		ColName:   "", // 可选：如果有列名信息
		SheetName: sheetName,
		ColIndex:  colIdx,
	}

	// 2. 获取列数据
	colData := cols[colIdx][startRowIdx:]

	// 3. 存储上下文到 adaptor
	ctx, cancel := diff.GlobalAdaptor.StoreContext(context.TODO(), sheetName, "", checkCtx)
	defer cancel()

	// 4. 转换数据格式: []string → map[int]string
	dataMap := make(map[int]string)
	for rowIdx, value := range colData {
		dataMap[rowIdx] = value
	}

	// 5. 调用新接口的规则
	result := a.rule.Check(dataMap)

	// 确保 context 被取消
	_ = ctx

	// 6. 转换结果格式
	if result.IsOk() {
		return nil // 检查通过，无错误
	}

	// 检查失败，返回错误信息
	// 注意：由于新接口只返回一个字符串原因，我们无法定位具体错误行
	// 这是简化接口的代价，如果需要详细错误信息，应该使用旧接口
	return []*json_rule.CellError{
		{
			Index:  0,
			Reason: result.GetReason(),
		},
	}
}
