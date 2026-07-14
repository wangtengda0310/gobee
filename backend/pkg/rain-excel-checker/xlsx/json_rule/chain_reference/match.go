// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现关系链检查的匹配阶段操作函数（对应前端"匹配卡片"）
// 匹配阶段用于两阶段门控模型的 Phase1（交汇判断）：
//   - 匹配成功 = 两链有交集/关联 → 进入 Phase2 比较阶段
//   - 匹配失败 = 两链无关联 → 跳过，不报错
//
// 所有操作语义统一为"从左链到右链"，与比较阶段一致
package chain_reference

import (
	"fmt"
	"strings"
)

// MatchByType 匹配阶段分发：根据匹配类型调用对应的匹配函数
//
// 语义分类：
//   - **门控类型**（matched=false 时静默跳过）：verify_exists、date_*
//   - **强制类型**（matched=false 时由 Compare 层报错）：verify_must_exist
//
// 强制类型返回 false + 详细 reason（缺失值列表），由 MatchHandler 写入 ctx.Violation/Reason
func MatchByType(matchType string, leftVals, rightVals []string) (bool, string) {
	switch matchType {
	case "date_equals":
		return matchDateEquals(leftVals, rightVals)
	case "date_before_or_equal":
		return matchDateBeforeOrEqual(leftVals, rightVals)
	case "date_after_or_equal":
		return matchDateAfterOrEqual(leftVals, rightVals)
	case "verify_exists":
		return matchVerifyExists(leftVals, rightVals)
	case "verify_must_exist":
		return matchVerifyMustExist(leftVals, rightVals)
	default:
		return true, fmt.Sprintf("未知的匹配类型: %s", matchType)
	}
}

// IsMatchTypeStrict 判断匹配类型是否为强制类型
// 强制类型在 Match 不通过时由 Compare 层报错，而不是静默跳过
func IsMatchTypeStrict(matchType string) bool {
	return matchType == "verify_must_exist"
}

// matchDateEquals 日期相同匹配（复用比较阶段函数）
func matchDateEquals(leftVals, rightVals []string) (bool, string) {
	matched, _ := compareDateEquals(leftVals, rightVals)
	return matched, ""
}

// matchDateBeforeOrEqual 日期早于或等于匹配（复用比较阶段函数）
func matchDateBeforeOrEqual(leftVals, rightVals []string) (bool, string) {
	matched, _ := compareDateBeforeOrEqual(leftVals, rightVals)
	return matched, ""
}

// matchDateAfterOrEqual 日期晚于或等于匹配（复用比较阶段函数）
func matchDateAfterOrEqual(leftVals, rightVals []string) (bool, string) {
	matched, _ := compareDateAfterOrEqual(leftVals, rightVals)
	return matched, ""
}

// matchVerifyExists 存在性验证匹配：左链值在右链中全部找到 → 门控通过，否则不通过
// 匹配阶段只需通过/不通过，不报告缺失值详情
func matchVerifyExists(leftVals, rightVals []string) (bool, string) {
	rightSet := make(map[string]bool)
	for _, v := range rightVals {
		rightSet[strings.TrimSpace(v)] = true
	}

	for _, lv := range leftVals {
		if !rightSet[strings.TrimSpace(lv)] {
			return false, ""
		}
	}
	return true, ""
}

// matchVerifyMustExist 引用完整性强制检查：左链值必须在右链中全部存在
// 与 matchVerifyExists 的差异：
//   - matchVerifyExists：缺失时返回 (false, "")，由 Compare 层静默跳过（门控语义）
//   - matchVerifyMustExist：缺失时返回 (false, "缺失值列表")，由 Compare 层报错（引用完整性语义）
//
// 用于检测"DrawSkin.byproduct 引用必须在 ShopGoods.Item 中存在"这类引用完整性约束
func matchVerifyMustExist(leftVals, rightVals []string) (bool, string) {
	rightSet := make(map[string]bool)
	for _, v := range rightVals {
		rightSet[strings.TrimSpace(v)] = true
	}

	var missing []string
	for _, lv := range leftVals {
		key := strings.TrimSpace(lv)
		if !rightSet[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return false, fmt.Sprintf("引用值 [%s] 在目标列中不存在", strings.Join(missing, ", "))
	}
	return true, ""
}
