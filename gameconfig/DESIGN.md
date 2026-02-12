# 游戏配置管理工具设计文档

## 1. 背景和需求

### 1.1 核心问题

游戏服务端需要处理大量配置数据，这些配置通常由策划同事使用 Excel 编辑。传统的配置管理方式存在以下问题：

- **开发效率低**：每次修改配置需要手动导出 CSV
- **版本控制困难**：Excel 二进制文件无法有效 diff
- **文档不同步**：配置说明分散在多个地方，容易过时
- **结构演进困难**：配置表结构变化时难以兼容旧数据

### 1.2 核心需求

1. **双模式支持**：开发环境直接读 Excel，生产环境读 CSV
2. **公式计算**：支持读取公式计算后的值
3. **批注作为活文档**：读取 Excel 批注，自动生成 API 文档
4. **Schema 迁移**：支持表结构演进和数据迁移
5. **热重载**：配置文件变化时自动重新加载

## 2. 关键选型

### 2.1 为什么选择双模式？

| 模式 | 读取方式 | 优点 | 缺点 | 适用场景 |
|------|---------|------|------|----------|
| **Excel** | 直接读取 .xlsx | 快速迭代，策划修改后立即可见 | 无法版本控制 | 开发环境 |
| **CSV** | 读取导出的 .csv | Git 版本控制，diff 友好 | 需要导出步骤 | 生产环境 |

**决策**：支持双模式，通过 `Mode` 配置项控制，默认 `ModeAuto` 自动检测。

### 2.2 为什么选择 excelize？

| 库 | 活跃度 | 公式支持 | 批注支持 | 性能 |
|----|--------|---------|---------|------|
| excelize | ⭐⭐⭐⭐⭐ | ✅ | ✅ | ⭐⭐⭐⭐ |
| tealeg/xlsx | ⭐⭐ | ❌ | ⚠️ | ⭐⭐⭐ |
| 360EntSecGroup-Skylar/excelize | ❌ 已废弃 | ✅ | ✅ | ⭐⭐⭐ |

**决策**：使用 `excelize/v2`，活跃维护，功能完整。

### 2.3 为什么支持 Schema 迁移？

游戏配置表结构会频繁演变：
- 新增列：新功能需要新的配置字段
- 删除列：旧功能废弃
- 重命名列：字段名调整

**决策**：实现 Schema 版本管理和迁移机制，支持向后兼容。

### 2.4 批注作为活文档

传统方式：策划在 Excel 中写好批注，开发手动复制到代码注释或 API 文档，容易不同步。

**决策**：程序直接读取 Excel 批注，自动生成或更新 API 文档，确保文档始终与配置同步。

## 3. 数据格式约定

### 3.1 Excel 文件结构

```
装备表.xlsx:
├── 武器 (Sheet)
└── 防具 (Sheet)
```

### 3.2 Sheet 数据格式

| 行 | 说明 | 示例 |
|----|------|------|
| 0 | 版本行 | `__version__ \| 2` |
| 1 | 变更说明（可选） | `__changes__ \| 新增 quality 列` |
| 2 | 字段名行 | `id \| name \| attack \| defense` |
| 3 | 类型行（可选） | `int \| string \| int \| int` |
| 4+ | 数据行 | `1001 \| 铁剑 \| 10 \| 5` |

### 3.3 CSV 导出格式

```
装备表/
├── 武器.csv
└── 防具.csv
```

每个 Sheet 导出为同名 CSV 文件，放在以 Excel 文件名命名的目录下。

### 3.4 Struct Tag 格式

```go
type Equipment struct {
    ID      int    `excel:"id"`              // 基本映射
    Name    string `excel:"name,required"`    // 必填字段
    Attack  int    `excel:"attack,default:0"` // 默认值
    Defense int    `excel:"-"`               // 跳过此字段
}
```

## 4. 架构设计

### 4.1 模块组织

```
gameconfig/
├── internal/config/    # 内部实现
│   ├── excel.go       # Excel 读取
│   ├── exporter.go    # Excel 导出
│   ├── csv.go         # CSV 读取
│   ├── loader.go      # 统一加载器
│   ├── schema.go      # Schema 管理
│   ├── watcher.go     # 热重载
│   ├── mapper.go      # 反射映射
│   ├── comment.go     # 批注处理
│   └── errors.go      # 错误处理
└── pkg/config/        # 对外 API
    └── config.go
```

### 4.2 核心类型

```go
// Mode 配置加载模式
type Mode string

const (
    ModeAuto   Mode = "auto"   // 自动检测
    ModeExcel  Mode = "excel"  // 强制 Excel
    ModeCSV    Mode = "csv"    // 强制 CSV
)

// Loader 配置加载器（泛型）
type Loader[T any] struct {
    basePath  string
    sheetName string
    mode      Mode
    options   LoadOptions
}

// ConfigWithComments 带批注的配置
type ConfigWithComments[T any] struct {
    Data     []T
    Comments map[string]string  // "字段名" -> "批注内容"
}

// SchemaVersion Schema 版本
type SchemaVersion struct {
    Version    int
    Migrations []Migration
}

// Watcher 文件监听器
type Watcher[T any] struct {
    loader   *Loader[T]
    callback func([]T)
    debounce time.Duration
}
```

### 4.3 加载流程

```
用户调用 Load()
    ↓
检测模式 (Auto/Excel/CSV)
    ↓
┌─────────────┬─────────────┐
│   Excel     │    CSV      │
│  读取器     │   读取器     │
└─────────────┴─────────────┘
    ↓
Schema 迁移（如果需要）
    ↓
反射映射到结构体
    ↓
读取批注（如果需要）
    ↓
返回数据
```

## 5. 错误处理策略

### 5.1 错误信息格式

```
配置错误 [装备表.xlsx] 行5 列3 (attack):
  无法将字符串 "high" 转换为 int32 类型
```

包含以下信息：
- 文件名
- 行号
- 列名
- 具体错误

### 5.2 错误类型

```go
var (
    ErrFileNotFound     = errors.New("文件不存在")
    ErrInvalidFormat    = errors.New("格式错误")
    ErrTypeMismatch    = errors.New("类型不匹配")
    ErrRequiredField   = errors.New("缺少必填字段")
    ErrInvalidVersion  = errors.New("无效的版本号")
    ErrMigrationFailed = errors.New("迁移失败")
)
```

## 6. 测试策略

### 6.1 单元测试

每个核心模块都有对应的测试文件：
- `excel_test.go`：Excel 读取测试
- `csv_test.go`：CSV 读取测试
- `loader_test.go`：加载器测试
- `schema_test.go`：Schema 迁移测试
- `comment_test.go`：批注测试
- `watcher_test.go`：热重载测试

### 6.2 集成测试

`tests/integration_test.go`：完整流程测试，覆盖：
- Excel → CSV → 加载 → 热重载

### 6.3 基准测试

`loader_bench_test.go`：性能监控

## 7. 性能考虑

### 7.1 缓存策略

- Excel 文件对象缓存（避免重复打开）
- CSV 解析结果缓存（可选）

### 7.2 热重载防抖

文件变化后等待 200ms 再重新加载，避免频繁触发。

## 8. 未来扩展

- 配置验证：支持自定义验证规则
- 多语言支持：自动提取多语言文本
- 数据合并：支持多个配置文件的合并
- 配置加密：敏感配置加密存储
