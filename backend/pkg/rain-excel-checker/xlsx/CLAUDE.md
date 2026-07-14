# xlsx 目录文档

Excel 配表检查核心模块。

## 子目录

| 目录 | 说明 | 文档 |
|------|------|------|
| [excelio/](./excelio/) | Excel 读取、解析、类型定义 | [CLAUDE.md](./excelio/CLAUDE.md) |
| [helpers/](./helpers/) | 通用辅助工具（列操作、参数解析、正则、拼音、路径、武将） | [CLAUDE.md](./helpers/CLAUDE.md) |
| [diff/](./diff/) | Excel 差异检测（Git 历史对比） | [CLAUDE.md](./diff/CLAUDE.md) |
| [ruleconfig/](./ruleconfig/) | 检查规则配置 I/O（JSON 读写） | [CLAUDE.md](./ruleconfig/CLAUDE.md) |
| [engine/](./engine/) | 检查器注册工厂 + 执行入口 | [CLAUDE.md](./engine/CLAUDE.md) |
| [coded_rules/](./coded_rules/) | 校验规则实现 | [CLAUDE.md](./coded_rules/CLAUDE.md) |
| [json_rule/](./json_rule/) | 规则定义和类型 | [CLAUDE.md](./json_rule/CLAUDE.md) |
| [workflow/](./workflow/) | 统一检查工作流入口（CLI/Wails/MCP 三端） | [CLAUDE.md](./workflow/CLAUDE.md) |
| [tests/](./tests/) | 单元测试 | [CLAUDE.md](./tests/CLAUDE.md) |

## 架构概览

```
                    ┌─────────────────┐
                    │   main.go       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │    engine       │
                    │ (检查入口/工厂)  │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐   ┌──────▼──────┐   ┌──────▼────────┐
│  coded_rules   │   │   helpers   │   │    excelio    │
│                │   │             │   │               │
│ cross_table/   │   │ 列操作工具   │   │ Excel 读取    │
│ table/         │   │ 参数解析     │   │ 类型定义      │
│ general/       │   │ 正则/拼音   │   │ 列辅助工具     │
│                │   │ 路径/武将   │   │               │
└────────────────┘   └─────────────┘   └───────────────┘
        │                    │
        │            ┌───────▼────────┐
        │            │      diff      │
        │            │ (Git 差异检测)  │
        │            └────────────────┘
        │
        ▼
┌─────────────────────────┐
│      json_rule          │
│    (规则定义/类型)       │
└─────────────────────────┘
```

## 包职责

| 包 | 职责 | 对外暴露 |
|----|------|----------|
| `excelio` | Excel 文件读取、数据结构定义、格式解析 | 是（rain-qa-func、rain-resources-checker 使用） |
| `helpers` | 通用工具函数（列操作、参数解析、正则、拼音、路径、武将辅助） | 是（coded_rules 使用） |
| `diff` | Excel 差异检测（Git 历史对比）、上下文适配器 | 是（coded_rules 使用） |
| `ruleconfig` | 检查规则 JSON 配置的保存和加载 | 是（main.go 使用） |
| `engine` | 检查器注册工厂 + 检查执行入口（CheckAll/CheckWithFilter/CheckWithGitHistory） | 是（main.go、workflow、rain-qa-func 使用） |

## 快速开始

### 执行检查

```go
// 读取 Excel 并执行规则检查
res, err := engine.CheckAll(excelPath, casePath)
```

### 添加新规则

详见 [coded_rules/CLAUDE.md](./coded_rules/CLAUDE.md)
