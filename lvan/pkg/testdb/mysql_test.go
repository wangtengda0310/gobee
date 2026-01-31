package testdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMySQLFramework 测试MySQL框架核心功能
func TestMySQLFramework(t *testing.T) {
	t.Run("创建MySQL测试数据库", func(t *testing.T) {
		// 创建测试配置
		config := CreateTestMySQLConfig()

		// 通过工厂创建MySQL测试数据库
		mysqlDB, err := GlobalFactory.CreateMySQL(config)
		require.NoError(t, err)
		assert.NotNil(t, mysqlDB)

		// 测试初始状态
		assert.Equal(t, config.Server.Database, mysqlDB.(*MySQLTestDatabaseImpl).config.Server.Database)
		assert.Equal(t, config.Server.Port, mysqlDB.(*MySQLTestDatabaseImpl).config.Server.Port)
	})

	t.Run("启动和停止MySQL测试数据库", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		mysqlDB, err := GlobalFactory.CreateMySQL(config)
		require.NoError(t, err)

		// 启动数据库（应该使用模拟模式，因为没有真实MySQL服务器）
		err = mysqlDB.Start()
		assert.NoError(t, err)
		assert.True(t, mysqlDB.(*MySQLTestDatabaseImpl).isStarted)

		// 获取连接
		conn := mysqlDB.GetMySQLConnection()
		assert.NotNil(t, conn)
		assert.Empty(t, conn.DSN) // 模拟模式下DSN为空

		// 停止数据库
		err = mysqlDB.Stop()
		assert.NoError(t, err)
		assert.False(t, mysqlDB.(*MySQLTestDatabaseImpl).isStarted)
	})

	t.Run("测试数据操作", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		mysqlDB, err := GlobalFactory.CreateMySQL(config)
		require.NoError(t, err)

		// 启动数据库
		err = mysqlDB.Start()
		require.NoError(t, err)
		defer mysqlDB.Stop()

		// 查询测试数据
		results, err := mysqlDB.Query("SELECT * FROM user")
		assert.NoError(t, err)
		assert.Len(t, results, 2) // 应该有2条测试数据

		// 验证第一条数据
		if len(results) > 0 {
			assert.Contains(t, results[0], "uid")
			assert.Contains(t, results[0], "accountid")
			assert.Equal(t, []byte("1"), results[0]["uid"])
			assert.Equal(t, []byte("test_user_001"), results[0]["accountid"])
		}
	})

	t.Run("测试插入新数据", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		mysqlDB, err := GlobalFactory.CreateMySQL(config)
		require.NoError(t, err)

		// 启动数据库
		err = mysqlDB.Start()
		require.NoError(t, err)
		defer mysqlDB.Stop()

		// 插入新数据
		newData := []map[string]interface{}{
			{
				"accountid":  "test_user_003",
				"data":       []byte("test data 003"),
				"created_at": "2025-01-10 12:00:00",
			},
		}

		err = mysqlDB.InsertData("user", newData)
		assert.NoError(t, err)

		// 查询验证数据
		results, err := mysqlDB.Query("SELECT * FROM user")
		assert.NoError(t, err)
		assert.Len(t, results, 3) // 现在应该有3条数据
	})

	t.Run("测试清理数据", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		mysqlDB, err := GlobalFactory.CreateMySQL(config)
		require.NoError(t, err)

		// 启动数据库
		err = mysqlDB.Start()
		require.NoError(t, err)
		defer mysqlDB.Stop()

		// 验证初始数据
		results, err := mysqlDB.Query("SELECT * FROM user")
		assert.NoError(t, err)
		assert.Greater(t, len(results), 0)

		// 清理数据
		err = mysqlDB.Clear()
		assert.NoError(t, err)

		// 验证数据已清理
		results, err = mysqlDB.Query("SELECT * FROM user")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("测试SQL构建", func(t *testing.T) {
		mysqlDB := &MySQLTestDatabaseImpl{}

		// 测试CREATE TABLE SQL构建
		schema := TableSchema{
			Name: "test_table",
			Columns: []ColumnSchema{
				{Name: "id", Type: "INT", Primary: true, Nullable: false, AutoIncrement: true},
				{Name: "name", Type: "VARCHAR(100)", Nullable: false},
				{Name: "email", Type: "VARCHAR(255)", Nullable: true, Unique: true},
			},
		}

		createSQL := mysqlDB.buildCreateTableSQL(schema)
		assert.Contains(t, createSQL, "CREATE TABLE test_table")
		assert.Contains(t, createSQL, "id INT PRIMARY KEY AUTO_INCREMENT")
		assert.Contains(t, createSQL, "name VARCHAR(100) NOT NULL")
		assert.Contains(t, createSQL, "email VARCHAR(255) UNIQUE")

		// 测试INSERT SQL构建
		data := map[string]interface{}{
			"name":  "test_user",
			"email": "test@example.com",
		}

		insertSQL, args := mysqlDB.buildInsertSQL("test_table", schema, data)
		assert.Contains(t, insertSQL, "INSERT INTO test_table")
		assert.Contains(t, insertSQL, "name")
		assert.Contains(t, insertSQL, "email")
		assert.Len(t, args, 2)
		assert.Equal(t, "test_user", args[0])
		assert.Equal(t, "test@example.com", args[1])

		// 测试CREATE INDEX SQL构建
		index := IndexSchema{
			Name:    "idx_name",
			Columns: []string{"name"},
			Unique:  false,
		}

		indexSQL := mysqlDB.buildCreateIndexSQL("test_table", index)
		assert.Equal(t, "CREATE INDEX idx_name ON test_table (name)", indexSQL)
	})
}

// TestMySQLConfigValidation 测试MySQL配置验证
func TestMySQLConfigValidation(t *testing.T) {
	t.Run("有效配置验证", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		err := ValidateMySQLConfig(config)
		assert.NoError(t, err)
	})

	t.Run("无效端口验证", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		config.Server.Port = 99999
		err := ValidateMySQLConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的端口号")
	})

	t.Run("缺少主键验证", func(t *testing.T) {
		config := MySQLConfig{
			Server: MySQLServerConfig{
				Port:     3307,
				Database: "testdb",
				User:     "root",
			},
			Schemas: []TableSchema{
				{
					Name: "test_table",
					Columns: []ColumnSchema{
						{Name: "name", Type: "VARCHAR(100)", Nullable: false},
					},
				},
			},
		}
		err := ValidateMySQLConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "需要一个主键")
	})
}

// TestMySQLFactoryPattern 测试MySQL工厂模式
func TestMySQLFactoryPattern(t *testing.T) {
	factory := GlobalFactory
	assert.NotNil(t, factory)

	t.Run("工厂创建MySQL实例", func(t *testing.T) {
		config := CreateTestMySQLConfig()
		mysqlDB, err := factory.CreateMySQL(config)
		assert.NoError(t, err)
		assert.NotNil(t, mysqlDB)

		// 验证接口实现
		var _ MySQLTestDatabase = mysqlDB
		var _ TestDatabase = mysqlDB
	})
}

// TestMySQLTypeMapping 测试MySQL类型映射
func TestMySQLTypeMapping(t *testing.T) {
	mysqlDB := &MySQLTestDatabaseImpl{}

	typeTests := []struct {
		input    string
		expected string
	}{
		{"INT", "INT"},
		{"VARCHAR", "VARCHAR(255)"},
		{"TEXT", "TEXT"},
		{"BLOB", "BLOB"},
		{"TIMESTAMP", "TIMESTAMP"},
		{"BOOLEAN", "BOOLEAN"},
		{"FLOAT", "FLOAT"},
		{"UNKNOWN", "VARCHAR(255)"},
	}

	for _, test := range typeTests {
		t.Run(test.input, func(t *testing.T) {
			result := mysqlDB.mapColumnType(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}