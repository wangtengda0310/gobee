package v2

import (
	"database/sql"
	"log"
)

// LegacyAdapter 提供从旧架构到新v2架构的适配器
// 这个适配器允许现有代码逐步迁移到新的访问者模式
type LegacyAdapter struct {
	v2Datasource Datasource
}

// NewLegacyAdapter 创建新架构的适配器
func NewLegacyAdapter(v2ds Datasource) *LegacyAdapter {
	return &LegacyAdapter{
		v2Datasource: v2ds,
	}
}

// ToLegacyDatasource 转换为旧式Datasource结构
func (a *LegacyAdapter) ToLegacyDatasource() *LegacyDatasource {
	// 从v2数据源获取配置信息
	var host string
	var port uint16
	var user, password, database, table string

	if mysqlDs, ok := a.v2Datasource.(MySQLDatasource); ok {
		host = mysqlDs.GetHost()
		port = uint16(mysqlDs.GetPort())
		user = mysqlDs.GetUser()
		password = mysqlDs.GetPassword()
		database = mysqlDs.GetDatabase()
		table = mysqlDs.GetTable()
	}

	// 创建旧式配置
	legacyConfig := &LegacyConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Table:    table,
	}

	return &LegacyDatasource{
		DB:          nil, // 当前实现中不包含实际连接
		LegacyConfig: legacyConfig,
	}
}

// LegacyDatasource 实现旧的dump.Datasource结构
// 这个结构体包装了新的v2数据源，使其兼容现有代码
type LegacyDatasource struct {
	*sql.DB
	*LegacyConfig
}

// LegacyConfig 模拟旧的Config结构
type LegacyConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	Table    string
}

// LegacyVisitorAdapter 将旧的访问者函数适配为新的Visitor接口
type LegacyVisitorAdapter struct {
	callback func(*LegacyDatasource)
}

// NewLegacyVisitorAdapter 从旧的访问者函数创建适配器
func NewLegacyVisitorAdapter(callback func(*LegacyDatasource)) *LegacyVisitorAdapter {
	return &LegacyVisitorAdapter{
		callback: callback,
	}
}

// VisitDatasource 实现新的Visitor接口
func (v *LegacyVisitorAdapter) VisitDatasource(ds Datasource) {
	// 将新数据源转换为旧数据源
	adapter := NewLegacyAdapter(ds)
	legacyDs := adapter.ToLegacyDatasource()

	// 调用旧的访问者函数
	v.callback(legacyDs)
}

// VisitMySQL 实现DataVisitor接口
func (v *LegacyVisitorAdapter) VisitMySQL(ds MySQLDatasource) {
	v.VisitDatasource(ds)
}

// VisitRedis 实现DataVisitor接口
func (v *LegacyVisitorAdapter) VisitRedis(ds RedisDatasource) {
	v.VisitDatasource(ds)
}

// LegacyActionBridge 提供从旧action模式到新访问者模式的桥梁
type LegacyActionBridge struct {
	v2Datasource Datasource
}

// NewLegacyActionBridge 创建旧action的桥梁
func NewLegacyActionBridge(v2ds Datasource) *LegacyActionBridge {
	return &LegacyActionBridge{
		v2Datasource: v2ds,
	}
}

// Export 使用新架构执行导出操作，但兼容旧的调用方式
func (b *LegacyActionBridge) Export(where string, ids ...string) {
	// 创建导出访问者
	options := &MySQLOptions{
		Where: where,
		IDs:   ids,
		Limit: 10000, // 默认限制
	}

	// 使用标准导出访问者
	visitor := NewMySQLExportVisitor(options)
	b.v2Datasource.Accept(visitor)
}

// MigrationHelper 提供迁移辅助功能
type MigrationHelper struct{}

// NewMigrationHelper 创建迁移辅助器
func NewMigrationHelper() *MigrationHelper {
	return &MigrationHelper{}
}

// MigrateCommand 将旧的命令迁移到新架构
func (h *MigrationHelper) MigrateCommand() {
	// 这个函数展示了如何将旧的命令迁移到新架构
	log.Println("开始迁移旧的命令架构到新的访问者模式")

	// 创建新的数据源（示例）
	config := NewMySQLConfig("localhost", 3306, "root", "", "gforge", "user")
	v2ds := NewMySQLDatasource(config)

	// 创建桥梁
	bridge := NewLegacyActionBridge(v2ds)

	// 模拟旧的调用方式
	bridge.Export("uid", "1", "2", "3")

	log.Println("迁移示例完成")
}

// WrapLegacyFunction 将旧的全局函数包装为新的访问者
func (h *MigrationHelper) WrapLegacyFunction(fn func(*LegacyDatasource)) Visitor {
	return &LegacyFunctionVisitor{
		callback: fn,
	}
}

// LegacyFunctionVisitor 包装旧函数的访问者
type LegacyFunctionVisitor struct {
	callback func(*LegacyDatasource)
}

func (v *LegacyFunctionVisitor) VisitDatasource(ds Datasource) {
	// 转换新数据源为旧数据源
	adapter := NewLegacyAdapter(ds)
	legacyDs := adapter.ToLegacyDatasource()

	// 调用旧函数
	v.callback(legacyDs)
}

func (v *LegacyFunctionVisitor) VisitMySQL(ds MySQLDatasource) {
	v.VisitDatasource(ds)
}

func (v *LegacyFunctionVisitor) VisitRedis(ds RedisDatasource) {
	v.VisitDatasource(ds)
}