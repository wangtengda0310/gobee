package csvunmarshaler

import (
	"testing"
)

// CommonTestUnmarshalerInterface 通用的接口测试函数，用于验证李氏替换原则
func CommonTestUnmarshalerInterface(t *testing.T, unmarshaler Unmarshaler) {
	// TestPerson 测试用的结构体
	type TestPerson struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	// TestData 测试数据
	const TestData = `name,age
John,30
Alice,25
Bob,35`

	t.Run("Interface Substitution Principle", func(t *testing.T) {
		// 测试数据
		testCases := []struct {
			name     string
			data     []byte
			expected []TestPerson
		}{
			{
				name: "normal case",
				data: []byte(TestData),
				expected: []TestPerson{
					{Name: "John", Age: 30},
					{Name: "Alice", Age: 25},
					{Name: "Bob", Age: 35},
				},
			},
			{
				name:     "empty data",
				data:     []byte("name,age\n"),
				expected: []TestPerson{},
			},
			{
				name: "single record",
				data: []byte("name,age\nJohn,30"),
				expected: []TestPerson{
					{Name: "John", Age: 30},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var result []TestPerson
				err := unmarshaler.Unmarshal(tc.data, &result)
				if err != nil {
					t.Errorf("Unmarshal failed: %v", err)
					return
				}

				if len(result) != len(tc.expected) {
					t.Errorf("Expected %d records, got %d", len(tc.expected), len(result))
					return
				}

				for j, expected := range tc.expected {
					if result[j].Name != expected.Name || result[j].Age != expected.Age {
						t.Errorf("Record %d: expected %+v, got %+v", j, expected, result[j])
					}
				}
			})
		}
	})

	t.Run("Error Handling Consistency", func(t *testing.T) {
		// 测试各种错误情况
		errorTestCases := []struct {
			name        string
			data        []byte
			expectError bool
		}{
			{
				name:        "invalid csv format",
				data:        []byte("name,age\nJohn,invalid_age"),
				expectError: true,
			},
			{
				name:        "missing columns",
				data:        []byte("name\nJohn"),
				expectError: false, // 某些库可能处理这种情况
			},
			{
				name:        "malformed csv",
				data:        []byte("name,age\nJohn,30,\nAlice,25"),
				expectError: true,
			},
		}

		for _, tc := range errorTestCases {
			t.Run(tc.name, func(t *testing.T) {
				var result []TestPerson
				err := unmarshaler.Unmarshal(tc.data, &result)
				if tc.expectError && err == nil {
					t.Errorf("Expected error but got none")
				}
				if !tc.expectError && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})
		}
	})
}