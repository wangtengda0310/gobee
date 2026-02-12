package config

// Migration 迁移规则
type Migration struct {
	FromVersion int
	ToVersion   int
	Migrate     func(map[string]string) map[string]string
	Description string
}

// SchemaVersion Schema 版本
type SchemaVersion struct {
	Version    int
	Migrations []Migration
}

// SchemaManager Schema 管理器
type SchemaManager struct {
	versions map[string]*SchemaVersion
}

// NewSchemaManager 创建 Schema 管理器
func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		versions: make(map[string]*SchemaVersion),
	}
}

// Register 注册 Schema
func (sm *SchemaManager) Register(table string, version *SchemaVersion) {
	sm.versions[table] = version
}

// Migrate 执行数据迁移
func (sm *SchemaManager) Migrate(table string, fromVersion int, row map[string]string) (map[string]string, error) {
	schema, ok := sm.versions[table]
	if !ok {
		return row, nil // 没有注册 Schema，不进行迁移
	}

	// 找到目标版本
	currentVersion := fromVersion
	for _, migration := range schema.Migrations {
		if migration.FromVersion == currentVersion {
			// 执行迁移
			newRow := migration.Migrate(row)
			currentVersion = migration.ToVersion
			row = newRow
		}
	}

	return row, nil
}

// GetVersion 获取表的最新版本
func (sm *SchemaManager) GetVersion(table string) int {
	schema, ok := sm.versions[table]
	if !ok {
		return 0
	}
	return schema.Version
}
