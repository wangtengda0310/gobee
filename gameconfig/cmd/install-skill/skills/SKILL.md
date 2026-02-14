---
name: gameconfig 配置管理工具
description: 游戏配置管理工具 - 支持 Excel/CSV 双模式加载、Mock 数据、条件字段、Schema 迁移
---

# gameconfig 配置管理工具

## 项目定位

游戏配置管理子模块，提供 AI 原生配置加载能力。

**Go Module**: `github.com/wangtengda0310/gobee/gameconfig`

**核心价值**: 开发读 Excel，生产读 CSV，测试用 Mock。

---

## 快速开始

### 最简用法

```go
import "github.com/wangtengda0310/gobee/gameconfig/pkg/config"

type Equipment struct {
    ID      int    `excel:"id"`
    Name    string `excel:"name,required"`
    Attack  int    `excel:"attack,default:0"`
}

loader := config.NewLoader[Equipment](
    "config/装备表.xlsx",
    "武器",
    config.LoadOptions{Mode: config.ModeAuto},
)

items, err := loader.Load()
```

---

## 核心概念

### 加载模式

| 模式 | 何时使用 |
|------|----------|
| `ModeAuto` | 默认，自动检测（优先 CSV） |
| `ModeExcel` | 开发环境直接读 Excel |
| `ModeCSV` | 生产环境读 CSV |
| `ModeMemory` | 测试环境使用 Mock 数据 |

### Excel 格式约定

```
行 0: __version__ | 2              ← 版本行（可选）
行 1: id | name | attack | defense ← 字段名行
行 2: int | string | int | int     ← 类型行（可选）
行 3+: 1001 | 铁剑 | 10 | 5        ← 数据行
```

### Struct Tag

| Tag | 说明 |
|-----|------|
| `excel:"id"` | 基本映射 |
| `excel:"name,required"` | 必填字段 |
| `excel:"atk,default:0"` | 默认值 |
| `excel:"-"` | 跳过此字段 |
| `excel:"field,when:type=1"` | 条件字段（仅 type=1 时加载） |

---

## 使用场景

### 场景 1: 基础加载

```go
// 自动模式：开发读 Excel，生产读 CSV
loader := config.NewLoader[Equipment](
    "config/装备表.xlsx",
    "武器",
    config.LoadOptions{Mode: config.ModeAuto},
)
items, err := loader.Load()
```

### 场景 2: 测试用 Mock 数据

```go
// 方式 1: 直接提供 MockData
loader := config.NewLoader[Equipment]("", "武器", config.LoadOptions{
    Mode: config.ModeMemory,
    MockData: [][]string{
        {"id", "name", "attack", "defense"},
        {"1001", "铁剑", "10", "5"},
        {"1002", "钢剑", "20", "8"},
    },
})

// 方式 2: 使用 SetMockData（适合动态测试）
loader := config.NewLoader[Equipment]("", "武器", config.LoadOptions{
    Mode: config.ModeMemory,
})
loader.SetMockData(mockData)
items, err := loader.Load()
```

### 场景 3: 条件字段（Phase 3-4 功能）

```go
type Equipment struct {
    ID      int    `excel:"id"`
    Type    int    `excel:"type"`           // 0:普通 1:武器 2:盔甲
    Attack  int    `excel:"attack,when:type=1"`  // 仅武器时加载
    Defense int    `excel:"defense,when:type=2"` // 仅盔甲时加载
}
```

### 场景 4: Schema 迁移

```go
schema := config.NewSchemaManager()
schema.Register("装备表.武器", &config.SchemaVersion{
    Version: 2,
    Migrations: []config.Migration{
        {
            FromVersion: 1,
            Transform: func(old, new interface{}) error {
                // 迁移逻辑
                return nil
            },
        },
    },
})
```

### 场景 5: 热重载

```go
watcher := config.NewWatcher(loader)
watcher.OnChange(func(data []Equipment) {
    log.Printf("配置已更新，共 %d 条", len(data))
})
go watcher.Watch(context.Background())
```

---

## 项目结构

```
gameconfig/
├── internal/config/
│   ├── loader.go      # 统一加载器
│   ├── excel.go       # Excel 读取
│   ├── csv.go         # CSV 读取
│   ├── mapper.go      # 结构体反射映射
│   ├── schema.go      # Schema 版本管理
│   ├── watcher.go     # 热重载
│   └── *_test.go      # 单元测试
├── pkg/config/
│   └── config.go      # 对外 API
├── cmd/xlsx2csv/      # 导出工具
└── tests/             # 集成测试
```

---

## 最佳实践

### DO 推荐

- ✅ 生产环境使用 `ModeCSV`（Git diff 友好）
- ✅ 测试环境使用 `ModeMemory`（无需文件）
- ✅ 使用 `when` 标签实现条件字段
- ✅ 为必填字段添加 `required` 标签
- ✅ 为数值字段添加合理的 `default` 值

### DON'T 避免

- ❌ 不要在生产环境直接读取 Excel
- ❌ 不要忘记处理 `Load()` 返回的 error
- ❌ 不要在多个 goroutine 中共享同一个 Loader（并发安全已在 v1.1.0+ 支持）
- ❌ 不要让 Mock 数据与实际数据结构不一致

---

## 常见问题

### Q: 如何导出 Excel 为 CSV？

```bash
# 使用内置工具
cd gameconfig/cmd/xlsx2csv
go run main.go -input ../../testdata/excel/装备表.xlsx -output ../../testdata/csv/
```

### Q: 条件字段不生效？

检查：
1. `when` 条件字段是否在条件字段之前定义
2. 条件值是否正确（如 `when:type=1`）
3. 参考文档: `gameconfig/internal/config/conditional_test.go:100`

### Q: 如何验证配置数据？

```go
// 使用 Validate 接口
type Equipment struct {
    // ... 字段定义
}

func (e *Equipment) Validate() error {
    if e.Attack < 0 || e.Attack > 10000 {
        return fmt.Errorf("attack 超出范围 [0,10000]: %d", e.Attack)
    }
    return nil
}
```

### Q: 测试文件在哪？

- 单元测试: `internal/config/*_test.go`
- 集成测试: `tests/integration_test.go`
- 测试数据: `testdata/excel/` 和 `testdata/csv/`

---

## 运行测试

```bash
# 单元测试
go test ./internal/config/... -v

# 集成测试
go test ./tests/... -v

# 代码覆盖率
go test -coverprofile=coverage.out ./...
```

---

## 依赖

```go
require (
    github.com/xuri/excelize/v2 v2.10.0    // Excel 读写
    github.com/fsnotify/fsnotify v1.9.0   // 文件监听
)
```

---

## 相关文档

- 项目 README: `gameconfig/README.md`
- 设计文档: `gameconfig/DESIGN.md`
- AI 原生规划: `gameconfig/AI_NATIVE.md`
- 变更日志: `gameconfig/CHANGELOG.md`
- **AI 使用指导**: `abilities/AI指导.md` ← AI 阅读此文件以了解如何协助用户
