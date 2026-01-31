package service

import (
	"database/sql"
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
