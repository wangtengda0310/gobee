package reflectparam

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"

	"github.com/jszwec/csvutil"
	"github.com/why2go/csv_parser"
)

// 生成测试CSV数据
func generateCSVData(rows int) string {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"name", "value"})
	for i := 0; i < rows; i++ {
		writer.Write([]string{"item", "123"})
	}
	writer.Flush()
	return buf.String()
}

// 基准测试: why2go/csv_parser
func BenchmarkCsvParser(b *testing.B) {
	data := generateCSVData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewBufferString(data))
		parser, err := csv_parser.NewCsvParser[Args](r)
		if err != nil {
			b.Fatal(err)
		}

		ctx := context.Background()
		count := 0
		for range parser.DataChan(ctx) {
			count++
		}
		parser.Close()
	}
}

// 基准测试: jszwec/csvutil
func BenchmarkCsvutil(b *testing.B) {
	data := generateCSVData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewBufferString(data))
		dec, err := csvutil.NewDecoder(r)
		if err != nil {
			b.Fatal(err)
		}

		var args Args
		for {
			err := dec.Decode(&args)
			if err != nil {
				break
			}
		}
	}
}

// 基准测试: 标准库 encoding/csv
func BenchmarkStdCSV(b *testing.B) {
	data := generateCSVData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewBufferString(data))
		records, err := r.ReadAll()
		if err != nil {
			b.Fatal(err)
		}
		// 跳过标题行
		for i := 1; i < len(records); i++ {
			var args Args
			if len(records[i]) >= 2 {
				args.S = records[i][0]
			}
		}
	}
}

// 基准测试: 反射解析
func BenchmarkParseReflection(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Parse("test 123")
	}
}

// 基准测试: 泛型反射解析
func BenchmarkParseGeneric(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ParseGeneric("test 123", func() Args { return Args{} })
	}
}

// 基准测试: 直接赋值
func BenchmarkParseDirect(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Args{S: "test", V: 123}
	}
}
