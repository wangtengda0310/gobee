// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现关系链检查的比较阶段操作函数（对应前端"比较卡片"）
// 所有操作语义统一为"从左链到右链"：
//   - leftVals = 左链提取的值
//   - rightVals = 右链提取的值
//   - 返回 (是否违规, 原因描述)
package chain_reference

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
)

// CompareByType 比较阶段分发：根据比较类型调用对应的比较函数
//
// 注：verify_must_exist 与 verify_exists 在 Compare 阶段语义相同（缺失即报错）；
// 两者的差异只体现在 Match 阶段（IsMatchTypeStrict 区分门控 vs 强制）。
func CompareByType(compareType string, leftVals, rightVals []string) (bool, string) {
	switch compareType {
	case "date_equals":
		return compareDateEquals(leftVals, rightVals)
	case "date_before_or_equal":
		return compareDateBeforeOrEqual(leftVals, rightVals)
	case "date_after_or_equal":
		return compareDateAfterOrEqual(leftVals, rightVals)
	case "verify_exists", "verify_must_exist":
		return compareVerifyExists(leftVals, rightVals)
	default:
		return true, fmt.Sprintf("未知的比较类型: %s", compareType)
	}
}

// compareDateEquals 日期相同比较（秒级精度）：左链与右链有相同时间点 → 违规
func compareDateEquals(leftVals, rightVals []string) (bool, string) {
	leftTimeSet := parseTimeSet(leftVals)
	rightTimeSet := parseTimeSet(rightVals)

	var matched []string
	for key := range rightTimeSet {
		if leftTimeSet[key] {
			matched = append(matched, key)
		}
	}
	if len(matched) > 0 {
		return true, fmt.Sprintf("关系链检查(date_equals): 时间点相同 [%s]", strings.Join(matched, ", "))
	}
	return false, ""
}

// compareDateBeforeOrEqual 日期早于或等于比较：右链时间早于或等于左链某时间 → 违规
func compareDateBeforeOrEqual(leftVals, rightVals []string) (bool, string) {
	leftTimes := parseTimeList(leftVals)
	rightTimes := parseTimeList(rightVals)

	var matched []string
	for _, rt := range rightTimes {
		for _, lt := range leftTimes {
			if rt.Before(lt) || rt.Equal(lt) {
				matched = append(matched, fmt.Sprintf("%s <= %s", rt.Format("2006-01-02 15:04:05"), lt.Format("2006-01-02 15:04:05")))
				break
			}
		}
	}
	if len(matched) > 0 {
		return true, fmt.Sprintf("关系链检查(date_before_or_equal): 右链时间早于或等于左链 [%s]", strings.Join(matched, ", "))
	}
	return false, ""
}

// compareDateAfterOrEqual 日期晚于或等于比较：右链时间晚于或等于左链某时间 → 违规
func compareDateAfterOrEqual(leftVals, rightVals []string) (bool, string) {
	leftTimes := parseTimeList(leftVals)
	rightTimes := parseTimeList(rightVals)

	var matched []string
	for _, rt := range rightTimes {
		for _, lt := range leftTimes {
			if rt.After(lt) || rt.Equal(lt) {
				matched = append(matched, fmt.Sprintf("%s >= %s", rt.Format("2006-01-02 15:04:05"), lt.Format("2006-01-02 15:04:05")))
				break
			}
		}
	}
	if len(matched) > 0 {
		return true, fmt.Sprintf("关系链检查(date_after_or_equal): 右链时间晚于或等于左链 [%s]", strings.Join(matched, ", "))
	}
	return false, ""
}

// CompareTwoPhase 两阶段比较（仅 time_overlap + 两链都有 compareCol 时使用）
// 第一阶段：匹配 Match 值（如武将ID），找到交集
// 第二阶段：比较交集项的 Compare 值（如时间），检查时间点是否重叠
func CompareTwoPhase(leftResult, rightResult *ChainResult, leftKey, rightKey string) (bool, string) {
	leftVals := leftResult.MatchValues()
	rightVals := rightResult.MatchValues()

	if len(leftVals) == 0 || len(rightVals) == 0 {
		return false, ""
	}

	leftMatchSet := make(map[string]bool)
	leftCompareMap := make(map[string]string)
	for _, v := range leftResult.Values {
		matchKey := strings.TrimSpace(v.Match)
		leftMatchSet[matchKey] = true
		leftCompareMap[matchKey] = strings.TrimSpace(v.Compare)
	}

	var matchedPairs []string
	for _, rv := range rightResult.Values {
		rvMatch := strings.TrimSpace(rv.Match)
		if !leftMatchSet[rvMatch] {
			continue
		}

		leftTimeStr := leftCompareMap[rvMatch]
		rightTimeStr := strings.TrimSpace(rv.Compare)

		leftTime := helpers.ParseDate(leftTimeStr)
		rightTime := helpers.ParseDate(rightTimeStr)

		if !leftTime.IsZero() && !rightTime.IsZero() {
			leftTimeKey := leftTime.Format("2006-01-02 15:04:05")
			rightTimeKey := rightTime.Format("2006-01-02 15:04:05")

			if leftTimeKey == rightTimeKey {
				matchedPairs = append(matchedPairs, fmt.Sprintf("%s(%s)", rvMatch, leftTimeKey))
			}
		} else if leftTimeStr == rightTimeStr && leftTimeStr != "" {
			matchedPairs = append(matchedPairs, fmt.Sprintf("%s(%s)", rvMatch, leftTimeStr))
		}
	}

	if len(matchedPairs) > 0 {
		return true, fmt.Sprintf("关系链检查(time_overlap): 时间点匹配 [%s]", strings.Join(matchedPairs, ", "))
	}

	return false, ""
}

// CompareTimeMatch 时间点匹配比较（用于两阶段 time_overlap）
func CompareTimeMatch(leftVals, rightVals []string, leftKey, rightKey string) (bool, string) {
	var leftTimes []time.Time
	for _, v := range leftVals {
		t := helpers.ParseDate(v)
		if !t.IsZero() {
			leftTimes = append(leftTimes, t)
		}
	}

	var rightTimes []time.Time
	for _, v := range rightVals {
		t := helpers.ParseDate(v)
		if !t.IsZero() {
			rightTimes = append(rightTimes, t)
		}
	}

	leftTimeSet := make(map[string]bool)
	for _, t := range leftTimes {
		leftTimeSet[t.Format("2006-01-02 15:04:05")] = true
	}

	var matched []string
	for _, rt := range rightTimes {
		key := rt.Format("2006-01-02 15:04:05")
		if leftTimeSet[key] {
			matched = append(matched, key)
		}
	}

	if len(matched) > 0 {
		return true, fmt.Sprintf("关系链检查(time_overlap): 时间点匹配 [%s]", strings.Join(matched, ", "))
	}

	return false, ""
}

// parseTimeSet 将字符串列表解析为时间集合（秒级字符串作为 key）
func parseTimeSet(vals []string) map[string]bool {
	result := make(map[string]bool)
	for _, v := range vals {
		t := helpers.ParseDate(strings.TrimSpace(v))
		if !t.IsZero() {
			result[t.Format("2006-01-02 15:04:05")] = true
		}
	}
	return result
}

// parseTimeList 将字符串列表解析为时间列表
func parseTimeList(vals []string) []time.Time {
	var result []time.Time
	for _, v := range vals {
		t := helpers.ParseDate(strings.TrimSpace(v))
		if !t.IsZero() {
			result = append(result, t)
		}
	}
	return result
}

// compareVerifyExists 存在性验证比较：左链值在右链集合中全部找到 → 不违规，有缺失 → 违规
// 语义：验证左链提取的每个值是否都存在于右链集合中，用于检查引用完整性等场景
func compareVerifyExists(leftVals, rightVals []string) (bool, string) {
	// 构建右链值集合用于快速查找
	rightSet := make(map[string]bool)
	for _, v := range rightVals {
		rightSet[strings.TrimSpace(v)] = true
	}

	// 遍历左链值，检查每个值是否在右链集合中
	var notFound []string
	for _, lv := range leftVals {
		lv = strings.TrimSpace(lv)
		if !rightSet[lv] {
			notFound = append(notFound, lv)
		}
	}

	if len(notFound) > 0 {
		return true, fmt.Sprintf("关系链检查(verify_exists): 左链值未在右链中找到 [%s]", strings.Join(notFound, ", "))
	}
	return false, ""
}
