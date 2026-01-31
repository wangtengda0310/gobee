# 架构重构实施计划

## 重构目标

将当前基于全局变量的访问者模式重构为基于 context.Context 的依赖传递模式。

---

## 重构步骤

### 步骤1：创建 service 层接口和实现

**文件**: `pkg/dump/service/datasource.go`

```go
package service

import (
    "context"
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
```

**文件**: `pkg/dump/service/mysql.go`

```go
package service

import (
    "context"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

// MySQLManager MySQL 数据源管理器
type MySQLManager struct {
    db     *sql.DB
    config Config
}

// NewMySQLManager 创建 MySQL 管理器
func NewMySQLManager(ctx context.Context, config Config) (*MySQLManager, error) {
    db, err := sql.Open("mysql", config.DSN())
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
```

---

### 步骤2：创建 context 管理包

**文件**: `cmd/context/context.go`

```go
package context

import (
    "context"

    "github.com/wangtengda0310/gobee/lvan/pkg/dump/service"
)

type contextKey string

const (
    ManagerKey contextKey = "datasource-manager"
)

// SetManager 设置数据源管理器到 context
func SetManager(ctx context.Context, mgr service.Manager) context.Context {
    return context.WithValue(ctx, ManagerKey, mgr)
}

// GetManager 从 context 获取数据源管理器
func GetManager(ctx context.Context) service.Manager {
    if mgr, ok := ctx.Value(ManagerKey).(service.Manager); ok {
        return mgr
    }
    return nil
}
```

---

### 步骤3：重构 mysql 命令

**修改**: `cmd/dumper/cmd/mysql.go`

**Before**:
```go
internal.Accept(func(visitor internal.Visitor, where string, args ...string) {
    conn := dump.ConnC(c)
    conn(func(db dump.Datasource) {
        visitor(db)
    })
})
```

**After**:
```go
// 创建 manager
mgr, err := service.NewMySQLManager(cmd.Context(), cfg)
if err != nil {
    return err
}

// 存储到 context
ctx := cmdcontext.SetManager(cmd.Context(), mgr)
cmd.SetContext(ctx)
```

---

### 步骤4：重构 dump 命令

**修改**: `cmd/dumper/cmd/dump.go`

**Before**:
```go
internal.VisitExport(func(db dump.Datasource) {
    columns := dump.Columns(db.DB, db.Database, db.Table)
    // ...
})
```

**After**:
```go
mgr := cmdcontext.GetManager(cmd.Context())
if mgr == nil {
    log.Panic("数据源未初始化")
}

db := mgr.GetDB()
cfg := mgr.GetConfig()

columns := dump.Columns(db, cfg.Database, cfg.Table)
// ...
```

---

### 步骤5：重构 import 命令

**修改**: `cmd/dumper/cmd/import.go`

类似 dump 命令的重构方式。

---

## 验证测试

### 测试用例

1. **Dump 功能测试**
```bash
go run lvan/cmd/dumper/main.go mysql dump \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user 1
```

2. **Import 功能测试**
```bash
go run lvan/cmd/dumper/main.go mysql import \
  -h 127.0.0.1 -P 3307 -u root \
  -d lvan_dumper_test -t user lvan_dumper_test.user.zip
```

3. **数据一致性验证**
- 检查导出的 ZIP 文件内容
- 验证 BLOB 数据完整性
- 验证 TIMESTAMP 格式正确

---

## 回滚计划

如果重构导致问题，可以通过 git 回滚：

```bash
git checkout HEAD~1 -- lvan/cmd/dumper/cmd/
git checkout HEAD~1 -- lvan/pkg/dump/
```

---

## 时间估算

| 步骤 | 预计时间 |
|------|---------|
| 创建 service 层 | 30 分钟 |
| 创建 context 包 | 15 分钟 |
| 重构 mysql 命令 | 20 分钟 |
| 重构 dump 命令 | 20 分钟 |
| 重构 import 命令 | 20 分钟 |
| 测试验证 | 30 分钟 |
| **总计** | **~2.5 小时** |
