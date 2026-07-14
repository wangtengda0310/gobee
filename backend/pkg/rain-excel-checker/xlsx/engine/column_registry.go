// Package engine 提供检查器管理功能
// 本包负责管理所有列级检查器的注册、获取和调度
package engine

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/base"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/business/calculation"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/business/date"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/business/numeric"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/business/pinyin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/business/reference"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check/datatype"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// Manager 全局列级检查器管理器实例
var Manager = NewCheckerManager()

// CheckerManager 列级检查器管理器
// 负责管理所有列级检查器的注册和获取
type CheckerManager struct {
	checkerMap map[json_rule.EColRule]Checker // 检查器映射表
}

// NewCheckerManager 创建新的列级检查器管理器
// 返回一个初始化完成的检查器管理器实例
func NewCheckerManager() *CheckerManager {
	return &CheckerManager{
		checkerMap: make(map[json_rule.EColRule]Checker),
	}
}

// Reg 注册列级检查器
// 参数:
//   - checkRule: 检查规则类型枚举
//   - checker: 对应的检查器实例
func (cm *CheckerManager) Reg(checkRule json_rule.EColRule, checker Checker) {
	cm.checkerMap[checkRule] = checker
}

// GetChecker 获取指定类型的检查器
// 参数:
//   - checkRule: 检查规则类型枚举
//
// 返回:
//   - 检查器实例，如果不存在则返回 nil
func (cm *CheckerManager) GetChecker(checkRule json_rule.EColRule) Checker {
	ck, exist := cm.checkerMap[checkRule]
	if !exist {
		return nil
	}
	return ck
}

// init 包初始化函数
// 自动注册所有列级检查器到管理器中
func init() {
	// 基础检查规则 (base)
	Manager.Reg(json_rule.ALL_BASE, new(base.AllBaseCheckRule))
	Manager.Reg(json_rule.INCREASE_ID, new(base.IncreaseCheckRule))
	Manager.Reg(json_rule.UNIQUE, new(base.UniqueCheckRule))
	Manager.Reg(json_rule.CHS_ONLY, new(base.CHSCheckRule))
	Manager.Reg(json_rule.NOT_EMPTY, new(base.NotEmptyCheckRule))
	Manager.Reg(json_rule.SERVER_OR_CLIENT, new(base.ServerOrClientCheckRule))

	// 数据类型检查规则 (datatype)
	Manager.Reg(json_rule.NUMERIC, new(datatype.NumericCheckRule)) // 检查是否为数值
	Manager.Reg(json_rule.DATE, new(datatype.DateCheckRule))       // 检查日期格式
	Manager.Reg(json_rule.BOOLEAN, new(datatype.BooleanCheckRule)) // 检查布尔值
	Manager.Reg(json_rule.STRING, new(datatype.StringCheckRule))   // 检查单元格应为字符串

	// 范围检查规则 (date)
	Manager.Reg(json_rule.DATE_DURATION, new(date.DateDurationCheckRule))
	Manager.Reg(json_rule.DATE_RANGE, new(date.DateRangeCheckRule))

	// 数值范围检查规则 (numeric)
	Manager.Reg(json_rule.NUMERIC_RANGE, new(numeric.NumericRangeCheckRule))
	Manager.Reg(json_rule.ENUM, new(numeric.EnumCheckRule))

	// 引用检查规则 (reference)
	Manager.Reg(json_rule.FOREIGN_KEY, new(reference.ForeignKeyCheckRule))
	Manager.Reg(json_rule.CROSS_REFERENCE, new(reference.CrossReferenceCheckRule))
	Manager.Reg(json_rule.SPLIT_REGERENCE, new(reference.SplitReferenceCheckRule))
	Manager.Reg(json_rule.CHAIN_REFERENCE, new(reference.ChainReferenceCheckRule))
	Manager.Reg(json_rule.REGEX, new(reference.RegexCheckRule))

	// 格式检查规则 (datatype)
	Manager.Reg(json_rule.SPECIAL_FORMAT, new(datatype.SpecialFormatCheckRule))

	// 复杂业务规则
	Manager.Reg(json_rule.WEIGHT_SUM, new(calculation.WeightSumCheckRule))
	Manager.Reg(json_rule.DATE_CONSISTENCY, new(calculation.DateConsistencyCheckRule))
	Manager.Reg(json_rule.RESOURCE, new(base.ResourceCheckRule))
	Manager.Reg(json_rule.PIN_YIN_CHS, new(pinyin.PinYinCHSCheckRule))
	Manager.Reg(json_rule.RICH_TEXT, new(datatype.RichTextCheckRule))
}
