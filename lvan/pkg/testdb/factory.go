package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// 全局工厂实例
	GlobalFactory TestDatabaseFactory = &testDatabaseFactoryImpl{}
)

// testDatabaseFactoryImpl 工厂实现
type testDatabaseFactoryImpl struct{}

// CreateMySQLFromConfig 从配置文件创建MySQL测试数据库
func (f *testDatabaseFactoryImpl) CreateMySQLFromConfig(configPath string) (MySQLTestDatabase, error) {
	config, err := LoadMySQLConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载MySQL配置失败: %v", err)
	}
	return f.CreateMySQL(config)
}

// CreateMySQL 创建MySQL测试数据库
func (f *testDatabaseFactoryImpl) CreateMySQL(config MySQLConfig) (MySQLTestDatabase, error) {
	return &MySQLTestDatabaseImpl{
		config: config,
	}, nil
}

// MySQLTestDatabaseImpl MySQL测试数据库实现
type MySQLTestDatabaseImpl struct {
	config      MySQLConfig
	conn        *MySQLConnection
	isStarted   bool
	data        map[string][]map[string]interface{} // 内存数据存储
	autoIncrement map[string]int                     // 自增ID计数器
}

// Start 启动MySQL测试数据库
func (m *MySQLTestDatabaseImpl) Start() error {
	if m.isStarted {
		return nil
	}

	// 验证配置
	if err := ValidateMySQLConfig(m.config); err != nil {
		return err
	}

	// 初始化内存存储
	m.data = make(map[string][]map[string]interface{})
	m.autoIncrement = make(map[string]int)

	// 创建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.config.Server.User,
		m.config.Server.Password,
		m.config.Server.Host,
		m.config.Server.Port,
		m.config.Server.Database)

	// 尝试连接真实数据库，如果失败则使用模拟模式
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// 连接失败，使用模拟模式
		m.conn = &MySQLConnection{
			DSN: "", // 模拟模式下DSN为空
		}

		// 模拟模式下也需要初始化表结构和数据
		if err := m.createSchemas(); err != nil {
			return fmt.Errorf("创建表结构失败: %v", err)
		}
		if err := m.insertTestData(); err != nil {
			return fmt.Errorf("插入测试数据失败: %v", err)
		}

		m.isStarted = true
		return nil
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		// 连接失败，关闭连接并使用模拟模式
		db.Close()
		m.conn = &MySQLConnection{
			DSN: "", // 模拟模式下DSN为空
		}

		// 模拟模式下也需要初始化表结构和数据
		if err := m.createSchemas(); err != nil {
			return fmt.Errorf("创建表结构失败: %v", err)
		}
		if err := m.insertTestData(); err != nil {
			return fmt.Errorf("插入测试数据失败: %v", err)
		}

		m.isStarted = true
		return nil
	}

	// 连接成功，使用真实数据库
	m.conn = &MySQLConnection{
		DB:  db,
		DSN: dsn,
	}

	// 创建表结构
	if err := m.createSchemas(); err != nil {
		return fmt.Errorf("创建表结构失败: %v", err)
	}

	// 插入测试数据
	if err := m.insertTestData(); err != nil {
		return fmt.Errorf("插入测试数据失败: %v", err)
	}

	m.isStarted = true
	return nil
}

// Stop 停止MySQL测试数据库
func (m *MySQLTestDatabaseImpl) Stop() error {
	if !m.isStarted {
		return nil
	}

	if m.conn != nil && m.conn.DB != nil {
		m.conn.DB.Close()
	}

	m.data = nil
	m.autoIncrement = nil
	m.isStarted = false
	return nil
}

// GetConnection 获取连接
func (m *MySQLTestDatabaseImpl) GetConnection() interface{} {
	return m.conn
}

// GetMySQLConnection 获取MySQL连接
func (m *MySQLTestDatabaseImpl) GetMySQLConnection() *MySQLConnection {
	return m.conn
}

// Clear 清理数据
func (m *MySQLTestDatabaseImpl) Clear() error {
	if !m.isStarted {
		return nil
	}

	if m.conn != nil && m.conn.DB != nil {
		// 真实数据库清理
		for _, schema := range m.config.Schemas {
			_, err := m.conn.DB.Exec(fmt.Sprintf("DELETE FROM %s", schema.Name))
			if err != nil {
				return fmt.Errorf("清理表 %s 失败: %v", schema.Name, err)
			}
		}
	} else {
		// 模拟模式清理
		for tableName := range m.data {
			m.data[tableName] = make([]map[string]interface{}, 0)
		}
		// 重置自增ID
		for tableName := range m.autoIncrement {
			m.autoIncrement[tableName] = 0
		}
	}

	return nil
}

// LoadConfig 加载配置
func (m *MySQLTestDatabaseImpl) LoadConfig(configPath string) error {
	config, err := LoadMySQLConfig(configPath)
	if err != nil {
		return err
	}
	m.config = config
	return nil
}

// CreateTable 创建表
func (m *MySQLTestDatabaseImpl) CreateTable(schema *TableSchema) error {
	// 允许在启动过程中创建表（用于初始化）
	if !m.isStarted && m.data == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if m.conn != nil && m.conn.DB != nil {
		// 真实数据库创建表
		createSQL := m.buildCreateTableSQL(*schema)
		_, err := m.conn.DB.Exec(createSQL)
		if err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}

		// 创建索引
		for _, index := range schema.Indexes {
			indexSQL := m.buildCreateIndexSQL(schema.Name, index)
			_, err := m.conn.DB.Exec(indexSQL)
			if err != nil {
				return fmt.Errorf("创建索引失败: %v", err)
			}
		}
	} else {
		// 模拟模式，初始化表数据存储
		m.data[schema.Name] = make([]map[string]interface{}, 0)
		m.autoIncrement[schema.Name] = 0
	}

	return nil
}

// InsertData 插入数据
func (m *MySQLTestDatabaseImpl) InsertData(tableName string, data []map[string]interface{}) error {
	// 允许在启动过程中插入数据（用于初始化）
	if !m.isStarted && m.data == nil {
		return fmt.Errorf("数据库未初始化")
	}

	if len(data) == 0 {
		return nil
	}

	// 获取表结构
	var schema *TableSchema
	for _, s := range m.config.Schemas {
		if s.Name == tableName {
			schema = &s
			break
		}
	}

	if schema == nil {
		return fmt.Errorf("未找到表 %s 的结构定义", tableName)
	}

	if m.conn != nil && m.conn.DB != nil {
		// 真实数据库插入
		for _, row := range data {
			insertSQL, args := m.buildInsertSQL(tableName, *schema, row)
			_, err := m.conn.DB.Exec(insertSQL, args...)
			if err != nil {
				return fmt.Errorf("插入数据失败: %v", err)
			}
		}
	} else {
		// 模拟模式插入
		for _, row := range data {
			// 处理自增字段
			insertData := make(map[string]interface{})
			for k, v := range row {
				insertData[k] = v
			}

			for _, col := range schema.Columns {
				if col.AutoIncrement {
					m.autoIncrement[tableName]++
					insertData[col.Name] = m.autoIncrement[tableName]
					break
				}
			}

			m.data[tableName] = append(m.data[tableName], insertData)
		}
	}

	return nil
}

// Query 查询数据
func (m *MySQLTestDatabaseImpl) Query(query string, args ...interface{}) ([]map[string][]byte, error) {
	if !m.isStarted {
		return nil, fmt.Errorf("数据库未启动")
	}

	if m.conn != nil && m.conn.DB != nil {
		// 真实数据库查询
		rows, err := m.conn.DB.Query(query, args...)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %v", err)
		}
		defer rows.Close()

		return m.scanRowsToBytes(rows)
	} else {
		// 模拟模式查询
		return m.mockQuery(query, args...)
	}
}

// scanRowsToBytes 扫描行数据为字节数组
func (m *MySQLTestDatabaseImpl) scanRowsToBytes(rows *sql.Rows) ([]map[string][]byte, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %v", err)
	}

	var results []map[string][]byte
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %v", err)
		}

		row := make(map[string][]byte)
		for i, col := range columns {
			if values[i] != nil {
				row[col] = m.valueToBytes(values[i])
			}
		}
		results = append(results, row)
	}

	return results, nil
}

// valueToBytes 将值转换为字节数组
func (m *MySQLTestDatabaseImpl) valueToBytes(value interface{}) []byte {
	switch v := value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case int, int32, int64:
		return []byte(fmt.Sprintf("%d", v))
	case float32, float64:
		return []byte(fmt.Sprintf("%f", v))
	case bool:
		return []byte(fmt.Sprintf("%t", v))
	case time.Time:
		return []byte(v.Format("2006-01-02 15:04:05"))
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}

// mockQuery 模拟查询
func (m *MySQLTestDatabaseImpl) mockQuery(query string, args ...interface{}) ([]map[string][]byte, error) {
	// 简单的SQL解析，只支持基本的SELECT
	query = strings.ToUpper(strings.TrimSpace(query))

	if strings.HasPrefix(query, "SELECT") {
		return m.mockSelect(query, args...)
	}

	if strings.HasPrefix(query, "INSERT") {
		return m.mockInsert(query, args...)
	}

	if strings.HasPrefix(query, "UPDATE") {
		return m.mockUpdate(query, args...)
	}

	if strings.HasPrefix(query, "DELETE") {
		return m.mockDelete(query, args...)
	}

	return nil, fmt.Errorf("不支持的SQL语句: %s", query)
}

// mockSelect 模拟SELECT查询
func (m *MySQLTestDatabaseImpl) mockSelect(query string, args ...interface{}) ([]map[string][]byte, error) {
	// 简单解析：SELECT * FROM table_name
	parts := strings.Fields(query)
	if len(parts) < 4 {
		return nil, fmt.Errorf("无效的SELECT语句: %s", query)
	}

	tableName := parts[3]
	// 移除可能的分号或逗号
	tableName = strings.TrimRight(tableName, ",;")

	
	// 尝试大小写不敏感的表名匹配
	tableData, exists := m.data[tableName]
	if !exists {
		// 尝试小写匹配
		tableData, exists = m.data[strings.ToLower(tableName)]
	}
	if !exists {
		// 遍历所有表名进行匹配
		for availableTable := range m.data {
			if strings.EqualFold(availableTable, tableName) {
				tableData = m.data[availableTable]
				exists = true
				break
			}
		}
	}

	if !exists {
		return []map[string][]byte{}, nil
	}

	var results []map[string][]byte
	for _, row := range tableData {
		byteRow := make(map[string][]byte)
		for col, val := range row {
			byteRow[col] = m.valueToBytes(val)
		}
		results = append(results, byteRow)
	}

	return results, nil
}

// mockInsert 模拟INSERT语句
func (m *MySQLTestDatabaseImpl) mockInsert(query string, args ...interface{}) ([]map[string][]byte, error) {
	// 简单解析：INSERT INTO table_name ...
	parts := strings.Fields(query)
	if len(parts) < 3 {
		return nil, fmt.Errorf("无效的INSERT语句")
	}

	tableName := parts[2]
	if strings.HasSuffix(tableName, ",") {
		tableName = tableName[:len(tableName)-1]
	}

	// 返回影响的行数
	return []map[string][]byte{
		{"affected_rows": []byte("1")},
	}, nil
}

// mockUpdate 模拟UPDATE语句
func (m *MySQLTestDatabaseImpl) mockUpdate(query string, args ...interface{}) ([]map[string][]byte, error) {
	return []map[string][]byte{
		{"affected_rows": []byte("1")},
	}, nil
}

// mockDelete 模拟DELETE语句
func (m *MySQLTestDatabaseImpl) mockDelete(query string, args ...interface{}) ([]map[string][]byte, error) {
	return []map[string][]byte{
		{"affected_rows": []byte("1")},
	}, nil
}

// createSchemas 创建所有表结构
func (m *MySQLTestDatabaseImpl) createSchemas() error {
	for _, schema := range m.config.Schemas {
		if err := m.CreateTable(&schema); err != nil {
			return fmt.Errorf("创建表 %s 失败: %v", schema.Name, err)
		}
	}
	return nil
}

// insertTestData 插入所有测试数据
func (m *MySQLTestDatabaseImpl) insertTestData() error {
	for _, schema := range m.config.Schemas {
		if len(schema.Data) > 0 {
			if err := m.InsertData(schema.Name, schema.Data); err != nil {
				return fmt.Errorf("插入表 %s 测试数据失败: %v", schema.Name, err)
			}
		}
	}
	return nil
}

// buildCreateTableSQL 构建CREATE TABLE SQL
func (m *MySQLTestDatabaseImpl) buildCreateTableSQL(schema TableSchema) string {
	sql := fmt.Sprintf("CREATE TABLE %s (", schema.Name)

	for i, col := range schema.Columns {
		colDef := fmt.Sprintf("%s %s", col.Name, m.mapColumnType(col.Type))

		if col.Primary {
			colDef += " PRIMARY KEY"
		}

		if col.AutoIncrement {
			colDef += " AUTO_INCREMENT"
		}

		if !col.Nullable && !col.Primary {
			colDef += " NOT NULL"
		}

		if col.Unique {
			colDef += " UNIQUE"
		}

		if col.Default != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", col.Default)
		}

		if i < len(schema.Columns)-1 {
			colDef += ","
		}

		sql += colDef
	}

	sql += ")"
	return sql
}

// buildCreateIndexSQL 构建CREATE INDEX SQL
func (m *MySQLTestDatabaseImpl) buildCreateIndexSQL(tableName string, index IndexSchema) string {
	unique := ""
	if index.Unique {
		unique = "UNIQUE "
	}

	columns := ""
	for i, col := range index.Columns {
		if i > 0 {
			columns += ", "
		}
		columns += col
	}

	return fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)", unique, index.Name, tableName, columns)
}

// buildInsertSQL 构建INSERT SQL
func (m *MySQLTestDatabaseImpl) buildInsertSQL(tableName string, schema TableSchema, data map[string]interface{}) (string, []interface{}) {
	columns := make([]string, 0)
	placeholders := make([]string, 0)
	args := make([]interface{}, 0)

	for _, col := range schema.Columns {
		if value, exists := data[col.Name]; exists {
			columns = append(columns, col.Name)
			placeholders = append(placeholders, "?")
			args = append(args, value)
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return sql, args
}

// getAvailableTables 获取可用表名列表
func (m *MySQLTestDatabaseImpl) getAvailableTables() []string {
	var tables []string
	for tableName := range m.data {
		tables = append(tables, tableName)
	}
	return tables
}

// mapColumnType 映射列类型
func (m *MySQLTestDatabaseImpl) mapColumnType(goType string) string {
	switch goType {
	case "INT", "int":
		return "INT"
	case "VARCHAR", "string":
		return "VARCHAR(255)"
	case "TEXT", "text":
		return "TEXT"
	case "BLOB", "blob":
		return "BLOB"
	case "TIMESTAMP", "timestamp":
		return "TIMESTAMP"
	case "BOOLEAN", "bool":
		return "BOOLEAN"
	case "FLOAT", "float":
		return "FLOAT"
	case "DOUBLE", "double":
		return "DOUBLE"
	default:
		// 如果已经包含长度信息，直接返回
		if strings.Contains(goType, "(") {
			return goType
		}
		return "VARCHAR(255)"
	}
}