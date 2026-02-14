# 游戏配置管理工具

为游戏服务端提供双模式配置管理：开发环境直接读取 Excel，生产环境读取 CSV。

**Go Module**: `github.com/wangtengda0310/gobee/gameconfig`

---

## 特性

- **双模式支持**：开发用 Excel（快速迭代），生产用 CSV（Git 友好）
- **Mock 数据**：测试环境无需文件，直接使用内存数据
- **条件字段**：根据条件动态加载字段（如 `when:type=1`）
- **Schema 迁移**：支持表结构演进和数据迁移
- **热重载**：配置文件变化时自动重新加载
- **类型推断**：自动推断 Go 类型，支持默认值和必填验证
- **批注支持**：读取 Excel 批注作为字段说明

---

## 安装

```bash
go get github.com/wangtengda0310/gobee/gameconfig
```

### Claude Code Skill（推荐）

安装 gameconfig 的 Claude Code Skill，让 AI 帮助你：

- 🔍 审查配置表，发现潜在问题
- 🧪 自动生成测试数据
- 📝 生成结构体定义
- 🔄 分析 Schema 变更

```bash
# 安装 skill 到全局
go install github.com/wangtengda0310/gobee/gameconfig/cmd/install-skill@latest
gameconfig-install-skill
```

安装后，在任何项目中直接与 AI 对话即可使用：
- "审查一下装备表配置"
- "生成测试数据"
- "创建配置结构体"

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/wangtengda0310/gobee/gameconfig/pkg/config"
)

type Equipment struct {
    ID      int    `excel:"id"`
    Name    string `excel:"name,required"`
    Attack  int    `excel:"attack,default:0"`
    Defense int    `excel:"defense,default:0"`
}

func main() {
    // 自动模式：开发环境读 Excel，生产环境读 CSV
    loader := config.NewLoader[Equipment](
        "config/装备表.xlsx",
        "武器",
        config.LoadOptions{
            Mode:      config.ModeAuto,
            HeaderRow: 0,
        },
    )

    equipments, err := loader.Load()
    if err != nil {
        panic(err)
    }

    for _, eq := range equipments {
        fmt.Printf("%s: 攻击=%d, 防御=%d\n", eq.Name, eq.Attack, eq.Defense)
    }
}
```

### 使用 Mock 数据（测试环境）

当没有策划提供 Excel 文件时，可以使用 Mock 数据：

```go
// 方式 1：直接提供 MockData
loader := config.NewLoader[Equipment]("", "武器", config.LoadOptions{
    Mode: config.ModeMemory,
    MockData: [][]string{
        {"id", "name", "attack", "defense"},
        {"1001", "铁剑", "10", "5"},
        {"1002", "钢剑", "25", "10"},
    },
})

items, err := loader.Load()

// 方式 2：使用 SetMockData 方法（适合动态测试）
loader := config.NewLoader[Equipment]("", "武器", config.LoadOptions{
    Mode: config.ModeMemory,
})
loader.SetMockData(mockData)
items, err := loader.Load()
```

### 条件字段

根据条件动态加载字段（仅当条件满足时才解析该字段）：

```go
type Equipment struct {
    ID      int    `excel:"id"`
    Type    int    `excel:"type"`                      // 0:普通 1:武器 2:盔甲
    Attack  int    `excel:"attack,when:type=1"`        // 仅武器时加载
    Defense int    `excel:"defense,when:type=2"`       // 仅盔甲时加载
}
```

**注意**：条件字段必须在依赖字段之后定义（如 `attack` 必须在 `type` 之后）。

### Schema 迁移

处理配置表结构演进：

```go
schema := config.NewSchemaManager()
schema.Register("装备表.武器", &config.SchemaVersion{
    Version: 2,
    Migrations: []config.Migration{
        {
            FromVersion: 1,
            ToVersion: 2,
            Migrate: func(row map[string]string) map[string]string {
                row["attack_power"] = row["attack"]  // 重命名
                delete(row, "attack")
                delete(row, "old_field")              // 删除
                row["quality"] = "common"             // 新增（默认值）
                return row
            },
            Description: "重命名 attack 为 attack_power，删除 old_field，新增 quality",
        },
    },
})

loader := config.NewLoader[Equipment]("config/装备表.xlsx", "武器")
loader.SetSchemaManager(schema)
equipments, err := loader.Load()
```

### 热重载

监听配置文件变化并自动重新加载：

```go
watcher := config.NewWatcher(loader)
watcher.OnChange(func(data []Equipment) {
    log.Printf("配置已更新，共 %d 条", len(data))
})

ctx := context.Background()
go watcher.Watch(ctx)

// 主程序继续运行...
```

---

## Excel 格式约定

### Sheet 数据格式

| 行 | 说明 | 示例 |
|----|------|------|
| 0 | 版本行（可选） | `__version__ \| 2` |
| 1 | 变更说明（可选） | `__changes__ \| 新增 quality 列` |
| 2 | 字段名行 | `id \| name \| attack \| defense` |
| 3 | 类型行（可选） | `int \| string \| int \| int` |
| 4+ | 数据行 | `1001 \| 铁剑 \| 10 \| 5` |

### Struct Tag 格式

```go
type Equipment struct {
    ID      int    `excel:"id"`              // 基本映射
    Name    string `excel:"name,required"`    // 必填字段
    Attack  int    `excel:"attack,default:0"` // 默认值
    Defense int    `excel:"-"`               // 跳过此字段
}
```

| Tag | 说明 |
|-----|------|
| `excel:"field"` | 基本映射 |
| `excel:"field,required"` | 必填字段（缺失时返回错误） |
| `excel:"field,default:value"` | 默认值（缺失或空时使用） |
| `excel:"field,when:condition"` | 条件字段（条件满足时才加载） |
| `excel:"-"` | 跳过此字段 |

---

## 配置模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| `ModeAuto` | 自动检测（优先 CSV） | 默认模式 |
| `ModeExcel` | 强制读取 Excel | 开发环境 |
| `ModeCSV` | 强制读取 CSV | 生产环境 |
| `ModeMemory` | 从内存数据加载 | 测试环境（Mock 数据） |

---

## 并发安全

gameconfig 设计了并发安全机制：

### ✅ 支持的场景

- **多 goroutine 同时读取**：每个 Loader 实例独立，可并发加载
- **同一个 Loader 并发读取**：使用 RWMutex 保护，安全无虞
- **热重载 + 读取**：Watcher 独立运行，不影响 Loader 读取

### ⚠️ 注意事项

- `SetMockData()` 方法有锁保护，但多 goroutine 同时设置时最终值不确定
- 建议在测试环境中使用，生产环境中慎用

### 推荐用法

```go
// ✅ 推荐：同一个 Loader 在不同 goroutine 中读取
loader := config.NewLoader[Equipment](path, "sheet", opts)
for i := 0; i < 10; i++ {
    go func() {
        items, _ := loader.Load()
        // 处理数据
    }()
}

// ⚠️ 注意：多 goroutine 写入 SetMockData 时最终值不确定
```

---

## Excel 导出工具

将 Excel 的每个 Sheet 导出为 CSV 文件：

```bash
go run github.com/wangtengda0310/gobee/gameconfig/cmd/xlsx2csv \
    -source config \
    -target config/csv
```

导出后的目录结构：

```
config/
├── 装备表.xlsx
└── csv/
    ├── 武器.csv
    └── 防具.csv
```

---

## 错误处理

```go
equipments, err := loader.Load()
if err != nil {
    // 友好的错误信息，包含源文件位置
    // 示例: 配置错误 [装备表.xlsx] 行5 列3 (attack):
    //       无法将字符串 "high" 转换为 int32 类型
    panic(err)
}
```

---

## 常见问题

### Q: 如何选择配置模式？

- **开发环境**：使用 `ModeExcel` 或 `ModeAuto`，直接读取 Excel 快速迭代
- **生产环境**：使用 `ModeCSV`，CSV 文件 Git diff 友好
- **测试环境**：使用 `ModeMemory` 配合 Mock 数据，无需文件

### Q: 条件字段不生效？

检查：
1. `when` 条件字段是否在条件字段之前定义（如 `type` 必须在 `attack` 之前）
2. 条件值是否正确（如 `when:type=1`）
3. 参考文档：`internal/config/conditional_test.go`

### Q: 如何验证配置数据？

实现 `Validate` 接口：

```go
func (e *Equipment) Validate() error {
    if e.Attack < 0 || e.Attack > 10000 {
        return fmt.Errorf("attack 超出范围 [0,10000]: %d", e.Attack)
    }
    return nil
}
```

### Q: CSV 文件编码问题？

确保 CSV 文件使用 UTF-8 编码，Excel 导出时会自动转换。

---

## 许可证

MIT License
