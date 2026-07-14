// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的比较处理器（CompareHandler）
// Compare 是洋葱链的倒数第二层（Validate 之后），在所有步骤执行完毕后做最终比较
package chain_reference

import (
	"strings"
	"time"
)

// CompareHandler 比较处理器
// 位于洋葱链的 Validate 之后、左链步骤之前
// 先调用 next 让内层（左链 + 右链 + Match）全部执行完毕，再做最终比较判定
//
// 比较模式：
//   - 通用两阶段门控（MatchCompareType 非空）：Phase1 由 Match handler 完成，
//     Phase2 比较当前列值 vs 右链 FirstStepInputValues
//   - time_overlap 特化（UseTimeOverlap=true）：使用 CompareTwoPhase 两阶段比较
//   - 单阶段退化：比较左链 MatchValues vs 右链 MatchValues
type CompareHandler struct {
	CompareType    string // 比较类型：verify_exists/date_*/time_overlap 等
	LeftKey        string // 左链时间比较键
	RightKey       string // 右链时间比较键
	UseTimeOverlap bool   // 是否使用 time_overlap 特化（两链都有 CompareCol）
}

// Handle 执行比较判定
//
// 执行流程：
//  1. 调用 next(ctx) — 执行内层所有步骤（左链 → Match → 右链）
//  2. 后置处理：
//     a. 如果 ctx.Err != nil：直接返回
//     b. 如果 ctx.Matched == false（两阶段门控未通过）：直接返回
//     c. 根据比较模式判定违规并写入 ctx.Violation 和 ctx.Reason
func (h *CompareHandler) Handle(ctx *ChainContext, next NextFunc) error {
	// 前向：先执行内层所有步骤
	if err := next(ctx); err != nil {
		return err
	}

	// 后置处理：内层执行完毕后进行最终比较

	// 如果内层产生错误，直接返回
	if ctx.Err != nil {
		return nil
	}

	// 两阶段门控未通过（Match handler 判定两链未交汇），不报错
	matchCompareType := ctx.MatchCompareType()
	if matchCompareType != "" && !ctx.Matched {
		// 例外：强制类型（verify_must_exist）已在 MatchHandler 设置 ctx.Violation/Reason，
		// 此时跳过比较逻辑（右链未执行，RightResult 为 nil），仅做预警窗口过滤
		if ctx.Violation {
			if ShouldSuppressByWarnBefore(ctx, time.Now()) {
				ctx.Violation = false
				ctx.Reason = ""
			}
		}
		return nil
	}

	// 根据比较模式判定违规
	var violation bool
	var reason string

	if matchCompareType != "" {
		// === 通用两阶段门控比较 ===
		// Phase 1 已由 Match handler 完成（ctx.Matched == true）
		// Phase 2: 比较当前列值 vs 右链第一步查找值
		violation, reason = h.compareTwoPhaseGate(ctx)
	} else if h.UseTimeOverlap {
		// === time_overlap 特化（两链都有 CompareCol）===
		violation, reason = CompareTwoPhase(ctx.LeftResult, ctx.RightResult, h.LeftKey, h.RightKey)
	} else if h.CompareType == "time_overlap" {
		// === time_overlap 但只有一链有 CompareCol ===
		leftVals := h.safeMatchValues(ctx.LeftResult)
		rightVals := h.safeMatchValues(ctx.RightResult)
		violation, reason = CompareTimeMatch(leftVals, rightVals, h.LeftKey, h.RightKey)
	} else {
		// === 单阶段退化：比较左链 MatchValues vs 右链 MatchValues ===
		leftVals := h.safeMatchValues(ctx.LeftResult)
		rightVals := h.safeMatchValues(ctx.RightResult)
		violation, reason = CompareByType(h.CompareType, leftVals, rightVals)
	}

	// 写入比较结果
	if violation {
		// 预警窗口过滤：如果违规目标的生效时间距今超过 warnBefore，静默不报错
		if ShouldSuppressByWarnBefore(ctx, time.Now()) {
			violation = false
			reason = ""
		}
	}
	if violation {
		ctx.Violation = true
		ctx.Reason = reason
	}

	return nil
}

// compareTwoPhaseGate 通用两阶段门控比较
// Phase 2: 用 chainCompare 比较当前列值与右链第一步 preCol 的查找值
// 当左链第一步 IsArray=true 时，先按逗号拆分当前列值再比较
func (h *CompareHandler) compareTwoPhaseGate(ctx *ChainContext) (bool, string) {
	currentColVal := ctx.CurrentCellValue()
	var rightFirstInputVals []string
	if ctx.RightResult != nil {
		rightFirstInputVals = ctx.RightResult.GetFirstStepInputValues()
	}

	// 左链第一步 isArray 控制当前列值的拆分
	var currentVals []string
	if len(ctx.Config.Left.Steps) > 0 && strings.ToLower(ctx.Config.Left.Steps[0].IsArray) == "true" {
		currentVals = SplitArrayElements(currentColVal, ",")
	} else {
		currentVals = []string{currentColVal}
	}

	return CompareByType(h.CompareType, currentVals, rightFirstInputVals)
}

// safeMatchValues 安全获取 ChainResult 的 MatchValues
// result 为 nil 时返回 nil 而非 panic
func (h *CompareHandler) safeMatchValues(result *ChainResult) []string {
	if result == nil {
		return nil
	}
	return result.MatchValues()
}
