---
name: multica-cli
description: |
  Multica CLI 流水线专用指南 — 在 Docker 流水线环境中使用 multica 命令行工具创建 issue。

  **必须使用此技能的场景**：
  - 用户要求使用 multica CLI 创建 issue
  - 用户提到 "multica"、"issue" 等关键词且涉及流水线操作

  **适用领域**：CI/CD 流水线、自动化 issue 创建
version: 1.1.0
tags: [multica, cli, issue-tracking, ci-cd]
---

# Multica CLI 流水线使用指南

> Multica 是一个任务协作平台 — 人类与 AI agent 在同一工作空间中协作。

## 环境说明

- 镜像已预置 `/home/analyzer/.multica/config.json`，**无需登录**
- `multica issue create` 是纯 API 调用，**无需启动 daemon**

---

## Issue 创建

### 基本用法

```bash
multica issue create --title "标题" --description "描述内容"
```

### 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `--title` | Issue 标题（必填） | `--title "配表提交影响分析"` |
| `--description` | Issue 描述 | `--description "详细描述..."` |
| `--description-stdin` | 从 stdin 读取描述（避免命令行长度限制） | `echo "描述" \| multica issue create --title "..." --description-stdin` |
| `--project` | 指定项目 ID（固定为 `rain-qa-func`） | `--project rain-qa-func` |
| `--status` | 初始状态 | `--status open` |

### 示例

```bash
# 基础创建（固定 project 为 rain-qa-func）
multica issue create --title "修复登录 bug" --description "用户报告无法登录..." --project rain-qa-func

# 从管道传入长描述
cat report.md | multica issue create --title "配表提交影响分析报告" --description-stdin --project rain-qa-func
```

创建成功时输出短 ID（如 `MUL-123`）。

---

## 故障排查

| 问题 | 解决方案 |
|------|----------|
| 认证失败 | 检查 `/home/analyzer/.multica/config.json` 是否存在且包含有效 PAT |
| 命令找不到 | 确认 CLI 已安装：`multica version` |
| 权限不足 | 检查 workspace 角色，部分操作需要 admin/owner |

---

## 相关资源

- 官方文档：https://multica.ai/docs
- CLI 文档：https://multica.ai/docs/cli
- 获取帮助：`multica <command> --help`
