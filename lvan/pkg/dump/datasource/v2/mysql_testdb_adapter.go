package v2

import (
	"fmt"
	"log"
	"sync"

	"github.com/wangtengda0310/gobee/lvan/pkg/testdb"
)

// MySQLTestDBAdapter MySQL测试数据库适配器
type MySQLTestDBAdapter struct {
	mysqlDB testdb.MySQLTestDatabase
	mutex   sync.RWMutex
}

var (
	// 全局适配器实例
	globalMySQLTestDBAdapter *MySQLTestDBAdapter
	mysqlAdapterOnce         sync.Once
)

// GetGlobalMySQLTestDBAdapter 获取全局MySQL测试数据库适配器实例
func GetGlobalMySQLTestDBAdapter() *MySQLTestDBAdapter {
	mysqlAdapterOnce.Do(func() {
		globalMySQLTestDBAdapter = &MySQLTestDBAdapter{}
	})
	return globalMySQLTestDBAdapter
}

// SetupMySQLTestDB 设置MySQL测试数据库
func (a *MySQLTestDBAdapter) SetupMySQLTestDB(configPath string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 如果已经有实例，先清理
	if a.mysqlDB != nil {
		if err := a.mysqlDB.Stop(); err != nil {
			log.Printf("停止MySQL测试数据库失败: %v", err)
		}
	}

	// 创建新的MySQL测试数据库
	mysqlDB, err := testdb.GlobalFactory.CreateMySQLFromConfig(configPath)
	if err != nil {
		return fmt.Errorf("创建MySQL测试数据库失败: %v", err)
	}

	// 启动数据库
	if err := mysqlDB.Start(); err != nil {
		return fmt.Errorf("启动MySQL测试数据库失败: %v", err)
	}

	a.mysqlDB = mysqlDB
	log.Printf("MySQL测试数据库已启动")
	return nil
}

// GetMySQLTestDB 获取MySQL测试数据库
func (a *MySQLTestDBAdapter) GetMySQLTestDB() testdb.MySQLTestDatabase {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.mysqlDB
}

// Cleanup 清理所有测试数据库
func (a *MySQLTestDBAdapter) Cleanup() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.mysqlDB != nil {
		if err := a.mysqlDB.Stop(); err != nil {
			log.Printf("停止MySQL测试数据库失败: %v", err)
		}
		a.mysqlDB = nil
	}

	log.Println("MySQL测试数据库已清理")
}

// CreateTestMySQLConfig 为测试创建MySQL配置
func CreateTestMySQLConfig() testdb.MySQLConfig {
	return testdb.CreateTestMySQLConfig()
}

// SetupTestMySQLEnvironment 设置MySQL测试环境
func SetupTestMySQLEnvironment() error {
	adapter := GetGlobalMySQLTestDBAdapter()

	// 创建测试配置
	mysqlConfig := CreateTestMySQLConfig()
	mysqlDB, err := testdb.GlobalFactory.CreateMySQL(mysqlConfig)
	if err != nil {
		return fmt.Errorf("创建MySQL测试数据库失败: %v", err)
	}

	// 启动数据库
	if err := mysqlDB.Start(); err != nil {
		return fmt.Errorf("启动MySQL测试数据库失败: %v", err)
	}

	adapter.mutex.Lock()
	adapter.mysqlDB = mysqlDB
	adapter.mutex.Unlock()

	log.Println("MySQL测试环境设置完成")
	return nil
}

// TeardownTestMySQLEnvironment 拆卸MySQL测试环境
func TeardownTestMySQLEnvironment() {
	if globalMySQLTestDBAdapter != nil {
		globalMySQLTestDBAdapter.Cleanup()
	}
}