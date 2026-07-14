# excel-test

Excel 配置表检查服务，支持全量/增量检查、预览、过滤和飞书通知。

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| CheckContext | context.go:9 | 检查上下文，统一收集检查结果 |
| ExcelCheckService | wails.go:84 | Excel 检查服务（核心业务逻辑） |
| ExcelConfig | wails.go:1462 | Excel 应用配置 |
| ExcelCheckResult | wails.go:243 | 检查结果 |
| SheetPreviewResult | wails.go:499 | Sheet 预览结果 |
| FilterCondition | wails.go:1201 | 过滤条件 |
| ExcelTestGameService | wails.go:1610 | 前端 Game 服务包装器 |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| NewCheckContext | context.go:16 | 创建检查上下文 |
| CheckAllExcelRules | wails.go:250 | 执行全量 Excel 规则检查 |
| CheckIncremental | wails.go:285 | 增量检查（基于 git diff） |
| GetAllExcels | wails.go:408 | 获取所有 Excel 文件信息 |
| PreviewExcelSheet | wails.go:529 | 预览 Excel Sheet 数据 |
| GetExcelColumnInfo | wails.go:725 | 获取列详细信息 |
| CreateExcelFile | wails.go:1099 | 创建符合项目规范的 Excel 文件 |
| FilterExcelData | wails.go:1227 | 根据条件过滤 Excel 数据 |
| QueryExcelRange | wails.go:1360 | 查询指定范围数据 |
| InitLLMTools | llm.go:17 | 初始化 LLM 工具注册 |

## 开发注意事项
- Excel 文件格式：前 4 行为表头（中文名/类型/字段名/导出标识）
- 支持两种模式：全量检查和增量检查（基于 git diff）
- CheckContext 统一收集结果，避免重复发送通知
- 飞书通知支持劫持模式，测试阶段使用 InterceptService

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/excel-test/excel-test.spec.ts`](../../../frontend/e2e/excel-test/excel-test.spec.ts) | 页面加载、规则检查、增量检查、Sheet预览、过滤查询、列信息 |

## 依赖关系
- 依赖：rain-excel-checker（检查逻辑）、common（配置服务/飞书通知）、excelize
