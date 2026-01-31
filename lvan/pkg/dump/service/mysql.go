package service

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLManager MySQL 数据源管理器
type MySQLManager struct {
	db     *sql.DB
	config Config
}

// NewMySQLManager 创建 MySQL 管理器
func NewMySQLManager(ctx context.Context, config Config) (*MySQLManager, error) {
	// 直接构建 DSN，不使用 Config.DSN() 方法
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping 失败: %w", err)
	}

	return &MySQLManager{
		db:     db,
		config: config,
	}, nil
}

func (m *MySQLManager) GetDB() *sql.DB {
	return m.db
}

func (m *MySQLManager) GetConfig() Config {
	return m.config
}

func (m *MySQLManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
