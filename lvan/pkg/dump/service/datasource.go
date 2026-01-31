package service

import (
	"database/sql"
	"fmt"
)

// Manager 数据源管理器接口
type Manager interface {
	GetDB() *sql.DB
	GetConfig() Config
	Close() error
}

// Config 数据源配置
type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	Table    string
}

// DSN 返回数据源连接字符串
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.Database)
}
