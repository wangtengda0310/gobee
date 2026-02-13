# 游戏配置管理工具

为游戏服务端提供双模式配置管理：开发环境直接读取 Excel，生产环境读取 CSV。

## 特性

- **双模式支持**：开发用 Excel（快速迭代），生产用 CSV（Git 友好）
- **公式计算**：支持读取公式计算后的值
- **批注作为活文档**：读取 Excel 批注，自动生成 API 文档
- **Schema 迁移**：支持表结构演进和数据迁移
- **热重载**：配置文件变化时自动重新加载
- **类型推断**：自动推断 Go 类型

## 安装

```bash
go get github.com/wangtengda0310/gobee/gameconfig
```

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
    Quality string `excel:"quality,default:common"`
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

### 读取批注

```go
result, err := config.LoadWithComments[Equipment](
    "config/装备表.xlsx",
    "武器",
)

if err != nil {
    panic(err)
}

// 访问批注
comment := result.Comments["attack"]
fmt.Println("attack 字段说明:", comment)
// 输出: attack 字段说明: 物理攻击力，范围 0-999
```

### Schema 迁移

```go
schema := config.NewSchemaManager()
schema.Register("装备表.武器", &config.SchemaVersion{
    Version: 2,
    Migrations: []config.Migration{
        {
            FromVersion: 1,
            ToVersion: 2,
            Migrate: func(row map[string]string) map[string]string {
                // 版本 1 -> 2 的迁移
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

```go
watcher := config.NewWatcher(loader)
watcher.OnChange(func(data []Equipment) {
    log.Printf("配置已更新，共 %d 条", len(data))
})

ctx := context.Background()
go watcher.Watch(ctx)

// 主程序继续运行...
```

## Excel 导出工具

将 Excel 的每个 Sheet 导出为 CSV 文件：

```bash
go run github.com/wangtengda0310/gobee/gameconfig/cmd/xlsx2csv \
    -source config \
    -target config/csv
```

## Excel 格式约定

### Sheet 数据格式

| 行 | 说明 | 示例 |
|----|------|------|
| 0 | 版本行 | `__version__ \| 2` |
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

### CSV 导出格式

```
装备表/
├── 武器.csv
└── 防具.csv
```

## 配置模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| `ModeAuto` | 自动检测 | 默认模式 |
| `ModeExcel` | 强制读取 Excel | 开发环境 |
| `ModeCSV` | 强制读取 CSV | 生产环境 |
| `ModeMemory` | 从内存数据加载 | 测试环境（使用 Mock 数据） |

### 并发安全特性

gameconfig 设计了并发安全机制，支持以下并发场景：

#### ✅ 支持的场景
- **多 goroutine 同时读取**：每个 Loader 实例独立，可并发加载
- **热重载 + 读取**：Watcher 独立运行，不影响 Loader 读取
- **StructMapper 并发使用**：反射操作后缓存不可变，天然线程安全

#### ⚠️ 注意：SetMockData 并发安全
- `SetMockData()` 方法使用 `sync.RWMutex` 保护数据访问
- 多个 goroutine 同时设置数据时，最终值不确定，但不保证数据一致性
- 建议在测试环境中使用，生产环境中慎用

#### 推荐用法
```go
// ✅ 推荐：每 goroutine 使用独立 Loader
for i := 0; i < 10; i++ {
    loader := config.NewLoader[Equipment](path, "sheet", opts)
    go func() {
        items, _ := loader.Load()
        // 处理数据
    }()
}

// ✅ 推荐：同一个 Loader 在不同 goroutine 中读取
loader := config.NewLoader[Equipment](path, "sheet", opts)
for i := 0; i < 10; i++ {
    go func() {
        items, _ := loader.Load()
        // 处理数据
    }()
}

// ⚠️ 不推荐：多 goroutine 写入同一 Loader
// SetMockData 没有原子操作保证
for i := 0; i < 10; i++ {
    go loader.SetMockData(data)  // 可能产生竞争
}
```

### 测试

运行并发测试（带竞态检测）：
```bash
# 运行并发测试
go test ./internal/config/... -run "TestConcurrent" -race -v

# 运行所有测试
go test ./... -v
```

### 使用 Mock 数据（测试环境）

当没有策划提供 Excel 文件时，可以使用 Mock 数据进行开发和测试：

```go
// 方式 1：直接在 LoadOptions 中提供 MockData
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

// 设置 mock 数据
loader.SetMockData([][]string{
    {"id", "name", "attack", "defense"},
    {"1001", "铁剑", "10", "5"},
})

items, err := loader.Load()
```

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

## 运行测试

```bash
# 单元测试
go test ./internal/config/... -v

# 集成测试
go test ./tests/... -v

# 基准测试
go test ./internal/config/... -bench=. -benchmem

# 代码覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 设计文档

详细设计文档请参阅 [DESIGN.md](DESIGN.md)。

## 许可证

MIT License
