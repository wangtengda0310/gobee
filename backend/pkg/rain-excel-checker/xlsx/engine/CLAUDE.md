# engine 包 — Excel 检查器管理核心功能

Excel 配表检查的核心执行引擎，负责管理所有检查器的注册、调度和执行。提供主要检查入口函数、检查器管理接口以及规则过滤功能。

## 文件结构

```
engine/
├── interfaces.go         # 检查器接口定义（Checker、TableChecker）
├── column_registry.go    # 列级检查器注册和管理
├── table_registry.go     # 表级检查器注册和管理
├── cache.go              # 检查结果缓存管理
├── adapter.go            # 列规则适配器
├── check.go              # Excel 读取和列级/表级检查执行
├── executor.go           # 主要检查入口函数
├── filter.go             # 规则过滤和表级规则补充
├── cache_test.go         # 缓存测试
├── col_rule_adapter_test.go  # 适配器测试
├── filter_test.go        # 过滤器测试
└── integration_test.go   # 集成测试
```

## 核心组件

### interfaces.go — 检查器接口定义

| 导出类型 | 说明 |
|----------|------|
| `Checker` | 列级检查器接口，所有列级检查器必须实现 |
| `TableChecker` | 表级检查器接口，所有表级检查器必须实现 |

### column_registry.go — 列级检查器管理

| 导出类型/变量/函数 | 说明 |
|-------------------|------|
| `Manager` | 全局列级检查器管理器实例 |
| `CheckerManager` | 列级检查器管理器，负责注册和获取检查器 |
| `NewCheckerManager()` | 创建新的管理器 |
| `Reg(checkRule, checker)` | 注册列级检查器 |
| `GetChecker(checkRule)` | 获取指定类型的检查器 |

### table_registry.go — 表级检查器管理

| 导出类型/变量/函数 | 说明 |
|-------------------|------|
| `TableManager` | 全局表级检查器管理器实例 |
| `TableCheckerManager` | 表级检查器管理器 |
| `NewTableCheckerManager()` | 创建新的管理器 |
| `Reg(rule, checker)` | 注册表级检查器 |
| `GetChecker(rule)` | 获取指定类型的表级检查器 |
| `GetAllMetas()` | 获取所有表级规则元数据 |

### executor.go — 检查入口函数

| 导出类型/函数 | 说明 |
|-------------|------|
| `CheckOption` | CheckWithGitHistory 的可选参数（函数类型） |
| `CheckStats` | 检查统计信息结构体 |
| `CheckAll()` | 全量检查入口 |
| `CheckWithFilter()` | 增量检查入口（变更文件过滤） |
| `CheckWithGitHistory()` | 增量检查入口（Git 历史版本） |
| `WithFallbackSheetMap()` | 设置预加载的 fallback sheetMap |
| `WithPreloadedRules()` | 使用预加载的规则列表 |
| `WithSheetMapOutput()` | 设置 sheetMap 导出目标 |

### check.go — Excel 检查执行

| 导出函数 | 说明 |
|----------|------|
| `ReadAndCheckXlsx()` | 读取 Excel 文件并执行列级检查 |
| `CheckTableRules()` | 检查表级规则 |
| `CheckSingleColumn()` | 检查单个列（前端"执行此字段检查"功能） |
| `CheckSheetCols()` | 检查表的所有列 |

### filter.go — 规则过滤和管理

| 导出函数 | 说明 |
|----------|------|
| `FilterRulesByChangedFiles()` | 根据变更文件过滤规则列表 |
| `SupplementDefaultParams()` | 为已有 JSON 配置的 TableRule 补充默认参数 |
| `filterSheetMapByRules()` | 根据规则列表过滤 sheetMap |
| `supplementDefaultTableRules()` | 为有默认表级规则但没有 JSON 文件的表创建 SheetRule |

### cache.go — 检查结果缓存

| 导出函数 | 说明 |
|----------|------|
| `StoreCheckResults()` | 缓存配表检查结果 |
| `GetCachedCheckResults()` | 读取缓存的检查结果 |

## 关键入口函数

| 功能 | 入口 |
|------|------|
| 全量检查 | `executor.go:CheckAll()` |
| 增量检查（本地文件） | `executor.go:CheckWithFilter()` |
| 增量检查（Git 历史） | `executor.go:CheckWithGitHistory()` |
| 单列检查 | `check.go:CheckSingleColumn()` |
| 规则过滤 | `filter.go:FilterRulesByChangedFiles()` |
| 检查器注册 | `column_registry.go:Reg()` / `table_registry.go:Reg()` |

## 包依赖

### 依赖
- `json_rule` — 规则类型定义和检查结果数据结构
- `excelio` — Excel 文件读取和解析
- `helpers` — 通用辅助工具
- `diff` — Git 差异检测和快照管理
- `ruleconfig` — 检查规则配置的保存和加载
- `gitutil` — Git 操作工具

### 被依赖
- `workflow` — 检查工作流编排
- `main.go` — 项目入口调用检查函数
