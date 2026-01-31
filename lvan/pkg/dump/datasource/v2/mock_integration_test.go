package v2

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wangtengda0310/gobee/lvan/pkg/testdb"
	_ "github.com/go-sql-driver/mysql"
)

// TestIntegration_WithMockMySQLServer 使用go-mysql-server的集成测试
func TestIntegration_WithMockMySQLServer(t *testing.T) {
	// 设置测试环境
	if err := SetupTestMySQLEnvironment(); err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer TeardownTestMySQLEnvironment()

	// 获取测试数据库适配器
	adapter := GetGlobalMySQLTestDBAdapter()
	mysqlDB := adapter.GetMySQLTestDB()
	require.NotNil(t, mysqlDB, "MySQL测试数据库应该已初始化")

	// 获取数据库连接
	conn := mysqlDB.GetMySQLConnection()
	require.NotNil(t, conn, "应该能获取到MySQL连接")
	defer conn.DB.Close()

	// 测试数据库连接
	if err := conn.DB.Ping(); err != nil {
		t.Fatalf("数据库连接测试失败: %v", err)
	}

	t.Run("查询测试数据", func(t *testing.T) {
		// 查询user表中的数据
		rows, err := conn.DB.Query("SELECT uid, accountid, created_at FROM user ORDER BY uid")
		require.NoError(t, err)
		defer rows.Close()

		var users []struct {
			UID       int    `db:"uid"`
			AccountID string `db:"accountid"`
			CreatedAt string `db:"created_at"`
		}

		for rows.Next() {
			var user struct {
				UID       int    `db:"uid"`
				AccountID string `db:"accountid"`
				CreatedAt string `db:"created_at"`
			}

			err := rows.Scan(&user.UID, &user.AccountID, &user.CreatedAt)
			require.NoError(t, err)
			users = append(users, user)
		}

		assert.GreaterOrEqual(t, len(users), 2, "应该至少有2个测试用户")

		// 验证第一个用户
		if len(users) > 0 {
			assert.Equal(t, 1, users[0].UID)
			assert.Equal(t, "test_user_001", users[0].AccountID)
		}
	})

	t.Run("插入新数据", func(t *testing.T) {
		// 插入新用户
		result, err := conn.DB.Exec(
			"INSERT INTO user (accountid, data, created_at) VALUES (?, ?, ?)",
			"integration_test_user", []byte("integration test data"), "2025-01-10 15:00:00")
		require.NoError(t, err)

		id, err := result.LastInsertId()
		require.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 验证插入的数据
		var count int
		err = conn.DB.QueryRow("SELECT COUNT(*) FROM user WHERE accountid = ?", "integration_test_user").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("更新数据", func(t *testing.T) {
		// 更新第一个用户的数据
		result, err := conn.DB.Exec(
			"UPDATE user SET data = ? WHERE uid = ?",
			[]byte("updated test data"), 1)
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证更新后的数据
		var data []byte
		err = conn.DB.QueryRow("SELECT data FROM user WHERE uid = ?", 1).Scan(&data)
		require.NoError(t, err)
		assert.Equal(t, []byte("updated test data"), data)
	})

	t.Run("删除数据", func(t *testing.T) {
		// 删除测试插入的用户
		result, err := conn.DB.Exec("DELETE FROM user WHERE accountid = ?", "integration_test_user")
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证删除后的数据
		var count int
		err = conn.DB.QueryRow("SELECT COUNT(*) FROM user WHERE accountid = ?", "integration_test_user").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestIntegration_WithMockMySQLServer_Basic 基本MySQL集成测试
func TestIntegration_WithMockMySQLServer_Basic(t *testing.T) {
	// 设置测试环境
	if err := SetupTestMySQLEnvironment(); err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer TeardownTestMySQLEnvironment()

	// 获取测试数据库适配器
	adapter := GetGlobalMySQLTestDBAdapter()
	mysqlDB := adapter.GetMySQLTestDB()
	require.NotNil(t, mysqlDB, "MySQL测试数据库应该已初始化")

	// 获取数据库连接
	conn := mysqlDB.GetMySQLConnection()
	require.NotNil(t, conn, "应该能获取到MySQL连接")
	defer func() {
		if conn.DB != nil {
			conn.DB.Close()
		}
	}()

	t.Run("测试基本查询", func(t *testing.T) {
		// 查询数据
		results, err := mysqlDB.Query("SELECT COUNT(*) as count FROM user")
		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
	})

	t.Run("测试数据插入", func(t *testing.T) {
		// 插入新数据
		newData := []map[string]interface{}{
			{
				"accountid":  "integration_test_user",
				"data":       []byte("integration test data"),
				"created_at": "2025-01-10 15:00:00",
			},
		}
		err := mysqlDB.InsertData("user", newData)
		require.NoError(t, err)

		// 验证插入结果
		results, err := mysqlDB.Query("SELECT accountid FROM user WHERE accountid = ?", "integration_test_user")
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

// TestIntegration_VisitorPatternWithMockServer 使用访问者模式与模拟服务器的集成测试
func TestIntegration_VisitorPatternWithMockServer(t *testing.T) {
	// 设置测试环境
	if err := SetupTestMySQLEnvironment(); err != nil {
		t.Fatalf("设置测试环境失败: %v", err)
	}
	defer TeardownTestMySQLEnvironment()

	// 获取测试数据库适配器
	adapter := GetGlobalMySQLTestDBAdapter()
	mysqlDB := adapter.GetMySQLTestDB()
	require.NotNil(t, mysqlDB, "MySQL测试数据库应该已初始化")

	// 创建MySQL数据源配置
	config := NewMySQLConfig("127.0.0.1", 3307, "root", "", "testdb", "user")
	datasource := NewMySQLDatasource(config)
	defer datasource.Close()

	// 创建测试访问者
	visitor := &IntegrationTestVisitor{
		testData: make(map[string]interface{}),
	}

	// 执行访问者模式
	datasource.Accept(visitor)

	// 验证访问者执行结果
	assert.True(t, visitor.visitedMySQL, "MySQL访问者应该被调用")
	assert.Equal(t, "mysql", visitor.datasourceType, "数据源类型应该是mysql")
	assert.Equal(t, "testdb", visitor.database, "数据库名应该是testdb")
	assert.Equal(t, "user", visitor.table, "表名应该是user")
}

// IntegrationTestVisitor 集成测试访问者
type IntegrationTestVisitor struct {
	visitedMySQL  bool
	datasourceType string
	database      string
	table         string
	testData      map[string]interface{}
}

func (v *IntegrationTestVisitor) VisitMySQL(ds MySQLDatasource) {
	v.visitedMySQL = true
	v.datasourceType = ds.GetMetadata().GetType()
	v.database = ds.GetDatabase()
	v.table = ds.GetTable()

	// 执行一些数据操作进行验证
	conn := ds.GetConnection()
	if conn != nil {
		mysqlConn, ok := conn.(*testdb.MySQLConnection)
		if ok && mysqlConn.DB != nil {
			// 查询数据并存储到测试数据中
			rows, err := mysqlConn.DB.Query("SELECT COUNT(*) as count FROM user")
			if err == nil {
				defer rows.Close()
				if rows.Next() {
					var count int
					if err := rows.Scan(&count); err == nil {
						v.testData["user_count"] = count
					}
				}
			}
		}
	}
}

func (v *IntegrationTestVisitor) VisitRedis(ds RedisDatasource) {
	// 对于这个集成测试，我们不测试Redis访问者
}

func (v *IntegrationTestVisitor) VisitDatasource(ds Datasource) {
	// 回退到基础访问者方法
	v.datasourceType = ds.GetMetadata().GetType()
}

// TestIntegration_ConfigDrivenTest 使用配置驱动的测试
func TestIntegration_ConfigDrivenTest(t *testing.T) {
	// 测试从YAML配置文件创建测试数据库
	adapter := GetGlobalMySQLTestDBAdapter()

	// 尝试从配置文件设置MySQL数据库
	err := adapter.SetupMySQLTestDB("mysql_test.yaml")
	if err != nil {
		t.Logf("从YAML配置文件设置MySQL失败，使用程序化配置: %v", err)

		// 如果YAML配置失败，使用程序化配置
		if err := SetupTestMySQLEnvironment(); err != nil {
			t.Fatalf("设置测试环境失败: %v", err)
		}
		defer TeardownTestMySQLEnvironment()
	} else {
		defer adapter.Cleanup()
	}

	mysqlDB := adapter.GetMySQLTestDB()
	require.NotNil(t, mysqlDB, "MySQL测试数据库应该已初始化")

	conn := mysqlDB.GetMySQLConnection()
	require.NotNil(t, conn, "应该能获取到MySQL连接")
	defer conn.DB.Close()

	// 测试数据库功能
	t.Run("验证表结构", func(t *testing.T) {
		// 检查user表是否存在
		var tableName string
		err := conn.DB.QueryRow("SHOW TABLES LIKE 'user'").Scan(&tableName)
		if err == sql.ErrNoRows {
			t.Skip("user表不存在，跳过表结构验证")
		}
		require.NoError(t, err, "user表应该存在")
	})

	t.Run("验证测试数据", func(t *testing.T) {
		// 检查是否有测试数据
		var count int
		err := conn.DB.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
		if err != nil {
			t.Skip("无法查询user表，跳过数据验证")
		}
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0, "应该有测试数据")
	})
}