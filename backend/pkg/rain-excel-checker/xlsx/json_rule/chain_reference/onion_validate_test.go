// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件测试洋葱模型参数校验处理器（ValidateHandler）
package chain_reference

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== 测试辅助 ====================

// validParams 返回一组合法的规则参数
// 包含完整的 chainSteps 配置和比较类型
func validParams() map[string]string {
	return map[string]string{
		"chainSteps": `{
			"left": {
				"steps": [
					{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "OnceDropRule"},
					{"sheet": "掉落规则表|DropRule", "preCol": "Id", "findVal": "self", "nextCol": "DropGroup"}
				]
			},
			"right": {
				"steps": [
					{"sheet": "掉落道具表|DropItem", "preCol": "Item", "findVal": "self", "nextCol": "DropGroup"},
					{"sheet": "掉落分组表|DropGroup", "preCol": "Id", "findVal": "self", "nextCol": "Id"}
				]
			}
		}`,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}
}

// validConfig 返回一组合法的两链配置
func validConfig() *ChainPairConfig {
	return &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
				{Sheet: "掉落规则表|DropRule", PreCol: "Id", FindVal: "self", NextCol: "DropGroup"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "掉落道具表|DropItem", PreCol: "Item", FindVal: "self", NextCol: "DropGroup"},
				{Sheet: "掉落分组表|DropGroup", PreCol: "Id", FindVal: "self", NextCol: "Id"},
			},
		},
	}
}

// makeCtx 创建测试用的 ChainContext
// params 和 config 可选传入，不传时使用合法默认值
func makeCtx(params map[string]string, config *ChainPairConfig) *ChainContext {
	if params == nil {
		params = validParams()
	}
	if config == nil {
		config = validConfig()
	}
	return NewChainContext(nil, nil, 0, 0, 0, nil, config, params, []string{"test"}, 0)
}

// ==================== 校验测试 ====================

// TestValidate_ValidConfig 合法配置通过校验
func TestValidate_ValidConfig(t *testing.T) {
	ctx := makeCtx(nil, nil)
	handler := &ValidateHandler{}

	// next 应该被调用（传入一个 marker next 函数验证）
	nextCalled := false
	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err, "合法配置不应返回错误")
	assert.True(t, nextCalled, "合法配置应调用 next 函数")
}

// TestValidate_EmptyChainSteps chainSteps 参数为空时报错
func TestValidate_EmptyChainSteps(t *testing.T) {
	params := map[string]string{
		"chainSteps":   "",
		"chainCompare": "verify_exists",
	}
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "chainSteps 为空应返回错误")
	assert.Contains(t, err.Error(), "chainSteps")
}

// TestValidate_InvalidJSON chainSteps JSON 格式错误时报错
func TestValidate_InvalidJSON(t *testing.T) {
	params := map[string]string{
		"chainSteps":   `{invalid json}`,
		"chainCompare": "verify_exists",
	}
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "无效 JSON 应返回错误")
	assert.Contains(t, err.Error(), "JSON")
}

// TestValidate_BothChainsEmpty 两条链都没有步骤时报错
func TestValidate_BothChainsEmpty(t *testing.T) {
	params := map[string]string{
		"chainSteps":   `{"left": {"steps": []}, "right": {"steps": []}}`,
		"chainCompare": "verify_exists",
	}
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "两条链都无步骤应返回错误")
	assert.Contains(t, err.Error(), "两条链都没有步骤")
}

// TestValidate_UnknownCompareType 未知的比较类型报错
func TestValidate_UnknownCompareType(t *testing.T) {
	params := validParams()
	params["chainCompare"] = "unknown_type"
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "未知比较类型应返回错误")
	assert.Contains(t, err.Error(), "未知的比较类型")
}

// TestValidate_UnknownMatchType 未知的匹配类型报错
func TestValidate_UnknownMatchType(t *testing.T) {
	params := validParams()
	params["chainMatchCompare"] = "unknown_match"
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "未知匹配类型应返回错误")
	assert.Contains(t, err.Error(), "未知的匹配类型")
}

// TestValidate_LeftStepSheetEmpty 左链第一步 Sheet 为空是合法的（仅取值模式）
func TestValidate_LeftStepSheetEmpty(t *testing.T) {
	params := validParams()
	// 左链第一步 sheet 为空是合法的（表示从当前表取值）
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	nextCalled := false
	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err, "左链第一步 Sheet 为空应合法")
	assert.True(t, nextCalled, "应调用 next 函数")
}

// TestValidate_InvalidRegex 正则表达式编译失败时报错
func TestValidate_InvalidRegex(t *testing.T) {
	params := map[string]string{
		"chainSteps": `{
			"left": {
				"steps": [
					{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "Col1", "pattern": "[invalid"}
				]
			},
			"right": {
				"steps": [
					{"sheet": "SomeSheet", "preCol": "Id", "findVal": "self", "nextCol": "Name"}
				]
			}
		}`,
		"chainCompare": "verify_exists",
	}
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "无效正则应返回错误")
	assert.Contains(t, err.Error(), "正则编译失败")
}

// TestValidate_MissingRequiredSheet 右链步骤缺少 Sheet 报错
func TestValidate_MissingRequiredSheet(t *testing.T) {
	params := map[string]string{
		"chainSteps": `{
			"left": {
				"steps": [
					{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "Col1"}
				]
			},
			"right": {
				"steps": [
					{"sheet": "", "preCol": "Id", "findVal": "self", "nextCol": "Name"}
				]
			}
		}`,
		"chainCompare": "verify_exists",
	}
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "右链步骤 Sheet 为空应返回错误")
	assert.Contains(t, err.Error(), "目标表")
}

// TestValidate_WarnBefore_AllParamsValid 预警参数三参数同时配置合法
func TestValidate_WarnBefore_AllParamsValid(t *testing.T) {
	params := validParams()
	params["chainWarnBefore"] = "168h"
	params["chainWarnSheet"] = "SeasonPass"
	params["chainWarnCol"] = "StartTime"
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	nextCalled := false
	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err, "预警参数三参数同时配置应合法")
	assert.True(t, nextCalled, "应调用 next 函数")
}

// TestValidate_WarnBefore_MissingOneParam 预警参数缺少一个报错
func TestValidate_WarnBefore_MissingOneParam(t *testing.T) {
	params := validParams()
	params["chainWarnBefore"] = "168h"
	params["chainWarnSheet"] = "SeasonPass"
	// 缺少 chainWarnCol
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "缺少 chainWarnCol 应返回错误")
	assert.Contains(t, err.Error(), "chainWarnCol")
}

// TestValidate_WarnBefore_InvalidDuration 格式错误报错
func TestValidate_WarnBefore_InvalidDuration(t *testing.T) {
	params := validParams()
	params["chainWarnBefore"] = "7days"
	params["chainWarnSheet"] = "SeasonPass"
	params["chainWarnCol"] = "StartTime"
	ctx := makeCtx(params, nil)
	handler := &ValidateHandler{}

	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		return nil
	})

	assert.Error(t, err, "无效 duration 格式应返回错误")
	assert.Contains(t, err.Error(), "chainWarnBefore")
}

// ==================== filterMode/filterDays 校验测试 ====================

// buildStepFilterParams 构造一个最简合法 chainSteps 但替换左链第一步的 filterMode/filterDays/filterCol
func buildStepFilterParams(filterMode, filterDays, filterCol string) map[string]string {
	chainSteps := `{"left":{"steps":[{"sheet":"","preCol":"","findVal":"col","nextCol":"X","filterCol":"` + filterCol + `","filterMode":"` + filterMode + `","filterDays":"` + filterDays + `"}]},"right":{"steps":[{"sheet":"Y","preCol":"P","findVal":"self","nextCol":"N"}]}}`
	return map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
	}
}

// TestValidate_FilterMode_Valid 合法 filterMode 应通过
func TestValidate_FilterMode_Valid(t *testing.T) {
	for _, mode := range []string{"", "multi", "withinDays"} {
		filterDays := "7"
		filterCol := "StartTime"
		if mode == "" || mode == "multi" {
			filterDays = "" // 非 withinDays 不要求 filterDays
		}
		ctx := &ChainContext{Params: buildStepFilterParams(mode, filterDays, filterCol)}
		err := validateConfig(ctx)
		assert.NoError(t, err, "合法 filterMode=%q 不应报错", mode)
	}
}

// TestValidate_FilterMode_Unknown 未知 filterMode 应报错
func TestValidate_FilterMode_Unknown(t *testing.T) {
	ctx := &ChainContext{Params: buildStepFilterParams("invalidMode", "", "StartTime")}
	err := validateConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未知的过滤模式")
	assert.Contains(t, err.Error(), "invalidMode")
}

// TestValidate_WithinDays_MissingFilterCol withinDays 模式 filterCol 为空应报错
func TestValidate_WithinDays_MissingFilterCol(t *testing.T) {
	ctx := &ChainContext{Params: buildStepFilterParams("withinDays", "7", "")}
	err := validateConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filterCol 不能为空")
}

// TestValidate_WithinDays_InvalidFilterDays 非数字 filterDays 应报错
func TestValidate_WithinDays_InvalidFilterDays(t *testing.T) {
	for _, days := range []string{"", "abc", "1.5", "  "} {
		ctx := &ChainContext{Params: buildStepFilterParams("withinDays", days, "StartTime")}
		err := validateConfig(ctx)
		assert.Error(t, err, "filterDays=%q 应报错", days)
		assert.Contains(t, err.Error(), "filterDays 必须是正整数")
	}
}

// TestValidate_WithinDays_NonPositiveFilterDays 0 或负数 filterDays 应报错
func TestValidate_WithinDays_NonPositiveFilterDays(t *testing.T) {
	for _, days := range []string{"0", "-1", "-100"} {
		ctx := &ChainContext{Params: buildStepFilterParams("withinDays", days, "StartTime")}
		err := validateConfig(ctx)
		assert.Error(t, err, "filterDays=%q 应报错", days)
		assert.Contains(t, err.Error(), "filterDays 必须>0")
	}
}

// TestValidate_WithinDays_RightStep 右链步骤的 filterMode 同样校验
func TestValidate_WithinDays_RightStep(t *testing.T) {
	chainSteps := `{"left":{"steps":[{"sheet":"","preCol":"","findVal":"col","nextCol":"X"}]},"right":{"steps":[{"sheet":"Y","preCol":"P","findVal":"self","nextCol":"N","filterCol":"","filterMode":"withinDays","filterDays":"7"}]}}`
	ctx := &ChainContext{Params: map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
	}}
	err := validateConfig(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "右链步骤1")
	assert.Contains(t, err.Error(), "filterCol 不能为空")
}
