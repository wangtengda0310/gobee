package csvutil

import (
	csvunmarshaler "csv-unmarshaler"
	"strings"
	"testing"
)

// TestPerson 测试用的结构体
type TestPerson struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

// TestCSVUtilUnmarshaler 使用通用测试函数测试csvutil实现的Unmarshaler
func TestCSVUtilUnmarshaler(t *testing.T) {
	unmarshaler := NewCSVUtilUnmarshaler()

	// 将我们的unmarshaler适配为通用接口
	adapter := csvunmarshaler.UnmarshalerFunc(unmarshaler.Unmarshal)

	csvunmarshaler.CommonTestUnmarshalerInterface(t, adapter)
}

// TestCSVUtilUnmarshalerSpecific csvutil特定的测试
func TestCSVUtilUnmarshalerSpecific(t *testing.T) {
	unmarshaler := NewCSVUtilUnmarshaler()

	t.Run("csvutil specific features", func(t *testing.T) {
		// 测试csvutil特有的功能或行为
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

	t.Run("test with header mapping", func(t *testing.T) {
		// 测试csvutil的header映射功能
		csvData := "Name,Age\nJohn,30\nAlice,25"
		var result []TestPerson
		err := unmarshaler.Unmarshal([]byte(csvData), &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}
	})

	t.Run("test with custom decoder", func(t *testing.T) {
		// 测试使用自定义decoder
		reader := strings.NewReader("name,age\nJohn,30\nAlice,25")
		decoder, err := NewCSVUtilDecoder(reader)
		if err != nil {
			t.Fatalf("Failed to create decoder: %v", err)
		}

		var result []TestPerson
		for {
			person := TestPerson{}
			if err := decoder.Decode(&person); err != nil {
				break
			}
			result = append(result, person)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 records, got %d", len(result))
		}
	})
}