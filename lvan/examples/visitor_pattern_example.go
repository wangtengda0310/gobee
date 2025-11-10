package main

import (
	"fmt"
	"log"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump/datasource/v2"
)

// ExampleVisitorPattern demonstrates how to use the new visitor pattern
func ExampleVisitorPattern() {
	fmt.Println("=== LVAN Dumper Visitor Pattern Example ===")

	// 1. 创建MySQL配置
	config := v2.NewMySQLConfig(
		"localhost", // 主机
		3306,        // 端口
		"root",      // 用户名
		"password",  // 密码
		"gforge",    // 数据库
		"user",      // 表
	)

	// 2. 创建数据源
	datasource := v2.NewMySQLDatasource(config)

	// 3. 使用不同的访问者进行操作
	visitors := []struct {
		name    string
		visitor v2.Visitor
	}{
		{
			name:    "基础信息访问者",
			visitor: &BasicInfoVisitor{},
		},
		{
			name:    "数据导出访问者",
			visitor: v2.NewMySQLExportVisitor(&v2.MySQLOptions{
				Where:        "uid > 0",
				IDs:          []string{"1", "2", "3"},
				Fields:       []string{"uid", "accountid", "data"},
				Limit:        100,
				OutputFormat: "json",
				OutputPath:   "/tmp/export.json",
			}),
		},
		{
			name:    "数据验证访问者",
			visitor: &DataValidationVisitor{},
		},
	}

	// 4. 应用每个访问者
	for _, v := range visitors {
		fmt.Printf("\n--- 使用 %s ---\n", v.name)
		datasource.Accept(v.visitor)
	}
}

// BasicInfoVisitor 基础信息访问者
type BasicInfoVisitor struct{}

func (v *BasicInfoVisitor) VisitDatasource(ds v2.Datasource) {
	metadata := ds.GetMetadata()
	fmt.Printf("数据源类型: %s\n", metadata.GetType())
	fmt.Printf("包含表: %v\n", metadata.GetTables())
}

func (v *BasicInfoVisitor) VisitMySQL(ds v2.MySQLDatasource) {
	fmt.Printf("MySQL连接信息:\n")
	fmt.Printf("  主机: %s:%d\n", ds.GetHost(), ds.GetPort())
	fmt.Printf("  数据库: %s\n", ds.GetDatabase())
	fmt.Printf("  表: %s\n", ds.GetTable())
	fmt.Printf("  用户: %s\n", ds.GetUser())

	// 调用基础方法
	v.VisitDatasource(ds)
}

func (v *BasicInfoVisitor) VisitRedis(ds v2.RedisDatasource) {
	fmt.Printf("Redis连接信息:\n")
	fmt.Printf("  主机: %s:%d\n", ds.GetHost(), ds.GetPort())
	fmt.Printf("  数据库: %d\n", ds.GetDatabase())
	fmt.Printf("  键模式: %s\n", ds.GetKeyPattern())

	// 调用基础方法
	v.VisitDatasource(ds)
}

// DataValidationVisitor 数据验证访问者
type DataValidationVisitor struct {
	validationResults []string
}

func (v *DataValidationVisitor) VisitDatasource(ds v2.Datasource) {
	fmt.Printf("开始验证数据源: %s\n", ds.GetMetadata().GetType())
}

func (v *DataValidationVisitor) VisitMySQL(ds v2.MySQLDatasource) {
	v.VisitDatasource(ds)

	// 验证MySQL特定配置
	validations := []string{}

	// 验证主机名
	if ds.GetHost() == "" {
		validations = append(validations, "❌ 主机名不能为空")
	} else {
		validations = append(validations, "✅ 主机名有效")
	}

	// 验证端口
	port := ds.GetPort()
	if port <= 0 || port > 65535 {
		validations = append(validations, "❌ 端口号无效")
	} else {
		validations = append(validations, "✅ 端口号有效")
	}

	// 验证数据库名
	if ds.GetDatabase() == "" {
		validations = append(validations, "❌ 数据库名不能为空")
	} else {
		validations = append(validations, "✅ 数据库名有效")
	}

	// 输出验证结果
	fmt.Println("配置验证结果:")
	for _, result := range validations {
		fmt.Printf("  %s\n", result)
	}

	v.validationResults = validations
}

func (v *DataValidationVisitor) VisitRedis(ds v2.RedisDatasource) {
	v.VisitDatasource(ds)

	// 验证Redis特定配置
	fmt.Printf("验证Redis配置: %s:%d, db=%d\n",
		ds.GetHost(), ds.GetPort(), ds.GetDatabase())

	v.validationResults = append(v.validationResults, "✅ Redis配置已验证")
}

// ExampleCustomVisitor shows how to create a custom visitor
func ExampleCustomVisitor() {
	fmt.Println("\n=== 自定义访问者示例 ===")

	config := v2.NewMySQLConfig("localhost", 3306, "user", "pass", "db", "table")
	datasource := v2.NewMySQLDatasource(config)

	// 创建自定义访问者
	customVisitor := &CustomReportVisitor{
		reportTitle: "数据源报告",
		sections:    []string{},
	}

	// 应用访问者
	datasource.Accept(customVisitor)

	// 打印报告
	fmt.Println(customVisitor.GenerateReport())
}

// CustomReportVisitor 自定义报告访问者
type CustomReportVisitor struct {
	reportTitle string
	sections    []string
}

func (v *CustomReportVisitor) VisitDatasource(ds v2.Datasource) {
	metadata := ds.GetMetadata()
	v.sections = append(v.sections, fmt.Sprintf("数据源类型: %s", metadata.GetType()))
}

func (v *CustomReportVisitor) VisitMySQL(ds v2.MySQLDatasource) {
	v.VisitDatasource(ds)

	v.sections = append(v.sections, fmt.Sprintf("连接信息: %s@%s:%d/%s.%s",
		ds.GetUser(), ds.GetHost(), ds.GetPort(), ds.GetDatabase(), ds.GetTable()))
	v.sections = append(v.sections, "状态: 已连接")
}

func (v *CustomReportVisitor) VisitRedis(ds v2.RedisDatasource) {
	v.VisitDatasource(ds)

	v.sections = append(v.sections, fmt.Sprintf("连接信息: %s:%d/%d",
		ds.GetHost(), ds.GetPort(), ds.GetDatabase()))
	v.sections = append(v.sections, "状态: 已连接")
}

func (v *CustomReportVisitor) GenerateReport() string {
	report := fmt.Sprintf("=== %s ===\n", v.reportTitle)
	for _, section := range v.sections {
		report += fmt.Sprintf("%s\n", section)
	}
	return report
}

// ExampleVisitorChaining demonstrates chaining multiple visitors
func ExampleVisitorChaining() {
	fmt.Println("\n=== 访问者链示例 ===")

	config := v2.NewMySQLConfig("localhost", 3306, "user", "pass", "db", "table")
	datasource := v2.NewMySQLDatasource(config)

	// 创建访问者链
	visitors := []v2.Visitor{
		&BasicInfoVisitor{},           // 首先获取基础信息
		&DataValidationVisitor{},      // 然后验证配置
		&CustomReportVisitor{          // 最后生成报告
			reportTitle: "综合数据源报告",
			sections:    []string{},
		},
	}

	// 依次应用每个访问者
	for i, visitor := range visitors {
		fmt.Printf("\n步骤 %d: ", i+1)
		datasource.Accept(visitor)
	}
}

// ExampleErrorHandling demonstrates error handling in visitor pattern
func ExampleErrorHandling() {
	fmt.Println("\n=== 错误处理示例 ===")

	// 测试无效配置
	testCases := []struct {
		name string
		config v2.MySQLConfig
	}{
		{
			name: "无效端口配置",
			config: v2.NewMySQLConfig("localhost", 99999, "user", "pass", "db", "table"),
		},
		{
			name: "空主机配置",
			config: v2.NewMySQLConfig("", 3306, "user", "pass", "db", "table"),
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n测试 %s:\n", tc.name)

		// 捕获panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("✅ 成功捕获配置错误: %v\n", r)
				}
			}()

			datasource := v2.NewMySQLDatasource(tc.config)
			fmt.Printf("❌ 应该因为配置错误而失败，但却创建了数据源: %v\n", datasource)
		}()
	}
}

func main() {
	// 运行所有示例
	ExampleVisitorPattern()
	ExampleCustomVisitor()
	ExampleVisitorChaining()
	ExampleErrorHandling()

	fmt.Println("\n=== 示例完成 ===")
	log.Println("所有访问者模式示例执行完毕")
}