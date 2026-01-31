package service_test

import (
	"fmt"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

// ExampleDSN 展示如何构建 MySQL DSN 连接字符串
//
// 这是一个参考示例，用于在其他项目中构建 MySQL 数据源连接字符串。
// 如果你的项目需要支持多种数据源，可以参考这个模式。
func ExampleDSN() {
	// 示例 1: 基本的 MySQL DSN 构建
	cfg := service.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "mydb",
	}

	// DSN 格式: user:password@tcp(host:port)/database?params
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	fmt.Println("MySQL DSN:", dsn)
	// 输出: root:@tcp(localhost:3306)/mydb?parseTime=true

	// 示例 2: 带密码的连接
	cfgWithPassword := service.Config{
		Host:     "192.168.1.100",
		Port:     3307,
		User:     "appuser",
		Password: "secretpass",
		Database: "production",
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=30s",
		cfgWithPassword.User, cfgWithPassword.Password,
		cfgWithPassword.Host, cfgWithPassword.Port, cfgWithPassword.Database)

	fmt.Println("MySQL DSN with password:", dsn)
	// 输出: appuser:secretpass@tcp(192.168.1.100:3307)/production?parseTime=true&timeout=30s

	// 示例 3: 使用 DSN 连接数据库
	// import (
	//     "database/sql"
	//     _ "github.com/go-sql-driver/mysql"
	// )
	//
	// db, err := sql.Open("mysql", dsn)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// defer db.Close()
	//
	// if err := db.Ping(); err != nil {
	//     log.Fatal(err)
	// }

	// 示例 4: 从环境变量或配置中心获取配置后构建
	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// database := os.Getenv("DB_NAME")
	//
	// cfg := service.Config{
	//     Host:     host,
	//     Port:     parsePort(port),
	//     User:     user,
	//     Password: password,
	//     Database: database,
	// }
	//
	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
	//     cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

// ExampleMySQLManager 展示如何创建 MySQL 管理器
//
// 这展示了在需要数据库连接的服务中，如何封装连接管理逻辑。
func ExampleMySQLManager() {
	// 示例 1: HTTP API 场景
	// 从 HTTP 请求或配置中心获取参数
	type DatabaseConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	httpConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "apiuser",
		Password: "apipass",
		Database: "apidb",
	}

	// 转换为 service.Config
	cfg := service.Config{
		Host:     httpConfig.Host,
		Port:     uint16(httpConfig.Port),
		User:     httpConfig.User,
		Password: httpConfig.Password,
		Database: httpConfig.Database,
		Table:    "", // API 场景可能不需要表名
	}

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	fmt.Println("API Service DSN:", dsn)

	// 示例 2: 批量导出服务
	// 需要根据动态参数创建多个连接
	configs := []service.Config{
		{Host: "db1.example.com", Port: 3306, User: "exporter", Password: "pass1", Database: "db1"},
		{Host: "db2.example.com", Port: 3306, User: "exporter", Password: "pass2", Database: "db2"},
	}

	for i, cfg := range configs {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		fmt.Printf("DB %d DSN: %s\n", i+1, dsn)
		// 可以并行创建多个连接
	}
}

// ExampleConfigFromCobra 展示如何从 Cobra 命令行构建 Config
//
// 这展示了在命令行工具中，如何从解析的参数构建 Config 对象。
func ExampleConfigFromCobra() {
	// 假设 Cobra 命令行参数已解析到 viper
	// --host localhost --port 3306 --user root --password "" --database mydb --table user

	// 从 viper 获取配置
	// cfg := dump.Config{}
	// viper.Unmarshal(&cfg)

	cfg := service.Config{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "mydb",
		Table:    "user",
	}

	// 使用 Config
	fmt.Printf("Connecting to %s:%d/%s.%s as %s\n",
		cfg.Host, cfg.Port, cfg.Database, cfg.Table, cfg.User)
}
