// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的参数校验处理器（ValidateHandler）
// Validate 是洋葱链的最外层，在所有业务逻辑之前执行参数合法性校验
package chain_reference

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ==================== 已知的比较/匹配/过滤类型 ====================

// knownFilterModes 已知的过滤模式集合
// 与 FilterRowsByConditionEx 中的 switch case 保持同步
var knownFilterModes = map[string]bool{
	"":           true, // 单值模式（默认）
	"multi":      true, // 多值模式
	"withinDays": true, // 时间窗口模式
}

// ==================== 已知的比较/匹配类型 ====================

// knownCompareTypes 已知的比较类型集合
// 与 CompareByType 中的 switch case 保持同步
var knownCompareTypes = map[string]bool{
	"date_equals":          true,
	"date_before_or_equal": true,
	"date_after_or_equal":  true,
	"time_overlap":         true,
	"verify_exists":        true,
	"verify_must_exist":    true,
}

// knownMatchTypes 已知的匹配类型集合
// 与 MatchByType 中的 switch case 保持同步
var knownMatchTypes = map[string]bool{
	"date_equals":          true,
	"date_before_or_equal": true,
	"date_after_or_equal":  true,
	"verify_exists":        true,
	"verify_must_exist":    true,
}

// ==================== 参数校验处理器 ====================

// ValidateHandler 参数校验处理器（洋葱模型最外层）
// 在执行任何业务逻辑之前校验 ChainContext 中的配置参数
// 校验失败时直接返回错误，不调用 next（阻止内层执行）
type ValidateHandler struct{}

// Handle 执行参数校验
// 校验通过后调用 next 将控制权传递给内层处理器
func (h *ValidateHandler) Handle(ctx *ChainContext, next NextFunc) error {
	if err := validateConfig(ctx); err != nil {
		return err
	}
	return next(ctx)
}

// ==================== 校验逻辑 ====================

// validateConfig 校验 ChainContext 中的配置参数
//
// 执行流程：
//  1. 校验 chainSteps 参数非空
//  2. 校验 chainSteps JSON 格式正确
//  3. 校验至少一条链有步骤
//  4. 校验右链步骤的 Sheet 非空（右链需要跳转到目标表）
//  5. 校验 chainCompare 是已知类型
//  6. 校验 chainMatchCompare 是已知类型（如非空）
//  7. 校验所有步骤中的正则模式可编译
func validateConfig(ctx *ChainContext) error {
	// 步骤 1: 校验 chainSteps 参数非空
	chainStepsJSON := ctx.ChainStepsJSON()
	if chainStepsJSON == "" {
		return fmt.Errorf("关系链参数 chainSteps 未配置")
	}

	// 步骤 2: 校验 chainSteps JSON 格式正确
	var rawConfig ChainPairConfig
	if err := json.Unmarshal([]byte(chainStepsJSON), &rawConfig); err != nil {
		return fmt.Errorf("关系链配置 JSON 格式错误: %w", err)
	}

	// 步骤 3: 校验至少一条链有步骤
	if len(rawConfig.Left.Steps) == 0 && len(rawConfig.Right.Steps) == 0 {
		return fmt.Errorf("关系链配置中两条链都没有步骤")
	}

	// 步骤 4: 校验右链步骤的 Sheet 非空（右链需要跳转到目标表）
	for i, step := range rawConfig.Right.Steps {
		if step.Sheet == "" {
			return fmt.Errorf("右链步骤%d: 目标表(sheet)不能为空", i+1)
		}
	}

	// 步骤 5: 校验 chainCompare 是已知类型
	compareType := ctx.CompareType()
	if !knownCompareTypes[compareType] {
		return fmt.Errorf("未知的比较类型: %s", compareType)
	}

	// 步骤 6: 校验 chainMatchCompare 是已知类型（如非空）
	matchType := ctx.MatchCompareType()
	if matchType != "" && !knownMatchTypes[matchType] {
		return fmt.Errorf("未知的匹配类型: %s", matchType)
	}

	// 步骤 7: 校验所有步骤中的正则模式可编译
	allSteps := append(rawConfig.Left.Steps, rawConfig.Right.Steps...)
	for i, step := range allSteps {
		if step.Pattern != "" {
			if _, err := regexp.Compile(step.Pattern); err != nil {
				return fmt.Errorf("步骤%d: 正则编译失败: %w", i+1, err)
			}
		}
	}

	// 步骤 7.5: 校验所有步骤的 filterMode/filterDays 配置合法
	// - filterMode 必须是已知类型（""/"multi"/"withinDays"）
	// - filterMode="withinDays" 时 filterDays 必须是正整数且 filterCol 非空
	// 防止配置错误导致过滤静默失效或全表降级
	if err := validateStepFilters(rawConfig.Left.Steps, "左链"); err != nil {
		return err
	}
	if err := validateStepFilters(rawConfig.Right.Steps, "右链"); err != nil {
		return err
	}

	// 步骤 8: 校验预警窗口参数（chainWarnBefore/chainWarnSheet/chainWarnCol）
	// 三参数必须同时配置或同时为空
	warnBeforeStr := ctx.Params["chainWarnBefore"]
	warnSheet := ctx.Params["chainWarnSheet"]
	warnCol := ctx.Params["chainWarnCol"]
	if warnBeforeStr != "" || warnSheet != "" || warnCol != "" {
		if warnBeforeStr == "" {
			return fmt.Errorf("预警窗口参数 chainWarnBefore 未配置（需与 chainWarnSheet、chainWarnCol 同时配置）")
		}
		if warnSheet == "" {
			return fmt.Errorf("预警窗口参数 chainWarnSheet 未配置（需与 chainWarnBefore、chainWarnCol 同时配置）")
		}
		if warnCol == "" {
			return fmt.Errorf("预警窗口参数 chainWarnCol 未配置（需与 chainWarnBefore、chainWarnSheet 同时配置）")
		}
		if _, err := time.ParseDuration(warnBeforeStr); err != nil {
			return fmt.Errorf("预警窗口参数 chainWarnBefore 格式错误: %w（示例: 168h 表示7天）", err)
		}
	}

	return nil
}

// validateStepFilters 校验链中各步骤的过滤条件配置合法
// 参数 sideLabel: "左链" 或 "右链"，用于错误信息定位
//
// 校验规则：
//  1. filterMode 必须是已知类型（""/"multi"/"withinDays"）
//  2. filterMode="withinDays" 时：
//     - filterCol 必须非空（指定时间列）
//     - filterDays 必须是正整数（去除空白后）
func validateStepFilters(steps []ChainStep, sideLabel string) error {
	for i, step := range steps {
		if !knownFilterModes[step.FilterMode] {
			return fmt.Errorf("%s步骤%d: 未知的过滤模式 filterMode=%q（合法值: \"\", \"multi\", \"withinDays\"）",
				sideLabel, i+1, step.FilterMode)
		}
		if step.FilterMode == "withinDays" {
			if step.FilterCol == "" {
				return fmt.Errorf("%s步骤%d: filterMode=\"withinDays\" 时 filterCol 不能为空（需指定时间列）",
					sideLabel, i+1)
			}
			days, err := strconv.Atoi(strings.TrimSpace(step.FilterDays))
			if err != nil {
				return fmt.Errorf("%s步骤%d: filterMode=\"withinDays\" 时 filterDays 必须是正整数，当前值=%q",
					sideLabel, i+1, step.FilterDays)
			}
			if days <= 0 {
				return fmt.Errorf("%s步骤%d: filterMode=\"withinDays\" 时 filterDays 必须>0，当前值=%d",
					sideLabel, i+1, days)
			}
		}
	}
	return nil
}
