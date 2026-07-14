# Merge 场景飞书消息汇聚（已实现）

## 设计目标

Merge 场景下将多个 commit 的检查结果汇聚为**一条**飞书消息（原行为：每个 commit 各发一条），避免消息泛滥。普通 commit 消息格式完全不变。

## 实现状态

✅ **已完成** — 所有代码已合并到主分支。

## 关键代码位置

| 组件 | 文件 | 说明 |
|------|------|------|
| 数据结构 | `feishu-lib/notification/event.go` | `CommitSection` 结构体 + `CheckResultEvent` 扩展 |
| 格式化 | `feishu-lib/notification/formatter.go` | `FormatMergeContent` / `FormatMergeConsoleOutput` |
| Handler | `feishu-lib/notification/handlers/*.go` | 三个 handler 的 merge 分支 |
| 分发逻辑 | `rain-excel-checker/main.go` | `dispatchMergeResults` 重写为"1 次 Dispatch" |

## 消息结构

1. Merge 概览（@ merge 操作者 + 分支 commit 摘要）
2. 各 commit 分段检查结果（每个段落 @ 对应提交者）
3. 检查时间

- `maxMergeDisplayCommits`（默认 5）控制最多展示详细分段的 commit 数量
- 超过限制的 commit 在概览区域简要列出
