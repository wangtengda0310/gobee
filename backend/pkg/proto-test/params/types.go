// Package params 提供参数迭代和变量定义的纯数据类型与算法。
//
// 职责边界：
//   - RangeValue / FieldValues：前端 4 态输入的数据结构
//   - IterationConfig / GenerateIterativeMessages：字段迭代展开算法
//   - VariableDef / builtinVariables：变量注册表和查找
//
// 本包不依赖 streamproto（帧解码、连接管理）或 streamproxy（Wails 服务层），
// 避免循环导入。运行时的 FrameMux 交互和 proto 提取留在 streamproto/variable.go。
package params

// RangeValue 范围输入值（对应前端 range-input 组件）
type RangeValue struct {
	Start int `json:"start"`
	Step  int `json:"step"`
	End   int `json:"end"`
}

// FieldValues 字段4态值（对应前端 FieldFourState 接口）
// 每个字段路径对应一组完整的输入状态：原始值、范围、枚举、组合
type FieldValues struct {
	RangeValue   RangeValue `json:"range_value"`
	EnumValue    []any      `json:"enum_value"`
	ComboValue   []any      `json:"combo_value"`
	InputType    string     `json:"input_type"`              // "original" | "range" | "enum" | "combo" | "variable"
	VariableName string     `json:"variable_name,omitempty"` // 变量短名（仅 variable 类型），如 "cityId"
}
