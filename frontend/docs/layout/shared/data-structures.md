# 共享数据结构

> 配表测试多个页面共用的后端返回数据结构

## TableCheckResult

表级检查结果，区分通知和错误：

| 字段 | 类型 | 说明 |
|------|------|------|
| ok | boolean | true=通知, false=错误 |
| sheetName | string | Sheet 名称 |
| ruleType | string | 规则类型 |
| displayName | string | 显示名称 |
| reason | string | 原因描述 |
| errCells | CellError[] | 错误单元格 |

## ColCheckResult

列级检查结果：

| 字段 | 类型 | 说明 |
|------|------|------|
| Ok | boolean | 是否通过 |
| ColIndex | number | 列索引 |
| ErrCells | CellError[] | 错误单元格 |
| SheetName | string | Sheet 名称 |

---
**Verification Date**: 2026-05-11
**Status**: 从 docs/layout/CLAUDE.md 迁移而来
