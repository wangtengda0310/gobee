package v2

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// MySQLOptions MySQL导出选项
type MySQLOptions struct {
	Where        string   // WHERE条件
	IDs          []string // ID列表
	Fields       []string // 字段列表
	Limit        int      // 限制条数
	OutputFormat string   // 输出格式
	OutputPath   string   // 输出路径
	Compression  string   // 压缩方式
}

// MySQLExportVisitor MySQL导出访问者
// 实现DataVisitor接口，专门处理MySQL数据导出
type MySQLExportVisitor interface {
	DataVisitor

	// 获取导出选项
	GetOptions() *MySQLOptions

	// 检查是否有过滤器
	HasFilter() bool

	// 检查输出格式是否有效
	IsValidOutputFormat() bool

	// 获取导出结果
	GetResults() []Record

	// 获取导出统计信息
	GetStats() *ExportStats
}

// mysqlExportVisitorImpl MySQL导出访问者的具体实现
type mysqlExportVisitorImpl struct {
	options *MySQLOptions
	results []Record
	stats   *ExportStats
}

// ExportStats 导出统计信息
type ExportStats struct {
	TotalRows     int64     // 总行数
	ExportedRows  int64     // 已导出行数
	FilteredRows  int64     // 被过滤的行数
	ErrorRows     int64     // 错误行数
	StartTime     string    // 开始时间
	EndTime       string    // 结束时间
	ProcessedSize int64     // 处理的数据大小（字节）
}

// NewMySQLExportVisitor 创建MySQL导出访问者
func NewMySQLExportVisitor(options *MySQLOptions) MySQLExportVisitor {
	// 设置默认值
	if options.Limit <= 0 {
		options.Limit = 10000 // 默认限制1万条
	}

	if options.OutputFormat == "" {
		options.OutputFormat = "dir" // 默认输出到目录
	}

	if options.Compression == "" {
		options.Compression = "gzip" // 默认gzip压缩
	}

	return &mysqlExportVisitorImpl{
		options: options,
		results: make([]Record, 0),
		stats: &ExportStats{
			TotalRows:     0,
			ExportedRows:  0,
			FilteredRows:  0,
			ErrorRows:     0,
			ProcessedSize: 0,
			StartTime:     time.Now().Format(time.RFC3339), // 初始化开始时间
		},
	}
}

// VisitMySQL 访问MySQL数据源
func (v *mysqlExportVisitorImpl) VisitMySQL(ds MySQLDatasource) {
	// 记录开始时间
	v.stats.StartTime = time.Now().Format(time.RFC3339)

	// 获取数据库连接
	conn := ds.GetConnection()
	if conn == nil {
		panic("MySQL数据源连接为空")
	}

	sqlConn, ok := conn.(*sql.DB)
	if !ok {
		panic("MySQL数据源连接类型错误")
	}

	// 获取数据库和表信息
	database := ds.GetDatabase()
	table := ds.GetTable()

	log.Printf("开始导出数据: %s.%s, 条件: %s, IDs: %v", database, table, v.options.Where, v.options.IDs)

	// 构建查询
	query := v.buildRealQuery(database, table)
	log.Printf("执行查询: %s", query)

	// 执行真实查询
	v.executeRealQuery(sqlConn, query)

	// 处理输出格式
	v.handleOutput()

	// 记录结束时间
	v.stats.EndTime = time.Now().Format(time.RFC3339)
	log.Printf("导出完成: 总行数=%d, 导出行数=%d, 耗时=%v",
		v.stats.TotalRows, v.stats.ExportedRows, time.Since(time.Now()))
}

// VisitRedis 访问Redis数据源（MySQL导出访问者不支持）
func (v *mysqlExportVisitorImpl) VisitRedis(ds RedisDatasource) {
	panic(fmt.Sprintf("MySQL导出访问者不支持Redis数据源"))
}

// VisitDatasource 访问通用数据源
func (v *mysqlExportVisitorImpl) VisitDatasource(ds Datasource) {
	panic(fmt.Sprintf("MySQL导出访问者需要具体的MySQL数据源，收到: %T", ds))
}

// GetOptions 获取导出选项
func (v *mysqlExportVisitorImpl) GetOptions() *MySQLOptions {
	return v.options
}

// HasFilter 检查是否有过滤器
func (v *mysqlExportVisitorImpl) HasFilter() bool {
	return v.options.Where != "" || len(v.options.IDs) > 0 || len(v.options.Fields) > 0
}

// IsValidOutputFormat 检查输出格式是否有效
func (v *mysqlExportVisitorImpl) IsValidOutputFormat() bool {
	validFormats := []string{"zip", "dir", "sql-tpl", "csv", "json"}
	for _, format := range validFormats {
		if v.options.OutputFormat == format {
			return true
		}
	}
	return false
}

// GetResults 获取导出结果
func (v *mysqlExportVisitorImpl) GetResults() []Record {
	return v.results
}

// GetStats 获取导出统计信息
func (v *mysqlExportVisitorImpl) GetStats() *ExportStats {
	return v.stats
}

// buildRealQuery 构建真实的SQL查询语句
func (v *mysqlExportVisitorImpl) buildRealQuery(database, table string) string {
	query := fmt.Sprintf("SELECT")

	// 构建字段列表
	if len(v.options.Fields) > 0 {
		query += " " + strings.Join(v.options.Fields, ", ")
	} else {
		query += " *"
	}

	query += fmt.Sprintf(" FROM %s.%s", database, table)

	// 添加WHERE条件
	conditions := []string{}

	if v.options.Where != "" {
		conditions = append(conditions, v.options.Where)
	}

	if len(v.options.IDs) > 0 {
		// 构建ID条件（假设主键名为id）
		idList := make([]string, len(v.options.IDs))
		for i, id := range v.options.IDs {
			idList[i] = fmt.Sprintf("'%s'", id)
		}
		idCondition := fmt.Sprintf("id IN (%s)", strings.Join(idList, ", "))
		conditions = append(conditions, idCondition)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// 添加LIMIT
	if v.options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", v.options.Limit)
	}

	return query
}

// executeRealQuery 执行真实的数据库查询
func (v *mysqlExportVisitorImpl) executeRealQuery(db *sql.DB, query string) {
	// 执行查询
	rows, err := db.Query(query)
	if err != nil {
		panic(fmt.Sprintf("执行查询失败: %s, 错误: %v", query, err))
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		panic(fmt.Sprintf("获取列信息失败: %v", err))
	}

	if len(columns) == 0 {
		log.Printf("查询结果为空: %s", query)
		return
	}

	log.Printf("查询到 %d 个列: %v", len(columns), columns)

	// 遍历结果集
	for rows.Next() {
		// 创建值容器
		values := make([][]byte, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(pointers...); err != nil {
			log.Printf("扫描行数据失败: %v", err)
			v.stats.ErrorRows++
			continue
		}

		// 创建Record - 将dump.Record转换为map[string]interface{}
		recordMap := make(map[string]interface{})
		for i, column := range columns {
			if values[i] != nil {
				recordMap[column] = string(values[i]) // 将[]byte转换为string
			} else {
				recordMap[column] = nil
			}
		}

		// 设置元数据
		recordMap["query"] = query
		recordMap["source"] = "mysql"
		recordMap["operation"] = "export"
		recordMap["export_time"] = time.Now().Format(time.RFC3339)

		// 创建Record实例
		newRecord := NewRecordWithData(recordMap)

		v.results = append(v.results, newRecord)
		v.stats.ExportedRows++
		v.stats.TotalRows++

		// 计算处理的数据大小
		for _, value := range values {
			v.stats.ProcessedSize += int64(len(value))
		}
	}

	// 检查遍历过程中的错误
	if err = rows.Err(); err != nil {
		log.Printf("遍历结果集时出错: %v", err)
	}

	log.Printf("查询完成，共处理 %d 行数据", v.stats.ExportedRows)
}

// BuildQuery 构建SQL查询语句（导出为公共方法用于测试，保持向后兼容）
func (v *mysqlExportVisitorImpl) BuildQuery(database, table string) string {
	return v.buildRealQuery(database, table)
}

// handleOutput 处理输出格式
func (v *mysqlExportVisitorImpl) handleOutput() {
	switch v.options.OutputFormat {
	case "zip":
		v.outputToZip()
	case "dir":
		v.outputToDir()
	case "sql-tpl":
		v.outputToSQLTemplate()
	case "csv":
		v.outputToCSV()
	case "json":
		v.outputToJSON()
	default:
		// 默认输出到目录
		v.outputToDir()
	}
}

// outputToZip 输出到ZIP文件
func (v *mysqlExportVisitorImpl) outputToZip() {
	// 模拟ZIP输出实现
	// 实际应该：
	// 1. 创建ZIP文件
	// 2. 将Record数据写入ZIP中的各个文件
	// 3. 添加元数据文件
}

// outputToDir 输出到目录
func (v *mysqlExportVisitorImpl) outputToDir() {
	// 模拟目录输出实现
	// 实际应该：
	// 1. 创建输出目录
	// 2. 为每个表创建单独的文件
	// 3. 写入数据文件
	// 4. 创建元数据文件
}

// outputToSQLTemplate 输出到SQL模板
func (v *mysqlExportVisitorImpl) outputToSQLTemplate() {
	// 模拟SQL模板输出实现
	// 实际应该：
	// 1. 读取SQL模板文件
	// 2. 解析模板变量
	// 3. 替换变量为实际数据
	// 4. 生成最终的SQL文件
}

// outputToCSV 输出到CSV文件
func (v *mysqlExportVisitorImpl) outputToCSV() {
	// 模拟CSV输出实现
	// 实际应该：
	// 1. 创建CSV文件
	// 2. 写入CSV头部
	// 3. 写入数据行
	// 4. 处理特殊字符转义
}

// outputToJSON 输出到JSON文件
func (v *mysqlExportVisitorImpl) outputToJSON() {
	// 模拟JSON输出实现
	// 实际应该：
	// 1. 创建JSON文件
	// 2. 序列化Record数据
	// 3. 写入JSON数组或对象
	// 4. 处理格式化选项
}