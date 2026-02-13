package config

import (
	"fmt"
	"testing"
)

// TestParseCondition 测试条件表达式解析
func TestParseCondition(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		// 基本比较操作符
		{"type>0", false},
		{"type=1", false},
		{"type==1", false},
		{"type!=0", false},
		{"level>=10", false},
		{"level<=5", false},
		{"level<100", false},

		// 逻辑操作符
		{"type>0 & level>10", false},
		{"type>0 && level>10", false},
		{"type>0 and level>10", false},
		{"type=1 | type=3", false},
		{"type=1 || type=3", false},
		{"type=1 or type=3", false},

		// 一元操作符
		{"!enabled", false},
		{"not enabled", false},

		// in 操作符
		{"type in [1,2,3]", false},
		{"type in [1]", false},

		// between 操作符
		{"level between 10,20", false},

		// 括号分组
		{"(type>0) & level>10", false},

		// 错误情况
		{"", true},           // 空条件
		{"type>", true},       // 缺少右操作数
		{"in [1,2,3]", true}, // in 操作符缺少左操作数
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseCondition(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCondition(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestConditionEvaluate 测试条件评估
func TestConditionEvaluate(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		values map[string]interface{}
		want   bool
	}{
		{
			name: "大于",
			expr: "level>10",
			values: map[string]interface{}{"level": 15},
			want:   true,
		},
		{
			name: "大于-不满足",
			expr: "level>10",
			values: map[string]interface{}{"level": 5},
			want:   false,
		},
		{
			name: "等于",
			expr: "type=1",
			values: map[string]interface{}{"type": 1},
			want:   true,
		},
		{
			name: "等于-不满足",
			expr: "type=1",
			values: map[string]interface{}{"type": 2},
			want:   false,
		},
		{
			name: "逻辑与",
			expr: "type>0 & level>10",
			values: map[string]interface{}{"type": 1, "level": 15},
			want:   true,
		},
		{
			name: "逻辑与-不满足",
			expr: "type>0 & level>10",
			values: map[string]interface{}{"type": 1, "level": 5},
			want:   false,
		},
		{
			name: "逻辑或",
			expr: "type=1 | type=3",
			values: map[string]interface{}{"type": 1},
			want:   true,
		},
		{
			name: "逻辑或-第二个满足",
			expr: "type=1 | type=3",
			values: map[string]interface{}{"type": 3},
			want:   true,
		},
		{
			name: "逻辑或-不满足",
			expr: "type=1 | type=3",
			values: map[string]interface{}{"type": 2},
			want:   false,
		},
		{
			name: "逻辑非",
			expr: "!enabled",
			values: map[string]interface{}{"enabled": false},
			want:   true,
		},
		{
			name: "逻辑非-不满足",
			expr: "!enabled",
			values: map[string]interface{}{"enabled": true},
			want:   false,
		},
		{
			name: "in操作符",
			expr: "type in [1,2,3]",
			values: map[string]interface{}{"type": 2},
			want:   true,
		},
		{
			name: "in操作符-不满足",
			expr: "type in [1,2,3]",
			values: map[string]interface{}{"type": 4},
			want:   false,
		},
		{
			name: "between操作符",
			expr: "level between 10,20",
			values: map[string]interface{}{"level": 15},
			want:   true,
		},
		{
			name: "between操作符-边界",
			expr: "level between 10,20",
			values: map[string]interface{}{"level": 10},
			want:   true,
		},
		{
			name: "between操作符-不满足",
			expr: "level between 10,20",
			values: map[string]interface{}{"level": 5},
			want:   false,
		},
		{
			name: "复杂表达式",
			expr: "(type>0 & level>10) | type=0",
			values: map[string]interface{}{"type": 1, "level": 15},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, err := ParseCondition(tt.expr)
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			ctx := NewEvalContext()
			for k, v := range tt.values {
				ctx.SetValue(k, v)
			}

			result, err := cond.Evaluate(ctx)
			if err != nil {
				t.Fatalf("评估失败: %v", err)
			}

			if result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestDependentFields 测试依赖字段提取
func TestDependentFields(t *testing.T) {
	tests := []struct {
		name             string
		expr             string
		wantFields       []string
	}{
		{
			name:       "单字段",
			expr:       "level>10",
			wantFields: []string{"level"},
		},
		{
			name:       "多字段",
			expr:       "type>0 & level>10",
			wantFields: []string{"type", "level"},
		},
		{
			name:       "重复字段",
			expr:       "type>0 & type<10",
			wantFields: []string{"type", "type"},
		},
		{
			name:       "括号分组",
			expr:       "(type>0 | type=0) & level>10",
			wantFields: []string{"type", "type", "level"},
		},
		{
			name:       "in操作符",
			expr:       "type in [1,2,3]",
			wantFields: []string{"type"},
		},
		{
			name:       "between操作符",
			expr:       "level between 10,20",
			wantFields: []string{"level"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, err := ParseCondition(tt.expr)
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			fields := cond.DependentFields()
			if len(fields) != len(tt.wantFields) {
				t.Errorf("DependentFields() 长度 = %d, want %d", len(fields), len(tt.wantFields))
			}

			for i, field := range fields {
				if i >= len(tt.wantFields) || field != tt.wantFields[i] {
					t.Errorf("DependentFields()[%d] = %s, want %s", i, field, tt.wantFields[i])
				}
			}
		})
	}
}

// TestConditionString 测试条件字符串表示
func TestConditionString(t *testing.T) {
	tests := []struct {
		expr  string
		want  string
	}{
		{"level>10", "(level > 10)"},
		{"type=1", "(type = 1)"},
		{"type>0 & level>10", "((type > 0) & (level > 10))"},
		{"type=1 | type=3", "((type = 1) | (type = 3))"},
		{"!enabled", "!enabled"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			cond, err := ParseCondition(tt.expr)
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			got := cond.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestToBool 测试布尔值转换
func TestToBool(t *testing.T) {
	tests := []struct {
		val  interface{}
		want bool
	}{
		{true, true},
		{false, false},
		{0, false},
		{1, true},
		{-1, true},
		{0.0, false},
		{1.5, true},
		{"", false},
		{"hello", true},
		{nil, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.val), func(t *testing.T) {
			got := toBool(tt.val)
			if got != tt.want {
				t.Errorf("toBool(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

// TestToFloat64 测试浮点数转换
func TestToFloat64(t *testing.T) {
	tests := []struct {
		val    interface{}
		want   float64
		wantErr bool
	}{
		{0, 0, false},
		{1, 1, false},
		{-1, -1, false},
		{0.0, 0, false},
		{1.5, 1.5, false},
		{"3.14", 3.14, false},
		{"hello", 0, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.val), func(t *testing.T) {
			got, err := toFloat64(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("toFloat64(%v) error = %v, wantErr %v", tt.val, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

// TestCompareValues 测试值比较
func TestCompareValues(t *testing.T) {
	tests := []struct {
		a, b  interface{}
		want  int
	}{
		{1, 2, -1},
		{2, 1, 1},
		{1, 1, 0},
		{1.5, 2.5, -1},
		{"a", "b", -1},
		{"b", "a", 1},
		{"a", "a", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v,%v", tt.a, tt.b), func(t *testing.T) {
			got := compareValues(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareValues(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
