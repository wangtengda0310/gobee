package params

import (
	"strconv"
	"strings"
)

// IterationConfig 字段迭代配置（内部使用，不直接暴露给前端）
// 前端使用 FieldValues（含 OriginalValue/RangeValue/EnumValue/ComboValue/InputType），
// 后端内部通过 FieldValuesToIterationConfig 转换为本结构。
type IterationConfig struct {
	Type    string // "range", "enum", "combo", "original"（与前端 input_type 一致）
	Start   *int   // 范围起始值（仅 range）
	End     *int   // 范围终值（仅 range）
	Step    *int   // 范围步长（仅 range）
	Values  []any  // 枚举值或组合值列表（enum/compose）
	Field   string // 字段名（顶层 key）
	VarName string // 变量名（仅 variable 类型），对应 variable 注册表中的 ShortName
}

// GenerateIterativeMessages 生成分组独立迭代消息列表
//
// 算法：各类型字段独立迭代，不交叉组合：
//   - range 字段：每个 range 字段逐值迭代，其他字段保持原始值
//   - enum 字段：每个 enum 字段逐值迭代，其他字段保持原始值
//   - compose 字段：仅 compose 字段之间做笛卡尔积，其他字段保持原始值
//
// 总消息数 = Σ(range_counts) + Σ(enum_counts) + compose_cartesian_product
//
// 注意：当前仅支持顶层字段迭代。嵌套路径迭代需要扩展 Field 为路径格式
// 并实现深拷贝和路径感知的 setter。
func GenerateIterativeMessages(baseMessage map[string]any, iterationConfig map[string]IterationConfig) []map[string]any {
	var messages []map[string]any

	// 按 field 首字母排序确保顺序稳定
	orderedFields := sortedFields(iterationConfig)

	// 分组：range/enum 各自独立迭代，combo 之间做笛卡尔积
	var rangeFields, enumFields, composeFields, variableFields []string
	for _, field := range orderedFields {
		cfg := iterationConfig[field]
		switch cfg.Type {
		case "range":
			rangeFields = append(rangeFields, field)
		case "enum":
			enumFields = append(enumFields, field)
		case "combo":
			composeFields = append(composeFields, field)
		case "variable":
			variableFields = append(variableFields, field)
		}
	}

	// range 字段独立迭代：每个 range 字段生成一组消息
	for _, field := range rangeFields {
		cfg := iterationConfig[field]
		values := expandRange(cfg)
		for _, v := range values {
			msg := copyBaseWithOverride(baseMessage, field, v)
			messages = append(messages, msg)
		}
	}

	// enum 字段独立迭代：每个 enum 字段生成一组消息
	for _, field := range enumFields {
		cfg := iterationConfig[field]
		for _, v := range cfg.Values {
			msg := copyBaseWithOverride(baseMessage, field, v)
			messages = append(messages, msg)
		}
	}

	// compose 字段做笛卡尔积（仅 compose 字段之间）
	if len(composeFields) > 0 {
		var composeConfigs []IterationConfig
		for _, field := range composeFields {
			composeConfigs = append(composeConfigs, iterationConfig[field])
		}
		combinations := cartesianProduct(composeConfigs)
		for _, combo := range combinations {
			msg := copyBase(baseMessage)
			for i, field := range composeFields {
				if i < len(combo) {
					msg[field] = combo[i]
				}
			}
			messages = append(messages, msg)
		}
	}

	// variable 字段：保持原始 payload 不做替换，运行时由 ResolveMessageVariables 根据 FieldValues 元数据解析
	// 一条消息中存在多个 variable 字段时，仍只生成一条消息，避免同一请求被重复发送
	if len(variableFields) > 0 {
		messages = append(messages, copyBase(baseMessage))
	}

	// 无迭代配置时返回原始消息
	if len(messages) == 0 {
		messages = append(messages, copyBase(baseMessage))
	}

	return messages
}

// expandRange 展开范围配置为值列表
func expandRange(cfg IterationConfig) []any {
	if cfg.Start == nil || cfg.End == nil {
		return nil
	}
	start := *cfg.Start
	end := *cfg.End
	step := 1
	if cfg.Step != nil && *cfg.Step > 0 {
		step = *cfg.Step
	}
	if start > end {
		return nil
	}
	var values []any
	for v := start; v <= end; v += step {
		values = append(values, v)
	}
	return values
}

// copyBase 浅拷贝基础消息
func copyBase(base map[string]any) map[string]any {
	msg := make(map[string]any, len(base))
	for k, v := range base {
		msg[k] = v
	}
	return msg
}

// copyBaseWithOverride 浅拷贝基础消息并覆盖指定字段
func copyBaseWithOverride(base map[string]any, field string, value any) map[string]any {
	msg := copyBase(base)
	msg[field] = value
	return msg
}

// sortedFields 返回按字母排序的字段名列表
func sortedFields(config map[string]IterationConfig) []string {
	fields := make([]string, 0, len(config))
	for f := range config {
		fields = append(fields, f)
	}
	for i := 1; i < len(fields); i++ {
		for j := i; j > 0 && fields[j] < fields[j-1]; j-- {
			fields[j], fields[j-1] = fields[j-1], fields[j]
		}
	}
	return fields
}

// cartesianProduct 笛卡尔积生成（仅用于 compose 字段之间）
func cartesianProduct(configs []IterationConfig) [][]any {
	if len(configs) == 0 {
		return [][]any{}
	}

	var result [][]any
	var current []any

	var generate func(idx int)
	generate = func(idx int) {
		if idx == len(configs) {
			cp := make([]any, len(current))
			copy(cp, current)
			result = append(result, cp)
			return
		}

		for _, v := range configs[idx].Values {
			current = append(current, v)
			generate(idx + 1)
			current = current[:len(current)-1]
		}
	}

	generate(0)
	return result
}

// FieldValuesToIterationConfig 将前端 FieldValues 转为内部 IterationConfig
// 过滤掉 input_type=original 的字段（不参与迭代）
func FieldValuesToIterationConfig(fv map[string]FieldValues) map[string]IterationConfig {
	result := make(map[string]IterationConfig)
	for field, v := range fv {
		if v.InputType == "original" {
			continue
		}
		cfg := IterationConfig{
			Type:  v.InputType,
			Field: field,
		}
		switch v.InputType {
		case "range":
			cfg.Start = &v.RangeValue.Start
			cfg.End = &v.RangeValue.End
			cfg.Step = &v.RangeValue.Step
		case "enum":
			cfg.Values = v.EnumValue
		case "combo":
			cfg.Values = v.ComboValue
		case "variable":
			cfg.VarName = v.VariableName
		}
		result[field] = cfg
	}
	return result
}

// ParseValueString 将字符串尝试解析为 JSON 类型（bool > float64 > string）
func ParseValueString(s string) any {
	// bool
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	// 数值
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		// 整数去掉小数点，保持 JSON 兼容
		if v == float64(int64(v)) && !strings.Contains(s, ".") {
			return float64(int64(v))
		}
		return v
	}
	return s
}
