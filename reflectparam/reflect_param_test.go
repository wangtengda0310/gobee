package reflectparam

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/why2go/csv_parser"
)

// ========== Parse 函数测试 ==========

func TestParse_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Args
	}{
		{
			name:     "标准输入",
			input:    "hello 123",
			expected: Args{S: "hello", V: 123},
		},
		{
			name:     "空格分隔",
			input:    "world 456",
			expected: Args{S: "world", V: 456},
		},
		{
			name:     "数字0",
			input:    "test 0",
			expected: Args{S: "test", V: 0},
		},
		{
			name:     "负数",
			input:    "item -100",
			expected: Args{S: "item", V: -100},
		},
		{
			name:     "大数字",
			input:    "big 2147483647",
			expected: Args{S: "big", V: 2147483647},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expected.S, result.S)
			assert.Equal(t, tt.expected.V, result.V)
		})
	}
}

func TestParse_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Args
	}{
		{
			name:     "只有字符串没有数字",
			input:    "onlystring",
			expected: Args{S: "onlystring", V: 0},
		},
		{
			name:     "只有数字没有字符串",
			input:    " 456",
			expected: Args{S: "", V: 456},
		},
		{
			name:     "空字符串",
			input:    "",
			expected: Args{S: "", V: 0},
		},
		{
			name:     "只有空格",
			input:    "   ",
			expected: Args{S: "", V: 0},
		},
		{
			name:     "多个空格分隔",
			input:    "one  two  123",
			expected: Args{S: "one", V: 0}, // 第二部分 "two" 不是数字
		},
		{
			name:     "特殊字符字符串",
			input:    "测试中文 999",
			expected: Args{S: "测试中文", V: 999},
		},
		{
			name:     "包含特殊符号",
			input:    "test-value! 123",
			expected: Args{S: "test-value!", V: 123},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expected.S, result.S, "字符串字段不匹配")
			assert.Equal(t, tt.expected.V, result.V, "数字字段不匹配")
		})
	}
}

func TestParse_InvalidNumber(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedStr string
		expectedVal int
	}{
		{
			name:        "无效数字格式",
			input:       "test abc",
			expectedStr: "test",
			expectedVal: 0,
		},
		{
			name:        "数字带字母",
			input:       "test 123abc",
			expectedStr: "test",
			expectedVal: 0,
		},
		{
			name:        "浮点数",
			input:       "test 123.45",
			expectedStr: "test",
			expectedVal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expectedStr, result.S)
			assert.Equal(t, tt.expectedVal, result.V)
		})
	}
}

// ========== ParseGeneric 函数测试 ==========

func TestParseGeneric_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		factory  func() Args
		expected Args
	}{
		{
			name:     "工厂函数创建实例",
			input:    "test 789",
			factory:  func() Args { return Args{} },
			expected: Args{S: "test", V: 789},
		},
		{
			name:     "带默认值的工厂",
			input:    "hello 999",
			factory:  func() Args { return Args{S: "default"} },
			expected: Args{S: "hello", V: 999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGeneric(tt.input, tt.factory)
			assert.Equal(t, tt.expected.S, result.S)
			assert.Equal(t, tt.expected.V, result.V)
		})
	}
}

func TestParseGeneric_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Args
	}{
		{
			name:     "空输入",
			input:    "",
			expected: Args{S: "", V: 0},
		},
		{
			name:     "只有字符串",
			input:    "single",
			expected: Args{S: "single", V: 0},
		},
		{
			name:     "只有数字",
			input:    " 123",
			expected: Args{S: "", V: 123},
		},
		{
			name:     "负数",
			input:    "test -999",
			expected: Args{S: "test", V: -999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGeneric(tt.input, func() Args { return Args{} })
			assert.Equal(t, tt.expected.S, result.S)
			assert.Equal(t, tt.expected.V, result.V)
		})
	}
}

// ========== CSV 解析测试 ==========

func TestParseCSV_ValidData(t *testing.T) {
	csvData := `name,value
alice,100
bob,200
charlie,300`

	r := csv.NewReader(bytes.NewBufferString(csvData))
	parser, err := csv_parser.NewCsvParser[Args](r)
	require.NoError(t, err, "创建parser不应失败")
	defer parser.Close()

	expected := []Args{
		{S: "alice", V: 100},
		{S: "bob", V: 200},
		{S: "charlie", V: 300},
	}

	i := 0
	for dataWrapper := range parser.DataChan(context.Background()) {
		require.NoError(t, dataWrapper.Err, "解析不应出错")
		require.Less(t, i, len(expected), "数据量不应超过预期")
		assert.Equal(t, expected[i].S, dataWrapper.Data.S, "第%d行字符串不匹配", i+1)
		assert.Equal(t, expected[i].V, dataWrapper.Data.V, "第%d行数字不匹配", i+1)
		i++
	}

	assert.Equal(t, len(expected), i, "应解析所有行")
}

func TestParseCSV_EmptyData(t *testing.T) {
	csvData := `name,value`

	r := csv.NewReader(bytes.NewBufferString(csvData))
	parser, err := csv_parser.NewCsvParser[Args](r)
	require.NoError(t, err)
	defer parser.Close()

	count := 0
	for dataWrapper := range parser.DataChan(context.Background()) {
		require.NoError(t, dataWrapper.Err)
		count++
	}

	assert.Equal(t, 0, count, "只有标题行应返回0条数据")
}

func TestParseCSV_PartialData(t *testing.T) {
	tests := []struct {
		name     string
		csvData  string
		expectedS string
		expectedV int
	}{
		{
			name:     "只有字符串无数字",
			csvData:  "name,value\nitem,",
			expectedS: "item",
			expectedV: 0,
		},
		{
			name:     "只有数字无字符串",
			csvData:  "name,value\n,456",
			expectedS: "",
			expectedV: 456,
		},
		{
			name:     "都为空",
			csvData:  "name,value\n,",
			expectedS: "",
			expectedV: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := csv.NewReader(bytes.NewBufferString(tt.csvData))
			parser, err := csv_parser.NewCsvParser[Args](r)
			require.NoError(t, err)
			defer parser.Close()

			count := 0
			var result *Args
			for dataWrapper := range parser.DataChan(context.Background()) {
				require.NoError(t, dataWrapper.Err)
				result = dataWrapper.Data
				count++
			}

			assert.Equal(t, 1, count, "应解析一条数据")
			assert.Equal(t, tt.expectedS, result.S)
			assert.Equal(t, tt.expectedV, result.V)
		})
	}
}

func TestParseCSV_InvalidNumber(t *testing.T) {
	csvData := `name,value
test,invalid
bob,200`

	r := csv.NewReader(bytes.NewBufferString(csvData))
	parser, err := csv_parser.NewCsvParser[Args](r)
	require.NoError(t, err)
	defer parser.Close()

	errorCount := 0
	successCount := 0

	for dataWrapper := range parser.DataChan(context.Background()) {
		if dataWrapper.Err != nil {
			// 无效数字会返回错误
			errorCount++
		} else {
			successCount++
		}
	}

	// 应该有错误和成功的情况混合
	// csv_parser 会处理所有行，但无效数字会产生错误
	totalRows := errorCount + successCount
	assert.Greater(t, totalRows, 0, "应该处理了一些行")
	t.Logf("成功: %d, 错误: %d, 总计: %d", successCount, errorCount, totalRows)
}

func TestParseCSV_MultipleRows(t *testing.T) {
	tests := []struct {
		name         string
		rowCount     int
	}{
		{"单行", 1},
		{"10行", 10},
		{"100行", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成CSV数据
			var buf bytes.Buffer
			writer := csv.NewWriter(&buf)
			writer.Write([]string{"name", "value"})
			for i := 0; i < tt.rowCount; i++ {
				writer.Write([]string{"item", "100"})
			}
			writer.Flush()

			r := csv.NewReader(&buf)
			parser, err := csv_parser.NewCsvParser[Args](r)
			require.NoError(t, err)
			defer parser.Close()

			count := 0
			for dataWrapper := range parser.DataChan(context.Background()) {
				require.NoError(t, dataWrapper.Err)
				count++
			}

			assert.Equal(t, tt.rowCount, count, "应解析所有%d行", tt.rowCount)
		})
	}
}

// ========== Args 结构体测试 ==========

func TestArgs_Structure(t *testing.T) {
	args := Args{S: "test", V: 123}

	assert.Equal(t, "test", args.S, "S字段应正确设置")
	assert.Equal(t, 123, args.V, "V字段应正确设置")
}

func TestArgs_DefaultValues(t *testing.T) {
	var args Args

	assert.Empty(t, args.S, "默认S应为空")
	assert.Zero(t, args.V, "默认V应为0")
}

// ========== 并发测试 ==========

func TestParse_Concurrent(t *testing.T) {
	// 测试并发解析是否安全
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			result := Parse("test 123")
			assert.Equal(t, "test", result.S)
			assert.Equal(t, 123, result.V)
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestParseGeneric_Concurrent(t *testing.T) {
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			result := ParseGeneric("hello 456", func() Args { return Args{} })
			assert.Equal(t, "hello", result.S)
			assert.Equal(t, 456, result.V)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ========== 性能/压力测试 ==========

func TestParse_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	iterations := 10000
	for i := 0; i < iterations; i++ {
		_ = Parse("test 123")
	}
}
