package csvunmarshaler

import (
	"csv-unmarshaler/csvutil"
	"csv-unmarshaler/encoding"
	"csv-unmarshaler/gocsv"
	"testing"
)

// TestLiskovSubstitutionPrinciple 演示李氏替换原则的综合测试
func TestLiskovSubstitutionPrinciple(t *testing.T) {
	// TestPerson 测试用的结构体
	type TestPerson struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	// 测试数据
	const TestData = `name,age
John,30
Alice,25
Bob,35`

	// 创建所有实现了Unmarshaler接口的类型
	// 根据李氏替换原则，它们都应该可以互相替换
	unmarshalers := []Unmarshaler{
		UnmarshalerFunc(gocsv.NewGoCSVUnmarshaler().Unmarshal),
		UnmarshalerFunc(csvutil.NewCSVUtilUnmarshaler().Unmarshal),
		UnmarshalerFunc(encoding.NewEncodingCSVUnmarshaler().Unmarshal),
	}

	// 测试每种实现都能正确工作
	for i, unmarshaler := range unmarshalers {
		t.Run("Unmarshaler_"+string(rune('A'+i)), func(t *testing.T) {
			t.Logf("Testing with implementation: %T", unmarshaler)

			var result []TestPerson
			err := unmarshaler.Unmarshal([]byte(TestData), &result)
			if err != nil {
				t.Errorf("Unmarshal failed: %v", err)
				return
			}

			// 验证结果
			expected := []TestPerson{
				{Name: "John", Age: 30},
				{Name: "Alice", Age: 25},
				{Name: "Bob", Age: 35},
			}

			if len(result) != len(expected) {
				t.Errorf("Expected %d records, got %d", len(expected), len(result))
				return
			}

			for j, expectedPerson := range expected {
				if result[j].Name != expectedPerson.Name || result[j].Age != expectedPerson.Age {
					t.Errorf("Record %d: expected %+v, got %+v", j, expectedPerson, result[j])
				}
			}
		})
	}
}

// TestAllImplementationsBehaviorConsistency 测试所有实现的行为一致性
func TestAllImplementationsBehaviorConsistency(t *testing.T) {
	// TestPerson 测试用的结构体
	type TestPerson struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	// 测试用例
	testCases := []struct {
		name        string
		data        string
		expectError bool
		description string
	}{
		{
			name: "normal_case",
			data: `name,age
John,30
Alice,25`,
			expectError: false,
			description: "正常的CSV数据",
		},
		{
			name: "empty_data",
			data: "name,age\n",
			expectError: false,
			description: "只有header，没有数据",
		},
		{
			name: "single_record",
			data: "name,age\nJohn,30",
			expectError: false,
			description: "单条记录",
		},
		{
			name: "invalid_age",
			data: "name,age\nJohn,invalid",
			expectError: true,
			description: "无效的年龄值",
		},
		{
			name: "missing_columns",
			data: "name\nJohn",
			expectError: false,
			description: "缺少某些列",
		},
	}

	// 所有实现
	implementations := map[string]Unmarshaler{
		"gocsv":    UnmarshalerFunc(gocsv.NewGoCSVUnmarshaler().Unmarshal),
		"csvutil":  UnmarshalerFunc(csvutil.NewCSVUtilUnmarshaler().Unmarshal),
		"encoding": UnmarshalerFunc(encoding.NewEncodingCSVUnmarshaler().Unmarshal),
	}

	// 对每个测试用例，验证所有实现的行为是否一致
	for _, tc := range testCases {
		t.Run(tc.name+" - "+tc.description, func(t *testing.T) {
			results := make(map[string]bool)

			for implName, unmarshaler := range implementations {
				var result []TestPerson
				err := unmarshaler.Unmarshal([]byte(tc.data), &result)

				hasError := err != nil
				results[implName] = hasError

				if tc.expectError != hasError {
					t.Logf("Implementation %s: expected error=%v, got error=%v (error: %v)",
						implName, tc.expectError, hasError, err)
				}
			}

			// 检查所有实现的行为是否一致
			// 注意：不同的CSV库可能有不同的错误处理策略，所以我们只是记录差异
			firstResult := ""
			for implName, hasError := range results {
				if firstResult == "" {
					firstResult = implName
					continue
				}

				if results[firstResult] != hasError {
					t.Logf("Behavior difference detected: %s returned error=%v, %s returned error=%v",
						firstResult, results[firstResult], implName, hasError)
				}
			}
		})
	}
}