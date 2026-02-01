package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSON编码示例
type User struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email,omitempty"`
}

func jsonExample() {
	user := User{
		Name: "张三",
		Age:  25,
		Email: "zhangsan@example.com",
	}

	// 编码为JSON
	jsonData, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("JSON编码错误: %v\n", err)
		return
	}
	fmt.Printf("JSON编码: %s\n", string(jsonData))

	// 从JSON解码
	var decodedUser User
	err = json.Unmarshal(jsonData, &decodedUser)
	if err != nil {
		fmt.Printf("JSON解码错误: %v\n", err)
		return
	}
	fmt.Printf("JSON解码: %+v\n", decodedUser)
}

// Base64编码示例
func base64Example() {
	data := "Hello, 世界!"

	// 标准Base64编码
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Printf("Base64编码: %s\n", encoded)

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Printf("Base64解码错误: %v\n", err)
		return
	}
	fmt.Printf("Base64解码: %s\n", string(decoded))
}

// 十六进制编码示例
func hexExample() {
	data := []byte{72, 101, 108, 108, 111, 44, 32, 228, 184, 150, 231, 149, 140, 33}

	// 编码为十六进制
	encoded := hex.EncodeToString(data)
	fmt.Printf("十六进制编码: %s\n", encoded)

	// 从十六进制解码
	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		fmt.Printf("十六进制解码错误: %v\n", err)
		return
	}
	fmt.Printf("十六进制解码: %s\n", string(decoded))
}

// CSV编码示例
func csvExample() {
	records := [][]string{
		{"姓名", "年龄", "城市"},
		{"张三", "25", "北京"},
		{"李四", "30", "上海"},
		{"王五", "28", "广州"},
	}

	// 创建CSV文件
	file, err := os.Create("example.csv")
	if err != nil {
		fmt.Printf("创建CSV文件错误: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV数据
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			fmt.Printf("写入CSV错误: %v\n", err)
			return
		}
	}

	fmt.Println("CSV文件已创建: example.csv")
}

func main() {
	fmt.Println("=== Go语言编码示例 ===\n")

	fmt.Println("1. JSON编码示例:")
	jsonExample()
	fmt.Println()

	fmt.Println("2. Base64编码示例:")
	base64Example()
	fmt.Println()

	fmt.Println("3. 十六进制编码示例:")
	hexExample()
	fmt.Println()

	fmt.Println("4. CSV编码示例:")
	csvExample()
	fmt.Println()
}