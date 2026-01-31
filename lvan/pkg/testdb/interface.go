package testdb

import (
	"database/sql"
)

// TestDatabase 测试数据库基础接口
type TestDatabase interface {
	Start() error
	Stop() error
	GetConnection() interface{}
	Clear() error
	LoadConfig(configPath string) error
}

// MySQLTestDatabase MySQL测试数据库接口
type MySQLTestDatabase interface {
	TestDatabase
	GetMySQLConnection() *MySQLConnection
	CreateTable(schema *TableSchema) error
	InsertData(tableName string, data []map[string]interface{}) error
	Query(query string, args ...interface{}) ([]map[string][]byte, error)
}

// MySQLConnection MySQL连接包装器
type MySQLConnection struct {
	DB *sql.DB
	DSN string
}

// TestDatabaseFactory 测试数据库工厂接口
type TestDatabaseFactory interface {
	CreateMySQLFromConfig(configPath string) (MySQLTestDatabase, error)
	CreateMySQL(config MySQLConfig) (MySQLTestDatabase, error)
}

// TableSchema 表结构定义
type TableSchema struct {
	Name    string         `yaml:"name"`
	Columns []ColumnSchema `yaml:"columns"`
	Indexes []IndexSchema  `yaml:"indexes"`
	Data    []map[string]interface{} `yaml:"data,omitempty"`
}

// ColumnSchema 列结构定义
type ColumnSchema struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	Primary       bool   `yaml:"primary,omitempty"`
	Nullable      bool   `yaml:"nullable,omitempty"`
	Unique        bool   `yaml:"unique,omitempty"`
	AutoIncrement bool   `yaml:"auto_increment,omitempty"`
	Default       string `yaml:"default,omitempty"`
}

// IndexSchema 索引结构定义
type IndexSchema struct {
	Name    string   `yaml:"name"`
	Columns []string `yaml:"columns"`
	Unique  bool     `yaml:"unique,omitempty"`
}