package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateIterativeMessages(t *testing.T) {
	tests := []struct {
		name            string
		baseMessage     map[string]any
		iterationConfig map[string]IterationConfig
		expectedCount   int
		validateFunc    func(*testing.T, []map[string]any)
	}{
		{
			name: "纯范围迭代",
			baseMessage: map[string]any{
				"a": 11, "b": 22, "c": 33, "d": 44,
			},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "range", Start: intPtr(1), End: intPtr(5), Step: intPtr(1), Field: "a"},
				"b": {Type: "original", Field: "b"},
				"c": {Type: "original", Field: "c"},
				"d": {Type: "original", Field: "d"},
			},
			expectedCount: 5,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Equal(t, 1, messages[0]["a"])
				assert.Equal(t, 22, messages[0]["b"])
				assert.Equal(t, 5, messages[4]["a"])
			},
		},
		{
			name: "纯枚举迭代",
			baseMessage: map[string]any{
				"a": 11, "b": 22, "c": 33, "d": 44,
			},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "original", Field: "a"},
				"b": {Type: "enum", Values: []any{"x", "y", "z"}, Field: "b"},
				"c": {Type: "original", Field: "c"},
				"d": {Type: "original", Field: "d"},
			},
			expectedCount: 3,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Equal(t, "x", messages[0]["b"])
				assert.Equal(t, "z", messages[2]["b"])
			},
		},
		{
			name: "纯组合笛卡尔积",
			baseMessage: map[string]any{
				"a": 11, "b": 22, "c": 33, "d": 44,
			},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "original", Field: "a"},
				"b": {Type: "original", Field: "b"},
				"c": {Type: "combo", Values: []any{10, 20}, Field: "c"},
				"d": {Type: "combo", Values: []any{100, 200}, Field: "d"},
			},
			expectedCount: 4,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Equal(t, 10, messages[0]["c"])
				assert.Equal(t, 200, messages[3]["d"])
			},
		},
		{
			name: "分组独立迭代 - range+enum+compose",
			baseMessage: map[string]any{
				"a": 11, "b": 22, "c": 33, "d": 44,
			},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "range", Start: intPtr(1), End: intPtr(5), Step: intPtr(1), Field: "a"},
				"b": {Type: "enum", Values: []any{"x", "y", "z"}, Field: "b"},
				"c": {Type: "combo", Values: []any{10, 20}, Field: "c"},
				"d": {Type: "combo", Values: []any{100, 200}, Field: "d"},
			},
			expectedCount: 12,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				for i := 0; i < 5; i++ {
					assert.Equal(t, i+1, messages[i]["a"])
				}
			},
		},
		{
			name:        "无迭代配置 - 返回原始消息",
			baseMessage: map[string]any{"a": 11, "b": 22, "c": 33, "d": 44},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "original", Field: "a"},
				"b": {Type: "original", Field: "b"},
				"c": {Type: "original", Field: "c"},
				"d": {Type: "original", Field: "d"},
			},
			expectedCount: 1,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Equal(t, 11, messages[0]["a"])
			},
		},
		{
			name:        "多个range字段 - 各自独立迭代",
			baseMessage: map[string]any{"a": 0, "b": 0},
			iterationConfig: map[string]IterationConfig{
				"a": {Type: "range", Start: intPtr(1), End: intPtr(3), Step: intPtr(1), Field: "a"},
				"b": {Type: "range", Start: intPtr(10), End: intPtr(12), Step: intPtr(1), Field: "b"},
			},
			expectedCount: 6,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Equal(t, 1, messages[0]["a"])
				assert.Equal(t, 12, messages[5]["b"])
			},
		},
		{
			name: "多个 variable 字段只生成一条消息",
			baseMessage: map[string]any{
				"beLookerID": map[string]any{"__var__": "roomCreator"},
				"roomID":     map[string]any{"__var__": "roomID"},
			},
			iterationConfig: map[string]IterationConfig{
				"beLookerID": {Type: "variable", VarName: "roomCreator", Field: "beLookerID"},
				"roomID":     {Type: "variable", VarName: "roomID", Field: "roomID"},
			},
			expectedCount: 1,
			validateFunc: func(t *testing.T, messages []map[string]any) {
				assert.Len(t, messages, 1)
				assert.Equal(t, map[string]any{"__var__": "roomCreator"}, messages[0]["beLookerID"])
				assert.Equal(t, map[string]any{"__var__": "roomID"}, messages[0]["roomID"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := GenerateIterativeMessages(tt.baseMessage, tt.iterationConfig)
			assert.Equal(t, tt.expectedCount, len(messages), "消息数量不匹配")
			tt.validateFunc(t, messages)
		})
	}
}

func TestCartesianProduct(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []IterationConfig
		expected [][]any
	}{
		{
			name: "单个配置",
			inputs: []IterationConfig{
				{Type: "combo", Values: []any{1, 2, 3}, Field: "x"},
			},
			expected: [][]any{{1}, {2}, {3}},
		},
		{
			name: "两个配置笛卡尔积",
			inputs: []IterationConfig{
				{Type: "combo", Values: []any{10, 20}, Field: "c"},
				{Type: "combo", Values: []any{100, 200}, Field: "d"},
			},
			expected: [][]any{{10, 100}, {10, 200}, {20, 100}, {20, 200}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cartesianProduct(tt.inputs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldValuesToIterationConfig(t *testing.T) {
	fv := map[string]FieldValues{
		"a": {InputType: "range", RangeValue: RangeValue{Start: 1, Step: 1, End: 5}},
		"b": {InputType: "enum", EnumValue: []any{"x", "y", "z"}},
		"c": {InputType: "original"},
		"d": {InputType: "combo", ComboValue: []any{10, 20}},
	}

	result := FieldValuesToIterationConfig(fv)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.NotContains(t, result, "c")
	assert.Equal(t, "range", result["a"].Type)
	assert.Equal(t, 1, *result["a"].Start)
}

func TestVariableTypeConfig(t *testing.T) {
	fv := map[string]FieldValues{
		"cityId": {InputType: "variable", VariableName: "cityId"},
		"a":      {InputType: "original"},
	}

	result := FieldValuesToIterationConfig(fv)

	assert.Len(t, result, 1)
	assert.Equal(t, "variable", result["cityId"].Type)
	assert.Equal(t, "cityId", result["cityId"].VarName)

	baseMessage := map[string]any{"cityId": map[string]any{"__var__": "cityId"}, "a": 1, "b": 2}
	messages := GenerateIterativeMessages(baseMessage, result)
	assert.Len(t, messages, 1)
	assert.Equal(t, map[string]any{"__var__": "cityId"}, messages[0]["cityId"])
}

func TestVariableWithCombo(t *testing.T) {
	baseMessage := map[string]any{
		"cityId": map[string]any{"__var__": "cityId"},
		"count":  0,
		"type":   "",
	}

	config := map[string]IterationConfig{
		"cityId": {Type: "variable", VarName: "cityId", Field: "cityId"},
		"count":  {Type: "range", Start: intPtr(1), End: intPtr(3), Step: intPtr(1), Field: "count"},
		"type":   {Type: "enum", Values: []any{"A", "B"}, Field: "type"},
	}

	messages := GenerateIterativeMessages(baseMessage, config)
	assert.Len(t, messages, 6) // 3(range) + 2(enum) + 1(variable)
}

func TestVariablePlaceholderInjection(t *testing.T) {
	baseMessage := map[string]any{"cityId": uint64(0), "count": 10}
	config := map[string]IterationConfig{
		"cityId": {Type: "variable", VarName: "cityId", Field: "cityId"},
	}

	messages := GenerateIterativeMessages(baseMessage, config)
	assert.Len(t, messages, 1)
	assert.Equal(t, uint64(0), messages[0]["cityId"])
	assert.Equal(t, 10, messages[0]["count"])
}

func TestVariablePlaceholderInjectionWithExistingPlaceholder(t *testing.T) {
	baseMessage := map[string]any{
		"cityId": map[string]any{"__var__": "cityId"},
		"count":  10,
	}

	config := map[string]IterationConfig{
		"cityId": {Type: "variable", VarName: "cityId", Field: "cityId"},
	}

	messages := GenerateIterativeMessages(baseMessage, config)
	assert.Len(t, messages, 1)
	assert.Equal(t, map[string]any{"__var__": "cityId"}, messages[0]["cityId"])
}

func intPtr(v int) *int { return &v }
