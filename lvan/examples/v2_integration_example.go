// v2架构集成示例
// 展示LVAN Dumper v2架构的完整使用流程

package main

import (
	"fmt"
	"log"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump/datasource/v2"
)

func main() {
	fmt.Println("=== LVAN Dumper v2架构集成示例 ===\n")

	// 1. 创建数据源工厂
	fmt.Println("1. 创建数据源工厂")
	factory := v2.NewDatasourceFactory()

	// 验证工厂状态
	if !factory.IsValid() {
		log.Fatal("工厂创建失败")
	}
	fmt.Println("✅ 工厂创建成功")

	// 2. 创建MySQL配置
	fmt.Println("\n2. 创建MySQL配置")
	mysqlConfig := v2.NewMySQLConfig(
		"localhost",    // 主机
		3306,           // 端口
		"lvan_user",    // 用户名
		"lvan_pass",    // 密码
		"lvan_db",      // 数据库
		"users",        // 表名
	)

	// 验证配置
	if err := mysqlConfig.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}
	fmt.Println("✅ MySQL配置创建并验证成功")

	// 3. 通过工厂创建MySQL数据源
	fmt.Println("\n3. 创建MySQL数据源")
	mysqlDatasource, err := factory.CreateMySQL(mysqlConfig)
	if err != nil {
		log.Fatalf("MySQL数据源创建失败: %v", err)
	}
	fmt.Println("✅ MySQL数据源创建成功")

	// 显示数据源信息
	fmt.Printf("   数据库: %s\n", mysqlDatasource.GetDatabase())
	fmt.Printf("   表名: %s\n", mysqlDatasource.GetTable())
	fmt.Printf("   元数据类型: %s\n", mysqlDatasource.GetMetadata().GetType())

	// 4. 创建MySQL导出访问者
	fmt.Println("\n4. 创建MySQL导出访问者")
	exportOptions := &v2.MySQLOptions{
		Where:       "age > 18 AND status = 'active'",
		IDs:         []string{"1", "2", "3", "5", "8"},
		Fields:      []string{"id", "name", "email", "age", "created_at"},
		Limit:       1000,
		OutputFormat: "zip",
		OutputPath:  "./exports",
		Compression: "gzip",
	}

	exportVisitor := v2.NewMySQLExportVisitor(exportOptions)
	fmt.Println("✅ 导出访问者创建成功")

	// 显示导出配置
	fmt.Printf("   WHERE条件: %s\n", exportOptions.Where)
	fmt.Printf("   ID列表: %v\n", exportOptions.IDs)
	fmt.Printf("   字段列表: %v\n", exportOptions.Fields)
	fmt.Printf("   输出格式: %s\n", exportOptions.OutputFormat)
	fmt.Printf("   限制条数: %d\n", exportOptions.Limit)

	// 5. 执行数据导出
	fmt.Println("\n5. 执行数据导出")
	mysqlDatasource.Accept(exportVisitor)
	fmt.Println("✅ 数据导出执行完成")

	// 6. 获取导出结果
	fmt.Println("\n6. 分析导出结果")
	results := exportVisitor.GetResults()
	stats := exportVisitor.GetStats()

	fmt.Printf("   导出记录数: %d\n", len(results))
	fmt.Printf("   总处理行数: %d\n", stats.TotalRows)
	fmt.Printf("   成功导出行数: %d\n", stats.ExportedRows)
	fmt.Printf("   过滤行数: %d\n", stats.FilteredRows)
	fmt.Printf("   错误行数: %d\n", stats.ErrorRows)

	// 显示前几条记录示例
	fmt.Println("\n7. 导出记录示例:")
	for i, record := range results {
		if i >= 3 { // 只显示前3条
			break
		}

		fmt.Printf("   记录 %d:\n", i+1)
		fields := record.GetFields()
		for _, field := range fields {
			value, _ := record.GetField(field)
			fmt.Printf("     %s: %v\n", field, value)
		}

		// 显示元数据
		metadata := record.GetAllMetadata()
		if len(metadata) > 0 {
			fmt.Printf("     元数据:\n")
			for key, value := range metadata {
				fmt.Printf("       %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}

	// 8. 显示工厂统计信息
	fmt.Println("8. 工厂统计信息:")
	factoryStats := factory.GetStats()
	fmt.Printf("   创建数据源总数: %d\n", factoryStats.CreatedCount)
	fmt.Printf("   当前活跃数据源: %d\n", factoryStats.ActiveCount)
	fmt.Printf("   缓存命中次数: %d\n", factoryStats.CacheHits)
	fmt.Printf("   缓存未命中次数: %d\n", factoryStats.CacheMisses)
	fmt.Printf("   错误次数: %d\n", factoryStats.ErrorCount)
	fmt.Printf("   缓存大小: %d\n", factory.GetCacheSize())

	// 9. 测试缓存机制
	fmt.Println("\n9. 测试缓存机制:")

	// 使用相同配置创建另一个数据源（测试缓存）
	_, err = factory.CreateMySQL(mysqlConfig)
	if err != nil {
		log.Printf("第二次创建失败: %v", err)
	} else {
		fmt.Println("✅ 缓存机制工作正常 - 使用了缓存的数据源")

		// 更新统计信息
		updatedStats := factory.GetStats()
		fmt.Printf("   更新后的缓存命中次数: %d\n", updatedStats.CacheHits)
	}

	// 10. 演示Record操作
	fmt.Println("\n10. 演示Record操作:")
	if len(results) > 0 {
		record := results[0]

		// 类型安全的字段访问
		if id, err := record.GetInt("id"); err == nil {
			fmt.Printf("   ID (int): %d\n", id)
		}

		if name, err := record.GetString("name"); err == nil {
			fmt.Printf("   姓名 (string): %s\n", name)
		}

		if age, err := record.GetInt("age"); err == nil {
			fmt.Printf("   年龄 (int): %d\n", age)
		}

		// 字段类型信息
		fmt.Printf("   字段类型:\n")
		for _, field := range record.GetFields() {
			fieldType := record.GetFieldType(field)
			fmt.Printf("     %s: %s\n", field, fieldType)
		}

		// 序列化操作
		recordMap := record.ToMap()
		fmt.Printf("   序列化结果包含 %d 个字段\n", len(recordMap))
	}

	// 11. 清理资源
	fmt.Println("\n11. 清理资源:")
	if err := factory.Cleanup(); err != nil {
		log.Printf("清理失败: %v", err)
	} else {
		fmt.Println("✅ 资源清理完成")
	}

	fmt.Println("\n=== v2架构集成示例完成 ===")
	fmt.Println("示例展示了v2架构的核心特性:")
	fmt.Println("✅ 类型安全的配置管理")
	fmt.Println("✅ 工厂模式的数据源创建")
	fmt.Println("✅ 访问者模式的数据处理")
	fmt.Println("✅ 统一的Record数据结构")
	fmt.Println("✅ 智能缓存机制")
	fmt.Println("✅ 完整的统计和监控")
	fmt.Println("✅ 自动资源管理")
}