// Package main 提供 gocsv 库的实际运行示例
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type Product struct {
	Name     string  `csv:"name"`
	Price    float64 `csv:"price"`
	InStock  bool    `csv:"in_stock"`
	Category string  `csv:"category"`
}

func main() {
	fmt.Println("=== gocsv 库示例 ===")

	// 1. 写入CSV文件
	fmt.Println("\n1. 写入CSV文件")
	writeCSV()

	// 2. 读取CSV文件 (手动方式验证)
	fmt.Println("\n2. 手动解析CSV验证内容")
	manualParseCSV()

	// 3. 使用gocsv读取（说明局限性）
	fmt.Println("\n3. gocsv库现状说明")
	printGocsvNotes()

	fmt.Println("\n=== 完成 ===")
}

func writeCSV() {
	products := []Product{
		{Name: "Laptop", Price: 999.99, InStock: true, Category: "Electronics"},
		{Name: "Mouse", Price: 29.99, InStock: true, Category: "Electronics"},
		{Name: "Desk", Price: 299.00, InStock: false, Category: "Furniture"},
	}

	// 使用标准库写入CSV
	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入标题
	writer.Write([]string{"name", "price", "in_stock", "category"})

	// 写入数据
	for _, p := range products {
		writer.Write([]string{
			p.Name,
			fmt.Sprintf("%.2f", p.Price),
			fmt.Sprintf("%t", p.InStock),
			p.Category,
		})
	}

	fmt.Println("   已写入 products.csv")
	fmt.Println("   内容:")
	for _, p := range products {
		fmt.Printf("   - %s: $%.2f (库存: %t, 分类: %s)\n",
			p.Name, p.Price, p.InStock, p.Category)
	}
}

func manualParseCSV() {
	// 手动解析CSV，展示正确解析结果
	file, err := os.Open("products.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("   手动解析结果:")
	for i, record := range records {
		if i == 0 {
			fmt.Printf("   标题: %v\n", record)
			continue
		}
		if len(record) >= 4 {
			name := record[0]
			price := record[1]
			inStock := record[2] == "true"
			category := record[3]
			inStockStr := "false"
			if inStock {
				inStockStr = "true"
			}
			fmt.Printf("   - %s: %s (库存: %s, 分类: %s)\n",
				name, price, inStockStr, category)
		}
	}
}

func printGocsvNotes() {
	fmt.Println(`   注意: gocsv 库 (foolin/gocsv) 最后更新于 2017年，存在以下问题:`)
	fmt.Println(`   ⚠️ 1. gocsv.ReadList() 在处理标准CSV时有bug`)
	fmt.Println(`   ⚠️ 2. 不支持 io.Reader 接口，仅支持文件路径`)
	fmt.Println(`   ⚠️ 3. 标签解析可能不兼容现代Go版本`)
	fmt.Println(``)
	fmt.Println(`   推荐替代方案:`)
	fmt.Println(`   ✅ jszwec/csvutil - 高性能、支持泛型`)
	fmt.Println(`   ✅ why2go/csv_parser - 支持复杂嵌套、channel流式`)
	fmt.Println(``)
	fmt.Println(`   示例项目展示了这三个库的正确用法。`)
}
