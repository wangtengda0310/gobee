package testdb

import (
	"fmt"
)

// MySQLConfig MySQL配置结构
type MySQLConfig struct {
	Server   MySQLServerConfig   `yaml:"server"`
	Schemas  []TableSchema       `yaml:"schemas"`
	Variables map[string]interface{} `yaml:"variables,omitempty"`
}

// MySQLServerConfig MySQL服务器配置
type MySQLServerConfig struct {
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host,omitempty"`
}


// ValidateMySQLConfig 验证MySQL配置
func ValidateMySQLConfig(config MySQLConfig) error {
	// 端口验证
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("无效的端口号: %d", config.Server.Port)
	}

	// 数据库名验证
	if config.Server.Database == "" {
		return fmt.Errorf("数据库名不能为空")
	}

	// 表结构验证
	for _, schema := range config.Schemas {
		if err := validateTableSchema(schema); err != nil {
			return fmt.Errorf("表 %s 配置错误: %v", schema.Name, err)
		}
	}

	return nil
}

// validateTableSchema 验证表结构
func validateTableSchema(schema TableSchema) error {
	if schema.Name == "" {
		return fmt.Errorf("表名不能为空")
	}

	if len(schema.Columns) == 0 {
		return fmt.Errorf("表必须至少有一列")
	}

	// 检查主键
	hasPrimary := false
	for _, col := range schema.Columns {
		if col.Primary {
			hasPrimary = true
			break
		}
	}

	if !hasPrimary {
		return fmt.Errorf("表 %s 需要一个主键", schema.Name)
	}

	return nil
}

// CreateTestMySQLConfig 创建测试用的MySQL配置
func CreateTestMySQLConfig() MySQLConfig {
	return MySQLConfig{
		Server: MySQLServerConfig{
			Host:     "127.0.0.1",
			Port:     3307,
			Database: "testdb",
			User:     "root",
			Password: "",
		},
		Schemas: []TableSchema{
			{
				Name: "user",
				Columns: []ColumnSchema{
					{Name: "uid", Type: "INT", Primary: true, Nullable: false, AutoIncrement: true},
					{Name: "accountid", Type: "VARCHAR(50)", Nullable: false, Unique: true},
					{Name: "data", Type: "BLOB", Nullable: true},
					{Name: "created_at", Type: "TIMESTAMP", Nullable: false},
				},
				Indexes: []IndexSchema{
					{Name: "idx_accountid", Columns: []string{"accountid"}, Unique: true},
				},
				Data: []map[string]interface{}{
					{
						"uid":        1,
						"accountid":  "test_user_001",
						"data":       []byte("test data 001"),
						"created_at": "2025-01-10 10:00:00",
					},
					{
						"uid":        2,
						"accountid":  "test_user_002",
						"data":       []byte("test data 002"),
						"created_at": "2025-01-10 11:00:00",
					},
				},
			},
		},
	}
}

