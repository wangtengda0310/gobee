package exceltest

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// CheckContext 检查上下文
// 统一收集检查过程中的所有结果，最后统一发送通知
type CheckContext struct {
	ParseErrors  []*SheetParseError
	ColResults   []*json_rule.ColCheckResult
	TableResults []*json_rule.TableCheckResult
}

// NewCheckContext 创建检查上下文
func NewCheckContext() *CheckContext {
	return &CheckContext{
		ParseErrors:  make([]*SheetParseError, 0),
		ColResults:   make([]*json_rule.ColCheckResult, 0),
		TableResults: make([]*json_rule.TableCheckResult, 0),
	}
}

// AddParseErrors 添加解析错误
func (c *CheckContext) AddParseErrors(errors []*SheetParseError) {
	if errors == nil {
		return
	}
	c.ParseErrors = append(c.ParseErrors, errors...)
}

// AddColResults 添加列级检查结果
func (c *CheckContext) AddColResults(results []*json_rule.ColCheckResult) {
	if results == nil {
		return
	}
	c.ColResults = append(c.ColResults, results...)
}

// AddTableResults 添加表级检查结果
func (c *CheckContext) AddTableResults(results []*json_rule.TableCheckResult) {
	if results == nil {
		return
	}
	c.TableResults = append(c.TableResults, results...)
}

// ToResult 转换为最终结果
func (c *CheckContext) ToResult() *ExcelCheckResult {
	return &ExcelCheckResult{
		ParseErrors:  c.ParseErrors,
		ColResults:   c.ColResults,
		TableResults: c.TableResults,
	}
}

// HasErrors 是否有错误
// 解析错误、列级检查失败、表级检查失败都算错误
func (c *CheckContext) HasErrors() bool {
	// 有解析错误
	if len(c.ParseErrors) > 0 {
		return true
	}

	// 有列级检查失败
	for _, r := range c.ColResults {
		if r != nil && !r.Ok {
			return true
		}
	}

	// 有表级检查失败（Ok=false 表示错误检测规则发现问题）
	for _, r := range c.TableResults {
		if r != nil && !r.Ok {
			return true
		}
	}

	return false
}
