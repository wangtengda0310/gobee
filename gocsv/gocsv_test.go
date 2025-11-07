package gocsv

import (
	csvunmarshaler "csv-unmarshaler"
	"testing"
)

// TestPerson 测试用的结构体
type TestPerson struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

// TestGoCSVUnmarshaler 使用通用测试函数测试gocsv实现的Unmarshaler
func TestGoCSVUnmarshaler(t *testing.T) {
	unmarshaler := NewGoCSVUnmarshaler()

	// 将我们的unmarshaler适配为通用接口
	adapter := csvunmarshaler.UnmarshalerFunc(unmarshaler.Unmarshal)

	csvunmarshaler.CommonTestUnmarshalerInterface(t, adapter)
}

// TestGoCSVUnmarshalerSpecific gocsv特定的测试
func TestGoCSVUnmarshalerSpecific(t *testing.T) {
	unmarshaler := NewGoCSVUnmarshaler()

	t.Run("gocsv specific features", func(t *testing.T) {
		// 测试gocsv特有的功能或行为
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
}