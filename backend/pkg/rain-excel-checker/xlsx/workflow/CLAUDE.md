# workflow 包 — 统一配表检查工作流入口

提供统一的配表检查工作流入口，封装 engine 的底层函数，为 CLI/Wails/MCP 三端提供一致的检查接口。

## 文件结构

```
workflow/
└── workflow.go    # 工作流实现（检查配置、统计信息、执行入口）
```

## 核心组件

### workflow.go — 工作流核心实现

| 导出类型 | 说明 |
|----------|------|
| `CheckMode` | 检查模式类型（CheckModeFull / CheckModeIncremental） |
| `CheckWorkflowConfig` | 检查工作流配置（ExcelPath、CasePath、Mode、Rules、ChangedFiles、Git 参数等） |
| `WorkflowStats` | 工作流统计信息（TotalRules、FilteredRules、ChangedFileCount 等） |
| `SheetParseError` | Sheet 解析错误（FileName、SheetName、Error） |
| `WorkflowResult` | 工作流执行结果（ColResults、TableResults、ParseErrors、Stats） |

| 导出函数 | 说明 |
|----------|------|
| `RunCheckWorkflow()` | 统一检查入口，根据 Mode 和参数自动选择执行路径 |

## 检查模式

| 模式 | 入口函数 | 数据来源 | 适用场景 |
|------|----------|----------|----------|
| 全量 | `runFullCheck` | 本地磁盘 | 本地全量验证 |
| 增量（Git 历史） | `runIncrementalCheck` → `CheckWithGitHistory` | git show | merge 场景 |
| 增量（变更文件） | `runIncrementalCheck` → `CheckWithFilter` | 本地磁盘 | 普通 commit |

## 包依赖

### 依赖
- `engine` — 核心检查引擎（CheckWithFilter、CheckWithGitHistory）
- `excelio` — Excel 文件读取和过滤
- `json_rule` — 规则类型定义

### 被依赖
- CLI（main.go）— 命令行入口
- Wails 前端 — 通过服务层调用
- MCP 服务器 — 通过服务层调用

## 关键行为

- **规则预加载**：`CheckWorkflowConfig.Rules` 非空时跳过磁盘加载，支持前端传入内存中的规则
- **自动路径选择**：根据参数自动选择 Git 历史检查或变更文件过滤检查
- **错误收集**：统一收集 Sheet 解析错误（`SheetParseError`），供前端显示
