package encoding

import (
	csvunmarshaler "csv-unmarshaler"
	"testing"
)

// TestPerson 测试用的结构体
type TestPerson struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

// TestEncodingCSVUnmarshaler 使用通用测试函数测试encoding/csv实现的Unmarshaler
func TestEncodingCSVUnmarshaler(t *testing.T) {
	unmarshaler := NewEncodingCSVUnmarshaler()

	// 将我们的unmarshaler适配为通用接口
	adapter := csvunmarshaler.UnmarshalerFunc(unmarshaler.Unmarshal)

	csvunmarshaler.CommonTestUnmarshalerInterface(t, adapter)
}

// TestEncodingCSVUnmarshalerSpecific encoding/csv特定的测试
func TestEncodingCSVUnmarshalerSpecific(t *testing.T) {
	unmarshaler := NewEncodingCSVUnmarshaler()

	t.Run("encoding/csv specific features", func(t *testing.T) {
		// 测试encoding/csv特有的功能或行为
		data := []byte("name,age\nJohn,30\nAlice,25")
		var result []TestPerson
		err := unmarshaler.Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}
	})

	t.Run("test with different field types", func(t *testing.T) {
		// 测试不同类型的字段
		type ComplexPerson struct {
			Name    string  `csv:"name"`
			Age     int     `csv:"age"`
			Height  float64 `csv:"height"`
			Active  bool    `csv:"active"`
		}

		csvData := `name,age,height,active
John,30,1.75,true
Alice,25,1.68,false`

		var result []ComplexPerson
		err := unmarshaler.Unmarshal([]byte(csvData), &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}

		// 验证第一个记录
		if result[0].Name != "John" || result[0].Age != 30 ||
		   result[0].Height != 1.75 || result[0].Active != true {
			t.Errorf("First record mismatch: %+v", result[0])
		}
	})

	t.Run("test with manual parsing", func(t *testing.T) {
		// 测试手动解析功能
		csvData := "name,age\nJohn,30\nAlice,25"
		records, err := unmarshaler.ParseCSV(csvData)
		if err != nil {
			t.Fatalf("ParseCSV failed: %v", err)
		}

		if len(records) != 3 { // header + 2 data rows
			t.Errorf("Expected 3 records, got %d", len(records))
		}

		if records[0][0] != "name" || records[0][1] != "age" {
			t.Errorf("Header mismatch: %v", records[0])
		}
	})
}