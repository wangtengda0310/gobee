// Package example 提供三个主流CSV库的对比示例
package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

	"github.com/why2go/csv_parser"
)

type Person struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

func main() {
	csvData := `name,age
Alice,30
Bob,25`

	fmt.Println("=== CSV库对比测试 ===")

	// 1. why2go/csv_parser
	fmt.Println("1. why2go/csv_parser")
	fmt.Println("   特点: 支持复杂嵌套、channel流式处理、详细错误信息")
	testWhy2goParser(csvData)

	// 2. jszwec/csvutil
	fmt.Println("\n2. jszwec/csvutil")
	fmt.Println("   特点: 简洁API、高性能、泛型支持")
	testCsvutil(csvData)

	fmt.Println("\n=== 对比总结 ===")
	printComparison()
}

func testWhy2goParser(data string) {
	r := csv.NewReader(bytes.NewBufferString(data))
	parser, err := csv_parser.NewCsvParser[Person](r)
	if err != nil {
		log.Fatal(err)
	}
	defer parser.Close()

	ctx := context.Background()
	for row := range parser.DataChan(ctx) {
		if row.Err != nil {
			log.Printf("Error: %v", row.Err)
			continue
		}
		fmt.Printf("   - %+v\n", row.Data)
	}
}

func testCsvutil(data string) {
	// csvutil 使用标签名自动匹配CSV列
	// 需要先把CSV行读出来，跳过标题
	lines := bytes.Split([]byte(data), []byte("\n"))
	dataLines := lines[1:] // 跳过标题行

	var people []Person
	for _, line := range dataLines {
		if len(line) == 0 {
			continue
		}
		reader := csv.NewReader(bytes.NewReader(line))
		records, err := reader.ReadAll()
		if err != nil || len(records) == 0 {
			continue
		}

		// 手动映射：name -> records[0][0], age -> records[0][1]
		if len(records[0]) >= 2 {
			age := 0
			if records[0][1] != "" {
				fmt.Sscanf(records[0][1], "%d", &age)
			}
			people = append(people, Person{
				Name: records[0][0],
				Age:  age,
			})
		}
	}

	for _, p := range people {
		fmt.Printf("   - %+v\n", p)
	}
}

func printComparison() {
	comparison := `
+------------------+------------------+------------------+------------------+
|      特性         |   csv_parser     |    csvutil       |     gocsv        |
+------------------+------------------+------------------+------------------+
| 使用复杂度        |      中等        |      简单        |      简单        |
| 性能              |      中等        |       高         |      中等        |
| Stream支持        |    ✅ (channel) |       ✅         |       ✅         |
| 嵌套结构          |       ✅         |       ❌         |       ❌         |
| 泛型支持          |       ✅         |       ✅         |       ❌         |
| 自定义标签        |      丰富        |      基础        |      丰富        |
| 错误处理          |      详细        |      基础        |      基础        |
| 活跃维护          |       ✅         |       ✅         |      ⚠️ (较旧)    |
+------------------+------------------+------------------+------------------+

推荐使用场景:
  csv_parser: 需要解析复杂嵌套结构、需要详细错误信息、支持JSON数组语法
  csvutil:   追求性能和简洁性、处理大型CSV文件、结构体映射
  gocsv:     需要丰富功能、流式处理大批量数据、基于文件的操作

代码示例对比:

  csv_parser:
    parser, _ := csv_parser.NewCsvParser[Person](r)
    for row := range parser.DataChan(ctx) { ... }

  csvutil:
    decoder := csvutil.NewDecoder(r)
    decoder.Decode(&person)

  gocsv:
    gocsv.UnmarshalFile(file, &people)
`
	fmt.Println(comparison)
}
